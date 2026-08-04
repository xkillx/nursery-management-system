ALTER TABLE payment_attempts ALTER COLUMN initiated_by_user_id DROP NOT NULL;
ALTER TABLE payment_attempts ALTER COLUMN initiated_by_membership_id DROP NOT NULL;
