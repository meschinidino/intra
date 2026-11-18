package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"intra/internal/audit"
	"intra/internal/broker"
	"intra/internal/guard"
	"intra/internal/pggateway"
	"intra/internal/policy"
	"intra/internal/providers"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func clientAddr(addr string) string {
	return strings.Replace(addr, "0.0.0.0", "127.0.0.1", 1)
}

func readTrim(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func serveControl(addr string, h http.Handler, data string) error {
	cert := filepath.Join(data, "certs", "guard.crt")
	key := filepath.Join(data, "certs", "guard.key")
	caPath := filepath.Join(data, "certs", "ca.crt")
	if _, err := os.Stat(cert); err != nil {
		return fmt.Errorf("control TLS cert required: %w", err)
	}
	srv, err := tlsServer(addr, h, cert, key, caPath)
	if err != nil {
		return err
	}
	return srv.ListenAndServeTLS("", "")
}

func controlClient(data string) *http.Client {
	certPath := filepath.Join(data, "certs", "control.crt")
	keyPath := filepath.Join(data, "certs", "control.key")
	if _, err := os.Stat(certPath); err != nil {
		certPath = filepath.Join(data, "certs", "owner.crt")
		keyPath = filepath.Join(data, "certs", "owner.key")
	}
	caPath := filepath.Join(data, "certs", "ca.crt")
	if _, err := os.Stat(certPath); err != nil {
		return http.DefaultClient
	}
	c, err := tlsClient(certPath, keyPath, caPath)
	if err != nil {
		log.Fatalf("control client: %v", err)
	}
	return c
}

func controlScheme(data string) string {
	if _, err := os.Stat(filepath.Join(data, "certs", "guard.crt")); err == nil {
		return "https"
	}
	return "http"
}

func mustKey(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read key %s: %v (run intra init-demo first)", path, err)
	}
	return b
}

