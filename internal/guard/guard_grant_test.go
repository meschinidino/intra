package guard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"intra/internal/policy"
)

func TestGuardGrantPathStripsToken(t *testing.T) {
	var sawAuth, sawActor string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		sawActor = r.Header.Get("X-Intra-Actor")
		w.WriteHeader(200)
	}))
	defer up.Close()
	p := &policy.Policy{
		Version: 1,
		Actors:  map[string]policy.Actor{"owner": {Name: "owner", Kind: "human", Owner: "f"}},
		Services: map[string]policy.Service{
			"sensitive": {Name: "sensitive", Kind: "http", Sensitive: true, Upstream: up.URL, Owner: "f", Actions: []policy.Action{policy.ActionRead}},
		},
		Rules: []policy.Rule{{Actor: "owner", Service: "sensitive", Actions: []policy.Action{policy.ActionRead}, Access: "grant"}},
	}
	g, err := New(p, fakeValidator{ok: true}, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.AllowHeaderID = true
	srv := httptest.NewServer(g.Handler())
	defer srv.Close()
	req, _ := http.NewRequest("GET", srv.URL+"/svc/sensitive/", nil)
	req.Header.Set("X-Intra-Actor", "owner")
	req.Header.Set("Authorization", "Bearer grant-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("status %v err %v", resp, err)
	}
	if sawAuth != "" || sawActor != "" {
		t.Fatalf("upstream leaked headers auth=%q actor=%q", sawAuth, sawActor)
	}

	g2, _ := New(p, fakeValidator{ok: false}, nil)
	g2.AllowHeaderID = true
	srv2 := httptest.NewServer(g2.Handler())
	defer srv2.Close()
	req2, _ := http.NewRequest("GET", srv2.URL+"/svc/sensitive/", nil)
	req2.Header.Set("X-Intra-Actor", "owner")
	resp2, _ := http.DefaultClient.Do(req2)
	if resp2.StatusCode != 403 {
		t.Fatalf("no-grant status %d", resp2.StatusCode)
	}
}
