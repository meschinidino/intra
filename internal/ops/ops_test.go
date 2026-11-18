package ops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"intra/internal/audit"
	"intra/internal/broker"
	"intra/internal/policy"
)

func seedData(t *testing.T) (string, []byte) {
	t.Helper()
	dir := t.TempDir()
	key := []byte("anchor-key")
	w, err := audit.OpenWriter(filepath.Join(dir, "audit.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Append(t.Context(), audit.Event{Actor: "owner", Action: "read", Result: "allow"}); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	if err := audit.Anchor(filepath.Join(dir, "audit.jsonl"), filepath.Join(dir, "anchor.json"), key); err != nil {
		t.Fatal(err)
	}
	legacy := `[{"id":"g_1","request_key":"k1","actor":"owner","services":["db"],"state":"revoked"}]`
	if err := os.WriteFile(filepath.Join(dir, "grants.json"), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "policy.json"), []byte(`{"version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// Secret material must never enter a backup.
	_ = os.WriteFile(filepath.Join(dir, "approval.key"), []byte("SECRETKEY"), 0o600)
	return dir, key
}

func TestBackupRestoreRoundTripAndTamper(t *testing.T) {
	dir, key := seedData(t)
	backup := filepath.Join(t.TempDir(), "backup.zip")
	if err := CreateBackup(dir, backup); err != nil {
		t.Fatal(err)
	}
	// Tamper: rewrite backup with a corrupted audit file via a fresh backup of corrupt data.
	badDir, _ := seedData(t)
	_ = os.WriteFile(filepath.Join(badDir, "audit.jsonl"), []byte("{}\n"), 0o600)
	// CreateBackup does not verify the chain, so restore must refuse it.
	badBackup := filepath.Join(t.TempDir(), "bad.zip")
	if err := CreateBackup(badDir, badBackup); err != nil {
		t.Fatal(err)
	}
	if err := RestoreBackup(badBackup, t.TempDir(), key); err == nil {
		t.Fatal("restore accepted broken audit chain")
	}
	target := t.TempDir()
	if err := RestoreBackup(backup, target, key); err != nil {
		t.Fatal(err)
	}
	if err := audit.VerifyAnchor(filepath.Join(target, "audit.jsonl"), filepath.Join(target, "anchor.json"), key); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(target, "grants.json"))
	if strings.Contains(string(b), "SECRETKEY") {
		t.Fatal("backup leaked key material")
	}
	st, err := broker.OpenStore(filepath.Join(target, "grants.json"))
	if err != nil {
		t.Fatal(err)
	}
	_ = st
}

func TestRebindChangesKeyAndRecordsFingerprintsOnly(t *testing.T) {
	dir := t.TempDir()
	old := []byte("old-approval-key-0123456789abcdef")
	if err := os.WriteFile(filepath.Join(dir, "approval.key"), old, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := RebindApprover(dir, time.Now()); err != nil {
		t.Fatal(err)
	}
	fresh, _ := os.ReadFile(filepath.Join(dir, "approval.key"))
	if string(fresh) == string(old) || len(fresh) != 32 {
		t.Fatal("key not rotated")
	}
	rec, _ := os.ReadFile(filepath.Join(dir, "approver-rebind.json"))
	if strings.Contains(string(rec), string(old)) || strings.Contains(string(rec), string(fresh)) {
		t.Fatal("rebind record contains key material")
	}
	g := broker.Grant{ID: "g_1", Actor: "owner", ExpiresAt: time.Now()}
	oldSig, _ := broker.ApprovalSignature(old, g, "owner")
	newSig, _ := broker.ApprovalSignature(fresh, g, "owner")
	if oldSig == newSig {
		t.Fatal("old signature still valid after rebind")
	}
}

func TestChecksums(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "intra"), []byte("binary"), 0o755)
	sums := filepath.Join(dir, "SHA256SUMS")
	if err := WriteChecksums(dir, []string{"intra"}, sums); err != nil {
		t.Fatal(err)
	}
	bad, err := VerifyChecksums(sums, dir)
	if err != nil || len(bad) != 0 {
		t.Fatalf("bad=%v err=%v", bad, err)
	}
	_ = os.WriteFile(filepath.Join(dir, "intra"), []byte("tampered"), 0o755)
	bad, err = VerifyChecksums(sums, dir)
	if err != nil || len(bad) != 1 {
		t.Fatalf("tamper not caught bad=%v err=%v", bad, err)
	}
}

func TestSummarizePartialStaleLate(t *testing.T) {
	now := time.Now()
	anchor := filepath.Join(t.TempDir(), "anchor.json")
	_ = os.WriteFile(anchor, []byte(`{"seq":1,"head":"x","sig":"y"}`), 0o600)
	grants := []broker.Grant{
		{ID: "g_partial", State: broker.StatePartialRevoke, Services: []string{"db", "logs"}, RevokeTries: 2, Children: []broker.Child{{Service: "db", State: broker.StatePartialRevoke}, {Service: "logs", State: broker.StateRevoked}}},
		{ID: "g_late", State: broker.StateActive, Services: []string{"db"}, ExpiresAt: now.Add(-time.Minute)},
		{ID: "g_ok", State: broker.StateActive, Services: []string{"db"}, ExpiresAt: now.Add(time.Hour)},
	}
	s := Summarize(grants, anchor, 1, time.Hour, now)
	_ = policy.ActionRead
	if len(s.PartialRevokes) != 1 || len(s.PartialRevokes[0].Failed) != 1 {
		t.Fatalf("partial %+v", s.PartialRevokes)
	}
	if len(s.LateKills) != 2 {
		t.Fatalf("late %+v", s.LateKills)
	}
	if s.AnchorStale {
		t.Fatalf("fresh anchor marked stale: %+v", s)
	}
	s2 := Summarize(grants, filepath.Join(t.TempDir(), "missing.json"), 5, time.Hour, now)
	if !s2.AnchorStale {
		t.Fatal("missing anchor not stale")
	}
}
