import { act, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import i18n from '../../i18n';
import { detectLang, useLanguage } from './language-context';
import { LanguageProvider } from './LanguageProvider';
import { useTheme } from './theme-context';
import { ThemeProvider } from './ThemeProvider';

function setNavigatorLanguage(lang: string) {
  vi.spyOn(navigator, 'language', 'get').mockReturnValue(lang);
}

function mockMatchMedia(dark: boolean) {
  const listeners = new Set<() => void>();
  const mq = {
    matches: dark,
    media: '(prefers-color-scheme: dark)',
    addEventListener: (_: string, l: () => void) => listeners.add(l),
    removeEventListener: (_: string, l: () => void) => listeners.delete(l),
  };
  vi.stubGlobal('matchMedia', () => mq);
  return {
    change(nowDark: boolean) {
      mq.matches = nowDark;
      listeners.forEach((l) => {
        l();
      });
    },
  };
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
  void i18n.changeLanguage('en');
});

function LangProbe() {
  const { lang, setLang } = useLanguage();
  return (
    <>
      <p data-testid="lang">{lang}</p>
      <p data-testid="ui">{i18n.t('nav.travel')}</p>
      <button
        type="button"
        onClick={() => {
          setLang('ru');
        }}
      >
        ru
      </button>
    </>
  );
}

describe('LanguageProvider', () => {
  it('detects ru from navigator.language = ru-RU', () => {
    setNavigatorLanguage('ru-RU');
    expect(detectLang()).toBe('ru');
    render(
      <LanguageProvider>
        <LangProbe />
      </LanguageProvider>,
    );
    expect(screen.getByTestId('lang')).toHaveTextContent('ru');
    expect(screen.getByTestId('ui')).toHaveTextContent('Путешествия');
    expect(document.documentElement.lang).toBe('ru');
  });

  it('falls back to en for other browser languages', () => {
    setNavigatorLanguage('de-DE');
    expect(detectLang()).toBe('en');
  });

  it('a stored choice wins over the browser language', () => {
    setNavigatorLanguage('ru-RU');
    localStorage.setItem('lang', 'en');
    expect(detectLang()).toBe('en');
  });

  it('ignores an invalid stored value', () => {
    setNavigatorLanguage('ru');
    localStorage.setItem('lang', 'fr');
    expect(detectLang()).toBe('ru');
  });

  it('switching updates <html lang>, storage and UI strings', () => {
    setNavigatorLanguage('en-US');
    render(
      <LanguageProvider>
        <LangProbe />
      </LanguageProvider>,
    );
    expect(document.documentElement.lang).toBe('en');
    act(() => {
      screen.getByRole('button', { name: 'ru' }).click();
    });
    expect(document.documentElement.lang).toBe('ru');
    expect(localStorage.getItem('lang')).toBe('ru');
    expect(i18n.language).toBe('ru');
  });

  it('works when localStorage throws', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('SecurityError');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('QuotaExceededError');
    });
    setNavigatorLanguage('ru-RU');
    render(
      <LanguageProvider>
        <LangProbe />
      </LanguageProvider>,
    );
    expect(screen.getByTestId('lang')).toHaveTextContent('ru');
    act(() => {
      screen.getByRole('button', { name: 'ru' }).click();
    });
    expect(document.documentElement.lang).toBe('ru');
  });
});

function ThemeProbe() {
  const { theme, toggleTheme } = useTheme();
  return (
    <button type="button" onClick={toggleTheme}>
      {theme}
    </button>
  );
}

describe('ThemeProvider', () => {
  beforeEach(() => {
    document.documentElement.removeAttribute('data-theme');
  });

  it('uses the theme set by the inline script', () => {
    mockMatchMedia(false);
    document.documentElement.setAttribute('data-theme', 'dark');
    render(
      <ThemeProvider>
        <ThemeProbe />
      </ThemeProvider>,
    );
    expect(screen.getByRole('button')).toHaveTextContent('dark');
  });

  it('falls back to stored, then OS theme', () => {
    mockMatchMedia(true);
    render(
      <ThemeProvider>
        <ThemeProbe />
      </ThemeProvider>,
    );
    expect(screen.getByRole('button')).toHaveTextContent('dark');
    expect(document.documentElement.dataset.theme).toBe('dark');
  });

  it('toggle updates <html data-theme>, color-scheme and storage', () => {
    mockMatchMedia(false);
    render(
      <ThemeProvider>
        <ThemeProbe />
      </ThemeProvider>,
    );
    act(() => {
      screen.getByRole('button').click();
    });
    expect(document.documentElement.dataset.theme).toBe('dark');
    expect(document.documentElement.style.colorScheme).toBe('dark');
    expect(localStorage.getItem('theme')).toBe('dark');
    act(() => {
      screen.getByRole('button').click();
    });
    expect(document.documentElement.dataset.theme).toBe('light');
    expect(localStorage.getItem('theme')).toBe('light');
  });

  it('follows OS changes until the user picks a theme', () => {
    const os = mockMatchMedia(false);
    render(
      <ThemeProvider>
        <ThemeProbe />
      </ThemeProvider>,
    );
    act(() => {
      os.change(true);
    });
    expect(document.documentElement.dataset.theme).toBe('dark');

    act(() => {
      screen.getByRole('button').click(); // explicit choice: light
    });
    act(() => {
      os.change(true);
    });
    expect(document.documentElement.dataset.theme).toBe('light');
  });

  it('works when localStorage throws', () => {
    mockMatchMedia(false);
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    render(
      <ThemeProvider>
        <ThemeProbe />
      </ThemeProvider>,
    );
    act(() => {
      screen.getByRole('button').click();
    });
    expect(document.documentElement.dataset.theme).toBe('dark');
  });
});
