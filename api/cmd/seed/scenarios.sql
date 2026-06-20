-- Day 4 core schema seed scenarios

-- Scenario: three roles in one scope
INSERT INTO tenants (id, name)
VALUES ('11111111-1111-1111-1111-111111111111', 'Seed Tenant')
ON CONFLICT (id) DO NOTHING;

INSERT INTO branches (id, tenant_id, name)
VALUES ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Seed Branch')
ON CONFLICT (tenant_id, name) DO UPDATE SET updated_at = now();

UPDATE branches SET core_hourly_rate_minor = 1200 WHERE id = '22222222-2222-2222-2222-222222222222';

INSERT INTO users (id, email, email_normalized, password_hash, is_active)
VALUES
  ('33333333-3333-3333-3333-333333333333', 'manager@seed.local', 'manager@seed.local', 'seed', true),
  ('44444444-4444-4444-4444-444444444444', 'practitioner@seed.local', 'practitioner@seed.local', 'seed', true),
  ('55555555-5555-5555-5555-555555555555', 'parent@seed.local', 'parent@seed.local', 'seed', true)
ON CONFLICT (email_normalized) DO NOTHING;

INSERT INTO memberships (id, tenant_id, branch_id, user_id, role, is_active, ended_at)
VALUES
  ('66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333', 'manager', true, NULL),
  ('77777777-7777-7777-7777-777777777777', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', '44444444-4444-4444-4444-444444444444', 'practitioner', true, NULL),
  ('88888888-8888-8888-8888-888888888888', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', '55555555-5555-5555-5555-555555555555', 'parent', true, NULL)
ON CONFLICT (tenant_id, branch_id, user_id) DO UPDATE SET role = EXCLUDED.role, is_active = true, ended_at = NULL, updated_at = now();

-- Scenario: child with a parent_carer contact and an active parent-membership mapping
INSERT INTO children (id, tenant_id, branch_id, first_name, date_of_birth, start_date, core_hourly_rate_minor, is_active)
VALUES ('99999999-9999-9999-9999-999999999999', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'Seed Child', '2021-01-01', '2024-01-01', 1200, true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO child_contacts (id, tenant_id, branch_id, child_id, contact_type, sort_order, full_name, email, telephone, has_parental_responsibility)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', '99999999-9999-9999-9999-999999999999', 'parent_carer', 1, 'Seed Parent Carer', 'parent@seed.local', '07123456789', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO parent_membership_children (id, tenant_id, branch_id, membership_id, child_id)
VALUES ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', '88888888-8888-8888-8888-888888888888', '99999999-9999-9999-9999-999999999999')
ON CONFLICT DO NOTHING;
