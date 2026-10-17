import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router';
import { describe, expect, it } from 'vitest';
import type { Lang } from '../../api/types';
import { renderWithQuery } from '../../test/render';
import { fixtures } from '../../test/server';
import TravelPage from './TravelPage';

function renderTravel(lang: Lang = 'en') {
  return renderWithQuery(
    <MemoryRouter initialEntries={['/travel']}>
      <Routes>
        <Route path="/travel" element={<TravelPage />} />
      </Routes>
    </MemoryRouter>,
    { lang },
  );
}

const listSection = async () =>
  (await screen.findByRole('heading', { level: 2, name: /Countries and places|Страны и места/ }))
    .parentElement as HTMLElement;

describe('PlaceList', () => {
  it.each(['en', 'ru'] as const)('groups places by country in API order (%s)', async (lang) => {
    const travel = fixtures.travel[lang];
    renderTravel(lang);
    const section = await listSection();

    const groups = within(section).getAllByRole('heading', { level: 3 });
    expect(groups.map((h) => h.textContent)).toEqual(
      travel.countries.map(
        (c) =>
          c.name +
          (lang === 'en'
            ? `${c.placeIds.length} ${c.placeIds.length === 1 ? 'place' : 'places'}`
            : `${c.placeIds.length} ${c.placeIds.length === 1 ? 'место' : 'места'}`),
      ),
    );

    for (const c of travel.countries) {
      const group = within(section)
        .getAllByRole('heading', { level: 3 })
        .find((h) => h.textContent.startsWith(c.name))?.parentElement;
      if (!group) throw new Error(`no group for ${c.iso2}`);
      const names = within(group)
        .getAllByRole('button')
        .map((b) => b.textContent);
      const want = c.placeIds.map((id) => travel.places.find((p) => p.id === id)?.name);
      expect(names.map((n) => want.find((w) => w && n.startsWith(w)))).toEqual(want);
      expect(group.querySelector('img')).toHaveAttribute('alt', ''); // decorative flag
    }
  });

  it('shows the latest visit year and announces it', async () => {
    renderTravel();
    const section = await listSection();
    const amsterdam = within(section).getByRole('button', { name: /^Amsterdam/ });
    expect(amsterdam).toHaveTextContent('2024');
    expect(amsterdam).toHaveAccessibleName('Amsterdam, last visit 2024');
  });

  it('activating a place with the keyboard opens the panel', async () => {
    const user = userEvent.setup();
    renderTravel();
    const section = await listSection();
    const tokyo = within(section).getByRole('button', { name: /^Tokyo/ });
    tokyo.focus();
    await user.keyboard('{Enter}');

    const aside = await screen.findByRole('complementary');
    expect(within(aside).getByRole('heading', { level: 2, name: 'Tokyo' })).toHaveFocus();
    expect(tokyo).toHaveAttribute('aria-current', 'true');

    await user.keyboard('{Escape}');
    await waitFor(() => {
      expect(screen.queryByRole('complementary')).toBeNull();
    });
    expect(tokyo).toHaveFocus(); // focus returns to the list button
  });
});
