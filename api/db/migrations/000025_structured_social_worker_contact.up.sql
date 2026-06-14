ALTER TABLE child_registration_profiles
    ADD COLUMN social_worker_name TEXT,
    ADD COLUMN social_worker_phone TEXT,
    ADD COLUMN social_worker_email TEXT;

ALTER TABLE child_registration_profiles
    DROP COLUMN social_worker_contact_details;
