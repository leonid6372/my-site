// Stand-in for the Leaflet map in jsdom tests (map rendering is checked manually,
// plan §7). It renders the same labelled region plus one button per pin, so tests
// can select places "on the map".
import { useTranslation } from 'react-i18next';
import type { Country, Place } from '../api/types';
import type { Selection } from '../features/travel/useSelection';

export default function MockTravelMap({
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
          data-testid={`pin-${p.id}`}
          onClick={() => {
            onSelect({ type: 'place', id: p.id });
          }}
        >
          pin {p.name}
        </button>
      ))}
    </div>
  );
}
