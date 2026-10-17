import { createContext, useContext } from 'react';
import type { Lang } from '../../api/types';

export interface LanguageState {
  lang: Lang;
  setLang: (lang: Lang) => void;
}

export const LanguageContext = createContext<LanguageState | null>(null);

/** The active UI/content language (see LanguageProvider). */
export function useLanguage(): LanguageState {
  const ctx = useContext(LanguageContext);
  if (!ctx) throw new Error('useLanguage must be used inside <LanguageProvider>');
  return ctx;
}
