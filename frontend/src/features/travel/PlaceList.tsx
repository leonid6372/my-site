import { useTranslation } from 'react-i18next';
import type { Country, Place } from '../../api/types';
import { latestVisitYear } from '../../lib/dates';
import { flagUrl } from './flags';
import styles from './PlaceList.module.css';
import type { Selection } from './useSelection';

/**
 * Countries and places as a list (spec §5.2): the accessible equivalent of the map.
 * Countries keep the API order (by localised name); places keep the API order
 * within each country.
 */
export function PlaceList({
  countries,
  places,
  selection,
  onSelect,
}: {
  countries: Country[];
  places: Place[];
  selection: Selection;
  onSelect: (s: Selection) => void;
}) {
  const { t } = useTranslation();
  if (countries.length === 0) return null;
  const byId = new Map(places.map((p) => [p.id, p]));

  return (
    <section aria-labelledby="place-list-heading" className={styles.section}>
      <h2 id="place-list-heading" className={styles.heading}>
        {t('travel.listHeading')}
      </h2>
      <ul className={styles.groups}>
        {countries.map((c) => {
          const inCountry = c.placeIds.flatMap((id) => byId.get(id) ?? []);
          const flag = flagUrl(c.iso2);
          return (
            <li key={c.iso2} className={styles.group}>
              <h3 className={styles.country}>
                {flag && <img src={flag} alt="" className={styles.flag} width={24} height={16} />}
                <span>{c.name}</span>
                <span className={styles.count}>
                  {t('travel.placeCount', { count: inCountry.length })}
                </span>
              </h3>
              <ul className={styles.places}>
                {inCountry.map((p) => {
                  const year = latestVisitYear(p.visits);
                  const current = selection?.type === 'place' && selection.id === p.id;
                  return (
                    <li key={p.id}>
                      <button
                        type="button"
                        className={styles.place}
                        aria-current={current ? 'true' : undefined}
                        onClick={() => {
                          onSelect({ type: 'place', id: p.id });
                        }}
                      >
                        <span>{p.name}</span>
                        {year !== null && (
                          <span className={styles.year}>
                            <span className="visually-hidden">
                              {`, ${t('travel.lastVisit', { year })}`}
                            </span>
                            <span aria-hidden="true">{year}</span>
                          </span>
                        )}
                      </button>
                    </li>
                  );
                })}
              </ul>
            </li>
          );
        })}
      </ul>
    </section>
  );
}
