import { createContext, useContext } from 'react';
import { STORAGE_KEYS, storageGet } from '../../lib/storage';

export type Theme = 'light' | 'dark';

export interface ThemeState {
  theme: Theme;
  toggleTheme: () => void;
}

export const ThemeContext = createContext<ThemeState | null>(null);

export function useTheme(): ThemeState {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error('useTheme must be used inside <ThemeProvider>');
  return ctx;
}

export const DARK_QUERY = '(prefers-color-scheme: dark)';

export function osTheme(): Theme {
  return typeof window.matchMedia === 'function' && window.matchMedia(DARK_QUERY).matches
    ? 'dark'
    : 'light';
}

export function storedTheme(): Theme | null {
  const v = storageGet(STORAGE_KEYS.theme);
  return v === 'light' || v === 'dark' ? v : null;
}

/** The theme already applied by the inline script in index.html, else stored → OS. */
export function initialTheme(): Theme {
  const applied = document.documentElement.getAttribute('data-theme');
  if (applied === 'light' || applied === 'dark') return applied;
  return storedTheme() ?? osTheme();
}
