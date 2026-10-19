package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"time"

	"github.com/leonid6372/my-site/backend/internal/config"
)

// tlsConfig returns the server TLS settings. With TLS_CERT_FILE and
// TLS_KEY_FILE it uses that key pair; otherwise it generates a self-signed
// certificate, which is enough for local development and for nginx talking
// to the API inside the Docker network.
func tlsConfig(cfg config.Config, log *slog.Logger) (*tls.Config, error) {
	var (
		cert tls.Certificate
		err  error
	)
	if cfg.TLSCertFile != "" {
		cert, err = tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS key pair: %w", err)
		}
	} else {
		log.Warn("TLS_CERT_FILE/TLS_KEY_FILE not set: using a self-signed certificate")
		cert, err = selfSignedCert()
		if err != nil {
			return nil, err
		}
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil
}

// selfSignedCert creates an in-memory ECDSA P-256 certificate for localhost
// and the "api" Docker service name, valid for one year.
func selfSignedCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate key: %w", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate serial: %w", err)
	}
	now := time.Now()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.AddDate(1, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost", "api"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, nil
}
