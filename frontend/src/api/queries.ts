import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { useLanguage } from '../app/providers/language-context';
import { apiGet } from './client';
import type { Cv, MainInfo, Travel } from './types';

// Keys include the language so each language is cached separately (spec §6), even
// though the language itself is sent in a header. keepPreviousData keeps the old
// language on screen until the new one arrives.
const common = { staleTime: 5 * 60_000, retry: 1, placeholderData: keepPreviousData } as const;

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
