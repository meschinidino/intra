package broker

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"intra/internal/audit"
	"intra/internal/policy"
)

func setup(t *testing.T) (*Server, func()) {
	t.Helper()
	dir := t.TempDir()
	w, err := audit.OpenWriter(filepath.Join(dir, "a.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer((&audit.Server{W: w}).Handler())
	st, err := OpenStore(filepath.Join(dir, "grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	p := &policy.Policy{
		Version: 1,
		Actors: map[string]policy.Actor{
			"owner": {Name: "owner", Kind: "human", Owner: "f"},
			"robot": {Name: "robot", Kind: "robot", Owner: "f"},
		},
		Services: map[string]policy.Service{
			"db": {Name: "db", Kind: "postgres", Sensitive: true, Owner: "f", Actions: []policy.Action{policy.ActionRead}},
		},
		Rules: []policy.Rule{
			{Actor: "owner", Service: "db", Actions: []policy.Action{policy.ActionRead}, Access: "grant"},
			{Actor: "robot", Service: "db", Actions: []policy.Action{policy.ActionRead}, Access: "template", TemplateID: "robot-read"},
		},
		Templates: []policy.Template{{ID: "robot-read", Actor: "robot", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, MaxTTLSeconds: 7200}},
	}
	s := New(p, st, &audit.Client{Base: srv.URL}, NoopKiller{}, []byte("approval-key"))
	return s, func() { srv.Close(); _ = w.Close() }
}

func TestGrantFlowDuplicateTouchRevoke(t *testing.T) {
	s, done := setup(t)
	defer done()
	ctx := context.Background()
	req := Request{RequestKey: "k1", Actor: "owner", Device: "mac", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "incident", TTLSeconds: 600}
	g, dup, err := s.RequestGrant(ctx, req)
	if err != nil || dup {
		t.Fatalf("request dup=%v err=%v", dup, err)
	}
	g2, dup, err := s.RequestGrant(ctx, req)
	if err != nil || !dup || g2.ID != g.ID {
		t.Fatalf("duplicate = %s dup=%v err=%v", g2.ID, dup, err)
	}
	if err := s.Approve(ctx, g.ID, "owner", s.SignApproval(g.ID, "owner"), false); err == nil {
		t.Fatal("sensitive approve without Touch ID passed")
	}
	if err := s.Approve(ctx, g.ID, "owner", s.SignApproval(g.ID, "owner"), true); err != nil {
		t.Fatal(err)
	}
	if !s.Validate("owner", "mac", "db", policy.ActionRead, g.Token()) {
		t.Fatal("active grant did not validate")
	}
	if err := s.Revoke(ctx, g.ID, "owner"); err != nil {
		t.Fatal(err)
	}
	if s.Validate("owner", "mac", "db", policy.ActionRead, g.Token()) {
		t.Fatal("revoked grant validates")
	}
	if s.Secret(g.ID) != "" {
		t.Fatal("secret survived revoke")
	}
}

func TestRobotCannotSelfApprove(t *testing.T) {
	s, done := setup(t)
	defer done()
	g, _, err := s.RequestGrant(context.Background(), Request{RequestKey: "r1", Actor: "robot", Device: "run1", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "task", TemplateID: "robot-read", TTLSeconds: 600})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Approve(context.Background(), g.ID, "robot", s.SignApproval(g.ID, "robot"), true); err == nil {
		t.Fatal("robot self-approved")
	}
}

func TestAuditFailureBlocksGrant(t *testing.T) {
	s, done := setup(t)
	defer done()
	s.Audit = &audit.Client{Base: "http://127.0.0.1:1"}
	_, _, err := s.RequestGrant(context.Background(), Request{RequestKey: "k2", Actor: "owner", Device: "mac", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "x", TTLSeconds: 60})
	if err == nil {
		t.Fatal("grant issued while audit down")
	}
}

func TestExpiry(t *testing.T) {
	s, done := setup(t)
	defer done()
	now := time.Now()
	s.Now = func() time.Time { return now }
	g, _, err := s.RequestGrant(context.Background(), Request{RequestKey: "e1", Actor: "owner", Device: "mac", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "x", TTLSeconds: 60})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Approve(context.Background(), g.ID, "owner", s.SignApproval(g.ID, "owner"), true); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	if n := s.SweepExpires(context.Background()); n != 1 {
		t.Fatalf("swept %d", n)
	}
}
