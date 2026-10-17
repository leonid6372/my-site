import { ArrowRight, MapPin } from 'lucide-react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router';
import { useMain } from '../../api/queries';
import type { MainInfo } from '../../api/types';
import { ChipList } from '../../components/Chip';
import { ErrorState } from '../../components/ErrorState';
import { RemoteImage } from '../../components/RemoteImage';
import { LoadingRegion, Skeleton } from '../../components/Skeleton';
import { SocialLinks } from '../../components/SocialLinks';
import { staggerContainer, useStaggerItem } from '../../lib/motion';
import { usePageTitle } from '../../lib/usePageTitle';
import styles from './MainPage.module.css';

export default function MainPage() {
  usePageTitle('main');
  const { data, isPending, isError, refetch, isFetching } = useMain();

  if (isPending) return <MainSkeleton />;
  if (isError) {
    return (
      <ErrorState
        onRetry={() => {
          void refetch();
        }}
        retrying={isFetching}
      />
    );
  }
  return <MainContent main={data} />;
}

function MainContent({ main }: { main: MainInfo }) {
  const { t } = useTranslation();
  const item = useStaggerItem();

  return (
    <motion.div className={styles.page} variants={staggerContainer} initial="hidden" animate="show">
      <div className={styles.body}>
        <motion.header className={styles.hero} variants={item}>
          <h1>{main.fullName}</h1>
          <p className={styles.headline}>{main.headline}</p>
          <p className={styles.location}>
            <MapPin size={16} aria-hidden />
            <span className="visually-hidden">{t('main.location')}: </span>
            {main.location}
          </p>
        </motion.header>

        <motion.p className={styles.currently} variants={item}>
          <span className={styles.currentlyLabel}>{t('main.currently')}</span>
          <span>{main.currently}</span>
        </motion.p>

        <motion.section aria-labelledby="about-heading" variants={item}>
          <h2 id="about-heading" className={styles.label}>
            {t('main.about')}
          </h2>
          <div className={styles.bio}>
            {main.bio.map((p) => (
              <p key={p}>{p}</p>
            ))}
          </div>
        </motion.section>

        {main.interests.length > 0 && (
          <motion.section aria-labelledby="interests-heading" variants={item}>
            <h2 id="interests-heading" className={styles.label}>
              {t('main.interests')}
            </h2>
            <ChipList items={main.interests} />
          </motion.section>
        )}

        <motion.div className={styles.ctas} variants={item}>
          <CtaCard to="/travel" title={t('main.ctaTravel')} text={t('main.ctaTravelText')} />
          <CtaCard to="/cv" title={t('main.ctaCv')} text={t('main.ctaCvText')} />
        </motion.div>
      </div>

      <motion.aside className={styles.side} variants={item}>
        <RemoteImage
          photo={main.photo}
          variant="portrait"
          aspect="square"
          className={styles.portrait}
        />
        <section aria-labelledby="contacts-heading" className={styles.contacts}>
          <h2 id="contacts-heading" className={styles.label}>
            {t('main.contacts')}
          </h2>
          <SocialLinks links={main.links} />
        </section>
      </motion.aside>
    </motion.div>
  );
}

function CtaCard({ to, title, text }: { to: string; title: string; text: string }) {
  return (
    <Link to={to} className={styles.card}>
      <span className={styles.cardTitle}>
        {title}
        <ArrowRight size={18} aria-hidden className={styles.cardArrow} />
      </span>
      <span className={styles.cardText}>{text}</span>
    </Link>
  );
}

function MainSkeleton() {
  return (
    <LoadingRegion>
      <div className={styles.page}>
        <div className={styles.side}>
          <Skeleton className={styles.portrait} height="auto" radius="50%" />
        </div>
        <div className={styles.body}>
          <Skeleton width="60%" height="2.5rem" />
          <Skeleton width="40%" height="1.25rem" />
          <Skeleton width="100%" height="2.5rem" />
          <Skeleton width="100%" height="6rem" />
          <Skeleton width="70%" height="2rem" />
        </div>
      </div>
    </LoadingRegion>
  );
}
