package service

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"sync"
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

func TestCacheBuildsOncePerKey(t *testing.T) {
	store := &countingStore{inner: fixtureStore(t)}
	cache := NewCache(New(store))
	ctx := context.Background()

	etags := map[cacheKey]string{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, ep := range []Endpoint{EndpointMain, EndpointCV, EndpointTravel} {
		for _, lang := range []i18n.Lang{i18n.EN, i18n.RU} {
			for range 100 {
				wg.Go(func() {
					resp, err := cache.Get(ctx, ep, lang)
					if err != nil {
						t.Error(err)
						return
					}
					mu.Lock()
					defer mu.Unlock()
					key := cacheKey{ep, lang}
					if prev, ok := etags[key]; ok && prev != resp.ETag {
						t.Errorf("%v: ETag changed from %s to %s", key, prev, resp.ETag)
					}
					etags[key] = resp.ETag
				})
			}
		}
	}
	wg.Wait()

	// One build per language for each endpoint.
	for name, n := range map[string]int32{"main": store.main.Load(), "cv": store.cv.Load(), "travel": store.travel.Load()} {
		if n != 2 {
			t.Errorf("%s: %d store calls for 200 requests, want 2 (en + ru)", name, n)
		}
	}
	for _, ep := range []Endpoint{EndpointMain, EndpointCV, EndpointTravel} {
		if etags[cacheKey{ep, i18n.EN}] == etags[cacheKey{ep, i18n.RU}] {
			t.Errorf("%s: EN and RU ETags are equal", ep)
		}
	}
}

func TestCacheResponse(t *testing.T) {
	cache := NewCache(New(fixtureStore(t)))
	resp, err := cache.Get(context.Background(), EndpointMain, i18n.RU)
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^W/"[0-9a-f]{16}"$`).MatchString(resp.ETag) {
		t.Errorf("ETag = %s, want W/\"<16 hex>\"", resp.ETag)
	}
	var m MainDTO
	if err := json.Unmarshal(resp.Body, &m); err != nil {
		t.Fatal(err)
	}
	if m.FullName != "Леонид Соловьёв" {
		t.Errorf("body fullName = %q, want the RU value", m.FullName)
	}
}

func TestCacheDoesNotKeepErrors(t *testing.T) {
	store := &countingStore{inner: fixtureStore(t)}
	store.fail.Store(1)
	cache := NewCache(New(store))
	ctx := context.Background()

	if _, err := cache.Get(ctx, EndpointMain, i18n.EN); !errors.Is(err, errFake) {
		t.Fatalf("first Get err = %v, want the store error", err)
	}
	if _, err := cache.Get(ctx, EndpointMain, i18n.EN); err != nil {
		t.Fatalf("second Get err = %v, want a fresh successful build", err)
	}
	if _, err := cache.Get(ctx, EndpointMain, i18n.EN); err != nil {
		t.Fatal(err)
	}
	if n := store.main.Load(); n != 2 {
		t.Errorf("store calls = %d, want 2 (failed build + one cached build)", n)
	}
}

func TestCacheUnknownEndpoint(t *testing.T) {
	if _, err := NewCache(New(fixtureStore(t))).Get(context.Background(), "nope", i18n.EN); err == nil {
		t.Fatal("expected an error")
	}
}
