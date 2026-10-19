# Frontend Implementation Plan

Status: **Implemented**  ·  Last updated: 2026-10-17
Implements: [frontend/spec.md](spec.md) · [api/contract.md](../api/contract.md)

This document describes **how** the frontend is built. Behaviour is defined in the spec;
if this plan contradicts the spec, the spec wins.

## 1. Constitution check

| Principle | How the plan complies |
|-----------|-----------------------|
| Simplicity | A Vite SPA, CSS Modules + CSS variables, a few focused libraries, no global state library. |
| Bilingual parity | `en.json`/`ru.json` key-parity test; content is always fetched per language. |
| Minimal design + subtle animation | Design tokens in one file; Motion only for enter/exit transitions; `prefers-reduced-motion` handled centrally. |
| Accessibility | Semantic markup, focus management in the panel/lightbox, a list alternative to the map. |
| Testing | Vitest + Testing Library + MSW; manual checklist for the map. |

## 2. Technology choices

| Concern | Choice | Why |
|---------|--------|-----|
| Runtime / tooling | Node.js active LTS, npm, Vite (latest), TypeScript `strict` | Fixed by the constitution; npm needs no extra install. |
| UI | React 19 | Current major; supported by all chosen libraries. |
| Routing | React Router v7 (library / "declarative" mode: `BrowserRouter`, `Routes`) | Standard; easy redirects and lazy routes. |
| Server state | TanStack Query v5 | Caching per `[key, lang]`, `placeholderData: keepPreviousData` for flicker-free language switches. |
| i18n (UI strings) | `i18next` + `react-i18next`, resources bundled (no HTTP backend) | Mature, handles RU plurals (`_one/_few/_many`). |
| Map | `leaflet` 1.9 + `react-leaflet` v5, **no tile layer** | Zoom/pan/markers/fly-to out of the box; vector countries only. |
| World shapes | `world-atlas` (`countries-50m.json`, TopoJSON) + `topojson-client` (`feature()`) | ~750 KB raw / ~230 KB gzipped; ids are ISO numeric codes. |
| Animations | `motion` (`motion/react`): `AnimatePresence`, `motion.div`, `useReducedMotion` | Page transitions and panel/sheet slide; CSS for hover. |
| Lightbox | `yet-another-react-lightbox` (+ `Captions`, `Counter` plugins) | Keyboard, swipe, focus trap and neighbour preloading (`carousel.preload: 1`) built in. |
| Icons | `lucide-react` for UI icons; 3 hand-made inline SVG components for GitHub, LinkedIn and Telegram | Lucide has no brand icons; inline SVGs keep it dependency-free. |
| Flags | `country-flag-icons` (per-country SVG React components) | Windows doesn't render flag emoji (spec clarified, see §9). |
| Font | `@fontsource-variable/inter` (latin + cyrillic subsets) | Self-hosted, no Google Fonts request. |
| Styling | CSS Modules + `styles/tokens.css` (custom properties) | Zero runtime, scoped, simple. |
| Testing | Vitest, `@testing-library/react`, `@testing-library/user-event`, `jsdom`, MSW v2 | MSW lets tests assert the `Accept-Language` header. |
| Lint / format | ESLint (flat config) + `typescript-eslint` + `eslint-plugin-react-hooks` + `jsx-a11y`; Prettier | Constitution requirement. |

## 3. Application structure

```
src/
├── main.tsx                     # mount, import fonts + global CSS
├── app/
│   ├── App.tsx                  # providers + <AppRoutes/>
│   ├── routes.tsx               # route table (lazy pages, /main → / redirect)
│   ├── Layout.tsx               # Header, <main id="content">, Footer, page transition
│   └── providers/
│       ├── QueryProvider.tsx
│       ├── LanguageProvider.tsx # lang state, localStorage, <html lang>, i18next.changeLanguage
│       └── ThemeProvider.tsx    # theme state, localStorage, data-theme on <html>
├── api/
│   ├── types.ts                 # copied 1:1 from api/contract.md §3
│   ├── client.ts                # apiGet<T>(path, lang): fetch + Accept-Language + ApiError
│   └── queries.ts               # useMain(), useCv(), useTravel() → useQuery([..., lang])
├── i18n/ { index.ts, en.json, ru.json }
├── lib/
│   ├── dates.ts                 # formatPeriod, formatDuration, formatVisit (Intl)
│   ├── motion.ts                # shared variants + reduced-motion helpers
│   └── storage.ts               # safe localStorage get/set (try/catch)
├── components/                  # Header, Nav, LangSwitch, ThemeToggle, Footer, SocialLinks,
│                                # RemoteImage, Skeleton, ErrorState, Chip, SkipLink, icons/
├── features/
│   ├── main/MainPage.tsx (+ Hero, Bio, Interests, CtaCards)
│   ├── cv/CvPage.tsx (+ CvHeader, Experience, Education, Skills, Languages)
│   └── travel/
│       ├── TravelPage.tsx       # data, selection state ↔ ?place=
│       ├── StatsBar.tsx, useCountUp.ts
│       ├── map/TravelMap.tsx, useWorldGeo.ts, styles.ts
│       ├── DetailsPanel.tsx     # side panel / bottom sheet; PlaceDetails | CountryDetails
│       ├── PhotoGrid.tsx        # first 9 + "Show all N"
│       ├── PhotoLightbox.tsx    # wraps yet-another-react-lightbox
│       └── PlaceList.tsx
├── pages/NotFoundPage.tsx
└── styles/ { tokens.css, global.css, print.css }
```

