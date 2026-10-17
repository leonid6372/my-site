// MSW server for tests. Fixtures were captured from the real API (placeholder content),
// which B21 verified against api/contract.md. The travel fixture has a 30-photo place
// (amsterdam) and a 0-photo place (rotterdam).
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import type { Cv, Lang, MainInfo, Travel } from '../api/types';
import cvEn from './fixtures/cv.en.json';
import cvRu from './fixtures/cv.ru.json';
import mainEn from './fixtures/main.en.json';
import mainRu from './fixtures/main.ru.json';
import travelEn from './fixtures/travel.en.json';
import travelRu from './fixtures/travel.ru.json';

export const fixtures = {
  main: { en: mainEn as MainInfo, ru: mainRu as MainInfo },
  cv: { en: cvEn as Cv, ru: cvRu as Cv },
  travel: { en: travelEn as Travel, ru: travelRu as Travel },
};

export type Endpoint = keyof typeof fixtures;

/** Every API request seen by the handlers, for header/URL assertions. */
export const requests: { endpoint: Endpoint; url: string; acceptLanguage: string | null }[] = [];

function handler(endpoint: Endpoint) {
  return http.get(`*/api/v1/${endpoint}`, ({ request }) => {
    const acceptLanguage = request.headers.get('accept-language');
    requests.push({ endpoint, url: request.url, acceptLanguage });
    const lang: Lang = acceptLanguage === 'ru' ? 'ru' : 'en';
    return HttpResponse.json(fixtures[endpoint][lang], {
      headers: { 'Content-Language': lang, Vary: 'Accept-Language' },
    });
  });
}

export const handlers = (Object.keys(fixtures) as Endpoint[]).map(handler);

export const server = setupServer(...handlers);

/** Makes `endpoint` answer with a contract error (status 500 by default). */
export function failEndpoint(endpoint: Endpoint, status = 500) {
  server.use(
    http.get(`*/api/v1/${endpoint}`, () =>
      HttpResponse.json(
        { error: { code: 'internal_error', message: 'internal server error' } },
        { status },
      ),
    ),
  );
}
