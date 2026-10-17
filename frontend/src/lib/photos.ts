import type { Photo } from '../api/types';

export type PhotoVariant = 'thumb' | 'full' | 'portrait';

/** Grids use the thumbnail when there is one; the lightbox and portrait use the full image. */
export function photoSrc(photo: Photo, variant: PhotoVariant): string {
  return variant === 'thumb' ? (photo.thumbUrl ?? photo.url) : photo.url;
}

/** alt → caption → "" (spec §4.5). */
export function photoAlt(photo: Photo): string {
  return photo.alt ?? photo.caption ?? '';
}
