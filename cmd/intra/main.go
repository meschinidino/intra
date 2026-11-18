package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"intra/internal/audit"
	"intra/internal/broker"
	"intra/internal/ops"
	"intra/internal/drill"
	"intra/internal/onboard"
	"intra/internal/policy"
	"intra/internal/providers"
	"intra/internal/simulator"
	"intra/internal/version"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func fatal(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "version":
		fmt.Println("intra", version.Version)
	case "init-demo":
		fatal(initDemo(os.Args[2:]))
	case "compile":
		fatal(cmdCompile(os.Args[2:]))
	case "request":
		fatal(cmdRequest(os.Args[2:]))
	case "approve":
		fatal(cmdApprove(os.Args[2:]))
	case "revoke":
		fatal(cmdRevoke(os.Args[2:]))
	case "grants":
		fatal(cmdGrants(os.Args[2:]))
	case "status":
		fatal(cmdStatus(os.Args[2:]))
	case "anchor":
		fatal(cmdAnchor(os.Args[2:], false))
	case "verify-anchor":
		fatal(cmdAnchor(os.Args[2:], true))
	case "replay":
		fatal(cmdReplay(os.Args[2:]))
	case "draft":
		fatal(cmdDraft(os.Args[2:]))
	case "simulate":
		fatal(cmdSimulate(os.Args[2:]))
	case "drills":
		fatal(cmdDrills(os.Args[2:]))
	case "import":
		fatal(cmdImport(os.Args[2:]))
	case "providers":
		fatal(cmdProviders(os.Args[2:]))
	case "latency":
		fatal(cmdLatency(os.Args[2:]))
	case "issue-identity":
		fatal(cmdIssueIdentity(os.Args[2:]))
	case "onboard":
		fatal(cmdOnboardApply(os.Args[2:]))
	case "backup":
		fatal(cmdBackup(os.Args[2:]))
	case "restore":
		fatal(cmdRestore(os.Args[2:]))
	case "rebind-approver":
		fatal(cmdRebind(os.Args[2:]))
	case "anchor-once":
		fatal(cmdAnchorOnce(os.Args[2:]))
	case "release-checksums":
		fatal(cmdReleaseChecksums(os.Args[2:], false))
	case "verify-release":
		fatal(cmdReleaseChecksums(os.Args[2:], true))
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("intra <version|init-demo|compile|request|approve|revoke|grants|status|anchor|verify-anchor|anchor-once|replay|draft|simulate|drills|import|providers|latency|issue-identity|onboard|backup|restore|rebind-approver|release-checksums|verify-release>")
}

func brokerBase() string {
	if v := os.Getenv("INTRA_BROKER_URL"); v != "" {
		return v
	}
	if _, err := os.Stat(filepath.Join(dataDir(), "certs", "guard.crt")); err == nil {
		return "https://127.0.0.1:18081"
	}
	return "http://127.0.0.1:18081"
}

func postJSON(url string, in, out any) error {
	b, _ := json.Marshal(in)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if tok := adminToken(); tok != "" {
		req.Header.Set("X-Intra-Admin", tok)
	}
	resp, err := brokerClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("%s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	if out != nil {
		return json.Unmarshal(body, out)
	}
	return nil
}

