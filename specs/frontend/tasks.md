# Frontend Tasks

Status: **Approved**  ·  Last updated: 2026-10-17
Implements: [frontend/spec.md](spec.md) · [frontend/plan.md](plan.md) · [api/contract.md](../api/contract.md)

## How to use this file

- Work top to bottom. A task can start when all of its **Depends on** tasks are done.
- `[P]` = can be done in parallel with other `[P]` tasks in the same phase.
- Every task ends with **Done when**; tick the box only when that check passes.
- One task is roughly one commit: `feat(frontend): … (F07)`.
- The frontend does **not** need a running backend until F18: all tests and local work use
  MSW fixtures built from the contract examples. So the frontend can be built in parallel
  with the backend.
- If a task reveals that the spec or plan is wrong, update them first (constitution §2).

---

## Phase 0 — Project setup

- [x] **F01 · Scaffold**
  - `npm create vite@latest frontend -- --template react-ts`; TS `strict`; ESLint flat config
    (typescript-eslint, react-hooks, jsx-a11y) + Prettier; npm scripts from plan §6.4;
    Vite proxy `/api → http://localhost:8080`.
  - Vitest (`jsdom`) + Testing Library + `src/test/setup.ts`; MSW v2 server with handlers
    returning `src/test/fixtures/{main,cv,travel}.{en,ru}.json` (based on the contract
    examples; travel fixture includes a 30-photo place and a 0-photo place).
  - Depends on: B01
  - Done when: `npm run lint`, `typecheck`, `test` (one sample test) and `build` all pass.

- [x] **F02 · Design tokens and global styles** `[P]`
  - `styles/tokens.css` (spec §4.3, light + `[data-theme="dark"]`), `global.css` (reset,
    typography scale, focus-visible ring, reduced-motion block), `print.css`.
  - `@fontsource-variable/inter` (latin + cyrillic). Inline theme script in `index.html` (plan §4.2).
  - Refs: FR-F06, US-03, US-04 AC3
  - Depends on: F01
  - Done when: a reload with a dark OS theme shows no flash of light; toggling `data-theme` in devtools re-colours the page; Cyrillic text renders in Inter.

## Phase 1 — Core infrastructure

- [x] **F03 · API types, client, queries** `[P]`
  - `api/types.ts` (copy of contract §3), `client.ts` (`apiGet` with `Accept-Language`, `ApiError`),
    `queries.ts` (`useMain`, `useCv`, `useTravel` keyed by `[name, lang]`, `keepPreviousData`).
  - Refs: FR-F05, US-02 AC5
  - Depends on: F01
  - Done when: MSW tests assert the `accept-language` header equals the active lang, no `lang` in the URL, an error JSON → `ApiError`, and a lang change → a new request.

- [x] **F04 · i18n setup** `[P]`
  - `i18n/index.ts` (i18next, bundled resources), `en.json` / `ru.json` with all UI strings:
    nav, buttons, section headings, stats labels, continents, "Present", durations (plural keys),
    errors, 404, "Show all {{count}} photos", aria-labels.
  - Refs: spec §4.4
  - Depends on: F01
  - Done when: the key-parity test passes; the RU plural test gives "1 год / 2 года / 5 лет".

- [x] **F05 · Language and theme providers**
  - `lib/storage.ts`; `LanguageProvider` (detect → persist → `<html lang>` → `i18n.changeLanguage`);
    `ThemeProvider` (`data-theme`, persist).
  - Refs: FR-F04, FR-F06, US-02 AC1/AC3/AC4, US-03
  - Depends on: F03, F04
  - Done when: tests cover `navigator.language = 'ru-RU'` → `ru`; a stored choice wins; a toggle updates `<html>` and storage; a throwing `localStorage` doesn't crash.

- [x] **F06 · Routing, layout, page transitions**
  - `routes.tsx` (`/` → `/main` replace, `/main`, `/cv`, lazy `/travel`, `*` → NotFound),
    `Layout` with `AnimatePresence`, `MotionConfig reducedMotion="user"`, per-page `document.title`.
  - Refs: FR-F03, FR-F07, US-01 AC3/AC4, US-04
  - Depends on: F05
  - Done when: routing tests pass (`/` → `/main`, unknown → 404 with a link to `/main`); manually, page transitions animate and are off with reduced motion.

- [x] **F07 · Header, footer, navigation**
  - `SkipLink`, `Header` (name → `/main`, `Nav` with `aria-current`, `LangSwitch` with
    `aria-pressed`, `ThemeToggle` with `aria-label`), a mobile menu < 640 px, `Footer`,
    `SocialLinks` + brand SVG icons (email = lucide `Mail`; `mailto:` same tab, others `_blank` + `rel`).
  - Refs: FR-F01, FR-F02, US-01 AC1/AC2, US-10 AC2
  - Depends on: F06
  - Done when: tests: the active link has `aria-current="page"`; the lang switch changes UI text; link targets/rels are correct per kind. Manually: everything is reachable by Tab, and the mobile menu works at 360 px.

