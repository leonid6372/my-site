import { useTranslation } from 'react-i18next';
import { useTravel } from '../../api/queries';
import type { Travel } from '../../api/types';
import { ErrorState } from '../../components/ErrorState';
import { LoadingRegion, Skeleton } from '../../components/Skeleton';
import { usePageTitle } from '../../lib/usePageTitle';
import TravelMap from './map/TravelMap';
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
  return (
    <>
      <StatsBar stats={travel.stats} />
      {travel.places.length === 0 && <p className={styles.empty}>{t('travel.noTrips')}</p>}
      <div className={styles.mapArea} data-selection={selectionKey(selection)}>
        <TravelMap
          countries={travel.countries}
          places={travel.places}
          selection={selection}
          onSelect={onSelect}
        />
      </div>
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