func cmdCompile(args []string) error {
	fs := flag.NewFlagSet("compile", flag.ExitOnError)
	path := fs.String("policy", ".intra/policy.json", "policy")
	out := fs.String("out", ".intra/compiled.json", "output")
	_ = fs.Parse(args)
	p, err := policy.Load(*path)
	if err != nil {
		return err
	}
	c, err := p.Compile()
	if err != nil {
		return err
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	if err := os.MkdirAll(filepath.Dir(*out), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(*out, append(b, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Println("hash", c.Hash)
	return nil
}

func split(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func cmdRequest(args []string) error {
	fs := flag.NewFlagSet("request", flag.ExitOnError)
	key := fs.String("request-key", "", "unique request key")
	actor := fs.String("actor", "owner", "actor")
	device := fs.String("device", "mac", "device")
	services := fs.String("services", "sensitive-db", "comma services")
	actions := fs.String("actions", "read", "comma actions")
	reason := fs.String("reason", "", "reason")
	incident := fs.String("incident", "", "incident id")
	tmpl := fs.String("template", "", "template id")
	ttl := fs.Int("ttl", 1800, "ttl seconds")
	_ = fs.Parse(args)
	if *key == "" {
		*key = fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	var acts []policy.Action
	for _, a := range split(*actions) {
		acts = append(acts, policy.Action(a))
	}
	var out map[string]any
	err := postJSON(brokerBase()+"/grants", broker.Request{
		RequestKey: *key, Actor: *actor, Device: *device, Services: split(*services),
		Actions: acts, Reason: *reason, IncidentID: *incident, TemplateID: *tmpl, TTLSeconds: *ttl,
	}, &out)
	if err != nil {
		return err
	}
	printJSON(out)
	return nil
}

func cmdApprove(args []string) error {
	fs := flag.NewFlagSet("approve", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	id := fs.String("id", "", "grant id")
	approver := fs.String("approver", "owner", "approver")
	helper := fs.String("touchid-helper", os.Getenv("INTRA_TOUCHID_HELPER"), "touch id helper binary")
	dev := fs.Bool("dev-approve", false, "DEMO ONLY: mark Touch ID passed without helper")
	_ = fs.Parse(args)
	if *id == "" {
		return fmt.Errorf("--id required")
	}
	touch := false
	if *helper != "" {
		if err := exec.Command(*helper).Run(); err != nil {
			return fmt.Errorf("touch id: %w", err)
		}
		touch = true
	} else if *dev {
		touch = true
	}
	key, err := os.ReadFile(filepath.Join(*data, "approval.key"))
	if err != nil {
		return err
	}
	var grant broker.Grant
	if code, err := getJSON(brokerBase()+"/grants/"+*id, &grant); err != nil || code/100 != 2 {
		if err != nil {
			return err
		}
		return fmt.Errorf("fetch grant status %d", code)
	}
	sig, err := broker.ApprovalSignature(key, grant, *approver)
	if err != nil {
		return err
	}
	var out broker.Grant
	err = postJSON(brokerBase()+"/grants/"+*id+"/approve", map[string]any{
		"approver": *approver, "signature": sig, "touch_id": touch,
	}, &out)
	if err != nil {
		return err
	}
	printJSON(out.Public())
	return nil
}

func cmdRevoke(args []string) error {
	fs := flag.NewFlagSet("revoke", flag.ExitOnError)
	id := fs.String("id", "", "grant id")
	by := fs.String("by", "owner", "by")
	_ = fs.Parse(args)
	var out broker.Grant
	if err := postJSON(brokerBase()+"/grants/"+*id+"/revoke", map[string]any{"by": *by}, &out); err != nil {
		return err
	}
	printJSON(out.Public())
	return nil
}

func cmdGrants(args []string) error {
	var out []broker.Grant
	code, err := getJSON(brokerBase()+"/grants", &out)
	if err != nil {
		return err
	}
	if code/100 != 2 {
		return fmt.Errorf("broker status %d", code)
	}
	printJSON(out)
	return nil
}

func getOK(url string) (int, []byte) {
	c := brokerClient()
	c.Timeout = 2 * time.Second
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	if tok := adminToken(); tok != "" {
		req.Header.Set("X-Intra-Admin", tok)
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, []byte(err.Error())
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, b
}

func cmdStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	_ = fs.Parse(args)
	out := map[string]any{}
	scheme := "http"
	if _, err := os.Stat(filepath.Join(*data, "certs", "guard.crt")); err == nil {
		scheme = "https"
	}
	auditAddr := env("INTRA_AUDIT_ADDR", "127.0.0.1:18082")
	guardURL := "http://" + env("INTRA_GUARD_HTTP_ADDR", "127.0.0.1:18080") + "/health"
	if scheme == "https" {
		guardURL = "https://127.0.0.1:18443/health"
	}
	for name, url := range map[string]string{
		"broker": brokerBase() + "/health",
		"audit":  scheme + "://" + auditAddr + "/health",
		"guard":  guardURL,
	} {
		code, _ := getOK(url)
		out[name] = code
	}
	if code, b := getOK(scheme + "://" + auditAddr + "/head"); code == 200 {
		var head any
		_ = json.Unmarshal(b, &head)
		out["audit_head"] = head
	}
	if code, b := getOK(brokerBase() + "/grants"); code == 200 {
		var grants []broker.Grant
		_ = json.Unmarshal(b, &grants)
		active := 0
		for _, g := range grants {
			if g.State == broker.StateActive {
				active++
			}
		}
		out["grants_total"] = len(grants)
		out["grants_active"] = active
	}
	anchor := filepath.Join(*data, "anchor.json")
	if fi, err := os.Stat(anchor); err == nil {
		out["anchor_age_seconds"] = int(time.Since(fi.ModTime()).Seconds())
	} else {
		out["anchor_age_seconds"] = -1
	}
	// Operator summary: partial revoke, stale anchor, and late kills.
	var grants []broker.Grant
	if code, b := getOK(brokerBase() + "/grants"); code == 200 {
		_ = json.Unmarshal(b, &grants)
	}
	var headSeq uint64
	if events, err := loadAudit(*data); err == nil && len(events) > 0 {
		headSeq = events[len(events)-1].Seq
	}
	maxAge := 24 * time.Hour
	if v := os.Getenv("INTRA_ANCHOR_MAX_AGE_SEC"); v != "" {
		var secs int
		if _, err := fmt.Sscan(v, &secs); err == nil && secs > 0 {
			maxAge = time.Duration(secs) * time.Second
		}
	}
	sum := ops.Summarize(grants, anchor, headSeq, maxAge, time.Now())
	out["grants_by_state"] = sum.GrantsByState
	out["partial_revokes"] = sum.PartialRevokes
	out["late_kills"] = sum.LateKills
	out["anchor_stale"] = sum.AnchorStale
	out["anchor_reason"] = sum.AnchorReason
	printJSON(out)
	return nil
}

func cmdBackup(args []string) error {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	out := fs.String("out", "", "backup zip path")
	_ = fs.Parse(args)
	if *out == "" {
		return fmt.Errorf("--out required")
	}
	if err := ops.CreateBackup(*data, *out); err != nil {
		return err
	}
	fmt.Println("backup written to", *out)
	return nil
}

func cmdRestore(args []string) error {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	from := fs.String("from", "", "backup zip path")
	force := fs.Bool("force", false, "replace existing grants and audit state")
	_ = fs.Parse(args)
	if *from == "" {
		return fmt.Errorf("--from required; stop intrad before restore")
	}
	if !*force {
		for _, name := range []string{"grants.json", "audit.jsonl"} {
			if _, err := os.Stat(filepath.Join(*data, name)); err == nil {
				return fmt.Errorf("%s exists in %s; stop intrad and pass --force to restore over it", name, *data)
			}
		}
	}
	key, _ := readAnchorKey(*data)
	if err := ops.RestoreBackup(*from, *data, key); err != nil {
		return err
	}
	fmt.Println("restore OK; restart intrad")
	return nil
}

func cmdRebind(args []string) error {
	fs := flag.NewFlagSet("rebind-approver", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	confirm := fs.Bool("confirm-old-revoked", false, "confirm the old approver device is revoked or declared lost")
	_ = fs.Parse(args)
	if !*confirm {
		return fmt.Errorf("--confirm-old-revoked required: revoke or declare the old device lost first")
	}
	if err := ops.RebindApprover(*data, time.Now()); err != nil {
		return err
	}
	fmt.Println("approver rebound; restart intrad to load the new key")
	return nil
}

func cmdAnchorOnce(args []string) error {
	fs := flag.NewFlagSet("anchor-once", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	maxAgeSec := fs.Int("max-age-sec", 86400, "max anchor age in seconds")
	_ = fs.Parse(args)
	key, err := readAnchorKey(*data)
	if err != nil {
		return err
	}
	sink := audit.MultiSink{audit.FileSink{Path: filepath.Join(*data, "alerts.jsonl")}}
	if hook := os.Getenv("INTRA_ALERT_WEBHOOK"); hook != "" {
		sink = append(sink, audit.WebhookSink{URL: hook})
	}
	rep, err := audit.AnchorOnce(context.Background(), filepath.Join(*data, "audit.jsonl"), filepath.Join(*data, "anchor.json"), key, time.Duration(*maxAgeSec)*time.Second, sink, time.Now())
	if err != nil {
		return err
	}
	printJSON(rep)
	return nil
}

func cmdReleaseChecksums(args []string, verify bool) error {
	fs := flag.NewFlagSet("release-checksums", flag.ExitOnError)
	dir := fs.String("dir", ".", "artifact directory")
	files := fs.String("files", "", "comma file names (generate mode)")
	out := fs.String("out", "", "checksums path (generate mode)")
	sums := fs.String("checksums", "", "checksums path (verify mode)")
	_ = fs.Parse(args)
	if verify {
		if *sums == "" {
			return fmt.Errorf("--checksums required")
		}
		bad, err := ops.VerifyChecksums(*sums, *dir)
		if err != nil {
			return err
		}
		if len(bad) > 0 {
			return fmt.Errorf("release verification failed for: %s", strings.Join(bad, ", "))
		}
		fmt.Println("release artifacts OK")
		return nil
	}
	if *out == "" || *files == "" {
		return fmt.Errorf("--files and --out required")
	}
	if err := ops.WriteChecksums(*dir, split(*files), *out); err != nil {
		return err
	}
	fmt.Println("checksums written to", *out)
	return nil
}

func cmdAnchor(args []string, verify bool) error {
	fs := flag.NewFlagSet("anchor", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	_ = fs.Parse(args)
	key, err := readAnchorKey(*data)
	if err != nil {
		return err
	}
	logPath := filepath.Join(*data, "audit.jsonl")
	anchorPath := filepath.Join(*data, "anchor.json")
	if verify {
		if err := audit.VerifyAnchor(logPath, anchorPath, key); err != nil {
			return err
		}
		fmt.Println("anchor OK")
		return nil
	}
	return audit.Anchor(logPath, anchorPath, key)
}

func loadAudit(data string) ([]audit.Event, error) {
	return audit.ReadAll(filepath.Join(data, "audit.jsonl"))
}

func cmdReplay(args []string) error {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	path := fs.String("policy", ".intra/policy.json", "candidate policy")
	_ = fs.Parse(args)
	events, err := loadAudit(*data)
	if err != nil {
		return err
	}
	p, err := policy.Load(*path)
	if err != nil {
		return err
	}
	printJSON(simulator.Replay(events, p))
	return nil
}

func cmdDraft(args []string) error {
	fs := flag.NewFlagSet("draft", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	_ = fs.Parse(args)
	events, err := loadAudit(*data)
	if err != nil {
		return err
	}
	printJSON(simulator.DraftRules(events))
	return nil
}

func cmdSimulate(args []string) error {
	fs := flag.NewFlagSet("simulate", flag.ExitOnError)
	path := fs.String("policy", ".intra/policy.json", "policy")
	_ = fs.Parse(args)
	p, err := policy.Load(*path)
	if err != nil {
		return err
	}
	printJSON(simulator.ReachPaths(p))
	return nil
}

func cmdDrills(args []string) error {
	fs := flag.NewFlagSet("drills", flag.ExitOnError)
	config := fs.String("config", "", "targets JSON file")
	_ = fs.Parse(args)
	if *config == "" {
		return fmt.Errorf("--config required")
	}
	b, err := os.ReadFile(*config)
	if err != nil {
		return err
	}
	var targets []drill.Target
	if err := json.Unmarshal(b, &targets); err != nil {
		return err
	}
	printJSON(drill.Run(context.Background(), targets, "CANARY-SECRET-123"))
	return nil
}

func cmdImport(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: intra import <wireguard|tailscale> <file>")
	}
	var out []onboard.ImportedPath
	var err error
	switch args[0] {
	case "wireguard":
		out, err = onboard.ImportWireGuard(args[1])
	case "tailscale":
		out, err = onboard.ImportTailscale(args[1])
	default:
		return fmt.Errorf("unknown import %s", args[0])
	}
	if err != nil {
		return err
	}
	printJSON(out)
	return nil
}

func cmdProviders(args []string) error {
	// Local simulator exercise only. Real tenant proof needs a new budget decision.
	entra := providers.NewSimulator("entra-sim", 0)
	google := providers.NewSimulator("google-sim", 0)
	out := map[string]any{}
	for _, a := range []providers.Adapter{entra, google} {
		g, err := a.Issue(context.Background(), "vendor@example.com", "read", "owner", time.Hour)
		if err != nil {
			return err
		}
		lat, err := a.KillLatency(context.Background(), g.ID)
		if err != nil {
			return err
		}
		out[a.Name()] = map[string]any{"grant": g.ID, "kill_ms": lat.Milliseconds(), "simulated": true}
	}
	printJSON(out)
	return nil
}

func cmdLatency(args []string) error {
	fs := flag.NewFlagSet("latency", flag.ExitOnError)
	direct := fs.String("direct", "", "direct URL")
	guarded := fs.String("guarded", "", "guarded URL")
	n := fs.Int("n", 50, "requests")
	actor := fs.String("actor", "owner", "actor header for guarded")
	caFile := fs.String("cacert", "", "CA file for guarded TLS")
	certFile := fs.String("cert", "", "client cert for guarded TLS")
	keyFile := fs.String("key", "", "client key for guarded TLS")
	_ = fs.Parse(args)
	client := http.DefaultClient
	if *caFile != "" {
		caPEM, err := os.ReadFile(*caFile)
		if err != nil {
			return err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caPEM)
		cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
		if err != nil {
			return err
		}
		client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}}}
	}
	measure := func(url string, guarded bool) (float64, error) {
		var ds []float64
		for i := 0; i < *n; i++ {
			req, _ := http.NewRequest(http.MethodGet, url, nil)
			if guarded {
				req.Header.Set("X-Intra-Actor", *actor)
				req.Header.Set("X-Intra-Device", "mac")
			}
			start := time.Now()
			c := http.DefaultClient
			if guarded {
				c = client
			}
			resp, err := c.Do(req)
			if err != nil {
				return 0, err
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			ds = append(ds, float64(time.Since(start).Microseconds())/1000)
		}
		sort.Float64s(ds)
		return ds[int(0.95*float64(len(ds)-1))], nil
	}
	d, err := measure(*direct, false)
	if err != nil {
		return err
	}
	g, err := measure(*guarded, true)
	if err != nil {
		return err
	}
	pass := g-d < 50
	printJSON(map[string]any{"direct_p95_ms": d, "guarded_p95_ms": g, "added_p95_ms": g - d, "budget_ms": 50, "pass": pass})
	if !pass {
		return fmt.Errorf("added p95 latency %.3f ms exceeds 50 ms budget", g-d)
	}
	return nil
}
