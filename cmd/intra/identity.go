package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"intra/internal/certs"
	"intra/internal/onboard"
	"intra/internal/policy"
)

type runManifest struct {
	Actor      string    `json:"actor"`
	RunID      string    `json:"run_id"`
	Device     string    `json:"device"`
	ExpiresAt  time.Time `json:"expires_at"`
	CertName   string    `json:"cert_name"`
	CertSHA256 string    `json:"cert_sha256"`
	Signature  string    `json:"signature"`
}

// cmdIssueIdentity creates a per-run client certificate and signed manifest.
// This is the Stage 7 attestation root for the local build: only the holder
// of the local CA plus approval key can mint a run identity.
func cmdIssueIdentity(args []string) error {
	fs := flag.NewFlagSet("issue-identity", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	actor := fs.String("actor", "robot", "actor")
	run := fs.String("run", "", "run id")
	ttl := fs.Int("ttl-hours", 24, "max identity age")
	_ = fs.Parse(args)
	if *run == "" {
		return fmt.Errorf("--run required")
	}
	if *ttl < 1 || *ttl > 24 {
		return fmt.Errorf("--ttl-hours must be 1..24")
	}
	certDir := filepath.Join(*data, "certs")
	ca, err := certs.LoadCA(filepath.Join(certDir, "ca.crt"), filepath.Join(certDir, "ca.key"))
	if err != nil {
		return err
	}
	name := strings.NewReplacer("/", "-", " ", "-").Replace(*actor + "-" + *run)
	if err := ca.IssueWithTTL(certDir, name, *actor+"/"+*run, false, time.Duration(*ttl)*time.Hour); err != nil {
		return err
	}
	key, err := os.ReadFile(filepath.Join(*data, "approval.key"))
	if err != nil {
		return err
	}
	certPEM, err := os.ReadFile(filepath.Join(certDir, name+".crt"))
	if err != nil {
		return err
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("bad issued cert")
	}
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		return err
	}
	fp := sha256.Sum256(block.Bytes)
	m := runManifest{Actor: *actor, RunID: *run, Device: *run, ExpiresAt: time.Now().Add(time.Duration(*ttl) * time.Hour).UTC(), CertName: name, CertSHA256: hex.EncodeToString(fp[:])}
	mac := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(mac, "%s:%s:%s", m.Actor, m.RunID, m.ExpiresAt.Format(time.RFC3339))
	m.Signature = hex.EncodeToString(mac.Sum(nil))
	b, _ := json.MarshalIndent(m, "", "  ")
	out := filepath.Join(certDir, name+".manifest.json")
	if err := os.WriteFile(out, append(b, '\n'), 0o600); err != nil {
		return err
	}
	printJSON(m)
	return nil
}

// cmdOnboardApply performs the first guided safe cut: it imports current
// reach, adds one selected service to policy, and compiles generated configs.
func cmdOnboardApply(args []string) error {
	fs := flag.NewFlagSet("onboard", flag.ExitOnError)
	data := fs.String("data", ".intra", "data")
	kind := fs.String("kind", "tailscale", "tailscale or wireguard")
	file := fs.String("file", "", "import file")
	name := fs.String("service", "imported", "new service name")
	upstream := fs.String("upstream", "", "service upstream URL")
	owner := fs.String("owner", "founder", "service owner")
	_ = fs.Parse(args)
	if *file == "" || *upstream == "" {
		return fmt.Errorf("--file and --upstream required")
	}
	var paths []onboard.ImportedPath
	var err error
	switch *kind {
	case "tailscale":
		paths, err = onboard.ImportTailscale(*file)
	case "wireguard":
		paths, err = onboard.ImportWireGuard(*file)
	default:
		return fmt.Errorf("unknown kind %s", *kind)
	}
	if err != nil {
		return err
	}
	polPath := filepath.Join(*data, "policy.json")
	p, err := policy.Load(polPath)
	if err != nil {
		return err
	}
	if _, exists := p.Services[*name]; exists {
		return fmt.Errorf("service %s exists", *name)
	}
	p.Services[*name] = policy.Service{Name: *name, Kind: "http", Sensitive: true, Upstream: *upstream, Owner: *owner, Actions: []policy.Action{policy.ActionRead}}
	p.Rules = append(p.Rules, policy.Rule{Actor: "owner", Service: *name, Actions: []policy.Action{policy.ActionRead}, Access: "grant"})
	if err := p.Validate(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(p, "", "  ")
	if err := os.WriteFile(polPath, append(b, '\n'), 0o644); err != nil {
		return err
	}
	c, err := p.Compile()
	if err != nil {
		return err
	}
	cb, _ := json.MarshalIndent(c, "", "  ")
	if err := os.WriteFile(filepath.Join(*data, "compiled.json"), append(cb, '\n'), 0o644); err != nil {
		return err
	}
	printJSON(map[string]any{"imported_paths": paths, "service_added": *name, "policy_hash": c.Hash, "next": "restart intrad, then run attack suite"})
	return nil
}
