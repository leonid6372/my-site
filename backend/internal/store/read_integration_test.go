//go:build integration

package store

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/leonid6372/my-site/backend/internal/content"
)

// assertEqual compares with reflect.DeepEqual and prints both sides as JSON.
func assertEqual(t *testing.T, name string, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		g, _ := json.MarshalIndent(got, "", "  ")
		w, _ := json.MarshalIndent(want, "", "  ")
		t.Errorf("%s mismatch\n got: %s\nwant: %s", name, g, w)
	}
}

func TestReadRoundTrip(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	want := loadFixture(t)
	if err := s.ReplaceAll(ctx, want); err != nil {
		t.Fatal(err)
	}

	gotMain, err := s.GetMain(ctx)
	if err != nil {
		t.Fatal(err)
	}
	gotMain.Links, err = s.ListLinks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, "main", gotMain, want.Main)

	gotCV, err := s.GetCV(ctx)
	if err != nil {
		t.Fatal(err)
	}
	byStartDesc := func(a, b *content.Date) int { return b.Compare(a.Time) }
	slices.SortStableFunc(want.CV.Experience, func(a, b content.Experience) int {
		return byStartDesc(a.StartDate, b.StartDate)
	})
	slices.SortStableFunc(want.CV.Education, func(a, b content.Education) int {
		return byStartDesc(a.StartDate, b.StartDate)
	})
	assertEqual(t, "cv", gotCV, want.CV)

	gotPlaces, err := s.ListPlaces(ctx)
	if err != nil {
		t.Fatal(err)
	}
	slices.SortFunc(want.Travel.Places, func(a, b content.Place) int { return cmp.Compare(a.ID, b.ID) })
	assertEqual(t, "places", gotPlaces, want.Travel.Places)
}

func TestReadEmptyDatabase(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	if _, err := s.pool.Exec(ctx, `TRUNCATE main_info, social_link, cv_experience, cv_education,
		cv_skill_group, cv_language, place, visit, photo`); err != nil {
		t.Fatal(err)
	}

	if _, err := s.GetMain(ctx); !errors.Is(err, ErrNotSeeded) {
		t.Errorf("GetMain err = %v, want ErrNotSeeded", err)
	}
	if _, err := s.GetCV(ctx); !errors.Is(err, ErrNotSeeded) {
		t.Errorf("GetCV err = %v, want ErrNotSeeded", err)
	}
	if places, err := s.ListPlaces(ctx); err != nil || places != nil {
		t.Errorf("ListPlaces = %v, %v; want nil, nil", places, err)
	}
}
