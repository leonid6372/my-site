# Backend Specification

Status: **Approved**  ·  Last updated: 2026-10-17
Related: [Constitution](../000-constitution.md) · [Requirements](../001-requirements.md) · [API contract](../api/contract.md)

## 1. Responsibilities

The backend is a single Go service (`api`) that:

1. **Imports content** from YAML files in `content/` into PostgreSQL (the `seed` command),
   validating it.
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
| FR-B06 | Return photo objects with external URLs exactly as given in the content (no download, proxying or rewriting). | US-10, US-22 |
| FR-B07 | Provide a `seed` command that loads all content files into the DB in a single transaction (full replace). | G4 |
| FR-B08 | The seed command validates content and fails (non-zero exit, no DB changes) on any error. See §6. | NFR-05 |
| FR-B09 | Provide country reference data (ISO2, ISO numeric, continent, EN/RU names) embedded in the binary. | US-21, US-23 |
| FR-B10 | Apply DB migrations automatically on startup (`serve`) and before `seed`. | NFR-06 |
| FR-B11 | Return JSON errors as defined in the contract; never leak internal error details. | NFR-07 |
| FR-B12 | Set `ETag` + `Cache-Control` on API responses and support `If-None-Match` (ETag differs per language). | NFR-01 |
| FR-B13 | Log every request as structured JSON (method, path, status, duration, resolved language). | — |

## 3. Commands

The binary `api` has subcommands:

