# Frontend Specification

Status: **Implemented**  ·  Last updated: 2026-10-17
Related: [Constitution](../000-constitution.md) · [Requirements](../001-requirements.md) · [API contract](../api/contract.md)

## 1. Responsibilities

A React + Vite + TypeScript single-page application that:

1. Renders the three pages (Main, Travel, CV) and a 404 page.
2. Fetches content from the backend API in the active language (sent in the
   `Accept-Language` header).
3. Provides EN/RU switching, a light/dark theme and subtle animations.
4. Renders the interactive travel map with countries, pins, stats, details and a photo lightbox.

## 2. Routes

| Path      | Page        | Data source |
|-----------|-------------|-------------|
| `/`       | —           | redirects (replace) to `/main` |
| `/main`   | Main (about me, start page) | `GET /api/v1/main` |
| `/travel` | Travel      | `GET /api/v1/travel` (+ static world TopoJSON) |
| `/cv`     | CV          | `GET /api/v1/main` + `GET /api/v1/cv` |
| `*`       | Not found   | — |

- Client-side routing with History API; nginx falls back to `index.html` for unknown paths.
- The language is **never** in the URL (no `/en/…`, no `?lang=`).
- `document.title` per page: `"<Page> — <Full name>"` (localised).
- Optional deep link: `/travel?place=<id>` opens the details of that place (and updates on selection).

## 3. Functional requirements

| ID     | Requirement | Traces to |
|--------|-------------|-----------|
| FR-F01 | App shell: header (site name, nav, language switch, theme toggle), main, footer. | US-01 |
| FR-F02 | The active nav link is marked with `aria-current="page"` and a visual indicator. | US-01 |
| FR-F03 | 404 page with a link to `/main`; `/` redirects to `/main`. | US-01 AC3, AC4 |
| FR-F04 | Language detection (stored choice → `navigator.language` → `en`), persisted to `localStorage`, `<html lang>` updated. | US-02 |
| FR-F05 | Every API request sends `Accept-Language: <active lang>`; a language change refetches content and swaps UI strings without reload. | US-02 AC2, AC5 |
| FR-F06 | Theme: stored choice → `prefers-color-scheme`; toggle persisted to `localStorage`; no flash of the wrong theme on load. | US-03 |
| FR-F07 | Animations per §7, disabled under `prefers-reduced-motion`. | US-04 |
| FR-F08 | Main page per §5.1. | US-10 |
| FR-F09 | Travel page per §5.2: stats, map, countries, pins, details panel, lightbox, list. | US-20…US-24 |
| FR-F10 | CV page per §5.3, with locale-aware dates and durations, and print styles. | US-30 |
| FR-F11 | Loading skeletons, error state with retry, and empty states for each page. | NFR-07 |
| FR-F12 | Responsive layout 360–1920 px. | NFR-02 |
| FR-F13 | Accessibility baseline (landmarks, focus management, alt text, contrast, keyboard). | NFR-04 |
| FR-F14 | External photos via a shared `<RemoteImage>` component per §4.5 (thumb/full, lazy load, aspect ratio, error placeholder). | US-22 AC4–AC5, NFR-08 |

## 4. Global UI

### 4.1 Header
- Left: site name (`fullName` or initials), linking to `/main`.
- Centre/right: nav links **Main · Travel · CV** (RU: **Главная · Путешествия · Резюме**).
- Right: language switch `EN | RU` (two buttons with `aria-pressed`) and theme toggle
  (icon button: sun/moon, with `aria-label`).
- On < 640 px: nav collapses into a menu button that opens a simple dropdown/sheet.

### 4.2 Footer
- `© <year> <fullName>` plus the social link icons.

### 4.3 Design system (minimal)
- **Typography:** one sans-serif family (e.g. Inter, self-hosted), Cyrillic subset included.
  Scale: 14 / 16 / 20 / 28 / 40 px. Body line-height 1.6.
- **Colours (CSS custom properties)**, two themes:

  | Token          | Light     | Dark      |
  |----------------|-----------|-----------|
  | `--bg`         | `#ffffff` | `#0f1115` |
  | `--surface`    | `#f6f7f9` | `#171a21` |
  | `--text`       | `#1a1d23` | `#e8eaed` |
  | `--muted`      | `#6b7280` | `#9aa0a6` |
  | `--border`     | `#e5e7eb` | `#2a2f3a` |
  | `--accent`     | `#2f6fed` | `#6c9bff` |
  | `--visited`    | accent at 55 % | accent at 55 % |
  | `--land`       | `#e9ecef` | `#232833` |

