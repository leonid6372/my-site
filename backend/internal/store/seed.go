package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/leonid6372/my-site/backend/internal/content"
)

// LText and *LText values are passed to pgx as-is: its JSON codec encodes
// them as JSONB {"en": ..., "ru": ...} and nil pointers as NULL.

// ReplaceAll replaces all content in one transaction: on any error the
// previous content is left untouched.
func (s *Store) ReplaceAll(ctx context.Context, all content.All) (err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback(ctx))
		}
	}()

	if _, err = tx.Exec(ctx, `TRUNCATE main_info, social_link, cv_experience, cv_education,
		cv_skill_group, cv_language, place, visit, photo RESTART IDENTITY CASCADE`); err != nil {
		return fmt.Errorf("truncate: %w", err)
	}

	b := &pgx.Batch{}
	queueMain(b, all.Main, all.CV.Summary)
	queueCV(b, all.CV)
	queueTravel(b, all.Travel)
	if err = tx.SendBatch(ctx, b).Close(); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func queueMain(b *pgx.Batch, m content.MainFile, summary content.LText) {
	b.Queue(`INSERT INTO main_info (full_name, headline, location, currently, bio, summary,
		photo_url, photo_thumb_url, photo_alt, photo_width, photo_height, interests)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		m.FullName, m.Headline, m.Location, m.Currently, m.Bio, summary,
		m.Photo.URL, nullString(m.Photo.ThumbURL), m.Photo.Alt, m.Photo.Width, m.Photo.Height,
		nonNil(m.Interests))
	for i, l := range m.Links {
		b.Queue(`INSERT INTO social_link (kind, label, url, sort_order) VALUES ($1, $2, $3, $4)`,
			l.Kind, l.Label, l.URL, i)
	}
}

func queueCV(b *pgx.Batch, cv content.CVFile) {
	for _, e := range cv.Experience {
		b.Queue(`INSERT INTO cv_experience (id, position, company, company_url, location,
			start_date, end_date, description, highlights)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			e.ID, e.Position, e.Company, nullString(e.CompanyURL), e.Location,
			date(e.StartDate), date(e.EndDate), e.Description, nonNil(e.Highlights))
	}
	for _, e := range cv.Education {
		b.Queue(`INSERT INTO cv_education (id, institution, degree, field, start_date, end_date, description)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			e.ID, e.Institution, e.Degree, e.Field, date(e.StartDate), date(e.EndDate), e.Description)
	}
	for i, g := range cv.Skills {
		b.Queue(`INSERT INTO cv_skill_group (category, items, sort_order) VALUES ($1, $2, $3)`,
			g.Category, g.Items, i)
	}
	for i, l := range cv.Languages {
		b.Queue(`INSERT INTO cv_language (name, level, sort_order) VALUES ($1, $2, $3)`,
			l.Name, l.Level, i)
	}
}

func queueTravel(b *pgx.Batch, t content.TravelFile) {
	for _, p := range t.Places {
		b.Queue(`INSERT INTO place (id, country_iso2, name, lat, lng, story)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			p.ID, p.Country, p.Name, p.Lat, p.Lng, p.Story)
		for _, v := range p.Visits {
			b.Queue(`INSERT INTO visit (place_id, start_date, end_date) VALUES ($1, $2, $3)`,
				p.ID, date(v.Start), date(v.End))
		}
		for i, ph := range p.Photos {
			b.Queue(`INSERT INTO photo (place_id, url, thumb_url, alt, caption, width, height, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				p.ID, ph.URL, nullString(ph.ThumbURL), ph.Alt, ph.Caption, ph.Width, ph.Height, i)
		}
	}
}

// date converts an optional content date to a DATE parameter (nil = NULL).
func date(d *content.Date) *time.Time {
	if d == nil {
		return nil
	}
	return &d.Time
}

// nullString maps an empty optional string to NULL.
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nonNil makes an absent list encode as JSON [] instead of null.
func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