- [x] **F08 · Shared components** `[P]`
  - `Skeleton`, `ErrorState` (message + Retry → `refetch`), `Chip`, `RemoteImage` (plan §4.4).
  - Refs: FR-F11, FR-F14
  - Depends on: F02
  - Done when: RemoteImage tests cover thumb vs. full `src`, the `thumbUrl` null fallback, `aspect-ratio` set only with width/height, `error` event → placeholder, the alt chain, `referrerPolicy`; an ErrorState test shows Retry calls the handler.

- [x] **F09 · Date helpers** `[P]`
  - `lib/dates.ts`: `formatPeriod`, `formatDuration`, `formatVisit` (plan §4.5).
  - Refs: US-30 AC2
  - Depends on: F04
  - Done when: table tests pass for EN/RU, an open-ended period ("Present"/"н. в."), durations of 1, 2, 5, 11, 12, 14, 21, 25 months.

## Phase 2 — Main page

- [x] **F10 · Main page**
  - Hero (portrait `RemoteImage` `fetchPriority="high"`, name h1, headline, location), "Currently"
    badge, bio paragraphs, interests chips, `SocialLinks`, two CTA cards; section stagger;
    skeleton/error states; two-column ≥ 768 px.
  - Refs: FR-F08, US-10
  - Depends on: F07, F08
  - Done when: tests render the EN and RU fixtures (all fields present, one h1, CTA hrefs `/travel` and `/cv`); error → Retry; manually OK at 360 / 1440 px in both themes.

## Phase 3 — CV page

- [x] **F11 · CV page**
  - Header (from `useMain`), summary, experience (period + duration, company link if `companyUrl`,
    highlights), education, skills chips, languages; section order per spec §5.3; print styles
    (URLs printed, plain email address).
  - Refs: FR-F10, US-30
  - Depends on: F08, F09, F10
  - Done when: tests check the section order, a `null` `companyUrl`/`description` not rendered, "Present" for a null `endDate`, the RU fixture; the print preview hides the nav/toggles and shows the link URLs.

## Phase 4 — Travel page

- [x] **F12 · Travel page shell, selection state, stats**
  - Lazy `TravelPage` with `useTravel`; `selection` state ↔ `?place=` (plan §5.3); `StatsBar` +
    `useCountUp` (IntersectionObserver, reduced motion → instant, `Intl.NumberFormat` for percent);
    empty state "No trips yet".
  - Refs: FR-F09, US-21
  - Depends on: F08
  - Done when: tests: stats show fixture numbers (count-up mocked); `?place=amsterdam` → selected; an unknown id → removed from the URL; 0 places → the empty note.

- [x] **F13 · World map**
  - `useWorldGeo` (TopoJSON as `?url` asset → `feature()`), `TravelMap` per plan §5.2: no tiles,
    visited/land fills from tokens (re-applied on theme change), hover tooltip + highlight,
    `CircleMarker` pins with tooltips, country click → `flyToBounds` + select, pin click →
    `flyTo` + select, `fitBounds` on load, scroll-wheel zoom only when focused, `aria-label`.
  - Refs: US-20, US-23, US-04 (fly-to)
  - Depends on: F12
  - Done when: manual checklist M-01…M-06 (below) passes; the Travel chunk is separate in `vite build` output; the main bundle doesn't contain Leaflet.

- [ ] **F14 · Details panel**
  - `DetailsPanel` (desktop side panel / mobile bottom sheet with Motion), `PlaceDetails`
    (name, country, visits with "×N" badge, story paragraphs), `CountryDetails` (name,
    continent, places list → select place); close button + `Esc`; focus to the heading on open,
    restore on close.
  - Refs: US-22 AC1/AC3, US-23 AC1
  - Depends on: F12
  - Done when: tests: selecting a place renders all fields; `Esc` closes it and focus returns to the trigger; a country selection lists its places, and clicking one switches to place details.

- [ ] **F15 · Photo grid and lightbox**
  - `PhotoGrid` (3/2 columns, first 9 + "Show all N", counter), `PhotoLightbox`
    (yet-another-react-lightbox with Captions + Counter, `preload: 1`, blurred thumb placeholder,
    error placeholder, starts at the clicked index, covers all photos).
  - Refs: US-22 AC2/AC4/AC5, NFR-08
  - Depends on: F14
  - Done when: tests: 30-photo fixture → 9 tiles + "Show all 30 photos" → 30 tiles; clicking tile 3 opens the lightbox at 3/30; `→` → 4/30; `Esc` closes the lightbox but **not** the panel; 0 photos → no grid.

