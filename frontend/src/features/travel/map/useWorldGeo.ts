import { useQuery } from '@tanstack/react-query';
import type { FeatureCollection, Geometry } from 'geojson';
import { feature } from 'topojson-client';
import type { Topology } from 'topojson-specification';
import { ABKHAZIA_RING } from './abkhazia';
import { fixAntimeridian } from './antimeridian';
// Imported as a URL: a separately cached static asset, not part of any JS chunk (plan §5.1).
import worldUrl from 'world-atlas/countries-50m.json?url';

/** `mask`: an opaque, non-interactive land underlay (see `extraFeatures`). */
export type WorldGeo = FeatureCollection<Geometry, { name?: string; mask?: boolean }>;
type WorldFeature = WorldGeo['features'][number];

/**
 * Places world-atlas doesn't have as their own feature, drawn on top of the country
 * they sit in. Their id is the ISO2 code. Each comes with a mask first: it hides the
 * parent's fill under them, so the parent and the place are marked visited separately.
 */
function extraFeatures(): WorldFeature[] {
  const geometry: Geometry = { type: 'Polygon', coordinates: [ABKHAZIA_RING] };
  return [
    { type: 'Feature', id: 'AB-mask', properties: { mask: true }, geometry },
    { type: 'Feature', id: 'AB', properties: { name: 'Abkhazia' }, geometry },
  ];
}

async function loadWorld(): Promise<WorldGeo> {
  const res = await fetch(worldUrl);
  if (!res.ok) throw new Error(`world map: HTTP ${String(res.status)}`);
  const topo = (await res.json()) as Topology;
  const countries = topo.objects.countries;
  if (!countries) throw new Error('world map: no countries object');
  const geo = feature(topo, countries) as WorldGeo;
  // Without this, rings crossing ±180° draw as bands across the whole map.
  return {
    ...geo,
    features: [
      ...geo.features.map((f) => ({ ...f, geometry: fixAntimeridian(f.geometry) })),
      ...extraFeatures(),
    ],
  };
}

/**
 * World countries as GeoJSON; feature ids are ISO 3166-1 numeric codes ("528"), or
 * ISO2 for places without one ("AB"). Match them with `featureKey`.
 */
export function useWorldGeo() {
  return useQuery({ queryKey: ['world-geo'], queryFn: loadWorld, staleTime: Infinity });
}

/** The world feature id of a country: its ISO numeric code, or ISO2 when it has none. */
export function featureKey(c: { iso2: string; isoNumeric: string }): string {
  return c.isoNumeric || c.iso2;
}
