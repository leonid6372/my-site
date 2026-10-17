import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, type RenderOptions } from '@testing-library/react';
import { useMemo, useState, type ReactNode } from 'react';
import type { Lang } from '../api/types';
import { LanguageContext } from '../app/providers/language-context';

export function testQueryClient() {
  // Hooks set retry: 1 per query; a zero delay keeps error tests fast.
  return new QueryClient({ defaultOptions: { queries: { retryDelay: 0 } } });
}

/** A language context without detection/persistence, for hook and component tests. */
export function TestLanguage({
  initial = 'en',
  children,
}: {
  initial?: Lang;
  children: ReactNode;
}) {
  const [lang, setLang] = useState<Lang>(initial);
  const value = useMemo(() => ({ lang, setLang }), [lang]);
  return <LanguageContext value={value}>{children}</LanguageContext>;
}

export function renderWithQuery(
  ui: ReactNode,
  { lang = 'en', ...options }: { lang?: Lang } & Omit<RenderOptions, 'wrapper'> = {},
) {
  const client = testQueryClient();
  return {
    client,
    ...render(ui, {
      wrapper: ({ children }) => (
        <QueryClientProvider client={client}>
          <TestLanguage initial={lang}>{children}</TestLanguage>
        </QueryClientProvider>
      ),
      ...options,
    }),
  };
}
