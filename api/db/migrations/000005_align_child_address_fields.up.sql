-- Add structured address columns to child_profiles (matching parent schema).
ALTER TABLE child_profiles ADD COLUMN address_line1 text;
ALTER TABLE child_profiles ADD COLUMN address_line2 text;
ALTER TABLE child_profiles ADD COLUMN address_city text;
ALTER TABLE child_profiles ADD COLUMN address_postcode text;

-- Migrate existing data: extract text from JSONB, copy postcode.
UPDATE child_profiles
SET address_line1 = home_address->>'text',
    address_postcode = home_postcode;

-- Drop the old JSONB column and its CHECK constraint.
ALTER TABLE child_profiles DROP CONSTRAINT IF EXISTS child_profiles_address_is_object;
ALTER TABLE child_profiles DROP COLUMN IF EXISTS home_address;

-- Drop the old postcode column.
ALTER TABLE child_profiles DROP COLUMN IF EXISTS home_postcode;
