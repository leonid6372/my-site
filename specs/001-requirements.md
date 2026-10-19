# 001 — Product Requirements

Status: **Implemented**  ·  Owner: Leonid Solovev  ·  Last updated: 2026-10-17

## 1. Overview

A minimal, bilingual (EN/RU) personal website with three pages: **Main** (about me),
**Travel** and **CV**. A small Go API serves the content from PostgreSQL. The content itself
is authored as YAML files in the repository; photos are referenced by external URLs.

## 2. Goals

- G1. Present who I am in a clean, personal way.
- G2. Visualise my travels on an interactive world map, with stories and photos.
- G3. Give recruiters and colleagues a clear, readable professional CV.
- G4. Keep content easy to update by editing text files.

## 3. Non-goals

- Admin UI, authentication, comments, contact form.
- CV PDF download.
- Projects/portfolio and certifications sections on the CV.
- HTTPS, a custom domain, analytics, SEO work.
- Uploading or hosting photo files on the server (photos are external URLs).

## 4. Audience

| Persona           | Need                                                          |
|-------------------|---------------------------------------------------------------|
| Recruiter / HR    | Quickly understand my experience, skills and how to reach me. |
| Friend / visitor  | Learn about me and browse my travels.                         |
| Me (site owner)   | Update content by editing YAML and redeploying.               |

## 5. User stories and acceptance criteria

### 5.1 Global

**US-01 — Navigation**
As a visitor, I want a persistent header so that I can move between pages.
- AC1: Given any page, when it loads, then the header shows the site name and links to
  Main, Travel and CV, and the current page link is visually marked as active.
- AC2: Given the header, when I use only the keyboard, then I can reach and activate all links.
- AC3: Given an unknown URL, when it loads, then a 404 page with a link back to Main is shown.
- AC4: Given I open the site root `/`, then the Main page is shown; the old `/main` URL redirects to `/`.

**US-02 — Language switch (EN/RU)**
As a visitor, I want to switch the language between English and Russian.
- AC1: Given my first visit, when the browser language starts with `ru`, then the site
  is shown in Russian; otherwise in English.
- AC2: Given any page, when I click the language switch, then all UI text and all content
  (bio, CV, travel stories) change language without a full page reload.
- AC3: Given I chose a language, when I come back later, then my choice is remembered.
- AC4: The `<html lang>` attribute matches the active language.
- AC5: The language is never part of the URL; the frontend sends it to the API in the
  `Accept-Language` request header.

**US-03 — Theme**
As a visitor, I want a light or dark theme.
- AC1: Given my first visit, then the theme follows the OS `prefers-color-scheme`.
- AC2: When I click the theme toggle, then the theme switches and my choice is remembered.

**US-04 — Animations**
As a visitor, I want subtle animations that make the site feel alive but calm.
- AC1: Page content fades/slides in on load and on navigation (≤ 400 ms).
- AC2: Links, buttons and cards have subtle hover/focus transitions.
- AC3: Given `prefers-reduced-motion: reduce`, then all non-essential animations are disabled.

### 5.2 Main page (`/`)

**US-10 — Learn about me**
As a visitor, I want to see who the owner is.
- AC1: The page shows: profile photo, full name, headline (one-line role), a short bio
  (1–3 paragraphs), current location, a "Currently…" line (what I am doing now),
  a list of interests/hobbies, and contact links (Email, GitHub, LinkedIn, Telegram).
- AC2: All links have accessible labels; GitHub, LinkedIn and Telegram open in a new tab;
  the email link uses `mailto:` and opens the visitor's mail client.
- AC3: The page contains call-to-action links to the Travel and CV pages.

### 5.3 Travel page (`/travel`)

**US-20 — See where I have been**
As a visitor, I want a world map that shows the countries and places I have visited.
- AC1: Visited countries are highlighted with a fill colour; other countries are neutral.
- AC2: Each visited place (city/location) is shown as a pin at its coordinates.
- AC3: The map supports zoom (buttons, scroll wheel, pinch) and pan (drag).

**US-21 — Travel statistics**
- AC1: Above the map, a stats summary shows: number of countries visited, number of
  continents visited, number of places visited, and the percentage of the world's
  countries visited (out of 195, one decimal).
- AC2: The numbers animate (count up) on first display, unless reduced motion is on.

**US-22 — Place details**
As a visitor, I want to see the details of a place.
- AC1: When I click a pin, then a details panel opens with: place name, country, visit
  date(s), a short story and a photo gallery.
- AC2: When I click a photo, then it opens enlarged (lightbox) with its caption;
  I can go to the next/previous photo and close it with `Esc`.
