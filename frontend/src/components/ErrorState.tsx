import { AlertCircle, RotateCw } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import styles from './ErrorState.module.css';

/** "Something went wrong" + Retry (spec §6, FR-F11). */
export function ErrorState({
  onRetry,
  retrying = false,
}: {
  onRetry: () => void;
  retrying?: boolean;
}) {
  const { t } = useTranslation();
  return (
    <div className={styles.box} role="alert">
      <AlertCircle size={28} aria-hidden className={styles.icon} />
      <h2 className={styles.title}>{t('common.errorTitle')}</h2>
      <p className={styles.text}>{t('common.errorText')}</p>
      <button type="button" className={styles.retry} onClick={onRetry} disabled={retrying}>
        <RotateCw size={16} aria-hidden />
        {t('common.retry')}
      </button>
    </div>
  );
}
