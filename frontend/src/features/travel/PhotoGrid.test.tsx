import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';
import { renderWithQuery } from '../../test/render';
import TravelPage from './TravelPage';

function renderTravel(path: string) {
  return renderWithQuery(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/travel" element={<TravelPage />} />
      </Routes>
    </MemoryRouter>,
  );
}

const tiles = (scope: HTMLElement) =>
  within(scope).getAllByRole('button', { name: /^Open photo \d+ of 30$/ });

describe('PhotoGrid and lightbox', () => {
  it('30 photos: 9 tiles + "Show all 30 photos" → 30 tiles', async () => {
    const user = userEvent.setup();
    renderTravel('/travel?place=amsterdam');
    const aside = await screen.findByRole('complementary');
    expect(within(aside).getByRole('heading', { level: 3, name: /Photos/ })).toHaveTextContent(
      '30 photos',
    );
    expect(tiles(aside)).toHaveLength(9);

    await user.click(within(aside).getByRole('button', { name: 'Show all 30 photos' }));
    expect(tiles(aside)).toHaveLength(30);
    expect(within(aside).queryByRole('button', { name: /Show all/ })).toBeNull();
  });

  it('opens at the clicked photo, → moves on, Esc closes the lightbox but not the panel', async () => {
    const user = userEvent.setup();
    renderTravel('/travel?place=amsterdam');
    const aside = await screen.findByRole('complementary');

    await user.click(within(aside).getByRole('button', { name: 'Open photo 3 of 30' }));
    const dialog = await screen.findByRole('dialog');
    expect(dialog).toHaveTextContent('3 / 30');

    await user.keyboard('{ArrowRight}');
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toHaveTextContent('4 / 30');
    });

    await user.keyboard('{Escape}');
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull();
    });
    expect(screen.getByRole('complementary')).toBeInTheDocument(); // panel still open

    await user.keyboard('{Escape}'); // now Esc closes the panel
    await waitFor(() => {
      expect(screen.queryByRole('complementary')).toBeNull();
    });
  });

  it('navigates over all photos, including those hidden behind "Show all"', async () => {
    const user = userEvent.setup();
    renderTravel('/travel?place=amsterdam');
    const aside = await screen.findByRole('complementary');
    await user.click(within(aside).getByRole('button', { name: 'Open photo 9 of 30' }));
    await screen.findByRole('dialog');
    await user.keyboard('{ArrowRight}');
    await waitFor(() => {
      expect(screen.getByRole('dialog')).toHaveTextContent('10 / 30');
    });
  });

  it('Esc still closes the lightbox when focus has drifted out of it', async () => {
    const user = userEvent.setup();
    renderTravel('/travel?place=amsterdam');
    const aside = await screen.findByRole('complementary');
    await user.click(within(aside).getByRole('button', { name: 'Open photo 2 of 30' }));
    await screen.findByRole('dialog');

    // Focus leaves the lightbox (e.g. Tab into the browser UI and back to <body>).
    (document.activeElement as HTMLElement | null)?.blur();
    expect(document.activeElement).toBe(document.body);

    await user.keyboard('{Escape}');
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull();
    });
    expect(screen.getByRole('complementary')).toBeInTheDocument(); // only the lightbox closed
  });

  it('shows no grid for a place without photos', async () => {
    renderTravel('/travel?place=rotterdam');
    const aside = await screen.findByRole('complementary');
    expect(within(aside).queryByRole('heading', { name: /Photos/ })).toBeNull();
    expect(within(aside).queryByRole('button', { name: /Open photo/ })).toBeNull();
  });
});
