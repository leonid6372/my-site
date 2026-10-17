import { act, renderHook, screen, waitFor, within } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router';
import { describe, expect, it } from 'vitest';
import type { Lang, Travel } from '../../api/types';
import { renderWithQuery } from '../../test/render';
import { fixtures, server } from '../../test/server';
import TravelPage from './TravelPage';
import { useSelection } from './useSelection';

function Location() {
  const { search } = useLocation();
  return <p data-testid="search">{search}</p>;
}

function renderTravel(path = '/travel', lang: Lang = 'en') {
  return renderWithQuery(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/travel" element={<TravelPage />} />
      </Routes>
      <Location />
    </MemoryRouter>,
    { lang },
  );
}

const mapArea = () => {
  const el = document.querySelector('[data-selection]');
  if (!(el instanceof HTMLElement)) throw new Error('map area not rendered');
  return el;
};

/** label → value pairs of the stats <dl> (the screen-reader value, i.e. the final number). */
function statPairs(bar: HTMLElement): Record<string, string> {
  const out: Record<string, string> = {};
  for (const dt of within(bar).getAllByRole('term')) {
    const dd = dt.nextElementSibling;
    out[dt.textContent] = dd?.querySelector('.visually-hidden')?.textContent ?? '';
  }
  return out;
}

describe('StatsBar', () => {
  it('shows the fixture numbers (count-up skipped without IntersectionObserver)', async () => {
    const { stats } = fixtures.travel.en; // 3 countries, 2 continents, 4 places, 1.5 %
    renderTravel();
    const bar = await screen.findByLabelText('Travel statistics');
    expect(statPairs(bar)).toEqual({
      countries: String(stats.countries),
      continents: String(stats.continents),
      places: String(stats.places),
      'of the world': '1.5%',
    });
  });

  it('uses RU plural labels and number format', async () => {
    renderTravel('/travel', 'ru');
    const bar = await screen.findByLabelText('Статистика путешествий');
    const pairs = statPairs(bar);
    // 3 страны, 2 континента, 4 места, 1,5 % мира
    expect(Object.keys(pairs)).toEqual(['страны', 'континента', 'места', 'мира']);
    expect(pairs['мира']?.replace(/\s/g, ' ')).toMatch(/^1,5 ?%$/);
  });
});

describe('selection and ?place=', () => {
  it('selects the place from ?place=amsterdam', async () => {
    renderTravel('/travel?place=amsterdam');
    await waitFor(() => {
      expect(mapArea()).toHaveAttribute('data-selection', 'place:amsterdam');
    });
    expect(screen.getByTestId('search')).toHaveTextContent('?place=amsterdam');
  });

  it('removes an unknown id from the URL', async () => {
    renderTravel('/travel?place=atlantis');
    await waitFor(() => {
      expect(screen.getByTestId('search')).toBeEmptyDOMElement();
    });
    expect(mapArea()).toHaveAttribute('data-selection', '');
  });

  it('mirrors only place selections to the URL', () => {
    const { result } = renderHook(
      () => ({ api: useSelection(fixtures.travel.en.places), location: useLocation() }),
      {
        wrapper: ({ children }) => (
          <MemoryRouter initialEntries={['/travel']}>{children}</MemoryRouter>
        ),
      },
    );
    act(() => {
      result.current.api.select({ type: 'place', id: 'tokyo' });
    });
    expect(result.current.location.search).toBe('?place=tokyo');
    expect(result.current.api.selection).toEqual({ type: 'place', id: 'tokyo' });

    act(() => {
      result.current.api.select({ type: 'country', iso2: 'NL' });
    });
    expect(result.current.location.search).toBe('');
    expect(result.current.api.selection).toEqual({ type: 'country', iso2: 'NL' });

    act(() => {
      result.current.api.select(null);
    });
    expect(result.current.api.selection).toBeNull();
  });
});

describe('empty and error states', () => {
  it('shows "No trips yet" and zero stats with 0 places', async () => {
    const empty: Travel = {
      stats: { countries: 0, continents: 0, places: 0, worldPercent: 0 },
      countries: [],
      places: [],
    };
    server.use(http.get('*/api/v1/travel', () => HttpResponse.json(empty)));
    renderTravel();
    expect(await screen.findByText('No trips yet')).toBeInTheDocument();
    expect(screen.getByLabelText('Travel statistics')).toHaveTextContent('0%');
    expect(screen.getByRole('region', { name: /World map/ })).toBeInTheDocument();
  });

  it('shows the error state', async () => {
    server.use(http.get('*/api/v1/travel', () => HttpResponse.json({}, { status: 500 })));
    renderTravel();
    expect(await screen.findByRole('alert')).toHaveTextContent('Something went wrong');
  });
});