- **Spacing:** 4 px base scale; content max width 960 px (CV 800 px); generous whitespace.
- **Radius:** 8 px; shadows only on floating elements (panel, lightbox).
- Icons: one lightweight icon set (SVG), including brand icons for social links.

### 4.4 Internationalisation
- UI strings (nav labels, headings, buttons, "Present", "Currently", stats labels,
  continent names, error messages) live in `src/i18n/en.json` and `src/i18n/ru.json`.
  A unit test checks that both files have identical key sets.
- Content strings come from the API, already localised.
- Dates are formatted with `Intl.DateTimeFormat` (month + year for CV, day + month + year
  for travel). Durations ("3 yrs 5 mos" / "3 г. 5 мес.") use a small helper with Russian
  plural rules (`Intl.PluralRules`).

### 4.5 External images (`<RemoteImage>`)

All photos are external URLs from the API (no images are served by our backend).

- **Source choice:** grids and thumbnails use `thumbUrl ?? url`; the lightbox and the main-page
  portrait use `url`.
- **Layout stability:** if `width`/`height` are present, set them (CSS `aspect-ratio`);
  otherwise use a fixed container ratio (grid tiles 1:1 with `object-fit: cover`, lightbox
  `object-fit: contain`).
- **Loading:** `loading="lazy"` and `decoding="async"` everywhere except the main-page portrait
  (`fetchpriority="high"`); a `--surface` skeleton is shown until `onLoad`, then fades in.
- **Errors:** on `onError`, show a neutral placeholder (image icon on `--surface`) and keep
  the caption; never show a broken-image icon.
- **Privacy:** `referrerPolicy="no-referrer"`, so third-party hosts don't receive page URLs.
- **Alt text:** `alt ?? caption ?? ""`.

## 5. Pages

### 5.1 Main (`/main`)

Layout (desktop: two columns; mobile: stacked):

1. **Hero:** round photo (`<RemoteImage>`, `url`), `fullName` (h1),
   `headline`, location with a pin icon.
2. **"Currently"** line, styled as a subtle highlighted badge.
3. **Bio:** paragraphs from `bio[]`.
4. **Interests:** a list of chips.
5. **Links:** Email, GitHub, LinkedIn and Telegram as icon + label buttons. GitHub,
   LinkedIn and Telegram open with `target="_blank" rel="noopener noreferrer"`; the email
   link is a plain `mailto:` link (same tab).
6. **CTAs:** two cards, "See where I've travelled →" (`/travel`) and "View my CV →" (`/cv`).

### 5.2 Travel (`/travel`)

```
┌──────────────────────────────────────────────────────────┐
│  12 countries · 3 continents · 15 places · 6.2% of world  │  ← stats bar
├──────────────────────────────────────────────┬───────────┤
│                                              │  Details  │
│                 World map                    │  panel    │
│     (visited countries filled + pins)        │ (on       │
│                                              │ selection)│
├──────────────────────────────────────────────┴───────────┤
│  Countries & places list (accessible alternative)         │
└──────────────────────────────────────────────────────────┘
```

**Map**
- Map library with zoom/pan support (Leaflet via react-leaflet is the expected choice, to be
  confirmed in `plan.md`), **without a tile server**: countries are drawn from a bundled
  world TopoJSON (`world-atlas` countries-50m, features keyed by ISO numeric id) on a plain
  `--bg` background, which gives a minimal look and no external dependency.
- Country styling: visited countries `--visited` fill; other countries `--land` fill;
  `--border` strokes; hovering a visited country raises its opacity and shows a tooltip
  with the country name.
- Pins: small circular markers (accent colour, white outline) at `lat/lng`. When markers
  overlap at low zoom they stay visible (no clustering needed for ~15 places).
  Tooltip on hover = place name.
- Initial view fits all places with padding; min zoom shows the whole world; max zoom ~ 10.
- Zoom controls are visible; scroll-wheel zoom is enabled only after the map is clicked or
  focused (avoids page-scroll trapping).

**Interactions**
- Click a pin → select the place → the details panel shows the place and the map flies to it.
- Click a visited country → fly to the country's bounds; the panel shows the country name,
  continent and a list of its places (click one → place details).
- Click a non-visited country → no action.
- `Esc` or the close button closes the panel and clears the selection; the URL `?place=` is
  updated accordingly.

**Details panel (place)**
- Desktop: a right-side panel (360 px) sliding in. Mobile (< 768 px): a bottom sheet
  (max 70 % height, scrollable).
