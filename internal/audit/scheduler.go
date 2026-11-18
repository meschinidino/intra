package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Alert is one operator alert. It holds metadata only.
type Alert struct {
	TS      time.Time `json:"ts"`
	Kind    string    `json:"kind"`
	Detail  string    `json:"detail"`
	Seq     uint64    `json:"seq,omitempty"`
	StaleBy string    `json:"stale_by,omitempty"`
}

type Sink interface {
	Send(ctx context.Context, a Alert) error
}

// FileSink appends alerts as JSON lines to an owner-only file.
type FileSink struct{ Path string }

func (s FileSink) Send(_ context.Context, a Alert) error {
	if s.Path == "" {
		return fmt.Errorf("alert file path empty")
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	b, _ := json.Marshal(a)
	f, err := os.OpenFile(s.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}

// WebhookSink posts alerts as JSON. A non-2xx response is an error.
type WebhookSink struct {
	URL    string
	Client *http.Client
}

func (s WebhookSink) Send(ctx context.Context, a Alert) error {
	if s.URL == "" {
		return fmt.Errorf("alert webhook URL empty")
	}
	b, _ := json.Marshal(a)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c := s.Client
	if c == nil {
		c = http.DefaultClient
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("alert webhook status %s", resp.Status)
	}
	return nil
}

// MultiSink tries every sink and returns the first error.
type MultiSink []Sink

func (m MultiSink) Send(ctx context.Context, a Alert) error {
	var first error
	for _, s := range m {
		if s == nil {
			continue
		}
		if err := s.Send(ctx, a); err != nil && first == nil {
			first = err
		}
	}
	return first
}

type AnchorReport struct {
	Anchored bool   `json:"anchored"`
	Stale    bool   `json:"stale"`
	Seq      uint64 `json:"seq"`
	Reason   string `json:"reason,omitempty"`
}

// AnchorOnce anchors when the anchor is missing, stale, or behind the chain.
// It sends one alert through sink when anchoring fails or the anchor is stale.
func AnchorOnce(ctx context.Context, logPath, anchorPath string, key []byte, maxAge time.Duration, sink Sink, now time.Time) (AnchorReport, error) {
	if err := Verify(logPath); err != nil {
		send(ctx, sink, Alert{TS: now, Kind: "anchor_failed", Detail: err.Error()})
		return AnchorReport{}, fmt.Errorf("refusing scheduled anchor: %w", err)
	}
	events, err := ReadAll(logPath)
	if err != nil {
		return AnchorReport{}, err
	}
	var headSeq uint64
	if n := len(events); n > 0 {
		headSeq = events[n-1].Seq
	}
	stale, reason := anchorStale(anchorPath, headSeq, maxAge, now)
	if !stale {
		return AnchorReport{Seq: headSeq}, nil
	}
	if err := Anchor(logPath, anchorPath, key); err != nil {
		send(ctx, sink, Alert{TS: now, Kind: "anchor_failed", Detail: err.Error(), Seq: headSeq})
		return AnchorReport{Stale: true, Seq: headSeq, Reason: reason}, err
	}
	return AnchorReport{Anchored: true, Stale: true, Seq: headSeq, Reason: reason}, nil
}

func anchorStale(anchorPath string, headSeq uint64, maxAge time.Duration, now time.Time) (bool, string) {
	fi, err := os.Stat(anchorPath)
	if err != nil {
		return true, "anchor missing"
	}
	b, err := os.ReadFile(anchorPath)
	if err != nil {
		return true, "anchor unreadable"
	}
	var rec struct {
		Seq uint64 `json:"seq"`
	}
	if err := json.Unmarshal(b, &rec); err != nil {
		return true, "anchor malformed"
	}
	if rec.Seq < headSeq {
		return true, "anchor behind chain head"
	}
	if maxAge > 0 && now.Sub(fi.ModTime()) > maxAge {
		return true, "anchor older than max age"
	}
	return false, ""
}

// StaleAnchor reports whether the anchor on disk is stale. It writes nothing.
func StaleAnchor(anchorPath string, headSeq uint64, maxAge time.Duration, now time.Time) (bool, string) {
	return anchorStale(anchorPath, headSeq, maxAge, now)
}

func send(ctx context.Context, sink Sink, a Alert) {
	if sink != nil {
		_ = sink.Send(ctx, a)
	}
}
