# my-site

Personal website with three pages: **Main** (about me), **Travel** (map of visited places)
and **CV**, in English and Russian. Built with spec-driven development: the specs in
[`specs/`](specs/README.md) are the source of truth for everything in this repo.

```
content/    YAML content: the only place you edit texts, places and photo URLs
backend/    Go API: seeds content/ into PostgreSQL, serves read-only JSON (/api/v1/*)
frontend/   React + Vite SPA, served by nginx (which also proxies /api)
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
| Frontend development | Node.js 24 LTS |
| `scripts/smoke.sh` | bash and curl (Git Bash works on Windows) |

## Quick start (everything in Docker)

```sh
git clone https://github.com/leonid6372/my-site.git && cd my-site
cp .env.example .env          # then set POSTGRES_PASSWORD
docker compose up -d --build  # db, api and web; the site is on https://localhost
scripts/smoke.sh              # end-to-end check through nginx
```

Open <https://localhost> (locally the certificate is self-signed, so the browser warns once;
<http://localhost> redirects there). Only nginx (ports 80 and 443) is published. For development, add
`-f docker-compose.yml -f docker-compose.dev.yml` to also publish the database on
`127.0.0.1:5432` and the API on `https://127.0.0.1:8080` (with CORS for the Vite dev server).

## Local development

### Backend

Start only the database in Docker and run the API with Go:

```sh
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d db
cd backend
export DATABASE_URL="postgres://site:<password>@localhost:5432/site?sslmode=disable"
export CONTENT_DIR=../content CORS_ORIGINS=http://localhost:5173
go run ./cmd/api start        # seed + serve HTTPS on :8080
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
`CORS_ORIGINS` (empty), `LOG_LEVEL` (`info`), `TLS_CERT_FILE` / `TLS_KEY_FILE` (empty).

The API serves **HTTPS only**. Set `TLS_CERT_FILE` and `TLS_KEY_FILE` (PEM) to use your own
certificate; without them it generates a self-signed one at startup (valid for `localhost`,
`127.0.0.1` and `api`), so use `curl -k` against it locally.

Tests and lint, from `backend/`:

```sh
go test ./...                              # unit tests
go test -tags=integration ./...            # + PostgreSQL tests (needs Docker, uses testcontainers)
BASE_URL=https://localhost:8080 go test -tags=live -run TestLive ./internal/service   # running API vs contract
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run --build-tags=integration ./...
go test ./internal/service -update         # regenerate the golden JSON after an intended change
```

### Frontend

Vite on <http://localhost:5173> proxies `/api` to `https://localhost:8080`, so start the API
first: `go run` as above, or the dev `api` container
(`docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d db api`).

```sh
cd frontend
npm ci
npm run dev        # http://localhost:5173
```

| Command | What it does |
|---------|--------------|
| `npm run dev` | Vite dev server with hot reload |
| `npm test` | Vitest + Testing Library; the API is mocked with MSW (fixtures in `src/test/fixtures`) |
| `npm run lint` | ESLint (incl. jsx-a11y) and Prettier check (`npm run format` fixes formatting) |
| `npm run typecheck` | TypeScript, strict |
| `npm run build` | type-check + production build to `dist/` |

UI strings live in `src/i18n/en.json` and `ru.json` (a test keeps their keys identical);
all content comes from the API in the language of the `Accept-Language` header.

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
cp .env.example .env          # set a strong POSTGRES_PASSWORD, DOMAIN and ACME_EMAIL
docker compose up -d --build
docker compose ps             # db, api and web should be "healthy"
scripts/smoke.sh              # optional end-to-end check
```

The site is served over **HTTPS** on port 443; port 80 only redirects to HTTPS (and answers
Let's Encrypt checks). Only nginx is published; the API and the database are reachable
inside the compose network only.

HTTPS certificate:

- Point your domain's DNS (`A`/`AAAA` record) at the server and open ports 80 and 443.
- Set `DOMAIN` and `ACME_EMAIL` in `.env`. The `certbot` service gets a free Let's Encrypt
  certificate on first start and checks for renewal every 12 h; nginx picks up a new
  certificate within 5 minutes, no restart needed.
- Until the certificate is issued, or with `DOMAIN` empty, nginx uses a self-signed
  certificate (browsers warn). `docker compose logs certbot` shows why issuing failed.

### Updating content

Content is bind-mounted into the container, so content changes need no image rebuild
([plan §9.3](specs/backend/plan.md#93-content-update-workflow-on-the-vps)):

```sh
git pull                                  # new YAML
docker compose run --rm api validate      # fix any errors before restarting
docker compose restart api                # start re-seeds the database
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
| `bind: address already in use` on 80 or 443 | Another web server uses the port: stop it, or change `ports` of `web` (e.g. `"8081:80", "8443:443"`). |
| Browser warns about the certificate | Self-signed fallback: set `DOMAIN`/`ACME_EMAIL` in `.env` and check `docker compose logs certbot`. |
| `/api/v1/health` returns 503 | The API is up but cannot reach PostgreSQL: check `docker compose ps db`. |
