ALTER TABLE child_registration_profiles
  ALTER COLUMN other_languages TYPE TEXT[]
    USING CASE
      WHEN other_languages IS NULL OR btrim(other_languages) = '' THEN ARRAY[]::TEXT[]
      ELSE ARRAY[other_languages]
    END,
  ALTER COLUMN other_languages SET DEFAULT '{}',
  ALTER COLUMN other_languages SET NOT NULL;
