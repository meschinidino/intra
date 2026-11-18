package policy

import "testing"

func TestEgressTemplateValidatesAndDecides(t *testing.T) {
	p := &Policy{
		Version: 1,
		Actors:  map[string]Actor{"robot": {Name: "robot", Kind: "robot", Owner: "founder"}},
		Services: map[string]Service{
			"internet": {Name: "internet", Kind: "egress", Sensitive: true, Owner: "founder", Actions: []Action{ActionEgress}},
		},
		Rules:     []Rule{{Actor: "robot", Service: "internet", Actions: []Action{ActionEgress}, Access: "template", TemplateID: "robot-egress"}},
		Templates: []Template{{ID: "robot-egress", Actor: "robot", Services: []string{"internet"}, Actions: []Action{ActionEgress}, MaxTTLSeconds: 7200}},
		Egress:    []EgressRule{{Actor: "robot", Hosts: []string{"example.com"}, Ports: []int{443}}},
	}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	d := p.Decide("robot", "internet", ActionEgress)
	if !d.Allowed || d.Access != "template" || d.Rule.TemplateID != "robot-egress" {
		t.Fatalf("decision %+v", d)
	}
	if !p.EgressAllowed("robot", "example.com", 443) {
		t.Fatal("egress host should match")
	}
}
