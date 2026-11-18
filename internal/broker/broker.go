package broker

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"intra/internal/audit"
	"intra/internal/policy"
)

type State string

const (
	StateRequested      State = "requested"
	StateApproved       State = "approved"
	StateActive         State = "active"
	StateDenied         State = "denied"
	StateExpired        State = "expired"
	StateRevoked        State = "revoked"
	StatePartialRevoke  State = "partial_revoke"
	StateRevokeEscalate State = "revoke_escalated"
	StateRevoking       State = "revoking"
	StateIssuing        State = "issuing"
)

type Grant struct {
	ID          string          `json:"id"`
	RequestKey  string          `json:"request_key"`
	Actor       string          `json:"actor"`
	Device      string          `json:"device"`
	Services    []string        `json:"services"`
	Actions     []policy.Action `json:"actions"`
	Reason      string          `json:"reason"`
	IncidentID  string          `json:"incident_id,omitempty"`
	TemplateID  string          `json:"template_id,omitempty"`
	State       State           `json:"state"`
	CreatedAt   time.Time       `json:"created_at"`
	ExpiresAt   time.Time       `json:"expires_at"`
	TokenHash   string          `json:"token_hash"`
	RequestHash string          `json:"request_hash"`
	DBRole      string          `json:"db_role,omitempty"`
	Children    []Child         `json:"children,omitempty"`
	Approver    string          `json:"approver,omitempty"`
	RevokeTries int             `json:"revoke_tries,omitempty"`

	secret string // DB password, memory only, never persisted
	token  string // shown to requester once in memory response only
}

type Child struct {
	Service string `json:"service"`
	State   State  `json:"state"`
}

func (g Grant) Public() Grant { g.secret = ""; g.token = ""; return g }
func (g Grant) Token() string { return g.token }

type SessionKiller interface {
	OpenGrant(ctx context.Context, g Grant) (secret string, err error)
	KillGrant(ctx context.Context, g Grant) error
}

type NoopKiller struct{ FailKill map[string]bool }

func (n NoopKiller) OpenGrant(context.Context, Grant) (string, error) { return randHex(24) }
func (n NoopKiller) KillGrant(_ context.Context, g Grant) error {
	if n.FailKill != nil && n.FailKill[g.ID] {
		return fmt.Errorf("kill failed")
	}
	return nil
}

type Store struct {
	mu     sync.Mutex
	path   string
	grants map[string]*Grant
	byKey  map[string]string
}

type diskGrant struct {
	Grant
}

// StoreFormatVersion is the current grant-store envelope version.
// Version 0 was the legacy bare JSON grant list from v0.1.0.
const StoreFormatVersion = 1

type storeEnvelope struct {
	Version int     `json:"version"`
	Grants  []Grant `json:"grants"`
}

// ParseStoreBytes decodes a grant store in legacy v0 or envelope v1 form.
// It rejects any newer format so an old binary never misreads a new store.
func ParseStoreBytes(b []byte) ([]Grant, error) {
	trimmed := bytesTrimSpace(b)
	if len(trimmed) == 0 {
		return nil, nil
	}
	if trimmed[0] == '[' {
		var list []Grant
		if err := json.Unmarshal(b, &list); err != nil {
			return nil, err
		}
		return list, nil
	}
	var env storeEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, err
	}
	if env.Version > StoreFormatVersion {
		return nil, fmt.Errorf("grant store format %d is newer than supported %d; upgrade Intra before opening this store", env.Version, StoreFormatVersion)
	}
	if env.Version < 1 {
		return nil, fmt.Errorf("malformed grant store envelope version %d", env.Version)
	}
	return env.Grants, nil
}

// OpenStoreBytes parses store bytes without opening a file. Backup restore
// uses it to validate a backed-up store before install.
func OpenStoreBytes(b []byte) ([]Grant, error) { return ParseStoreBytes(b) }

func bytesTrimSpace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\n' || b[0] == '\r' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}

func OpenStore(path string) (*Store, error) {
	s := &Store{path: path, grants: map[string]*Grant{}, byKey: map[string]string{}}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	list, err := ParseStoreBytes(b)
	if err != nil {
		return nil, err
	}
	for i := range list {
		g := list[i]
		s.grants[g.ID] = &g
		s.byKey[g.RequestKey] = g.ID
	}
	return s, nil
}

