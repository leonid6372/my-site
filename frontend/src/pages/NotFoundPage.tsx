import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';
import { usePageTitle } from '../lib/usePageTitle';
import styles from './NotFoundPage.module.css';

export default function NotFoundPage() {
  const { t } = useTranslation();
  usePageTitle('notFound');
  return (
    <div className={styles.page}>
      <p className={styles.code} aria-hidden="true">
        404
      </p>
      <h1>{t('notFound.title')}</h1>
      <p className={styles.text}>{t('notFound.text')}</p>
      <Link to="/main" className={styles.back}>
        {t('notFound.back')}
      </Link>
    </div>
  );
}
