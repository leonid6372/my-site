import { describe, expect, it } from 'vitest';
import en from './en.json';
import i18n from './index';
import ru from './ru.json';

const PLURAL_SUFFIX = /_(zero|one|two|few|many|other)$/;

/** Flattened keys with plural suffixes removed ("duration.years_few" → "duration.years"). */
function keys(obj: object, prefix = ''): Set<string> {
  const out = new Set<string>();
  for (const [k, v] of Object.entries(obj)) {
    const path = prefix + k;
    if (typeof v === 'object' && v !== null) {
      for (const sub of keys(v as object, path + '.')) out.add(sub);
    } else {
      out.add(path.replace(PLURAL_SUFFIX, ''));
    }
  }
  return out;
}

describe('i18n resources', () => {
  it('en.json and ru.json have the same keys', () => {
    expect([...keys(ru)].sort()).toEqual([...keys(en)].sort());
  });

  it('has every plural form each language needs', () => {
    const plural = (lang: string) => new Intl.PluralRules(lang).resolvedOptions().pluralCategories;
    for (const [lang, res] of [
      ['en', en],
      ['ru', ru],
    ] as const) {
      const flat = JSON.stringify(res);
      for (const base of [
        'duration.years',
        'duration.months',
        'stats.countries',
        'travel.showAll',
      ]) {
        const [section, key] = base.split('.') as [string, string];
        for (const cat of plural(lang)) {
          expect(flat, `${lang} ${base}_${cat}`).toContain(`"${key}_${cat}"`);
        }
        expect(section in res).toBe(true);
      }
    }
  });

  it('has no empty strings', () => {
    const empty = (obj: object): boolean =>
      Object.values(obj).some((v) => (typeof v === 'object' ? empty(v as object) : v === ''));
    expect(empty(en)).toBe(false);
    expect(empty(ru)).toBe(false);
  });
});

describe('plurals', () => {
  const ru = i18n.getFixedT('ru');
  const en = i18n.getFixedT('en');

  it('RU years: 1 год / 2 года / 5 лет', () => {
    expect([1, 2, 5].map((count) => ru('duration.years', { count }))).toEqual([
      '1 год',
      '2 года',
      '5 лет',
    ]);
    expect([11, 21, 22, 25].map((count) => ru('duration.years', { count }))).toEqual([
      '11 лет',
      '21 год',
      '22 года',
      '25 лет',
    ]);
  });

  it('EN years and months', () => {
    expect(en('duration.years', { count: 1 })).toBe('1 yr');
    expect(en('duration.years', { count: 3 })).toBe('3 yrs');
    expect(en('duration.months', { count: 1 })).toBe('1 mo');
    expect(en('duration.months', { count: 5 })).toBe('5 mos');
  });

  it('RU stats labels', () => {
    expect([1, 3, 12].map((count) => ru('stats.countries', { count }))).toEqual([
      'страна',
      'страны',
      'стран',
    ]);
  });

  it('interpolates without HTML escaping', () => {
    expect(en('footer.copyright', { year: 2026, name: 'A & B' })).toBe('© 2026 A & B');
  });
});
