import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, type RenderOptions } from '@testing-library/react';
import type { ReactNode } from 'react';
import type { Lang } from '../api/types';
import { LanguageProvider } from '../app/providers/LanguageProvider';
import { ThemeProvider } from '../app/providers/ThemeProvider';

export function testQueryClient() {
  // Hooks set retry: 1 per query; a zero delay keeps error tests fast.
  return new QueryClient({ defaultOptions: { queries: { retryDelay: 0 } } });
}

/**
 * Renders with the real Query, Theme and Language providers. The language is chosen
 * the way a returning visitor's would be: from the stored choice.
 */
export function renderWithQuery(
  ui: ReactNode,
  { lang = 'en', ...options }: { lang?: Lang } & Omit<RenderOptions, 'wrapper'> = {},
) {
  localStorage.setItem('lang', lang);
  const client = testQueryClient();
  return {
    client,
    ...render(ui, {
      wrapper: ({ children }) => (
        <QueryClientProvider client={client}>
          <ThemeProvider>
            <LanguageProvider>{children}</LanguageProvider>
          </ThemeProvider>
        </QueryClientProvider>
      ),
      ...options,
    }),
  };
}
