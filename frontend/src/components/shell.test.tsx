import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';
import { AppRoutes } from '../app/routes';
import { renderWithQuery } from '../test/render';
import { fixtures, requests } from '../test/server';
import { printTarget } from '../lib/links';
import { SocialLinks } from './SocialLinks';

function renderApp(path = '/main', lang: 'en' | 'ru' = 'en') {
  return renderWithQuery(
    <MemoryRouter initialEntries={[path]}>
      <AppRoutes />
    </MemoryRouter>,
    { lang },
  );
}

describe('header navigation', () => {
  it.each([
    ['/main', 'Main'],
    ['/travel', 'Travel'],
    ['/cv', 'CV'],
  ])('marks %s as the current page', async (path, label) => {
    renderApp(path);
    const nav = screen.getByRole('navigation', { name: 'Main navigation' });
    await waitFor(() => {
      expect(within(nav).getByRole('link', { name: label })).toHaveAttribute(
        'aria-current',
        'page',
      );
    });
    const others = within(nav)
      .getAllByRole('link')
      .filter((a) => a.textContent !== label);
    others.forEach((a) => {
      expect(a).not.toHaveAttribute('aria-current');
    });
  });

  it('links the site name to /main', async () => {
    renderApp('/cv');
    const home = await screen.findByRole('link', {
      name: `${fixtures.main.en.fullName}, main page`,
    });
    expect(home).toHaveAttribute('href', '/main');
  });

  it('navigates between pages', async () => {
    const user = userEvent.setup();
    renderApp('/main');
    const nav = screen.getByRole('navigation', { name: 'Main navigation' });
    await user.click(within(nav).getByRole('link', { name: 'CV' }));
    expect(
      await screen.findByRole('heading', { level: 2, name: 'Experience' }),
    ).toBeInTheDocument();
  });
});

describe('language switch', () => {
  it('has aria-pressed buttons and switches UI text and content', async () => {
    const user = userEvent.setup();
    renderApp('/main', 'en');
    const group = screen.getByRole('group', { name: 'Language' });
    const en = within(group).getByRole('button', { name: 'EN' });
    const ru = within(group).getByRole('button', { name: 'RU' });
    expect(en).toHaveAttribute('aria-pressed', 'true');
    expect(ru).toHaveAttribute('aria-pressed', 'false');

    await user.click(ru);

    expect(
      within(screen.getByRole('group', { name: 'Язык' })).getByRole('button', { name: 'RU' }),
    ).toHaveAttribute('aria-pressed', 'true');
    const nav = screen.getByRole('navigation', { name: 'Основная навигация' });
    expect(within(nav).getByRole('link', { name: 'Путешествия' })).toBeInTheDocument();
    await screen.findByRole('link', { name: `${fixtures.main.ru.fullName}, главная страница` });
    expect(requests.some((r) => r.acceptLanguage === 'ru')).toBe(true);
    expect(document.documentElement.lang).toBe('ru');
  });
});

describe('theme toggle', () => {
  it('has a descriptive label that flips with the theme', async () => {
    const user = userEvent.setup();
    renderApp();
    const toggle = screen.getByRole('button', { name: 'Switch to dark theme' });
    await user.click(toggle);
    expect(document.documentElement.dataset.theme).toBe('dark');
    expect(screen.getByRole('button', { name: 'Switch to light theme' })).toBeInTheDocument();
  });
});

describe('mobile menu', () => {
  it('toggles aria-expanded, closes on Esc and after navigation', async () => {
    const user = userEvent.setup();
    renderApp('/main');
    const button = screen.getByRole('button', { name: 'Open menu' });
    const nav = screen.getByRole('navigation', { name: 'Main navigation' });
    expect(button).toHaveAttribute('aria-expanded', 'false');
    expect(button).toHaveAttribute('aria-controls', nav.id);

    await user.click(button);
    expect(button).toHaveAttribute('aria-expanded', 'true');
    expect(nav).toHaveAttribute('data-open');

    await user.keyboard('{Escape}');
    expect(button).toHaveAttribute('aria-expanded', 'false');
    expect(button).toHaveFocus();

    await user.click(button);
    await user.click(within(nav).getByRole('link', { name: 'Travel' }));
    await waitFor(() => {
      expect(button).toHaveAttribute('aria-expanded', 'false');
    });
  });
});

describe('skip link and landmarks', () => {
  it('is the first focusable element and moves focus to main', async () => {
    const user = userEvent.setup();
    renderApp();
    await user.tab();
    const skip = screen.getByRole('link', { name: 'Skip to content' });
    expect(skip).toHaveFocus();
    await user.keyboard('{Enter}');
    expect(screen.getByRole('main')).toHaveFocus();
  });

  it('has banner, navigation, main and contentinfo landmarks', async () => {
    renderApp();
    expect(screen.getByRole('banner')).toBeInTheDocument();
    expect(screen.getByRole('main')).toBeInTheDocument();
    expect(screen.getByRole('contentinfo')).toBeInTheDocument();
    const footer = screen.getByRole('contentinfo');
    await within(footer).findByText(`© ${new Date().getFullYear()} ${fixtures.main.en.fullName}`);
  });
});

describe('SocialLinks', () => {
  const links = fixtures.main.en.links;

  it('opens mailto in the same tab and the others in a new tab with rel', () => {
    renderWithQuery(<SocialLinks links={links} />);
    for (const link of links) {
      const a = screen.getByRole('link', { name: new RegExp(`^${link.label}`) });
      expect(a).toHaveAttribute('href', link.url);
      if (link.kind === 'email') {
        expect(a.getAttribute('href')).toMatch(/^mailto:/);
        expect(a).not.toHaveAttribute('target');
        expect(a).not.toHaveAttribute('rel');
      } else {
        expect(a).toHaveAttribute('target', '_blank');
        expect(a).toHaveAttribute('rel', 'noopener noreferrer');
        expect(a).toHaveTextContent('(opens in a new tab)');
      }
    }
  });

  it('icon variant still has an accessible name per link', () => {
    renderWithQuery(<SocialLinks links={links} variant="icons" label="Contact links" />);
    const list = screen.getByRole('list', { name: 'Contact links' });
    expect(within(list).getAllByRole('link')).toHaveLength(links.length);
    expect(within(list).getByRole('link', { name: 'GitHub' })).toBeInTheDocument();
  });

  it('prints the plain email address and URLs when printUrls is set', () => {
    renderWithQuery(<SocialLinks links={links} printUrls />);
    const email = links.find((l) => l.kind === 'email');
    expect(email).toBeDefined();
    if (!email) return;
    expect(printTarget(email)).toBe(email.url.replace('mailto:', ''));
    const a = screen.getByRole('link', { name: new RegExp(`^${email.label}`) });
    expect(a).toHaveAttribute('data-print-url', email.url.replace('mailto:', ''));
  });
});
