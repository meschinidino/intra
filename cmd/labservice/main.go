package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var hits []string
var mu sync.Mutex

func main() {
	mode := env("LAB_MODE", "normal")
	addr := env("LAB_ADDR", ":18090")
	token := os.Getenv("LAB_VIEWER_TOKEN")
	canary := env("LAB_CANARY", "CANARY-SECRET-123")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if mode == "attacker" {
			body := make([]byte, 0, 4096)
			buf := make([]byte, 4096)
			for {
				n, err := r.Body.Read(buf)
				body = append(body, buf[:n]...)
				if err != nil {
					break
				}
			}
			mu.Lock()
			hits = append(hits, r.URL.String()+" "+string(body))
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"recorded": true})
			return
		}
		// Logs and metrics require a viewer token. Write and admin always deny.
		if mode == "logs" || mode == "metrics" {
			if strings.Contains(r.URL.Path, "admin") || (r.Method != http.MethodGet && r.Method != http.MethodPost) {
				http.Error(w, "read only", 403)
				return
			}
			if token == "" || r.Header.Get("Authorization") != "Bearer "+token {
				http.Error(w, "viewer token required", 401)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"mode": mode, "entry": "fake-" + mode + " line"})
			return
		}
		if mode == "sensitive-http" {
			// The guard is the only intended caller in compose. Direct calls
			// still need the viewer token, which only the guard config knows.
			if token == "" || r.Header.Get("X-Lab-Viewer") != token {
				http.Error(w, "direct access denied", 403)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"secret": canary})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"service": "normal", "ok": true})
	})
	if mode == "attacker" {
		mux.HandleFunc("/hits", func(w http.ResponseWriter, _ *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			_ = json.NewEncoder(w).Encode(hits)
		})
	}
	log.Printf("labservice mode=%s addr=%s", mode, addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
