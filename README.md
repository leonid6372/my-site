# my-site

Personal website with three pages: **Main** (about me), **Travel** (map of visited places)
and **CV**, in English and Russian. Built with spec-driven development: the specs in
[`specs/`](specs/README.md) are the source of truth for everything in this repo.

```
content/    YAML content: the only place you edit texts, places and photo URLs
backend/    Go API: seeds content/ into PostgreSQL, serves read-only JSON (/api/v1/*)
frontend/   React + Vite SPA (in progress, see specs/frontend/tasks.md)
specs/      constitution, requirements, API contract, specs, plans, tasks
scripts/    smoke.sh: end-to-end check of a running stack
```

How it fits together: on every start the API validates `content/`, replaces the database
content in one transaction and then serves it. Photos are external `https://` URLs; the
server never stores image files.

## Prerequisites

| For | You need |
|-----|----------|
| Running the site (local or VPS) | Docker with the Compose plugin (`docker compose version`) |
| Backend development | Go 1.26 |
| Frontend development | Node.js 24 LTS (frontend in progress) |
| `scripts/smoke.sh` | bash and curl (Git Bash works on Windows) |

## Quick start (everything in Docker)

```sh
git clone https://github.com/leonid6372/my-site.git && cd my-site
cp .env.example .env          # then set POSTGRES_PASSWORD
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d --build db api
BASE_URL=http://localhost:8080 scripts/smoke.sh
curl -H "Accept-Language: ru" http://localhost:8080/api/v1/main
```

`docker-compose.dev.yml` publishes the database on `127.0.0.1:5432` and the API on
`127.0.0.1:8080` (with CORS for the Vite dev server). Without it, nothing but the web
server is reachable from outside.

## Local development

### Backend

Start only the database in Docker and run the API with Go:

```sh
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d db
cd backend
export DATABASE_URL="postgres://site:<password>@localhost:5432/site?sslmode=disable"
export CONTENT_DIR=../content CORS_ORIGINS=http://localhost:5173
go run ./cmd/api start        # seed + serve on :8080
```

(PowerShell: `$env:DATABASE_URL = "..."` etc.) If the `api` container is running, stop it
first (`docker compose stop api`): both use port 8080.

| Command | What it does |
|---------|--------------|
| `api start` | migrate → validate → seed → serve (the container default) |
| `api serve` | migrate → serve, without re-seeding |
| `api seed` | migrate → validate → replace all content |
| `api validate` | check `content/` only, no database needed |
| `api migrate` | apply database migrations |
| `api healthcheck` | exit 0 if the local `/api/v1/health` returns 200 |

Configuration (environment variables, [spec §4](specs/backend/spec.md#4-configuration-environment-variables)):
`DATABASE_URL` (required), `HTTP_ADDR` (`:8080`), `CONTENT_DIR` (`/app/content`),
`CORS_ORIGINS` (empty), `LOG_LEVEL` (`info`).

Tests and lint, from `backend/`:

```sh
go test ./...                              # unit tests
go test -tags=integration ./...            # + PostgreSQL tests (needs Docker, uses testcontainers)
BASE_URL=http://localhost:8080 go test -tags=live -run TestLive ./internal/service   # running API vs contract
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run --build-tags=integration ./...
go test ./internal/service -update         # regenerate the golden JSON after an intended change
```

### Frontend

In progress: the setup will be documented here with frontend task F01. It will talk to the
API at `http://localhost:8080` (either `go run` or the dev `api` container above).

## Editing content

All content lives in three files in [`content/`](content/): `main.yaml`, `cv.yaml` and
`travel.yaml` (`.json` with the same structure also works, but not both for one file).
The field reference is in [backend spec §5](specs/backend/spec.md#5-content-files-source-of-truth)
and the rules in [§6](specs/backend/spec.md#6-validation-rules-seed--validate). The essentials:

- Every visible text is `{ en: "...", ru: "..." }`; both languages are required.
- Dates are `YYYY-MM-DD`. A CV entry with `endDate: null` means "present".
- Photos are `https://` URLs (`thumbUrl`, `alt`, `caption`, `width`+`height` optional),
  up to 30 per place.
- Places use ISO 3166-1 alpha-2 country codes (`NL`, `JP`, …) and need `lat`/`lng`.
- Paragraphs are separated by a blank line.

Check your edits before deploying; errors show the exact location:

```sh
docker compose run --rm api validate
# travel.yaml:places[3].country: unknown ISO 3166-1 alpha-2 code "ZZ"
```

## Deploying to a VPS

One-time setup on the server (any Linux with Docker and the Compose plugin):

```sh
git clone https://github.com/leonid6372/my-site.git && cd my-site
cp .env.example .env          # set a strong POSTGRES_PASSWORD
docker compose up -d --build
docker compose ps             # db and api should be "healthy"
```

The site is plain HTTP for now (no domain or TLS yet). Until the frontend's `web` service
exists (task F18), the production compose file starts `db` and `api` only, and no port is
published.

### Updating content

Content is bind-mounted into the container, so content changes need no image rebuild
([plan §9.3](specs/backend/plan.md#93-content-update-workflow-on-the-vps)):

```sh
git pull                                  # new YAML
docker compose run --rm api validate      # fix any errors before restarting
docker compose restart api                # start re-seeds and clears the response cache
```

If a bad edit slips through, `api start` refuses to serve and the container restarts in a
loop. The database keeps the previous content (the seed is transactional): revert the YAML
and restart again. `docker compose logs api` shows the reason.

### Updating code

```sh
git pull && docker compose up -d --build
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `required variable POSTGRES_DB is missing a value` | Create `.env` from `.env.example` (in the repo root). |
| `api` keeps restarting | `docker compose logs api`: usually a content validation error. |
| `bind: address already in use` on 8080 | Stop the other API (`docker compose stop api` or the `go run`). |
| `/api/v1/health` returns 503 | The API is up but cannot reach PostgreSQL: check `docker compose ps db`. |