func (s *Store) saveLocked() error {
	list := make([]Grant, 0, len(s.grants))
	for _, g := range s.grants {
		list = append(list, g.Public())
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	b, err := json.MarshalIndent(storeEnvelope{Version: StoreFormatVersion, Grants: list}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	// Write in place: the lab mounts grants.json as one file, so a sibling
	// temp file may not be creatable in a read-only assembled /data view.
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

type Server struct {
	Policy       *policy.Policy
	Store        *Store
	Audit        *audit.Client
	Killer       SessionKiller
	ApprovalKey  []byte
	AdminToken   string
	LocalTouch   func() bool
	Now          func() time.Time
	mu           sync.Mutex
	sessions     map[string]int
	KillDeadline time.Duration
}

func New(p *policy.Policy, st *Store, au *audit.Client, killer SessionKiller, key []byte) *Server {
	if killer == nil {
		killer = NoopKiller{}
	}
	return &Server{Policy: p, Store: st, Audit: au, Killer: killer, ApprovalKey: key, Now: time.Now, sessions: map[string]int{}, KillDeadline: time.Minute}
}

func randHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }

type Request struct {
	RequestKey string          `json:"request_key"`
	Actor      string          `json:"actor"`
	Device     string          `json:"device"`
	Services   []string        `json:"services"`
	Actions    []policy.Action `json:"actions"`
	Reason     string          `json:"reason"`
	IncidentID string          `json:"incident_id"`
	TemplateID string          `json:"template_id"`
	TTLSeconds int             `json:"ttl_seconds"`
}

func requestHash(req Request) string {
	svcs := append([]string{}, req.Services...)
	sort.Strings(svcs)
	acts := make([]string, 0, len(req.Actions))
	for _, a := range req.Actions {
		acts = append(acts, string(a))
	}
	sort.Strings(acts)
	raw := strings.Join([]string{req.Actor, req.Device, strings.Join(svcs, ","), strings.Join(acts, ","), req.Reason, req.IncidentID, req.TemplateID, fmt.Sprint(req.TTLSeconds)}, "|")
	return hash(raw)
}

func maxTTL(p *policy.Policy, req Request, decisions []policy.Decision) int {
	sensitive := false
	for _, s := range req.Services {
		if svc, ok := p.Services[s]; ok && svc.Sensitive {
			sensitive = true
		}
	}
	actor := p.Actors[req.Actor]
	capTTL := 24 * 3600
	if sensitive || actor.Kind == "robot" {
		capTTL = 8 * 3600
	}
	if req.IncidentID != "" && capTTL > 8*3600 {
		capTTL = 8 * 3600
	}
	for _, d := range decisions {
		if d.Rule != nil && d.Rule.MaxTTLSeconds > 0 && d.Rule.MaxTTLSeconds < capTTL {
			capTTL = d.Rule.MaxTTLSeconds
		}
	}
	if req.TemplateID != "" {
		if t := p.Template(req.TemplateID); t != nil && t.MaxTTLSeconds < capTTL {
			capTTL = t.MaxTTLSeconds
		}
	}
	return capTTL
}

func ApprovalSignature(key []byte, g Grant, approver string) (string, error) {
	if len(key) == 0 {
		return "", fmt.Errorf("approval key unavailable")
	}
	svcs := append([]string{}, g.Services...)
	sort.Strings(svcs)
	acts := make([]string, 0, len(g.Actions))
	for _, a := range g.Actions {
		acts = append(acts, string(a))
	}
	sort.Strings(acts)
	m := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(m, "%s:%s:%s:%s:%s:%s:%s:%s", g.ID, approver, g.Actor, g.Device, strings.Join(svcs, ","), strings.Join(acts, ","), g.ExpiresAt.UTC().Format(time.RFC3339), g.TemplateID)
	return hex.EncodeToString(m.Sum(nil)), nil
}

func (s *Server) audit(ctx context.Context, e audit.Event) error {
	if s.Audit == nil {
		return fmt.Errorf("audit unavailable")
	}
	return s.Audit.Append(ctx, e)
}

func (s *Server) RequestGrant(ctx context.Context, req Request) (Grant, bool, error) {
	if strings.TrimSpace(req.RequestKey) == "" || strings.TrimSpace(req.Reason) == "" || len(req.Services) == 0 || len(req.Actions) == 0 {
		return Grant{}, false, fmt.Errorf("request_key, reason, services and actions are required")
	}
	reqHash := requestHash(req)
	s.Store.mu.Lock()
	if id, ok := s.Store.byKey[req.RequestKey]; ok {
		g := *s.Store.grants[id]
		s.Store.mu.Unlock()
		if g.RequestHash != reqHash {
			return Grant{}, false, fmt.Errorf("request_key already used for a different request")
		}
		return g.Public(), true, nil
	}
	s.Store.mu.Unlock()

	access := ""
	var decisions []policy.Decision
	for _, svc := range req.Services {
		for _, act := range req.Actions {
			d := s.Policy.Decide(req.Actor, svc, act)
			if !d.Allowed {
				_ = s.audit(ctx, audit.Event{Actor: req.Actor, Device: req.Device, Service: svc, Action: string(act), Rule: "none", Result: "deny:" + d.Reason})
				return Grant{}, false, fmt.Errorf("denied: %s", d.Reason)
			}
			if req.TemplateID != "" && (d.Access != "template" || d.Rule == nil || d.Rule.TemplateID != req.TemplateID) {
				return Grant{}, false, fmt.Errorf("template does not cover this request")
			}
			decisions = append(decisions, d)
			if access == "" {
				access = d.Access
			}
		}
	}
	if req.TemplateID != "" {
		t := s.Policy.Template(req.TemplateID)
		if t == nil || t.Actor != req.Actor {
			return Grant{}, false, fmt.Errorf("template does not cover this request")
		}
		for _, svc := range req.Services {
			found := false
			for _, allowed := range t.Services {
				if allowed == svc {
					found = true
				}
			}
			if !found {
				return Grant{}, false, fmt.Errorf("template does not cover service %s", svc)
			}
		}
		for _, act := range req.Actions {
			found := false
			for _, allowed := range t.Actions {
				if allowed == act {
					found = true
				}
			}
			if !found {
				return Grant{}, false, fmt.Errorf("template does not cover action %s", act)
			}
		}
		access = "template"
	}
	capTTL := maxTTL(s.Policy, req, decisions)
	ttl := req.TTLSeconds
	if ttl <= 0 || ttl > capTTL {
		return Grant{}, false, fmt.Errorf("ttl_seconds must be 1..%d", capTTL)
	}
	id, err := randHex(8)
	if err != nil {
		return Grant{}, false, err
	}
	token, err := randHex(24)
	if err != nil {
		return Grant{}, false, err
	}
	now := s.Now()
	g := &Grant{ID: "g_" + id, RequestKey: req.RequestKey, Actor: req.Actor, Device: req.Device,
		Services: append([]string{}, req.Services...), Actions: append([]policy.Action{}, req.Actions...),
		Reason: req.Reason, IncidentID: req.IncidentID, TemplateID: req.TemplateID,
		State: StateRequested, CreatedAt: now, ExpiresAt: now.Add(time.Duration(ttl) * time.Second),
		TokenHash: hash(token), RequestHash: reqHash, token: token}
	for _, svc := range req.Services {
		g.Children = append(g.Children, Child{Service: svc, State: StateRequested})
	}
	if access == "static" || access == "template" {
		// Approval stays explicit even for templates; templates only bound scope.
	}
	// Audit before persisting an issuable grant. D13: audit failure blocks issue.
	if err := s.audit(ctx, audit.Event{Actor: req.Actor, Device: req.Device, Service: strings.Join(req.Services, ","), Action: "request", Rule: access, Result: "requested", GrantID: g.ID}); err != nil {
		return Grant{}, false, fmt.Errorf("audit paused: %w", err)
	}
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	if id, ok := s.Store.byKey[req.RequestKey]; ok {
		existing := s.Store.grants[id]
		if existing.RequestHash != reqHash {
			return Grant{}, false, fmt.Errorf("request_key already used for a different request")
		}
		return existing.Public(), true, nil
	}
	s.Store.grants[g.ID] = g
	s.Store.byKey[g.RequestKey] = g.ID
	if err := s.Store.saveLocked(); err != nil {
		delete(s.Store.grants, g.ID)
		delete(s.Store.byKey, g.RequestKey)
		return Grant{}, false, err
	}
	out := g.Public()
	out.token = token
	return out, false, nil
}

func (s *Server) SignApproval(grantID, approver string) string {
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	g, ok := s.Store.grants[grantID]
	if !ok {
		return ""
	}
	sig, _ := ApprovalSignature(s.ApprovalKey, *g, approver)
	return sig
}

func (s *Server) Approve(ctx context.Context, id, approver, sig string, sensitiveOK bool) error {
	actor, known := s.Policy.Actors[approver]
	if !known || actor.Kind != "human" {
		return fmt.Errorf("approver must be a known human")
	}
	s.Store.mu.Lock()
	g, ok := s.Store.grants[id]
	if !ok {
		s.Store.mu.Unlock()
		return fmt.Errorf("unknown grant")
	}
	if g.State != StateRequested {
		s.Store.mu.Unlock()
		return fmt.Errorf("grant is %s", g.State)
	}
	expected, err := ApprovalSignature(s.ApprovalKey, *g, approver)
	if err != nil || !hmac.Equal([]byte(sig), []byte(expected)) {
		s.Store.mu.Unlock()
		return fmt.Errorf("invalid approval signature")
	}
	sensitive := false
	hasPostgres := false
	for _, name := range g.Services {
		if s.Policy.Services[name].Sensitive {
			sensitive = true
		}
		if s.Policy.Services[name].Kind == "postgres" {
			hasPostgres = true
		}
	}
	if sensitive && s.LocalTouch != nil {
		sensitiveOK = s.LocalTouch()
	}
	if sensitive && !sensitiveOK {
		s.Store.mu.Unlock()
		return fmt.Errorf("Touch ID confirmation required")
	}
	if approver == g.Actor && s.Policy.Actors[g.Actor].Kind == "robot" {
		s.Store.mu.Unlock()
		return fmt.Errorf("robots cannot self-approve")
	}
	g.State = StateIssuing
	snapshot := *g
	if err := s.Store.saveLocked(); err != nil {
		s.Store.mu.Unlock()
		return err
	}
	s.Store.mu.Unlock()

	secret := ""
	if hasPostgres {
		secret, err = s.Killer.OpenGrant(ctx, snapshot)
		if err != nil {
			s.setState(id, StateRequested)
			return err
		}
	}
	if err := s.audit(ctx, audit.Event{Actor: approver, Service: strings.Join(snapshot.Services, ","), Action: "approve", Rule: "approval", Result: "approved", GrantID: snapshot.ID}); err != nil {
		if hasPostgres {
			_ = s.Killer.KillGrant(context.Background(), snapshot)
		}
		s.setState(id, StateRequested)
		return fmt.Errorf("audit paused: %w", err)
	}

	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	g = s.Store.grants[id]
	g.secret = secret
	g.Approver = approver
	g.State = StateActive
	for i := range g.Children {
		g.Children[i].State = StateActive
	}
	if err := s.Store.saveLocked(); err != nil {
		g.State = StateIssuing
		return err
	}
	return nil
}

func (s *Server) setState(id string, state State) {
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	if g, ok := s.Store.grants[id]; ok {
		g.State = state
		_ = s.Store.saveLocked()
	}
}

// RestorePostgresSecrets re-issues in-memory secrets for active database
// grants after a broker restart. Persisted grants never store the secret.
func (s *Server) RestorePostgresSecrets(ctx context.Context) error {
	s.Store.mu.Lock()
	var snaps []Grant
	for _, g := range s.Store.grants {
		if g.State != StateActive {
			continue
		}
		for _, name := range g.Services {
			if s.Policy.Services[name].Kind == "postgres" {
				snaps = append(snaps, *g)
				break
			}
		}
	}
	s.Store.mu.Unlock()
	for _, snap := range snaps {
		secret, err := s.Killer.OpenGrant(ctx, snap)
		if err != nil {
			return err
		}
		s.Store.mu.Lock()
		if g, ok := s.Store.grants[snap.ID]; ok {
			g.secret = secret
		}
		s.Store.mu.Unlock()
	}
	return nil
}

func (s *Server) Secret(id string) string {
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	if g, ok := s.Store.grants[id]; ok {
		return g.secret
	}
	return ""
}

func (s *Server) SetDBRole(id, role string) {
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	if g, ok := s.Store.grants[id]; ok {
		g.DBRole = role
		_ = s.Store.saveLocked()
	}
}

func (s *Server) Validate(actor, device, service string, action policy.Action, token string) bool {
	now := s.Now()
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	want := hash(token)
	for _, g := range s.Store.grants {
		if !hmac.Equal([]byte(g.TokenHash), []byte(want)) || g.Actor != actor {
			continue
		}
		if device == "" || g.Device != device {
			return false
		}
		if g.State != StateActive || now.After(g.ExpiresAt) {
			return false
		}
		for _, svc := range g.Services {
			if svc != service {
				continue
			}
		}
		foundSvc := false
		for _, svc := range g.Services {
			if svc == service {
				foundSvc = true
			}
		}
		if !foundSvc {
			return false
		}
		for _, a := range g.Actions {
			if a == action {
				return true
			}
		}
	}
	return false
}

func (s *Server) ByToken(token string) (Grant, bool) {
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	want := hash(token)
	for _, g := range s.Store.grants {
		if hmac.Equal([]byte(g.TokenHash), []byte(want)) {
			return *g, true
		}
	}
	return Grant{}, false
}

func (s *Server) RegisterSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id]++
}
func (s *Server) UnregisterSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessions[id] > 0 {
		s.sessions[id]--
	}
}
func (s *Server) Sessions(id string) int { s.mu.Lock(); defer s.mu.Unlock(); return s.sessions[id] }