- Content: place name (h2), country name, visits list (localised date ranges; "×2" style
  badge if several), story paragraphs, and a photo grid.
- **Photo grid (up to 30 photos):** 3 columns on desktop and 2 on mobile, square tiles using
  `<RemoteImage>` with `thumbUrl ?? url`, lazy-loaded. Only the first 9 photos are rendered
  at first, with a "Show all N photos" button that reveals the rest. A photo counter
  is shown in the grid header.
- Focus moves into the panel on open and returns to the triggering element on close.

**Lightbox**
- Opens on photo click: the full image (`url`) fitted to the viewport, a caption, a counter
  "3 / 30", prev/next buttons, keyboard `←`/`→`/`Esc`, swipe on touch, and a backdrop click
  to close. Navigation covers all photos of the place, including those still hidden behind
  "Show all".
- Only the current image and its immediate neighbours (prev/next) are preloaded.
- While the full image loads, the thumbnail is shown blurred as a placeholder; on error,
  the placeholder from §4.5 is shown.
- Focus is trapped while open.

**Stats bar**
- Four stat items (number + localised label). Numbers count up from 0 over ~800 ms on
  first view (skipped under reduced motion).

**List**
- Grouped by country (localised name, SVG flag from ISO2, place count) → places
  (name + latest visit year). Each place is a button that selects it and scrolls the map
  into view.

### 5.3 CV (`/cv`)

Single column, max width 800 px, ordered sections (each with an `h2`):

1. **Header:** `fullName` (h1), `headline`, location, and Email/GitHub/LinkedIn/Telegram links
   (from `/api/v1/main`).
2. **Professional summary**
3. **Experience:** timeline-style list. Each item: position (bold), company (linked if
   `companyUrl`), location, period "May 2021 – Present · 3 yrs 5 mos", description,
   bullet highlights.
4. **Education:** degree, field, institution, period, optional description.
5. **Skills:** category label + chips.
6. **Languages:** name — level.

Print styles (`@media print`): hide header nav, toggles and footer; black text on white;
no animations; links printed with their URL for GitHub/LinkedIn/Telegram and the
plain email address (without `mailto:`) for Email.

## 6. Data fetching and state

