# Backend Implementation Plan

Status: **Approved**  ·  Last updated: 2026-10-17
Implements: [backend/spec.md](spec.md) · [api/contract.md](../api/contract.md)

This document describes **how** the backend is built. Behaviour is defined in the spec;
if this plan contradicts the spec, the spec wins.

## 1. Constitution check

| Principle | How the plan complies |
|-----------|-----------------------|
| Simplicity | One Go binary, standard library HTTP router, no ORM, no framework. |
| Content as code | YAML/JSON in `content/` is the only source of content; the DB is a derived copy that is rebuilt on every start. |
| Bilingual parity | The validator rejects any `ltext` missing `en` or `ru`. |
| Read-only API | Only `GET`/`HEAD` routes are registered; everything else → `405`. |
| No secrets in git | All configuration via env; `.env.example` committed, `.env` git-ignored. |
| Testing | Unit + `httptest` + PostgreSQL integration tests (testcontainers). |

## 2. Technology choices

| Concern | Choice | Why |
|---------|--------|-----|
| Language | Go — latest stable release, pinned in `go.mod` (`go` + `toolchain` lines) and the Dockerfile | Required by the constitution. |
| HTTP routing | `net/http` `ServeMux` with method + path patterns (`GET /api/v1/main`) | Built in; enough for 4 routes. |
| PostgreSQL driver | `github.com/jackc/pgx/v5` + `pgxpool` | De-facto standard, fast, native JSONB/array support. |
| Migrations | `github.com/pressly/goose/v3` with SQL files embedded via `embed.FS` | Simple SQL migrations, runs in-process. |
| YAML | `go.yaml.in/yaml/v3` (the maintained fork of `gopkg.in/yaml.v3`), `Decoder.KnownFields(true)` | Strict decoding catches unknown keys. |
| JSON content | `encoding/json` with `DisallowUnknownFields()` | Same strictness as YAML. |
| Accept-Language | `golang.org/x/text/language` (`ParseAcceptLanguage`) + own selection logic | Robust RFC parsing; selection rules from the spec. |
| Logging | `log/slog` JSON handler | Standard library. |
| Gzip | `github.com/klauspost/compress/gzhttp` | Well-tested gzip middleware. |
| CLI | `os.Args` switch + `flag` per subcommand | No CLI framework needed. |
| Integration tests | `github.com/testcontainers/testcontainers-go/modules/postgres` | Real PostgreSQL 16 in tests. |
| Linting | `golangci-lint` v2 (`govet`, `staticcheck`, `errcheck`, `revive`, `gosec`) | Constitution requirement. |

## 3. Architecture

```
          ┌──────────────────────── api binary ────────────────────────┐
content/  │  content.Load ─► content.Validate ─► store.ReplaceAll (tx)  │──► PostgreSQL
(yaml)    │                                                            │
          │  HTTP ─► middleware ─► handler ─► service ─► store.Read*   │◄── PostgreSQL
          │                       (i18n lang)   (localise, stats,      │
          │                                      sort, cache)          │
          └────────────────────────────────────────────────────────────┘
```

### 3.1 Packages

| Package | Responsibility | Depends on |
|---------|----------------|-----------|
| `cmd/api` | Subcommand dispatch, wiring, signal handling. | all |
| `internal/config` | Parse and validate env vars into a `Config` struct. | — |
| `internal/content` | Content structs (`MainFile`, `CVFile`, `TravelFile`, `LText`, `Photo`), loader (YAML or JSON per file), validator returning `[]ValidationError`. | `geo` |
| `internal/geo` | Embedded `countries.json`; `Lookup(iso2) (Country, bool)`. | — |
| `internal/i18n` | `Negotiate(header string) Lang`, middleware storing `Lang` in the context, `FromContext(ctx)`. | `x/text` |
| `internal/store` | `pgxpool` setup, embedded migrations, `ReplaceAll(ctx, content.All)`, read queries returning DB models. | `pgx`, `goose` |
| `internal/service` | Builds localised response DTOs (matching the contract) from store models, stats, sorting; response cache. | `store`, `geo`, `i18n` |
| `internal/httpapi` | Router, handlers, middleware (request ID, recover, logging, CORS, gzip), JSON/error helpers, ETag. | `service`, `i18n` |

Dependencies flow downward only: `httpapi → service → store`, and `content` is used only
by the seed path.

### 3.2 Key types

```go
// internal/content
type LText struct{ EN, RU string }               // yaml/json: {en: "...", ru: "..."}
type Photo struct {
    URL, ThumbURL string; Alt, Caption *LText; Width, Height *int
}

// internal/i18n
type Lang string                                  // "en" | "ru"
const EN, RU Lang = "en", "ru"

// internal/service – DTOs mirror api/contract.md exactly (json tags camelCase)
type MainDTO struct{ FullName string `json:"fullName"`; /* … */ }
```

