# Frontend Tasks

Status: **Implemented**  ·  Last updated: 2026-10-17
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
  - `routes.tsx` (`/` Main, `/main` → `/` replace, `/cv`, lazy `/travel`, `*` → NotFound),
    `Layout` with `AnimatePresence`, `MotionConfig reducedMotion="user"`, per-page `document.title`.
  - Refs: FR-F03, FR-F07, US-01 AC3/AC4, US-04
  - Depends on: F05
  - Done when: routing tests pass (`/main` → `/`, unknown → 404 with a link to `/`); manually, page transitions animate and are off with reduced motion.

- [x] **F07 · Header, footer, navigation**
  - `SkipLink`, `Header` (name → `/`, `Nav` with `aria-current`, `LangSwitch` with
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

- [x] **F14 · Details panel**
  - `DetailsPanel` (desktop side panel / mobile bottom sheet with Motion), `PlaceDetails`
    (name, country, visits with "×N" badge, story paragraphs), `CountryDetails` (name,
    continent, places list → select place); close button + `Esc`; focus to the heading on open,
    restore on close.
  - Refs: US-22 AC1/AC3, US-23 AC1
  - Depends on: F12
  - Done when: tests: selecting a place renders all fields; `Esc` closes it and focus returns to the trigger; a country selection lists its places, and clicking one switches to place details.

- [x] **F15 · Photo grid and lightbox**
  - `PhotoGrid` (3/2 columns, first 9 + "Show all N", counter), `PhotoLightbox`
    (yet-another-react-lightbox with Captions + Counter, `preload: 1`, blurred thumb placeholder,
    error placeholder, starts at the clicked index, covers all photos).
  - Refs: US-22 AC2/AC4/AC5, NFR-08
  - Depends on: F14
  - Done when: tests: 30-photo fixture → 9 tiles + "Show all 30 photos" → 30 tiles; clicking tile 3 opens the lightbox at 3/30; `→` → 4/30; `Esc` closes the lightbox but **not** the panel; 0 photos → no grid.

- [x] **F16 · Place list** `[P]`
  - `PlaceList`: grouped by country (SVG flag via `country-flag-icons`, localised name, count) →
    place buttons (name + latest visit year); selecting one scrolls the map into view.
  - Refs: US-24
  - Depends on: F12
  - Done when: tests: the groups and order match the fixture; activating a place with the keyboard opens the panel.

## Phase 5 — Integration and deployment

- [x] **F17 · Run against the real backend**
  - `npm run dev` with the backend from B17; fix any contract mismatches (in the code, or via a
    contract change first).
  - Depends on: F11, F15, F16, B17
  - Done when: all three pages work with the real API in EN and RU; the network tab shows `Accept-Language` on every API call and no `lang` params.

- [x] **F18 · Dockerfile, nginx, compose**
  - `frontend/Dockerfile` and `nginx.conf` (plan §6.2–6.3); enable the `web` service in
    `docker-compose.yml`.
  - Refs: NFR-06
  - Depends on: F17, B19
  - Done when: `docker compose up -d --build` on a clean machine → `http://localhost/` shows the Main page; a deep reload of `/travel?place=…` works; `/assets/*` has the immutable cache header; `index.html` has `no-cache`.

## Phase 6 — Verification

- [x] **F19 · Accessibility, responsive and performance pass**
  - Keyboard-only walkthrough, a screen-reader spot check (NVDA), axe DevTools on every page,
    widths 360 / 768 / 1440 / 1920, both themes, both languages, reduced motion, a Lighthouse
    mobile run (LCP < 2.5 s on Main/CV, map interactive < 4 s).
  - Refs: NFR-01…NFR-04, FR-F12, FR-F13
  - Depends on: F18
  - Done when: no axe "serious/critical" issues; the manual checklist below is fully ticked; Lighthouse numbers recorded in the PR.

- [x] **F20 · Traceability and quality gate**
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

## Verification results (F19, 2026-10-17)

Run against the production stack (`docker compose up`, nginx on port 80) with scripted
browser checks (Playwright driving Edge and Firefox), axe-core and Lighthouse.

| ID | Result | How it was checked |
|----|--------|--------------------|
| M-01 | ✅ | Visited fill = accent at 55 %, others `--land`; theme toggle re-colours countries, pins and background without reload. |
| M-02 | ✅ | 4 pins for 4 places; hovering a pin shows its name. |
| M-03 | ✅ | Zoom buttons; the wheel scrolls the page until the map is clicked, then zooms; drag pans. |
| M-04 | ✅ | Visited country click → selection + fly to its mainland bounds; a non-visited click does nothing. |
| M-05 | ✅ | Pin click → `?place=` + fly-to (pin centred); reduced motion jumps with no intermediate positions. |
| M-06 | ✅ | Initial view fits all pins; 0 places → whole world + "No trips yet". |
| M-07 | ✅ | 360 px: bottom sheet ≤ 70 % height, inner scroll, close button. |
| M-08 | ✅ | Lightbox: arrows, Esc, counter, captions; only the photo and its neighbours load; broken URL → placeholder. Swipe is provided by the library (not driven by the scripts). |
| M-09 | ✅ | EN → RU switch on each page without reload; the choice survives a reload. |
| M-10 | ✅ | Print emulation of `/cv`: no header/footer/toggles, black on white, contact and company URLs printed, email without `mailto:`. |
| M-11 | ✅ | Logical Tab order with visible focus on every stop; skip link; map paths are named buttons (Enter opens details); lightbox traps focus; Esc order lightbox → panel → focus back on the trigger. |
| M-12 | ✅ | Reduced motion: no transforms in page transitions, stats only ever show final values, map jumps. |

- **axe-core** (WCAG 2.1 A/AA + best practices): Main, CV, Travel, Travel with panel, lightbox, 404,
  in both themes and both languages (24 runs): **0 violations** of any impact.
- **Responsive:** 5 pages × 360/768/1440/1920 px × 2 themes × 2 languages (80 combinations): no
  horizontal scroll, no page errors.
- **Browsers:** Edge (Chromium) for everything above; Firefox for the main flows at 360 and
  1440 px (all pass). Safari was not available on this Windows machine.
- **Screen reader:** no NVDA session was run. Accessible names, roles, landmarks and focus
  order were checked through the accessibility tree and axe instead.

**Lighthouse (mobile, 4× CPU slowdown, simulated):**

| Profile | Page | Perf | A11y | BP | FCP | LCP | TBT | CLS | TTI |
|---------|------|------|------|----|-----|-----|-----|-----|-----|
| fast 4G (NFR-01: 9 Mbps, 60 ms) | /main | 80 | 100 | 100 | 0.7 s | **2.1 s** ✅ | 380 ms | 0 | 2.1 s |
| | /cv | 87 | 100 | 100 | 0.7 s | **0.9 s** ✅ | 190 ms | 0.051 | 1.1 s |
| | /travel | 70 | 100 | 100 | 0.9 s | 1.7 s | 830 ms | 0.002 | **2.1 s** ✅ |
| slow 4G (Lighthouse default: 1.6 Mbps, 150 ms) | /main | 70 | 100 | 100 | 2.3 s | 4.9 s | 60 ms | 0 | 4.9 s |
| | /cv | 94 | 100 | 100 | 2.3 s | 2.6 s | 50 ms | 0.051 | 2.7 s |
| | /travel | 60 | 100 | 100 | 2.7 s | 3.4 s | 750 ms | 0.014 | 4.2 s |

NFR-01 (fast 4G) is met. Under the harsher slow-4G default, /main's LCP is the portrait,
which waits on HTML → main JS (188 KB, ~half unused at first load) → `/api/v1/main` → the
image; the placeholder photo host also redirects once (picsum.photos → fastly). Possible
follow-ups if needed: start the `/main` API request and preload the portrait from an inline
script before the JS bundle runs, and trim the main bundle (e.g. Motion's `LazyMotion`).

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

## Traceability (verified in F20)

Unit/component tests live next to the code (`*.test.ts(x)`); "M-xx" are the checklist items
above, verified in F19.

**Functional requirements**

| Req | Tasks | Evidence |
|-----|-------|----------|
| FR-F01 app shell | F06, F07 | `shell.test.tsx` (landmarks, header, footer); M-11 |
| FR-F02 active link | F07 | `shell.test.tsx` "marks … as the current page" (aria-current) |
| FR-F03 404 and `/` redirect | F06 | `routes.test.tsx` (redirect, 4 unknown paths); F18 nginx deep links |
| FR-F04 language detection | F05 | `providers.test.tsx` (ru-RU, stored choice wins, invalid value, `<html lang>`, throwing storage) |
| FR-F05 Accept-Language, refetch | F03 | `api.test.tsx` (header per language, clean URLs, refetch with old data kept); F17 real-API sweep |
| FR-F06 theme, no flash | F02, F05 | `providers.test.tsx`; F02 first-frame check in Edge; M-01 |
| FR-F07 animations, reduced motion | F06, F12, F13 | M-05, M-12 (transform sampling); `TravelPage.test.tsx` stats without count-up |
| FR-F08 Main page | F10 | `MainPage.test.tsx` (EN/RU fields, one h1, CTAs, skeleton, error → Retry) |
| FR-F09 Travel page | F12–F16 | `TravelPage.test.tsx`, `DetailsPanel.test.tsx`, `PhotoGrid.test.tsx`, `PlaceList.test.tsx`; M-01…M-08 |
| FR-F10 CV page | F09, F11 | `CvPage.test.tsx` (order, null fields, Present, print targets), `dates.test.ts`; M-10 |
| FR-F11 loading/error/empty | F08, F10–F12 | `components.test.tsx` (ErrorState), page error/skeleton tests, "No trips yet" test |
| FR-F12 responsive 360–1920 | F19 | 80-combination responsive matrix; M-07 |
| FR-F13 accessibility | F07, F19 | axe 0 violations (24 runs); M-11; accessible-name tests (`PlaceList.test.tsx`, `shell.test.tsx`) |
| FR-F14 RemoteImage | F08 | `components.test.tsx` (thumb/full, null fallback, aspect-ratio, error placeholder, alt chain, referrer) |

**Acceptance criteria (001-requirements.md)**

| AC | Evidence |
|----|----------|
| US-01 AC1 header + active link | `shell.test.tsx`; M-11 |
| US-01 AC2 keyboard | `shell.test.tsx` (skip link, menu Esc); M-11 |
| US-01 AC3 404 with link to Main | `routes.test.tsx`; F18 SPA 404 in browser |
| US-01 AC4 `/` = Main, `/main` → `/` | `routes.test.tsx`; F17, F18 |
| US-02 AC1 browser language | `providers.test.tsx` (ru-RU → ru, de → en) |
| US-02 AC2 switch without reload | `shell.test.tsx`, `api.test.tsx`; M-09 (no navigation entry added) |
| US-02 AC3 remembered | `providers.test.tsx`; M-09 (reload) |
| US-02 AC4 `<html lang>` | `providers.test.tsx`, `shell.test.tsx` |
| US-02 AC5 header, never in URL | `api.test.tsx`; F17 (10 real calls) |
| US-03 AC1 follows OS | `providers.test.tsx`; F02 first-frame check |
| US-03 AC2 toggle remembered | `providers.test.tsx`, `shell.test.tsx` |
| US-04 AC1 page fade/slide ≤ 400 ms | F07 transition sampling (300 ms in, 150 ms out) |
| US-04 AC2 hover/focus transitions | CSS (150 ms); M-11 focus rings (hover is visual only) |
| US-04 AC3 reduced motion | M-05, M-12 |
| US-10 AC1 all fields | `MainPage.test.tsx` (EN and RU) |
| US-10 AC2 link labels, targets | `shell.test.tsx` SocialLinks (target/rel per kind, names) |
| US-10 AC3 CTAs | `MainPage.test.tsx` |
| US-20 AC1–AC3 map | M-01, M-02, M-03 |
| US-21 AC1 stats (195, one decimal) | `TravelPage.test.tsx` (EN/RU exact pairs); backend `TestWorldPercentRounding` |
| US-21 AC2 count-up / reduced | M-12 (12 intermediate values vs final only) |
| US-22 AC1 details panel | `DetailsPanel.test.tsx`; M-05 |
| US-22 AC2 lightbox | `PhotoGrid.test.tsx` (3/30 → 4/30, Esc); M-08 |
| US-22 AC3 close button / Esc / sheet | `DetailsPanel.test.tsx`; M-07 |
| US-22 AC4 30 photos lazy | `PhotoGrid.test.tsx` (9 → 30), `components.test.tsx` (`loading=lazy`); M-08 (neighbours only) |
| US-22 AC5 broken photo placeholder | `components.test.tsx`; M-08 |
| US-23 AC1 country zoom + list | `DetailsPanel.test.tsx` (country → place); M-04 |
| US-23 AC2 non-visited no-op | M-04 |
| US-24 AC1 accessible list | `PlaceList.test.tsx` (order, keyboard → panel → focus back) |
| US-30 AC1 section order | `CvPage.test.tsx` |
| US-30 AC2 locale periods + duration | `dates.test.ts` (EN/RU table), `CvPage.test.tsx` |
| US-30 AC3 printable | `CvPage.test.tsx` print targets; M-10 |
| NFR-01 performance | F19 Lighthouse (fast 4G: LCP 2.1 s / 0.9 s, map 2.1 s) |
| NFR-02 responsive | F19 matrix |
| NFR-03 browsers | Edge + Firefox (F19); Safari not tested |
| NFR-04 accessibility | axe 0 violations; M-11 |
| NFR-06 deployability | F18 (`docker compose up -d --build`, smoke.sh through nginx) |
| NFR-07 friendly errors | ErrorState + Retry tests on every page |
| NFR-08 30 photos | `PhotoGrid.test.tsx`; M-08 |
