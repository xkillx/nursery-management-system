ALTER TABLE child_registration_profiles
  ALTER COLUMN other_languages DROP DEFAULT,
  ALTER COLUMN other_languages DROP NOT NULL,
  ALTER COLUMN other_languages TYPE TEXT
    USING NULLIF(array_to_string(other_languages[1:1], ''), '');
