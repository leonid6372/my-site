import { useTranslation } from 'react-i18next';
import { usePageTitle } from '../../lib/usePageTitle';

// Placeholder until F12. Loaded lazily: Leaflet and the map data live in this chunk.
export default function TravelPage() {
  const { t } = useTranslation();
  usePageTitle('travel');
  return <h1>{t('titles.travel')}</h1>;
}
