// UI strings (spec §4.4). Content strings come from the API, already localised.
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import type { Lang } from '../api/types';
import en from './en.json';
import ru from './ru.json';

export const resources = { en: { translation: en }, ru: { translation: ru } } as const;

void i18n.use(initReactI18next).init({
  resources,
  lng: 'en' satisfies Lang,
  fallbackLng: 'en',
  supportedLngs: ['en', 'ru'],
  interpolation: { escapeValue: false }, // React already escapes
  react: { useSuspense: false },
  initAsync: false, // resources are bundled: initialise synchronously
});

export default i18n;
