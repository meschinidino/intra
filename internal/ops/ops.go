// Package ops holds local pilot-operations helpers: backup, restore,
// approver rebinding, release checksum checks, and operator summaries.
// Every helper fails closed on a checksum, chain, or format mismatch.
package ops

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"intra/internal/audit"
	"intra/internal/broker"
	"intra/internal/version"
)

var backupFiles = []string{"grants.json", "audit.jsonl", "anchor.json", "policy.json", "compiled.json"}

type manifestFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

type manifest struct {
	Format  int            `json:"format"`
	Version string         `json:"version"`
	Created time.Time      `json:"created"`
	Files   []manifestFile `json:"files"`
}

func sha256Hex(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// CreateBackup writes one zip backup with a checksum manifest.
// It includes grant, audit, anchor, and policy state only.
// It never includes keys, certificates, or tokens.
func CreateBackup(dataDir, outPath string) error {
	var m manifest
	m.Format, m.Version, m.Created = 1, version.Version, time.Now().UTC()
	var bufs []struct {
		name string
		b    []byte
	}
	for _, name := range backupFiles {
		b, err := os.ReadFile(filepath.Join(dataDir, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		m.Files = append(m.Files, manifestFile{Path: name, SHA256: sha256Hex(b), Size: int64(len(b))})
		bufs = append(bufs, struct {
			name string
			b    []byte
		}{name, b})
	}
	if len(bufs) == 0 {
		return fmt.Errorf("nothing to back up in %s", dataDir)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	mb, _ := json.MarshalIndent(m, "", "  ")
	for _, item := range append([]struct {
		name string
		b    []byte
	}{{"manifest.json", append(mb, '\n')}}, bufs...) {
		w, err := zw.Create(item.name)
		if err != nil {
			return err
		}
		if _, err := w.Write(item.b); err != nil {
			return err
		}
	}
	return zw.Close()
}

// RestoreBackup verifies a backup, then installs it into dataDir.
// It refuses a mismatched checksum, a broken audit chain, and a bad anchor.
// The caller must stop intrad before restore.
func RestoreBackup(backupPath, dataDir string, anchorKey []byte) error {
	r, err := zip.OpenReader(backupPath)
	if err != nil {
		return err
	}
	defer r.Close()
	contents := map[string][]byte{}
	for _, zf := range r.File {
		if strings.Contains(zf.Name, "..") || filepath.IsAbs(zf.Name) {
			return fmt.Errorf("unsafe backup entry %s", zf.Name)
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		b, err := io.ReadAll(io.LimitReader(rc, 64<<20))
		_ = rc.Close()
		if err != nil {
			return err
		}
		contents[zf.Name] = b
	}
	mb, ok := contents["manifest.json"]
	if !ok {
		return fmt.Errorf("backup has no manifest.json")
	}
	var m manifest
	if err := json.Unmarshal(mb, &m); err != nil {
		return fmt.Errorf("backup manifest: %w", err)
	}
	if m.Format != 1 {
		return fmt.Errorf("unsupported backup format %d", m.Format)
	}
	for _, mf := range m.Files {
		b, ok := contents[mf.Path]
		if !ok {
			return fmt.Errorf("backup missing %s", mf.Path)
		}
		if sha256Hex(b) != mf.SHA256 || int64(len(b)) != mf.Size {
			return fmt.Errorf("backup checksum mismatch for %s", mf.Path)
		}
	}
	// Verify the audit chain and anchor from backup bytes before install.
	tmp, err := os.MkdirTemp("", "intra-restore-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	for name, b := range contents {
		if name == "manifest.json" {
			continue
		}
		if err := os.WriteFile(filepath.Join(tmp, name), b, 0o600); err != nil {
			return err
		}
	}
	logPath := filepath.Join(tmp, "audit.jsonl")
	if _, err := os.Stat(logPath); err == nil {
		if err := audit.Verify(logPath); err != nil {
			return fmt.Errorf("backup audit chain: %w", err)
		}
		if _, err := os.Stat(filepath.Join(tmp, "anchor.json")); err == nil && len(anchorKey) > 0 {
			if err := audit.VerifyAnchor(logPath, filepath.Join(tmp, "anchor.json"), anchorKey); err != nil {
				return fmt.Errorf("backup anchor: %w", err)
			}
		}
	}
	if gb, ok := contents["grants.json"]; ok {
		if _, err := broker.OpenStoreBytes(gb); err != nil {
			return fmt.Errorf("backup grant store: %w", err)
		}
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	for _, mf := range m.Files {
		if err := os.WriteFile(filepath.Join(dataDir, mf.Path), contents[mf.Path], 0o600); err != nil {
			return err
		}
	}
	return nil
}

// RebindApprover replaces the approval key and writes a fingerprint record.
// The old key leaves no copy. The record holds SHA-256 fingerprints only.
// Restart intrad after rebind so the broker loads the new key.
func RebindApprover(dataDir string, now time.Time) error {
	keyPath := filepath.Join(dataDir, "approval.key")
	old, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("read current approval key: %w", err)
	}
	fresh := make([]byte, 32)
	if _, err := rand.Read(fresh); err != nil {
		return err
	}
	if bytes.Equal(old, fresh) {
		return fmt.Errorf("new approval key equals old key")
	}
	if err := os.WriteFile(keyPath, fresh, 0o600); err != nil {
		return err
	}
	rec := map[string]any{
		"ts": now.UTC(), "old_key_sha256": sha256Hex(old), "new_key_sha256": sha256Hex(fresh),
		"note": "old approver device declared revoked; restart intrad to load the new key",
	}
	b, _ := json.MarshalIndent(rec, "", "  ")
	return os.WriteFile(filepath.Join(dataDir, "approver-rebind.json"), append(b, '\n'), 0o600)
}

// WriteChecksums writes a SHA256SUMS file for the named files in dir.
func WriteChecksums(dir string, names []string, outPath string) error {
	var lines []string
	for _, name := range names {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		lines = append(lines, fmt.Sprintf("%s  %s", sha256Hex(b), name))
	}
	return os.WriteFile(outPath, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

// VerifyChecksums checks every entry in a SHA256SUMS file under baseDir.
// It returns the list of mismatched or missing file names.
func VerifyChecksums(checksumsPath, baseDir string) ([]string, error) {
	b, err := os.ReadFile(checksumsPath)
	if err != nil {
		return nil, err
	}
	var bad []string
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed checksum line %q", line)
		}
		want, name := parts[0], parts[1]
		seen[name] = true
		data, err := os.ReadFile(filepath.Join(baseDir, name))
		if err != nil || sha256Hex(data) != want {
			bad = append(bad, name)
		}
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("no checksums in %s", checksumsPath)
	}
	return bad, nil
}

// StatusSummary is the operator view of grant and anchor health.
type StatusSummary struct {
	GrantsByState  map[string]int  `json:"grants_by_state"`
	PartialRevokes []PartialRevoke `json:"partial_revokes,omitempty"`
	LateKills      []LateKill      `json:"late_kills,omitempty"`
	AnchorStale    bool            `json:"anchor_stale"`
	AnchorReason   string          `json:"anchor_reason,omitempty"`
}

type PartialRevoke struct {
	GrantID     string   `json:"grant_id"`
	Services    []string `json:"services"`
	Failed      []string `json:"failed_children,omitempty"`
	RevokeTries int      `json:"revoke_tries"`
	Escalated   bool     `json:"escalated"`
}

type LateKill struct {
	GrantID        string   `json:"grant_id"`
	State          string   `json:"state"`
	Services       []string `json:"services"`
	OverdueSeconds int      `json:"overdue_seconds"`
}

// Summarize builds the operator summary from public grant records.
// A late kill is an expired active grant or a grant stuck in a revoke state.
func Summarize(grants []broker.Grant, anchorPath string, headSeq uint64, maxAnchorAge time.Duration, now time.Time) StatusSummary {
	out := StatusSummary{GrantsByState: map[string]int{}}
	for _, g := range grants {
		out.GrantsByState[string(g.State)]++
		switch g.State {
		case broker.StatePartialRevoke, broker.StateRevokeEscalate:
			pr := PartialRevoke{GrantID: g.ID, Services: g.Services, RevokeTries: g.RevokeTries, Escalated: g.State == broker.StateRevokeEscalate}
			for _, c := range g.Children {
				if c.State == broker.StatePartialRevoke || c.State == broker.StateRevokeEscalate {
					pr.Failed = append(pr.Failed, c.Service)
				}
			}
			out.PartialRevokes = append(out.PartialRevokes, pr)
			out.LateKills = append(out.LateKills, LateKill{GrantID: g.ID, State: string(g.State), Services: g.Services, OverdueSeconds: 0})
		case broker.StateRevoking:
			out.LateKills = append(out.LateKills, LateKill{GrantID: g.ID, State: string(g.State), Services: g.Services})
		case broker.StateActive:
			if now.After(g.ExpiresAt) {
				out.LateKills = append(out.LateKills, LateKill{GrantID: g.ID, State: string(g.State), Services: g.Services, OverdueSeconds: int(now.Sub(g.ExpiresAt).Seconds())})
			}
		}
	}
	out.AnchorStale, out.AnchorReason = audit.StaleAnchor(anchorPath, headSeq, maxAnchorAge, now)
	return out
}
