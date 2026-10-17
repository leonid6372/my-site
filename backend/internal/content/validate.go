package content

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/leonid6372/my-site/backend/internal/geo"
)

// MaxPhotosPerPlace is the photo limit per place (NFR-08).
const MaxPhotosPerPlace = 30

// LinkKinds are the allowed contact link kinds, in display order.
var LinkKinds = []string{"email", "github", "linkedin", "telegram"}

var idPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// ValidationError is one broken rule, located by file and field path.
type ValidationError struct {
	File string // e.g. "travel.yaml"
	Path string // e.g. "places[3].photos[2].url"
	Msg  string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s:%s: %s", e.File, e.Path, e.Msg)
}

// Validate checks all rules of spec §6 and returns every violation found.
// An empty result means the content is valid.
func Validate(all All) []ValidationError {
	var errs []ValidationError
	file := func(name string) *checker {
		f := all.Files[name]
		if f == "" {
			f = name + ".yaml"
		}
		return &checker{file: f, errs: &errs}
	}
	validateMain(file(FileMain), all.Main)
	validateCV(file(FileCV), all.CV)
	validateTravel(file(FileTravel), all.Travel)
	return errs
}

func validateMain(c *checker, m MainFile) {
	c.ltext("fullName", m.FullName)
	c.ltext("headline", m.Headline)
	c.ltext("location", m.Location)
	c.ltext("currently", m.Currently)
	c.ltext("bio", m.Bio)
	c.photo("photo", m.Photo)
	for i, in := range m.Interests {
		c.ltext(idx("interests", i), in)
	}

	seen := map[string]bool{}
	for i, l := range m.Links {
		p := idx("links", i)
		c.ltext(p+".label", l.Label)
		switch {
		case !slices.Contains(LinkKinds, l.Kind):
			c.add(p+".kind", "must be one of: %s", strings.Join(LinkKinds, ", "))
		case seen[l.Kind]:
			c.add(p+".kind", "duplicate link kind %q", l.Kind)
		}
		seen[l.Kind] = true

		if l.Kind == "email" {
			c.mailto(p+".url", l.URL)
		} else {
			c.httpsURL(p+".url", l.URL, true)
		}
	}
}

func validateCV(c *checker, cv CVFile) {
	c.ltext("summary", cv.Summary)

	ids := map[string]bool{}
	for i, e := range cv.Experience {
		p := idx("experience", i)
		c.id(p+".id", e.ID, ids)
		c.ltext(p+".position", e.Position)
		c.required(p+".company", e.Company)
		c.httpsURL(p+".companyUrl", e.CompanyURL, false)
		c.ltext(p+".location", e.Location)
		c.period(p, "startDate", "endDate", e.StartDate, e.EndDate, false)
		c.optionalLText(p+".description", e.Description)
		for j, h := range e.Highlights {
			c.ltext(idx(p+".highlights", j), h)
		}
	}

	ids = map[string]bool{}
	for i, e := range cv.Education {
		p := idx("education", i)
		c.id(p+".id", e.ID, ids)
		c.ltext(p+".institution", e.Institution)
		c.ltext(p+".degree", e.Degree)
		c.ltext(p+".field", e.Field)
		c.period(p, "startDate", "endDate", e.StartDate, e.EndDate, false)
		c.optionalLText(p+".description", e.Description)
	}

	for i, s := range cv.Skills {
		p := idx("skills", i)
		c.ltext(p+".category", s.Category)
		if len(s.Items) == 0 {
			c.add(p+".items", "required")
		}
		for j, item := range s.Items {
			c.required(idx(p+".items", j), item)
		}
	}

	for i, l := range cv.Languages {
		p := idx("languages", i)
		c.ltext(p+".name", l.Name)
		c.ltext(p+".level", l.Level)
	}
}

