package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/leonid6372/my-site/backend/internal/content"
	"github.com/leonid6372/my-site/backend/internal/i18n"
)

var update = flag.Bool("update", false, "regenerate golden files in testdata/golden")

// fakeStore serves content in file order, as loaded from a content directory.
type fakeStore struct {
	all content.All
	err error
}

func (f *fakeStore) GetMain(context.Context) (content.MainFile, error) { return f.all.Main, f.err }
func (f *fakeStore) ListLinks(context.Context) ([]content.Link, error) {
	return f.all.Main.Links, f.err
}
func (f *fakeStore) GetCV(context.Context) (content.CVFile, error) { return f.all.CV, f.err }
func (f *fakeStore) ListPlaces(context.Context) ([]content.Place, error) {
	return f.all.Travel.Places, f.err
}

func fixtureStore(t *testing.T) *fakeStore {
	t.Helper()
	all, err := content.Load(filepath.Join("..", "content", "testdata", "valid"))
	if err != nil {
		t.Fatal(err)
	}
	return &fakeStore{all: all}
}

func marshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(b, '\n')
}

// checkGolden compares got with testdata/golden/<name>, or rewrites it with -update.
func checkGolden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name)
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test ./internal/service -update to create it)", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s differs from golden file\n got:\n%s\nwant:\n%s", name, got, want)
	}
}

func TestGolden(t *testing.T) {
	ctx := context.Background()
	svc := New(fixtureStore(t))
	for _, lang := range []i18n.Lang{i18n.EN, i18n.RU} {
		m, err := svc.Main(ctx, lang)
		if err != nil {
			t.Fatal(err)
		}
		checkGolden(t, "main."+string(lang)+".json", marshal(t, m))

		cv, err := svc.CV(ctx, lang)
		if err != nil {
			t.Fatal(err)
		}
		checkGolden(t, "cv."+string(lang)+".json", marshal(t, cv))
	}
}

// contractExample returns the first ```json block after the given heading in
// specs/api/contract.md.
func contractExample(t *testing.T, heading string) any {
	t.Helper()
	doc, err := os.ReadFile(filepath.Join("..", "..", "..", "specs", "api", "contract.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(doc), "\r\n", "\n")
	_, section, ok := strings.Cut(text, "\n"+heading)
	if !ok {
		t.Fatalf("heading %q not found in contract", heading)
	}
	m := regexp.MustCompile("(?s)```json\n(.*?)\n```").FindStringSubmatch(section)
	if m == nil {
		t.Fatalf("no json example under %q", heading)
	}
	var v any
	if err := json.Unmarshal([]byte(m[1]), &v); err != nil {
		t.Fatalf("contract example under %q: %v", heading, err)
	}
	return v
}

// sameShape reports where got and want differ in structure: object key sets
// must be equal, array elements are checked against the first example
// element, and scalar types must match unless either side is null.
func sameShape(path string, got, want any) []string {
	if got == nil || want == nil {
		return nil
	}
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			return []string{path + ": want object"}
		}
		var diffs []string
		for k := range w {
			if _, ok := g[k]; !ok {
				diffs = append(diffs, path+"."+k+": missing")
			}
		}
		for k := range g {
			if _, ok := w[k]; !ok {
				diffs = append(diffs, path+"."+k+": not in contract")
			} else {
				diffs = append(diffs, sameShape(path+"."+k, g[k], w[k])...)
			}
		}
		return diffs
	case []any:
		g, ok := got.([]any)
		if !ok {
			return []string{path + ": want array"}
		}
		var diffs []string
		if len(w) > 0 {
			for i, e := range g {
				diffs = append(diffs, sameShape(path+"["+strconv.Itoa(i)+"]", e, w[0])...)
			}
		}
		return diffs
	default:
		if reflect.TypeOf(got) != reflect.TypeOf(want) {
			return []string{path + ": type " + reflect.TypeOf(got).String() + ", want " + reflect.TypeOf(want).String()}
		}
		return nil
	}
}

func TestShapesMatchContract(t *testing.T) {
	ctx := context.Background()
	svc := New(fixtureStore(t))
	m, err := svc.Main(ctx, i18n.EN)
	if err != nil {
		t.Fatal(err)
	}
	cv, err := svc.CV(ctx, i18n.EN)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		heading string
		dto     any
	}{
		{"### 2.2", m},
		{"### 2.3", cv},
	} {
		var got any
		if err := json.Unmarshal(marshal(t, tc.dto), &got); err != nil {
			t.Fatal(err)
		}
		diffs := sameShape("$", got, contractExample(t, tc.heading))
		slices.Sort(diffs)
		for _, d := range diffs {
			t.Errorf("contract %s: %s", tc.heading, d)
		}
	}
}

