import { ImageOff } from 'lucide-react';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import Lightbox, { type Slide } from 'yet-another-react-lightbox';
import Captions from 'yet-another-react-lightbox/plugins/captions';
import 'yet-another-react-lightbox/plugins/captions.css';
import Counter from 'yet-another-react-lightbox/plugins/counter';
import 'yet-another-react-lightbox/plugins/counter.css';
import 'yet-another-react-lightbox/styles.css';
import type { Photo } from '../../api/types';
import { photoAlt } from '../../lib/photos';
import styles from './PhotoLightbox.module.css';

type PhotoSlide = Slide & { photo: Photo };

/**
 * Full-size photos (spec §5.2): counter, captions, arrows/swipe/Esc, focus trap.
 * Only the neighbours are preloaded; the thumbnail is shown blurred until the full
 * image arrives; a broken URL shows the placeholder.
 */
export function PhotoLightbox({
  photos,
  index,
  onClose,
}: {
  photos: Photo[];
  index: number | null;
  onClose: () => void;
}) {
  const { t } = useTranslation();
  const slides: PhotoSlide[] = photos.map((photo) => ({
    src: photo.url,
    alt: photoAlt(photo),
    description: photo.caption ?? undefined,
    width: photo.width ?? undefined,
    height: photo.height ?? undefined,
    photo,
  }));

  return (
    <Lightbox
      open={index !== null}
      index={index ?? 0}
      close={onClose}
      slides={slides}
      plugins={[Captions, Counter]}
      carousel={{ preload: 1, finite: false }}
      controller={{ closeOnBackdropClick: true }}
      captions={{ descriptionTextAlign: 'center' }}
      labels={{
        Previous: t('lightbox.previous'),
        Next: t('lightbox.next'),
        Close: t('lightbox.close'),
      }}
      render={{
        // rect = the space the lightbox gives this slide; the photo is fitted into it.
        slide: ({ slide, rect }) => (
          <LightboxSlide
            photo={(slide as PhotoSlide).photo}
            width={rect.width}
            height={rect.height}
          />
        ),
      }}
    />
  );
}

function LightboxSlide({ photo, width, height }: { photo: Photo; width: number; height: number }) {
  const { t } = useTranslation();
  const [status, setStatus] = useState<'loading' | 'loaded' | 'error'>('loading');
  const alt = photoAlt(photo);

  if (status === 'error') {
    return (
      <div
        className={styles.placeholder}
        role="img"
        aria-label={alt || t('common.imageUnavailable')}
      >
        <ImageOff size={40} aria-hidden />
      </div>
    );
  }
  return (
    <div className={styles.slide} style={{ width, height }}>
      {status === 'loading' && photo.thumbUrl && (
        <img
          src={photo.thumbUrl}
          alt=""
          aria-hidden="true"
          className={styles.blur}
          referrerPolicy="no-referrer"
        />
      )}
      <img
        src={photo.url}
        alt={alt}
        className={styles.full}
        data-status={status}
        decoding="async"
        referrerPolicy="no-referrer"
        onLoad={() => {
          setStatus('loaded');
        }}
        onError={() => {
          setStatus('error');
        }}
      />
    </div>
  );
}
