// Package content defines the site content model. The data lives only in
// PostgreSQL; this package has no loading or seeding logic.
package content

import (
	"encoding/json"
	"fmt"
	"time"
)

// LText is a localised text; both languages are required (constitution §5.3).
type LText struct {
	EN string `json:"en"`
	RU string `json:"ru"`
}

// Date is a calendar date written as YYYY-MM-DD.
type Date struct{ time.Time }

const dateLayout = "2006-01-02"

func parseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("invalid date %q, want YYYY-MM-DD", s)
	}
	return Date{t}, nil
}

// String formats the date as YYYY-MM-DD.
func (d Date) String() string { return d.Format(dateLayout) }

// UnmarshalJSON accepts only "YYYY-MM-DD" strings.
func (d *Date) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("date must be a string: %w", err)
	}
	parsed, err := parseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalJSON writes the date as "YYYY-MM-DD".
func (d Date) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }

// Photo is an externally hosted image.
type Photo struct {
	URL      string `json:"url"`
	ThumbURL string `json:"thumbUrl,omitempty"`
	Alt      *LText `json:"alt,omitempty"`
	Caption  *LText `json:"caption,omitempty"`
	Width    *int   `json:"width,omitempty"`
	Height   *int   `json:"height,omitempty"`
}

// Link is a contact link on the main page and the CV header.
type Link struct {
	Kind  string `json:"kind"`
	Label LText  `json:"label"`
	URL   string `json:"url"`
}

// MainFile is the main page content (tables main_info and social_link).
type MainFile struct {
	FullName  LText   `json:"fullName"`
	Headline  LText   `json:"headline"`
	Location  LText   `json:"location"`
	Currently LText   `json:"currently"`
	Bio       LText   `json:"bio"`
	Photo     Photo   `json:"photo"`
	Interests []LText `json:"interests"`
	Links     []Link  `json:"links"`
}

// Experience is one job on the CV.
type Experience struct {
	ID          string  `json:"id"`
	Position    LText   `json:"position"`
	Company     string  `json:"company"`
	CompanyURL  string  `json:"companyUrl,omitempty"`
	Location    LText   `json:"location"`
	StartDate   *Date   `json:"startDate"`
	EndDate     *Date     `json:"endDate"` // nil = present
	Projects    []Project `json:"projects,omitempty"`
}

// Project is one description within a job, with its own highlights.
type Project struct {
	Description *LText  `json:"description,omitempty"`
	Highlights  []LText `json:"highlights,omitempty"`
}

// Education is one education entry on the CV.
type Education struct {
	ID          string `json:"id"`
	Institution LText  `json:"institution"`
	Degree      LText  `json:"degree"`
	Field       LText  `json:"field"`
	StartDate   *Date  `json:"startDate"`
	EndDate     *Date  `json:"endDate"`
	Description *LText `json:"description,omitempty"`
}

// SkillGroup is a skill category with non-localised items.
type SkillGroup struct {
	Category LText    `json:"category"`
	Items    []string `json:"items"`
}

// SpokenLanguage is a language and level on the CV.
type SpokenLanguage struct {
	Name  LText `json:"name"`
	Level LText `json:"level"`
}

// CVFile is the CV content (cv_* tables).
type CVFile struct {
	Summary    LText            `json:"summary"`
	Experience []Experience     `json:"experience"`
	Education  []Education      `json:"education"`
	Skills     []SkillGroup     `json:"skills"`
	Languages  []SpokenLanguage `json:"languages"`
}

// Visit is one trip to a place.
type Visit struct {
	Start *Date `json:"start"`
	End   *Date `json:"end"`
}

// Place is a visited city or location.
type Place struct {
	ID      string   `json:"id"`
	Country string   `json:"country"`
	Name    LText    `json:"name"`
	Lat     *float64 `json:"lat"`
	Lng     *float64 `json:"lng"`
	Story   LText    `json:"story"`
	Visits  []Visit  `json:"visits"`
	Photos  []Photo  `json:"photos,omitempty"`
}

// TravelFile is the travel content (tables place, visit and photo).
type TravelFile struct {
	Places []Place `json:"places"`
}

// All is the complete site content.
type All struct {
	Main   MainFile   `json:"main"`
	CV     CVFile     `json:"cv"`
	Travel TravelFile `json:"travel"`
}
