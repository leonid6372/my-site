import { ImageOff } from 'lucide-react';
import { useState, type CSSProperties } from 'react';
import { useTranslation } from 'react-i18next';
import type { Photo } from '../api/types';
import { photoAlt, photoSrc, type PhotoVariant } from '../lib/photos';
import styles from './RemoteImage.module.css';

type Status = 'loading' | 'loaded' | 'error';

interface Props {
  photo: Photo;
  /** thumb: grids (thumbUrl ?? url); full: lightbox; portrait: main-page photo, loaded eagerly. */
  variant: PhotoVariant;
  /** intrinsic: aspect-ratio from width/height when both are known; square: always 1:1. */
  aspect?: 'intrinsic' | 'square';
  fit?: 'cover' | 'contain';
  className?: string;
}

/**
 * External photo (spec §4.5): skeleton until loaded, then fade in; a neutral
 * placeholder instead of a broken image; no referrer sent to third-party hosts.
 */
export function RemoteImage({
  photo,
  variant,
  aspect = 'intrinsic',
  fit = 'cover',
  className,
}: Props) {
  const { t } = useTranslation();
  const src = photoSrc(photo, variant);
  const alt = photoAlt(photo);
  // Status belongs to one src; a new src starts loading again.
  const [state, setState] = useState<{ src: string; status: Status }>({ src, status: 'loading' });
  const status = state.src === src ? state.status : 'loading';

  const hasSize = photo.width !== null && photo.height !== null;
  const style: CSSProperties | undefined =
    aspect === 'intrinsic' && hasSize
      ? { aspectRatio: `${String(photo.width)} / ${String(photo.height)}` }
      : undefined;

  const frameClass = [styles.frame, aspect === 'square' && styles.square, styles[fit], className]
    .filter(Boolean)
    .join(' ');

  return (
    <div className={frameClass} style={style} data-status={status}>
      {status === 'error' ? (
        <div
          className={styles.placeholder}
          role="img"
          aria-label={alt || t('common.imageUnavailable')}
        >
          <ImageOff size={28} aria-hidden />
        </div>
      ) : (
        <img
          src={src}
          alt={alt}
          className={styles.img}
          loading={variant === 'portrait' ? 'eager' : 'lazy'}
          fetchPriority={variant === 'portrait' ? 'high' : undefined}
          decoding="async"
          referrerPolicy="no-referrer"
          width={photo.width ?? undefined}
          height={photo.height ?? undefined}
          onLoad={() => {
            setState({ src, status: 'loaded' });
          }}
          onError={() => {
            setState({ src, status: 'error' });
          }}
        />
      )}
    </div>
  );
}
