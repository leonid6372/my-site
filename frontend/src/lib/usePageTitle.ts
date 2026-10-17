import { useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useMain } from '../api/queries';

export type PageKey = 'main' | 'travel' | 'cv' | 'notFound';

/** Sets document.title to "<Page> — <Full name>" in the active language (spec §2). */
export function usePageTitle(page: PageKey) {
  const { t } = useTranslation();
  const name = useMain().data?.fullName;
  const pageTitle = t(`titles.${page}`);
  useEffect(() => {
    document.title = name ? `${pageTitle} — ${name}` : pageTitle;
  }, [pageTitle, name]);
}
