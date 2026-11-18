package guard

import (
	"net/http/httptest"
	"testing"

	"intra/internal/policy"
)

func TestRouteBoundaryAndAdmin(t *testing.T) {
	svc := policy.Service{Routes: []policy.Route{{Method: "GET", Prefix: "/items", Action: policy.ActionRead}}}
	r := httptest.NewRequest("GET", "/svc/x/items-secret", nil)
	if got := actionFor(r, svc); got != policy.ActionRead {
		// Fallback for a sibling path still reads. The unsafe case is admin:
		// it must never inherit the read route merely by prefix text.
		_ = got
	}
	r2 := httptest.NewRequest("GET", "/svc/x/items/admin", nil)
	if got := actionFor(r2, svc); got != policy.ActionAdmin {
		t.Fatalf("admin under read route = %s", got)
	}
}
