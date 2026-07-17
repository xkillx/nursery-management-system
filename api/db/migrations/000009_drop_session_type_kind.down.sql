ALTER TABLE session_types ADD COLUMN kind text NOT NULL DEFAULT 'standard';
ALTER TABLE session_types ADD CONSTRAINT session_types_kind_check CHECK (kind IN ('standard', 'wraparound_before', 'wraparound_after', 'core', 'extended'));
