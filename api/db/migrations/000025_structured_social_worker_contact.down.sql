ALTER TABLE child_registration_profiles
    ADD COLUMN social_worker_contact_details TEXT;

ALTER TABLE child_registration_profiles
    DROP COLUMN social_worker_name,
    DROP COLUMN social_worker_phone,
    DROP COLUMN social_worker_email;
