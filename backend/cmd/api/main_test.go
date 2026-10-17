package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRepoContentIsValid keeps the committed content/ valid.
func TestRepoContentIsValid(t *testing.T) {
	var out bytes.Buffer
	if _, err := loadContent(filepath.Join("..", "..", "..", "content"), &out); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
}

func TestLoadContentPrintsErrors(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join("..", "..", "internal", "content", "testdata", "valid")
	for _, name := range []string{"main.yaml", "cv.yaml", "travel.yaml"} {
		data, err := os.ReadFile(filepath.Join(src, name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "main.yaml" {
			data = bytes.Replace(data, []byte(`ru: "Инженер-программист"`), []byte(`ru: ""`), 1)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	var out bytes.Buffer
	_, err := loadContent(dir, &out)
	if err == nil || err.Error() != "content has 1 validation error(s)" {
		t.Fatalf("err = %v", err)
	}
	if got := strings.TrimSpace(out.String()); got != "main.yaml:headline.ru: required" {
		t.Errorf("output = %q", got)
	}
}

func TestServeHTTPShutsDownGracefully(t *testing.T) {
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
	go func() { done <- serveHTTP(ctx, ln, h, slog.New(slog.DiscardHandler)) }()

	type result struct {
		body string
		err  error
	}
	got := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
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
	if _, err := http.Get("http://" + ln.Addr().String() + "/"); err == nil {
		t.Error("server still accepts connections after shutdown")
	}
}

func TestHealthURL(t *testing.T) {
	for in, want := range map[string]string{
		":8080":          "http://127.0.0.1:8080/api/v1/health",
		"0.0.0.0:9000":   "http://127.0.0.1:9000/api/v1/health",
		"[::]:8080":      "http://127.0.0.1:8080/api/v1/health",
		"localhost:8081": "http://localhost:8081/api/v1/health",
		"bad":            "http://127.0.0.1:8080/api/v1/health",
	} {
		if got := healthURL(in); got != want {
			t.Errorf("healthURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHealthcheck(t *testing.T) {
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
