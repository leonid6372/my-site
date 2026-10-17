package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/leonid6372/my-site/backend/internal/i18n"
)

// Endpoint names a cached response.
type Endpoint string

// Cached endpoints.
const (
	EndpointMain   Endpoint = "main"
	EndpointCV     Endpoint = "cv"
	EndpointTravel Endpoint = "travel"
)

// Response is a serialised response body with its weak ETag.
type Response struct {
	Body []byte
	ETag string
}

type cacheKey struct {
	endpoint Endpoint
	lang     i18n.Lang
}

type cacheEntry struct {
	mu   sync.Mutex
	resp *Response
}

// Cache holds the serialised JSON per (endpoint, lang) (plan §4.3). Content
// only changes when the process restarts, so entries never expire. Each
// entry is built lazily on first use; concurrent first requests wait for one
// build. Failed builds are not cached, so a short DB outage does not stick.
type Cache struct {
	svc     *Service
	mu      sync.Mutex
	entries map[cacheKey]*cacheEntry
}

// NewCache returns an empty cache in front of svc.
func NewCache(svc *Service) *Cache {
	return &Cache{svc: svc, entries: map[cacheKey]*cacheEntry{}}
}

// Get returns the cached response for (ep, lang), building it on first use.
func (c *Cache) Get(ctx context.Context, ep Endpoint, lang i18n.Lang) (Response, error) {
	key := cacheKey{ep, lang}
	c.mu.Lock()
	e, ok := c.entries[key]
	if !ok {
		e = &cacheEntry{}
		c.entries[key] = e
	}
	c.mu.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()
	if e.resp != nil {
		return *e.resp, nil
	}
	resp, err := c.build(ctx, ep, lang)
	if err != nil {
		return Response{}, err
	}
	e.resp = &resp
	return resp, nil
}

func (c *Cache) build(ctx context.Context, ep Endpoint, lang i18n.Lang) (Response, error) {
	var (
		dto any
		err error
	)
	switch ep {
	case EndpointMain:
		dto, err = c.svc.Main(ctx, lang)
	case EndpointCV:
		dto, err = c.svc.CV(ctx, lang)
	case EndpointTravel:
		dto, err = c.svc.Travel(ctx, lang)
	default:
		return Response{}, fmt.Errorf("unknown endpoint %q", ep)
	}
	if err != nil {
		return Response{}, err
	}
	body, err := json.Marshal(dto)
	if err != nil {
		return Response{}, fmt.Errorf("encode %s: %w", ep, err)
	}
	return Response{Body: body, ETag: etag(body)}, nil
}

// etag returns W/"<first 16 hex chars of sha256(body)>".
func etag(body []byte) string {
	sum := sha256.Sum256(body)
	return `W/"` + hex.EncodeToString(sum[:])[:16] + `"`
}
