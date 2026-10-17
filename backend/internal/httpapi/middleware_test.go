package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/leonid6372/my-site/backend/internal/i18n"
)

func jsonLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

func ok(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func TestRecoverReturns500WithoutDetails(t *testing.T) {
	var logs bytes.Buffer
	h := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("secret: password=hunter2")
	}), RequestID, Recover(jsonLogger(&logs)))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/main", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != contentTypeJSON {
		t.Errorf("Content-Type = %q", ct)
	}
	want := `{"error":{"code":"internal_error","message":"internal server error"}}`
	if got := rec.Body.String(); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
	if !strings.Contains(logs.String(), "hunter2") || !strings.Contains(logs.String(), `"stack"`) {
		t.Errorf("panic value and stack should be logged, got %s", logs.String())
	}
	if !strings.Contains(logs.String(), rec.Header().Get(requestIDHeader)) {
		t.Errorf("log should contain the request id")
	}
}

func TestRecoverRepanicsOnAbortHandler(t *testing.T) {
	h := Recover(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	defer func() {
		if v := recover(); v != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler", v)
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestLoggingLineHasAllFields(t *testing.T) {
	var logs bytes.Buffer
	h := Chain(http.HandlerFunc(ok), RequestID, Logging(jsonLogger(&logs)), i18n.Middleware)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/main", nil)
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")
	req.Header.Set(requestIDHeader, "abc-123")
	h.ServeHTTP(httptest.NewRecorder(), req)

	var line map[string]any
	if err := json.Unmarshal(logs.Bytes(), &line); err != nil {
		t.Fatalf("log is not one JSON line: %v\n%s", err, logs.String())
	}
	want := map[string]any{
		"msg": "request", "id": "abc-123", "method": "GET", "path": "/api/v1/main",
		"status": float64(200), "bytes": float64(len(`{"status":"ok"}`)), "lang": "ru",
	}
	for k, v := range want {
		if line[k] != v {
			t.Errorf("log[%q] = %v, want %v", k, line[k], v)
		}
	}
	if _, ok := line["duration_ms"].(float64); !ok {
		t.Errorf("log lacks numeric duration_ms: %v", line)
	}
}

func TestLoggingRecordsErrorStatus(t *testing.T) {
	var logs bytes.Buffer
	h := Logging(jsonLogger(&logs))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, CodeNotFound, "resource not found")
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/nope", nil))
	if !strings.Contains(logs.String(), `"status":404`) || !strings.Contains(logs.String(), `"lang":""`) {
		t.Errorf("log = %s", logs.String())
	}
}

func TestRequestID(t *testing.T) {
	var seen string
	h := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))
	for _, tc := range []struct{ in, want string }{
		{"keep-me.1", "keep-me.1"},
		{"", ""},                      // generated
		{"bad id\nX-Injected: 1", ""}, // replaced
		{strings.Repeat("a", 65), ""}, // too long, replaced
	} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(requestIDHeader, tc.in)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		got := rec.Header().Get(requestIDHeader)
		if got != seen {
			t.Errorf("header %q != context %q", got, seen)
		}
		if tc.want != "" && got != tc.want {
			t.Errorf("in %q: id = %q, want %q", tc.in, got, tc.want)
		}
		if tc.want == "" && (got == tc.in || len(got) != 16) {
			t.Errorf("in %q: id = %q, want a new 16-char id", tc.in, got)
		}
	}
}

const devOrigin = "http://localhost:5173"

func TestCORSPreflightFromAllowedOrigin(t *testing.T) {
	h := CORS([]string{devOrigin})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("preflight must not reach the handler")
	}))
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/main", nil)
	req.Header.Set("Origin", devOrigin)
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "accept-language")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rec.Code)
	}
	for k, v := range map[string]string{
		"Access-Control-Allow-Origin":  devOrigin,
		"Access-Control-Allow-Methods": "GET, HEAD, OPTIONS",
		"Access-Control-Allow-Headers": "Accept-Language",
	} {
		if got := rec.Header().Get(k); got != v {
			t.Errorf("%s = %q, want %q", k, got, v)
		}
	}
	if vary := rec.Header().Values("Vary"); !strings.Contains(strings.Join(vary, ","), "Origin") {
		t.Errorf("Vary = %v, want Origin", vary)
	}
}

func TestCORSSimpleRequestFromAllowedOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/main", nil)
	req.Header.Set("Origin", devOrigin)
	rec := httptest.NewRecorder()
	CORS([]string{devOrigin})(http.HandlerFunc(ok)).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || rec.Header().Get("Access-Control-Allow-Origin") != devOrigin {
		t.Errorf("status %d, headers %v", rec.Code, rec.Header())
	}
	if !strings.Contains(rec.Header().Get("Access-Control-Expose-Headers"), "Content-Language") {
		t.Errorf("Content-Language should be exposed: %v", rec.Header())
	}
}

func TestCORSOtherOriginOrDisabled(t *testing.T) {
	for name, origins := range map[string][]string{
		"other origin":   {devOrigin},
		"not configured": nil,
	} {
		t.Run(name, func(t *testing.T) {
			reached := false
			h := CORS(origins)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached = true }))
			req := httptest.NewRequest(http.MethodOptions, "/api/v1/main", nil)
			req.Header.Set("Origin", "http://evil.example")
			req.Header.Set("Access-Control-Request-Method", "GET")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			for k := range rec.Header() {
				if strings.HasPrefix(k, "Access-Control-") {
					t.Errorf("unexpected CORS header %s", k)
				}
			}
			if !reached {
				t.Error("request should be passed on (the router answers 405)")
			}
		})
	}
}

func TestGzipOnlyLargeBodies(t *testing.T) {
	for _, tc := range []struct {
		size     int
		wantGzip bool
	}{
		{100, false},
		{4096, true},
	} {
		body := strings.Repeat("a", tc.size)
		h := Gzip()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", contentTypeJSON)
			_, _ = io.WriteString(w, body)
		}))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		gz := rec.Header().Get("Content-Encoding") == "gzip"
		if gz != tc.wantGzip {
			t.Errorf("%d bytes: gzip = %v, want %v", tc.size, gz, tc.wantGzip)
		}
	}
}

func TestWriteErrorShape(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("ETag", `W/"x"`)
	writeError(rec, http.StatusServiceUnavailable, CodeUnavailable, "database unavailable")
	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 503 || body.Error.Code != "unavailable" || body.Error.Message != "database unavailable" {
		t.Errorf("got %d %+v", rec.Code, body)
	}
	if rec.Header().Get("ETag") != "" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("errors must not be cached: %v", rec.Header())
	}
}
