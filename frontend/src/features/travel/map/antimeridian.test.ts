import type { MultiPolygon, Polygon, Position } from 'geojson';
import { feature } from 'topojson-client';
import type { Topology } from 'topojson-specification';
import world from 'world-atlas/countries-50m.json';
import { describe, expect, it } from 'vitest';
import { fixAntimeridian, unwrapRing } from './antimeridian';

const maxJump = (ring: Position[]) =>
  ring.reduce(
    (m, p, i) => (i === 0 ? m : Math.max(m, Math.abs((p[0] ?? 0) - (ring[i - 1]?.[0] ?? 0)))),
    0,
  );

describe('unwrapRing', () => {
  it('leaves an ordinary ring untouched', () => {
    const ring = [
      [0, 0],
      [10, 0],
      [10, 10],
      [0, 0],
    ];
    expect(unwrapRing(ring)).toBe(ring);
  });

  it('unwraps a ring that jumps across ±180°', () => {
    const ring = [
      [170, 60],
      [179.9, 60],
      [-179.9, 65],
      [-170, 66],
      [170, 60],
    ];
    const out = unwrapRing(ring);
    expect(out.map((p) => p[0])).toEqual([170, 179.9, 180.1, 190, 170]);
    expect(maxJump(out)).toBeLessThan(180);
  });

  it('keeps a ring that circles the globe (like Antarctica’s coast) as it is', () => {
    // Shape of the real data: all longitudes, closing with a ~360° jump.
    const ring = [
      [-180, -80],
      [-90, -70],
      [0, -70],
      [90, -70],
      [179.6, -80],
      [-180, -80],
    ];
    expect(unwrapRing(ring)).toBe(ring);
  });
});

describe('world-atlas countries', () => {
  const topo = world as unknown as Topology;
  const countries = topo.objects.countries;
  if (!countries) throw new Error('no countries object');
  const geo = feature(topo, countries) as unknown as {
    features: { id?: string; geometry: Polygon | MultiPolygon }[];
  };

  /** Ids of features that have a ring jumping across ±180°. */
  const crossing = (fix: boolean) =>
    geo.features
      .filter((f) => {
        const g = fix ? fixAntimeridian(f.geometry) : f.geometry;
        const polys = g.type === 'Polygon' ? [g.coordinates] : g.coordinates;
        return polys.some((poly) => poly.some((r) => maxJump(r) > 180));
      })
      .map((f) => String(f.id));

  it('had crossing rings in the raw data (Russia 643, Fiji 242)', () => {
    expect(crossing(false)).toEqual(expect.arrayContaining(['643', '242']));
  });

  it('after fixing, only Antarctica (whose coast circles the globe) keeps one', () => {
    expect(crossing(true)).toEqual(['010']);
  });
});
