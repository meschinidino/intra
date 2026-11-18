package certs

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestIssueWithTTL(t *testing.T) {
	dir := t.TempDir()
	ca, err := MakeCA(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := ca.IssueWithTTL(dir, "run1", "robot/run1", false, 2*time.Hour); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "run1.crt"))
	block, _ := pem.Decode(b)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if time.Until(cert.NotAfter) > 3*time.Hour {
		t.Fatalf("cert lives too long: %v", time.Until(cert.NotAfter))
	}
}
