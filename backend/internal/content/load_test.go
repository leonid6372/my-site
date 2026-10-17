package content

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const validDir = "testdata/valid"

func TestLoadValidYAML(t *testing.T) {
	all, err := Load(validDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := all.Main.FullName.RU; got != "Леонид Соловьёв" {
		t.Errorf("main.fullName.ru = %q", got)
	}
	if got := len(all.Main.Links); got != 4 {
		t.Errorf("links = %d, want 4", got)
	}
	exp := all.CV.Experience[0]
	if exp.StartDate.String() != "2021-05-01" || exp.EndDate != nil {
		t.Errorf("experience[0] dates = %v, %v", exp.StartDate, exp.EndDate)
	}
	p := all.Travel.Places[0]
	if *p.Lat != 52.3676 || len(p.Visits) != 2 || len(p.Photos) != 2 || *p.Photos[0].Width != 1600 {
		t.Errorf("places[0] = %+v", p)
	}
	if *all.Travel.Places[2].Lat != -0.1807 {
		t.Errorf("negative latitude not decoded")
	}
	if all.Files[FileTravel] != "travel.yaml" {
		t.Errorf("Files = %v", all.Files)
	}
}

// TestLoadJSONEquivalent writes the YAML fixture as JSON and checks that
// loading it yields the same content.
func TestLoadJSONEquivalent(t *testing.T) {
	want, err := Load(validDir)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "main.json"), want.Main)
	writeJSON(t, filepath.Join(dir, "cv.json"), want.CV)
	writeJSON(t, filepath.Join(dir, "travel.json"), want.Travel)

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load JSON: %v", err)
	}
	if !reflect.DeepEqual(got.Main, want.Main) || !reflect.DeepEqual(got.CV, want.CV) ||
		!reflect.DeepEqual(got.Travel, want.Travel) {
		t.Error("JSON content differs from YAML content")
	}
	if got.Files[FileMain] != "main.json" {
		t.Errorf("Files = %v", got.Files)
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(t *testing.T, dir string)
		wantErr string
	}{
		{
			name: "unknown yaml key",
			mutate: func(t *testing.T, dir string) {
				appendFile(t, filepath.Join(dir, "main.yaml"), "nickname: Leo\n")
			},
			wantErr: "main.yaml",
		},
		{
			name: "unknown json key",
			mutate: func(t *testing.T, dir string) {
				remove(t, filepath.Join(dir, "cv.yaml"))
				write(t, filepath.Join(dir, "cv.json"), `{"summary":{"en":"a","ru":"б"},"extra":1}`)
			},
			wantErr: `cv.json: json: unknown field "extra"`,
		},
		{
			name: "ambiguous source",
			mutate: func(t *testing.T, dir string) {
				write(t, filepath.Join(dir, "main.json"), `{}`)
			},
			wantErr: "main: ambiguous source, found main.yaml and main.json",
		},
		{
			name:    "missing file",
			mutate:  func(t *testing.T, dir string) { remove(t, filepath.Join(dir, "travel.yaml")) },
			wantErr: "travel: no travel.yaml",
		},
		{
			name:    "empty file",
			mutate:  func(t *testing.T, dir string) { write(t, filepath.Join(dir, "cv.yaml"), "  \n") },
			wantErr: "cv.yaml: file is empty",
		},
		{
			name: "bad date",
			mutate: func(t *testing.T, dir string) {
				replaceIn(t, filepath.Join(dir, "travel.yaml"), "2023-04-10", "2023-04-10T10:00:00Z")
			},
			wantErr: `travel.yaml: line 17: invalid date "2023-04-10T10:00:00Z"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := copyValid(t)
			tt.mutate(t, dir)
			_, err := Load(dir)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("got error %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// --- helpers -----------------------------------------------------------------

func copyValid(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	entries, err := os.ReadDir(validDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(validDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(dir, e.Name()), string(data))
	}
	return dir
}

func write(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, string(data))
}

func appendFile(t *testing.T, path, data string) {
	t.Helper()
	old, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, string(old)+data)
}

func replaceIn(t *testing.T, path, old, repl string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	write(t, path, strings.Replace(string(data), old, repl, 1))
}

func remove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
}
