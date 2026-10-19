package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// insecureClient trusts any certificate (the test servers are self-signed).
var insecureClient = &http.Client{Transport: &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test only
}}

func TestServeHTTPShutsDownGracefully(t *testing.T) {
	cert, err := selfSignedCert()
	if err != nil {
		t.Fatal(err)
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release // an in-flight request during shutdown
		_, _ = io.WriteString(w, "done")
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- serveHTTP(ctx, ln, h, tlsCfg, slog.New(slog.DiscardHandler)) }()

	type result struct {
		body string
		err  error
	}
	got := make(chan result, 1)
	go func() {
		resp, err := insecureClient.Get("https://" + ln.Addr().String() + "/")
		if err != nil {
			got <- result{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		b, err := io.ReadAll(resp.Body)
		got <- result{string(b), err}
	}()

	<-started
	cancel() // like SIGTERM while a request is running
	time.Sleep(50 * time.Millisecond)
	close(release)

	if r := <-got; r.err != nil || r.body != "done" {
		t.Errorf("in-flight request: %q, %v; want it to complete", r.body, r.err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("serveHTTP = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serveHTTP did not return after shutdown")
	}
	if _, err := insecureClient.Get("https://" + ln.Addr().String() + "/"); err == nil {
		t.Error("server still accepts connections after shutdown")
	}
}

func TestHealthURL(t *testing.T) {
	for in, want := range map[string]string{
		":8080":          "https://127.0.0.1:8080/api/v1/health",
		"0.0.0.0:9000":   "https://127.0.0.1:9000/api/v1/health",
		"[::]:8080":      "https://127.0.0.1:8080/api/v1/health",
		"localhost:8081": "https://localhost:8081/api/v1/health",
		"bad":            "https://127.0.0.1:8080/api/v1/health",
	} {
		if got := healthURL(in); got != want {
			t.Errorf("healthURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHealthcheck(t *testing.T) {
	status := http.StatusOK
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	}))
	defer srv.Close()

	if err := healthcheck(srv.URL); err != nil {
		t.Errorf("200: %v", err)
	}
	status = http.StatusServiceUnavailable
	if err := healthcheck(srv.URL); err == nil {
		t.Error("503: want an error")
	}
	srv.Close()
	if err := healthcheck(srv.URL); err == nil {
		t.Error("server down: want an error")
	}
}

func TestRunUsage(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
	}{
		{"no args", nil, 2, "usage: api <command>"},
		{"unknown command", []string{"foo"}, 2, `unknown command "foo"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if code := run(tt.args, &stderr); code != tt.wantCode {
				t.Errorf("exit code = %d, want %d", code, tt.wantCode)
			}
			if !strings.Contains(stderr.String(), tt.wantOut) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tt.wantOut)
			}
		})
	}
}
