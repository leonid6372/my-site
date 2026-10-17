import { createContext, useContext } from 'react';
import type { Lang } from '../../api/types';
import { STORAGE_KEYS, storageGet } from '../../lib/storage';

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

/** Stored choice → browser language → English (FR-F04). */
export function detectLang(): Lang {
  const stored = storageGet(STORAGE_KEYS.lang);
  if (stored === 'en' || stored === 'ru') return stored;
  const browser = typeof navigator === 'undefined' ? '' : navigator.language;
  return browser.toLowerCase().startsWith('ru') ? 'ru' : 'en';
}
