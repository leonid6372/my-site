/// <reference types="vitest/config" />
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

// Plan §6.1. In development the API runs on https://localhost:8080 (go run or the dev
// compose stack) with a self-signed certificate, so the proxy does not verify it.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: { '/api': { target: 'https://localhost:8080', secure: false } },
  },
  build: {
    target: 'es2022',
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['src/test/setup.ts'],
    css: { modules: { classNameStrategy: 'non-scoped' } },
  },
});
