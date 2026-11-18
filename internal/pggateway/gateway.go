package pggateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"intra/internal/audit"
	"intra/internal/broker"
	"intra/internal/policy"
)

type Gateway struct {
	Broker  *broker.Server
	Base    *pgx.ConnConfig // backend database location and database name
	Service string          // policy service name, default "sensitive-db"
	Audit   *audit.Client
	TLS     *tls.Config // when set, SSLRequest upgrades and client certs identify actor/device

	mu       sync.Mutex
	sessions map[*session]struct{}
}

type session struct {
	gw      *Gateway
	front   *pgproto3.Backend
	conn    net.Conn
	backend *pgx.Conn
	grantID string
	actor   string
	device  string
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	once    sync.Once
}

func (g *Gateway) service() string {
	if g.Service == "" {
		return "sensitive-db"
	}
	return g.Service
}

func (g *Gateway) CloseGrant(id string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for s := range g.sessions {
		if s.grantID == id {
			s.close()
		}
	}
}

func (s *session) close() {
	s.once.Do(func() {
		close(s.done)
		if s.cancel != nil {
			s.cancel()
		}
		if s.conn != nil {
			_ = s.conn.Close()
		}
		// Backend close happens in handle after loop returns. Closing it here
		// would race a live pgx query; ctx cancellation stops that query first.
	})
}

func (g *Gateway) Serve(ln net.Listener) error {
	for {
		c, err := ln.Accept()
		if err != nil {
			return err
		}
		go g.handle(c)
	}
}

func (g *Gateway) handle(conn net.Conn) {
	be := pgproto3.NewBackend(conn, conn)
	var start *pgproto3.StartupMessage
	for {
		msg, err := be.ReceiveStartupMessage()
		if err != nil {
			log.Printf("pggateway startup: %v", err)
			_ = conn.Close()
			return
		}
		switch m := msg.(type) {
		case *pgproto3.StartupMessage:
			start = m
		case *pgproto3.SSLRequest:
			if g.TLS == nil {
				_, _ = conn.Write([]byte("N"))
				continue
			}
			_, _ = conn.Write([]byte("S"))
			tlsConn := tls.Server(conn, g.TLS)
			if err := tlsConn.Handshake(); err != nil {
				log.Printf("pggateway tls handshake: %v", err)
				_ = tlsConn.Close()
				return
			}
			conn = tlsConn
			be = pgproto3.NewBackend(conn, conn)
			continue
		case *pgproto3.GSSEncRequest:
			_, _ = conn.Write([]byte("N"))
			continue
		default:
			_ = conn.Close()
			return
		}
		break
	}
	certActor, certDevice := "", ""
	if tc, ok := conn.(*tls.Conn); ok {
		if pcs := tc.ConnectionState().PeerCertificates; len(pcs) > 0 {
			cn := pcs[0].Subject.CommonName
			if i := strings.Index(cn, "/"); i > 0 {
				certActor, certDevice = cn[:i], cn[i+1:]
			} else {
				certActor = cn
			}
		}
	} else if g.TLS != nil {
		_ = conn.Close()
		return
	}
	grantID := start.Parameters["user"]
	be.Send(&pgproto3.AuthenticationCleartextPassword{})
	_ = be.Flush()
	msg, err := be.Receive()
	if err != nil {
		log.Printf("pggateway password receive: %v", err)
		_ = conn.Close()
		return
	}
	pw, ok := msg.(*pgproto3.PasswordMessage)
	if !ok {
		be.Send(&pgproto3.ErrorResponse{Severity: "FATAL", Code: "28P01", Message: "password required"})
		_ = be.Flush()
		_ = conn.Close()
		return
	}
	grant, ok := g.Broker.ByToken(pw.Password)
	if !ok || grant.ID != grantID || grant.State != broker.StateActive {
		be.Send(&pgproto3.ErrorResponse{Severity: "FATAL", Code: "28P01", Message: "invalid or expired grant"})
		_ = be.Flush()
		_ = conn.Close()
		return
	}
	if certActor != "" && (certActor != grant.Actor || (certDevice != "" && certDevice != grant.Device)) {
		be.Send(&pgproto3.ErrorResponse{Severity: "FATAL", Code: "42501", Message: "certificate identity does not match grant"})
		_ = be.Flush()
		_ = conn.Close()
		return
	}
	if !g.Broker.Validate(grant.Actor, grant.Device, g.service(), policy.ActionRead, pw.Password) {
		be.Send(&pgproto3.ErrorResponse{Severity: "FATAL", Code: "42501", Message: "grant does not cover this database"})
		_ = be.Flush()
		_ = conn.Close()
		return
	}
	secret := g.Broker.Secret(grant.ID)
	if secret == "" || g.Base == nil {
		be.Send(&pgproto3.ErrorResponse{Severity: "FATAL", Code: "08006", Message: "grant session unavailable"})
		_ = be.Flush()
		_ = conn.Close()
		return
	}
	cfg := g.Base.Copy()
	cfg.User = RoleForGrant(grant.ID)
	cfg.Password = secret
	cfg.Database = g.Base.Database // pinned; client database parameter is ignored
	cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	backend, err := pgx.ConnectConfig(context.Background(), cfg)
	if err != nil {
		log.Printf("pggateway backend connect: %v", err)
		be.Send(&pgproto3.ErrorResponse{Severity: "FATAL", Code: "08006", Message: "backend unavailable"})
		_ = be.Flush()
		_ = conn.Close()
		return
	}
	sctx, cancel := context.WithCancel(context.Background())
	s := &session{gw: g, front: be, conn: conn, backend: backend, grantID: grant.ID, actor: grant.Actor, device: grant.Device, ctx: sctx, cancel: cancel, done: make(chan struct{})}
	g.mu.Lock()
	if g.sessions == nil {
		g.sessions = map[*session]struct{}{}
	}
	g.sessions[s] = struct{}{}
	g.mu.Unlock()
	g.Broker.RegisterSession(grant.ID)
	defer func() {
		g.mu.Lock()
		delete(g.sessions, s)
		g.mu.Unlock()
		g.Broker.UnregisterSession(grant.ID)
		s.close()
	}()
	go g.watch(s)

	be.Send(&pgproto3.AuthenticationOk{})
	for k, v := range map[string]string{
		"server_version": "16.0", "server_encoding": "UTF8", "client_encoding": "UTF8",
		"standard_conforming_strings": "on", "integer_datetimes": "on",
		"DateStyle": "ISO, MDY", "TimeZone": "UTC",
	} {
		be.Send(&pgproto3.ParameterStatus{Name: k, Value: v})
	}
	be.Send(&pgproto3.BackendKeyData{ProcessID: 1, SecretKey: []byte{0, 0, 0, 1}})
	be.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	_ = be.Flush()
	s.loop()
	_ = backend.Close(context.Background())
}

