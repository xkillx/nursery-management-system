CREATE TABLE rooms (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    branch_id UUID NOT NULL REFERENCES branches(id),
    name TEXT NOT NULL,
    description TEXT,
    age_group TEXT NOT NULL,
    capacity INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_rooms_tenant_id ON rooms (tenant_id);
CREATE INDEX idx_rooms_branch_id ON rooms (branch_id);
CREATE INDEX idx_rooms_active ON rooms (branch_id, is_active);
CREATE UNIQUE INDEX idx_rooms_active_name_per_branch ON rooms (branch_id, name) WHERE is_active = true;
