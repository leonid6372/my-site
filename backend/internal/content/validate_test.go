package content

import (
	"fmt"
	"testing"
)

func loadValid(t *testing.T) All {
	t.Helper()
	all, err := Load(validDir)
	if err != nil {
		t.Fatal(err)
	}
	return all
}

func TestValidateValidFixture(t *testing.T) {
	if errs := Validate(loadValid(t)); len(errs) != 0 {
		t.Fatalf("valid fixture has errors: %v", errs)
	}
}

func TestValidateRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(a *All)
		want   string // exact "file:path: msg" of the single expected error
	}{
		{"missing translation", func(a *All) { a.Main.Headline.RU = " " },
			"main.yaml:headline.ru: required"},
		{"missing required string", func(a *All) { a.CV.Experience[1].Company = "" },
			"cv.yaml:experience[1].company: required"},
		{"bad id pattern", func(a *All) { a.Travel.Places[0].ID = "Amsterdam" },
			"travel.yaml:places[0].id: must match ^[a-z0-9-]+$"},
		{"duplicate id", func(a *All) { a.Travel.Places[1].ID = "amsterdam" },
			`travel.yaml:places[1].id: duplicate id "amsterdam"`},
		{"end before start", func(a *All) {
			a.CV.Education[0].EndDate = new(mustDate("2012-01-01"))
		}, "cv.yaml:education[0].endDate: must not be before startDate (2013-09-01)"},
		{"missing start", func(a *All) { a.CV.Experience[0].StartDate = nil },
			"cv.yaml:experience[0].startDate: required"},
		{"visit without end", func(a *All) { a.Travel.Places[0].Visits[1].End = nil },
			"travel.yaml:places[0].visits[1].end: required"},
		{"unknown country", func(a *All) { a.Travel.Places[2].Country = "XX" },
			`travel.yaml:places[2].country: unknown ISO 3166-1 alpha-2 code "XX"`},
		{"latitude out of range", func(a *All) { a.Travel.Places[0].Lat = new(91.0) },
			"travel.yaml:places[0].lat: must be between -90 and 90"},
		{"missing longitude", func(a *All) { a.Travel.Places[0].Lng = nil },
			"travel.yaml:places[0].lng: required"},
		{"no visits", func(a *All) { a.Travel.Places[1].Visits = nil },
			"travel.yaml:places[1].visits: at least one visit is required"},
		{"too many photos", func(a *All) {
			ps := make([]Photo, 31)
			for i := range ps {
				ps[i] = Photo{URL: fmt.Sprintf("https://img.example/%d.jpg", i)}
			}
			a.Travel.Places[1].Photos = ps
		}, "travel.yaml:places[1].photos: at most 30 photos allowed, got 31"},
		{"http photo url", func(a *All) { a.Travel.Places[0].Photos[1].URL = "http://img.example/a.jpg" },
			"travel.yaml:places[0].photos[1].url: must be an absolute https URL"},
		{"relative thumb url", func(a *All) { a.Travel.Places[0].Photos[0].ThumbURL = "/thumbs/a.jpg" },
			"travel.yaml:places[0].photos[0].thumbUrl: must be an absolute https URL"},
		{"missing main photo url", func(a *All) { a.Main.Photo.URL = "" },
			"main.yaml:photo.url: required"},
		{"width without height", func(a *All) { a.Main.Photo.Height = nil },
			"main.yaml:photo: width and height must be given together"},
		{"non-positive width", func(a *All) { a.Travel.Places[0].Photos[0].Width = new(0) },
			"travel.yaml:places[0].photos[0].width: must be a positive integer"},
		{"caption missing translation", func(a *All) { a.Travel.Places[0].Photos[0].Caption.EN = "" },
			"travel.yaml:places[0].photos[0].caption.en: required"},
		{"unknown link kind", func(a *All) { a.Main.Links[3].Kind = "instagram" },
			"main.yaml:links[3].kind: must be one of: email, github, linkedin, telegram"},
		{"duplicate link kind", func(a *All) { a.Main.Links[2].Kind = "github" },
			`main.yaml:links[2].kind: duplicate link kind "github"`},
		{"bad mailto", func(a *All) { a.Main.Links[0].URL = "mailto:not-an-email" },
			`main.yaml:links[0].url: invalid email address "not-an-email"`},
		{"email without mailto", func(a *All) { a.Main.Links[0].URL = "me@example.com" },
			"main.yaml:links[0].url: must be a mailto: URL"},
		{"non-https link", func(a *All) { a.Main.Links[1].URL = "github.com/example" },
			"main.yaml:links[1].url: must be an absolute https URL"},
		{"empty skill items", func(a *All) { a.CV.Skills[1].Items = nil },
			"cv.yaml:skills[1].items: required"},
		{"json file name used", func(a *All) {
			a.Files[FileCV] = "cv.json"
			a.CV.Summary.EN = ""
		}, "cv.json:summary.en: required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			all := loadValid(t)
			tt.mutate(&all)
			errs := Validate(all)
			if len(errs) != 1 || errs[0].Error() != tt.want {
				t.Fatalf("got %v, want exactly [%s]", errs, tt.want)
			}
		})
	}
}

func TestValidateCollectsAllErrors(t *testing.T) {
	all := loadValid(t)
	all.Main.FullName.RU = ""
	all.CV.Summary.EN = ""
	all.Travel.Places[0].Country = "ZZ"
	if errs := Validate(all); len(errs) != 3 {
		t.Fatalf("got %d errors, want 3: %v", len(errs), errs)
	}
}

func mustDate(s string) Date {
	d, err := parseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}