`LText` is stored in PostgreSQL as JSONB `{"en": "...", "ru": "..."}` through a
`Value()/Scan()` implementation, so the store layer stays type-safe.

## 4. Request handling details

### 4.1 Middleware chain (outer → inner)

1. `requestID`: reads `X-Request-ID` or generates a random one; sets the response header.
2. `recover`: on panic, log the stack and return `500 internal_error`.
3. `logging`: `slog` line with `id, method, path, status, bytes, duration_ms, lang`.
4. `cors`: only if `CORS_ORIGINS` is set; allows `GET, HEAD, OPTIONS` and the header
   `Accept-Language`; answers preflight `OPTIONS` with `204`.
5. `i18n`: `Negotiate(r.Header.Get("Accept-Language"))` → context; sets
   `Content-Language` and `Vary: Accept-Language` on localised routes.
6. `gzhttp`: compresses responses > 1 KB.

### 4.2 Language negotiation algorithm

```
tags, qs, err := language.ParseAcceptLanguage(header)
if err != nil or header == "" → EN
best := none; bestQ := 0
for i, tag in tags (header order):
    base := tag.Base()            // "ru-RU" → "ru"
    if base in {en, ru} and qs[i] > bestQ: best, bestQ = base, qs[i]
return best or EN
```

Ties keep the earlier tag (strict `>`), which matches "ties resolved by header order".

### 4.3 Response cache and ETag

Content only changes when the container restarts (the seed runs on every start), so the
service **caches the serialised JSON per (endpoint, lang)** in memory:

- built lazily on the first request, guarded by `sync.Once` per key;
- each entry holds `body []byte` and `etag = W/"<first 16 hex of sha256(body)>"`;
- the handler compares `If-None-Match` and returns `304` with an empty body when it matches;
- headers: `Content-Type: application/json; charset=utf-8`, `Cache-Control: public, max-age=300`.

So the DB is queried at most once per endpoint and language per process lifetime.
`/health` is never cached (it pings the DB with a 1 s timeout).

### 4.4 Errors

`httpapi.writeError(w, status, code, msg)` is the only way errors leave the service.
Store and service errors are logged with the request ID and mapped to `500 internal_error`
(or `503 unavailable` for `pgconn` connection errors). Messages never include internals.

## 5. Seed pipeline

```
api start  =  migrate → load → validate → ReplaceAll (one tx) → serve
```

1. **Load:** for each of `main`, `cv`, `travel`, find exactly one of `<name>.yaml`,
   `<name>.yml` or `<name>.json` in `CONTENT_DIR` (none or more than one → error). Decode strictly.
2. **Validate:** walk the structs and collect `ValidationError{File, Path, Msg}` for every
   rule in spec §6. Paths use JSON-pointer-like notation: `places[3].photos[2].url`.
   Email validation uses `net/mail.ParseAddress` on the part after `mailto:`.
3. **ReplaceAll:** in one `pgx.Tx`:
   `TRUNCATE main_info, social_link, cv_experience, cv_education, cv_skill_group, cv_language, place, visit, photo RESTART IDENTITY CASCADE;`
   then bulk insert with `pgx.Batch` (or `CopyFrom` for photos). Any error → rollback; the
   previous content stays intact.
4. Exit code `1` with the printed error list on failure; `api start` does **not** start the
   server if the seed fails (the container restarts, and the logs show the reason).

## 6. Country reference data

- `internal/geo/countries.json`, ~250 entries:
  `{ "iso2": "NL", "isoNumeric": "528", "continent": "europe", "name": { "en": "Netherlands", "ru": "Нидерланды" } }`.
- Generated **once** by `tools/gencountries` (a small Go program, not part of the binary)
  from the `mledoze/countries` dataset (fields `cca2`, `ccn3`, `region`/`subregion`,
  `name.common`, `translations.rus.common`). The output is committed.
- Continent mapping: `Europe → europe`, `Asia → asia`, `Africa → africa`,
  `Oceania → oceania`, `Antarctic → antarctica`, `Americas` → by subregion
  (`South America → south-america`, everything else (`North America`, `Central America`,
  `Caribbean`) → `north-america`).
- Loaded via `//go:embed` into a `map[string]Country` at init.

## 7. Database

- Migration `00001_init.sql` creates exactly the schema from spec §7.
- Pool: `MaxConns = 5` (tiny workload), `HealthCheckPeriod = 30s`.
- On startup, the connection is retried with backoff for up to 30 s (the DB may still be
  starting), in addition to the compose `depends_on: condition: service_healthy`.

