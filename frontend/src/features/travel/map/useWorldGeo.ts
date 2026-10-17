import { useQuery } from '@tanstack/react-query';
import type { FeatureCollection, Geometry } from 'geojson';
import { feature } from 'topojson-client';
import type { Topology } from 'topojson-specification';
import { fixAntimeridian } from './antimeridian';
// Imported as a URL: a separately cached static asset, not part of any JS chunk (plan §5.1).
import worldUrl from 'world-atlas/countries-50m.json?url';

export type WorldGeo = FeatureCollection<Geometry, { name?: string }>;

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
    features: geo.features.map((f) => ({ ...f, geometry: fixAntimeridian(f.geometry) })),
  };
}

/** World countries as GeoJSON; feature ids are ISO 3166-1 numeric codes ("528"). */
export function useWorldGeo() {
  return useQuery({ queryKey: ['world-geo'], queryFn: loadWorld, staleTime: Infinity });
}