- [ ] **F16 · Place list** `[P]`
  - `PlaceList`: grouped by country (SVG flag via `country-flag-icons`, localised name, count) →
    place buttons (name + latest visit year); selecting one scrolls the map into view.
  - Refs: US-24
  - Depends on: F12
  - Done when: tests: the groups and order match the fixture; activating a place with the keyboard opens the panel.

## Phase 5 — Integration and deployment

- [ ] **F17 · Run against the real backend**
  - `npm run dev` with the backend from B17; fix any contract mismatches (in the code, or via a
    contract change first).
  - Depends on: F11, F15, F16, B17
  - Done when: all three pages work with the real API in EN and RU; the network tab shows `Accept-Language` on every API call and no `lang` params.

- [ ] **F18 · Dockerfile, nginx, compose**
  - `frontend/Dockerfile` and `nginx.conf` (plan §6.2–6.3); enable the `web` service in
    `docker-compose.yml`.
  - Refs: NFR-06
  - Depends on: F17, B19
  - Done when: `docker compose up -d --build` on a clean machine → `http://localhost/` redirects to `/main`; a deep reload of `/travel?place=…` works; `/assets/*` has the immutable cache header; `index.html` has `no-cache`.

## Phase 6 — Verification

- [ ] **F19 · Accessibility, responsive and performance pass**
  - Keyboard-only walkthrough, a screen-reader spot check (NVDA), axe DevTools on every page,
    widths 360 / 768 / 1440 / 1920, both themes, both languages, reduced motion, a Lighthouse
    mobile run (LCP < 2.5 s on Main/CV, map interactive < 4 s).
  - Refs: NFR-01…NFR-04, FR-F12, FR-F13
  - Depends on: F18
  - Done when: no axe "serious/critical" issues; the manual checklist below is fully ticked; Lighthouse numbers recorded in the PR.

- [ ] **F20 · Traceability and quality gate**
  - Every FR-F01…FR-F14 and every acceptance criterion in `001-requirements.md` maps to a test
    or a checklist item; `lint`, `typecheck`, `test`, `build` are green.
  - Depends on: F19
  - Done when: all checks are green → set the frontend `spec.md`, `plan.md` and `tasks.md` (and `001-requirements.md`, once B21 is also done) to **Implemented**.

---

## Manual checklist (used by F13 and F19)

| ID   | Check |
|------|-------|
| M-01 | Visited countries are filled and the others neutral, in both themes; the theme toggle re-colours the map without a reload. |
| M-02 | Each place shows a pin; hovering shows its name. |
| M-03 | Zoom with buttons, pinch (touch) and scroll (after focusing the map); pan by dragging; the page doesn't scroll-trap. |
| M-04 | Clicking a visited country flies to it and lists its places; clicking a non-visited country does nothing. |
| M-05 | Clicking a pin flies to it and opens the details; with reduced motion, the map jumps instantly. |
| M-06 | Initial view fits all places; with 0 places the whole world is shown. |
| M-07 | Mobile (360 px): the details panel is a bottom sheet, scrollable, and closable. |
| M-08 | Lightbox: swipe on touch, arrows, Esc, counter, captions; a broken photo URL shows a placeholder. |
| M-09 | Language switch on each page changes all texts without a reload; the choice survives a reload. |
| M-10 | Print preview of `/cv` is clean (no nav/toggles, URLs visible). |
| M-11 | All pages: Tab order is logical, focus is visible, "Skip to content" works. |
| M-12 | Reduced motion: no slide/fade transforms, no count-up. |

## Dependency overview

```
F01 ─┬─ F02 ─────────────── F08 ─┬────────────── F10 → F11 ─┐
     ├─ F03 ─┐                   │                           │
     └─ F04 ─┼─ F05 → F06 → F07 ─┘                           │
             └─ F09 ─────────────────────────── (F11)        │
                                 F08 → F12 ─┬─ F13           │
                                            ├─ F14 → F15 ────┼─ F17 (+B17) → F18 (+B19) → F19 → F20
                                            └─ F16 ──────────┘
```

## Traceability

| Requirement | Tasks |
|-------------|-------|
| US-01 Navigation | F06, F07 |
| US-02 Language | F03, F04, F05, F07, F17 |
| US-03 Theme | F02, F05, F07 |
| US-04 Animations | F02, F06, F10, F12, F13, F14 |
| US-10 Main page | F07, F10 |
| US-20 Map | F13 |
| US-21 Stats | F12 |
| US-22 Place details | F14, F15 |
| US-23 Country focus | F13, F14 |
| US-24 Place list | F16 |
| US-30 CV | F09, F11 |
| NFR-01…04 | F19 |
| NFR-06 | F18 |
| NFR-07 | F08 |
| NFR-08 | F15 |
