-- Reverse dependency order
DROP TRIGGER IF EXISTS trg_invoice_lines_immutability ON invoice_lines;
DROP FUNCTION IF EXISTS protect_issued_invoice_lines();

DROP TRIGGER IF EXISTS trg_invoice_immutability ON invoices;
DROP FUNCTION IF EXISTS protect_issued_invoice_immutability();

DROP TRIGGER IF EXISTS trg_invoice_status_transition ON invoices;
DROP FUNCTION IF EXISTS enforce_invoice_status_transition();

DROP TABLE IF EXISTS invoice_number_sequences;
DROP TABLE IF EXISTS invoice_lines;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS invoice_runs;
