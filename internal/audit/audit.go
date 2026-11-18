package audit

import (
	"bufio"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Event struct {
	TS        time.Time `json:"ts"`
	Actor     string    `json:"actor"`
	Device    string    `json:"device"`
	Service   string    `json:"service"`
	Action    string    `json:"action"`
	Rule      string    `json:"rule"`
	Result    string    `json:"result"`
	GrantID   string    `json:"grant_id,omitempty"`
	QueryHash string    `json:"query_hash,omitempty"`
	Seq       uint64    `json:"seq"`
	Prev      string    `json:"prev"`
	Hash      string    `json:"hash"`
}

type Writer struct {
	mu     sync.Mutex
	path   string
	seq    uint64
	prev   string
	f      *os.File
	lock   *os.File
	broken error
}

func OpenWriter(path string) (*Writer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = lock.Close()
		return nil, fmt.Errorf("audit writer already active: %w", err)
	}
	if err := Verify(path); err != nil {
		_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		_ = lock.Close()
		return nil, fmt.Errorf("audit chain: %w", err)
	}
	w := &Writer{path: path, prev: "GENESIS", lock: lock}
	events, err := ReadAll(path)
	if err != nil {
		_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		_ = lock.Close()
		return nil, err
	}
	if n := len(events); n > 0 {
		w.seq = events[n-1].Seq
		w.prev = events[n-1].Hash
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		_ = lock.Close()
		return nil, err
	}
	w.f = f
	return w, nil
}

func hashEvent(e Event) string {
	c := e
	c.Hash = ""
	b, _ := json.Marshal(c)
	h := sha256.Sum256(append([]byte(e.Prev), b...))
	return hex.EncodeToString(h[:])
}

func (w *Writer) Append(ctx context.Context, e Event) (Event, error) {
	if err := ctx.Err(); err != nil {
		return Event{}, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.broken != nil {
		return Event{}, w.broken
	}
	if e.TS.IsZero() {
		e.TS = time.Now().UTC()
	}
	w.seq++
	e.Seq = w.seq
	e.Prev = w.prev
	e.Hash = hashEvent(e)
	b, err := json.Marshal(e)
	if err != nil {
		w.seq--
		return Event{}, err
	}
	if _, err := w.f.Write(append(b, '\n')); err != nil {
		w.broken = fmt.Errorf("audit writer broken: %w", err)
		return Event{}, w.broken
	}
	if err := w.f.Sync(); err != nil {
		w.broken = fmt.Errorf("audit writer broken: %w", err)
		return Event{}, w.broken
	}
	w.prev = e.Hash
	return e, nil
}

func (w *Writer) Head() (uint64, string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.seq, w.prev
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	if w.lock != nil {
		_ = syscall.Flock(int(w.lock.Fd()), syscall.LOCK_UN)
		_ = w.lock.Close()
		w.lock = nil
	}
	return err
}

func ReadAll(path string) ([]Event, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return nil, fmt.Errorf("audit line: %w", err)
		}
		out = append(out, e)
	}
	return out, sc.Err()
}

func Verify(path string) error {
	events, err := ReadAll(path)
	if err != nil {
		return err
	}
	prev := "GENESIS"
	for i, e := range events {
		if e.Seq != uint64(i+1) {
			return fmt.Errorf("audit seq gap at %d", i+1)
		}
		if e.Prev != prev {
			return fmt.Errorf("audit prev mismatch at seq %d", e.Seq)
		}
		if hashEvent(e) != e.Hash {
			return fmt.Errorf("audit hash mismatch at seq %d", e.Seq)
		}
		prev = e.Hash
	}
	return nil
}

// Anchor writes an HMAC-signed chain head outside the audit store.
func Anchor(path, anchorPath string, key []byte) error {
	if err := Verify(path); err != nil {
		return fmt.Errorf("refusing to anchor unverified chain: %w", err)
	}
	events, err := ReadAll(path)
	if err != nil {
		return err
	}
	head := "GENESIS"
	var seq uint64
	if n := len(events); n > 0 {
		head = events[n-1].Hash
		seq = events[n-1].Seq
	}
	m := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(m, "%d:%s", seq, head)
	rec := map[string]any{"seq": seq, "head": head, "sig": hex.EncodeToString(m.Sum(nil)), "ts": time.Now().UTC()}
	b, _ := json.Marshal(rec)
	return os.WriteFile(anchorPath, append(b, '\n'), 0o600)
}

func VerifyAnchor(path, anchorPath string, key []byte) error {
	b, err := os.ReadFile(anchorPath)
	if err != nil {
		return err
	}
	var rec struct {
		Seq  uint64 `json:"seq"`
		Head string `json:"head"`
		Sig  string `json:"sig"`
	}
	if err := json.Unmarshal(b, &rec); err != nil {
		return err
	}
	m := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(m, "%d:%s", rec.Seq, rec.Head)
	if !hmac.Equal([]byte(hex.EncodeToString(m.Sum(nil))), []byte(rec.Sig)) {
		return fmt.Errorf("anchor signature mismatch")
	}
	if err := Verify(path); err != nil {
		return err
	}
	events, _ := ReadAll(path)
	if rec.Seq > 0 {
		if int(rec.Seq) > len(events) || events[rec.Seq-1].Hash != rec.Head {
			return fmt.Errorf("anchor head not in chain")
		}
	}
	return nil
}

type Server struct {
	W          *Writer
	AdminToken string
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if s.AdminToken != "" && r.Header.Get("X-Intra-Admin") != s.AdminToken {
			http.Error(w, "forbidden", 403)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method", 405)
			return
		}
		var e Event
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&e); err != nil {
			http.Error(w, "bad event", 400)
			return
		}
		out, err := s.W.Append(r.Context(), e)
		if err != nil {
			http.Error(w, "audit unavailable", 503)
			return
		}
		_ = json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("/head", func(w http.ResponseWriter, r *http.Request) {
		if s.AdminToken != "" && r.Header.Get("X-Intra-Admin") != s.AdminToken {
			http.Error(w, "forbidden", 403)
			return
		}
		seq, head := s.W.Head()
		_ = json.NewEncoder(w).Encode(map[string]any{"seq": seq, "head": head})
	})
	return mux
}

type Client struct {
	Base       string
	HTTP       *http.Client
	AdminToken string
}

func (c *Client) Append(ctx context.Context, e Event) error {
	if c.Base == "" {
		return fmt.Errorf("audit unavailable")
	}
	b, _ := json.Marshal(e)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Base+"/events", bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.AdminToken != "" {
		req.Header.Set("X-Intra-Admin", c.AdminToken)
	}
	h := c.HTTP
	if h == nil {
		h = http.DefaultClient
	}
	resp, err := h.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("audit unavailable: %s", resp.Status)
	}
	return nil
}
