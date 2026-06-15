ALTER TABLE children
  ADD COLUMN first_name TEXT,
  ADD COLUMN middle_name TEXT,
  ADD COLUMN last_name TEXT;

UPDATE children
SET first_name = full_name
WHERE first_name IS NULL;

ALTER TABLE children
  ALTER COLUMN first_name SET NOT NULL,
  ADD CONSTRAINT children_first_name_not_blank_check CHECK (btrim(first_name) <> '');

ALTER TABLE children
  DROP COLUMN full_name;
