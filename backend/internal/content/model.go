// Package content defines the structure of the content files (spec §5),
// loads them from YAML or JSON and validates them (spec §6).
package content

import (
	"encoding/json"
	"fmt"
	"time"

	"go.yaml.in/yaml/v3"
)

// LText is a localised text; both languages are required (constitution §5.3).
type LText struct {
	EN string `yaml:"en" json:"en"`
	RU string `yaml:"ru" json:"ru"`
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

// UnmarshalYAML accepts only plain YYYY-MM-DD dates.
func (d *Date) UnmarshalYAML(n *yaml.Node) error {
	parsed, err := parseDate(n.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w", n.Line, err)
	}
	*d = parsed
	return nil
}

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

// Photo is an externally hosted image (spec §5.1).
type Photo struct {
	URL      string `yaml:"url" json:"url"`
	ThumbURL string `yaml:"thumbUrl,omitempty" json:"thumbUrl,omitempty"`
	Alt      *LText `yaml:"alt,omitempty" json:"alt,omitempty"`
	Caption  *LText `yaml:"caption,omitempty" json:"caption,omitempty"`
	Width    *int   `yaml:"width,omitempty" json:"width,omitempty"`
	Height   *int   `yaml:"height,omitempty" json:"height,omitempty"`
}

// Link is a contact link on the main page and the CV header.
type Link struct {
	Kind  string `yaml:"kind" json:"kind"`
	Label LText  `yaml:"label" json:"label"`
	URL   string `yaml:"url" json:"url"`
}

// MainFile is main.yaml (spec §5.2).
type MainFile struct {
	FullName  LText   `yaml:"fullName" json:"fullName"`
	Headline  LText   `yaml:"headline" json:"headline"`
	Location  LText   `yaml:"location" json:"location"`
	Currently LText   `yaml:"currently" json:"currently"`
	Bio       LText   `yaml:"bio" json:"bio"`
	Photo     Photo   `yaml:"photo" json:"photo"`
	Interests []LText `yaml:"interests" json:"interests"`
	Links     []Link  `yaml:"links" json:"links"`
}

// Experience is one job in cv.yaml.
type Experience struct {
	ID          string  `yaml:"id" json:"id"`
	Position    LText   `yaml:"position" json:"position"`
	Company     string  `yaml:"company" json:"company"`
	CompanyURL  string  `yaml:"companyUrl,omitempty" json:"companyUrl,omitempty"`
	Location    LText   `yaml:"location" json:"location"`
	StartDate   *Date   `yaml:"startDate" json:"startDate"`
	EndDate     *Date   `yaml:"endDate" json:"endDate"` // nil = present
	Description *LText  `yaml:"description,omitempty" json:"description,omitempty"`
	Highlights  []LText `yaml:"highlights,omitempty" json:"highlights,omitempty"`
}

// Education is one education entry in cv.yaml.
type Education struct {
	ID          string `yaml:"id" json:"id"`
	Institution LText  `yaml:"institution" json:"institution"`
	Degree      LText  `yaml:"degree" json:"degree"`
	Field       LText  `yaml:"field" json:"field"`
	StartDate   *Date  `yaml:"startDate" json:"startDate"`
	EndDate     *Date  `yaml:"endDate" json:"endDate"`
	Description *LText `yaml:"description,omitempty" json:"description,omitempty"`
}

// SkillGroup is a skill category with non-localised items.
type SkillGroup struct {
	Category LText    `yaml:"category" json:"category"`
	Items    []string `yaml:"items" json:"items"`
}

// SpokenLanguage is a language and level on the CV.
type SpokenLanguage struct {
	Name  LText `yaml:"name" json:"name"`
	Level LText `yaml:"level" json:"level"`
}

// CVFile is cv.yaml (spec §5.3).
type CVFile struct {
	Summary    LText            `yaml:"summary" json:"summary"`
	Experience []Experience     `yaml:"experience" json:"experience"`
	Education  []Education      `yaml:"education" json:"education"`
	Skills     []SkillGroup     `yaml:"skills" json:"skills"`
	Languages  []SpokenLanguage `yaml:"languages" json:"languages"`
}

// Visit is one trip to a place.
type Visit struct {
	Start *Date `yaml:"start" json:"start"`
	End   *Date `yaml:"end" json:"end"`
}

// Place is a visited city or location in travel.yaml.
type Place struct {
	ID      string   `yaml:"id" json:"id"`
	Country string   `yaml:"country" json:"country"`
	Name    LText    `yaml:"name" json:"name"`
	Lat     *float64 `yaml:"lat" json:"lat"`
	Lng     *float64 `yaml:"lng" json:"lng"`
	Story   LText    `yaml:"story" json:"story"`
	Visits  []Visit  `yaml:"visits" json:"visits"`
	Photos  []Photo  `yaml:"photos,omitempty" json:"photos,omitempty"`
}

// TravelFile is travel.yaml (spec §5.4).
type TravelFile struct {
	Places []Place `yaml:"places" json:"places"`
}

// All is the complete site content.
type All struct {
	Main   MainFile
	CV     CVFile
	Travel TravelFile

	// Files maps a logical name ("main", "cv", "travel") to the file name it
	// was loaded from, e.g. "travel.yaml". Used in validation messages.
	Files map[string]string
}
