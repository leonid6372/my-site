import { AnimatePresence, useReducedMotion } from 'motion/react';
import { useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useTravel } from '../../api/queries';
import type { Travel } from '../../api/types';
import { ErrorState } from '../../components/ErrorState';
import { LoadingRegion, Skeleton } from '../../components/Skeleton';
import { usePageTitle } from '../../lib/usePageTitle';
import { DetailsPanel } from './DetailsPanel';
import TravelMap from './map/TravelMap';
import { PhotoGrid } from './PhotoGrid';
import { PlaceList } from './PlaceList';
import { StatsBar } from './StatsBar';
import styles from './TravelPage.module.css';
import { useSelection, type Selection } from './useSelection';

// Loaded lazily (routes.tsx): Leaflet and the map data live in this chunk.
export default function TravelPage() {
  usePageTitle('travel');
  const { t } = useTranslation();
  const travel = useTravel();
  const { selection, select } = useSelection(travel.data?.places);

  return (
    <div className={styles.page}>
      <h1>{t('titles.travel')}</h1>
      {travel.isError ? (
        <ErrorState
          onRetry={() => {
            void travel.refetch();
          }}
          retrying={travel.isFetching}
        />
      ) : travel.isPending ? (
        <TravelSkeleton />
      ) : (
        <TravelContent travel={travel.data} selection={selection} onSelect={select} />
      )}
    </div>
  );
}

function selectionKey(s: Selection): string {
  if (!s) return '';
  return s.type === 'place' ? `place:${s.id}` : `country:${s.iso2}`;
}

function TravelContent({
  travel,
  selection,
  onSelect,
}: {
  travel: Travel;
  selection: Selection;
  onSelect: (s: Selection) => void;
}) {
  const { t } = useTranslation();
  const reduced = useReducedMotion() ?? false;
  // While the lightbox is open, its Esc closes the lightbox only, not the panel.
  const [lightboxOpen, setLightboxOpen] = useState(false);
  const mapArea = useRef<HTMLDivElement>(null);

  // Choosing a place in the list selects it and brings the map into view (US-24).
  const selectFromList = (s: Selection) => {
    onSelect(s);
    mapArea.current?.scrollIntoView({ behavior: reduced ? 'auto' : 'smooth', block: 'center' });
  };

  return (
    <>
      <StatsBar stats={travel.stats} />
      {travel.places.length === 0 && <p className={styles.empty}>{t('travel.noTrips')}</p>}
      <div ref={mapArea} className={styles.mapArea} data-selection={selectionKey(selection)}>
        <TravelMap
          countries={travel.countries}
          places={travel.places}
          selection={selection}
          onSelect={onSelect}
        />
        <AnimatePresence>
          {selection && (
            <DetailsPanel
              key="details"
              selection={selection}
              countries={travel.countries}
              places={travel.places}
              onSelect={onSelect}
              escapeDisabled={lightboxOpen}
              renderPhotos={(place) => (
                <PhotoGrid
                  key={place.id}
                  photos={place.photos}
                  onLightboxChange={setLightboxOpen}
                />
              )}
            />
          )}
        </AnimatePresence>
      </div>
      <PlaceList
        countries={travel.countries}
        places={travel.places}
        selection={selection}
        onSelect={selectFromList}
      />
    </>
  );
}

function TravelSkeleton() {
  return (
    <LoadingRegion>
      <div className={styles.skeleton}>
        <Skeleton height="4.5rem" />
        <Skeleton height="min(60vh, 520px)" />
      </div>
    </LoadingRegion>
  );
}
