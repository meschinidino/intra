package guard

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// VerifyRunManifest checks the signed manifest for actor/device against the
// presented peer certificate and the current time.
func VerifyRunManifest(dir string, key []byte, actor, device string, peer *x509.Certificate, now time.Time) error {
	if peer == nil {
		return fmt.Errorf("peer certificate required")
	}
	name := strings.NewReplacer("/", "-", " ", "-").Replace(actor + "-" + device)
	b, err := os.ReadFile(filepath.Join(dir, name+".manifest.json"))
	if err != nil {
		return err
	}
	var m runManifest
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if m.Actor != actor || m.RunID != device || m.Device != device {
		return fmt.Errorf("manifest identity mismatch")
	}
	if now.After(m.ExpiresAt) {
		return fmt.Errorf("manifest expired")
	}
	fp := sha256.Sum256(peer.Raw)
	if hex.EncodeToString(fp[:]) != m.CertSHA256 {
		return fmt.Errorf("certificate fingerprint mismatch")
	}
	mac := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(mac, "%s:%s:%s", m.Actor, m.RunID, m.ExpiresAt.UTC().Format(time.RFC3339))
	if !hmac.Equal([]byte(hex.EncodeToString(mac.Sum(nil))), []byte(m.Signature)) {
		return fmt.Errorf("manifest signature mismatch")
	}
	return nil
}
