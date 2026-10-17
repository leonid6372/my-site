import { useTranslation } from 'react-i18next';
import styles from './SkipLink.module.css';

/** First focusable element: jumps to <main id="content"> (spec §8). */
export function SkipLink() {
  const { t } = useTranslation();
  return (
    <a
      href="#content"
      className={styles.skip}
      data-print="hide"
      onClick={(e) => {
        // Move focus explicitly: some browsers only scroll on a fragment link.
        e.preventDefault();
        const main = document.getElementById('content');
        main?.focus();
        main?.scrollIntoView();
      }}
    >
      {t('header.skipToContent')}
    </a>
  );
}
