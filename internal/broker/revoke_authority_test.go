package broker

import (
	"context"
	"testing"

	"intra/internal/policy"
)

func TestRevokeRejectsUnknownAndRobotOther(t *testing.T) {
	s, done := setup(t)
	defer done()
	g, _, err := s.RequestGrant(context.Background(), Request{RequestKey: "ra", Actor: "owner", Device: "mac", Services: []string{"db"}, Actions: []policy.Action{policy.ActionRead}, Reason: "x", TTLSeconds: 60})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Revoke(context.Background(), g.ID, "ghost"); err == nil {
		t.Fatal("unknown revoker accepted")
	}
	if err := s.Revoke(context.Background(), g.ID, "robot"); err == nil {
		t.Fatal("robot revoked owner grant")
	}
}
