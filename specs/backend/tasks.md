# Backend Tasks

Status: **Approved**  ·  Last updated: 2026-10-17
Implements: [backend/spec.md](spec.md) · [backend/plan.md](plan.md) · [api/contract.md](../api/contract.md)

## How to use this file

- Work top to bottom. A task can start when all of its **Depends on** tasks are done.
- `[P]` = can be done in parallel with other `[P]` tasks in the same phase.
- Every task ends with **Done when**: the check that proves it is finished. Tick the box
  only when that check passes.
- One task is roughly one commit: `feat(backend): … (B07)`.
- If a task reveals that the spec or plan is wrong, stop, update the spec or plan first,
  then continue (constitution §2).

Repo-wide setup (compose, `.env`, content) lives here because the backend owns the runtime.

---

## Phase 0 — Repository and project setup

- [x] **B01 · Repo scaffolding**
  - Add a root `.gitignore` (`.env`, `node_modules/`, `dist/`, `*.local`, coverage),
    `.gitattributes` (`* text=auto eol=lf`, so the shell scripts and YAML work in Linux containers),
    and `.env.example` (`POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`).
  - Depends on: —
  - Done when: `git status` is clean after copying `.env.example` to `.env`; `.env` is ignored.

- [x] **B02 · Go module and CLI skeleton**
  - `backend/go.mod` (module `github.com/leonid6372/my-site/backend`, pinned `go`/`toolchain`).
  - `cmd/api/main.go`: dispatch on `os.Args[1]` → `serve | seed | validate | migrate | start | healthcheck`;
    unknown/missing → usage + exit 2. Stubs return "not implemented".
  - `.golangci.yml` with the linters from plan §2.
  - Depends on: B01
  - Done when: `go build ./...`, `go vet ./...` and `golangci-lint run` pass; `go run ./cmd/api foo` prints usage and exits 2.

- [x] **B03 · Config package** `[P]`
  - `internal/config`: `Load() (Config, error)` for `HTTP_ADDR`, `DATABASE_URL`, `CONTENT_DIR`,
    `CORS_ORIGINS`, `LOG_LEVEL` with the defaults from spec §4. `DATABASE_URL` is required only
    by the commands that use the DB (`validate` does not need it).
  - Refs: spec §4
  - Depends on: B02
  - Done when: unit tests cover defaults, the missing required var error, CORS list parsing, and an invalid log level.

## Phase 1 — Domain foundations (no DB)

- [x] **B04 · Country reference data** `[P]`
  - `tools/gencountries/main.go`: download/read `mledoze/countries` JSON → write
    `internal/geo/countries.json` with the continent mapping from plan §6. Commit the output.
  - `internal/geo`: `//go:embed`, `Lookup(iso2)`, `All()`.
  - Refs: FR-B09
  - Depends on: B02
  - Done when: tests show `NL → 528, europe, "Нидерланды"`; `BR → south-america`; `MX → north-america`; `AQ → antarctica`; unknown `XX` → `false`; the file has ≥ 240 entries.

- [x] **B05 · Language negotiation** `[P]`
  - `internal/i18n`: `Lang`, `Negotiate(header) Lang` (plan §4.2), `Middleware` (puts the lang
    in the context, sets `Content-Language` + `Vary: Accept-Language`), `FromContext`.
  - Refs: FR-B05, contract §1.1
  - Depends on: B02
  - Done when: all the table cases from spec §12 pass, plus ties (`en, ru` → `en`; `ru;q=0.8, en;q=0.8` → `ru`) and `*` → `en`.

- [x] **B06 · Content model and loader**
  - `internal/content`: `LText`, `Photo`, `Link`, `MainFile`, `CVFile`, `TravelFile`, `All`.
    Date type that decodes `YYYY-MM-DD` from YAML and JSON.
  - `Load(dir) (All, []ValidationError, error)`: exactly one of `<name>.yaml|.yml|.json` per
    file; strict decoding (`KnownFields(true)` / `DisallowUnknownFields`).
  - Refs: spec §5
  - Depends on: B02
  - Done when: tests load `testdata/valid/` (YAML) and an equivalent JSON fixture into equal structs; unknown key → error with the file name; both `main.yaml` and `main.json` → "ambiguous source" error; missing file → error.

- [x] **B07 · Content validator**
  - `Validate(All) []ValidationError` implementing **every** rule of spec §6, with paths like
    `places[3].photos[2].url`. Email via `net/mail.ParseAddress`; country via `geo.Lookup`.
  - Refs: FR-B08, NFR-05
  - Depends on: B04, B06
  - Done when: table-driven tests have one broken fixture per rule (≥ 14 cases incl. 31 photos, `http://` URL, width without height, duplicate link kind, a bad `mailto:`, a missing `ru`, end < start, unknown country, lat 91), and each asserts the exact file + path; the valid fixture gives zero errors.

