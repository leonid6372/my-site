import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MotionConfig } from 'motion/react';
import { useState } from 'react';
import { BrowserRouter } from 'react-router';
import { LanguageProvider } from './providers/LanguageProvider';
import { ThemeProvider } from './providers/ThemeProvider';
import { AppRoutes } from './routes';

export function App() {
  const [queryClient] = useState(() => new QueryClient());
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <LanguageProvider>
          {/* reducedMotion="user": no transforms when the OS asks for reduced motion */}
          <MotionConfig reducedMotion="user">
            <BrowserRouter>
              <AppRoutes />
            </BrowserRouter>
          </MotionConfig>
        </LanguageProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}
