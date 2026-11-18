package policy

import "testing"

func testPolicy() *Policy {
	return &Policy{
		Version: 1,
		Actors: map[string]Actor{
			"owner": {Name: "owner", Kind: "human", Owner: "f"},
			"robot": {Name: "robot", Kind: "robot", Owner: "f"},
		},
		Services: map[string]Service{
			"db": {Name: "db", Kind: "postgres", Sensitive: true, Owner: "f", Actions: []Action{ActionRead, ActionWrite}},
		},
		Rules:  []Rule{{Actor: "owner", Service: "db", Actions: []Action{ActionRead}, Access: "grant"}},
		Egress: []EgressRule{{Actor: "robot", Hosts: []string{"example.com"}, Ports: []int{443}}},
	}
}

func TestDecide(t *testing.T) {
	p := testPolicy()
	if d := p.Decide("owner", "db", ActionRead); !d.Allowed || d.Access != "grant" {
		t.Fatalf("owner read = %+v", d)
	}
	if d := p.Decide("robot", "db", ActionRead); d.Allowed {
		t.Fatalf("robot read allowed")
	}
	if d := p.Decide("ghost", "db", ActionRead); d.Allowed {
		t.Fatalf("ghost allowed")
	}
}

func TestClassifySQLNoLiteralLeak(t *testing.T) {
	a, fp := ClassifySQL("SELECT * FROM secrets WHERE value='CANARY-SECRET-123'")
	if a != ActionRead || fp == "" {
		t.Fatalf("classify = %s %q", a, fp)
	}
	a, _ = ClassifySQL("DELETE FROM secrets")
	if a != ActionWrite {
		t.Fatalf("delete class = %s", a)
	}
}

func TestCompileDeterministic(t *testing.T) {
	p := testPolicy()
	c1, err := p.Compile()
	if err != nil {
		t.Fatal(err)
	}
	c2, _ := p.Compile()
	if c1.Hash != c2.Hash {
		t.Fatal("hash drift")
	}
}
