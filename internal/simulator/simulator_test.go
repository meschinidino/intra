package simulator

import (
	"testing"

	"intra/internal/audit"
	"intra/internal/policy"
)

func TestReplayAndScore(t *testing.T) {
	p := &policy.Policy{Version: 1, Actors: map[string]policy.Actor{"o": {Name: "o"}}, Services: map[string]policy.Service{"db": {Name: "db", Owner: "f", Actions: []policy.Action{policy.ActionRead}}}}
	events := []audit.Event{{Actor: "o", Service: "db", Action: "read", Result: "allow"}}
	r := Replay(events, p)
	if r.FalseBlocks != 1 {
		t.Fatalf("replay %+v", r)
	}
	if score, ok := StandingScore(0, nil); ok || score != 100 {
		t.Fatalf("empty inventory score %v ok=%v", score, ok)
	}
	if score, ok := StandingScore(10, []string{"a", "a"}); ok || score != 10 {
		t.Fatalf("score %v ok=%v", score, ok)
	}
}
