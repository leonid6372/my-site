import { describe, expect, it } from 'vitest';
import { ABKHAZIA_INNER_BORDER, ABKHAZIA_OUTER_BORDER, ABKHAZIA_RING } from './abkhazia';

describe('Abkhazia outline', () => {
  it('is a closed ring made of the two borders', () => {
    expect(ABKHAZIA_RING[0]).toEqual(ABKHAZIA_RING.at(-1));
    expect(ABKHAZIA_INNER_BORDER.at(-1)).toEqual(ABKHAZIA_OUTER_BORDER[0]);
    expect(ABKHAZIA_RING).toHaveLength(
      ABKHAZIA_INNER_BORDER.length + ABKHAZIA_OUTER_BORDER.length - 1,
    );
  });

  it('lies in the western Caucasus', () => {
    for (const [lng = 0, lat = 0] of ABKHAZIA_RING) {
      expect(lng).toBeGreaterThan(39.9);
      expect(lng).toBeLessThan(42.2);
      expect(lat).toBeGreaterThan(42.4);
      expect(lat).toBeLessThan(43.6);
    }
  });
});
