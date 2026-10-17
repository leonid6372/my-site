import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import { STORAGE_KEYS, storageSet } from '../../lib/storage';
import { DARK_QUERY, initialTheme, storedTheme, ThemeContext, type Theme } from './theme-context';

function apply(theme: Theme) {
  const root = document.documentElement;
  root.setAttribute('data-theme', theme);
  root.style.colorScheme = theme;
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setTheme] = useState<Theme>(initialTheme);

  // Every change applies the attribute *before* the re-render, so children that read
  // computed token values while rendering (the map's colours) see the new theme.
  const change = useCallback((next: Theme) => {
    apply(next);
    setTheme(next);
  }, []);

  // Keep <html> in sync on mount (the inline script normally did this already).
  useEffect(() => {
    apply(theme);
  }, [theme]);

  // Without an explicit choice, follow OS theme changes live (FR-F06).
  useEffect(() => {
    if (typeof window.matchMedia !== 'function') return;
    const mq = window.matchMedia(DARK_QUERY);
    const onChange = () => {
      if (!storedTheme()) change(mq.matches ? 'dark' : 'light');
    };
    mq.addEventListener('change', onChange);
    return () => {
      mq.removeEventListener('change', onChange);
    };
  }, [change]);

  const toggleTheme = useCallback(() => {
    const next = theme === 'dark' ? 'light' : 'dark';
    storageSet(STORAGE_KEYS.theme, next); // an explicit choice: stop following the OS
    change(next);
  }, [theme, change]);

  const value = useMemo(() => ({ theme, toggleTheme }), [theme, toggleTheme]);
  return <ThemeContext value={value}>{children}</ThemeContext>;
}