func (g *Gateway) watch(s *session) {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-t.C:
			grant, ok := g.Broker.Get(s.grantID)
			if !ok || grant.State != broker.StateActive || time.Now().After(grant.ExpiresAt) {
				s.close()
				return
			}
		}
	}
}

type stmt struct {
	sql    string
	params []string
}

func (s *session) loop() {
	stmts := map[string]*stmt{}
	portals := map[string]*stmt{}
	for {
		msg, err := s.front.Receive()
		if err != nil {
			return
		}
		switch m := msg.(type) {
		case *pgproto3.Query:
			s.exec(m.String, nil)
			s.front.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
			_ = s.front.Flush()
		case *pgproto3.Parse:
			stmts[m.Name] = &stmt{sql: m.Query}
			s.front.Send(&pgproto3.ParseComplete{})
		case *pgproto3.Bind:
			st := stmts[m.PreparedStatement]
			if st == nil {
				st = &stmt{}
			}
			cp := *st
			for _, p := range m.Parameters {
				cp.params = append(cp.params, string(p))
			}
			portals[m.DestinationPortal] = &cp
			s.front.Send(&pgproto3.BindComplete{})
		case *pgproto3.Describe:
			if m.ObjectType == 'S' {
				st := stmts[m.Name]
				n := 0
				if st != nil {
					n = countParams(st.sql)
				}
				oids := make([]uint32, n)
				for i := range oids {
					oids[i] = 25
				}
				s.front.Send(&pgproto3.ParameterDescription{ParameterOIDs: oids})
			}
			s.front.Send(&pgproto3.NoData{})
		case *pgproto3.Execute:
			st := portals[m.Portal]
			if st == nil {
				s.front.Send(&pgproto3.ErrorResponse{Severity: "ERROR", Code: "34000", Message: "unknown portal"})
			} else {
				args := make([]any, len(st.params))
				for i, p := range st.params {
					args[i] = p
				}
				s.exec(st.sql, args)
			}
		case *pgproto3.Sync:
			s.front.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
			_ = s.front.Flush()
		case *pgproto3.Flush:
			_ = s.front.Flush()
		case *pgproto3.Terminate:
			return
		default:
			s.front.Send(&pgproto3.ErrorResponse{Severity: "ERROR", Code: "0A000", Message: "unsupported message"})
			_ = s.front.Flush()
		}
	}
}

