package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/leonid6372/my-site/backend/internal/content"
	"github.com/leonid6372/my-site/backend/internal/i18n"
)

// countingStore counts the first store call of every endpoint build and can
// fail the next n calls.
type countingStore struct {
	inner                  Store
	main, cv, travel, fail atomic.Int32
}

var errFake = errors.New("db down")

func (s *countingStore) failing() bool { return s.fail.Add(-1) >= 0 }

func (s *countingStore) GetMain(ctx context.Context) (content.MainFile, error) {
	s.main.Add(1)
	if s.failing() {
		return content.MainFile{}, errFake
	}
	return s.inner.GetMain(ctx)
}

func (s *countingStore) ListLinks(ctx context.Context) ([]content.Link, error) {
	return s.inner.ListLinks(ctx)
}

func (s *countingStore) GetCV(ctx context.Context) (content.CVFile, error) {
	s.cv.Add(1)
	return s.inner.GetCV(ctx)
}

func (s *countingStore) ListPlaces(ctx context.Context) ([]content.Place, error) {
	s.travel.Add(1)
	return s.inner.ListPlaces(ctx)
}

// Every call must go to the store: nothing is cached.
func TestRenderQueriesStoreEveryTime(t *testing.T) {
	store := &countingStore{inner: fixtureStore(t)}
	svc := New(store)
	ctx := context.Background()

	const calls = 5
	for _, ep := range []Endpoint{EndpointMain, EndpointCV, EndpointTravel} {
		for _, lang := range []i18n.Lang{i18n.EN, i18n.RU} {
			for range calls {
				if _, err := svc.Render(ctx, ep, lang); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	for name, n := range map[string]int32{"main": store.main.Load(), "cv": store.cv.Load(), "travel": store.travel.Load()} {
		if n != 2*calls {
			t.Errorf("%s: %d store calls, want %d", name, n, 2*calls)
		}
	}
}

func TestRenderBody(t *testing.T) {
	body, err := New(fixtureStore(t)).Render(context.Background(), EndpointMain, i18n.RU)
	if err != nil {
		t.Fatal(err)
	}
	var m MainDTO
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	if m.FullName != "Леонид Соловьёв" {
		t.Errorf("body fullName = %q, want the RU value", m.FullName)
	}
}

func TestRenderRecoversAfterError(t *testing.T) {
	store := &countingStore{inner: fixtureStore(t)}
	store.fail.Store(1)
	svc := New(store)
	ctx := context.Background()

	if _, err := svc.Render(ctx, EndpointMain, i18n.EN); !errors.Is(err, errFake) {
		t.Fatalf("first Render err = %v, want the store error", err)
	}
	if _, err := svc.Render(ctx, EndpointMain, i18n.EN); err != nil {
		t.Fatalf("second Render err = %v, want success", err)
	}
}

func TestRenderUnknownEndpoint(t *testing.T) {
	if _, err := New(fixtureStore(t)).Render(context.Background(), "nope", i18n.EN); err == nil {
		t.Fatal("expected an error")
	}
}
