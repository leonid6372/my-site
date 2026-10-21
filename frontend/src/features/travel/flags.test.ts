import { describe, expect, it } from 'vitest';
import { flagUrl } from './flags';

describe('flagUrl', () => {
  it('finds flags from country-flag-icons', () => {
    expect(flagUrl('nl')).toMatch(/NL\.svg/);
  });

  it('has the Abkhazia flag for "AB"', () => {
    expect(flagUrl('AB')).toMatch(/AB\.svg/);
  });

  it('returns undefined for an unknown code', () => {
    expect(flagUrl('ZZ')).toBeUndefined();
  });
});
