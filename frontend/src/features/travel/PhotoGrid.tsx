import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { Photo } from '../../api/types';
import { RemoteImage } from '../../components/RemoteImage';
import styles from './PhotoGrid.module.css';
import { PhotoLightbox } from './PhotoLightbox';

const INITIAL = 9;

/**
 * Up to 30 photos of a place (spec §5.2): square thumbnails, the first 9 at first
 * plus "Show all N photos"; a click opens the lightbox over all photos.
 */
export function PhotoGrid({
  photos,
  onLightboxChange,
}: {
  photos: Photo[];
  onLightboxChange?: (open: boolean) => void;
}) {
  const { t } = useTranslation();
  const [showAll, setShowAll] = useState(false);
  const [index, setIndex] = useState<number | null>(null);

  if (photos.length === 0) return null;
  const visible = showAll ? photos : photos.slice(0, INITIAL);

  const open = (i: number) => {
    setIndex(i);
    onLightboxChange?.(true);
  };
  const close = () => {
    setIndex(null);
    onLightboxChange?.(false);
  };

  return (
    <section aria-labelledby="photos-heading" className={styles.section}>
      <h3 id="photos-heading" className={styles.label}>
        {t('travel.photos')}
        <span className={styles.count}>{t('travel.photoCount', { count: photos.length })}</span>
      </h3>
      <ul className={styles.grid}>
        {visible.map((photo, i) => (
          <li key={photo.url}>
            <button
              type="button"
              className={styles.tile}
              aria-label={t('travel.openPhoto', { index: i + 1, count: photos.length })}
              onClick={() => {
                open(i);
              }}
            >
              <RemoteImage photo={photo} variant="thumb" aspect="square" />
            </button>
          </li>
        ))}
      </ul>
      {!showAll && photos.length > INITIAL && (
        <button
          type="button"
          className={styles.more}
          onClick={() => {
            setShowAll(true);
          }}
        >
          {t('travel.showAll', { count: photos.length })}
        </button>
      )}
      <PhotoLightbox photos={photos} index={index} onClose={close} />
    </section>
  );
}
