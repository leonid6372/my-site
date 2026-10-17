// Package i18n negotiates the response language from the Accept-Language
// request header (contract §1.1, plan §4.2).
package i18n

import (
	"context"
	"net/http"

	"golang.org/x/text/language"
)

// Lang is a supported content language.
type Lang string

// Supported languages. EN is the fallback.
const (
	EN Lang = "en"
	RU Lang = "ru"
)

var supported = map[string]Lang{"en": EN, "ru": RU}

// Negotiate picks the supported language with the highest q-value from an
// Accept-Language header, matching on the primary subtag (ru-RU → ru).
// Ties keep header order. An empty, malformed or unsupported header yields EN.
func Negotiate(header string) Lang {
	if header == "" {
		return EN
	}
	tags, qs, err := language.ParseAcceptLanguage(header)
	if err != nil {
		return EN
	}
	best, bestQ := EN, float32(0)
	for i, tag := range tags {
		base, _ := tag.Base()
		lang, ok := supported[base.String()]
		if ok && qs[i] > bestQ {
			best, bestQ = lang, qs[i]
		}
	}
	return best
}

type ctxKey struct{}

// WithLang returns a copy of ctx carrying lang.
func WithLang(ctx context.Context, lang Lang) context.Context {
	return context.WithValue(ctx, ctxKey{}, lang)
}

// FromContext returns the negotiated language, or EN if none was set.
func FromContext(ctx context.Context) Lang {
	if lang, ok := ctx.Value(ctxKey{}).(Lang); ok {
		return lang
	}
	return EN
}

// Middleware negotiates the language for localised routes, stores it in the
// request context and sets the Content-Language and Vary response headers.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := Negotiate(r.Header.Get("Accept-Language"))
		w.Header().Set("Content-Language", string(lang))
		w.Header().Add("Vary", "Accept-Language")
		next.ServeHTTP(w, r.WithContext(WithLang(r.Context(), lang)))
	})
}
