package guard

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"intra/internal/policy"
)

func TestPrivateIPRanges(t *testing.T) {
	for _, s := range []string{"100.64.0.1", "198.18.0.1", "240.0.0.1", "224.0.0.1", "169.254.169.254", "192.0.0.1", "198.51.100.1", "2001:db8::1", "fc00::1", "fe80::1"} {
		if !privateIP(net.ParseIP(s)) {
			t.Errorf("%s should be non-global", s)
		}
	}
	if privateIP(net.ParseIP("93.184.216.34")) {
		t.Fatal("public IP marked private")
	}
}

func TestEgressRejectsForgedHeaderIdentity(t *testing.T) {
	p := &policy.Policy{Version: 1, Actors: map[string]policy.Actor{"robot": {Name: "robot"}}, Services: map[string]policy.Service{}, Egress: []policy.EgressRule{{Actor: "robot", Hosts: []string{"example.com"}, Ports: []int{443}}}}
	e := &Egress{Policy: p}
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("X-Intra-Actor", "robot")
	rec := httptest.NewRecorder()
	e.Handler().ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("forged header status %d", rec.Code)
	}
}