func countParams(sql string) int {
	max := 0
	for i := 0; i < len(sql)-1; i++ {
		if sql[i] != '$' {
			continue
		}
		n := 0
		j := i + 1
		for j < len(sql) && sql[j] >= '0' && sql[j] <= '9' {
			n = n*10 + int(sql[j]-'0')
			j++
		}
		if n > max {
			max = n
		}
	}
	return max
}

func multiStatement(sql string) bool {
	trimmed := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(sql), ";"))
	return strings.Contains(trimmed, ";")
}

func (s *session) exec(sql string, args []any) {
	ctx := s.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	action, fingerprint := policy.ClassifySQL(sql)
	upper := strings.ToUpper(strings.TrimSpace(sql))
	if multiStatement(sql) || strings.HasPrefix(upper, "COPY") || strings.HasPrefix(upper, "SET") {
		s.report(action, fingerprint, fmt.Errorf("statement class blocked"))
		s.sendErr(fmt.Errorf("statement class blocked at gateway"))
		return
	}
	if s.gw != nil && s.gw.Broker != nil {
		grant, ok := s.gw.Broker.Get(s.grantID)
		if !ok || grant.State != broker.StateActive {
			s.sendErr(fmt.Errorf("grant no longer active"))
			return
		}
		allowed := false
		for _, a := range grant.Actions {
			if a == action {
				allowed = true
			}
		}
		if !allowed {
			s.report(action, fingerprint, fmt.Errorf("action outside grant"))
			s.sendErr(fmt.Errorf("action outside grant"))
			return
		}
	}
	if s.gw != nil && s.gw.Audit != nil {
		if err := s.gw.Audit.Append(ctx, audit.Event{Actor: s.actor, Device: s.device, Service: s.gw.service(), Action: string(action), Rule: "grant", Result: "allow", GrantID: s.grantID, QueryHash: fingerprint}); err != nil {
			s.sendErr(fmt.Errorf("audit unavailable"))
			return
		}
	}
	isRead := strings.HasPrefix(upper, "SELECT") || strings.HasPrefix(upper, "WITH") || strings.HasPrefix(upper, "SHOW") || strings.HasPrefix(upper, "EXPLAIN")
	if !isRead {
		_, err := s.backend.Exec(ctx, sql, args...)
		s.report(action, fingerprint, err)
		if err != nil {
			s.sendErr(err)
			return
		}
		s.front.Send(&pgproto3.CommandComplete{CommandTag: []byte("OK")})
		_ = s.front.Flush()
		return
	}
	rows, err := s.backend.Query(ctx, sql, args...)
	s.report(action, fingerprint, err)
	if err != nil {
		s.sendErr(err)
		return
	}
	defer rows.Close()
	fields := rows.FieldDescriptions()
	rd := pgproto3.RowDescription{}
	for _, f := range fields {
		rd.Fields = append(rd.Fields, pgproto3.FieldDescription{
			Name: []byte(f.Name), TableOID: f.TableOID, TableAttributeNumber: f.TableAttributeNumber,
			DataTypeOID: f.DataTypeOID, DataTypeSize: f.DataTypeSize, TypeModifier: f.TypeModifier, Format: 0,
		})
	}
	s.front.Send(&rd)
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			s.sendErr(err)
			return
		}
		dr := pgproto3.DataRow{}
		for _, v := range vals {
			if v == nil {
				dr.Values = append(dr.Values, nil)
				continue
			}
			dr.Values = append(dr.Values, []byte(fmt.Sprint(v)))
		}
		s.front.Send(&dr)
	}
	if err := rows.Err(); err != nil {
		s.sendErr(err)
		return
	}
	s.front.Send(&pgproto3.CommandComplete{CommandTag: []byte(rows.CommandTag().String())})
	_ = s.front.Flush()
}

func (s *session) report(action policy.Action, fingerprint string, err error) {
	if s.gw == nil || s.gw.Audit == nil {
		return
	}
	result := "allow"
	if err != nil {
		result = "error"
	}
	_ = s.gw.Audit.Append(context.Background(), audit.Event{
		Actor: s.actor, Service: s.gw.service(), Action: string(action),
		Rule: "grant", Result: result, GrantID: s.grantID, QueryHash: fingerprint,
	})
}

func (s *session) sendErr(err error) {
	e := &pgproto3.ErrorResponse{Severity: "ERROR", Code: "XX000", Message: "database error"}
	if pe, ok := err.(*pgconn.PgError); ok {
		e.Code = pe.Code
		e.Message = pe.Message
		e.Severity = pe.Severity
	}
	s.front.Send(e)
	_ = s.front.Flush()
}