| Command        | Behaviour |
|----------------|-----------|
| `api serve`    | Run migrations, then start the HTTP server. |
| `api seed`     | Run migrations, validate `content/`, replace all content in the DB in one transaction. |
| `api validate` | Validate `content/` only (no DB access). Usable in CI. |
| `api migrate`  | Run migrations only. |
| `api start`    | `seed`, then `serve` in the same process (the container's default command). |
| `api healthcheck` | Calls the local `/api/v1/health`; exit 0 if healthy (used by the Docker `HEALTHCHECK`, since the image has no shell or curl). |

In Docker the container runs `api start`, so every start re-imports the current content.

## 4. Configuration (environment variables)

| Variable         | Default            | Description |
|------------------|--------------------|-------------|
| `HTTP_ADDR`      | `:8080`            | Listen address. |
| `DATABASE_URL`   | — (required)       | e.g. `postgres://site:secret@db:5432/site?sslmode=disable` |
| `CONTENT_DIR`    | `/app/content`     | Path to the YAML content files. |
| `CORS_ORIGINS`   | empty              | Comma-separated allowed origins (dev: `http://localhost:5173`). |
| `LOG_LEVEL`      | `info`             | `debug`, `info`, `warn`, `error`. |

Startup fails fast with a clear message if a required variable is missing.

## 5. Content files (source of truth)

All texts that are shown to visitors are localised objects: `{ en: "...", ru: "..." }`.
Both keys are **required** and must be non-empty.

```
content/
├── main.yaml      # main page info (about me) + social links
├── cv.yaml
└── travel.yaml
```

Content may also be written as `.json` with the same structure (`main.json`, …). If both a
`.yaml` and a `.json` file exist for the same name, validation fails (ambiguous source).

### 5.1 Photo object (used in `main.yaml` and `travel.yaml`)

```yaml
url: "https://images.example.com/amsterdam/01.jpg"         # required, absolute https URL
thumbUrl: "https://images.example.com/amsterdam/01_s.jpg"  # optional, smaller version
alt: { en: "...", ru: "..." }                              # optional
caption: { en: "...", ru: "..." }                          # optional
width: 1600                                                # optional, with height
height: 1067                                               # optional, with width
```

Photos can be hosted anywhere public (e.g. an image hosting service, a cloud bucket, a CDN).
The backend never fetches them.

### 5.2 `main.yaml`

```yaml
fullName: { en: "Leonid Solovev", ru: "Леонид Соловьёв" }
headline: { en: "Software Engineer", ru: "Инженер-программист" }
location: { en: "Amsterdam, Netherlands", ru: "Амстердам, Нидерланды" }
currently: { en: "Building my personal site", ru: "Делаю свой личный сайт" }
bio:
  en: |
    First paragraph.

    Second paragraph.
  ru: |
    Первый абзац.

    Второй абзац.
photo:
  url: "https://images.example.com/me.jpg"
  alt: { en: "Portrait of Leonid", ru: "Портрет Леонида" }
  width: 800
  height: 800
interests:
  - { en: "Travel", ru: "Путешествия" }
  - { en: "Photography", ru: "Фотография" }
links:
  - { kind: email,    label: { en: "Email",    ru: "Почта" },    url: "mailto:me@example.com" }
  - { kind: github,   label: { en: "GitHub",   ru: "GitHub" },   url: "https://github.com/example" }
  - { kind: linkedin, label: { en: "LinkedIn", ru: "LinkedIn" }, url: "https://www.linkedin.com/in/example" }
  - { kind: telegram, label: { en: "Telegram", ru: "Telegram" }, url: "https://t.me/example" }
```

`bio` is split into paragraphs by blank lines when it is served.

### 5.3 `cv.yaml`

```yaml
summary: { en: "...", ru: "..." }
experience:
  - id: acme-2021
    position: { en: "Senior Software Engineer", ru: "Старший инженер-программист" }
    company: "Acme Corp"                 # not localised
    companyUrl: "https://acme.example"   # optional
    location: { en: "Amsterdam, Netherlands", ru: "Амстердам, Нидерланды" }
    startDate: 2021-05-01
    endDate: null                        # null = present
    description: { en: "...", ru: "..." }   # optional
    highlights:
      - { en: "Did X", ru: "Сделал X" }
education:
  - id: msu-2015
    institution: { en: "Lomonosov Moscow State University", ru: "МГУ им. М. В. Ломоносова" }
    degree: { en: "MSc", ru: "Магистр" }
    field: { en: "Computer Science", ru: "Информатика" }
    startDate: 2013-09-01
    endDate: 2015-06-30
skills:
  - category: { en: "Programming languages", ru: "Языки программирования" }
    items: ["Go", "TypeScript", "SQL"]      # not localised
languages:
  - name: { en: "Russian", ru: "Русский" }
    level: { en: "Native", ru: "Родной" }
```

### 5.4 `travel.yaml`

```yaml
places:
  - id: amsterdam                # unique slug, [a-z0-9-]+
    country: NL                  # ISO 3166-1 alpha-2
    name: { en: "Amsterdam", ru: "Амстердам" }
    lat: 52.3676
    lng: 4.9041
    story:
      en: |
        Short story.
      ru: |
        Короткая история.
    visits:
      - { start: 2023-04-10, end: 2023-04-15 }
    photos:                      # 0..30 photo objects (§5.1)
      - url: "https://images.example.com/amsterdam/01.jpg"
        thumbUrl: "https://images.example.com/amsterdam/01_s.jpg"
        caption: { en: "Canals at sunset", ru: "Каналы на закате" }
      - url: "https://images.example.com/amsterdam/02.jpg"
```

## 6. Validation rules (seed / validate)

Validation collects **all** errors and prints them as `file:path: message`
(e.g. `travel.yaml:places[3].photos[2].url: must be an absolute https URL`) before it exits
with code 1.

| Rule | Applies to |
|------|-----------|
| Required fields are present and non-empty. | all |
| Every localised object has non-empty `en` **and** `ru`. | all |
| `id` values are unique within their list and match `^[a-z0-9-]+$`. | experience, education, places |
| Dates are valid `YYYY-MM-DD`; `end ≥ start`; `end` may be null only for experience/education. | cv, travel |
| `country` is a known ISO 3166-1 alpha-2 code from the embedded reference data. | travel |
| `lat ∈ [-90, 90]`, `lng ∈ [-180, 180]`. | travel |
| Each place has ≥ 1 visit and ≤ 30 photos. | travel |
| Photo `url` and `thumbUrl` are syntactically valid absolute `https://` URLs (not fetched). | main, travel |
| Photo `width` and `height` are both given or both omitted; if given, positive integers. | main, travel |
| `links[].kind` is one of `email`, `github`, `linkedin`, `telegram`; each kind appears at most once. For `email` the `url` is `mailto:<valid address>`; for the others it is an absolute `https://` URL. | main |
| No unknown keys in YAML/JSON (catches typos). | all |

## 7. Data model (PostgreSQL)

Localised text is stored as `JSONB` objects `{"en": "...", "ru": "..."}` (type alias **ltext**).
The seed command replaces all rows in one transaction (`TRUNCATE … RESTART IDENTITY CASCADE`,
then insert).

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
- **ETag:** a weak ETag computed from the SHA-256 of the response body (so it naturally
  differs per language).

## 9. HTTP server behaviour

- Router mounts `/api/v1/*`; everything else → `404 not_found` JSON. Non-`GET`/`HEAD` → `405`.
- Middleware order: request ID → recover (panic → 500) → structured logging → CORS (if
  configured; allows the `Accept-Language` header) → language negotiation → gzip.
- Timeouts: read 5 s, write 10 s, idle 60 s. Graceful shutdown on SIGTERM (max 10 s).

## 10. Proposed code layout

```
backend/
├── cmd/api/main.go          # subcommands: serve, seed, validate, migrate
├── internal/
│   ├── config/              # env parsing
│   ├── content/             # YAML/JSON structs, loader, validator
│   ├── geo/                 # embedded countries.json + lookup
│   ├── i18n/                # Accept-Language parsing + negotiation middleware
│   ├── store/               # PostgreSQL repository (pgx) + migrations (embedded)
│   ├── service/             # localisation, stats, sorting
│   └── httpapi/             # router, handlers, middleware, error helpers
├── migrations/              # SQL migration files
├── Dockerfile               # multi-stage, distroless/alpine final image
└── go.mod
```

(The final library choices, e.g. router, pgx and the migration tool, are fixed in `plan.md`.)

## 11. Runtime and deployment

`docker-compose.yml` (repo root) defines three services:

| Service | Image / build        | Ports        | Notes |
|---------|----------------------|--------------|-------|
| `db`    | `postgres:16-alpine` | internal only| volume `pgdata`; healthcheck `pg_isready`. |
| `api`   | `./backend`          | internal `8080` | depends on `db` healthy; mounts `./content` read-only at `/app/content`; command `api start`; healthcheck `api healthcheck`. |
| `web`   | `./frontend`         | host `80:80` | nginx serves the SPA and proxies `/api` to `api:8080` (see frontend spec). |

Secrets (`POSTGRES_PASSWORD`, etc.) come from `.env` (not committed); `.env.example` is committed.
The site is served over plain HTTP on the VPS IP (no TLS, by decision).

## 12. Testing and acceptance

| Area | Test |
|------|------|
| Content validator | Table-driven unit tests: one valid fixture plus one fixture per validation rule (§6), including > 30 photos, a non-https URL, width without height, an unknown link kind, and a duplicate yaml+json source. |
| Language negotiation | Table tests: no header → `en`; `ru` → `ru`; `ru-RU,ru;q=0.9,en;q=0.8` → `ru`; `en;q=0.5, ru;q=0.8` → `ru`; `de-DE` → `en`; malformed header → `en`. The response has `Content-Language` and `Vary: Accept-Language`. |
| Service logic | Unit tests for stats, localisation, paragraph splitting, sorting and the photo null fallbacks. |
| Handlers | `httptest` tests for every endpoint: 200 with the correct shape in EN and RU (compared to golden JSON), `404`, `405`, `304` with ETag, and a different ETag per language. |
| Store | Integration tests against PostgreSQL in Docker: seed → read back → equality; the seed is idempotent; a failed seed leaves the previous data untouched. |
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
