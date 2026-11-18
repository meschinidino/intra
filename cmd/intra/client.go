package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func dataDir() string { return env("INTRA_DATA", ".intra") }

func anchorKeyPath() string {
	if v := os.Getenv("INTRA_ANCHOR_KEY"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(dataDir(), "anchor.key")
	}
	return filepath.Join(home, ".config", "intra", "anchor.key")
}

func readAnchorKey(data string) ([]byte, error) {
	if b, err := os.ReadFile(anchorKeyPath()); err == nil {
		return b, nil
	}
	return os.ReadFile(filepath.Join(data, "anchor.key"))
}

func adminToken() string {
	if v := os.Getenv("INTRA_ADMIN_TOKEN"); v != "" {
		return v
	}
	b, err := os.ReadFile(filepath.Join(dataDir(), "admin.token"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func brokerClient() *http.Client {
	if !strings.HasPrefix(brokerBase(), "https://") {
		return http.DefaultClient
	}
	dir := dataDir()
	caPEM, err := os.ReadFile(filepath.Join(dir, "certs", "ca.crt"))
	if err != nil {
		return http.DefaultClient
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caPEM)
	cert, err := tls.LoadX509KeyPair(filepath.Join(dir, "certs", "owner.crt"), filepath.Join(dir, "certs", "owner.key"))
	if err != nil {
		return http.DefaultClient
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool, Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}}}
}

func getJSON(url string, out any) (int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	if tok := adminToken(); tok != "" {
		req.Header.Set("X-Intra-Admin", tok)
	}
	resp, err := brokerClient().Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if out != nil && len(b) > 0 {
		_ = json.Unmarshal(b, out)
	}
	return resp.StatusCode, nil
}