func TestOptionalFieldsAreNullNotOmitted(t *testing.T) {
	svc := New(&fakeStore{all: content.All{CV: content.CVFile{
		Experience: []content.Experience{{ID: "x", StartDate: &content.Date{}}},
		Education:  []content.Education{{ID: "y", StartDate: &content.Date{}}},
	}}})
	cv, err := svc.CV(context.Background(), i18n.EN)
	if err != nil {
		t.Fatal(err)
	}
	b := string(marshal(t, cv))
	for _, want := range []string{`"companyUrl": null`, `"endDate": null`, `"description": null`, `"highlights": []`} {
		if !strings.Contains(b, want) {
			t.Errorf("CV JSON lacks %s:\n%s", want, b)
		}
	}

	m, err := New(&fakeStore{}).Main(context.Background(), i18n.EN)
	if err != nil {
		t.Fatal(err)
	}
	b = string(marshal(t, m))
	for _, want := range []string{`"thumbUrl": null`, `"alt": null`, `"caption": null`, `"width": null`,
		`"height": null`, `"bio": []`, `"interests": []`, `"links": []`} {
		if !strings.Contains(b, want) {
			t.Errorf("main JSON lacks %s:\n%s", want, b)
		}
	}
}

func TestCVSortsByStartDateDesc(t *testing.T) {
	date := func(s string) *content.Date {
		var d content.Date
		if err := json.Unmarshal([]byte(`"`+s+`"`), &d); err != nil {
			t.Fatal(err)
		}
		return &d
	}
	svc := New(&fakeStore{all: content.All{CV: content.CVFile{
		Experience: []content.Experience{
			{ID: "old", StartDate: date("2015-01-01")},
			{ID: "b-new", StartDate: date("2022-03-01")},
			{ID: "a-new", StartDate: date("2022-03-01")},
			{ID: "mid", StartDate: date("2019-06-15")},
		},
		Education: []content.Education{
			{ID: "bsc", StartDate: date("2009-09-01")},
			{ID: "msc", StartDate: date("2013-09-01")},
		},
	}}})
	cv, err := svc.CV(context.Background(), i18n.EN)
	if err != nil {
		t.Fatal(err)
	}
	var exp, edu []string
	for _, e := range cv.Experience {
		exp = append(exp, e.ID)
	}
	for _, e := range cv.Education {
		edu = append(edu, e.ID)
	}
	if want := []string{"a-new", "b-new", "mid", "old"}; !slices.Equal(exp, want) {
		t.Errorf("experience order = %v, want %v", exp, want)
	}
	if want := []string{"msc", "bsc"}; !slices.Equal(edu, want) {
		t.Errorf("education order = %v, want %v", edu, want)
	}
}

func TestParagraphs(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", []string{}},
		{"  \n ", []string{}},
		{"One.", []string{"One."}},
		{"First.\n\nSecond.\n", []string{"First.", "Second."}},
		{"First.\r\n\r\nSecond.", []string{"First.", "Second."}},
		{"A\n  \n\n\nB", []string{"A", "B"}},
		{"Wrapped\nline.\n\nNext.", []string{"Wrapped line.", "Next."}},
	}
	for _, tt := range tests {
		if got := paragraphs(tt.in); !slices.Equal(got, tt.want) {
			t.Errorf("paragraphs(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLocalisesEveryText(t *testing.T) {
	ctx := context.Background()
	store := fixtureStore(t)
	svc := New(store)
	m, err := svc.Main(ctx, i18n.RU)
	if err != nil {
		t.Fatal(err)
	}
	if m.FullName != store.all.Main.FullName.RU || m.Links[0].Label != store.all.Main.Links[0].Label.RU ||
		*m.Photo.Alt != store.all.Main.Photo.Alt.RU {
		t.Errorf("main not localised to ru: %+v", m)
	}
	// The response holds one language only.
	if b := marshal(t, m); bytes.Contains(b, []byte(store.all.Main.Headline.EN)) {
		t.Errorf("RU response contains EN headline: %s", b)
	}
}

func TestStoreErrorsPropagate(t *testing.T) {
	want := errors.New("db down")
	svc := New(&fakeStore{err: want})
	if _, err := svc.Main(context.Background(), i18n.EN); !errors.Is(err, want) {
		t.Errorf("Main err = %v", err)
	}
	if _, err := svc.CV(context.Background(), i18n.EN); !errors.Is(err, want) {
		t.Errorf("CV err = %v", err)
	}
}
