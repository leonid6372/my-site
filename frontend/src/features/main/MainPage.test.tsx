import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router';
import { describe, expect, it } from 'vitest';
import type { Lang } from '../../api/types';
import { renderWithQuery } from '../../test/render';
import { failEndpoint, fixtures, server } from '../../test/server';
import MainPage from './MainPage';

function renderMain(lang: Lang = 'en') {
  return renderWithQuery(
    <MemoryRouter>
      <MainPage />
    </MemoryRouter>,
    { lang },
  );
}

describe('MainPage', () => {
  it.each(['en', 'ru'] as const)('renders every field of the %s fixture', async (lang) => {
    const main = fixtures.main[lang];
    renderMain(lang);

    const h1 = await screen.findByRole('heading', { level: 1 });
    expect(h1).toHaveTextContent(main.fullName);
    expect(screen.getAllByRole('heading', { level: 1 })).toHaveLength(1);
    expect(screen.getByText(main.headline)).toBeInTheDocument();
    expect(screen.getByText(main.location)).toBeInTheDocument();
    expect(screen.getByText(main.currently)).toBeInTheDocument();
    for (const p of main.bio) expect(screen.getByText(p)).toBeInTheDocument();
    for (const i of main.interests) expect(screen.getByText(i)).toBeInTheDocument();
    for (const l of main.links) {
      expect(screen.getByRole('link', { name: new RegExp(`^${l.label}`) })).toHaveAttribute(
        'href',
        l.url,
      );
    }
    const portrait = screen.getByRole('img', { name: main.photo.alt ?? '' });
    expect(portrait).toHaveAttribute('src', main.photo.url);
    expect(portrait).toHaveAttribute('fetchpriority', 'high');
  });

  it('links the CTA cards to /travel and /cv', async () => {
    renderMain();
    await screen.findByRole('heading', { level: 1 });
    expect(screen.getByRole('link', { name: /See where I've travelled/ })).toHaveAttribute(
      'href',
      '/travel',
    );
    expect(screen.getByRole('link', { name: /View my CV/ })).toHaveAttribute('href', '/cv');
  });

  it('uses RU UI labels for the RU fixture', async () => {
    renderMain('ru');
    await screen.findByRole('heading', { level: 1 });
    expect(screen.getByRole('heading', { name: 'Обо мне' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Посмотреть резюме/ })).toBeInTheDocument();
  });

  it('shows a skeleton while loading', () => {
    renderMain();
    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true');
  });

  it('shows the error state and recovers with Retry', async () => {
    failEndpoint('main');
    renderMain();
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Something went wrong');

    server.resetHandlers(); // the API is back
    await userEvent.click(within(alert).getByRole('button', { name: 'Retry' }));
    expect(await screen.findByRole('heading', { level: 1 })).toHaveTextContent(
      fixtures.main.en.fullName,
    );
  });
});
