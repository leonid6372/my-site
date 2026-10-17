package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leonid6372/my-site/backend/internal/content"
	"github.com/leonid6372/my-site/backend/internal/service"
	"github.com/leonid6372/my-site/backend/internal/store"
)

type fakeStore struct {
	all content.All
	err error
}

func (f *fakeStore) GetMain(context.Context) (content.MainFile, error) { return f.all.Main, f.err }
func (f *fakeStore) ListLinks(context.Context) ([]content.Link, error) {
	return f.all.Main.Links, f.err
}
func (f *fakeStore) GetCV(context.Context) (content.CVFile, error) { return f.all.CV, f.err }
func (f *fakeStore) ListPlaces(context.Context) ([]content.Place, error) {
	return f.all.Travel.Places, f.err
}

type fakeDB struct{ err error }

func (f fakeDB) Ping(context.Context) error { return f.err }

func loadFixture(t *testing.T) content.All {
	t.Helper()
	all, err := content.Load(filepath.Join("..", "content", "testdata", "valid"))
	if err != nil {
		t.Fatal(err)
	}
	return all
}

func newHandler(s service.Store, db Pinger) http.Handler {
	return New(Options{
		Cache: service.NewCache(service.New(s)),
		DB:    db,
		Log:   slog.New(slog.DiscardHandler),
	})
}

func do(h http.Handler, method, path string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestLocalisedEndpoints(t *testing.T) {
	all := loadFixture(t)
	h := newHandler(&fakeStore{all: all}, fakeDB{})

	// One language-specific value per endpoint, read from the response body.
	endpoints := []struct {
		path  string
		text  content.LText
		field func(map[string]any) any
	}{
		{"/api/v1/main", all.Main.FullName, func(m map[string]any) any { return m["fullName"] }},
		{"/api/v1/cv", all.CV.Summary, func(m map[string]any) any { return m["summary"] }},
		{"/api/v1/travel", all.Travel.Places[0].Name, func(m map[string]any) any {
			for _, p := range m["places"].([]any) {
				if p := p.(map[string]any); p["id"] == all.Travel.Places[0].ID {
					return p["name"]
				}
			}
			return nil
		}},
	}
	headers := []struct {
		accept, wantLang string
	}{
		{"", "en"}, {"en", "en"}, {"ru", "ru"}, {"ru-RU,ru;q=0.9", "ru"}, {"de", "en"},
	}

	for _, ep := range endpoints {
		etags := map[string]string{}
		for _, hd := range headers {
			t.Run(fmt.Sprintf("%s/%q", ep.path, hd.accept), func(t *testing.T) {
				var rec *httptest.ResponseRecorder
				if hd.accept == "" {
					rec = do(h, http.MethodGet, ep.path)
				} else {
					rec = do(h, http.MethodGet, ep.path, "Accept-Language", hd.accept)
				}
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
				}
				hdr := rec.Header()
				if got := hdr.Get("Content-Language"); got != hd.wantLang {
					t.Errorf("Content-Language = %q, want %q", got, hd.wantLang)
				}
				if !strings.Contains(strings.Join(hdr.Values("Vary"), ","), "Accept-Language") {
					t.Errorf("Vary = %v, want Accept-Language", hdr.Values("Vary"))
				}
				if hdr.Get("Content-Type") != contentTypeJSON || hdr.Get("Cache-Control") != cacheControl {
					t.Errorf("headers = %v", hdr)
				}
				etag := hdr.Get("ETag")
				if !strings.HasPrefix(etag, `W/"`) {
					t.Errorf("ETag = %q", etag)
				}
				if prev, ok := etags[hd.wantLang]; ok && prev != etag {
					t.Errorf("ETag for %s changed: %s vs %s", hd.wantLang, prev, etag)
				}
				etags[hd.wantLang] = etag

				var body map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				want := ep.text.EN
				if hd.wantLang == "ru" {
					want = ep.text.RU
				}
				if got := ep.field(body); got != want {
					t.Errorf("body value = %v, want %q", got, want)
				}
			})
		}
		if etags["en"] == etags["ru"] {
			t.Errorf("%s: EN and RU ETags are equal", ep.path)
		}
	}
}

// FR-B05: the language comes only from Accept-Language, never from the URL.
func TestLangQueryParamIsIgnored(t *testing.T) {
	h := newHandler(&fakeStore{all: loadFixture(t)}, fakeDB{})
	if got := do(h, http.MethodGet, "/api/v1/main?lang=ru").Header().Get("Content-Language"); got != "en" {
		t.Errorf("?lang=ru: Content-Language = %q, want en", got)
	}
	rec := do(h, http.MethodGet, "/api/v1/main?lang=en", "Accept-Language", "ru")
	if got := rec.Header().Get("Content-Language"); got != "ru" {
		t.Errorf("?lang=en with Accept-Language ru: Content-Language = %q, want ru", got)
	}
}