func (s *Server) Revoke(ctx context.Context, id, by string) error {
	s.Store.mu.Lock()
	g, ok := s.Store.grants[id]
	if !ok {
		s.Store.mu.Unlock()
		return fmt.Errorf("unknown grant")
	}
	if g.State == StateRevoked || g.State == StateExpired {
		s.Store.mu.Unlock()
		return nil
	}
	if by != "system" {
		revoker, known := s.Policy.Actors[by]
		if !known || (revoker.Kind != "human" && by != g.Actor) {
			s.Store.mu.Unlock()
			return fmt.Errorf("revoker not authorized")
		}
	}
	// Flip and persist a non-usable state before any external call. Validate
	// only accepts StateActive, so new work stops even if audit is down.
	g.State = StateRevoking
	g.secret = ""
	if err := s.Store.saveLocked(); err != nil {
		s.Store.mu.Unlock()
		return err
	}
	snapshot := *g
	s.Store.mu.Unlock()

	auditErr := s.audit(ctx, audit.Event{Actor: by, Service: strings.Join(snapshot.Services, ","), Action: "revoke", Rule: "revoke", Result: "requested", GrantID: id})
	kctx, cancel := context.WithTimeout(ctx, s.KillDeadline)
	defer cancel()
	err := s.Killer.KillGrant(kctx, snapshot)

	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	g = s.Store.grants[id]
	g.RevokeTries++
	if err != nil {
		if g.RevokeTries >= 3 {
			g.State = StateRevokeEscalate
		} else {
			g.State = StatePartialRevoke
		}
		for i := range g.Children {
			if s.Policy.Services[g.Children[i].Service].Kind == "postgres" {
				g.Children[i].State = StatePartialRevoke
			} else {
				g.Children[i].State = StateRevoked
			}
		}
		_ = s.Store.saveLocked()
		s.Store.mu.Unlock()
		_ = s.audit(ctx, audit.Event{Actor: by, Service: strings.Join(snapshot.Services, ","), Action: "revoke", Rule: "revoke", Result: "partial_revoke_alert", GrantID: id})
		s.Store.mu.Lock()
		return fmt.Errorf("partial revoke: %w", err)
	}
	g.State = StateRevoked
	g.secret = ""
	for i := range g.Children {
		g.Children[i].State = StateRevoked
	}
	if err := s.Store.saveLocked(); err != nil {
		return err
	}
	if auditErr != nil {
		return fmt.Errorf("revoked but audit unavailable: %w", auditErr)
	}
	return nil
}

// SweepExpires moves active grants past expiry through KillGrant.
func (s *Server) SweepExpires(ctx context.Context) int {
	now := s.Now()
	var ids []string
	s.Store.mu.Lock()
	for id, g := range s.Store.grants {
		if g.State == StateActive && now.After(g.ExpiresAt) {
			ids = append(ids, id)
		}
	}
	s.Store.mu.Unlock()
	n := 0
	for _, id := range ids {
		if err := s.Revoke(ctx, id, "system"); err == nil {
			s.Store.mu.Lock()
			s.Store.grants[id].State = StateExpired
			_ = s.Store.saveLocked()
			s.Store.mu.Unlock()
			n++
		}
	}
	return n
}

func (s *Server) Get(id string) (Grant, bool) {
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	g, ok := s.Store.grants[id]
	if !ok {
		return Grant{}, false
	}
	return g.Public(), true
}

func (s *Server) List() []Grant {
	s.Store.mu.Lock()
	defer s.Store.mu.Unlock()
	out := make([]Grant, 0, len(s.Store.grants))
	for _, g := range s.Store.grants {
		out = append(out, g.Public())
	}
	return out
}
