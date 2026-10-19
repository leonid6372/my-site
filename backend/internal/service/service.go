// Package service builds the localised API responses (specs/api/contract.md)
// from the content read by the store.
package service

import (
	"cmp"
	"context"
	"regexp"
	"slices"
	"strings"

	"github.com/leonid6372/my-site/backend/internal/content"
	"github.com/leonid6372/my-site/backend/internal/i18n"
)

// Store is the subset of store.Store the service reads from.
type Store interface {
	GetMain(ctx context.Context) (content.MainFile, error)
	ListLinks(ctx context.Context) ([]content.Link, error)
	GetCV(ctx context.Context) (content.CVFile, error)
	ListPlaces(ctx context.Context) ([]content.Place, error)
}

// Service builds response DTOs.
type Service struct {
	store Store
}

// New returns a Service reading from s.
func New(s Store) *Service { return &Service{store: s} }

// Main builds the /main response in lang.
func (s *Service) Main(ctx context.Context, lang i18n.Lang) (MainDTO, error) {
	m, err := s.store.GetMain(ctx)
	if err != nil {
		return MainDTO{}, err
	}
	links, err := s.store.ListLinks(ctx)
	if err != nil {
		return MainDTO{}, err
	}

	dto := MainDTO{
		FullName:  loc(m.FullName, lang),
		Headline:  loc(m.Headline, lang),
		Location:  loc(m.Location, lang),
		Currently: loc(m.Currently, lang),
		Bio:       paragraphs(loc(m.Bio, lang)),
		Photo:     photo(m.Photo, lang),
		Interests: locAll(m.Interests, lang),
		Links:     make([]LinkDTO, 0, len(links)),
	}
	for _, l := range links {
		dto.Links = append(dto.Links, LinkDTO{Kind: l.Kind, Label: loc(l.Label, lang), URL: l.URL})
	}
	return dto, nil
}

// CV builds the /cv response in lang. Experience and education are sorted by
// start date, newest first; skills and languages keep content order.
func (s *Service) CV(ctx context.Context, lang i18n.Lang) (CVDTO, error) {
	cv, err := s.store.GetCV(ctx)
	if err != nil {
		return CVDTO{}, err
	}

	dto := CVDTO{
		Summary:    loc(cv.Summary, lang),
		Experience: make([]ExperienceDTO, 0, len(cv.Experience)),
		Education:  make([]EducationDTO, 0, len(cv.Education)),
		Skills:     make([]SkillGroupDTO, 0, len(cv.Skills)),
		Languages:  make([]LanguageDTO, 0, len(cv.Languages)),
	}
	for _, e := range cv.Experience {
		dto.Experience = append(dto.Experience, ExperienceDTO{
			ID:          e.ID,
			Position:    loc(e.Position, lang),
			Company:     e.Company,
			CompanyURL:  optString(e.CompanyURL),
			Location:    loc(e.Location, lang),
			StartDate:   e.StartDate.String(),
			EndDate:     optDate(e.EndDate),
			Projects:    projects(e.Projects, lang),
		})
	}
	slices.SortStableFunc(dto.Experience, func(a, b ExperienceDTO) int {
		return byStartDesc(a.StartDate, b.StartDate, a.ID, b.ID)
	})
	for _, e := range cv.Education {
		dto.Education = append(dto.Education, EducationDTO{
			ID:          e.ID,
			Institution: loc(e.Institution, lang),
			Degree:      loc(e.Degree, lang),
			Field:       loc(e.Field, lang),
			StartDate:   e.StartDate.String(),
			EndDate:     optDate(e.EndDate),
			Description: optLoc(e.Description, lang),
		})
	}
	slices.SortStableFunc(dto.Education, func(a, b EducationDTO) int {
		return byStartDesc(a.StartDate, b.StartDate, a.ID, b.ID)
	})
	for _, g := range cv.Skills {
		dto.Skills = append(dto.Skills, SkillGroupDTO{Category: loc(g.Category, lang), Items: nonNil(g.Items)})
	}
	for _, l := range cv.Languages {
		dto.Languages = append(dto.Languages, LanguageDTO{Name: loc(l.Name, lang), Level: loc(l.Level, lang)})
	}
	return dto, nil
}

// byStartDesc orders YYYY-MM-DD dates newest first (they sort as strings),
// then by id for a deterministic order.
func byStartDesc(aDate, bDate, aID, bID string) int {
	return cmp.Or(cmp.Compare(bDate, aDate), cmp.Compare(aID, bID))
}

// loc resolves a localised text to lang.
func loc(t content.LText, lang i18n.Lang) string {
	if lang == i18n.RU {
		return t.RU
	}
	return t.EN
}

func optLoc(t *content.LText, lang i18n.Lang) *string {
	if t == nil {
		return nil
	}
	s := loc(*t, lang)
	return &s
}

func locAll(ts []content.LText, lang i18n.Lang) []string {
	out := make([]string, 0, len(ts))
	for _, t := range ts {
		out = append(out, loc(t, lang))
	}
	return out
}

func projects(ps []content.Project, lang i18n.Lang) []ProjectDTO {
	out := make([]ProjectDTO, 0, len(ps))
	for _, p := range ps {
		out = append(out, ProjectDTO{
			Description: optLoc(p.Description, lang),
			Highlights:  locAll(p.Highlights, lang),
		})
	}
	return out
}

// photo passes URLs and size through, localises alt and caption, and maps
// missing optional values to null.
func photo(p content.Photo, lang i18n.Lang) PhotoDTO {
	return PhotoDTO{
		URL:      p.URL,
		ThumbURL: optString(p.ThumbURL),
		Alt:      optLoc(p.Alt, lang),
		Caption:  optLoc(p.Caption, lang),
		Width:    p.Width,
		Height:   p.Height,
	}
}

func optString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optDate(d *content.Date) *string {
	if d == nil {
		return nil
	}
	s := d.String()
	return &s
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

var blankLines = regexp.MustCompile(`\n[ \t]*\n\s*`)

// paragraphs splits text on blank lines. Line breaks inside a paragraph
// (wrapped lines) become spaces.
func paragraphs(text string) []string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n"))
	out := []string{}
	if text == "" {
		return out
	}
	for _, p := range blankLines.Split(text, -1) {
		out = append(out, strings.Join(strings.Fields(p), " "))
	}
	return out
}