- [x] **B08 · Placeholder content + `api validate`**
  - `content/main.yaml`, `content/cv.yaml` (2 jobs, 1 education, 2 skill groups, 2 languages),
    `content/travel.yaml` (4 places in 3 countries on 2 continents; one place with 30 photos,
    one with 0, one with 2 visits) using `https://picsum.photos/id/<n>/1600/1067` and `/400/400` thumbs.
  - Wire `api validate`: print errors as `file:path: message`, exit 1 on errors, 0 and `content OK` otherwise.
  - Refs: A3, spec §3
  - Depends on: B07
  - Done when: `go run ./cmd/api validate` with `CONTENT_DIR=../content` prints `content OK`; breaking a `ru` field prints the right path and exits 1.

## Phase 2 — Database

- [x] **B09 · Pool, migrations, `api migrate`**
  - `internal/store`: `Open(ctx, url)` with retry/backoff up to 30 s, `MaxConns=5`;
    `migrations/00001_init.sql` = spec §7 schema (goose, embedded); `Migrate(ctx)`.
  - `docker-compose.dev.yml` publishing `db` on `localhost:5432` (needed for local runs).
  - Refs: FR-B10
  - Depends on: B03
  - Done when: `api migrate` against the dev DB creates all 9 tables; running it twice is a no-op; an integration test (`-tags=integration`, testcontainers) asserts the tables exist.

- [x] **B10 · `ReplaceAll` + `api seed`**
  - `ReplaceAll(ctx, content.All)`: one tx → `TRUNCATE … RESTART IDENTITY CASCADE` → batch
    inserts (`LText` ↔ JSONB via `Value/Scan`; `sort_order` from list position).
  - `api seed` = migrate → load → validate → ReplaceAll; exit 1 on any error.
  - Refs: FR-B07, FR-B08
  - Depends on: B08, B09
  - Done when: integration tests pass: seed the valid fixture → row counts match; seed twice → same counts (idempotent); force an insert error mid-tx → the previous data is unchanged.

- [x] **B11 · Read queries**
  - `GetMain`, `ListLinks`, `GetCV` (experience, education, skills, languages), `ListPlaces`
    (with visits and photos, no N+1: 3 queries + in-memory join).
  - Depends on: B10
  - Done when: an integration test seeds the fixture and reads back structs equal to the input (round-trip), including null optionals and photo order.

## Phase 3 — Service and HTTP API

- [x] **B12 · Service: Main and CV DTOs** `[P]`
  - DTO structs mirroring contract §2.2/§2.3 exactly (camelCase tags, `null` not omitted).
  - Localise `LText` by lang; split `bio` on blank lines; sort experience/education by
    `startDate` desc; photo null fallbacks.
  - The service depends on a `Store` interface (so it can be faked in tests).
  - Refs: FR-B02, FR-B03, FR-B06
  - Depends on: B11
  - Done when: golden tests `main.{en,ru}.json`, `cv.{en,ru}.json` match (`-update` flag regenerates them); a key-by-key check shows the shapes equal the contract examples.

- [x] **B13 · Service: Travel DTO and stats** `[P]`
  - Countries (from `geo`, localised name, sorted by localised name, `placeIds`), places sorted by
    latest visit desc, visits sorted desc, story with `\n\n`, stats incl. `worldPercent` rounding.
  - Refs: FR-B04, US-21
  - Depends on: B11
  - Done when: golden `travel.{en,ru}.json` match; unit tests for stats (0 places → all zeros, `worldPercent` 12/195 → 6.2) and the RU sort order (Cyrillic collation via `x/text/collate`).

- [x] **B14 · Response cache + ETag**
  - Per-(endpoint, lang) lazy cache: `body`, weak ETag `W/"<16 hex of sha256>"` (plan §4.3).
  - Refs: FR-B12
  - Depends on: B12, B13
  - Done when: a test with a counting fake store shows 1 store call for 100 requests per key; EN and RU ETags differ.

- [x] **B15 · HTTP helpers and middleware** `[P]`
  - `writeJSON`, `writeError` (contract §1.2 codes), request ID, recover, slog logging (incl. `lang`),
    CORS (only when configured; `Accept-Language` allowed; `OPTIONS` → 204), `gzhttp`.
  - Refs: FR-B11, FR-B13
  - Depends on: B05
  - Done when: tests show a panic → 500 JSON without details; the log line contains all the fields; preflight from `http://localhost:5173` → 204 with the right headers; another origin gets no CORS headers.

