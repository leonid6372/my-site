import { screen, waitFor } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router';
import { describe, expect, it } from 'vitest';
import { renderWithQuery } from '../test/render';
import { fixtures } from '../test/server';
import { AppRoutes } from './routes';

function CurrentPath() {
  const { pathname, search } = useLocation();
  return <p data-testid="path">{pathname + search}</p>;
}

function renderAt(path: string, lang: 'en' | 'ru' = 'en') {
  return renderWithQuery(
    <MemoryRouter initialEntries={[path]}>
      <AppRoutes />
      <CurrentPath />
    </MemoryRouter>,
    { lang },
  );
}

describe('routes', () => {
  // Pages are identified by their document.title ("<Page> — <Full name>").
  const titleOf = (page: string) => `${page} — ${fixtures.main.en.fullName}`;

  it('redirects / to /main', async () => {
    renderAt('/');
    expect(await screen.findByTestId('path')).toHaveTextContent(/^\/main$/);
    await waitFor(() => {
      expect(document.title).toBe(titleOf('About'));
    });
  });

  it.each([
    ['/main', 'About'],
    ['/cv', 'CV'],
    ['/travel', 'Travel'],
  ])('renders %s', async (path, page) => {
    renderAt(path);
    // /travel is a lazy chunk: its first dynamic import can be slow under a full test run.
    await waitFor(
      () => {
        expect(document.title).toBe(titleOf(page));
      },
      { timeout: 5000 },
    );
    expect(await screen.findByRole('heading', { level: 1 })).toBeInTheDocument();
  });

  it.each(['/nope', '/main/extra', '/en/main', '/profile'])(
    'shows the 404 page with a link to /main for %s',
    async (path) => {
      renderAt(path);
      expect(
        await screen.findByRole('heading', { level: 1, name: 'Page not found' }),
      ).toBeInTheDocument();
      expect(screen.getByRole('link', { name: 'Back to the main page' })).toHaveAttribute(
        'href',
        '/main',
      );
    },
  );

  it('sets a localised document.title with the full name', async () => {
    renderAt('/cv', 'ru');
    await waitFor(() => {
      expect(document.title).toBe(`Резюме — ${fixtures.main.ru.fullName}`);
    });
  });

  it('has exactly one h1 and a focusable main landmark', async () => {
    renderAt('/main');
    await screen.findByRole('heading', { level: 1 });
    expect(screen.getAllByRole('heading', { level: 1 })).toHaveLength(1);
    expect(screen.getByRole('main')).toHaveAttribute('id', 'content');
  });
});
