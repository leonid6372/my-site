import { Menu, Moon, Sun, X } from 'lucide-react';
import { useEffect, useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Link, NavLink } from 'react-router';
import { useMain } from '../api/queries';
import type { Lang } from '../api/types';
import { useLanguage } from '../app/providers/language-context';
import { useTheme } from '../app/providers/theme-context';
import styles from './Header.module.css';

const NAV = [
  { to: '/', key: 'nav.main' },
  { to: '/travel', key: 'nav.travel' },
  { to: '/cv', key: 'nav.cv' },
] as const;

const LANGS: { lang: Lang; label: string }[] = [
  { lang: 'en', label: 'EN' },
  { lang: 'ru', label: 'RU' },
];

export function LangSwitch() {
  const { t } = useTranslation();
  const { lang, setLang } = useLanguage();
  return (
    <div role="group" aria-label={t('header.language')} className={styles.langSwitch}>
      {LANGS.map((l) => (
        <button
          key={l.lang}
          type="button"
          lang={l.lang}
          aria-pressed={lang === l.lang}
          className={styles.langButton}
          onClick={() => {
            setLang(l.lang);
          }}
        >
          {l.label}
        </button>
      ))}
    </div>
  );
}

export function ThemeToggle() {
  const { t } = useTranslation();
  const { theme, toggleTheme } = useTheme();
  const label = theme === 'dark' ? t('header.toLight') : t('header.toDark');
  return (
    <button
      type="button"
      className={styles.iconButton}
      onClick={toggleTheme}
      aria-label={label}
      title={label}
    >
      {theme === 'dark' ? <Sun size={18} aria-hidden /> : <Moon size={18} aria-hidden />}
    </button>
  );
}

/** Site header (spec §4.1): name → /, nav, language switch, theme toggle. */
export function Header() {
  const { t } = useTranslation();
  const name = useMain().data?.fullName;
  const [menuOpen, setMenuOpen] = useState(false);
  const menuId = useId();
  const menuButton = useRef<HTMLButtonElement>(null);
  const closeMenu = () => {
    setMenuOpen(false);
  };

  // Esc closes the mobile menu and returns focus to its button.
  useEffect(() => {
    if (!menuOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setMenuOpen(false);
        menuButton.current?.focus();
      }
    };
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('keydown', onKey);
    };
  }, [menuOpen]);

  return (
    <header className={styles.header} data-print="hide">
      <div className={styles.inner}>
        <Link
          to="/"
          className={styles.brand}
          aria-label={name ? t('header.home', { name }) : undefined}
          onClick={closeMenu}
        >
          {name ?? <>&nbsp;</>}
        </Link>

        <nav
          id={menuId}
          aria-label={t('nav.label')}
          className={styles.nav}
          data-open={menuOpen || undefined}
        >
          <ul className={styles.navList}>
            {NAV.map((item) => (
              <li key={item.to}>
                {/* NavLink sets aria-current="page" on the active link (FR-F02).
                    `end` stops "/" from matching every page. */}
                <NavLink to={item.to} end className={styles.navLink} onClick={closeMenu}>
                  {t(item.key)}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>

        <div className={styles.controls}>
          <LangSwitch />
          <ThemeToggle />
          <button
            ref={menuButton}
            type="button"
            className={`${styles.iconButton} ${styles.menuButton}`}
            aria-expanded={menuOpen}
            aria-controls={menuId}
            aria-label={menuOpen ? t('nav.closeMenu') : t('nav.openMenu')}
            onClick={() => {
              setMenuOpen((o) => !o);
            }}
          >
            {menuOpen ? <X size={20} aria-hidden /> : <Menu size={20} aria-hidden />}
          </button>
        </div>
      </div>
    </header>
  );
}
