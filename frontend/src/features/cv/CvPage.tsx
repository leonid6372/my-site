import { useTranslation } from 'react-i18next';
import { usePageTitle } from '../../lib/usePageTitle';

// Placeholder until F11.
export default function CvPage() {
  const { t } = useTranslation();
  usePageTitle('cv');
  return <h1>{t('titles.cv')}</h1>;
}
