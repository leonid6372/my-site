import { useTranslation } from 'react-i18next';
import { usePageTitle } from '../../lib/usePageTitle';

// Placeholder until F10.
export default function MainPage() {
  const { t } = useTranslation();
  usePageTitle('main');
  return <h1>{t('titles.main')}</h1>;
}
