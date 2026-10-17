import { useEffect, useMemo, useState, type ReactNode } from 'react';
import type { Lang } from '../../api/types';
import i18n from '../../i18n';
import { STORAGE_KEYS, storageSet } from '../../lib/storage';
import { detectLang, LanguageContext } from './language-context';

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLang] = useState<Lang>(() => {
    const initial = detectLang();
    // Switch the bundled resources before the first render, so it is never in the wrong language.
    void i18n.changeLanguage(initial);
    return initial;
  });

  useEffect(() => {
    document.documentElement.lang = lang;
    storageSet(STORAGE_KEYS.lang, lang);
    if (i18n.language !== lang) void i18n.changeLanguage(lang);
  }, [lang]);

  const value = useMemo(() => ({ lang, setLang }), [lang]);
  return <LanguageContext value={value}>{children}</LanguageContext>;
}