func main() {
	mode := "all"
	if len(os.Args) > 1 {
		mode = os.Args[1]
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	}
	data := flag.String("data", env("INTRA_DATA", ".intra"), "data dir")
	policyPath := flag.String("policy", "", "policy path")
	flag.Parse()
	if *policyPath == "" {
		*policyPath = filepath.Join(*data, "policy.json")
	}
	pol, err := policy.Load(*policyPath)
	if err != nil {
		log.Fatalf("policy: %v", err)
	}
	if mode != "providers" && readTrim(filepath.Join(*data, "admin.token")) == "" {
		log.Fatal("admin token required: run intra init-demo")
	}
	ctx := context.Background()
	errs := make(chan error, 16)
	run := func(name string, fn func() error) {
		go func() {
			log.Printf("%s starting", name)
			errs <- fmt.Errorf("%s: %w", name, fn())
		}()
	}

	var auditClient *audit.Client
	if mode == "all" || mode == "audit" {
		w, err := audit.OpenWriter(filepath.Join(*data, "audit.jsonl"))
		if err != nil {
			log.Fatal(err)
		}
		defer w.Close()
		srv := &audit.Server{W: w, AdminToken: readTrim(filepath.Join(*data, "admin.token"))}
		run("audit", func() error { return serveControl(env("INTRA_AUDIT_ADDR", "127.0.0.1:18082"), srv.Handler(), *data) })
		time.Sleep(100 * time.Millisecond)
	}
	auditClient = &audit.Client{Base: controlScheme(*data) + "://" + clientAddr(env("INTRA_AUDIT_ADDR", "127.0.0.1:18082")), HTTP: controlClient(*data), AdminToken: readTrim(filepath.Join(*data, "admin.token"))}

	var b *broker.Server
	var gw *pggateway.Gateway
	if mode == "all" || mode == "broker" {
		store, err := broker.OpenStore(filepath.Join(*data, "grants.json"))
		if err != nil {
			log.Fatal(err)
		}
		gw = &pggateway.Gateway{Audit: auditClient}
		if os.Getenv("INTRA_PG_ADMIN_URL") != "" {
			tc, err := serverTLSConfig(filepath.Join(*data, "certs", "guard.crt"), filepath.Join(*data, "certs", "guard.key"), filepath.Join(*data, "certs", "ca.crt"))
			if err != nil {
				log.Fatalf("pg gateway TLS: %v", err)
			}
			gw.TLS = tc
		}
		killer := &pggateway.PGKiller{Gateway: gw, Objects: pol.Services["sensitive-db"].Objects}
		b = broker.New(pol, store, auditClient, killer, mustKey(filepath.Join(*data, "approval.key")))
		b.AdminToken = readTrim(filepath.Join(*data, "admin.token"))
		if helper := os.Getenv("INTRA_TOUCHID_HELPER"); helper != "" {
			b.LocalTouch = func() bool { return exec.Command(helper).Run() == nil }
		}
		gw.Broker = b

		if adminURL := os.Getenv("INTRA_PG_ADMIN_URL"); adminURL != "" {
			pool, err := pggateway.ConnectAdmin(ctx, adminURL)
			if err != nil {
				log.Fatalf("pg admin: %v", err)
			}
			killer.Admin = pool
			base, err := pggateway.ParseBase(env("INTRA_PG_BASE_URL", adminURL))
			if err != nil {
				log.Fatalf("pg base: %v", err)
			}
			gw.Base = base
			if err := b.RestorePostgresSecrets(ctx); err != nil {
				log.Fatalf("restore pg grants: %v", err)
			}
			ln, err := net.Listen("tcp", env("INTRA_PG_ADDR", "127.0.0.1:15432"))
			if err != nil {
				log.Fatal(err)
			}
			run("pggateway", func() error { return gw.Serve(ln) })
		}
		go func() {
			t := time.NewTicker(200 * time.Millisecond)
			for range t.C {
				b.SweepExpires(context.Background())
			}
		}()
		run("broker", func() error { return serveControl(env("INTRA_BROKER_ADDR", "127.0.0.1:18081"), b.Handler(), *data) })
		time.Sleep(100 * time.Millisecond)
	}

	if mode == "all" || mode == "guard" {
		if b == nil {
			// Standalone guard still needs a broker view. Open a fresh store
			// only when broker runs elsewhere in another process with shared
			// state over HTTP in compose. For local all mode b is set.
			log.Fatal("guard mode requires broker in same process for this build; use all")
		}
		g, err := guard.New(pol, b, auditClient)
		if err != nil {
			log.Fatal(err)
		}
		g.AllowHeaderID = os.Getenv("INTRA_ALLOW_HEADER_ID") == "1"
		if b, err := os.ReadFile(filepath.Join(*data, "viewer.token")); err == nil {
			g.LabViewerToken = strings.TrimSpace(string(b))
		}
		if os.Getenv("INTRA_VERIFY_MANIFESTS") == "1" {
			g.ManifestDir = filepath.Join(*data, "certs")
			g.ManifestKey = mustKey(filepath.Join(*data, "approval.key"))
		}
		run("guard", func() error { return serveGuard(g, *data) })
		eg := &guard.Egress{Policy: pol, Validator: b, Audit: auditClient}
		run("egress", func() error { return serveControl(env("INTRA_EGRESS_ADDR", "127.0.0.1:18083"), eg.Handler(), *data) })
	}

	if mode == "all" || mode == "providers" {
		entra := providers.NewSimulator("entra-sim", 50*time.Millisecond)
		google := providers.NewSimulator("google-sim", 50*time.Millisecond)
		run("provider-entra-sim", func() error {
			return http.ListenAndServe(env("INTRA_ENTRA_SIM_ADDR", "127.0.0.1:18084"), entra.Handler())
		})
		run("provider-google-sim", func() error {
			return http.ListenAndServe(env("INTRA_GOOGLE_SIM_ADDR", "127.0.0.1:18085"), google.Handler())
		})
	}

	log.Fatal(<-errs)
}

func serveGuard(g *guard.Guard, data string) error {
	addr := env("INTRA_GUARD_ADDR", "0.0.0.0:18443")
	cert := filepath.Join(data, "certs", "guard.crt")
	key := filepath.Join(data, "certs", "guard.key")
	caPath := filepath.Join(data, "certs", "ca.crt")
	if _, err := os.Stat(cert); err == nil {
		srv, err := tlsServer(addr, g.Handler(), cert, key, caPath)
		if err != nil {
			return err
		}
		return srv.ListenAndServeTLS("", "")
	}
	// Plain HTTP exists only for unit/dev wiring where certs were not made.
	return http.ListenAndServe(env("INTRA_GUARD_HTTP_ADDR", "127.0.0.1:18080"), g.Handler())
}
