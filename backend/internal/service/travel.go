package service

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"slices"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"

	"github.com/leonid6372/my-site/backend/internal/geo"
	"github.com/leonid6372/my-site/backend/internal/i18n"
)

// worldCountries is the denominator of stats.worldPercent (contract §2.4).
const worldCountries = 195

// StatsDTO is travel.stats.
type StatsDTO struct {
	Countries    int     `json:"countries"`
	Continents   int     `json:"continents"`
	Places       int     `json:"places"`
	WorldPercent float64 `json:"worldPercent"`
}

// CountryDTO is a visited country.
type CountryDTO struct {
	ISO2       string   `json:"iso2"`
	ISONumeric string   `json:"isoNumeric"`
	Name       string   `json:"name"`
	Continent  string   `json:"continent"`
	PlaceIDs   []string `json:"placeIds"`
}

// VisitDTO is one trip.
type VisitDTO struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

// PlaceDTO is a visited place.
type PlaceDTO struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	CountryISO2 string     `json:"countryIso2"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	Story       string     `json:"story"`
	Visits      []VisitDTO `json:"visits"`
	Photos      []PhotoDTO `json:"photos"`
}

// TravelDTO is the GET /api/v1/travel response (contract §2.4).
type TravelDTO struct {
	Stats     StatsDTO     `json:"stats"`
	Countries []CountryDTO `json:"countries"`
	Places    []PlaceDTO   `json:"places"`
}

// Travel builds the /travel response in lang. Places are sorted by latest
// visit (newest first), their visits by start date (newest first), and
// countries by localised name.
func (s *Service) Travel(ctx context.Context, lang i18n.Lang) (TravelDTO, error) {
	places, err := s.store.ListPlaces(ctx)
	if err != nil {
		return TravelDTO{}, err
	}

	dto := TravelDTO{Places: make([]PlaceDTO, 0, len(places))}
	for _, p := range places {
		pd := PlaceDTO{
			ID:          p.ID,
			Name:        loc(p.Name, lang),
			CountryISO2: p.Country,
			Lat:         deref(p.Lat),
			Lng:         deref(p.Lng),
			Story:       strings.Join(paragraphs(loc(p.Story, lang)), "\n\n"),
			Visits:      make([]VisitDTO, 0, len(p.Visits)),
			Photos:      make([]PhotoDTO, 0, len(p.Photos)),
		}
		for _, v := range p.Visits {
			pd.Visits = append(pd.Visits, VisitDTO{StartDate: v.Start.String(), EndDate: v.End.String()})
		}
		slices.SortStableFunc(pd.Visits, func(a, b VisitDTO) int { return cmp.Compare(b.StartDate, a.StartDate) })
		for _, ph := range p.Photos {
			pd.Photos = append(pd.Photos, photo(ph, lang))
		}
		dto.Places = append(dto.Places, pd)
	}
	slices.SortStableFunc(dto.Places, func(a, b PlaceDTO) int {
		return cmp.Or(cmp.Compare(latestVisit(b), latestVisit(a)), cmp.Compare(a.ID, b.ID))
	})

	dto.Countries, err = countries(dto.Places, lang)
	if err != nil {
		return TravelDTO{}, err
	}
	dto.Stats = stats(dto.Countries, len(dto.Places))
	return dto, nil
}

// latestVisit is the newest start date of a place whose visits are sorted.
func latestVisit(p PlaceDTO) string {
	if len(p.Visits) == 0 {
		return ""
	}
	return p.Visits[0].StartDate
}

// countries groups places by country (placeIds keep the order of places)
// and sorts the countries by localised name.
func countries(places []PlaceDTO, lang i18n.Lang) ([]CountryDTO, error) {
	out := []CountryDTO{}
	index := map[string]int{}
	for _, p := range places {
		i, ok := index[p.CountryISO2]
		if !ok {
			c, found := geo.Lookup(p.CountryISO2)
			if !found {
				return nil, fmt.Errorf("place %q: unknown country %q", p.ID, p.CountryISO2)
			}
			name := c.Name.EN
			if lang == i18n.RU {
				name = c.Name.RU
			}
			i = len(out)
			index[p.CountryISO2] = i
			out = append(out, CountryDTO{
				ISO2: c.ISO2, ISONumeric: c.ISONumeric, Name: name, Continent: c.Continent, PlaceIDs: []string{},
			})
		}
		out[i].PlaceIDs = append(out[i].PlaceIDs, p.ID)
	}
	sortByName(out, lang)
	return out, nil
}

// sortByName sorts countries with the collation rules of lang (e.g. "Ё" next
// to "Е" in Russian, "Å" next to "A" in English), not by byte order.
func sortByName(cs []CountryDTO, lang i18n.Lang) {
	tag := language.English
	if lang == i18n.RU {
		tag = language.Russian
	}
	// A Collator is not safe for concurrent use, so create one per call.
	col := collate.New(tag)
	slices.SortStableFunc(cs, func(a, b CountryDTO) int {
		return cmp.Or(col.CompareString(a.Name, b.Name), cmp.Compare(a.ISO2, b.ISO2))
	})
}

func stats(cs []CountryDTO, places int) StatsDTO {
	continents := map[string]struct{}{}
	for _, c := range cs {
		continents[c.Continent] = struct{}{}
	}
	return StatsDTO{
		Countries:    len(cs),
		Continents:   len(continents),
		Places:       places,
		WorldPercent: math.Round(float64(len(cs))/worldCountries*100*10) / 10,
	}
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
