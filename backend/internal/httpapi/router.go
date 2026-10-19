package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
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
	Service     *service.Service
	DB          Pinger
	Log         *slog.Logger
	CORSOrigins []string
}

const (
	cacheControl  = "no-store"
	healthTimeout = time.Second
)

type api struct {
	svc *service.Service
	db  Pinger
	log *slog.Logger
}

// New returns the complete HTTP handler: routes plus the middleware chain
// request ID → recover → logging → CORS → gzip (plan §4.1). Localised routes
// are additionally wrapped in i18n.Middleware.
func New(o Options) http.Handler {
	a := &api{svc: o.Service, db: o.DB, log: o.Log}
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

// content serves a localised response built from the database on every
// request; clients are told not to store it.
func (a *api) content(ep service.Endpoint) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := a.svc.Render(r.Context(), ep, i18n.FromContext(r.Context()))
		if err != nil {
			a.fail(w, r, err)
			return
		}
		h := w.Header()
		h.Set("Cache-Control", cacheControl)
		h.Set("Content-Type", contentTypeJSON)
		_, _ = w.Write(body) // net/http drops the body for HEAD
	})
}

type healthBody struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// health pings the database with a 1 s timeout.
func (a *api) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()
	w.Header().Set("Cache-Control", cacheControl)
	if err := a.db.Ping(ctx); err != nil {
		a.log.WarnContext(r.Context(), "health: database down", "id", RequestIDFrom(r.Context()), "err", err)
		writeJSON(w, http.StatusServiceUnavailable, healthBody{Status: "degraded", DB: "down"})
		return
	}
	writeJSON(w, http.StatusOK, healthBody{Status: "ok", DB: "ok"})
}

// fail logs err with the request id and maps it to a contract error:
// 503 unavailable when the database is unreachable or has no content, otherwise
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
	return errors.Is(err, store.ErrNoContent) || errors.As(err, &connErr) || errors.As(err, &netErr) ||
		errors.Is(err, context.DeadlineExceeded)
}
