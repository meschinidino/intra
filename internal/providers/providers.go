// Package providers holds local SaaS provider simulators plus the adapter
// contract a real Entra ID or Google Workspace integration must satisfy.
// The simulators never claim real-tenant proof.
package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Grant struct {
	ID        string    `json:"id"`
	Provider  string    `json:"provider"`
	Subject   string    `json:"subject"`
	Scope     string    `json:"scope"`
	ExpiresAt time.Time `json:"expires_at"`
	SessionID string    `json:"session_id"`
	Revoked   bool      `json:"revoked"`
	LastUsed  time.Time `json:"last_used"`
	Owner     string    `json:"owner"`
}

type Adapter interface {
	Name() string
	Issue(ctx context.Context, subject, scope, owner string, ttl time.Duration) (Grant, error)
	Revoke(ctx context.Context, id string) error
	KillLatency(ctx context.Context, id string) (time.Duration, error)
	Inventory(ctx context.Context) ([]Grant, error)
}

// Simulator is an in-memory provider used for local Stage 8 exercises.
type Simulator struct {
	Provider string
	KillWait time.Duration
	mu       sync.Mutex
	grants   map[string]*Grant
	seq      int
}

func NewSimulator(provider string, killWait time.Duration) *Simulator {
	return &Simulator{Provider: provider, KillWait: killWait, grants: map[string]*Grant{}}
}

func (s *Simulator) Name() string { return s.Provider }

func (s *Simulator) Issue(_ context.Context, subject, scope, owner string, ttl time.Duration) (Grant, error) {
	if ttl <= 0 || ttl > 24*time.Hour {
		return Grant{}, fmt.Errorf("ttl out of range")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	g := &Grant{
		ID: fmt.Sprintf("%s-%d", s.Provider, s.seq), Provider: s.Provider,
		Subject: subject, Scope: scope, Owner: owner,
		ExpiresAt: time.Now().Add(ttl), SessionID: fmt.Sprintf("sess-%d", s.seq),
		LastUsed: time.Now(),
	}
	s.grants[g.ID] = g
	return *g, nil
}

func (s *Simulator) Revoke(ctx context.Context, id string) error {
	if s.KillWait > 0 {
		select {
		case <-time.After(s.KillWait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	g, ok := s.grants[id]
	if !ok {
		return fmt.Errorf("unknown grant")
	}
	g.Revoked = true
	g.SessionID = ""
	return nil
}

func (s *Simulator) KillLatency(ctx context.Context, id string) (time.Duration, error) {
	start := time.Now()
	if err := s.Revoke(ctx, id); err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func (s *Simulator) Inventory(context.Context) ([]Grant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Grant, 0, len(s.grants))
	for _, g := range s.grants {
		out = append(out, *g)
	}
	return out, nil
}

func (s *Simulator) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		inv, _ := s.Inventory(r.Context())
		_ = json.NewEncoder(w).Encode(inv)
	})
	mux.HandleFunc("/grants", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Subject string `json:"subject"`
			Scope   string `json:"scope"`
			Owner   string `json:"owner"`
			TTL     int    `json:"ttl_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		g, err := s.Issue(r.Context(), body.Subject, body.Scope, body.Owner, time.Duration(body.TTL)*time.Second)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		_ = json.NewEncoder(w).Encode(g)
	})
	mux.HandleFunc("/grants/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/grants/")
		if r.Method != http.MethodDelete {
			w.WriteHeader(405)
			return
		}
		if err := s.Revoke(r.Context(), id); err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		w.WriteHeader(204)
	})
	return mux
}
