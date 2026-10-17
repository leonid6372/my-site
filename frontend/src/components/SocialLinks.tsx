import { Mail } from 'lucide-react';
import type { ComponentType } from 'react';
import { useTranslation } from 'react-i18next';
import type { Link } from '../api/types';
import { printTarget } from '../lib/links';
import { GitHubIcon, LinkedInIcon, TelegramIcon } from './icons/BrandIcons';
import styles from './SocialLinks.module.css';

const ICONS: Record<Link['kind'], ComponentType<{ size?: number; 'aria-hidden'?: boolean }>> = {
  email: Mail,
  github: GitHubIcon,
  linkedin: LinkedInIcon,
  telegram: TelegramIcon,
};

interface Props {
  links: Link[];
  /** "icons": compact icon-only row (footer); "buttons": icon + label (main page, CV). */
  variant?: 'icons' | 'buttons';
  /** Print the target after each link (CV print styles, spec §5.3). */
  printUrls?: boolean;
  label?: string;
}

/**
 * Contact links (spec §5.1): mailto opens in the same tab; GitHub, LinkedIn and
 * Telegram open in a new tab with rel="noopener noreferrer".
 */
export function SocialLinks({ links, variant = 'buttons', printUrls = false, label }: Props) {
  const { t } = useTranslation();
  return (
    <ul className={styles[variant]} aria-label={label}>
      {links.map((link) => {
        const Icon = ICONS[link.kind];
        const external = link.kind !== 'email';
        return (
          <li key={link.kind}>
            <a
              href={link.url}
              className={styles.link}
              {...(external ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
              {...(printUrls ? { 'data-print-url': printTarget(link) } : {})}
              aria-label={variant === 'icons' ? link.label : undefined}
              title={variant === 'icons' ? link.label : undefined}
            >
              <Icon size={variant === 'icons' ? 20 : 18} aria-hidden />
              {variant === 'buttons' && <span>{link.label}</span>}
              {external && <span className="visually-hidden"> {t('links.opensInNewTab')}</span>}
            </a>
          </li>
        );
      })}
    </ul>
  );
}
