// Command gencountries generates internal/geo/countries.json from the
// mledoze/countries dataset (plan §6). Run it from the backend directory:
//
//	go run ./tools/gencountries
//
// The output is committed; the api binary never downloads anything.
package main

import (
	"cmp"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"time"
)

const defaultSource = "https://raw.githubusercontent.com/mledoze/countries/master/countries.json"

type sourceCountry struct {
	CCA2      string `json:"cca2"`
	CCN3      string `json:"ccn3"`
	Region    string `json:"region"`
	Subregion string `json:"subregion"`
	Name      struct {
		Common string `json:"common"`
	} `json:"name"`
	Translations map[string]struct {
		Common string `json:"common"`
	} `json:"translations"`
}

type country struct {
	ISO2       string            `json:"iso2"`
	ISONumeric string            `json:"isoNumeric"`
	Continent  string            `json:"continent"`
	Name       map[string]string `json:"name"`
}

// extras are places the source dataset doesn't list. Their codes are not
// official ISO codes, so they must not clash with one.
var extras = []country{
	{ISO2: "AB", Continent: "asia", Name: map[string]string{"en": "Abkhazia", "ru": "Абхазия"}},
}

func main() {
	src := flag.String("src", defaultSource, "source URL or local file path")
	out := flag.String("out", "internal/geo/countries.json", "output file")
	flag.Parse()

	if err := run(*src, *out); err != nil {
		fmt.Fprintln(os.Stderr, "gencountries:", err)
		os.Exit(1)
	}
}

func run(src, out string) error {
	raw, err := read(src)
	if err != nil {
		return err
	}
	var in []sourceCountry
	if err := json.Unmarshal(raw, &in); err != nil {
		return fmt.Errorf("decode source: %w", err)
	}

	result := make([]country, 0, len(in))
	for _, c := range in {
		cont, err := continent(c.Region, c.Subregion)
		if err != nil {
			return fmt.Errorf("%s: %w", c.CCA2, err)
		}
		ru := c.Translations["rus"].Common
		if ru == "" {
			return fmt.Errorf("%s: missing Russian name", c.CCA2)
		}
		result = append(result, country{
			ISO2:       c.CCA2,
			ISONumeric: c.CCN3, // empty for e.g. Kosovo (no official code)
			Continent:  cont,
			Name:       map[string]string{"en": c.Name.Common, "ru": ru},
		})
	}
	result = append(result, extras...)
	slices.SortFunc(result, func(a, b country) int { return cmp.Compare(a.ISO2, b.ISO2) })

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, append(data, '\n'), 0o600); err != nil {
		return err
	}
	fmt.Printf("wrote %d countries to %s\n", len(result), out)
	return nil
}

// continent maps the dataset's region/subregion to the contract's continent ids.
func continent(region, subregion string) (string, error) {
	switch region {
	case "Europe":
		return "europe", nil
	case "Asia":
		return "asia", nil
	case "Africa":
		return "africa", nil
	case "Oceania":
		return "oceania", nil
	case "Antarctic":
		return "antarctica", nil
	case "Americas":
		if subregion == "South America" {
			return "south-america", nil
		}
		return "north-america", nil // North America, Central America, Caribbean
	}
	return "", fmt.Errorf("unknown region %q", region)
}

func read(src string) ([]byte, error) {
	if _, err := os.Stat(src); err == nil {
		return os.ReadFile(src) //nolint:gosec // path comes from the developer's flag
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(src)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", src, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
