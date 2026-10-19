package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"slices"
	"time"

	"github.com/klauspost/compress/gzhttp"
)

// Middleware wraps a handler.
type Middleware func(http.Handler) http.Handler

// Chain applies mws so that the first one is the outermost.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}

const requestIDHeader = "X-Request-ID"

// validRequestID limits incoming ids to safe characters, so they can be
// echoed and logged as-is.
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type requestIDKey struct{}

// RequestIDFrom returns the request id stored by RequestID, or "".
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// RequestID keeps a valid incoming X-Request-ID or generates a random one,
// stores it in the context and sets it on the response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if !validRequestID.MatchString(id) {
			id = newRequestID()
		}
		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // never fails (crypto/rand docs)
	return hex.EncodeToString(b)
}

// Recover turns a panic into a 500 internal_error response without details
// and logs the panic value with the stack trace.
func Recover(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler { // net/http's signal to abort silently
					panic(v)
				}
				log.ErrorContext(r.Context(), "panic",
					"id", RequestIDFrom(r.Context()), "panic", v, "stack", string(debug.Stack()))
				writeError(w, http.StatusInternalServerError, CodeInternal, "internal server error")
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder captures the status code and body size for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.status == 0 {
		s.status = code
	}
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if s.status == 0 {
		s.status = http.StatusOK
	}
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// Logging writes one structured log line per request (FR-B13). The language
// is read from the Content-Language response header set by i18n.Middleware,
// so it is empty for routes that are not localised.
func Logging(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			log.InfoContext(r.Context(), "request",
				"id", RequestIDFrom(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"bytes", rec.bytes,
				"duration_ms", float64(time.Since(start).Microseconds())/1000,
				"lang", w.Header().Get("Content-Language"),
			)
		})
	}
}

// CORS allows cross-origin GET/HEAD requests from the configured origins
// (development only). With no origins it is a no-op. Requests from other
// origins get no CORS headers, so the browser blocks them.
func CORS(origins []string) Middleware {
	return func(next http.Handler) http.Handler {
		if len(origins) == 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" || !slices.Contains(origins, origin) {
				next.ServeHTTP(w, r)
				return
			}
			h := w.Header()
			h.Add("Vary", "Origin")
			h.Set("Access-Control-Allow-Origin", origin)
			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				h.Add("Vary", "Access-Control-Request-Method")
				h.Add("Vary", "Access-Control-Request-Headers")
				h.Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Accept-Language")
				h.Set("Access-Control-Max-Age", "600")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h.Set("Access-Control-Expose-Headers", "Content-Language, X-Request-ID")
			next.ServeHTTP(w, r)
		})
	}
}

// gzipMinSize is the smallest body that is compressed (plan §4.1).
const gzipMinSize = 1024

// Gzip compresses responses larger than 1 KB for clients that accept gzip.
func Gzip() Middleware {
	wrap, err := gzhttp.NewWrapper(gzhttp.MinSize(gzipMinSize))
	if err != nil {
		panic(err) // only fails on invalid static options
	}
	return func(next http.Handler) http.Handler { return wrap(next) }
}
