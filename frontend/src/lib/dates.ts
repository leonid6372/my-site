// Date formatting (spec §4.4, plan §4.5). API dates are "YYYY-MM-DD" calendar dates;
// they are handled in UTC so the user's time zone never shifts a day or month.
import type { Lang, Visit } from '../api/types';
import i18n from '../i18n';

export function parseDate(iso: string): Date {
  const [y = 0, m = 1, d = 1] = iso.split('-').map(Number);
  return new Date(Date.UTC(y, m - 1, d));
}

const periodFormat = (lang: Lang) =>
  new Intl.DateTimeFormat(lang, { month: 'short', year: 'numeric', timeZone: 'UTC' });

const visitFormat = (lang: Lang) =>
  new Intl.DateTimeFormat(lang, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    timeZone: 'UTC',
  });

/** CV period: "May 2021 – Present" / "май 2021 г. – н. в.". */
export function formatPeriod(start: string, end: string | null, lang: Lang): string {
  const fmt = periodFormat(lang);
  const to = end ? fmt.format(parseDate(end)) : i18n.getFixedT(lang)('common.present');
  return `${fmt.format(parseDate(start))} – ${to}`;
}

/**
 * Whole months between two dates, counting both the start and the end month
 * (2021-05 … 2021-05 = 1 month), the way CVs usually count.
 */
export function monthsBetween(start: string, end: string | null, now: Date = new Date()): number {
  const s = parseDate(start);
  const e = end ? parseDate(end) : now;
  const months =
    (e.getUTCFullYear() - s.getUTCFullYear()) * 12 + (e.getUTCMonth() - s.getUTCMonth()) + 1;
  return Math.max(1, months);
}

/** "3 yrs 5 mos" / "3 года 5 мес." with plural rules from i18next. */
export function formatMonths(total: number, lang: Lang): string {
  const t = i18n.getFixedT(lang);
  const years = Math.floor(total / 12);
  const months = total % 12;
  const parts: string[] = [];
  if (years > 0) parts.push(t('duration.years', { count: years }));
  if (months > 0 || years === 0) parts.push(t('duration.months', { count: Math.max(months, 1) }));
  return parts.join(' ');
}

/** Duration of a CV entry; an open end runs until `now`. */
export function formatDuration(
  start: string,
  end: string | null,
  lang: Lang,
  now: Date = new Date(),
): string {
  return formatMonths(monthsBetween(start, end, now), lang);
}

/** Travel visit: "Aug 1 – 4, 2024" / "1–4 авг. 2024 г."; one day → a single date. */
export function formatVisit(visit: Visit, lang: Lang): string {
  const fmt = visitFormat(lang);
  const start = parseDate(visit.startDate);
  if (!visit.endDate || visit.endDate === visit.startDate) return fmt.format(start);
  return fmt.formatRange(start, parseDate(visit.endDate));
}

/** Year of the most recent visit (visits are sorted newest first by the API). */
export function latestVisitYear(visits: Visit[]): number | null {
  const first = visits[0];
  return first ? parseDate(first.startDate).getUTCFullYear() : null;
}
