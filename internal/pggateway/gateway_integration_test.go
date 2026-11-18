//go:build integration

package pggateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"intra/internal/audit"
	"intra/internal/broker"
	"intra/internal/certs"
	"intra/internal/policy"
)

func TestGatewayGrantLifecycle(t *testing.T) {
	adminURL := os.Getenv("INTRA_TEST_PG_ADMIN_URL")
	baseURL := os.Getenv("INTRA_TEST_PG_BASE_URL")
	if adminURL == "" || baseURL == "" {
		t.Skip("set INTRA_TEST_PG_ADMIN_URL and INTRA_TEST_PG_BASE_URL")
	}
	ctx := context.Background()
	admin, err := ConnectAdmin(ctx, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if _, err := admin.Exec(ctx, "SELECT value FROM secrets"); err == nil {
		t.Fatal("broker admin can read protected data")
	}

	dir := t.TempDir()
	aw, err := audit.OpenWriter(filepath.Join(dir, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	defer aw.Close()
	asrv := httptest.NewServer((&audit.Server{W: aw}).Handler())
	defer asrv.Close()
	st, err := broker.OpenStore(filepath.Join(dir, "grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	p := &policy.Policy{
		Version: 1,
		Actors:  map[string]policy.Actor{"owner": {Name: "owner", Kind: "human", Owner: "f"}},
		Services: map[string]policy.Service{
			"sensitive-db": {Name: "sensitive-db", Kind: "postgres", Sensitive: true, Owner: "f", Actions: []policy.Action{policy.ActionRead, policy.ActionWrite}},
		},
		Rules: []policy.Rule{{Actor: "owner", Service: "sensitive-db", Actions: []policy.Action{policy.ActionRead, policy.ActionWrite}, Access: "grant"}},
	}
	certDir := filepath.Join(dir, "certs")
	ca, err := certs.MakeCA(certDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range []struct {
		name, cn string
		server   bool
	}{{"guard", "localhost", true}, {"owner", "owner/mac", false}, {"robot", "robot/run1", false}} {
		if err := ca.Issue(certDir, spec.name, spec.cn, spec.server); err != nil {
			t.Fatal(err)
		}
	}
	serverCert, err := tls.LoadX509KeyPair(filepath.Join(certDir, "guard.crt"), filepath.Join(certDir, "guard.key"))
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca.PEM)
	gw := &Gateway{Audit: &audit.Client{Base: asrv.URL}, TLS: &tls.Config{Certificates: []tls.Certificate{serverCert}, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool, MinVersion: tls.VersionTLS12}}
	killer := &PGKiller{Admin: admin, Gateway: gw}
	b := broker.New(p, st, &audit.Client{Base: asrv.URL}, killer, []byte("k"))
	gw.Broker = b
	base, err := ParseBase(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	gw.Base = base
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() { _ = gw.Serve(ln) }()

	g, _, err := b.RequestGrant(ctx, broker.Request{RequestKey: "it1", Actor: "owner", Device: "mac", Services: []string{"sensitive-db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "test", TTLSeconds: 600})
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Approve(ctx, g.ID, "owner", b.SignApproval(g.ID, "owner"), true); err != nil {
		t.Fatal(err)
	}

	cfg, err := pgx.ParseConfig(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	host, port, _ := net.SplitHostPort(ln.Addr().String())
	cfg.Host = host
	pn, _ := net.LookupPort("tcp", port)
	cfg.Port = uint16(pn)
	cfg.User = g.ID
	cfg.Password = g.Token()
	clientCert, err := tls.LoadX509KeyPair(filepath.Join(certDir, "owner.crt"), filepath.Join(certDir, "owner.key"))
	if err != nil {
		t.Fatal(err)
	}
	cfg.TLSConfig = &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{clientCert}, ServerName: "localhost", MinVersion: tls.VersionTLS12}
	wrong := cfg.Copy()
	robotCert, err := tls.LoadX509KeyPair(filepath.Join(certDir, "robot.crt"), filepath.Join(certDir, "robot.key"))
	if err != nil {
		t.Fatal(err)
	}
	wrong.TLSConfig = &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{robotCert}, ServerName: "localhost", MinVersion: tls.VersionTLS12}
	if bad, err := pgx.ConnectConfig(ctx, wrong); err == nil {
		_ = bad.Close(ctx)
		t.Fatal("robot certificate opened owner grant")
	}
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect through gateway: %v", err)
	}
	var v string
	if err := conn.QueryRow(ctx, "SELECT value FROM secrets WHERE id=1").Scan(&v); err != nil {
		t.Fatalf("read: %v", err)
	}
	if v != "CANARY-SECRET-123" {
		t.Fatalf("canary %q", v)
	}
	if _, err := conn.Exec(ctx, "INSERT INTO secrets VALUES (2,'x')"); err == nil {
		t.Fatal("write succeeded under read role")
	}
	// Prove the role itself, not only the gateway classifier, denies writes.
	directCfg, _ := pgx.ParseConfig(baseURL)
	directCfg.User = RoleForGrant(g.ID)
	directCfg.Password = b.Secret(g.ID)
	direct, err := pgx.ConnectConfig(ctx, directCfg)
	if err != nil {
		t.Fatalf("direct role connect: %v", err)
	}
	if _, err := direct.Exec(ctx, "INSERT INTO secrets VALUES (3,'y')"); err == nil {
		t.Fatal("direct write succeeded under grant role")
	}
	_ = direct.Close(ctx)

	done := make(chan error, 1)
	go func() {
		var v string
		done <- conn.QueryRow(context.Background(), "SELECT pg_sleep(5)::text").Scan(&v)
	}()
	time.Sleep(300 * time.Millisecond)
	killStart := time.Now()
	if err := b.Revoke(ctx, g.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("live query survived revoke")
		}
		if time.Since(killStart) > 2*time.Second {
			t.Fatalf("kill took %v, want early termination", time.Since(killStart))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("live query still running after revoke")
	}
	_ = conn.Close(ctx)

	if _, err := pgx.ConnectConfig(ctx, cfg); err == nil {
		t.Fatal("revoked grant reconnected")
	}
	if err := audit.Verify(filepath.Join(dir, "audit.jsonl")); err != nil {
		t.Fatal(err)
	}
}
