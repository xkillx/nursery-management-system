ALTER TABLE ONLY children DROP CONSTRAINT IF EXISTS children_current_term_fkey;
ALTER TABLE children DROP COLUMN IF EXISTS current_term_id;

DROP TABLE IF EXISTS invoice_run_advance;
DROP TABLE IF EXISTS term_schedule_change;
DROP TABLE IF EXISTS term;
