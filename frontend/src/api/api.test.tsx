import { act, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { describe, expect, it } from 'vitest';
import { useLanguage } from '../app/providers/language-context';
import { renderWithQuery } from '../test/render';
import { failEndpoint, fixtures, requests, server } from '../test/server';
import { ApiError, apiGet } from './client';
import { useCv, useMain, useTravel } from './queries';

describe('apiGet', () => {
  it.each(['en', 'ru'] as const)(
    'sends Accept-Language: %s and no lang in the URL',
    async (lang) => {
      const data = await apiGet('/main', lang);
      expect(data).toEqual(fixtures.main[lang]);
      expect(requests).toHaveLength(1);
      const req = requests[0];
      expect(req?.acceptLanguage).toBe(lang);
      const url = new URL(req?.url ?? '');
      expect(url.pathname).toBe('/api/v1/main');
      expect(url.search).toBe('');
    },
  );

  it('turns a contract error body into ApiError', async () => {
    failEndpoint('cv', 503);
    const err = await apiGet('/cv', 'en').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({
      status: 503,
      code: 'internal_error',
      message: 'internal server error',
    });
  });

  it('turns a non-JSON error into ApiError with http_error', async () => {
    server.use(
      http.get('*/api/v1/main', () => new HttpResponse('<h1>Bad gateway</h1>', { status: 502 })),
    );
    await expect(apiGet('/main', 'en')).rejects.toMatchObject({ status: 502, code: 'http_error' });
  });

  it('turns a network failure into ApiError with network_error', async () => {
    server.use(http.get('*/api/v1/main', () => HttpResponse.error()));
    await expect(apiGet('/main', 'en')).rejects.toMatchObject({ status: 0, code: 'network_error' });
  });
});

function Probe() {
  const { lang, setLang } = useLanguage();
  const main = useMain();
  const cv = useCv();
  const travel = useTravel();
  return (
    <div>
      <p data-testid="name">{main.data?.fullName}</p>
      <p data-testid="summary">{cv.data?.summary}</p>
      <p data-testid="places">{travel.data?.stats.places}</p>
      <p data-testid="placeholder">{String(main.isPlaceholderData)}</p>
      <button
        type="button"
        onClick={() => {
          setLang(lang === 'en' ? 'ru' : 'en');
        }}
      >
        switch
      </button>
    </div>
  );
}

describe('query hooks', () => {
  it('fetch every endpoint with the active language', async () => {
    renderWithQuery(<Probe />, { lang: 'ru' });
    await waitFor(() => {
      expect(screen.getByTestId('name')).toHaveTextContent(fixtures.main.ru.fullName);
    });
    await waitFor(() => {
      expect(screen.getByTestId('places')).toHaveTextContent(
        String(fixtures.travel.ru.stats.places),
      );
    });
    expect(screen.getByTestId('summary')).toHaveTextContent(fixtures.cv.ru.summary);
    expect(requests.map((r) => r.acceptLanguage)).toEqual(['ru', 'ru', 'ru']);
    expect(requests.every((r) => !new URL(r.url).searchParams.has('lang'))).toBe(true);
  });

  it('refetches on a language change and keeps the old data until the new arrives', async () => {
    renderWithQuery(<Probe />, { lang: 'en' });
    await waitFor(() => {
      expect(screen.getByTestId('name')).toHaveTextContent(fixtures.main.en.fullName);
    });
    expect(requests.filter((r) => r.endpoint === 'main')).toHaveLength(1);

    act(() => {
      screen.getByRole('button', { name: 'switch' }).click();
    });
    // Previous (EN) data stays visible as placeholder while RU loads: no empty flash.
    expect(screen.getByTestId('name')).toHaveTextContent(fixtures.main.en.fullName);
    expect(screen.getByTestId('placeholder')).toHaveTextContent('true');

    await waitFor(() => {
      expect(screen.getByTestId('name')).toHaveTextContent(fixtures.main.ru.fullName);
    });
    const mainRequests = requests.filter((r) => r.endpoint === 'main');
    expect(mainRequests.map((r) => r.acceptLanguage)).toEqual(['en', 'ru']);
  });

  it('serves a language from the cache when switching back', async () => {
    renderWithQuery(<Probe />, { lang: 'en' });
    const button = screen.getByRole('button', { name: 'switch' });
    await waitFor(() => {
      expect(screen.getByTestId('name')).toHaveTextContent(fixtures.main.en.fullName);
    });
    act(() => {
      button.click();
    });
    await waitFor(() => {
      expect(screen.getByTestId('name')).toHaveTextContent(fixtures.main.ru.fullName);
    });
    act(() => {
      button.click();
    });
    expect(screen.getByTestId('name')).toHaveTextContent(fixtures.main.en.fullName);
    expect(requests.filter((r) => r.endpoint === 'main')).toHaveLength(2); // en, ru — no third fetch
  });
});
