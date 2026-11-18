package broker

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func certActor(r *http.Request) string {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return ""
	}
	cn := r.TLS.PeerCertificates[0].Subject.CommonName
	if i := strings.Index(cn, "/"); i > 0 {
		return cn[:i]
	}
	return cn
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/grants", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, 200, s.List())
		case http.MethodPost:
			var req Request
			if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
				writeJSON(w, 400, map[string]string{"error": "bad request"})
				return
			}
			if op := certActor(r); op != "" && op != req.Actor && op != "owner" {
				writeJSON(w, 403, map[string]string{"error": "certificate actor cannot request for this actor"})
				return
			}
			g, dup, err := s.RequestGrant(r.Context(), req)
			if err != nil {
				writeJSON(w, 403, map[string]string{"error": err.Error()})
				return
			}
			code := 201
			if dup {
				code = 200
			}
			writeJSON(w, code, map[string]any{"grant": g.Public(), "token": g.Token(), "duplicate": dup})
		default:
			w.WriteHeader(405)
		}
	})
	mux.HandleFunc("/grants/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/grants/"), "/")
		if len(parts) == 0 {
			w.WriteHeader(404)
			return
		}
		id := parts[0]
		if len(parts) == 1 && r.Method == http.MethodGet {
			g, ok := s.Get(id)
			if !ok {
				w.WriteHeader(404)
				return
			}
			writeJSON(w, 200, g)
			return
		}
		if len(parts) != 2 || r.Method != http.MethodPost {
			w.WriteHeader(404)
			return
		}
		var body struct {
			Approver  string `json:"approver"`
			Signature string `json:"signature"`
			TouchID   bool   `json:"touch_id"`
			By        string `json:"by"`
		}
		_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body)
		if op := certActor(r); op != "" {
			body.Approver = op
			body.By = op
		}
		switch parts[1] {
		case "approve":
			if err := s.Approve(r.Context(), id, body.Approver, body.Signature, body.TouchID); err != nil {
				writeJSON(w, 403, map[string]string{"error": err.Error()})
				return
			}
		case "revoke":
			if err := s.Revoke(r.Context(), id, first(body.By, body.Approver, "owner")); err != nil {
				writeJSON(w, 409, map[string]string{"error": err.Error()})
				return
			}
		default:
			w.WriteHeader(404)
			return
		}
		g, _ := s.Get(id)
		writeJSON(w, 200, g)
	})
	mux.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		ok := s.Validate(q.Get("actor"), q.Get("device"), q.Get("service"), actionOf(q.Get("action")), strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if !ok {
			w.WriteHeader(403)
			return
		}
		w.WriteHeader(204)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" && s.AdminToken != "" && r.Header.Get("X-Intra-Admin") != s.AdminToken {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func first(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}
