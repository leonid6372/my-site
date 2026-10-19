# API Contract — v1

Status: **Approved**  ·  Last updated: 2026-10-17

This is the single source of truth for the HTTP interface between `frontend` and `backend`.
Both the backend and frontend specs refer to this file. Change it here first.

## 1. Conventions

- **Base path:** `/api/v1`. In production nginx proxies `/api/*` to the backend.
- **Format:** JSON, UTF-8, `camelCase` field names.
- **Methods:** `GET` only (the public API is read-only).
- **Dates:** `YYYY-MM-DD` strings. `endDate: null` means "present".
- **Images:** all image fields are **absolute external `https://` URLs** taken as-is from the
  content files. The backend never stores, proxies or serves image files.
- **Caching:** none. Every response is built from the database on each request and sent
  with `Cache-Control: no-store` (no `ETag`, no `304`).
- **CORS:** not required in production (same origin). In development the backend allows
  the origin `http://localhost:5173` and the request header `Accept-Language`.

### 1.1 Language negotiation (`Accept-Language` header)

Endpoints that return text pick the language from the standard **`Accept-Language`
request header**. There is **no** `lang` query parameter, and the language is never part
of the URL.

- Supported languages: `en`, `ru`.
- The frontend always sends exactly one value: `Accept-Language: en` or `Accept-Language: ru`.
- The backend also accepts a full browser-style header (e.g. `ru-RU,ru;q=0.9,en;q=0.8`):
  it picks the supported language with the highest `q` value, matching on the primary
  subtag (`ru-RU` → `ru`).
- Missing header, or no supported language in it → **`en`** (fallback, no error).
- Every localised response includes:
  - `Content-Language: en` or `Content-Language: ru` (the language actually used);
  - `Vary: Accept-Language` (the response depends on the language).
- The response contains strings for **one** language only (already resolved).

### 1.2 Error format

All errors use this shape:

```json
{
  "error": {
    "code": "not_found",
    "message": "resource not found"
  }
}
```

| HTTP | code             | When                                  |
|------|------------------|---------------------------------------|
| 404  | `not_found`      | unknown route or resource             |
| 405  | `method_not_allowed` | any method other than `GET`/`HEAD` |
| 500  | `internal_error` | unexpected server error (no details)  |
| 503  | `unavailable`    | database not reachable                |

## 2. Endpoints

| Method | Path              | Localised | Used by |
|--------|-------------------|-----------|---------|
| GET    | `/api/v1/health`  | no        | Docker healthcheck |
| GET    | `/api/v1/main`    | yes       | Main page, CV header |
| GET    | `/api/v1/cv`      | yes       | CV page |
| GET    | `/api/v1/travel`  | yes       | Travel page |

### 2.1 `GET /api/v1/health`

Liveness and readiness check.

`200 OK`
```json
{ "status": "ok", "db": "ok" }
```
`503` when the DB is unreachable: `{ "status": "degraded", "db": "down" }`.

---

### 2.2 `GET /api/v1/main`

Data for the Main (about me) page and the CV header.

Request:
```http
GET /api/v1/main HTTP/1.1
Accept-Language: en
```

`200 OK` (`Content-Language: en`)
```json
{
  "fullName": "Leonid Solovyev",
  "headline": "Software Engineer",
  "location": "Amsterdam, Netherlands",
  "currently": "Building a personal website with spec-driven development",
  "bio": [
    "First paragraph of the bio.",
    "Second paragraph of the bio."
  ],
  "photo": {
    "url": "https://images.example.com/me.jpg",
    "thumbUrl": null,
    "alt": "Portrait of Leonid",
    "caption": null,
    "width": 800,
    "height": 800
  },
  "interests": ["Travel", "Photography", "Hiking"],
  "links": [
    { "kind": "email",    "label": "Email",    "url": "mailto:me@example.com" },
    { "kind": "github",   "label": "GitHub",   "url": "https://github.com/example" },
    { "kind": "linkedin", "label": "LinkedIn", "url": "https://www.linkedin.com/in/example" },
    { "kind": "telegram", "label": "Telegram", "url": "https://t.me/example" }
  ]
}
```

- `bio`: array of paragraphs (plain text, no HTML).
- `links[].url`: `mailto:` for `email`, an absolute `https://` URL for the others.
- `links[].kind`: one of `email`, `github`, `linkedin`, `telegram` (the frontend uses it to pick an
  icon). Links are ordered as in the content file.

---

### 2.3 `GET /api/v1/cv`

Everything on the CV page except the header (the header comes from `/main`).

`200 OK`
```json
{
  "summary": "Professional summary paragraph.",
  "experience": [
    {
      "id": "acme-2021",
      "position": "Senior Software Engineer",
      "company": "Acme Corp",
      "companyUrl": "https://acme.example",
      "location": "Amsterdam, Netherlands",
      "startDate": "2021-05-01",
      "endDate": null,
      "projects": [
        {
          "description": "Short description of one project or area of the role.",
          "highlights": ["Did X which improved Y by 30%", "Led Z"]
        }
      ]
    }
  ],
  "education": [
    {
      "id": "msu-2015",
      "institution": "Lomonosov Moscow State University",
      "degree": "MSc",
      "field": "Computer Science",
      "startDate": "2013-09-01",
      "endDate": "2015-06-30",
      "description": null
    }
  ],
  "skills": [
    { "category": "Programming languages", "items": ["Go", "TypeScript", "SQL"] },
    { "category": "Tools",                 "items": ["Docker", "PostgreSQL", "Git"] }
  ],
  "languages": [
    { "name": "Russian", "level": "Native" },
    { "name": "English", "level": "C1" }
  ]
}
```

