-- Delete an invoice and all related records.
-- Usage: psql -h localhost -p 5432 -U mac -d nursery_management -v invoice_id="'019fb536-c371-7dc7-9c80-5dff68e9e216'" -f api/db/scripts/delete_invoice.sql

BEGIN;

DELETE FROM payment_attempts WHERE invoice_id = :invoice_id;
DELETE FROM payment_reconciliation_records WHERE invoice_id = :invoice_id;
DELETE FROM invoice_reminder_log WHERE invoice_id = :invoice_id;
DELETE FROM invoices WHERE id = :invoice_id;

COMMIT;
