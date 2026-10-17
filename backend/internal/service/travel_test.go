package service

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/leonid6372/my-site/backend/internal/content"
	"github.com/leonid6372/my-site/backend/internal/i18n"
)

func mustDate(t *testing.T, s string) *content.Date {
	t.Helper()
	var d content.Date
	if err := json.Unmarshal([]byte(`"`+s+`"`), &d); err != nil {
		t.Fatal(err)
	}
	return &d
}

func place(t *testing.T, id, country string, starts ...string) content.Place {
	t.Helper()
	p := content.Place{ID: id, Country: country, Lat: new(1.0), Lng: new(2.0)}
	for _, s := range starts {
		p.Visits = append(p.Visits, content.Visit{Start: mustDate(t, s), End: mustDate(t, s)})
	}
	return p
}

func travel(t *testing.T, lang i18n.Lang, places ...content.Place) TravelDTO {
	t.Helper()
	dto, err := New(&fakeStore{all: content.All{Travel: content.TravelFile{Places: places}}}).
		Travel(context.Background(), lang)
	if err != nil {
		t.Fatal(err)
	}
	return dto
}

func TestStatsEmpty(t *testing.T) {
	dto := travel(t, i18n.EN)
	if dto.Stats != (StatsDTO{}) {
		t.Errorf("stats = %+v, want all zeros", dto.Stats)
	}
	b := string(marshal(t, dto))
	if !strings.Contains(b, `"countries": []`) || !strings.Contains(b, `"places": []`) {
		t.Errorf("empty lists must be [], got:\n%s", b)
	}
}

func TestStatsCounts(t *testing.T) {
	// 12 countries on 3 continents (europe, asia, south-america); NL has 2 places.
	codes := []string{"NL", "DE", "FR", "ES", "IT", "PT", "BE", "AT", "JP", "TH", "EC", "PE"}
	var places []content.Place
	for i, c := range codes {
		places = append(places, place(t, strings.ToLower(c), c, fmt.Sprintf("2020-01-%02d", i+1)))
	}
	places = append(places, place(t, "rotterdam", "NL", "2021-01-01"))

	got := travel(t, i18n.EN, places...).Stats
	want := StatsDTO{Countries: 12, Continents: 3, Places: 13, WorldPercent: 6.2} // 12/195 = 6.15…
	if got != want {
		t.Errorf("stats = %+v, want %+v", got, want)
	}
}

func TestWorldPercentRounding(t *testing.T) {
	tests := []struct {
		countries int
		want      float64
	}{
		{0, 0}, {1, 0.5}, {12, 6.2}, {39, 20}, {195, 100},
	}
	for _, tt := range tests {
		got := stats(make([]CountryDTO, tt.countries), 0).WorldPercent
		if got != tt.want {
			t.Errorf("%d countries: worldPercent = %v, want %v", tt.countries, got, tt.want)
		}
	}
}

func TestPlacesAndVisitsSortedNewestFirst(t *testing.T) {
	dto := travel(t, i18n.EN,
		place(t, "old", "NL", "2015-05-01"),
		place(t, "multi", "NL", "2016-01-01", "2024-06-01", "2020-03-01"),
		place(t, "b-recent", "ES", "2023-01-01"),
		place(t, "a-recent", "ES", "2023-01-01"),
	)
	var ids []string
	for _, p := range dto.Places {
		ids = append(ids, p.ID)
	}
	if want := []string{"multi", "a-recent", "b-recent", "old"}; !slices.Equal(ids, want) {
		t.Errorf("place order = %v, want %v", ids, want)
	}
	var starts []string
	for _, v := range dto.Places[0].Visits {
		starts = append(starts, v.StartDate)
	}
	if want := []string{"2024-06-01", "2020-03-01", "2016-01-01"}; !slices.Equal(starts, want) {
		t.Errorf("visit order = %v, want %v", starts, want)
	}
	// placeIds follow the order of places.
	for _, c := range dto.Countries {
		if c.ISO2 == "NL" && !slices.Equal(c.PlaceIDs, []string{"multi", "old"}) {
			t.Errorf("NL placeIds = %v", c.PlaceIDs)
		}
	}
}

func TestCountriesSortedByLocalisedName(t *testing.T) {
	// EN: Ecuador < Japan < Netherlands; RU: Нидерланды < Эквадор < Япония.
	places := []content.Place{
		place(t, "tokyo", "JP", "2020-01-01"),
		place(t, "quito", "EC", "2020-01-02"),
		place(t, "amsterdam", "NL", "2020-01-03"),
	}
	for lang, want := range map[i18n.Lang][]string{
		i18n.EN: {"Ecuador", "Japan", "Netherlands"},
		i18n.RU: {"Нидерланды", "Эквадор", "Япония"},
	} {
		var got []string
		for _, c := range travel(t, lang, places...).Countries {
			got = append(got, c.Name)
		}
		if !slices.Equal(got, want) {
			t.Errorf("%s: countries = %v, want %v", lang, got, want)
		}
	}
}

func TestSortByNameUsesCollation(t *testing.T) {
	names := func(cs []CountryDTO) []string {
		var out []string
		for _, c := range cs {
			out = append(out, c.Name)
		}
		return out
	}
	mk := func(ns ...string) []CountryDTO {
		var cs []CountryDTO
		for _, n := range ns {
			cs = append(cs, CountryDTO{Name: n})
		}
		return cs
	}

	// Byte order would put "Ёлка" (U+0401) before "Абхазия"; Russian collation
	// treats Ё as a variant of Е (and к < ь decides Ёлка < Ель).
	ru := mk("Жук", "Ель", "Ёлка", "Абхазия")
	sortByName(ru, i18n.RU)
	if want := []string{"Абхазия", "Ёлка", "Ель", "Жук"}; !slices.Equal(names(ru), want) {
		t.Errorf("RU order = %v, want %v", names(ru), want)
	}

	// Byte order would put "Åland Islands" after "Zambia".
	en := mk("Zambia", "Åland Islands", "Austria", "albania")
	sortByName(en, i18n.EN)
	if want := []string{"Åland Islands", "albania", "Austria", "Zambia"}; !slices.Equal(names(en), want) {
		t.Errorf("EN order = %v, want %v", names(en), want)
	}
}

func TestStoryKeepsBlankLineSeparators(t *testing.T) {
	p := place(t, "x", "NL", "2020-01-01")
	p.Story = content.LText{EN: "First\nline.\n\n\nSecond.\n", RU: "Один."}
	if got := travel(t, i18n.EN, p).Places[0].Story; got != "First line.\n\nSecond." {
		t.Errorf("story = %q", got)
	}
}

func TestUnknownCountryIsAnError(t *testing.T) {
	_, err := New(&fakeStore{all: content.All{Travel: content.TravelFile{
		Places: []content.Place{place(t, "x", "ZZ", "2020-01-01")},
	}}}).Travel(context.Background(), i18n.EN)
	if err == nil {
		t.Fatal("expected an error for an unknown country")
	}
}
