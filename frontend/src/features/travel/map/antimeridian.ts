import type { Geometry, Position } from 'geojson';

// Some countries (Russia's Chukotka, Fiji) have rings that cross the antimeridian:
// consecutive points jump from +180° to -180°. Leaflet draws that jump as a straight
// edge across the whole world, which shows up as horizontal bands and lines on a
// zoomed-out map. Unwrapping such rings (west longitudes + 360°) keeps each ring
// continuous; it then simply extends past 180°, which the map's bounds allow.

const JUMP = 180;

function maxJump(ring: Position[]): number {
  let max = 0;
  for (let i = 1; i < ring.length; i++) {
    const a = ring[i - 1];
    const b = ring[i];
    if (a && b) max = Math.max(max, Math.abs((b[0] ?? 0) - (a[0] ?? 0)));
  }
  return max;
}

/**
 * Unwraps a ring that jumps across ±180°. A ring that really circles the globe
 * (Antarctica's coast spans all 360°) cannot be unwrapped: shifting it only moves
 * the jump elsewhere, so it is kept as it is.
 */
export function unwrapRing(ring: Position[]): Position[] {
  if (maxJump(ring) <= JUMP) return ring;
  const unwrapped = ring.map(([lng = 0, lat = 0, ...rest]) => [
    lng < 0 ? lng + 360 : lng,
    lat,
    ...rest,
  ]);
  return maxJump(unwrapped) <= JUMP ? unwrapped : ring;
}

/** Returns the geometry with antimeridian-crossing rings unwrapped. */
export function fixAntimeridian<G extends Geometry>(g: G): G {
  if (g.type === 'Polygon') {
    return { ...g, coordinates: g.coordinates.map(unwrapRing) };
  }
  if (g.type === 'MultiPolygon') {
    return { ...g, coordinates: g.coordinates.map((poly) => poly.map(unwrapRing)) };
  }
  return g;
}
