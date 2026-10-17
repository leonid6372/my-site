import { useTranslation } from 'react-i18next';
import { useMain } from '../api/queries';
import styles from './Footer.module.css';
import { SocialLinks } from './SocialLinks';

/** Footer (spec §4.2): © year name + social icons. */
export function Footer() {
  const { t } = useTranslation();
  const main = useMain().data;
  return (
    <footer className={styles.footer} data-print="hide">
      <div className={styles.inner}>
        <p className={styles.copy}>
          {t('footer.copyright', { year: new Date().getFullYear(), name: main?.fullName ?? '' })}
        </p>
        {main && <SocialLinks links={main.links} variant="icons" label={t('footer.social')} />}
      </div>
    </footer>
  );
}
