//go:build integration

package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/leonid6372/my-site/backend/internal/content"
)

func loadFixture(t *testing.T) content.All {
	t.Helper()
	all, err := content.Load(filepath.Join("..", "content", "testdata", "valid"))
	if err != nil {
		t.Fatal(err)
	}
	if errs := content.Validate(all); len(errs) > 0 {
		t.Fatalf("fixture is invalid: %v", errs)
	}
	return all
}

// expectedCounts derives the row count of every table from the content.
func expectedCounts(all content.All) map[string]int {
	var visits, photos int
	for _, p := range all.Travel.Places {
		visits += len(p.Visits)
		photos += len(p.Photos)
	}
	return map[string]int{
		"main_info":      1,
		"social_link":    len(all.Main.Links),
		"cv_experience":  len(all.CV.Experience),
		"cv_education":   len(all.CV.Education),
		"cv_skill_group": len(all.CV.Skills),
		"cv_language":    len(all.CV.Languages),
		"place":          len(all.Travel.Places),
		"visit":          visits,
		"photo":          photos,
	}
}

func assertCounts(t *testing.T, s *Store, want map[string]int) {
	t.Helper()
	for table, n := range want {
		var got int
		// table names come from the fixed map above, not from input.
		if err := s.pool.QueryRow(context.Background(), "SELECT count(*) FROM "+table).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != n {
			t.Errorf("%s: %d rows, want %d", table, got, n)
		}
	}
}

func TestReplaceAllIsIdempotent(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	all := loadFixture(t)
	want := expectedCounts(all)
	if want["visit"] == 0 || want["photo"] == 0 {
		t.Fatal("fixture should contain visits and photos")
	}

	for i := range 2 {
		if err := s.ReplaceAll(ctx, all); err != nil {
			t.Fatalf("seed #%d: %v", i+1, err)
		}
		assertCounts(t, s, want)
	}

	// Identities restart, so sort_order and ids stay stable across seeds.
	var minID int
	if err := s.pool.QueryRow(ctx, "SELECT min(id) FROM social_link").Scan(&minID); err != nil {
		t.Fatal(err)
	}
	if minID != 1 {
		t.Errorf("social_link min(id) = %d, want 1 (RESTART IDENTITY)", minID)
	}
}

func TestReplaceAllKeepsPreviousDataOnError(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	all := loadFixture(t)
	if err := s.ReplaceAll(ctx, all); err != nil {
		t.Fatal(err)
	}
	want := expectedCounts(all)

	// Break the last row inserted (a photo with width but no height violates a
	// CHECK constraint), so the failure happens after everything else was queued.
	broken := loadFixture(t)
	broken.Main.Links = broken.Main.Links[:1]
	places := broken.Travel.Places
	last := &places[len(places)-1]
	last.Photos = append(last.Photos, content.Photo{URL: "https://example.com/x.jpg", Width: new(10)})

	if err := s.ReplaceAll(ctx, broken); err == nil {
		t.Fatal("expected an insert error")
	}
	assertCounts(t, s, want)
}
