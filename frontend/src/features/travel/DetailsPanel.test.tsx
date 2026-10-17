import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';
import type { Lang } from '../../api/types';
import { renderWithQuery } from '../../test/render';
import { fixtures } from '../../test/server';
import TravelPage from './TravelPage';

function renderTravel(path = '/travel', lang: Lang = 'en') {
  return renderWithQuery(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/travel" element={<TravelPage />} />
      </Routes>
    </MemoryRouter>,
    { lang },
  );
}

const panel = () => screen.getByRole('complementary');
const amsterdam = fixtures.travel.en.places.find((p) => p.id === 'amsterdam');
if (!amsterdam) throw new Error('fixture without amsterdam');

describe('DetailsPanel', () => {
  it('shows every field of a selected place and focuses its heading', async () => {
    const user = userEvent.setup();
    renderTravel();
    await user.click(await screen.findByRole('button', { name: 'pin amsterdam' }));

    const aside = panel();
    const heading = within(aside).getByRole('heading', { level: 2, name: amsterdam.name });
    expect(heading).toHaveFocus();
    expect(aside).toHaveTextContent('Netherlands');
    // Two visits: "×2" badge and both date ranges.
    expect(within(aside).getByText('×2')).toBeInTheDocument();
    const visits = within(aside).getAllByRole('listitem');
    expect(visits.length).toBeGreaterThanOrEqual(amsterdam.visits.length);
    expect(aside).toHaveTextContent(/Aug 1\W+4, 2024/);
    // Story paragraphs split on blank lines.
    for (const p of amsterdam.story.split(/\n{2,}/)) {
      expect(within(aside).getByText(p)).toBeInTheDocument();
    }
  });

  it('Esc closes it and returns focus to the trigger', async () => {
    const user = userEvent.setup();
    renderTravel();
    const pin = await screen.findByRole('button', { name: 'pin tokyo' });
    await user.click(pin);
    expect(panel()).toBeInTheDocument();

    await user.keyboard('{Escape}');
    await waitFor(() => {
      expect(screen.queryByRole('complementary')).toBeNull();
    });
    expect(pin).toHaveFocus();
    expect(document.querySelector('[data-selection]')).toHaveAttribute('data-selection', '');
  });

  it('the close button closes it too', async () => {
    const user = userEvent.setup();
    renderTravel('/travel?place=tokyo');
    await user.click(await screen.findByRole('button', { name: 'Close details' }));
    await waitFor(() => {
      expect(screen.queryByRole('complementary')).toBeNull();
    });
  });

  it('a country lists its places; clicking one switches to place details', async () => {
    const user = userEvent.setup();
    renderTravel();
    await user.click(await screen.findByRole('button', { name: 'country NL' }));

    const aside = panel();
    expect(within(aside).getByRole('heading', { level: 2, name: 'Netherlands' })).toHaveFocus();
    expect(aside).toHaveTextContent('Continent: Europe');
    const list = within(aside)
      .getAllByRole('button')
      .filter((b) => b.textContent !== '');
    const names = list.map((b) => b.textContent);
    expect(names.some((n) => n.startsWith('Amsterdam'))).toBe(true);
    expect(names.some((n) => n.startsWith('Rotterdam'))).toBe(true);
    expect(names.some((n) => n.startsWith('Tokyo'))).toBe(false);

    await user.click(within(aside).getByRole('button', { name: /^Rotterdam/ }));
    expect(within(panel()).getByRole('heading', { level: 2, name: 'Rotterdam' })).toHaveFocus();
  });

  it('is localised', async () => {
    renderTravel('/travel?place=amsterdam', 'ru');
    const aside = await screen.findByRole('complementary');
    expect(within(aside).getByRole('heading', { level: 2, name: 'Амстердам' })).toBeInTheDocument();
    expect(aside).toHaveTextContent('Нидерланды');
    expect(aside).toHaveTextContent('Поездки');
    expect(within(aside).getByRole('button', { name: 'Закрыть подробности' })).toBeInTheDocument();
  });
});
