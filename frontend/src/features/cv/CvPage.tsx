import { MapPin } from 'lucide-react';
import { motion } from 'motion/react';
import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { useCv, useMain } from '../../api/queries';
import type { Cv, Education, Experience, Lang, MainInfo } from '../../api/types';
import { useLanguage } from '../../app/providers/language-context';
import { ChipList } from '../../components/Chip';
import { ErrorState } from '../../components/ErrorState';
import { LoadingRegion, Skeleton } from '../../components/Skeleton';
import { SocialLinks } from '../../components/SocialLinks';
import { formatDuration, formatPeriod } from '../../lib/dates';
import { staggerContainer, useStaggerItem } from '../../lib/motion';
import { usePageTitle } from '../../lib/usePageTitle';
import styles from './CvPage.module.css';

export default function CvPage() {
  usePageTitle('cv');
  const main = useMain();
  const cv = useCv();

  if (main.isError || cv.isError) {
    return (
      <ErrorState
        onRetry={() => {
          if (main.isError) void main.refetch();
          if (cv.isError) void cv.refetch();
        }}
        retrying={main.isFetching || cv.isFetching}
      />
    );
  }
  if (main.isPending || cv.isPending) return <CvSkeleton />;
  return <CvContent main={main.data} cv={cv.data} />;
}

function Section({ id, title, children }: { id: string; title: string; children: ReactNode }) {
  const item = useStaggerItem();
  return (
    <motion.section className={styles.section} aria-labelledby={id} variants={item}>
      <h2 id={id} className={styles.sectionTitle}>
        {title}
      </h2>
      {children}
    </motion.section>
  );
}

/** Spec §5.3: header, summary, experience, education, skills, languages. */
function CvContent({ main, cv }: { main: MainInfo; cv: Cv }) {
  const { t } = useTranslation();
  const { lang } = useLanguage();
  const item = useStaggerItem();

  return (
    <motion.article
      className={styles.page}
      variants={staggerContainer}
      initial="hidden"
      animate="show"
    >
      <motion.header className={styles.header} variants={item}>
        <h1>{main.fullName}</h1>
        <p className={styles.headline}>{main.headline}</p>
        <p className={styles.location}>
          <MapPin size={16} aria-hidden />
          <span className="visually-hidden">{t('main.location')}: </span>
          {main.location}
        </p>
        <SocialLinks links={main.links} printUrls label={t('cv.contacts')} />
      </motion.header>

      <Section id="cv-summary" title={t('cv.summary')}>
        <p className={styles.summary}>{cv.summary}</p>
      </Section>

      {cv.experience.length > 0 && (
        <Section id="cv-experience" title={t('cv.experience')}>
          <ol className={styles.timeline}>
            {cv.experience.map((e) => (
              <ExperienceItem key={e.id} item={e} lang={lang} />
            ))}
          </ol>
        </Section>
      )}

      {cv.education.length > 0 && (
        <Section id="cv-education" title={t('cv.education')}>
          <ol className={styles.timeline}>
            {cv.education.map((e) => (
              <EducationItem key={e.id} item={e} lang={lang} />
            ))}
          </ol>
        </Section>
      )}

      {cv.skills.length > 0 && (
        <Section id="cv-skills" title={t('cv.skills')}>
          <dl className={styles.skills}>
            {cv.skills.map((g) => (
              <div key={g.category} className={styles.skillGroup}>
                <dt className={styles.skillCategory}>{g.category}</dt>
                <dd>
                  <ChipList items={g.items} />
                </dd>
              </div>
            ))}
          </dl>
        </Section>
      )}

      {cv.languages.length > 0 && (
        <Section id="cv-languages" title={t('cv.languages')}>
          <ul className={styles.languages}>
            {cv.languages.map((l) => (
              <li key={l.name}>
                <span className={styles.languageName}>{l.name}</span>
                <span className={styles.dash} aria-hidden="true">
                  {' — '}
                </span>
                <span className="visually-hidden">: </span>
                <span className={styles.muted}>{l.level}</span>
              </li>
            ))}
          </ul>
        </Section>
      )}
    </motion.article>
  );
}

function Period({ start, end, lang }: { start: string; end: string | null; lang: Lang }) {
  return (
    <p className={styles.period}>
      <span>{formatPeriod(start, end, lang)}</span>
      <span aria-hidden="true"> · </span>
      <span className="visually-hidden">, </span>
      <span>{formatDuration(start, end, lang)}</span>
    </p>
  );
}

function ExperienceItem({ item, lang }: { item: Experience; lang: Lang }) {
  const { t } = useTranslation();
  return (
    <li className={styles.entry}>
      <h3 className={styles.entryTitle}>{item.position}</h3>
      <p className={styles.entryMeta}>
        {item.companyUrl ? (
          <a
            href={item.companyUrl}
            target="_blank"
            rel="noopener noreferrer"
            data-print-url={item.companyUrl}
            className={styles.company}
          >
            {item.company}
            <span className="visually-hidden"> {t('links.opensInNewTab')}</span>
          </a>
        ) : (
          <span className={styles.company}>{item.company}</span>
        )}
        <span className={styles.muted}> · {item.location}</span>
      </p>
      <Period start={item.startDate} end={item.endDate} lang={lang} />
      {item.projects.map((p, i) => (
        <div key={i} className={styles.project}>
          {p.description && <p className={styles.description}>{p.description}</p>}
          {p.highlights.length > 0 && (
            <ul className={styles.highlights}>
              {p.highlights.map((h) => (
                <li key={h}>{h}</li>
              ))}
            </ul>
          )}
        </div>
      ))}
    </li>
  );
}

function EducationItem({ item, lang }: { item: Education; lang: Lang }) {
  return (
    <li className={styles.entry}>
      <h3 className={styles.entryTitle}>
        {item.degree}, {item.field}
      </h3>
      <p className={styles.entryMeta}>
        <span className={styles.company}>{item.institution}</span>
      </p>
      <Period start={item.startDate} end={item.endDate} lang={lang} />
      {item.description && <p className={styles.description}>{item.description}</p>}
    </li>
  );
}

function CvSkeleton() {
  return (
    <LoadingRegion>
      <div className={styles.page}>
        <Skeleton width="55%" height="2.5rem" />
        <Skeleton width="35%" height="1.25rem" />
        <Skeleton width="100%" height="5rem" />
        <Skeleton width="100%" height="9rem" />
        <Skeleton width="100%" height="9rem" />
      </div>
    </LoadingRegion>
  );
}
