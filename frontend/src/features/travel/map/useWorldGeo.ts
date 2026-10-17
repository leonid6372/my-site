import { useQuery } from '@tanstack/react-query';
import type { FeatureCollection, Geometry } from 'geojson';
import { feature } from 'topojson-client';
import type { Topology } from 'topojson-specification';
// Imported as a URL: a separately cached static asset, not part of any JS chunk (plan §5.1).
import worldUrl from 'world-atlas/countries-50m.json?url';

export type WorldGeo = FeatureCollection<Geometry, { name?: string }>;

async function loadWorld(): Promise<WorldGeo> {
  const res = await fetch(worldUrl);
  if (!res.ok) throw new Error(`world map: HTTP ${String(res.status)}`);
  const topo = (await res.json()) as Topology;
  const countries = topo.objects.countries;
  if (!countries) throw new Error('world map: no countries object');
  return feature(topo, countries) as WorldGeo;
}

/** World countries as GeoJSON; feature ids are ISO 3166-1 numeric codes ("528"). */
export function useWorldGeo() {
  return useQuery({ queryKey: ['world-geo'], queryFn: loadWorld, staleTime: Infinity });
}
