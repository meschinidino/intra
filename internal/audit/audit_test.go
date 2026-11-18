package audit

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestChainConcurrentAppendAndAnchor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.jsonl")
	w, err := OpenWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = w.Append(context.Background(), Event{Actor: "owner", Service: "db", Action: "read", Result: "allow"})
		}()
	}
	wg.Wait()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := Verify(path); err != nil {
		t.Fatal(err)
	}
	key := []byte("test-anchor-key")
	anchor := filepath.Join(dir, "anchor.json")
	if err := Anchor(path, anchor, key); err != nil {
		t.Fatal(err)
	}
	if err := VerifyAnchor(path, anchor, key); err != nil {
		t.Fatal(err)
	}
	// Tamper: rewrite history, anchor must catch it.
	b, _ := os.ReadFile(path)
	_ = os.WriteFile(path, b[:len(b)/2], 0o600)
	if err := VerifyAnchor(path, anchor, key); err == nil {
		t.Fatal("tampered chain verified")
	}
}
