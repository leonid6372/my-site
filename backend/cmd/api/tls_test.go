package main

import (
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/leonid6372/my-site/backend/internal/config"
)

func TestTLSConfigSelfSigned(t *testing.T) {
	c, err := tlsConfig(config.Config{}, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(c.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, host := range []string{"localhost", "api", "127.0.0.1"} {
		if err := leaf.VerifyHostname(host); err != nil {
			t.Errorf("certificate does not cover %s: %v", host, err)
		}
	}
}

func TestTLSConfigFromFiles(t *testing.T) {
	cert, err := selfSignedCert()
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certFile, keyFile := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	writePEM(t, certFile, "CERTIFICATE", cert.Certificate[0])
	writePEM(t, keyFile, "PRIVATE KEY", keyDER)

	log := slog.New(slog.DiscardHandler)
	c, err := tlsConfig(config.Config{TLSCertFile: certFile, TLSKeyFile: keyFile}, log)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Certificates) != 1 {
		t.Fatalf("got %d certificates", len(c.Certificates))
	}
	if _, err := tlsConfig(config.Config{TLSCertFile: certFile, TLSKeyFile: certFile}, log); err == nil {
		t.Error("bad key file: expected an error")
	}
}

func writePEM(t *testing.T, path, typ string, der []byte) {
	t.Helper()
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
}