## 4. Cross-cutting implementation details

### 4.1 API client

```ts
export async function apiGet<T>(path: string, lang: Lang, signal?: AbortSignal): Promise<T> {
  const res = await fetch(`/api/v1${path}`, { headers: { 'Accept-Language': lang }, signal });
  if (!res.ok) throw await toApiError(res);   // parses { error: { code, message } } or falls back
  return res.json() as Promise<T>;
}
// queries.ts
export const useMain = () => {
  const { lang } = useLanguage();
  return useQuery({ queryKey: ['main', lang], queryFn: ({ signal }) => apiGet<MainInfo>('/main', lang, signal),
                    staleTime: 0, gcTime: 0, retry: 1, placeholderData: keepPreviousData });
};
```

`Accept-Language` is CORS-safelisted, so no preflight is needed even in development.

### 4.2 Language and theme

- **Initial language:** `storage.get('lang')` → `navigator.language.startsWith('ru') ? 'ru' : 'en'`.
- `LanguageProvider` sets `document.documentElement.lang`, calls `i18n.changeLanguage`,
  and persists the choice. Query keys change → TanStack Query fetches the new language
  while keeping the old data visible.
- **Theme without flash:** an inline script in `index.html` (before CSS) reads
  `localStorage.theme`, falls back to `matchMedia('(prefers-color-scheme: dark)')`, and sets
  `<html data-theme="…">`. `tokens.css` defines `:root[data-theme="dark"] { … }`.
  `ThemeProvider` only toggles that attribute afterwards.

### 4.3 Animations

- `Layout` wraps the route outlet in `<AnimatePresence mode="wait">` keyed by `location.pathname`:
  enter `{opacity:0,y:8} → {opacity:1,y:0}` 300 ms; exit `opacity → 0` 150 ms.
- Section stagger: a `motion` container with `staggerChildren: 0.06`.
- Hover effects: pure CSS transitions (150 ms) in component modules.
- Reduced motion: `<MotionConfig reducedMotion="user">` at the root, plus a CSS block
  `@media (prefers-reduced-motion: reduce) { *{transition-duration:.01ms!important;animation:none!important} }`.
  `useCountUp` and the map `flyTo` check `useReducedMotion()` and jump directly instead.

### 4.4 `RemoteImage`

```tsx
<RemoteImage photo={p} variant="thumb" | "full" | "portrait" />
```
- `src = variant === 'thumb' ? (p.thumbUrl ?? p.url) : p.url`.
- The wrapper sets `aspect-ratio` from `width/height` when present, else 1/1 (thumb/portrait).
- State machine `loading → loaded | error`; skeleton background while loading; fade-in on
  load; placeholder (lucide `ImageOff`) on error.
- Attributes: `loading="lazy"` (except portrait: `fetchPriority="high"`), `decoding="async"`,
  `referrerPolicy="no-referrer"`, `alt = p.alt ?? p.caption ?? ''`.

### 4.5 Dates

`lib/dates.ts` uses `Intl.DateTimeFormat(lang, { month: 'short', year: 'numeric' })` for CV
periods and `{ day: 'numeric', month: 'short', year: 'numeric' }` for visits. The duration is
computed in whole months, then rendered through i18next plural keys
(`duration.years_one/_few/_many/_other`, `duration.months_…`), giving "1 yr 2 mos" and
"1 г. 2 мес." / "5 лет".

## 5. Travel page implementation

### 5.1 Code splitting

