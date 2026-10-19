-- +goose Up
-- A job can have several descriptions, each with its own highlights:
-- projects is a JSON array of {"description": ltext | null, "highlights": [ltext, ...]}.
ALTER TABLE cv_experience ADD COLUMN projects JSONB NOT NULL DEFAULT '[]';

UPDATE cv_experience
SET projects = jsonb_build_array(jsonb_build_object('description', description, 'highlights', highlights))
WHERE description IS NOT NULL OR highlights <> '[]';

ALTER TABLE cv_experience DROP COLUMN description, DROP COLUMN highlights;

-- +goose Down
ALTER TABLE cv_experience
  ADD COLUMN description JSONB,
  ADD COLUMN highlights  JSONB NOT NULL DEFAULT '[]';

-- Only the first project survives a rollback.
UPDATE cv_experience
SET description = NULLIF(projects -> 0 -> 'description', 'null'::jsonb),
    highlights  = COALESCE(projects -> 0 -> 'highlights', '[]')
WHERE projects <> '[]';

ALTER TABLE cv_experience DROP COLUMN projects;
