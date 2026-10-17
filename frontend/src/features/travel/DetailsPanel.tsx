import { MapPin, X } from 'lucide-react';
import { motion, useReducedMotion } from 'motion/react';
import { useEffect, useRef, useState, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import type { Country, Place } from '../../api/types';
import { useLanguage } from '../../app/providers/language-context';
import { formatVisit, latestVisitYear } from '../../lib/dates';
import { panelTransition } from '../../lib/motion';
import { useMediaQuery } from '../../lib/useMediaQuery';
import styles from './DetailsPanel.module.css';
import type { Selection } from './useSelection';

interface Props {
  selection: NonNullable<Selection>;
  countries: Country[];
  places: Place[];
  onSelect: (s: Selection) => void;
  /** Content below the place details (the photo grid, F15). */
  renderPhotos?: (place: Place) => ReactNode;
  /** While true (e.g. the lightbox is open) Esc does not close the panel. */
  escapeDisabled?: boolean;
}

/**
 * Place or country details (spec §5.2): a side panel on desktop, a bottom sheet on
 * mobile. Focus moves to the heading on open and back to the trigger on close; Esc
 * closes it.
 */
export function DetailsPanel({
  selection,
  countries,
  places,
  onSelect,
  renderPhotos,
  escapeDisabled = false,
}: Props) {
  const { t } = useTranslation();
  const desktop = useMediaQuery('(min-width: 768px)');
  const reduced = useReducedMotion() ?? false;
  const heading = useRef<HTMLHeadingElement>(null);
  // The element that had focus when the panel opened (a pin, a list button…),
  // captured once at mount, before focus moves into the panel.
  const [trigger] = useState(() => document.activeElement);

  const key = selection.type === 'place' ? `p:${selection.id}` : `c:${selection.iso2}`;
  useEffect(() => {
    heading.current?.focus();
  }, [key]);

  const close = () => {
    onSelect(null);
    if (trigger instanceof HTMLElement && trigger.isConnected) trigger.focus();
  };

  useEffect(() => {
    if (escapeDisabled) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') close();
    };
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('keydown', onKey);
    };
  });

  const offscreen = desktop ? { x: '100%' } : { y: '100%' };
  const motionProps = reduced
    ? {
        initial: { opacity: 0 },
        animate: { opacity: 1 },
        exit: { opacity: 0 },
        transition: { duration: 0.1 },
      }
    : { initial: offscreen, animate: { x: 0, y: 0 }, exit: offscreen, transition: panelTransition };

  let body: ReactNode = null;
  let title = '';
  if (selection.type === 'place') {
    const place = places.find((p) => p.id === selection.id);
    if (place) {
      title = place.name;
      body = (
        <PlaceDetails place={place} country={countries.find((c) => c.iso2 === place.countryIso2)}>
          {renderPhotos?.(place)}
        </PlaceDetails>
      );
    }
  } else {
    const country = countries.find((c) => c.iso2 === selection.iso2);
    if (country) {
      title = country.name;
      body = <CountryDetails country={country} places={places} onSelect={onSelect} />;
    }
  }

  return (
    <motion.aside
      className={desktop ? styles.side : styles.sheet}
      aria-labelledby="details-heading"
      {...motionProps}
    >
      <div className={styles.head}>
        <h2 id="details-heading" ref={heading} tabIndex={-1} className={styles.title}>
          {title}
        </h2>
        <button
          type="button"
          className={styles.close}
          onClick={close}
          aria-label={t('travel.closeDetails')}
        >
          <X size={18} aria-hidden />
        </button>
      </div>
      <div className={styles.body}>{body}</div>
    </motion.aside>
  );
}

function PlaceDetails({
  place,
  country,
  children,
}: {
  place: Place;
  country: Country | undefined;
  children?: ReactNode;
}) {
  const { t } = useTranslation();
  const { lang } = useLanguage();
  return (
    <>
      {country && (
        <p className={styles.meta}>
          <MapPin size={14} aria-hidden />
          {country.name}
        </p>
      )}
      <section aria-labelledby="visits-heading">
        <h3 id="visits-heading" className={styles.label}>
          {t('travel.visits')}
          {place.visits.length > 1 && (
            <span className={styles.badge}>
              {t('travel.visitCount', { count: place.visits.length })}
            </span>
          )}
        </h3>
        <ul className={styles.visits}>
          {place.visits.map((v) => (
            <li key={v.startDate}>{formatVisit(v, lang)}</li>
          ))}
        </ul>
      </section>
      <div className={styles.story}>
        {place.story.split(/\n{2,}/).map((p) => (
          <p key={p}>{p}</p>
        ))}
      </div>
      {children}
    </>
  );
}

function CountryDetails({
  country,
  places,
  onSelect,
}: {
  country: Country;
  places: Place[];
  onSelect: (s: Selection) => void;
}) {
  const { t } = useTranslation();
  const inCountry = places.filter((p) => country.placeIds.includes(p.id));
  return (
    <>
      <p className={styles.meta}>
        {t('travel.continent')}: {t(`continents.${country.continent}`)}
      </p>
      <section aria-labelledby="country-places-heading">
        <h3 id="country-places-heading" className={styles.label}>
          {t('travel.places')}
        </h3>
        <ul className={styles.placeList}>
          {inCountry.map((p) => (
            <li key={p.id}>
              <button
                type="button"
                className={styles.placeButton}
                onClick={() => {
                  onSelect({ type: 'place', id: p.id });
                }}
              >
                <span>{p.name}</span>
                <span className={styles.year}>{latestVisitYear(p.visits)}</span>
              </button>
            </li>
          ))}
        </ul>
      </section>
    </>
  );
}
