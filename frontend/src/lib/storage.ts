// Safe localStorage access: storage can be missing or throw (private mode, blocked
// site data), and the app must work without it. Keys are shared with the inline
// theme script in index.html.

export const STORAGE_KEYS = { lang: 'lang', theme: 'theme' } as const;

export function storageGet(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function storageSet(key: string, value: string): void {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // ignore: the choice just isn't remembered
  }
}
