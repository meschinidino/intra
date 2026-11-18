package providers

import (
	"context"
	"testing"
	"time"
)

func TestSimulatorLifecycle(t *testing.T) {
	s := NewSimulator("entra-sim", 0)
	g, err := s.Issue(context.Background(), "a@b.c", "read", "owner", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	lat, err := s.KillLatency(context.Background(), g.ID)
	if err != nil || lat > 15*time.Minute {
		t.Fatalf("lat %v err %v", lat, err)
	}
	inv, _ := s.Inventory(context.Background())
	if len(inv) != 1 || !inv[0].Revoked {
		t.Fatalf("inventory %+v", inv)
	}
}
