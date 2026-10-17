import { lazy, Suspense } from 'react';
import { Navigate, Route, Routes } from 'react-router';
import CvPage from '../features/cv/CvPage';
import MainPage from '../features/main/MainPage';
import NotFoundPage from '../pages/NotFoundPage';
import { Layout } from './Layout';

// The Travel page (Leaflet, map data) is a separate chunk, loaded on first visit.
const TravelPage = lazy(() => import('../features/travel/TravelPage'));

/** Route table (spec §2). `/` redirects (replace) to `/main`. */
export function AppRoutes() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Navigate to="/main" replace />} />
        <Route path="main" element={<MainPage />} />
        <Route path="cv" element={<CvPage />} />
        <Route
          path="travel"
          element={
            <Suspense fallback={null}>
              <TravelPage />
            </Suspense>
          }
        />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