- One typed API client (`src/api/`) using the types from [api/contract.md §3](../api/contract.md#3-typescript-types-for-the-frontend).
- The API client adds the header `Accept-Language: <lang>` (exactly `en` or `ru`) to every
  request. No language query parameter is ever sent.
- Server state via a query library (e.g. TanStack Query): query keys still include `lang`
  (`['main', lang]`, `['cv', lang]`, `['travel', lang]`) so each language is cached
  separately, even though `lang` travels in a header. `staleTime` 5 min, 1 retry.
- Changing the language keeps previous data on screen until the new data arrives (no flicker).
- UI state (language, theme) lives in a small React context; selected place/country lives
  in Travel page state, synced with the `?place=` query parameter.
- API base URL: relative `/api/v1` (Vite dev server proxies to `http://localhost:8080`).

### Loading / error / empty states

| State   | Behaviour |
|---------|-----------|
| Loading | Skeleton blocks matching the page layout (no spinners on whole pages). |
| Error   | A message ("Something went wrong" / "Что-то пошло не так") and a **Retry** button. |
| Empty   | Travel with 0 places: the map is still shown and the stats show zeros, with a "No trips yet" note. |
| 404 API | Treated as an error state. |

## 7. Animations

Implemented with a lightweight animation library (e.g. Motion / Framer Motion) or CSS
transitions. All durations ≤ 400 ms except the count-up.

| Where            | Animation |
|------------------|-----------|
| Page enter       | fade + translateY 8 px → 0, 300 ms, ease-out. |
| Route change     | outgoing page fades out 150 ms, incoming fades in. |
| Sections         | stagger children by 60 ms on first render. |
| Hover            | links: underline slide-in; cards/chips: lift 2 px + border colour, 150 ms. |
| Map              | `flyTo` / `flyToBounds` ~ 800 ms; country hover fill transition. |
| Details panel    | slide in from the right (desktop) / from the bottom (mobile), 250 ms. |
| Stats            | count-up ~ 800 ms. |
| Theme switch     | colour transition 200 ms on `background-color` and `color`. |

With `prefers-reduced-motion: reduce`: no transforms, no count-up, the map jumps instantly,
and only opacity changes ≤ 100 ms are allowed.

## 8. Accessibility

- Landmarks: `header`, `nav`, `main`, `footer`; one `h1` per page; logical heading order.
- All interactive elements are reachable by keyboard, with a visible `:focus-visible` ring.
- A "Skip to content" link as the first focusable element.
- Images: `alt` from content (profile alt, photo captions; decorative images `alt=""`).
- The map container has `aria-label`; the list (§5.2) is the accessible equivalent of the map.
- Colour contrast meets WCAG AA in both themes.

## 9. Code layout

```
frontend/
├── public/favicon.svg
├── src/
│   ├── api/                       # client, types (from contract), query hooks
│   ├── app/                       # App, routes, Layout, providers (language, theme)
│   ├── components/                # Header, Footer, SocialLinks, SkipLink, RemoteImage, Skeleton, ErrorState, Chip, icons/
│   ├── features/
│   │   ├── main/                  # MainPage
│   │   ├── travel/                # TravelPage, StatsBar, DetailsPanel, PhotoGrid, PhotoLightbox, PlaceList,
│   │   │   └── map/               #   TravelMap, world TopoJSON loader, antimeridian fix, colours
│   │   └── cv/                    # CvPage
│   ├── i18n/                      # en.json, ru.json, setup
│   ├── lib/                       # dates, motion, storage, links, photos, media query, page title
│   ├── pages/                     # NotFoundPage
│   ├── styles/                    # tokens.css, global.css, print.css
│   ├── test/                      # setup, MSW server + fixtures, render helper, map stand-in
│   └── main.tsx
├── nginx.conf
├── Dockerfile                     # build with node, serve with nginx:alpine
├── index.html                     # incl. the inline theme script
└── vite.config.ts
```

The world TopoJSON is imported from `world-atlas` as a hashed static asset (plan §5.1), not
copied to `public/`.

## 10. Build and serving

- `npm run dev`: Vite on `:5173` with a proxy `/api` → `http://localhost:8080`.
- `npm run build`: type-check + production build to `dist/`.
- Docker: multi-stage build (node → `nginx:alpine`). `nginx.conf`:
  - `location /` → `try_files $uri /index.html;`
  - `location /api/` → `proxy_pass http://api:8080;` (the `Accept-Language` header is
    forwarded unchanged).
  - hashed assets under `/assets/` → `Cache-Control: public, max-age=31536000, immutable`;
    `index.html` → `no-cache`.
  - gzip on for JS/CSS/JSON/SVG.
- A small inline script in `index.html` sets the theme class before first paint (FR-F06).

## 11. Testing and acceptance

| Area | Test |
|------|------|
| i18n | The key sets of `en.json` and `ru.json` are identical. |
| lib  | Date range and duration formatting in EN and RU (incl. plural forms 1/2/5, "Present"). |
| API client | Every request carries `Accept-Language` equal to the active language; switching the language triggers a refetch; no `lang` in any request URL. |
| Routing | `/` redirects to `/main`; unknown path → 404. |
| Components | Header active link and language/theme toggles (persistence, `<html lang>`); ErrorState retry calls refetch. |
| RemoteImage | Uses `thumbUrl` in grids and `url` in the lightbox; falls back to `url` when `thumbUrl` is null; shows the placeholder on error; `alt` fallback chain. |
| Main / CV | Render from fixture API responses; section order; `null` optional fields are not rendered. |
| Travel | Stats rendering; clicking a list item opens the panel; `Esc` closes it; `?place=` deep link opens the panel; a 30-photo place shows 9 photos and then "Show all 30"; the lightbox keyboard navigation over all 30. (Map rendering itself is checked manually.) |
| Manual checklist | Each acceptance criterion in [001-requirements.md](../001-requirements.md) is verified in Chrome and Firefox at 360 px and 1440 px, in both themes and languages, and with reduced motion. |

**Definition of done:** all FR-F items are implemented, `tsc`, ESLint and tests pass, and the
manual checklist is complete.

## Changelog

| Date       | Change |
|------------|--------|
| 2026-10-17 | Initial spec. |
| 2026-10-17 | Route `/main` (`/` redirects); `Accept-Language` header; external photos via `RemoteImage`; up to 30 photos per place; Email/GitHub/LinkedIn/Telegram links. |
| 2026-10-17 | (Plan phase clarification) Flags rendered as SVG instead of emoji, which Windows doesn't render. |
| 2026-10-17 | Implemented (F01–F20). Clarifications from implementation: RU durations use full year words ("1 год / 2 года / 5 лет", as task F04 specified) rather than "г."; map pins and visited countries are also keyboard-operable buttons; rings crossing the antimeridian are unwrapped so the map has no world-spanning bands; §9 updated to the actual layout. |
