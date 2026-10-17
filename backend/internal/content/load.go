package content

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Logical content file names.
const (
	FileMain   = "main"
	FileCV     = "cv"
	FileTravel = "travel"
)

var extensions = []string{".yaml", ".yml", ".json"}

// Load reads main, cv and travel from dir. Each must exist exactly once as
// .yaml, .yml or .json. Decoding is strict: unknown keys are errors.
// Load does not apply the validation rules; call Validate for that.
func Load(dir string) (All, error) {
	all := All{Files: map[string]string{}}
	targets := []struct {
		name string
		dst  any
	}{
		{FileMain, &all.Main},
		{FileCV, &all.CV},
		{FileTravel, &all.Travel},
	}
	var errs []error
	for _, t := range targets {
		file, err := findSource(dir, t.name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		all.Files[t.name] = filepath.Base(file)
		if err := decodeFile(file, t.dst); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(file), err))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return All{}, err
	}
	return all, nil
}

func findSource(dir, name string) (string, error) {
	var found []string
	for _, ext := range extensions {
		p := filepath.Join(dir, name+ext)
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf("%s: no %s.yaml, %s.yml or %s.json in %s", name, name, name, name, dir)
	case 1:
		return found[0], nil
	default:
		names := make([]string, len(found))
		for i, f := range found {
			names[i] = filepath.Base(f)
		}
		return "", fmt.Errorf("%s: ambiguous source, found %s", name, strings.Join(names, " and "))
	}
}

func decodeFile(path string, dst any) error {
	data, err := os.ReadFile(path) //nolint:gosec // path is built from CONTENT_DIR and fixed names
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return errors.New("file is empty")
	}
	if strings.EqualFold(filepath.Ext(path), ".json") {
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.DisallowUnknownFields()
		if err := dec.Decode(dst); err != nil {
			return err
		}
		if dec.More() {
			return errors.New("unexpected data after the JSON value")
		}
		return nil
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(dst); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