`TravelPage` is loaded with `React.lazy`; Leaflet, its CSS and `topojson-client` are only in
that chunk. The world TopoJSON is imported as a **URL** (`world-atlas/countries-50m.json?url`),
so it is a separately cached static asset, fetched by `useWorldGeo()` (a TanStack Query
query with `staleTime: Infinity`) and converted once with `feature(topo, topo.objects.countries)`.

### 5.2 Map setup

```tsx
<MapContainer crs={L.CRS.EPSG3857} zoomControl minZoom={1.5} maxZoom={10} zoomSnap={0.25}
              worldCopyJump={false} maxBounds={[[-85, -200], [85, 200]]}
              scrollWheelZoom={false} attributionControl={false}
              style={{ background: 'var(--bg)' }}>
  <GeoJSON data={countries} style={styleFor} onEachFeature={bindCountryEvents} />
  {places.map(p => <CircleMarker key={p.id} center={[p.lat, p.lng]} radius={6} … />)}
</MapContainer>
```

- No `TileLayer`, so there is no network dependency and the look stays minimal.
- `visitedNumeric = new Set(countries.map(c => c.isoNumeric))`; `styleFor(feature)` returns
  the `--visited` or `--land` fill. CSS variable values are read once via `getComputedStyle`
  and re-read when the theme changes (theme in the effect dependencies → `geoJsonRef.setStyle`).
- Scroll-wheel zoom is enabled on map `focus`/`click` and disabled on `blur`/mouse leave.
- **Pins:** `CircleMarker` (SVG, crisp, themable) with `Tooltip` = place name; click → `select({ type: 'place', id })`.
- **Country click:** only for visited features → `map.flyToBounds(layer.getBounds(), { padding: [40, 40], duration: 0.8 })`, then `select({ type: 'country', iso2 })`.
- **Place selection:** `map.flyTo([lat, lng], Math.max(map.getZoom(), 5), { duration: 0.8 })`.
- **Initial view:** `fitBounds` of all place coordinates (padding 40); with 0 places, the world view.
- **Countries without an ISO numeric id** in world-atlas (e.g. Kosovo, N. Cyprus): the polygon
  is not highlighted, but the pin and list entry still work (acceptable, documented).

### 5.3 Selection state

`TravelPage` owns `selection: { type: 'place', id } | { type: 'country', iso2 } | null`.
Only a place selection is mirrored to `?place=<id>` (`useSearchParams`, `replace: true`).
An unknown `?place=` id is ignored and removed from the URL.

### 5.4 Details panel

- Desktop (≥ 768 px): `motion.aside` fixed to the right of the map area, `x: 100% → 0`, 250 ms.
- Mobile: the same component rendered as a bottom sheet (`y: 100% → 0`), `max-height: 70vh`,
  internal scroll; drag-to-close is not required.
- Focus: on open, focus the panel heading (`tabIndex={-1}`); on close, restore focus to the
  stored trigger element. `Esc` closes it (listener active only while open and the lightbox is closed).
- `PhotoGrid`: `useState(showAll)`; renders `photos.slice(0, showAll ? 30 : 9)` and a
  "Show all {{count}} photos" button when `photos.length > 9`.
- `PhotoLightbox`: `slides = photos.map(p => ({ src: p.url, alt, description: caption, width, height }))`
  over **all** photos; `index` = the clicked index; `carousel.preload = 1` (neighbours only);
  custom `render.slide` uses `RemoteImage` with the thumb shown blurred until the full image loads.

### 5.5 Stats count-up

`useCountUp(target, 800ms)`: `requestAnimationFrame` with ease-out; starts when the stats bar
enters the viewport (`IntersectionObserver`); returns the target immediately under reduced motion.
`worldPercent` is formatted with `Intl.NumberFormat(lang, { maximumFractionDigits: 1 })`.

## 6. Build, serve, deploy

### 6.1 `vite.config.ts`

- `server.proxy: { '/api': { target: 'https://localhost:8080', secure: false } }`.
- `build.target: 'es2022'`, default hashed assets under `/assets/`.
- Vitest config in the same file (`environment: 'jsdom'`, `setupFiles: ['src/test/setup.ts']`).

### 6.2 `frontend/Dockerfile`

```dockerfile
FROM node:<lts>-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build          # tsc -b && vite build

FROM nginx:alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
```

### 6.3 `frontend/nginx.conf`

