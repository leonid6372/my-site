// Package contenttest provides a fixed sample of site content for tests.
package contenttest

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/leonid6372/my-site/backend/internal/content"
)

//go:embed fixture.json
var fixture []byte

// Fixture returns a fresh copy of the sample content.
func Fixture(t testing.TB) content.All {
	t.Helper()
	var all content.All
	if err := json.Unmarshal(fixture, &all); err != nil {
		t.Fatal(err)
	}
	return all
}
