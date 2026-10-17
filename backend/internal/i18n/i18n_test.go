package i18n

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNegotiate(t *testing.T) {
	tests := []struct {
		header string
		want   Lang
	}{
		{"", EN},
		{"en", EN},
		{"ru", RU},
		{"ru-RU,ru;q=0.9,en;q=0.8", RU},
		{"en;q=0.5, ru;q=0.8", RU},
		{"de-DE", EN},
		{"de-DE,ru;q=0.3", RU},
		{"en, ru", EN},             // tie → header order
		{"ru;q=0.8, en;q=0.8", RU}, // tie → header order
		{"*", EN},
		{"ru;q=0", EN},        // q=0 means "not acceptable"
		{"ru;q=abc,,,;;", EN}, // malformed
		{"EN-gb", EN},
	}
	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			if got := Negotiate(tt.header); got != tt.want {
				t.Errorf("Negotiate(%q) = %q, want %q", tt.header, got, tt.want)
			}
		})
	}
}

func TestFromContextDefault(t *testing.T) {
	if got := FromContext(context.Background()); got != EN {
		t.Errorf("got %q, want en", got)
	}
}

func TestMiddleware(t *testing.T) {
	var seen Lang
	h := Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = FromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/main", nil)
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if seen != RU {
		t.Errorf("context lang = %q, want ru", seen)
	}
	if got := rec.Header().Get("Content-Language"); got != "ru" {
		t.Errorf("Content-Language = %q, want ru", got)
	}
	if got := rec.Header().Get("Vary"); got != "Accept-Language" {
		t.Errorf("Vary = %q, want Accept-Language", got)
	}
}
