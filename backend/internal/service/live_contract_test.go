//go:build live

package service

// Compares a running API with specs/api/contract.md (task B21):
//
//	BASE_URL=https://localhost:8080 go test -tags=live -run TestLive ./internal/service
//
// It checks the shape of every response against the contract examples and the
// ordering rules on the data in the API's database.

import (
	"cmp"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"slices"
	"testing"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/leonid6372/my-site/backend/internal/i18n"
)

func langOf(s string) i18n.Lang { return i18n.Lang(s) }

// liveClient accepts the API's self-signed development certificate.
var liveClient = &http.Client{Transport: &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test against a local API
}}

func liveGet(t *testing.T, path, lang string, v any) http.Header {
	t.Helper()
	base := cmp.Or(os.Getenv("BASE_URL"), "https://localhost:8080")
	req, err := http.NewRequest(http.MethodGet, base+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept-Language", lang)
	resp, err := liveClient.Do(req)
	if err != nil {
		t.Fatalf("%s: %v (is the stack running? set BASE_URL)", path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s: status %d: %s", path, resp.StatusCode, body)
	}
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return resp.Header
}

func TestLiveShapesMatchContract(t *testing.T) {
	for _, tc := range []struct{ path, heading string }{
		{"/api/v1/health", "### 2.1"},
		{"/api/v1/main", "### 2.2"},
		{"/api/v1/cv", "### 2.3"},
		{"/api/v1/travel", "### 2.4"},
	} {
		for _, lang := range []string{"en", "ru"} {
			var got any
			liveGet(t, tc.path, lang, &got)
			diffs := sameShape("$", got, contractExample(t, tc.heading))
			slices.Sort(diffs)
			for _, d := range diffs {
				t.Errorf("%s [%s] vs contract %s: %s", tc.path, lang, tc.heading, d)
			}
		}
	}
}

// TestLiveOrderingMatchesContract checks the ordering rules of the contract
// on whatever data the running API's database holds.
func TestLiveOrderingMatchesContract(t *testing.T) {
	for _, lang := range []string{"en", "ru"} {
		var cv CVDTO
		liveGet(t, "/api/v1/cv", lang, &cv)
		if !slices.IsSortedFunc(cv.Experience, func(a, b ExperienceDTO) int { return cmp.Compare(b.StartDate, a.StartDate) }) {
			t.Errorf("[%s] experience not sorted by startDate desc", lang)
		}
		if !slices.IsSortedFunc(cv.Education, func(a, b EducationDTO) int { return cmp.Compare(b.StartDate, a.StartDate) }) {
			t.Errorf("[%s] education not sorted by startDate desc", lang)
		}

		var tr TravelDTO
		liveGet(t, "/api/v1/travel", lang, &tr)
		tag := language.English
		if lang == "ru" {
			tag = language.Russian
		}
		col := collate.New(tag)
		if !slices.IsSortedFunc(tr.Countries, func(a, b CountryDTO) int { return col.CompareString(a.Name, b.Name) }) {
			t.Errorf("[%s] countries not sorted by localised name", lang)
		}
		if !slices.IsSortedFunc(tr.Places, func(a, b PlaceDTO) int { return cmp.Compare(latestVisit(b), latestVisit(a)) }) {
			t.Errorf("[%s] places not sorted by latest visit desc", lang)
		}
		for _, p := range tr.Places {
			if !slices.IsSortedFunc(p.Visits, func(a, b VisitDTO) int { return cmp.Compare(b.StartDate, a.StartDate) }) {
				t.Errorf("[%s] %s visits not sorted desc", lang, p.ID)
			}
		}
		if tr.Stats.Places != len(tr.Places) {
			t.Errorf("[%s] stats.places = %d, response has %d places", lang, tr.Stats.Places, len(tr.Places))
		}
	}
}
