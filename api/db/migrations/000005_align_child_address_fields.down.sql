-- Re-add old columns.
ALTER TABLE child_profiles ADD COLUMN home_address jsonb DEFAULT '{}'::jsonb NOT NULL;
ALTER TABLE child_profiles ADD COLUMN home_postcode text;

-- Restore data: pack address_line1 back into JSONB, copy postcode.
UPDATE child_profiles
SET home_address = jsonb_build_object('text', COALESCE(address_line1, '')),
    home_postcode = address_postcode;

-- Re-add the CHECK constraint.
ALTER TABLE child_profiles ADD CONSTRAINT child_profiles_address_is_object CHECK ((jsonb_typeof(home_address) = 'object'::text));

-- Drop the structured address columns.
ALTER TABLE child_profiles DROP COLUMN IF EXISTS address_line1;
ALTER TABLE child_profiles DROP COLUMN IF EXISTS address_line2;
ALTER TABLE child_profiles DROP COLUMN IF EXISTS address_city;
ALTER TABLE child_profiles DROP COLUMN IF EXISTS address_postcode;
