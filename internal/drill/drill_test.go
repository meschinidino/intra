package drill

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunDenyAndCanary(t *testing.T) {
	deny := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(403) }))
	defer deny.Close()
	open := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("CANARY-SECRET-123")) }))
	defer open.Close()
	rep := Run(context.Background(), []Target{
		{Name: "closed", URL: deny.URL, Sensitive: true, Owner: "f"},
		{Name: "open", URL: open.URL, Sensitive: true, Owner: "f"},
	}, "CANARY-SECRET-123")
	if !rep.Blocked || rep.Score != 50 {
		t.Fatalf("report %+v", rep)
	}
}
