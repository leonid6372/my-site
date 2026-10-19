// Copied 1:1 from specs/api/contract.md §3. Change the contract first, then this file.

export type Lang = 'en' | 'ru';

export interface Photo {
  url: string;
  thumbUrl: string | null;
  alt: string | null;
  caption: string | null;
  width: number | null;
  height: number | null;
}
export interface Link {
  kind: 'email' | 'github' | 'linkedin' | 'telegram';
  label: string;
  url: string;
}

export interface MainInfo {
  fullName: string;
  headline: string;
  location: string;
  currently: string;
  bio: string[];
  photo: Photo;
  interests: string[];
  links: Link[];
}

export interface Experience {
  id: string;
  position: string;
  company: string;
  companyUrl: string | null;
  location: string;
  startDate: string;
  endDate: string | null;
  projects: Project[];
}
/** One description within a job, with its own highlights. */
export interface Project {
  description: string | null;
  highlights: string[];
}
export interface Education {
  id: string;
  institution: string;
  degree: string;
  field: string;
  startDate: string;
  endDate: string | null;
  description: string | null;
}
export interface Cv {
  summary: string;
  experience: Experience[];
  education: Education[];
  skills: { category: string; items: string[] }[];
  languages: { name: string; level: string }[];
}

export type Continent =
  'africa' | 'antarctica' | 'asia' | 'europe' | 'north-america' | 'oceania' | 'south-america';
export interface Visit {
  startDate: string;
  endDate: string | null;
}
export interface Place {
  id: string;
  name: string;
  countryIso2: string;
  lat: number;
  lng: number;
  story: string;
  visits: Visit[];
  photos: Photo[];
}
export interface Country {
  iso2: string;
  isoNumeric: string;
  name: string;
  continent: Continent;
  placeIds: string[];
}
export interface Travel {
  stats: { countries: number; continents: number; places: number; worldPercent: number };
  countries: Country[];
  places: Place[];
}

export interface ApiError {
  error: { code: string; message: string };
}
