package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"intra/internal/certs"
	"intra/internal/policy"
)

func randBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func demoPolicy(normalUp, logsUp, metricsUp string) *policy.Policy {
	return &policy.Policy{
		Version: 1,
		Actors: map[string]policy.Actor{
			"owner": {Name: "owner", Kind: "human", Owner: "founder"},
			"robot": {Name: "robot", Kind: "robot", Owner: "founder"},
			"drill": {Name: "drill", Kind: "drill", Owner: "founder"},
		},
		Services: map[string]policy.Service{
			"normal":       {Name: "normal", Kind: "http", Upstream: normalUp, Actions: []policy.Action{policy.ActionRead, policy.ActionWrite, policy.ActionAdmin, policy.ActionExport}, Owner: "founder"},
			"sensitive-db": {Name: "sensitive-db", Kind: "postgres", Sensitive: true, Actions: []policy.Action{policy.ActionRead, policy.ActionWrite, policy.ActionAdmin, policy.ActionExport}, Owner: "founder", Objects: []string{"secrets"}},
			"logs":         {Name: "logs", Kind: "logs", Sensitive: true, Upstream: logsUp, Actions: []policy.Action{policy.ActionRead, policy.ActionWrite, policy.ActionAdmin}, Owner: "founder"},
			"metrics":      {Name: "metrics", Kind: "metrics", Sensitive: true, Upstream: metricsUp, Actions: []policy.Action{policy.ActionRead, policy.ActionWrite, policy.ActionAdmin}, Owner: "founder"},
			"internet":     {Name: "internet", Kind: "egress", Sensitive: true, Actions: []policy.Action{policy.ActionEgress}, Owner: "founder"},
		},
		Rules: []policy.Rule{
			{Actor: "owner", Service: "normal", Actions: []policy.Action{policy.ActionRead}, Access: "static"},
			{Actor: "owner", Service: "sensitive-db", Actions: []policy.Action{policy.ActionRead}, Access: "grant", MaxTTLSeconds: 8 * 3600},
			{Actor: "owner", Service: "logs", Actions: []policy.Action{policy.ActionRead}, Access: "grant"},
			{Actor: "owner", Service: "metrics", Actions: []policy.Action{policy.ActionRead}, Access: "grant"},
			{Actor: "robot", Service: "normal", Actions: []policy.Action{policy.ActionRead}, Access: "static"},
			{Actor: "robot", Service: "sensitive-db", Actions: []policy.Action{policy.ActionRead}, Access: "template", TemplateID: "robot-read"},
			{Actor: "robot", Service: "internet", Actions: []policy.Action{policy.ActionEgress}, Access: "template", TemplateID: "robot-egress"},
		},
		Templates: []policy.Template{
			{ID: "robot-read", Actor: "robot", Services: []string{"sensitive-db"}, Actions: []policy.Action{policy.ActionRead}, MaxTTLSeconds: 7200},
			{ID: "robot-egress", Actor: "robot", Services: []string{"internet"}, Actions: []policy.Action{policy.ActionEgress}, MaxTTLSeconds: 7200},
		},
		Egress: []policy.EgressRule{
			{Actor: "robot", Hosts: []string{"example.com"}, Ports: []int{443}},
		},
	}
}

func initDemo(args []string) error {
	fs := flag.NewFlagSet("init-demo", flag.ExitOnError)
	data := fs.String("data", ".intra", "data dir")
	normal := fs.String("normal", "http://127.0.0.1:18090", "normal upstream")
	logs := fs.String("logs", "http://127.0.0.1:18091", "logs upstream")
	metrics := fs.String("metrics", "http://127.0.0.1:18092", "metrics upstream")
	force := fs.Bool("force", false, "overwrite keys and policy")
	_ = fs.Parse(args)

	if err := os.MkdirAll(*data, 0o700); err != nil {
		return err
	}
	polPath := filepath.Join(*data, "policy.json")
	if _, err := os.Stat(polPath); err == nil && !*force {
		return fmt.Errorf("%s exists; pass --force to replace demo state", polPath)
	}
	if *force {
		for _, name := range []string{"grants.json", "grants.json.tmp", "audit.jsonl", "anchor.json"} {
			_ = os.Remove(filepath.Join(*data, name))
		}
	}
	_ = os.WriteFile(filepath.Join(*data, "audit.jsonl.lock"), nil, 0o600)
	_ = os.WriteFile(filepath.Join(*data, "grants.json"), []byte("[]\n"), 0o600)
	if _, err := os.Stat(filepath.Join(*data, "audit.jsonl")); err != nil {
		_ = os.WriteFile(filepath.Join(*data, "audit.jsonl"), nil, 0o600)
	}
	p := demoPolicy(*normal, *logs, *metrics)
	if err := p.Validate(); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(p, "", "  ")
	if err := os.WriteFile(polPath, append(b, '\n'), 0o644); err != nil {
		return err
	}
	anchorKey := anchorKeyPath()
	if _, err := os.Stat(anchorKey); err != nil {
		if err := os.MkdirAll(filepath.Dir(anchorKey), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(anchorKey, randBytes(32), 0o600); err != nil {
			return err
		}
	}
	for name, content := range map[string][]byte{
		"approval.key": randBytes(32),
		"viewer.token": []byte(hex.EncodeToString(randBytes(18))),
		"admin.token":  []byte(hex.EncodeToString(randBytes(24))),
	} {
		if err := os.WriteFile(filepath.Join(*data, name), content, 0o600); err != nil {
			return err
		}
	}
	certDir := filepath.Join(*data, "certs")
	ca, err := certs.MakeCA(certDir)
	if err != nil {
		return err
	}
	for _, spec := range []struct {
		name, cn string
		server   bool
	}{
		{"guard", "localhost", true},
		{"owner", "owner/mac", false},
		{"control", "owner/control", false},
		{"robot", "robot/run1", false},
		{"drill", "drill/host", false},
	} {
		if err := ca.Issue(certDir, spec.name, spec.cn, spec.server); err != nil {
			return err
		}
	}
	compiled, err := p.Compile()
	if err != nil {
		return err
	}
	cb, _ := json.MarshalIndent(compiled, "", "  ")
	if err := os.WriteFile(filepath.Join(*data, "compiled.json"), append(cb, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("demo ready in %s (policy hash %s)\n", *data, compiled.Hash)
	return nil
}