```nginx
server {                       # port 80: ACME challenge + redirect
  listen 80;
  location /.well-known/acme-challenge/ { root /var/www/certbot; }
  location / { return 301 https://$host$request_uri; }
}

server {
  listen 443 ssl;
  http2 on;
  root /usr/share/nginx/html;
  ssl_certificate     /etc/nginx/certs/fullchain.pem;   # Let's Encrypt or self-signed,
  ssl_certificate_key /etc/nginx/certs/privkey.pem;     # linked by docker/40-tls-certs.sh
  ssl_protocols TLSv1.2 TLSv1.3;

  gzip on;
  gzip_types text/css application/javascript application/json image/svg+xml;

  location /api/ {
    proxy_pass https://api:8080;
    proxy_ssl_verify off;
    proxy_set_header Host $host;
    proxy_set_header X-Request-ID $request_id;
  }
  location /assets/ {
    add_header Cache-Control "public, max-age=31536000, immutable";
    try_files $uri =404;
  }
  location = /index.html { add_header Cache-Control "no-cache"; }
  location / { try_files $uri /index.html; }
}
```

nginx forwards `Accept-Language` by default (all client headers are proxied).

### 6.4 npm scripts

| Script | Command |
|--------|---------|
| `dev` | `vite` |
| `build` | `tsc -b && vite build` |
| `preview` | `vite preview` |
| `lint` | `eslint . && prettier --check .` |
| `typecheck` | `tsc -b --noEmit` |
| `test` | `vitest run` |

## 7. Testing strategy

| Area | Approach |
|------|----------|
| i18n parity | A test flattens the keys of `en.json` and `ru.json` and expects equal sets. |
| `lib/dates` | Table tests: EN/RU periods, "Present"/"н. в.", durations 1/2/5/11/21 months and years (RU plural forms). |
| API client | MSW handler records `request.headers.get('accept-language')`; asserts `en`/`ru` and that the URL has no `lang`. Error JSON → `ApiError`. |
| Routing | `/` → Main; `/main` → `/`; unknown path → NotFound. |
| Providers | Language/theme persistence (mocked `localStorage`), `<html lang>` and `data-theme` updates. |
| RemoteImage | thumb/full source selection, null fallbacks, error placeholder (fire `error` event), alt chain. |
| Pages | Main and CV render from MSW fixtures (fixtures = the contract examples); section order; null fields hidden; ErrorState + Retry. |
| Travel (non-map) | `TravelMap` is mocked in tests; PlaceList → panel opens; `Esc` closes; `?place=` deep link; 30-photo fixture → 9 visible, then "Show all 30"; lightbox arrows/Esc. |
| Manual | Map rendering, fly-to, theme recolouring, mobile bottom sheet, reduced motion, print preview: the checklist from the spec §11. |

## 8. Risks and mitigations

| Risk | Mitigation |
|------|-----------|
| Leaflet without tiles looks odd at high zoom (50 m borders are coarse) | `maxZoom: 10`; minimal style hides it; the 10 m dataset can be swapped in later if needed. |
| TopoJSON size hurts the Travel page load | Lazy chunk + separately cached asset + gzip; ~230 KB gzipped is within NFR-01. |
| Leaflet CSS/markers vs. dark theme | Only vector layers (`CircleMarker`, `GeoJSON`) styled from tokens; no default PNG marker icons. |
| `react-leaflet` and React 19 compatibility | react-leaflet v5 targets React 19; pin versions in `package-lock.json`. |
| Third-party photo hosts are slow | Lazy loading, thumbnails, 9-photo initial grid, neighbour-only preloading. |

## 9. Spec clarifications made in this phase

- **Flags:** spec §5.2 said "flag emoji from ISO2". Windows doesn't render flag emoji (it shows
  letters), so flags are rendered as SVGs (`country-flag-icons`). The spec is updated.

## Changelog

| Date       | Change |
|------------|--------|
| 2026-10-17 | Initial plan. |
| 2026-10-17 | Implemented (F01–F20). Notes: the map reads token colours via a probe element (SVG attributes cannot use CSS variables) and keeps pins in their own pane; flags are globbed SVG URL assets (`?url&no-inline`); the lightbox slide is sized from the slide rect; ThemeProvider applies `data-theme` before re-rendering; verification results are in tasks.md. |
| 2026-10-17 | (F01) Actual versions: Vite 8, TypeScript 6, React Router 8 (declarative API unchanged), MSW 3 (`onUnhandledFrame` replaces `onUnhandledRequest`), Vitest 5. The Vite template now ships oxlint; replaced by ESLint as planned, pinned to ESLint 9 because `eslint-plugin-jsx-a11y` does not support ESLint 10 yet. Test fixtures are captured from the real API (placeholder content, contract-verified in B21). |
