# Backend Specification

Status: **Implemented**  ·  Last updated: 2026-10-19
Related: [Constitution](../000-constitution.md) · [Requirements](../001-requirements.md) · [API contract](../api/contract.md)

## 1. Responsibilities

The backend is a single Go service (`api`) that:

1. **Reads content from PostgreSQL**, the only place data is stored. The API never writes
   content: it creates the schema (migrations) and reads the tables on every request.
2. **Serves the read-only JSON API** defined in [api/contract.md](../api/contract.md), in
   the language negotiated from the `Accept-Language` header.
3. **Computes derived data**: travel statistics, visited countries and continents.

It does **not** handle authentication, writes from the public, email, file uploads, or
storing/serving images. Photos are external URLs that are passed through unchanged.

## 2. Functional requirements

| ID     | Requirement | Traces to |
|--------|-------------|-----------|
| FR-B01 | Expose `GET /api/v1/health` reporting service and DB status. | NFR-06 |
| FR-B02 | Expose `GET /api/v1/main` returning the main-page info (profile, photo, interests, links). | US-10, US-30 |
| FR-B03 | Expose `GET /api/v1/cv` returning summary, experience, education, skills and spoken languages. | US-30 |
| FR-B04 | Expose `GET /api/v1/travel` returning stats, visited countries and places with visits and photos. | US-20…US-24 |
| FR-B05 | Negotiate the language from `Accept-Language` per contract §1.1 (fallback `en`); set `Content-Language` and `Vary: Accept-Language`. No `lang` query parameter. | US-02 |
| FR-B06 | Return photo objects with external URLs exactly as stored in the DB (no download, proxying or rewriting). | US-10, US-22 |
| FR-B07 | *(Removed 2026-10-19)* No content import: the database is the only source of data. | G4 |
| FR-B08 | *(Removed 2026-10-19)* No content validation command; data rules are in §6. | NFR-05 |
| FR-B09 | Provide country reference data (ISO2, ISO numeric, continent, EN/RU names) embedded in the binary. | US-21, US-23 |
| FR-B10 | Apply DB migrations automatically on startup (`serve`). Startup never inserts or changes content. | NFR-06 |
| FR-B11 | Return JSON errors as defined in the contract; never leak internal error details. | NFR-07 |
| FR-B12 | No caching: every API response is read from the DB and sent with `Cache-Control: no-store`. | NFR-01 |
| FR-B13 | Log every request as structured JSON (method, path, status, duration, resolved language). | — |

## 3. Commands

The binary `api` has subcommands:

