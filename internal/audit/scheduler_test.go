package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type recordSink struct{ alerts []Alert }

func (r *recordSink) Send(_ context.Context, a Alert) error {
	r.alerts = append(r.alerts, a)
	return nil
}

func TestAnchorOnceStaleThenFreshAndWebhook(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	w, err := OpenWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Append(context.Background(), Event{Actor: "owner", Action: "read", Result: "allow"}); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	anchor := filepath.Join(dir, "anchor.json")
	key := []byte("k")
	sink := &recordSink{}
	rep, err := AnchorOnce(context.Background(), path, anchor, key, time.Hour, sink, time.Now())
	if err != nil || !rep.Anchored {
		t.Fatalf("rep=%+v err=%v", rep, err)
	}
	rep, err = AnchorOnce(context.Background(), path, anchor, key, time.Hour, sink, time.Now())
	if err != nil || rep.Anchored {
		t.Fatalf("fresh rep=%+v err=%v", rep, err)
	}
	// Webhook sink delivers.
	var got int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got++; w.WriteHeader(204) }))
	defer srv.Close()
	if err := (WebhookSink{URL: srv.URL}).Send(context.Background(), Alert{Kind: "test"}); err != nil || got != 1 {
		t.Fatalf("webhook got=%d err=%v", got, err)
	}
	// File sink appends.
	fs := FileSink{Path: filepath.Join(dir, "alerts.jsonl")}
	if err := fs.Send(context.Background(), Alert{Kind: "anchor_failed"}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(fs.Path)
	if len(b) == 0 {
		t.Fatal("no alert written")
	}
}

func TestAnchorOnceBrokenChainAlertsAndRefuses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	if err := os.WriteFile(path, []byte("not json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	sink := &recordSink{}
	if _, err := AnchorOnce(context.Background(), path, filepath.Join(dir, "a.json"), []byte("k"), time.Hour, sink, time.Now()); err == nil {
		t.Fatal("broken chain anchored")
	}
	if len(sink.alerts) != 1 || sink.alerts[0].Kind != "anchor_failed" {
		t.Fatalf("alerts %+v", sink.alerts)
	}
}
