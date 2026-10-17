package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