- `experience` and `education` are sorted by `startDate` descending.
- `skills` and `languages` keep the order defined in the content files.
- Optional fields (`companyUrl`, `description`) are `null` when absent, never omitted.
- `projects` keeps the order stored in the database; a job without projects has `[]`.

---

### 2.4 `GET /api/v1/travel`

All travel data in one response (~15 places × up to 30 photos is ~450 photo objects,
~100 KB gzipped at most, so pagination is not needed).

`200 OK`
```json
{
  "stats": {
    "countries": 12,
    "continents": 3,
    "places": 15,
    "worldPercent": 6.2
  },
  "countries": [
    {
      "iso2": "NL",
      "isoNumeric": "528",
      "name": "Netherlands",
      "continent": "europe",
      "placeIds": ["amsterdam", "rotterdam"]
    }
  ],
  "places": [
    {
      "id": "amsterdam",
      "name": "Amsterdam",
      "countryIso2": "NL",
      "lat": 52.3676,
      "lng": 4.9041,
      "story": "A short story about the trip.\n\nSecond paragraph.",
      "visits": [
        { "startDate": "2023-04-10", "endDate": "2023-04-15" }
      ],
      "photos": [
        {
          "url": "https://images.example.com/amsterdam/01.jpg",
          "thumbUrl": "https://images.example.com/amsterdam/01_small.jpg",
          "alt": null,
          "caption": "Canals at sunset",
          "width": 1600,
          "height": 1067
        }
      ]
    }
  ]
}
```

- `stats.worldPercent` = `countries / 195 * 100`, rounded to 1 decimal.
- `countries[].isoNumeric`: a 3-digit, zero-padded ISO 3166-1 numeric code. The frontend
  uses it to match features in the world TopoJSON.
- `countries[].continent`: one of `africa`, `antarctica`, `asia`, `europe`,
  `north-america`, `oceania`, `south-america` (the frontend translates the label).
- `countries` sorted by localised name; `places` sorted by latest visit, newest first.
- `visits` sorted by `startDate` descending; `photos` keep the order in the content file
  (max 30 per place).
- `story`: plain text; paragraphs are separated by a blank line (`\n\n`).

### 2.5 Photo object

Used by `main.photo` and `places[].photos[]`.

| Field      | Type            | Notes |
|------------|-----------------|-------|
| `url`      | string          | Full-size image, absolute `https://` URL. Required. |
| `thumbUrl` | string \| null  | Optional smaller version for grids. If `null`, the frontend uses `url`. |
| `alt`      | string \| null  | Optional alt text; if `null`, the frontend falls back to `caption`, then `""`. |
| `caption`  | string \| null  | Optional localised caption. |
| `width`    | number \| null  | Optional intrinsic size in px (prevents layout shift). |
| `height`   | number \| null  | Must be given together with `width` or both omitted. |

## 3. TypeScript types (for the frontend)

```ts
export type Lang = 'en' | 'ru';

export interface Photo {
  url: string; thumbUrl: string | null; alt: string | null; caption: string | null;
  width: number | null; height: number | null;
}
export interface Link { kind: 'email' | 'github' | 'linkedin' | 'telegram'; label: string; url: string; }

export interface MainInfo {
  fullName: string; headline: string; location: string; currently: string;
  bio: string[]; photo: Photo; interests: string[]; links: Link[];
}

export interface Experience {
  id: string; position: string; company: string; companyUrl: string | null; location: string;
  startDate: string; endDate: string | null; projects: Project[];
}
export interface Project { description: string | null; highlights: string[]; }
export interface Education {
  id: string; institution: string; degree: string; field: string;
  startDate: string; endDate: string | null; description: string | null;
}
export interface Cv {
  summary: string; experience: Experience[]; education: Education[];
  skills: { category: string; items: string[] }[];
  languages: { name: string; level: string }[];
}

export type Continent = 'africa'|'antarctica'|'asia'|'europe'|'north-america'|'oceania'|'south-america';
export interface Visit { startDate: string; endDate: string | null; }
export interface Place {
  id: string; name: string; countryIso2: string; lat: number; lng: number;
  story: string; visits: Visit[]; photos: Photo[];
}
export interface Country { iso2: string; isoNumeric: string; name: string; continent: Continent; placeIds: string[]; }
export interface Travel {
  stats: { countries: number; continents: number; places: number; worldPercent: number };
  countries: Country[]; places: Place[];
}

export interface ApiError { error: { code: string; message: string } }
```

## Changelog

| Date       | Change |
|------------|--------|
| 2026-10-17 | Initial contract. |
| 2026-10-17 | `/profile` → `/main`; `lang` query param → `Accept-Language` header; `/media` removed, photos are external URLs (optional `thumbUrl`, `width`, `height`); links limited to GitHub/LinkedIn/Telegram. |
| 2026-10-17 | Email link (`kind: email`, `mailto:` URL) added back. |
