package broker

import (
	"context"
	"testing"

	"intra/internal/policy"
)

func TestPartialRevokeEscalates(t *testing.T) {
	s, done := setup(t)
	defer done()
	g, _, err := s.RequestGrant(context.Background(), Request{RequestKey: "p1", Actor: "owner", Device: "mac", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "x", TTLSeconds: 600})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Approve(context.Background(), g.ID, "owner", s.SignApproval(g.ID, "owner"), true); err != nil {
		t.Fatal(err)
	}
	s.Killer = NoopKiller{FailKill: map[string]bool{g.ID: true}}
	if err := s.Revoke(context.Background(), g.ID, "owner"); err == nil {
		t.Fatal("revoke should fail")
	}
	got, _ := s.Get(g.ID)
	if got.State != StatePartialRevoke {
		t.Fatalf("state %s", got.State)
	}
}
