// Package geo provides the embedded country reference data (FR-B09):
// ISO codes, continent and English/Russian names. Regenerate countries.json
// with `go run ./tools/gencountries`.
package geo

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// Continent ids as used in the API contract (§2.4).
const (
	Africa       = "africa"
	Antarctica   = "antarctica"
	Asia         = "asia"
	Europe       = "europe"
	NorthAmerica = "north-america"
	Oceania      = "oceania"
	SouthAmerica = "south-america"
)

// Country is one entry of the reference data.
type Country struct {
	ISO2       string `json:"iso2"`
	ISONumeric string `json:"isoNumeric"` // empty when there is no official code (e.g. XK)
	Continent  string `json:"continent"`
	Name       struct {
		EN string `json:"en"`
		RU string `json:"ru"`
	} `json:"name"`
}

//go:embed countries.json
var countriesJSON []byte

var (
	all    []Country
	byISO2 map[string]Country
)

func init() {
	if err := json.Unmarshal(countriesJSON, &all); err != nil {
		panic(fmt.Sprintf("geo: invalid embedded countries.json: %v", err))
	}
	byISO2 = make(map[string]Country, len(all))
	for _, c := range all {
		byISO2[c.ISO2] = c
	}
}

// Lookup returns the country with the given ISO 3166-1 alpha-2 code.
func Lookup(iso2 string) (Country, bool) {
	c, ok := byISO2[iso2]
	return c, ok
}

// All returns every country, sorted by ISO2. The slice must not be modified.
func All() []Country {
	return all
}
