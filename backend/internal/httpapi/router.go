package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/leonid6372/my-site/backend/internal/i18n"
	"github.com/leonid6372/my-site/backend/internal/service"
	"github.com/leonid6372/my-site/backend/internal/store"
)

// Pinger checks that the database is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Options configures the HTTP handler.
type Options struct {
	Cache       *service.Cache
	DB          Pinger
	Log         *slog.Logger
	CORSOrigins []string
}

const (
	cacheControl  = "public, max-age=300"
	healthTimeout = time.Second
)

type api struct {
	cache *service.Cache
	db    Pinger
	log   *slog.Logger
}

// New returns the complete HTTP handler: routes plus the middleware chain
// request ID → recover → logging → CORS → gzip (plan §4.1). Localised routes
// are additionally wrapped in i18n.Middleware.
func New(o Options) http.Handler {
	a := &api{cache: o.Cache, db: o.DB, log: o.Log}
	mux := http.NewServeMux()
	// Patterns have no method, so non-GET requests reach getOnly and get the
	// contract's JSON 405 instead of ServeMux's plain-text one.
	mux.Handle("/api/v1/health", getOnly(http.HandlerFunc(a.health)))
	mux.Handle("/api/v1/main", getOnly(i18n.Middleware(a.content(service.EndpointMain))))
	mux.Handle("/api/v1/cv", getOnly(i18n.Middleware(a.content(service.EndpointCV))))
	mux.Handle("/api/v1/travel", getOnly(i18n.Middleware(a.content(service.EndpointTravel))))
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, CodeNotFound, "resource not found")
	})

	return Chain(mux, RequestID, Recover(o.Log), Logging(o.Log), CORS(o.CORSOrigins), Gzip())
}

// getOnly answers every method except GET and HEAD with 405.
func getOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			writeError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method not allowed")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// content serves a cached, localised response with ETag revalidation.
func (a *api) content(ep service.Endpoint) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := a.cache.Get(r.Context(), ep, i18n.FromContext(r.Context()))
		if err != nil {
			a.fail(w, r, err)
			return
		}
		h := w.Header()
		h.Set("ETag", resp.ETag)
		h.Set("Cache-Control", cacheControl)
		if etagMatches(r.Header.Get("If-None-Match"), resp.ETag) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		h.Set("Content-Type", contentTypeJSON)
		_, _ = w.Write(resp.Body) // net/http drops the body for HEAD
	})
}

// etagMatches implements the weak comparison of If-None-Match (RFC 9110
// §13.1.2): "*" or any listed tag equal to etag, ignoring the W/ prefix.
func etagMatches(header, etag string) bool {
	if header == "" {
		return false
	}
	want := strings.TrimPrefix(etag, "W/")
	for tag := range strings.SplitSeq(header, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "*" || strings.TrimPrefix(tag, "W/") == want {
			return true
		}
	}
	return false
}

type healthBody struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// health pings the database with a 1 s timeout; it is never cached.
func (a *api) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()
	w.Header().Set("Cache-Control", "no-store")
	if err := a.db.Ping(ctx); err != nil {
		a.log.WarnContext(r.Context(), "health: database down", "id", RequestIDFrom(r.Context()), "err", err)
		writeJSON(w, http.StatusServiceUnavailable, healthBody{Status: "degraded", DB: "down"})
		return
	}
	writeJSON(w, http.StatusOK, healthBody{Status: "ok", DB: "ok"})
}

// fail logs err with the request id and maps it to a contract error:
// 503 unavailable when the database is unreachable or not seeded, otherwise
// 500 internal_error. The response never contains err itself.
func (a *api) fail(w http.ResponseWriter, r *http.Request, err error) {
	a.log.ErrorContext(r.Context(), "request failed", "id", RequestIDFrom(r.Context()), "err", err)
	if unavailable(err) {
		writeError(w, http.StatusServiceUnavailable, CodeUnavailable, "service temporarily unavailable")
		return
	}
	writeError(w, http.StatusInternalServerError, CodeInternal, "internal server error")
}

func unavailable(err error) bool {
	var (
		connErr *pgconn.ConnectError
		netErr  net.Error
	)
	return errors.Is(err, store.ErrNotSeeded) || errors.As(err, &connErr) || errors.As(err, &netErr) ||
		errors.Is(err, context.DeadlineExceeded)
}
