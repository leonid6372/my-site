import type { CSSProperties, ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import styles from './Skeleton.module.css';

interface Props {
  width?: CSSProperties['width'];
  height?: CSSProperties['height'];
  radius?: CSSProperties['borderRadius'];
  className?: string;
}

/** A neutral placeholder block (FR-F11: skeletons, no page spinners). */
export function Skeleton({ width = '100%', height = '1rem', radius, className }: Props) {
  return (
    <span
      className={[styles.block, className].filter(Boolean).join(' ')}
      style={{ width, height, borderRadius: radius }}
      aria-hidden="true"
    />
  );
}

/** Wraps a page skeleton: announces loading to screen readers once. */
export function LoadingRegion({ children }: { children: ReactNode }) {
  const { t } = useTranslation();
  return (
    <div aria-busy="true" role="status">
      <span className="visually-hidden">{t('common.loading')}</span>
      {children}
    </div>
  );
}
