-- Day 8: Attendance schema

CREATE TABLE attendance_sessions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    status TEXT NOT NULL,
    check_in_at TIMESTAMPTZ NOT NULL,
    check_out_at TIMESTAMPTZ,
    check_in_local_date DATE NOT NULL,
    check_out_local_date DATE,
    check_in_event_id UUID,
    check_out_event_id UUID,
    corrected_by_event_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT attendance_sessions_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT attendance_sessions_child_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT attendance_sessions_status_check
        CHECK (status IN ('open', 'complete', 'corrected')),
    CONSTRAINT attendance_sessions_open_shape_check
        CHECK (
            (status = 'open' AND check_out_at IS NULL AND check_out_local_date IS NULL)
            OR
            (status IN ('complete', 'corrected') AND check_out_at IS NOT NULL AND check_out_local_date IS NOT NULL)
        ),
    CONSTRAINT attendance_sessions_time_order_check
        CHECK (check_out_at IS NULL OR check_out_at > check_in_at),
    CONSTRAINT attendance_sessions_scope_id_unique
        UNIQUE (tenant_id, branch_id, id)
);

CREATE TABLE attendance_events (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    session_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    local_date DATE NOT NULL,
    recorded_by_user_id UUID NOT NULL REFERENCES users(id),
    recorded_by_membership_id UUID NOT NULL,
    request_id TEXT,
    reason_code TEXT,
    reason_note TEXT,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT attendance_events_branch_scope_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT attendance_events_child_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT attendance_events_session_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, session_id) REFERENCES attendance_sessions(tenant_id, branch_id, id),
    CONSTRAINT attendance_events_membership_scope_fkey
        FOREIGN KEY (tenant_id, branch_id, recorded_by_membership_id) REFERENCES memberships(tenant_id, branch_id, id),
    CONSTRAINT attendance_events_type_check
        CHECK (event_type IN ('check_in', 'check_out', 'correction')),
    CONSTRAINT attendance_events_correction_reason_check
        CHECK (event_type <> 'correction' OR reason_code IN ('missed_check_in', 'missed_check_out', 'incorrect_time', 'duplicate_entry', 'other')),
    CONSTRAINT attendance_events_other_reason_note_check
        CHECK (reason_code <> 'other' OR NULLIF(reason_note, '') IS NOT NULL)
);

-- Unique index needed before deferred FK references
CREATE UNIQUE INDEX attendance_events_scope_id_unique
    ON attendance_events (tenant_id, branch_id, id);

-- Deferred relationship constraints after both tables exist
ALTER TABLE attendance_sessions
ADD CONSTRAINT attendance_sessions_check_in_event_fkey
    FOREIGN KEY (tenant_id, branch_id, check_in_event_id) REFERENCES attendance_events(tenant_id, branch_id, id),
ADD CONSTRAINT attendance_sessions_check_out_event_fkey
    FOREIGN KEY (tenant_id, branch_id, check_out_event_id) REFERENCES attendance_events(tenant_id, branch_id, id),
ADD CONSTRAINT attendance_sessions_corrected_by_event_fkey
    FOREIGN KEY (tenant_id, branch_id, corrected_by_event_id) REFERENCES attendance_events(tenant_id, branch_id, id);

-- Indexes
CREATE UNIQUE INDEX idx_attendance_sessions_one_open_child
    ON attendance_sessions (tenant_id, branch_id, child_id)
    WHERE status = 'open';

CREATE INDEX idx_attendance_sessions_child_date
    ON attendance_sessions (tenant_id, branch_id, child_id, check_in_local_date DESC);

CREATE INDEX idx_attendance_sessions_open_scope
    ON attendance_sessions (tenant_id, branch_id, status, check_in_at)
    WHERE status = 'open';

CREATE INDEX idx_attendance_events_session_time
    ON attendance_events (tenant_id, branch_id, session_id, occurred_at);

CREATE INDEX idx_attendance_events_child_date
    ON attendance_events (tenant_id, branch_id, child_id, local_date DESC);
