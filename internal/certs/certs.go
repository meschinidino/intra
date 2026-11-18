package certs

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

type CA struct {
	Cert *x509.Certificate
	Key  *ecdsa.PrivateKey
	PEM  []byte
}

func write(path string, b []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, b, mode)
}

func MakeCA(dir string) (*CA, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "intra-lab-ca"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(365 * 24 * time.Hour),
		IsCA: true, KeyUsage: x509.KeyUsageCertSign, BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	cert, _ := x509.ParseCertificate(der)
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, _ := x509.MarshalECPrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if err := write(filepath.Join(dir, "ca.crt"), pemBytes, 0o644); err != nil {
		return nil, err
	}
	if err := write(filepath.Join(dir, "ca.key"), keyPEM, 0o600); err != nil {
		return nil, err
	}
	return &CA{Cert: cert, Key: key, PEM: pemBytes}, nil
}

func LoadCA(crtPath, keyPath string) (*CA, error) {
	cb, err := os.ReadFile(crtPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(cb)
	if block == nil {
		return nil, fmt.Errorf("bad ca cert")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	kb, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	kblock, _ := pem.Decode(kb)
	key, err := x509.ParseECPrivateKey(kblock.Bytes)
	if err != nil {
		return nil, err
	}
	return &CA{Cert: cert, Key: key, PEM: cb}, nil
}

// Issue writes a 90-day leaf cert. CN uses actor/device for clients.
func (ca *CA) Issue(dir, name, cn string, server bool) error {
	return ca.IssueWithTTL(dir, name, cn, server, 90*24*time.Hour)
}

// IssueWithTTL writes a leaf cert that expires with the granted identity.
func (ca *CA) IssueWithTTL(dir, name, cn string, server bool, ttl time.Duration) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour), NotAfter: time.Now().Add(ttl),
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	if server {
		tpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
		tpl.DNSNames = []string{"localhost", "guard", "broker", "audit", "intra"}
		tpl.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	} else {
		tpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, ca.Cert, &key.PublicKey, ca.Key)
	if err != nil {
		return err
	}
	keyDER, _ := x509.MarshalECPrivateKey(key)
	if err := write(filepath.Join(dir, name+".crt"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644); err != nil {
		return err
	}
	return write(filepath.Join(dir, name+".key"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600)
}
