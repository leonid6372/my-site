// SVG flags as separate static files (spec §5.2 clarification: Windows doesn't render
// flag emoji). `no-inline` keeps Vite from base64-inlining all ~265 small SVGs into
// the chunk: the chunk only holds this URL map, and the browser fetches the few
// flags actually shown.
import abkhaziaUrl from './flags/AB.svg?url&no-inline';

const files = import.meta.glob<string>('/node_modules/country-flag-icons/3x2/*.svg', {
  query: '?url&no-inline',
  import: 'default',
  eager: true,
});

const byIso2 = new Map(
  Object.entries(files).map(([path, url]) => [path.slice(-6, -4).toUpperCase(), url]),
);

// Flags missing from country-flag-icons. "AB" (Abkhazia) is not an ISO code; its
// flag comes from Wikimedia Commons.
byIso2.set('AB', abkhaziaUrl);

/** URL of the 3:2 SVG flag for an ISO 3166-1 alpha-2 code, if there is one. */
export function flagUrl(iso2: string): string | undefined {
  return byIso2.get(iso2.toUpperCase());
}
