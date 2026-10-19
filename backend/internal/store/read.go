package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/leonid6372/my-site/backend/internal/content"
)

// ErrNoContent is returned by GetMain and GetCV when the database holds no
// content yet. Data is added to the database directly; the API never writes it.
var ErrNoContent = errors.New("database has no content")

// GetMain returns the main page info. Links are read separately (ListLinks).
func (s *Store) GetMain(ctx context.Context) (content.MainFile, error) {
	var (
		m        content.MainFile
		thumbURL *string
	)
	err := s.pool.QueryRow(ctx, `SELECT full_name, headline, location, currently, bio,
		photo_url, photo_thumb_url, photo_alt, photo_width, photo_height, interests
		FROM main_info WHERE id = 1`).Scan(
		&m.FullName, &m.Headline, &m.Location, &m.Currently, &m.Bio,
		&m.Photo.URL, &thumbURL, &m.Photo.Alt, &m.Photo.Width, &m.Photo.Height, &m.Interests)
	if errors.Is(err, pgx.ErrNoRows) {
		return content.MainFile{}, ErrNoContent
	}
	if err != nil {
		return content.MainFile{}, fmt.Errorf("get main: %w", err)
	}
	m.Photo.ThumbURL = deref(thumbURL)
	m.Interests = nilIfEmpty(m.Interests)
	return m, nil
}

// ListLinks returns the contact links in content order.
func (s *Store) ListLinks(ctx context.Context) ([]content.Link, error) {
	rows, _ := s.pool.Query(ctx, `SELECT kind, label, url FROM social_link ORDER BY sort_order`)
	links, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (content.Link, error) {
		var l content.Link
		err := r.Scan(&l.Kind, &l.Label, &l.URL)
		return l, err
	})
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	return nilIfEmpty(links), nil
}

// GetCV returns the CV. Experience and education are ordered by start date,
// newest first; skills and languages keep content order.
func (s *Store) GetCV(ctx context.Context) (content.CVFile, error) {
	var cv content.CVFile
	err := s.pool.QueryRow(ctx, `SELECT summary FROM main_info WHERE id = 1`).Scan(&cv.Summary)
	if errors.Is(err, pgx.ErrNoRows) {
		return content.CVFile{}, ErrNoContent
	}
	if err != nil {
		return content.CVFile{}, fmt.Errorf("get cv summary: %w", err)
	}

	rows, _ := s.pool.Query(ctx, `SELECT id, position, company, company_url, location,
		start_date, end_date, description, highlights
		FROM cv_experience ORDER BY start_date DESC, id`)
	cv.Experience, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (content.Experience, error) {
		var (
			e          content.Experience
			companyURL *string
			start, end *time.Time
		)
		err := r.Scan(&e.ID, &e.Position, &e.Company, &companyURL, &e.Location,
			&start, &end, &e.Description, &e.Highlights)
		e.CompanyURL, e.StartDate, e.EndDate = deref(companyURL), toDate(start), toDate(end)
		e.Highlights = nilIfEmpty(e.Highlights)
		return e, err
	})
	if err != nil {
		return content.CVFile{}, fmt.Errorf("list experience: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `SELECT id, institution, degree, field, start_date, end_date, description
		FROM cv_education ORDER BY start_date DESC, id`)
	cv.Education, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (content.Education, error) {
		var (
			e          content.Education
			start, end *time.Time
		)
		err := r.Scan(&e.ID, &e.Institution, &e.Degree, &e.Field, &start, &end, &e.Description)
		e.StartDate, e.EndDate = toDate(start), toDate(end)
		return e, err
	})
	if err != nil {
		return content.CVFile{}, fmt.Errorf("list education: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `SELECT category, items FROM cv_skill_group ORDER BY sort_order`)
	cv.Skills, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (content.SkillGroup, error) {
		var g content.SkillGroup
		err := r.Scan(&g.Category, &g.Items)
		return g, err
	})
	if err != nil {
		return content.CVFile{}, fmt.Errorf("list skills: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `SELECT name, level FROM cv_language ORDER BY sort_order`)
	cv.Languages, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (content.SpokenLanguage, error) {
		var l content.SpokenLanguage
		err := r.Scan(&l.Name, &l.Level)
		return l, err
	})
	if err != nil {
		return content.CVFile{}, fmt.Errorf("list languages: %w", err)
	}

	cv.Experience, cv.Education = nilIfEmpty(cv.Experience), nilIfEmpty(cv.Education)
	cv.Skills, cv.Languages = nilIfEmpty(cv.Skills), nilIfEmpty(cv.Languages)
	return cv, nil
}

// ListPlaces returns all places ordered by id, each with its visits (in
// content order) and photos (in sort_order). It runs three queries and joins
// them in memory instead of one query per place.
func (s *Store) ListPlaces(ctx context.Context) ([]content.Place, error) {
	rows, _ := s.pool.Query(ctx, `SELECT id, country_iso2, name, lat, lng, story FROM place ORDER BY id`)
	places, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (content.Place, error) {
		var (
			p        content.Place
			lat, lng float64
		)
		err := r.Scan(&p.ID, &p.Country, &p.Name, &lat, &lng, &p.Story)
		p.Lat, p.Lng = &lat, &lng
		return p, err
	})
	if err != nil {
		return nil, fmt.Errorf("list places: %w", err)
	}
	byID := make(map[string]*content.Place, len(places))
	for i := range places {
		byID[places[i].ID] = &places[i]
	}

	rows, _ = s.pool.Query(ctx, `SELECT place_id, start_date, end_date FROM visit ORDER BY id`)
	var (
		placeID    string
		start, end time.Time
	)
	_, err = pgx.ForEachRow(rows, []any{&placeID, &start, &end}, func() error {
		p := byID[placeID]
		p.Visits = append(p.Visits, content.Visit{Start: &content.Date{Time: start}, End: &content.Date{Time: end}})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list visits: %w", err)
	}

	rows, _ = s.pool.Query(ctx, `SELECT place_id, url, thumb_url, alt, caption, width, height
		FROM photo ORDER BY place_id, sort_order`)
	var (
		ph       content.Photo
		thumbURL *string
	)
	_, err = pgx.ForEachRow(rows, []any{&placeID, &ph.URL, &thumbURL, &ph.Alt, &ph.Caption, &ph.Width, &ph.Height},
		func() error {
			ph.ThumbURL = deref(thumbURL)
			p := byID[placeID]
			p.Photos = append(p.Photos, ph)
			// Reset the scan target so pointer fields are never shared between photos.
			ph, thumbURL = content.Photo{}, nil
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list photos: %w", err)
	}
	return nilIfEmpty(places), nil
}

func toDate(t *time.Time) *content.Date {
	if t == nil {
		return nil
	}
	return &content.Date{Time: *t}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nilIfEmpty normalises empty lists to nil, matching content loaded from a
// file where the key is absent.
func nilIfEmpty[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return s
}