func validateTravel(c *checker, t TravelFile) {
	ids := map[string]bool{}
	for i, pl := range t.Places {
		p := idx("places", i)
		c.id(p+".id", pl.ID, ids)

		if pl.Country == "" {
			c.add(p+".country", "required")
		} else if _, ok := geo.Lookup(pl.Country); !ok {
			c.add(p+".country", "unknown ISO 3166-1 alpha-2 code %q", pl.Country)
		}

		c.ltext(p+".name", pl.Name)
		c.coordinate(p+".lat", pl.Lat, 90)
		c.coordinate(p+".lng", pl.Lng, 180)
		c.ltext(p+".story", pl.Story)

		if len(pl.Visits) == 0 {
			c.add(p+".visits", "at least one visit is required")
		}
		for j, v := range pl.Visits {
			c.period(idx(p+".visits", j), "start", "end", v.Start, v.End, true)
		}

		if len(pl.Photos) > MaxPhotosPerPlace {
			c.add(p+".photos", "at most %d photos allowed, got %d", MaxPhotosPerPlace, len(pl.Photos))
		}
		for j, ph := range pl.Photos {
			c.photo(idx(p+".photos", j), ph)
		}
	}
}

// --- checker -------------------------------------------------------------------

type checker struct {
	file string
	errs *[]ValidationError
}

func (c *checker) add(path, format string, args ...any) {
	*c.errs = append(*c.errs, ValidationError{File: c.file, Path: path, Msg: fmt.Sprintf(format, args...)})
}

func (c *checker) required(path, v string) {
	if strings.TrimSpace(v) == "" {
		c.add(path, "required")
	}
}

func (c *checker) ltext(path string, t LText) {
	c.required(path+".en", t.EN)
	c.required(path+".ru", t.RU)
}

func (c *checker) optionalLText(path string, t *LText) {
	if t != nil {
		c.ltext(path, *t)
	}
}

func (c *checker) id(path, id string, seen map[string]bool) {
	switch {
	case id == "":
		c.add(path, "required")
	case !idPattern.MatchString(id):
		c.add(path, "must match %s", idPattern)
	case seen[id]:
		c.add(path, "duplicate id %q", id)
	}
	seen[id] = true
}

// period checks start (required) and end (required when endRequired) with end ≥ start.
func (c *checker) period(path, startKey, endKey string, start, end *Date, endRequired bool) {
	if start == nil {
		c.add(path+"."+startKey, "required")
	}
	if end == nil {
		if endRequired {
			c.add(path+"."+endKey, "required")
		}
		return
	}
	if start != nil && end.Before(start.Time) {
		c.add(path+"."+endKey, "must not be before %s (%s)", startKey, start)
	}
}

func (c *checker) coordinate(path string, v *float64, limit float64) {
	switch {
	case v == nil:
		c.add(path, "required")
	case *v < -limit || *v > limit:
		c.add(path, "must be between %g and %g", -limit, limit)
	}
}

func (c *checker) photo(path string, p Photo) {
	c.httpsURL(path+".url", p.URL, true)
	c.httpsURL(path+".thumbUrl", p.ThumbURL, false)
	c.optionalLText(path+".alt", p.Alt)
	c.optionalLText(path+".caption", p.Caption)
	if (p.Width == nil) != (p.Height == nil) {
		c.add(path, "width and height must be given together")
	}
	if p.Width != nil && *p.Width <= 0 {
		c.add(path+".width", "must be a positive integer")
	}
	if p.Height != nil && *p.Height <= 0 {
		c.add(path+".height", "must be a positive integer")
	}
}

func (c *checker) httpsURL(path, raw string, required bool) {
	if raw == "" {
		if required {
			c.add(path, "required")
		}
		return
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		c.add(path, "must be an absolute https URL")
	}
}

func (c *checker) mailto(path, raw string) {
	addr, ok := strings.CutPrefix(raw, "mailto:")
	if !ok {
		c.add(path, "must be a mailto: URL")
		return
	}
	if a, err := mail.ParseAddress(addr); err != nil || a.Address != addr {
		c.add(path, "invalid email address %q", addr)
	}
}

func idx(path string, i int) string { return fmt.Sprintf("%s[%d]", path, i) }
