ALTER TABLE child_registration_consent_records
ADD COLUMN information_sharing_consent BOOLEAN NOT NULL DEFAULT true;
