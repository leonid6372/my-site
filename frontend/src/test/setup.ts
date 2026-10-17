import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterAll, afterEach, beforeAll, vi } from 'vitest';
import i18n from '../i18n';
import { requests, server } from './server';

// Leaflet needs real layout; every test uses a stand-in map (map rendering is a
// manual check, plan §7).
vi.mock('../features/travel/map/TravelMap', () => import('./mockTravelMap'));

// jsdom does not implement layout APIs.
Element.prototype.scrollIntoView = function scrollIntoView() {};
window.scrollTo = () => {};

beforeAll(() => {
  // MSW 3 name for v2's onUnhandledRequest: any request without a handler fails the test.
  server.listen({ onUnhandledFrame: 'error' });
});

afterEach(() => {
  cleanup();
  server.resetHandlers();
  requests.length = 0;
  localStorage.clear();
  document.documentElement.removeAttribute('data-theme');
  document.documentElement.lang = '';
  void i18n.changeLanguage('en');
});

afterAll(() => {
  server.close();
});
