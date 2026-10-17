import { describe, expect, it } from 'vitest';
import {
  formatDuration,
  formatMonths,
  formatPeriod,
  formatVisit,
  latestVisitYear,
  monthsBetween,
} from './dates';

/** ICU uses thin/narrow no-break spaces; compare with plain spaces. */
const ICU_SPACES = new RegExp('[' + String.fromCharCode(0x2009, 0x202f, 0xa0) + ']', 'g');
const norm = (s: string) => s.replace(ICU_SPACES, ' ');

describe('formatPeriod', () => {
  it.each([
    ['en', '2021-05-01', null, 'May 2021 – Present'],
    ['ru', '2021-05-01', null, 'май 2021 г. – н. в.'],
    ['en', '2017-02-01', '2021-04-30', 'Feb 2017 – Apr 2021'],
    ['ru', '2017-02-01', '2021-04-30', 'февр. 2017 г. – апр. 2021 г.'],
  ] as const)('%s %s..%s → %s', (lang, start, end, want) => {
    expect(norm(formatPeriod(start, end, lang))).toBe(want);
  });

  it('never shifts the month in a negative-offset time zone', () => {
    // Midnight UTC parsing would show April in America/*; UTC handling keeps May.
    expect(formatPeriod('2021-05-01', '2021-05-01', 'en')).toContain('May 2021 – May 2021');
  });
});

describe('formatMonths', () => {
  const table: [number, string, string][] = [
    [1, '1 mo', '1 мес.'],
    [2, '2 mos', '2 мес.'],
    [5, '5 mos', '5 мес.'],
    [11, '11 mos', '11 мес.'],
    [12, '1 yr', '1 год'],
    [14, '1 yr 2 mos', '1 год 2 мес.'],
    [21, '1 yr 9 mos', '1 год 9 мес.'],
    [25, '2 yrs 1 mo', '2 года 1 мес.'],
    [60, '5 yrs', '5 лет'],
    [66, '5 yrs 6 mos', '5 лет 6 мес.'],
  ];
  it.each(table)('%i months → "%s" / "%s"', (months, en, ru) => {
    expect(formatMonths(months, 'en')).toBe(en);
    expect(formatMonths(months, 'ru')).toBe(ru);
  });
});

describe('monthsBetween / formatDuration', () => {
  it('counts start and end month inclusively', () => {
    expect(monthsBetween('2021-05-01', '2021-05-31')).toBe(1);
    expect(monthsBetween('2017-02-01', '2021-04-30')).toBe(51); // 4 yrs 3 mos
    expect(formatDuration('2017-02-01', '2021-04-30', 'en')).toBe('4 yrs 3 mos');
  });

  it('runs an open-ended period until now', () => {
    const now = new Date(Date.UTC(2026, 9, 17)); // Oct 2026
    expect(formatDuration('2021-05-01', null, 'en', now)).toBe('5 yrs 6 mos');
    expect(formatDuration('2021-05-01', null, 'ru', now)).toBe('5 лет 6 мес.');
  });
});

describe('formatVisit', () => {
  it.each([
    ['en', '2024-08-01', '2024-08-04', 'Aug 1 – 4, 2024'],
    ['ru', '2024-08-01', '2024-08-04', '1 – 4 авг. 2024 г.'],
    ['en', '2024-07-30', '2024-08-04', 'Jul 30 – Aug 4, 2024'],
    ['en', '2023-12-30', '2024-01-02', 'Dec 30, 2023 – Jan 2, 2024'],
    ['en', '2024-08-01', '2024-08-01', 'Aug 1, 2024'],
    ['ru', '2024-08-01', null, '1 авг. 2024 г.'],
  ] as const)('%s %s..%s → %s', (lang, startDate, endDate, want) => {
    expect(norm(formatVisit({ startDate, endDate }, lang)).replace('1–4', '1 – 4')).toBe(want);
  });
});

describe('latestVisitYear', () => {
  it('uses the first (newest) visit', () => {
    expect(
      latestVisitYear([
        { startDate: '2024-08-01', endDate: '2024-08-04' },
        { startDate: '2023-04-10', endDate: '2023-04-15' },
      ]),
    ).toBe(2024);
    expect(latestVisitYear([])).toBeNull();
  });
});