## 8. Commands (final)

Besides `serve`, `seed`, `validate` and `migrate`, the distroless image (no shell, no curl)
needs two helpers, which were added to the spec during this phase:

| Command | Purpose |
|---------|---------|
| `api start` | `seed` then `serve` in one process: the container entrypoint (replaces `api seed && api serve`). |
| `api healthcheck` | `GET http://127.0.0.1:8080/api/v1/health`; exit 0 on 200, else 1. Used by the Docker `HEALTHCHECK`. |

They are recorded in the spec changelog (a clarification, with no change to the API behaviour).

## 9. Docker and deployment

### 9.1 `backend/Dockerfile`

```dockerfile
FROM golang:<pinned>-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
USER nonroot
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD ["/api", "healthcheck"]
ENTRYPOINT ["/api"]
CMD ["start"]
```

`ENTRYPOINT` + `CMD` lets other subcommands run in the same image, e.g.
`docker compose run --rm api validate`.

### 9.2 `docker-compose.yml` (repo root)

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes: [pgdata:/var/lib/postgresql/data]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $${POSTGRES_USER} -d $${POSTGRES_DB}"]
      interval: 5s
      retries: 10
    restart: unless-stopped

  api:
    build: ./backend
    environment:
      DATABASE_URL: postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB}?sslmode=disable
      CONTENT_DIR: /app/content
      LOG_LEVEL: info
    volumes: ["./content:/app/content:ro"]
    depends_on: { db: { condition: service_healthy } }
    restart: unless-stopped

  web:
    build: ./frontend
    ports: ["80:80"]
    depends_on: { api: { condition: service_healthy } }
    restart: unless-stopped

volumes:
  pgdata:
```

`docker-compose.dev.yml` (used only locally) publishes `db` on `localhost:5432`.

### 9.3 Content update workflow (on the VPS)

```
git pull                                   # new YAML content
docker compose run --rm api validate       # check content first; fix errors before restarting
docker compose restart api                 # start → re-seed → serve with fresh cache
```

The content is bind-mounted, so content changes need **no image rebuild**. Code changes:
`git pull && docker compose up -d --build`.

## 10. Local development

```
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d db
cd backend
set DATABASE_URL / CONTENT_DIR=../content / CORS_ORIGINS=http://localhost:5173
go run ./cmd/api start
```

Placeholder content in `content/` uses stable `https://picsum.photos/id/<n>/1600/1067`
URLs (plus `/400/400` thumbs), so the site works out of the box.

## 11. Testing strategy

| Layer | Approach | Location |
|-------|----------|----------|
| `content` validator | Table-driven; `testdata/valid/` plus one minimal broken fixture per rule; asserts the exact `File:Path` of each error. | `internal/content/*_test.go` |
| `i18n` | Table tests for every case listed in spec §12. | `internal/i18n/*_test.go` |
| `service` | Fake store (interface) → DTOs; golden JSON files in `testdata/golden/{main,cv,travel}.{en,ru}.json`; `-update` flag regenerates them. | `internal/service` |
| `httpapi` | `httptest.Server` with the fake service: status codes, headers (`Content-Language`, `Vary`, `ETag`, `Cache-Control`), `304`, `404`, `405`, CORS preflight. | `internal/httpapi` |
| `store` | testcontainers PostgreSQL: migrate → `ReplaceAll(valid)` → read back equality; run twice (idempotent); a failing insert mid-tx leaves the previous data. Guarded by `-tags=integration`. | `internal/store` |
| Smoke | Script `scripts/smoke.sh`: compose up, curl health + `Accept-Language: ru` main. | repo root |

CI (later, optional): `go vet`, `golangci-lint run`, `go test ./...`,
`go test -tags=integration ./...`, `api validate` against `content/`.

## 12. Risks and mitigations

| Risk | Mitigation |
|------|-----------|
| TopoJSON country ids not matching (e.g. Kosovo has no ISO numeric code) | `geo` data includes `isoNumeric` only where official; the frontend shows the pin even if the country polygon can't be highlighted. Documented in the frontend plan. |
| A bad content edit makes `api start` fail, which takes the API down (the seed runs before serving) | Always run `api validate` before deploying: locally, in CI, or as `docker compose run --rm api validate` on the VPS before the restart. The failed seed is transactional, so the DB keeps the previous content, and reverting the YAML restores the site. Documented in the README. |
| External photo hosts go down | Out of our control by decision; the frontend shows placeholders. |
| Russian plural/declension in the API | No such logic in the backend; all wording is in content or frontend i18n. |

## Changelog

| Date       | Change |
|------------|--------|
| 2026-10-17 | Initial plan. |