- AC3: The panel can be closed with a close button or `Esc`; on mobile it is a bottom sheet.
- AC4: Given a place with up to 30 photos, the panel stays responsive: photos load lazily
  as they scroll into view.
- AC5: Given a photo URL that fails to load, then a neutral placeholder is shown instead of
  a broken image.

**US-23 — Country focus**
- AC1: When I click a visited country, then the map smoothly zooms to that country and
  the panel lists the places I visited in it; clicking an item opens the place details (US-22).
- AC2: Clicking a non-visited country does nothing (no error).

**US-24 — Place list (accessible alternative)**
- AC1: Below the map there is a list of visited countries and their places, which also
  works for keyboard and screen-reader users; selecting a place opens its details (US-22).

### 5.4 CV page (`/cv`)

**US-30 — Read the CV**
As a recruiter, I want a clear résumé.
- AC1: The page shows, in this order:
  1. Header: name, headline, location, contact links (Email, GitHub, LinkedIn, Telegram).
  2. **Professional summary** (a short paragraph).
  3. **Experience**: position, company, location, period (start – end/"Present"),
     description and bullet highlights, newest first.
  4. **Education**: degree, field, institution, period, optional description, newest first.
  5. **Skills**: grouped by category (e.g. Languages, Frameworks, Tools).
  6. **Languages** (spoken): language and level (e.g. "Russian — Native", "English — C1").
- AC2: Periods are formatted for the active locale (e.g. "May 2021 – Present" /
  "май 2021 – н. в.") and show the computed duration (e.g. "3 yrs 5 mos" / "3 г. 5 мес.").
- AC3: The page is printable: the browser print view hides navigation and toggles.

## 6. Non-functional requirements

| ID      | Requirement |
|---------|-------------|
| NFR-01  | **Performance:** on a fast 4G mobile connection, Largest Contentful Paint < 2.5 s on Main and CV; the map is interactive < 4 s. |
| NFR-02  | **Responsive:** layouts work from 360 px to 1920 px wide; no horizontal scroll. |
| NFR-03  | **Browsers:** last 2 versions of Chrome, Firefox, Safari and Edge. |
| NFR-04  | **Accessibility:** semantic landmarks, keyboard operable, visible focus, alt text, WCAG AA contrast. |
| NFR-05  | **Content validation:** the seed process fails with a clear message if any required field or translation is missing, a country code is unknown, or a photo URL is not a valid `https://` URL. (Photo URLs are not fetched; their availability is not checked.) |
| NFR-06  | **Deployability:** `docker compose up -d --build` on a clean VPS starts the whole site, served over HTTP on port 80. |
| NFR-07  | **Reliability:** the API returns well-formed JSON errors; the frontend shows a friendly error state with a retry button instead of a blank page. |
| NFR-08  | **Data volume:** designed for ~15 places with **up to 30 photos each**, but must work without changes up to ~200 places. |

## 7. Assumptions (defaults chosen; change if needed)

| #  | Assumption |
|----|------------|
| A1 | The theme follows the OS setting, with a manual light/dark toggle (US-03). |
| A2 | Animations: page fade-in, page transitions, hover effects, map fly-to, stats count-up. |
| A3 | The repository starts with **placeholder seed content**; the owner replaces it with real data. |
| A4 | The owner writes both the EN and RU texts. |
| A5 | Monorepo: `frontend/`, `backend/`, `content/` and `specs/` with one `docker-compose.yml`. |
| A6 | Visit dates are stored as full dates (`YYYY-MM-DD`); a place can have several visits. |
| A7 | Photos are hosted elsewhere (any public `https://` image URL) and only referenced in YAML. Optionally, a smaller `thumbUrl` and `width`/`height` can be given per photo for faster loading. |

## 8. Decisions log (resolved questions)

| #  | Question | Decision |
|----|----------|----------|
| D1 | Which contact links? | Email, GitHub, LinkedIn, Telegram. |
| D2 | Photo volume? | Up to 30 photos per place. |
| D3 | Where are photos stored? | External URLs in YAML; nothing is uploaded to the server. |
| D4 | How is the language passed to the API? | `Accept-Language` request header, not a URL parameter. |
| D5 | Start page name? | Frontend route `/main` (`/` redirects there); API endpoint `/api/v1/main`. |

## 9. Open questions

- None at the moment.

## 10. Glossary

- **Place:** a city or location visited, with coordinates.
- **Visit:** one trip to a place, with start and end dates.
- **Visited country:** a country that has at least one place.
- **Content files:** YAML files in `content/` that are the source of all site content.
