package broker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreUpgradeLegacyToEnvelopeAndRejectNewer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "grants.json")
	legacy := `[{"id":"g_old","request_key":"k-old","actor":"owner","state":"revoked"}]`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.grants["g_old"]; !ok {
		t.Fatal("legacy grant not loaded")
	}
	st.mu.Lock()
	if err := st.saveLocked(); err != nil {
		st.mu.Unlock()
		t.Fatal(err)
	}
	st.mu.Unlock()
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), `"version": 1`) {
		t.Fatalf("store not upgraded: %s", b)
	}
	st2, err := OpenStore(path)
	if err != nil || len(st2.grants) != 1 {
		t.Fatalf("reopen upgraded store: %v grants=%d", err, len(st2.grants))
	}
	if _, err := ParseStoreBytes([]byte(`{"version":99,"grants":[]}`)); err == nil {
		t.Fatal("newer store format accepted")
	}
	if _, err := ParseStoreBytes([]byte(`{"version":0,"grants":[]}`)); err == nil {
		t.Fatal("malformed envelope accepted")
	}
}
