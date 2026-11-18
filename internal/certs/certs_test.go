package certs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCAIssueAndLoad(t *testing.T) {
	dir := t.TempDir()
	ca, err := MakeCA(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := ca.Issue(dir, "owner", "owner/mac", false); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCA(filepath.Join(dir, "ca.crt"), filepath.Join(dir, "ca.key")); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(filepath.Join(dir, "owner.key"))
	if err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("key mode %v err %v", fi.Mode(), err)
	}
}
