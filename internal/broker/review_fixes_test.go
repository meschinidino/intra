package broker

import (
	"context"
	"testing"

	"intra/internal/audit"
	"intra/internal/policy"
)

func TestDuplicateNeverReturnsTokenAndBindsShape(t *testing.T) {
	s, done := setup(t)
	defer done()
	ctx := context.Background()
	req := Request{RequestKey: "dup", Actor: "owner", Device: "mac", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "a", TTLSeconds: 60}
	g, _, err := s.RequestGrant(ctx, req)
	if err != nil || g.Token() == "" {
		t.Fatalf("first token err=%v token empty=%v", err, g.Token() == "")
	}
	g2, dup, err := s.RequestGrant(ctx, req)
	if err != nil || !dup || g2.Token() != "" || g2.Public().Token() != "" {
		t.Fatalf("duplicate leaked token dup=%v token=%q err=%v", dup, g2.Token(), err)
	}
	changed := req
	changed.Reason = "different"
	if _, _, err := s.RequestGrant(ctx, changed); err == nil {
		t.Fatal("same key with different shape was accepted")
	}
	if got, _ := s.Get(g.ID); got.Token() != "" {
		t.Fatal("Get returned token")
	}
}

func TestTemplateScopeMustFit(t *testing.T) {
	s, done := setup(t)
	defer done()
	// setup has only db. A write action is outside the read-only template.
	_, _, err := s.RequestGrant(context.Background(), Request{RequestKey: "t-out", Actor: "robot", Device: "r", Services: []string{"db"}, Actions: []policy.Action{policy.ActionWrite}, Reason: "x", TemplateID: "robot-read", TTLSeconds: 60})
	if err == nil {
		t.Fatal("write escaped read template")
	}
}

func TestRevokeFlipsStateWhenAuditDown(t *testing.T) {
	s, done := setup(t)
	defer done()
	ctx := context.Background()
	g, _, err := s.RequestGrant(ctx, Request{RequestKey: "rv", Actor: "owner", Device: "mac", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "x", TTLSeconds: 600})
	if err != nil {
		t.Fatal(err)
	}
	sig, err := ApprovalSignature([]byte("approval-key"), s.mustGrant(t, g.ID), "owner")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Approve(ctx, g.ID, "owner", sig, true); err != nil {
		t.Fatal(err)
	}
	s.Audit = &audit.Client{Base: "http://127.0.0.1:1"}
	_ = s.Revoke(ctx, g.ID, "owner")
	got, _ := s.Get(g.ID)
	if got.State == StateActive {
		t.Fatal("grant stayed active while audit was down")
	}
	if s.Validate("owner", "mac", "db", policy.ActionRead, g.Token()) {
		t.Fatal("grant validated during failed revoke")
	}
}

func (s *Server) mustGrant(t *testing.T, id string) Grant {
	t.Helper()
	g, ok := s.Get(id)
	if !ok {
		t.Fatalf("grant %s missing", id)
	}
	return g
}
