-- Reverse baseline: drop all schema objects created by 000001_baseline.up.sql.
-- CASCADE so trigger bindings, FK dependents, and indexes fall with the tables.

DROP TABLE IF EXISTS payment_reconciliation_records CASCADE;
DROP TABLE IF EXISTS payment_attempts CASCADE;
DROP TABLE IF EXISTS invoice_lines CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS invoice_runs CASCADE;
DROP TABLE IF EXISTS invoice_number_sequences CASCADE;
DROP TABLE IF EXISTS invoice_run_advance CASCADE;
DROP TABLE IF EXISTS term_schedule_change CASCADE;
DROP TABLE IF EXISTS term CASCADE;
DROP TABLE IF EXISTS session_template_entries CASCADE;
DROP TABLE IF EXISTS session_templates CASCADE;
DROP TABLE IF EXISTS child_booking_pattern_entries CASCADE;
DROP TABLE IF EXISTS child_booking_patterns CASCADE;
DROP TABLE IF EXISTS session_types CASCADE;
DROP TABLE IF EXISTS parent_membership_children CASCADE;
DROP TABLE IF EXISTS manager_invites CASCADE;
DROP TABLE IF EXISTS stripe_webhook_events CASCADE;
DROP TABLE IF EXISTS password_reset_tokens CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS child_leaving_records CASCADE;
DROP TABLE IF EXISTS child_billing_profiles CASCADE;
DROP TABLE IF EXISTS child_room_assignments CASCADE;
DROP TABLE IF EXISTS rooms CASCADE;
DROP TABLE IF EXISTS child_collection_settings CASCADE;
DROP TABLE IF EXISTS child_consent_records CASCADE;
DROP TABLE IF EXISTS child_funding_records CASCADE;
DROP TABLE IF EXISTS child_safeguarding_profiles CASCADE;
DROP TABLE IF EXISTS child_health_profiles CASCADE;
DROP TABLE IF EXISTS child_contacts CASCADE;
DROP TABLE IF EXISTS child_profiles CASCADE;
DROP TABLE IF EXISTS absence_markers CASCADE;
DROP TABLE IF EXISTS funding_profiles CASCADE;
DROP TABLE IF EXISTS attendance_events CASCADE;
DROP TABLE IF EXISTS attendance_sessions CASCADE;
DROP TABLE IF EXISTS children CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS memberships CASCADE;
DROP TABLE IF EXISTS rooms CASCADE;
DROP TABLE IF EXISTS branches CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;

DROP FUNCTION IF EXISTS enforce_invoice_status_transition() CASCADE;
DROP FUNCTION IF EXISTS protect_issued_invoice_immutability() CASCADE;
DROP FUNCTION IF EXISTS protect_issued_invoice_lines() CASCADE;
DROP FUNCTION IF EXISTS enforce_parent_membership_child_role() CASCADE;
DROP FUNCTION IF EXISTS enforce_parent_membership_child_scope() CASCADE;
DROP FUNCTION IF EXISTS cascade_parent_membership_child_end() CASCADE;

DROP TYPE IF EXISTS child_contact_type CASCADE;
DROP TYPE IF EXISTS lifecycle_reason_code CASCADE;
DROP TYPE IF EXISTS registration_term_time_only_status CASCADE;
