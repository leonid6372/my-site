# 000 — Project Constitution

> The constitution holds the non-negotiable principles and conventions of the project.
> Every spec, plan, task and line of code must follow it. Change it only deliberately,
> and record each change in the changelog at the bottom.

## 1. Purpose

`my-site` is a personal website with three pages:

1. **Main** (`/`, the start page): who I am.
2. **Travel**: an interactive map of the places I have visited.
3. **CV**: my professional résumé.

## 2. Development method: Spec-Driven Development (SDD)

1. **Spec first.** No code is written for behaviour that is not described in a spec.
2. **Workflow per change:** `specify → plan → tasks → implement → verify`.
   - `spec.md`: *what* and *why* (behaviour, acceptance criteria).
   - `plan.md`: *how* (technical design, libraries, structure).
   - `tasks.md`: ordered, small, verifiable implementation steps.
3. **Specs are the source of truth.** If code and spec disagree, the spec wins. Either
   fix the code, or update the spec first in a separate change and then change the code.
4. **Traceability.** Every requirement has an ID (`FR-…`, `NFR-…`, `US-…`). Tasks and
   tests reference these IDs.
5. **Acceptance criteria** use the *Given / When / Then* form and must be testable.

## 3. Repository layout (monorepo)

```
my-site/
├── specs/                 # all SDD documents (this folder)
│   ├── 000-constitution.md
│   ├── 001-requirements.md
│   ├── api/contract.md    # shared HTTP contract between frontend and backend
│   ├── backend/           # spec.md, plan.md, tasks.md
│   └── frontend/          # spec.md, plan.md, tasks.md
├── content/               # site content as YAML (edited by hand)
├── backend/               # Go API
├── frontend/              # React + Vite + TypeScript SPA
├── docker-compose.yml
└── .env.example
```

## 4. Technology stack (fixed)

| Layer      | Choice                                             |
|------------|----------------------------------------------------|
| Frontend   | React 18+, Vite, TypeScript (strict)               |
| Backend    | Go (latest stable), REST/JSON API                  |
| Database   | PostgreSQL 16                                      |
| Content    | YAML files in `content/`, imported into PostgreSQL |
| Runtime    | Docker + Docker Compose on a single VPS            |
| Web server | nginx (serves the SPA, proxies `/api`)             |
| Images     | External URLs only (no file uploads/storage on the server) |

Detailed library choices go into the `plan.md` files, not here.

## 5. Principles

1. **Simplicity over cleverness.** It is a small personal site. No microservices, no
   message queues, no premature abstractions.
2. **Content as code.** All content lives in version-controlled YAML files. Photos are
   **not** stored on the server; content references them by external URL. There is no
   admin UI. Content is updated by editing the files and redeploying.
3. **Bilingual parity (EN + RU).** Every user-visible text exists in both English and
   Russian. Missing translations are a **build/seed error**, not a silent fallback.
4. **Minimal design.** Plenty of whitespace, clean typography, restrained colours, and a
   small set of subtle animations. Animations must respect `prefers-reduced-motion`.
5. **Read-only public API.** The backend exposes only `GET` endpoints to the public.
   The content language is negotiated with the `Accept-Language` request header, not with
   URL parameters.
6. **No secrets in git.** Configuration comes from environment variables; `.env.example`
   documents them.
7. **Accessibility baseline.** Semantic HTML, keyboard navigation, alt texts for images,
   and sufficient colour contrast.

## 6. Conventions

- **Language of code, comments and specs:** English.
- **Git:** small commits; message style `type(scope): summary`
  (e.g. `feat(backend): add travel stats endpoint`). Types: `feat`, `fix`, `docs`,
  `refactor`, `test`, `chore`, `spec`.
- **Formatting / linting:**
  - Go: `gofmt`, `go vet`, `golangci-lint`.
  - TS: ESLint + Prettier, `tsc --noEmit` must pass.
- **Testing:**
  - Backend: `go test ./...`. Handlers are tested with `httptest`, and the repository
    layer against a real PostgreSQL in Docker.
  - Frontend: Vitest + React Testing Library for components and hooks.
  - Each acceptance criterion in a spec has at least one test, or a documented manual check.
- **API versioning:** all endpoints live under `/api/v1`. Breaking changes require a new
  version and a spec update first.

## 7. Out of scope (project-wide)

- Authentication, admin panel, user accounts.
- Analytics, SEO optimisation, comments, a contact form, a CV PDF download.

## Changelog

| Date       | Change                    |
|------------|---------------------------|
| 2026-10-17 | Initial constitution.     |
| 2026-10-17 | Photos by external URL only; language via `Accept-Language` header. |