- [x] **B16 · Router and handlers**
  - Routes `GET /api/v1/{health,main,cv,travel}`; `If-None-Match` → 304; `Cache-Control`;
    unknown → 404 JSON; non-GET/HEAD → 405 JSON; middleware order per plan §4.1.
  - Refs: FR-B01…FR-B05, FR-B11, FR-B12
  - Depends on: B14, B15
  - Done when: `httptest` tests for each endpoint × {no header, `en`, `ru`, `ru-RU,ru;q=0.9`, `de`} check the status, `Content-Language`, `Vary`, `ETag`, body language; 304 on a matching ETag; 404; 405 on `POST`.

- [x] **B17 · `serve`, `start`, `healthcheck`, health endpoint**
  - `/api/v1/health`: DB ping with 1 s timeout → 200 `{status:ok,db:ok}` / 503 `{status:degraded,db:down}`.
  - `serve`: migrate → HTTP server with the timeouts from spec §9 → graceful shutdown on SIGTERM/SIGINT (10 s).
  - `start`: seed → serve (the server doesn't start if the seed fails). `healthcheck`: GET local health, exit 0/1.
  - Refs: FR-B01, spec §3, §9
  - Depends on: B10, B16
  - Done when: `go run ./cmd/api start` serves the placeholder content; `curl -H "Accept-Language: ru" localhost:8080/api/v1/main` returns Russian; stopping the DB → health 503 and `api healthcheck` exits 1; Ctrl+C shuts down cleanly.

## Phase 4 — Containers and deployment

- [x] **B18 · Backend Dockerfile**
  - Multi-stage per plan §9.1 (`ENTRYPOINT ["/api"]`, `CMD ["start"]`, `HEALTHCHECK`).
  - Depends on: B17
  - Done when: `docker build ./backend` succeeds; the image is < 30 MB; `docker run … validate` with the content mounted prints `content OK`.

- [ ] **B19 · docker-compose.yml**
  - `db`, `api`, `web` per plan §9.2 (`web` is added by F18; until then, comment it out or use a placeholder).
  - Depends on: B18
  - Done when: `docker compose up -d --build db api` → `docker compose ps` shows both as healthy; `docker compose logs api` shows the seed succeeded.

- [ ] **B20 · Smoke script and docs**
  - `scripts/smoke.sh`: wait for health, assert `/api/v1/main` EN and RU, `/api/v1/travel` has `stats`.
  - Root `README.md`: prerequisites, local dev, deploy to the VPS, the content update workflow
    (validate → restart, plan §9.3), the YAML field reference (link to spec §5).
  - Refs: NFR-06
  - Depends on: B19
  - Done when: `scripts/smoke.sh` passes against the compose stack; a new reader can follow the README from a clean clone.

## Phase 5 — Verification

- [ ] **B21 · Traceability and quality gate**
  - Check that every FR-B01…FR-B13 has at least one test (list the test names in the PR description).
  - `go vet`, `golangci-lint run`, `go test ./...`, `go test -tags=integration ./...` all pass.
  - Compare the live responses with every example in `api/contract.md` (field names, nulls, ordering).
  - Depends on: B20
  - Done when: all checks are green → set `backend/spec.md`, `plan.md` and `tasks.md` status to **Implemented**.

---

## Dependency overview

```
B01 → B02 ─┬─ B03 ──────────────── B09 ─┐
           ├─ B04 ─┐                     │
           ├─ B05 ─┼──────────── B15 ─┐  │
           └─ B06 ─┴─ B07 → B08 ──────┼──┴─ B10 → B11 ─┬─ B12 ─┐
                                      │                └─ B13 ─┴─ B14 ─┐
                                      └────────────────────────────────┴─ B16 → B17 → B18 → B19 → B20 → B21
```

## Traceability

| Requirement | Tasks |
|-------------|-------|
| FR-B01 health | B16, B17 |
| FR-B02 main | B12, B16 |
| FR-B03 cv | B12, B16 |
| FR-B04 travel | B13, B16 |
| FR-B05 Accept-Language | B05, B16 |
| FR-B06 photo pass-through | B12, B13 |
| FR-B07 seed | B10 |
| FR-B08 validation | B07, B08, B10 |
| FR-B09 country data | B04 |
| FR-B10 migrations | B09, B17 |
| FR-B11 errors | B15, B16 |
| FR-B12 ETag / cache | B14, B16 |
| FR-B13 logging | B15 |
| NFR-05 | B07 |
| NFR-06 | B18, B19, B20 |