| Command        | Behaviour |
|----------------|-----------|
| `api serve`    | Run migrations, then start the HTTPS server (the container's default command). Never writes content. |
| `api migrate`  | Run migrations only. |
| `api healthcheck` | Calls the local `/api/v1/health`; exit 0 if healthy (used by the Docker `HEALTHCHECK`, since the image has no shell or curl). |

In Docker the container runs `api serve`. A fresh database is empty until data is added.

## 4. Configuration (environment variables)

| Variable         | Default            | Description |
|------------------|--------------------|-------------|
| `HTTP_ADDR`      | `:8080`            | Listen address (HTTPS). |
| `TLS_CERT_FILE`  | —                  | PEM certificate for HTTPS. Set together with `TLS_KEY_FILE`; if both are empty a self-signed certificate is generated at startup. |
| `TLS_KEY_FILE`   | —                  | PEM private key for `TLS_CERT_FILE`. |
| `DATABASE_URL`   | — (required)       | e.g. `postgres://site:secret@db:5432/site?sslmode=disable` |
| `CORS_ORIGINS`   | empty              | Comma-separated allowed origins (dev: `http://localhost:5173`). |
| `LOG_LEVEL`      | `info`             | `debug`, `info`, `warn`, `error`. |

Startup fails fast with a clear message if a required variable is missing.

## 5. Content data (source of truth: PostgreSQL)

All content lives in the PostgreSQL tables of §7 and is edited there directly (SQL or any
database client). There are no content files and no import step. Changes are visible on the
next request.

All texts shown to visitors are `ltext` JSONB objects `{"en": "...", "ru": "..."}`.

| Table | Holds | Order in the API |
|-------|-------|------------------|
| `main_info` | single row (`id = 1`): name, headline, location, currently, bio, CV summary, photo, interests | — |
| `social_link` | contact links: `kind` (`email`, `github`, `linkedin`, `telegram`), `label`, `url` | `sort_order` |
| `cv_experience` | jobs; `end_date` `NULL` = present | `start_date` desc |
| `cv_education` | education entries | `start_date` desc |
| `cv_skill_group` | skill categories with non-localised `items` | `sort_order` |
| `cv_language` | spoken languages with level | `sort_order` |
| `place` | visited places: `country_iso2`, `name`, `lat`, `lng`, `story` | latest visit desc |
| `visit` | visits of a place (`start_date`, `end_date`) | `start_date` desc |
| `photo` | photos of a place: `url`, optional `thumb_url`, `alt`, `caption`, `width`, `height` | `sort_order` |

If `main_info` has no row, `/api/v1/main` and `/api/v1/cv` answer `503 unavailable`.

## 6. Data rules (kept by whoever edits the database)

The API does not validate content. The schema enforces what it can (`NOT NULL`, `CHECK`,
foreign keys, unique `kind`); the rest must be kept by the editor:

| Rule | Applies to |
|------|-----------|
| Every `ltext` has non-empty `en` **and** `ru`. | all |
| `id` values match `^[a-z0-9-]+$`. | experience, education, places |
| `end ≥ start`; `end` may be `NULL` only for experience/education. | cv, travel |
| `country_iso2` is a known ISO 3166-1 alpha-2 code (an unknown code makes `/api/v1/travel` fail with 500). | travel |
| `lat ∈ [-90, 90]`, `lng ∈ [-180, 180]`. | travel |
| Each place has ≥ 1 visit and ≤ 30 photos. | travel |
| Photo `url` and `thumb_url` are absolute `https://` URLs (not fetched). | main, travel |
| Photo `width` and `height` are both set or both `NULL`. | main, travel |
| For `email` links the `url` is `mailto:<address>`; for the others an absolute `https://` URL. | main |
| Paragraphs in `bio` and `story` are separated by a blank line. | main, travel |

## 7. Data model (PostgreSQL)

Localised text is stored as `JSONB` objects `{"en": "...", "ru": "..."}` (type alias **ltext**).
The schema is created by the embedded migrations; the application only reads these tables.

```sql
-- singleton: main page info
CREATE TABLE main_info (
  id             SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  full_name      JSONB NOT NULL,   -- ltext
  headline       JSONB NOT NULL,
  location       JSONB NOT NULL,
  currently      JSONB NOT NULL,
  bio            JSONB NOT NULL,   -- ltext (paragraphs separated by blank lines)
  summary        JSONB NOT NULL,   -- CV professional summary
  photo_url      TEXT  NOT NULL,
  photo_thumb_url TEXT,
  photo_alt      JSONB,
  photo_width    INT,
  photo_height   INT,
  interests      JSONB NOT NULL,   -- array of ltext
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE social_link (
  id          SERIAL PRIMARY KEY,
  kind        TEXT  NOT NULL UNIQUE CHECK (kind IN ('email','github','linkedin','telegram')),
  label       JSONB NOT NULL,
  url         TEXT  NOT NULL,
  sort_order  INT   NOT NULL
);

CREATE TABLE cv_experience (
  id           TEXT PRIMARY KEY,
  position     JSONB NOT NULL,
  company      TEXT  NOT NULL,
  company_url  TEXT,
  location     JSONB NOT NULL,
  start_date   DATE  NOT NULL,
  end_date     DATE,
  description  JSONB,
  highlights   JSONB NOT NULL DEFAULT '[]'  -- array of ltext
);

CREATE TABLE cv_education (
  id           TEXT PRIMARY KEY,
  institution  JSONB NOT NULL,
  degree       JSONB NOT NULL,
  field        JSONB NOT NULL,
  start_date   DATE  NOT NULL,
  end_date     DATE,
  description  JSONB
);

CREATE TABLE cv_skill_group (
  id          SERIAL PRIMARY KEY,
  category    JSONB NOT NULL,
  items       TEXT[] NOT NULL,
  sort_order  INT NOT NULL
);

CREATE TABLE cv_language (
  id          SERIAL PRIMARY KEY,
  name        JSONB NOT NULL,
  level       JSONB NOT NULL,
  sort_order  INT NOT NULL
);

CREATE TABLE place (
  id            TEXT PRIMARY KEY,
  country_iso2  CHAR(2) NOT NULL,
  name          JSONB NOT NULL,
  lat           DOUBLE PRECISION NOT NULL,
  lng           DOUBLE PRECISION NOT NULL,
  story         JSONB NOT NULL
);

CREATE TABLE visit (
  id          SERIAL PRIMARY KEY,
  place_id    TEXT NOT NULL REFERENCES place(id) ON DELETE CASCADE,
  start_date  DATE NOT NULL,
  end_date    DATE NOT NULL,
  CHECK (end_date >= start_date)
);

CREATE TABLE photo (
  id          SERIAL PRIMARY KEY,
  place_id    TEXT NOT NULL REFERENCES place(id) ON DELETE CASCADE,
  url         TEXT NOT NULL,
  thumb_url   TEXT,
  alt         JSONB,
  caption     JSONB,
  width       INT,
  height      INT,
  sort_order  INT NOT NULL,
  CHECK ((width IS NULL) = (height IS NULL))
);

CREATE INDEX ON visit (place_id);
CREATE INDEX ON photo (place_id);
```

Country reference data (≈250 rows: ISO2, ISO numeric, continent, EN/RU names) is **not**
stored in the DB. It is a JSON file embedded in the binary (`go:embed`) and looked up in memory.

## 8. Business logic

- **Language negotiation:** middleware parses `Accept-Language` (RFC 9110: comma-separated
  tags with optional `;q=`), selects the best supported language (`en`/`ru`, primary-subtag
  match, highest `q`, ties resolved by header order), falls back to `en`, and stores it in
  the request context. Handlers never read the header themselves.
- **Localisation:** handlers resolve every `ltext` to the negotiated language before
  serialising. The response never contains both languages.
- **Bio / story paragraphs:** the main-page `bio` is split on blank lines into `string[]`;
  the travel `story` is returned as one string with `\n\n` separators (see contract).
- **Photos:** pass-through of `url`, `thumbUrl`, `width` and `height`; `alt` and `caption`
  are localised; missing optional values → `null`.
- **Travel stats:**
  - `countries` = number of distinct `place.country_iso2`;
  - `continents` = number of distinct continents of those countries;
  - `places` = number of places;
  - `worldPercent` = `round(countries / 195 * 100, 1)`.
- **Sorting:** as defined in the contract (§2.3, §2.4).
- **No caching:** every request reads the DB; responses are `Cache-Control: no-store`.

## 9. HTTP server behaviour

- Router mounts `/api/v1/*`; everything else → `404 not_found` JSON. Non-`GET`/`HEAD` → `405`.
- Middleware order: request ID → recover (panic → 500) → structured logging → CORS (if
  configured; allows the `Accept-Language` header) → language negotiation → gzip.
- Timeouts: read 5 s, write 10 s, idle 60 s. Graceful shutdown on SIGTERM (max 10 s).

## 10. Code layout

```
backend/
├── cmd/api/main.go          # subcommands: serve, migrate, healthcheck
├── internal/
│   ├── config/              # env parsing
│   ├── content/             # content model structs (contenttest/: JSON fixture for tests)
│   ├── geo/                 # embedded countries.json + lookup
│   ├── i18n/                # Accept-Language parsing + negotiation middleware
│   ├── store/               # PostgreSQL repository (pgx)
│   │   └── migrations/      # SQL migrations (goose), embedded in the binary
│   ├── service/             # localisation, stats, sorting, JSON rendering
│   └── httpapi/             # router, handlers, middleware, error helpers
├── tools/gencountries/      # one-off generator for internal/geo/countries.json
├── Dockerfile               # multi-stage, distroless final image
└── go.mod
```

(The final library choices, e.g. router, pgx and the migration tool, are fixed in `plan.md`.)

## 11. Runtime and deployment

`docker-compose.yml` (repo root) defines three services:

| Service | Image / build        | Ports        | Notes |
|---------|----------------------|--------------|-------|
| `db`    | `postgres:16-alpine` | internal only| volume `pgdata`; healthcheck `pg_isready`. |
| `api`   | `./backend`          | internal `8080` | depends on `db` healthy; command `api serve`; healthcheck `api healthcheck`. No content is mounted. |
| `web`   | `./frontend`         | host `80:80`, `443:443` | nginx serves the SPA over HTTPS and proxies `/api` to `https://api:8080`; port 80 redirects to HTTPS (see frontend spec). |
| `certbot` | `certbot/certbot`  | none         | gets and renews the Let's Encrypt certificate for `DOMAIN` (webroot challenge); idle when `DOMAIN` is empty. |

Secrets (`POSTGRES_PASSWORD`, etc.) come from `.env` (not committed); `.env.example` is committed.
Visitors get HTTPS: a Let's Encrypt certificate for `DOMAIN` (from `.env`), or a self-signed fallback until it is issued.

## 12. Testing and acceptance

| Area | Test |
|------|------|
| Language negotiation | Table tests: no header → `en`; `ru` → `ru`; `ru-RU,ru;q=0.9,en;q=0.8` → `ru`; `en;q=0.5, ru;q=0.8` → `ru`; `de-DE` → `en`; malformed header → `en`. The response has `Content-Language` and `Vary: Accept-Language`. |
| Service logic | Unit tests for stats, localisation, paragraph splitting, sorting and the photo null fallbacks. |
| Handlers | `httptest` tests for every endpoint: 200 with the correct shape in EN and RU (compared to golden JSON), `404`, `405`, and `Cache-Control: no-store` with fresh data on every request. |
| Store | Integration tests against PostgreSQL in Docker: insert the test fixture → read back → equality; an empty DB → `ErrNoContent`; migrating an empty DB adds no data. |
| Smoke | `docker compose up` → `curl /api/v1/health` returns `ok`; `curl -H 'Accept-Language: ru' /api/v1/main` returns Russian text. |

**Definition of done:** all FR-B items are implemented and tested, `go vet` and
`golangci-lint` are clean, and responses match [api/contract.md](../api/contract.md) exactly.

## Changelog

| Date       | Change |
|------------|--------|
| 2026-10-17 | Initial spec. |
| 2026-10-17 | `/profile` → `/main` (`main.yaml`, table `main_info`); language from `Accept-Language`; removed `/media` serving and `MEDIA_DIR`; photos are external URLs (optional thumb/size), max 30 per place; links limited to GitHub/LinkedIn/Telegram; JSON content files allowed. |
| 2026-10-17 | Email link (`kind: email`, `mailto:` URL) added back. |
| 2026-10-17 | (Plan phase clarification) Added the `api start` and `api healthcheck` commands for the shell-less container image. |
| 2026-10-17 | Implemented (B01–B21). §10 updated to the actual layout (migrations embedded under `internal/store/`). No behaviour changes. |
| 2026-10-19 | Removed the YAML content files and the `seed`, `validate` and `start` commands: PostgreSQL is the only source of data and startup never writes content. §5 and §6 now describe the tables and the data rules. Removed the response cache and ETag (`no-store`). |
