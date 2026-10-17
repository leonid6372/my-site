import { useRef } from 'react';
import { useTranslation } from 'react-i18next';
import type { Travel } from '../../api/types';
import { useLanguage } from '../../app/providers/language-context';
import styles from './StatsBar.module.css';
import { useCountUp } from './useCountUp';

function Stat({
  value,
  label,
  format,
}: {
  value: number;
  label: string;
  format: (n: number) => string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const shown = useCountUp(value, ref);
  return (
    <div ref={ref} className={styles.stat}>
      {/* Screen readers get the final value, not the animation frames. */}
      <dt className={styles.label}>{label}</dt>
      <dd className={styles.value}>
        <span aria-hidden="true">{format(shown)}</span>
        <span className="visually-hidden">{format(value)}</span>
      </dd>
    </div>
  );
}

/** Four counters (spec §5.2): countries, continents, places, share of the world. */
export function StatsBar({ stats }: { stats: Travel['stats'] }) {
  const { t } = useTranslation();
  const { lang } = useLanguage();
  const int = new Intl.NumberFormat(lang, { maximumFractionDigits: 0 });
  const pct = new Intl.NumberFormat(lang, { style: 'percent', maximumFractionDigits: 1 });

  return (
    <dl className={styles.bar} aria-label={t('stats.label')}>
      <Stat
        value={stats.countries}
        label={t('stats.countries', { count: stats.countries })}
        format={(n) => int.format(n)}
      />
      <Stat
        value={stats.continents}
        label={t('stats.continents', { count: stats.continents })}
        format={(n) => int.format(n)}
      />
      <Stat
        value={stats.places}
        label={t('stats.places', { count: stats.places })}
        format={(n) => int.format(n)}
      />
      <Stat
        value={stats.worldPercent}
        label={t('stats.world')}
        format={(n) => pct.format(n / 100)}
      />
    </dl>
  );
}
