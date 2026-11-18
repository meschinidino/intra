package guard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"intra/internal/policy"
)

type fakeValidator struct{ ok bool }

func (f fakeValidator) Validate(string, string, string, policy.Action, string) bool { return f.ok }

func TestGuardStaticAndGrant(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	defer up.Close()
	p := &policy.Policy{
		Version: 1,
		Actors:  map[string]policy.Actor{"owner": {Name: "owner", Kind: "human", Owner: "f"}, "robot": {Name: "robot", Kind: "robot", Owner: "f"}},
		Services: map[string]policy.Service{
			"normal": {Name: "normal", Kind: "http", Upstream: up.URL, Owner: "f", Actions: []policy.Action{policy.ActionRead}},
		},
		Rules: []policy.Rule{{Actor: "owner", Service: "normal", Actions: []policy.Action{policy.ActionRead}, Access: "static"}},
	}
	g, err := New(p, fakeValidator{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.AllowHeaderID = true
	srv := httptest.NewServer(g.Handler())
	defer srv.Close()

	req, _ := http.NewRequest("GET", srv.URL+"/svc/normal/", nil)
	req.Header.Set("X-Intra-Actor", "owner")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("owner status %v err %v", resp, err)
	}
	req2, _ := http.NewRequest("GET", srv.URL+"/svc/normal/", nil)
	req2.Header.Set("X-Intra-Actor", "robot")
	resp2, _ := http.DefaultClient.Do(req2)
	if resp2.StatusCode != 403 {
		t.Fatalf("robot status %d", resp2.StatusCode)
	}
	req3, _ := http.NewRequest("GET", srv.URL+"/svc/normal/", nil)
	resp3, _ := http.DefaultClient.Do(req3)
	if resp3.StatusCode != 403 {
		t.Fatalf("anon status %d", resp3.StatusCode)
	}
}
