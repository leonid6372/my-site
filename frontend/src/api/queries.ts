import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useLanguage } from '../app/providers/language-context';
import { apiGet } from './client';
import type { Cv, MainInfo, Travel } from './types';

// No client-side caching: data is always stale (staleTime 0) and dropped as soon as
// nothing uses it (gcTime 0), so every page view fetches fresh data from the API.
// Keys include the language because it is sent in a header. keepPreviousData keeps
// the old language on screen until the new one arrives.
const common = { staleTime: 0, gcTime: 0, retry: 1, placeholderData: keepPreviousData } as const;

export function useMain() {
  const { lang } = useLanguage();
  return useQuery({
    queryKey: ['main', lang],
    queryFn: ({ signal }) => apiGet<MainInfo>('/main', lang, signal),
    ...common,
  });
}

export function useCv() {
  const { lang } = useLanguage();
  return useQuery({
    queryKey: ['cv', lang],
    queryFn: ({ signal }) => apiGet<Cv>('/cv', lang, signal),
    ...common,
  });
}

export function useTravel() {
  const { lang } = useLanguage();
  return useQuery({
    queryKey: ['travel', lang],
    queryFn: ({ signal }) => apiGet<Travel>('/travel', lang, signal),
    ...common,
  });
}