func TestConditionalGet(t *testing.T) {
	h := newHandler(&fakeStore{all: loadFixture(t)}, fakeDB{})
	etag := do(h, http.MethodGet, "/api/v1/cv", "Accept-Language", "ru").Header().Get("ETag")

	for _, inm := range []string{etag, strings.TrimPrefix(etag, "W/"), `W/"other", ` + etag, "*"} {
		rec := do(h, http.MethodGet, "/api/v1/cv", "Accept-Language", "ru", "If-None-Match", inm)
		if rec.Code != http.StatusNotModified || rec.Body.Len() != 0 {
			t.Errorf("If-None-Match %s: status %d, body %d bytes; want 304 and empty", inm, rec.Code, rec.Body.Len())
		}
		if rec.Header().Get("ETag") != etag {
			t.Errorf("304 must repeat the ETag")
		}
	}

	// The RU ETag does not match the EN representation.
	if rec := do(h, http.MethodGet, "/api/v1/cv", "Accept-Language", "en", "If-None-Match", etag); rec.Code != http.StatusOK {
		t.Errorf("EN with the RU ETag: status %d, want 200", rec.Code)
	}
}

func TestHead(t *testing.T) {
	h := newHandler(&fakeStore{all: loadFixture(t)}, fakeDB{})
	// httptest.ResponseRecorder keeps the body; the real server drops it for
	// HEAD, so check only the status and headers here.
	rec := do(h, http.MethodHead, "/api/v1/main")
	if rec.Code != http.StatusOK || rec.Header().Get("ETag") == "" {
		t.Errorf("HEAD: status %d, headers %v", rec.Code, rec.Header())
	}
}

func assertError(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Errorf("status = %d, want %d", rec.Code, status)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contentTypeJSON {
		t.Errorf("Content-Type = %q", ct)
	}
	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body %q: %v", rec.Body, err)
	}
	if body.Error.Code != code || body.Error.Message == "" {
		t.Errorf("error = %+v, want code %q", body.Error, code)
	}
}

func TestNotFound(t *testing.T) {
	h := newHandler(&fakeStore{all: loadFixture(t)}, fakeDB{})
	for _, path := range []string{"/", "/api/v1/nope", "/api/v1/main/extra", "/api/v2/main", "/index.html"} {
		t.Run(path, func(t *testing.T) { assertError(t, do(h, http.MethodGet, path), http.StatusNotFound, CodeNotFound) })
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newHandler(&fakeStore{all: loadFixture(t)}, fakeDB{})
	for _, path := range []string{"/api/v1/health", "/api/v1/main", "/api/v1/cv", "/api/v1/travel"} {
		for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions} {
			t.Run(m+" "+path, func(t *testing.T) {
				rec := do(h, m, path)
				assertError(t, rec, http.StatusMethodNotAllowed, CodeMethodNotAllowed)
				if rec.Header().Get("Allow") != "GET, HEAD" {
					t.Errorf("Allow = %q", rec.Header().Get("Allow"))
				}
			})
		}
	}
}

func TestHealth(t *testing.T) {
	for _, tc := range []struct {
		err        error
		status     int
		wantStatus string
		wantDB     string
	}{
		{nil, http.StatusOK, "ok", "ok"},
		{errors.New("connection refused"), http.StatusServiceUnavailable, "degraded", "down"},
	} {
		rec := do(newHandler(&fakeStore{}, fakeDB{tc.err}), http.MethodGet, "/api/v1/health")
		var body healthBody
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if rec.Code != tc.status || body.Status != tc.wantStatus || body.DB != tc.wantDB {
			t.Errorf("ping err %v: %d %+v", tc.err, rec.Code, body)
		}
		if rec.Header().Get("Cache-Control") != "no-store" || rec.Header().Get("Content-Language") != "" {
			t.Errorf("health headers = %v", rec.Header())
		}
	}
}

func TestStoreErrorsMapToContractErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"generic", errors.New("syntax error at or near SELECT"), http.StatusInternalServerError, CodeInternal},
		{"not seeded", store.ErrNotSeeded, http.StatusServiceUnavailable, CodeUnavailable},
		{"connect", fmt.Errorf("get main: %w", &pgconn.ConnectError{}), http.StatusServiceUnavailable, CodeUnavailable},
		{"timeout", fmt.Errorf("get main: %w", context.DeadlineExceeded), http.StatusServiceUnavailable, CodeUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(newHandler(&fakeStore{err: tc.err}, fakeDB{}), http.MethodGet, "/api/v1/main")
			assertError(t, rec, tc.status, tc.code)
			if strings.Contains(rec.Body.String(), "SELECT") || rec.Header().Get("ETag") != "" {
				t.Errorf("error leaked details or an ETag: %s %v", rec.Body, rec.Header())
			}
		})
	}
}

func TestRealServerHeadAndGzip(t *testing.T) {
	srv := httptest.NewServer(newHandler(&fakeStore{all: loadFixture(t)}, fakeDB{}))
	defer srv.Close()

	resp, err := http.Head(srv.URL + "/api/v1/main")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	// Content-Length announces the GET size; the body itself must be empty.
	if resp.StatusCode != http.StatusOK || len(body) != 0 || resp.Header.Get("ETag") == "" {
		t.Errorf("HEAD: status %d, %d body bytes, headers %v", resp.StatusCode, len(body), resp.Header)
	}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/travel", nil)
	req.Header.Set("Accept-Encoding", "gzip") // set explicitly: the client then does not decompress
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("travel (> 1 KB) not gzipped: %v", resp.Header)
	}
	if !strings.Contains(strings.Join(resp.Header.Values("Vary"), ","), "Accept-Encoding") {
		t.Errorf("Vary lacks Accept-Encoding: %v", resp.Header.Values("Vary"))
	}
}
