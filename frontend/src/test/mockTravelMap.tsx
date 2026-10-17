// Stand-in for the Leaflet map in jsdom tests (map rendering is checked manually,
// plan §7). It renders the same labelled region plus one button per pin and per
// visited country, so tests can select things "on the map".
import { useTranslation } from 'react-i18next';
import type { Country, Place } from '../api/types';
import type { Selection } from '../features/travel/useSelection';

export default function MockTravelMap({
  countries,
  places,
  onSelect,
}: {
  countries: Country[];
  places: Place[];
  selection: Selection;
  onSelect: (s: Selection) => void;
}) {
  const { t } = useTranslation();
  return (
    <div role="region" aria-label={t('travel.mapLabel')} data-testid="map">
      {places.map((p) => (
        <button
          key={p.id}
          type="button"
          onClick={() => {
            onSelect({ type: 'place', id: p.id });
          }}
        >
          pin {p.id}
        </button>
      ))}
      {countries.map((c) => (
        <button
          key={c.iso2}
          type="button"
          onClick={() => {
            onSelect({ type: 'country', iso2: c.iso2 });
          }}
        >
          country {c.iso2}
        </button>
      ))}
    </div>
  );
}
