package onboard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImports(t *testing.T) {
	dir := t.TempDir()
	wg := filepath.Join(dir, "wg.conf")
	_ = os.WriteFile(wg, []byte("[Peer]\nPublicKey = abcdefghijklmnop\nAllowedIPs = 10.0.0.0/24, 10.0.1.5/32\n"), 0o644)
	out, err := ImportWireGuard(wg)
	if err != nil || len(out) != 2 {
		t.Fatalf("wg %v %v", out, err)
	}
	ts := filepath.Join(dir, "ts.json")
	_ = os.WriteFile(ts, []byte(`{"acls":[{"src":["tag:dev"],"dst":["tag:db:5432"]}]}`), 0o644)
	out, err = ImportTailscale(ts)
	if err != nil || len(out) != 1 {
		t.Fatalf("ts %v %v", out, err)
	}
}
