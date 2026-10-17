import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterAll, afterEach, beforeAll } from 'vitest';
import i18n from '../i18n';
import { requests, server } from './server';

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
