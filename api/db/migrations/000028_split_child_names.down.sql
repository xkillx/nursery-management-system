ALTER TABLE children
  ADD COLUMN full_name TEXT;

UPDATE children
SET full_name = concat_ws(' ', first_name, middle_name, last_name);

ALTER TABLE children
  ALTER COLUMN full_name SET NOT NULL;

ALTER TABLE children
  DROP CONSTRAINT children_first_name_not_blank_check,
  DROP COLUMN first_name,
  DROP COLUMN middle_name,
  DROP COLUMN last_name;
