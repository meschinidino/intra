package audit

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnchorCatchesCleanTruncation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.jsonl")
	w, err := OpenWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := w.Append(context.Background(), Event{Actor: "o", Service: "db", Action: "read", Result: "allow"}); err != nil {
			t.Fatal(err)
		}
	}
	_ = w.Close()
	key := []byte("anchor-key")
	anchor := filepath.Join(dir, "anchor.json")
	if err := Anchor(path, anchor, key); err != nil {
		t.Fatal(err)
	}
	f, _ := os.Open(path)
	sc := bufio.NewScanner(f)
	sc.Scan()
	first := sc.Text()
	_ = f.Close()
	if err := os.WriteFile(path, []byte(first+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(first) == "" {
		t.Fatal("empty first line")
	}
	if err := VerifyAnchor(path, anchor, key); err == nil {
		t.Fatal("clean truncation verified against anchor")
	}
	if _, err := OpenWriter(path); err != nil {
		// A truncated-but-valid prefix is a valid chain on its own. The
		// external anchor above is the control that catches the loss.
		t.Fatalf("prefix chain should open: %v", err)
	}
}
