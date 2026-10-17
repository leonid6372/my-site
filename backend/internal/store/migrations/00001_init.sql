-- +goose Up
-- Schema from specs/backend/spec.md §7. Localised text (ltext) is JSONB {"en": "...", "ru": "..."}.

-- singleton: main page info
CREATE TABLE main_info (
  id              SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  full_name       JSONB NOT NULL,
  headline        JSONB NOT NULL,
  location        JSONB NOT NULL,
  currently       JSONB NOT NULL,
  bio             JSONB NOT NULL,
  summary         JSONB NOT NULL,
  photo_url       TEXT  NOT NULL,
  photo_thumb_url TEXT,
  photo_alt       JSONB,
  photo_width     INT,
  photo_height    INT,
  interests       JSONB NOT NULL,
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE social_link (
  id          SERIAL PRIMARY KEY,
  kind        TEXT  NOT NULL UNIQUE CHECK (kind IN ('email','github','linkedin','telegram')),
  label       JSONB NOT NULL,
  url         TEXT  NOT NULL,
  sort_order  INT   NOT NULL
);

CREATE TABLE cv_experience (
  id           TEXT PRIMARY KEY,
  position     JSONB NOT NULL,
  company      TEXT  NOT NULL,
  company_url  TEXT,
  location     JSONB NOT NULL,
  start_date   DATE  NOT NULL,
  end_date     DATE,
  description  JSONB,
  highlights   JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE cv_education (
  id           TEXT PRIMARY KEY,
  institution  JSONB NOT NULL,
  degree       JSONB NOT NULL,
  field        JSONB NOT NULL,
  start_date   DATE  NOT NULL,
  end_date     DATE,
  description  JSONB
);

CREATE TABLE cv_skill_group (
  id          SERIAL PRIMARY KEY,
  category    JSONB NOT NULL,
  items       TEXT[] NOT NULL,
  sort_order  INT NOT NULL
);

CREATE TABLE cv_language (
  id          SERIAL PRIMARY KEY,
  name        JSONB NOT NULL,
  level       JSONB NOT NULL,
  sort_order  INT NOT NULL
);

CREATE TABLE place (
  id            TEXT PRIMARY KEY,
  country_iso2  CHAR(2) NOT NULL,
  name          JSONB NOT NULL,
  lat           DOUBLE PRECISION NOT NULL,
  lng           DOUBLE PRECISION NOT NULL,
  story         JSONB NOT NULL
);

CREATE TABLE visit (
  id          SERIAL PRIMARY KEY,
  place_id    TEXT NOT NULL REFERENCES place(id) ON DELETE CASCADE,
  start_date  DATE NOT NULL,
  end_date    DATE NOT NULL,
  CHECK (end_date >= start_date)
);

CREATE TABLE photo (
  id          SERIAL PRIMARY KEY,
  place_id    TEXT NOT NULL REFERENCES place(id) ON DELETE CASCADE,
  url         TEXT NOT NULL,
  thumb_url   TEXT,
  alt         JSONB,
  caption     JSONB,
  width       INT,
  height      INT,
  sort_order  INT NOT NULL,
  CHECK ((width IS NULL) = (height IS NULL))
);

CREATE INDEX ON visit (place_id);
CREATE INDEX ON photo (place_id);

-- +goose Down
DROP TABLE photo;
DROP TABLE visit;
DROP TABLE place;
DROP TABLE cv_language;
DROP TABLE cv_skill_group;
DROP TABLE cv_education;
DROP TABLE cv_experience;
DROP TABLE social_link;
DROP TABLE main_info;
