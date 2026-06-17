# Plan: Child Management Refactor — Drop the Registration Model, Adopt Direct Child Creation

Status: ready for execution
Audience: another AI agent that will execute it end-to-end without further human confirmation.

## 0. Product decisions captured in the interview

These are the resolved product decisions. The plan must implement these, not re-litigate them.

- **Scope**: big-bang rename, single release. No strangler pattern, no dual-write window.
- **Create endpoint**: there is exactly one manager endpoint that creates a child: `POST /api/v1/children`. It runs the full `CreateChildWithFullProfile` use case in one DB transaction. The legacy `POST /api/v1/children/with-registration` route is removed.
- **Module shape**: the `registrationprofiles` Go module is deleted. Its domain is merged into the `children` module. The `children` module becomes the single owner of the child bounded context (identity, profile, contacts, health, safeguarding, consent, funding, collection, room, leaving, billing).
- **Manager UI**: keep the guided stepper, retitle it to "Add child" / "Edit child". The stepper is the only edit surface for the new sub-records. The per-section editor that lived on the child detail page is deleted. The child detail page becomes a read-only summary with a single "Edit" button that opens the stepper pre-filled.
- **`children` table cleanup**: drop `primary_room_id`, `core_hourly_rate_minor`, `left_at`, `left_reason_code`, `left_reason_note`. Move their meaning to `child_room_assignments`, `child_billing_profiles`, `child_leaving_records`.
- **Funding**: the new `child_funding_records` is a per-child eligibility record (15/30h, 2yo, tax-free childcare, benefits, support notes), one row per child. It is distinct from the existing `funding_profiles` (per-billing-month funded allowance minutes used by invoicing). Both tables coexist.
- **Consent history**: collapse the current versioned `child_registration_consent_records` to a single current row per child in `child_consent_records`. No version column. New consent POST updates the row in place.
- **Completion attestation**: drop the `child_registration_completion_attestations` table and the `POST /api/v1/children/:id/registration-completion-attestations` endpoint. There is no separate "mark complete" action. The create transaction is the implicit attestation.
- **Collection password**: move off `child_profiles` into a new `child_collection_settings` table (one-to-one with child). Endpoint becomes `PUT /api/v1/children/:id/collection-settings`.
- **Health & safeguarding**: split into two 1:1 tables — `child_health_profiles` (medical/allergy/dietary/GP/health visitor) and `child_safeguarding_profiles` (social services, six concern flags, professional referrals, restricted notes). `child_profiles` keeps demographics/disability/access/routine care/GDPR declaration/registration date.
- **Authorisation**: manager only on the create and edit endpoints. Practitioner and parent access does not change in this tranche.
- **Create transaction rules**: the only fields required at create are `first_name`, `date_of_birth`, `start_date`, and `consents.safeguarding_reporting_acknowledgement = true`. `primary_room_id` is required (room placement is operational). Everything else is optional. The stepper in the UI keeps its "answer required to proceed" rule, but the API does not enforce that.
- **Migration policy**: drop the four registration tables (`child_registration_profiles`, `child_registration_contacts`, `child_registration_consent_records`, `child_registration_completion_attestations`) and rebuild the new shape. No data preservation. Backfill `child_room_assignments`, `child_billing_profiles`, `child_leaving_records`, `child_collection_settings` from existing `children` rows for the columns being dropped (one row per currently-active child that has data in those columns).
- **Audit log**: reuse the existing platform `audit_logs` table. The `CreateChildWithFullProfile` use case writes one `child_created` row with a `details` JSON that lists which sub-records were created.

## 1. Implementation decisions the agent owns (do not re-ask)

- **API path layout**: child sub-resource routes go under `/api/v1/children/:child_id/...` with the resource names from the new model (`profile`, `contacts`, `health`, `safeguarding`, `consent`, `funding`, `collection-settings`, `room-assignments`, `leaving`). Manager role on all of them.
- **HTTP method choice**: GET + PATCH for sub-resources that are 1:1 and mutable. POST/DELETE for `room-assignments` (1:many). GET + PATCH for `consent`. PUT for `collection-settings` (single hash).
- **Go package layout inside the `children` module**:
  - `internal/modules/children/domain/` — entity files split per concept: `child.go` (identity), `child_profile.go`, `child_contact.go`, `child_health_profile.go`, `child_safeguarding_profile.go`, `child_consent.go`, `child_funding_record.go`, `child_collection_setting.go`, `child_room_assignment.go`, `child_billing_profile.go`, `child_leaving_record.go`, `repository.go` (composite), `errors.go`.
  - `internal/modules/children/application/` — one use case per file: `create_child_with_full_profile.go`, `get_child.go`, `update_child.go`, `list_children.go`, `mark_inactive.go`, `list_attendance.go`, `get_profile.go`, `update_profile.go`, `get_contacts.go`, `replace_contacts.go`, `get_health.go`, `update_health.go`, `get_safeguarding.go`, `update_safeguarding.go`, `get_consent.go`, `update_consent.go`, `get_funding.go`, `update_funding.go`, `get_collection_setting.go`, `set_collection_password.go`, `list_room_assignments.go`, `create_room_assignment.go`, `close_room_assignment.go`, `get_billing_profile.go`, `update_billing_profile.go`, `get_leaving_record.go`, `record_leaving.go`.
  - `internal/modules/children/infrastructure/postgres/` — one repository implementation per concept, sharing one `*pgxpool.Pool` and one `sqlc.Querier`; or one repository file with grouped methods (agent's choice, keep it local and small).
  - `internal/modules/children/interfaces/http/` — one handler per resource group, with one DTO file per resource. The current single-file `handler.go` and `dto.go` split.
- **sqlc**: rewrite `api/db/query/children.sql` to a clean set of `:one`/`:many`/`:exec`/`:execrows` queries for the new shape. Keep one file per concept (`children_identity.sql`, `child_profiles.sql`, `child_contacts.sql`, `child_health_profiles.sql`, `child_safeguarding_profiles.sql`, `child_consents.sql`, `child_funding_records.sql`, `child_collection_settings.sql`, `child_room_assignments.sql`, `child_billing_profiles.sql`, `child_leaving_records.sql`) and `go tool sqlc generate`.
- **Transaction boundary**: every write use case that touches more than one table runs inside `txMgr.ExecTx(ctx, func(tx pgx.Tx) error{...})`. Read use cases use the pool directly.
- **Audit fields**: the existing `audit.WriteParams.ActionType`/`EntityType`/`EntityID`/`Details` shape stays. New action types introduced by this refactor: `child_created`, `child_profile_updated`, `child_contacts_replaced`, `child_health_updated`, `child_safeguarding_updated`, `child_consent_updated`, `child_funding_updated`, `child_collection_password_set`, `child_room_assigned`, `child_room_unassigned`, `child_marked_inactive`. No sensitive values (medical notes, safeguarding notes, contact details, collection password) in `Details` (existing redaction policy).
- **Validation**: keep the existing pattern of `domainerrors.Validation` and `domainerrors.ValidationWithFields`. The new `CreateChildWithFullProfile` use case validates required fields up front and then writes inside a single transaction.
- **Naming inside code** (old → new):
  - `SubmitCompleteRegistration` → `CreateChildWithFullProfile`
  - `CompleteRegistrationInput` → `CreateChildFullInput`
  - `ChildRegistrationInfo` → `ChildIdentityInput`
  - `RegistrationProfile` → `ChildProfile`
  - `ContactEntry` → `ChildContact`
  - `ConsentRecord` → `ChildConsent`
  - `CompletionAttestation` → removed
  - `Profile` repository → `ChildProfileRepository`, etc.
- **Frontend naming**:
  - Routes: `manager-registration-intake` → `manager-child-edit` (component also renamed). The "add" mode is when no `child_id` is in the URL; the "edit" mode is `/children/:child_id/edit`.
  - Models: `registration-profile.models.ts` → `child-profile.models.ts`, etc. Domain types (`RegistrationProfile`, `ConsentRecord`, `ContactEntry`, `CompleteRegistrationPayload`, `CompleteRegistrationResponse`) get the new names. The old `manager-child-registration` page is deleted (was the section editor on the child detail page).
  - Service: `getRegistrationProfile` → `getChildProfile`, `patchRegistrationProfile` → `patchChildProfile`, `setRegistrationCollectionPassword` → `setChildCollectionPassword`, `getRegistrationConsents` → `getChildConsent`, `createRegistrationConsent` → `updateChildConsent`, `getRegistrationWorkflowStatus` → dropped (no more workflow status), `createRegistrationCompletionAttestation` → dropped, `submitCompleteRegistration` → `createChildWithFullProfile`.
- **Test names**: `TestSubmitCompleteRegistrationValidateInput_*` → `TestCreateChildWithFullProfileValidateInput_*`. Update imports from `registrationprofiles/...` to `children/...`.

## 2. Out of scope for this plan

- Practitioner and parent access to any of the new sub-resources. Owner read-only overview is unchanged.
- Time-versioned consent history (collapsed in §0).
- A separate "pending application" or server-side draft object. The stepper keeps its localStorage draft pattern.
- Per-child custom billing rate wiring. `child_billing_profiles` exists but every row is `billing_basis = 'site_rate'` and `custom_rate_minor IS NULL`. The endpoint exists to switch the basis in the future.
- Digital parent journey, paper evidence uploads, Ofsted reporting.
- ADR. None of the §1 implementation choices pass the ADR test (none are hard to reverse, none are surprising without context, none are the result of a real trade-off). Skip.

## 3. File-by-file change list

### 3.1 Migrations

Add one forward-and-back pair, in numeric sequence after `000033_drop_consent_paper_form_on_file`:

`api/db/migrations/000034_refactor_to_child_management_model.up.sql`

```sql
-- 1. Drop the registration tables. No data preservation.
DROP TABLE IF EXISTS child_registration_contacts CASCADE;
DROP TABLE IF EXISTS child_registration_consent_records CASCADE;
DROP TABLE IF EXISTS child_registration_profiles CASCADE;
DROP TABLE IF EXISTS child_registration_completion_attestations CASCADE;
DROP TYPE  IF EXISTS registration_yes_no_unknown CASCADE;
DROP TYPE  IF EXISTS registration_immunisation_status CASCADE;
DROP TYPE  IF EXISTS registration_contact_type CASCADE;

-- 2. Drop the operational columns we are moving off `children`.
ALTER TABLE children
    DROP COLUMN IF EXISTS primary_room_id,
    DROP COLUMN IF EXISTS core_hourly_rate_minor,
    DROP COLUMN IF EXISTS left_at,
    DROP COLUMN IF EXISTS left_reason_code,
    DROP COLUMN IF EXISTS left_reason_note;

-- 3. New child_profiles table (1:1 with children).
CREATE TABLE child_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    -- Demographics / home
    sex TEXT,
    religion TEXT,
    ethnic_origin TEXT,
    first_language TEXT,
    other_languages TEXT,
    home_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    home_postcode TEXT,
    home_telephone TEXT,

    -- Disability / access
    disability_status TEXT NOT NULL DEFAULT 'unknown',  -- 'unknown' | 'no' | 'yes'
    disability_notes TEXT,
    access_requirements TEXT,

    -- Routine care
    routine_care_notes TEXT,

    -- GDPR declaration (paper form metadata)
    gdpr_declared_by_name TEXT,
    gdpr_declared_at TIMESTAMPTZ,
    gdpr_declaration_date DATE,

    -- Registration date captured at intake (paper form "today" the family registered)
    registration_date DATE,

    -- Section review booleans
    demographics_home_reviewed BOOLEAN NOT NULL DEFAULT false,
    medical_dietary_reviewed  BOOLEAN NOT NULL DEFAULT false,
    health_contacts_reviewed  BOOLEAN NOT NULL DEFAULT false,
    social_development_reviewed BOOLEAN NOT NULL DEFAULT false,
    parent_responsibility_reviewed BOOLEAN NOT NULL DEFAULT false,
    emergency_collection_reviewed BOOLEAN NOT NULL DEFAULT false,
    routine_care_reviewed BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_profiles_address_is_object
        CHECK (jsonb_typeof(home_address) = 'object')
);

CREATE INDEX idx_child_profiles_child ON child_profiles (tenant_id, branch_id, child_id);

-- 4. child_contacts: parent/carer, emergency, authorised collector. 1:many.
CREATE TYPE child_contact_type AS ENUM ('parent_carer', 'emergency_contact', 'authorised_collector');

CREATE TABLE child_contacts (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    contact_type child_contact_type NOT NULL,
    sort_order INTEGER NOT NULL,
    full_name TEXT NOT NULL CHECK (btrim(full_name) <> ''),
    relationship_to_child TEXT,
    address JSONB NOT NULL DEFAULT '{}'::jsonb,
    telephone TEXT,
    email TEXT,
    work_address JSONB NOT NULL DEFAULT '{}'::jsonb,
    has_parental_responsibility BOOLEAN,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_contacts_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_contacts_address_is_object
        CHECK (jsonb_typeof(address) = 'object'),
    CONSTRAINT child_contacts_work_address_is_object
        CHECK (jsonb_typeof(work_address) = 'object'),
    CONSTRAINT child_contacts_sort_order_nonneg
        CHECK (sort_order >= 0),
    CONSTRAINT child_contacts_type_sort_unique
        UNIQUE (tenant_id, branch_id, child_id, contact_type, sort_order)
);

CREATE INDEX idx_child_contacts_child_type
    ON child_contacts (tenant_id, branch_id, child_id, contact_type, sort_order);

-- 5. child_health_profiles: 1:1 with children. Medical, allergy, dietary, GP, health visitor.
CREATE TABLE child_health_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    medical_conditions_status TEXT NOT NULL DEFAULT 'unknown',  -- unknown | no | yes
    medical_conditions_notes TEXT,
    prescribed_medication_status TEXT NOT NULL DEFAULT 'unknown',
    medication_notes TEXT,
    immunisation_status TEXT NOT NULL DEFAULT 'unknown',  -- unknown | up_to_date | refused | partial | not_recorded
    immunisation_country TEXT,
    illness_diagnosis_history TEXT,
    dietary_requirements_status TEXT NOT NULL DEFAULT 'unknown',
    dietary_requirements_notes TEXT,
    dietary_side_effects TEXT,

    doctor_name TEXT,
    doctor_address TEXT,
    doctor_phone TEXT,
    health_visitor_name TEXT,
    health_visitor_address TEXT,
    health_visitor_phone TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_health_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_health_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id)
);

CREATE INDEX idx_child_health_profiles_child ON child_health_profiles (tenant_id, branch_id, child_id);

-- 6. child_safeguarding_profiles: 1:1 with children. Social services, concerns, professional referrals, restricted notes.
CREATE TABLE child_safeguarding_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    social_services_status TEXT NOT NULL DEFAULT 'unknown',  -- unknown | no | yes
    social_services_notes TEXT,
    social_worker_name TEXT,
    social_worker_phone TEXT,
    social_worker_email TEXT,
    concern_walking TEXT NOT NULL DEFAULT 'unknown',
    concern_speech_language TEXT NOT NULL DEFAULT 'unknown',
    concern_hearing TEXT NOT NULL DEFAULT 'unknown',
    concern_sight TEXT NOT NULL DEFAULT 'unknown',
    concern_emotional_wellbeing TEXT NOT NULL DEFAULT 'unknown',
    concern_behaviour TEXT NOT NULL DEFAULT 'unknown',
    professional_referrals JSONB NOT NULL DEFAULT '[]'::jsonb,
    restricted_notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_safeguarding_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_safeguarding_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_safeguarding_profiles_referrals_is_array
        CHECK (jsonb_typeof(professional_referrals) = 'array')
);

CREATE INDEX idx_child_safeguarding_profiles_child ON child_safeguarding_profiles (tenant_id, branch_id, child_id);

-- 7. child_consent_records: 1:1 with children. Single current row (no version).
CREATE TABLE child_consent_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    urgent_medical_treatment BOOLEAN NOT NULL,
    urgent_medical_treatment_exceptions TEXT,
    plasters BOOLEAN NOT NULL,
    safeguarding_reporting_acknowledgement BOOLEAN NOT NULL,
    information_sharing_consent BOOLEAN NOT NULL,
    gdpr_data_processing_consent BOOLEAN NOT NULL,
    area_senco_liaison BOOLEAN NOT NULL,
    health_visitor_liaison BOOLEAN NOT NULL,
    transition_documents BOOLEAN NOT NULL,
    local_outings BOOLEAN NOT NULL,
    face_painting BOOLEAN NOT NULL,
    parent_supplied_sun_cream BOOLEAN NOT NULL,
    parent_supplied_nappy_cream BOOLEAN NOT NULL,
    development_profile_photos BOOLEAN NOT NULL,
    nursery_display_boards BOOLEAN NOT NULL,
    promotional_literature BOOLEAN NOT NULL,
    nursery_website BOOLEAN NOT NULL,
    staff_student_coursework BOOLEAN NOT NULL,
    social_media BOOLEAN NOT NULL,
    social_media_channel_notes TEXT,

    notes_exceptions TEXT,

    -- Paper-form signer metadata (still useful, not stored in profile)
    signer_name TEXT NOT NULL,
    signed_date DATE NOT NULL,
    paper_form_on_file BOOLEAN NOT NULL,

    -- Last write actor
    entered_by_user_id UUID NOT NULL REFERENCES users(id),
    entered_by_membership_id UUID NOT NULL REFERENCES memberships(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_consent_records_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_consent_records_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id)
);

CREATE INDEX idx_child_consent_records_child ON child_consent_records (tenant_id, branch_id, child_id);

-- 8. child_funding_records: 1:1 with children. Eligibility/intent (15/30h, 2yo, tax-free, benefits).
CREATE TABLE child_funding_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    benefits_contribute_to_fees TEXT NOT NULL DEFAULT 'unknown',  -- unknown | no | yes
    working_tax_credit         TEXT NOT NULL DEFAULT 'unknown',
    college_uni_paid_to_parent TEXT NOT NULL DEFAULT 'unknown',
    college_uni_paid_to_nursery TEXT NOT NULL DEFAULT 'unknown',
    funding_3yo_term_time      TEXT NOT NULL DEFAULT 'unknown',
    funding_2yo_term_time      TEXT NOT NULL DEFAULT 'unknown',
    funding_support_notes      TEXT,
    funding_support_reviewed   BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_funding_records_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_funding_records_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id)
);

CREATE INDEX idx_child_funding_records_child ON child_funding_records (tenant_id, branch_id, child_id);

-- 9. child_collection_settings: 1:1 with children. The collection password hash and metadata.
CREATE TABLE child_collection_settings (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    over_18_collection_acknowledged BOOLEAN NOT NULL DEFAULT false,
    collection_password_hash TEXT,
    collection_password_updated_at TIMESTAMPTZ,
    collection_password_updated_by_user_id UUID REFERENCES users(id),
    collection_password_updated_by_membership_id UUID REFERENCES memberships(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_collection_settings_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_collection_settings_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_collection_settings_password_consistency
        CHECK (
            (collection_password_hash IS NULL
             AND collection_password_updated_at IS NULL
             AND collection_password_updated_by_user_id IS NULL
             AND collection_password_updated_by_membership_id IS NULL)
            OR
            (collection_password_hash IS NOT NULL
             AND collection_password_updated_at IS NOT NULL
             AND collection_password_updated_by_user_id IS NOT NULL
             AND collection_password_updated_by_membership_id IS NOT NULL)
        )
);

CREATE INDEX idx_child_collection_settings_child ON child_collection_settings (tenant_id, branch_id, child_id);

-- 10. child_room_assignments: 1:many per child. History of room placements.
CREATE TABLE child_room_assignments (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms(id),
    start_date DATE NOT NULL,
    end_date DATE,  -- NULL = current
    is_current BOOLEAN NOT NULL GENERATED ALWAYS AS (end_date IS NULL) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_room_assignments_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_room_assignments_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_room_assignments_dates_check
        CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE INDEX idx_child_room_assignments_child_current
    ON child_room_assignments (tenant_id, branch_id, child_id) WHERE is_current;
CREATE INDEX idx_child_room_assignments_room_current
    ON child_room_assignments (tenant_id, branch_id, room_id) WHERE is_current;

-- 11. child_billing_profiles: 1:1 with children. Per-child billing basis.
CREATE TABLE child_billing_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    billing_basis TEXT NOT NULL DEFAULT 'site_rate',  -- 'site_rate' | 'custom'
    custom_rate_minor INTEGER,                        -- nullable, only when billing_basis='custom'
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_billing_profiles_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_billing_profiles_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_billing_profiles_basis_check
        CHECK (billing_basis IN ('site_rate', 'custom')),
    CONSTRAINT child_billing_profiles_custom_rate_consistency
        CHECK (
            (billing_basis = 'site_rate' AND custom_rate_minor IS NULL)
            OR
            (billing_basis = 'custom' AND custom_rate_minor IS NOT NULL AND custom_rate_minor > 0)
        )
);

CREATE INDEX idx_child_billing_profiles_child ON child_billing_profiles (tenant_id, branch_id, child_id);

-- 12. child_leaving_records: 1:1 with children. Set when child is marked inactive.
CREATE TABLE child_leaving_records (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    branch_id UUID NOT NULL,
    child_id UUID NOT NULL UNIQUE,

    left_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reason_code TEXT NOT NULL,  -- duplicate_record | entered_in_error | left_nursery | safeguarding_direction | contact_update | access_revoked | other
    reason_note TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT child_leaving_records_branch_fkey
        FOREIGN KEY (tenant_id, branch_id) REFERENCES branches(tenant_id, id),
    CONSTRAINT child_leaving_records_child_fkey
        FOREIGN KEY (tenant_id, branch_id, child_id) REFERENCES children(tenant_id, branch_id, id),
    CONSTRAINT child_leaving_records_reason_check
        CHECK (reason_code IN ('duplicate_record','entered_in_error','left_nursery','safeguarding_direction','contact_update','access_revoked','other'))
);

CREATE INDEX idx_child_leaving_records_child ON child_leaving_records (tenant_id, branch_id, child_id);
```

`api/db/migrations/000034_refactor_to_child_management_model.down.sql` must reverse this entire migration: drop the new tables in reverse order, re-add the dropped `children` columns (`left_reason_code` as enum + `left_reason_note` text + `left_at` timestamptz + `core_hourly_rate_minor` int + `primary_room_id` uuid), re-create the four `child_registration_*` tables and three enum types from `000016/000020/000021/000023/000024/000025/000026`.

### 3.2 sqlc query files

Replace `api/db/query/children.sql` with the new split (keep the filename for the identity table; add per-concept files):

- `api/db/query/children.sql` — keep `ChildrenList`, `ChildrenGetByID`, `ChildrenGetByIDForUpdate`, `ChildrenExistsInScope`, `ChildrenListAttendance`, `ChildrenGetForCorrection`. Rewrite to drop the columns we removed (`primary_room_id`, `core_hourly_rate_minor`, `left_at`, `left_reason_code`, `left_reason_note`). Replace `site_core_hourly_rate_minor` lookup with a JOIN to `branches.core_hourly_rate_minor`. Drop `left_at` from list. Add `has_current_room` via EXISTS on `child_room_assignments`. Update `ChildrenUpdate` to drop primary_room_id, core_hourly_rate_minor, notes paths that are gone. Update `ChildrenMarkInactive` to be a no-op (we just flip `is_active` in the use case; the leaving record is written via a separate `ChildLeavingRecordsInsert` query). The new `ChildrenCreate` should drop the old primary_room_id, core_hourly_rate_minor parameters and add `is_active true` only.
- `api/db/query/child_profiles.sql` — `ChildProfileGetByChild`, `ChildProfileGetForUpdate`, `ChildProfileInsert`, `ChildProfileUpdate`, `ChildProfileDelete` (not used; left out).
- `api/db/query/child_contacts.sql` — `ChildContactsListByChild`, `ChildContactsListByChildAndType`, `ChildContactsReplaceForTypes` (delete by child_id + type list, then batch insert).
- `api/db/query/child_health_profiles.sql` — `ChildHealthProfileGetByChild`, `ChildHealthProfileUpsert`.
- `api/db/query/child_safeguarding_profiles.sql` — `ChildSafeguardingProfileGetByChild`, `ChildSafeguardingProfileUpsert`.
- `api/db/query/child_consents.sql` — `ChildConsentGetByChild`, `ChildConsentInsert`, `ChildConsentUpdate` (single row per child).
- `api/db/query/child_funding_records.sql` — `ChildFundingRecordGetByChild`, `ChildFundingRecordUpsert`.
- `api/db/query/child_collection_settings.sql` — `ChildCollectionSettingGetByChild`, `ChildCollectionSettingUpsert`, `ChildCollectionSettingSetPassword`.
- `api/db/query/child_room_assignments.sql` — `ChildRoomAssignmentsListByChild`, `ChildRoomAssignmentsGetCurrentByChild`, `ChildRoomAssignmentsInsert`, `ChildRoomAssignmentsCloseCurrent` (set end_date = today for the current row in the same transaction), `ChildRoomAssignmentsGetCurrentByRoom` (for capacity checks).
- `api/db/query/child_billing_profiles.sql` — `ChildBillingProfileGetByChild`, `ChildBillingProfileUpsert`.
- `api/db/query/child_leaving_records.sql` — `ChildLeavingRecordGetByChild`, `ChildLeavingRecordInsert`.

Delete `api/db/query/registration_profiles.sql`, `api/db/query/registration_consents.sql`, `api/db/query/registration_completion.sql` (no longer used; the migration is the only thing that references the old tables).

Run `make sqlc-generate` (which runs `go tool sqlc generate`). Update `api/internal/platform/db/sqlc/` accordingly.

### 3.3 Go code

#### Delete

- `api/internal/modules/registrationprofiles/` (entire module).
- `api/cmd/seed/...` references to `registrationprofiles` if any (verify with `rg "registrationprofiles" api/cmd`).
- `api/internal/app/bootstrap/people_routes_test.go` and any other test that imports `registrationprofiles` (verify with `rg "registrationprofiles" api`).

#### Move into the children module

Each of these use cases (the file names from §1) gets created under `api/internal/modules/children/application/`. The corresponding domain entity file goes under `api/internal/modules/children/domain/`. The corresponding repo methods go under `api/internal/modules/children/infrastructure/postgres/repository.go` (or split files; agent's choice).

Specifically, these existing files are deleted and replaced:

- `application/submit_complete_registration.go` → `application/create_child_with_full_profile.go`. New signature:
  - Input: `CreateChildFullInput { Child ChildIdentityInput; Profile *ChildProfileInput; Health *ChildHealthProfileInput; Safeguarding *ChildSafeguardingProfileInput; Contacts []ChildContactInput (pre-bucketed by type); Consent *ChildConsentInput; Funding *ChildFundingRecordInput; CollectionSettings *ChildCollectionSettingsInput (password only); Room *ChildRoomAssignmentInput; }`.
  - Validation: first_name, date_of_birth, start_date, room_id, and `consent.safeguarding_reporting_acknowledgement = true`. Everything else optional.
  - Transaction: insert `children` (identity + is_active=true), insert `child_collection_settings` (empty), insert `child_room_assignments` (current row), insert `child_profile` if Profile present, insert `child_health_profile` if Health present, insert `child_safeguarding_profile` if Safeguarding present, replace `child_contacts` for provided types, insert `child_consent_records` (single row, must be present), insert `child_funding_records` if Funding present, insert `child_billing_profiles` (defaults: basis='site_rate', custom_rate_null), audit log `child_created`.
  - Result: `ChildCreationResult { ChildID, FirstName, MiddleName, LastName, StartDate, CreatedSubRecords []string }`.
- `application/submit_complete_registration_test.go` → renamed to `application/create_child_with_full_profile_test.go`. The existing three test cases stay, retitled, with new field expectations (no `registration_date` validation, no `paper_form_on_file` validation; collection password optional; room is a hard requirement; safeguarding ack is a hard requirement).

#### Replace the children HTTP handler

- `interfaces/http/handler.go` — keep the existing `listChildren`, `getChild`, `markInactive`, `listAttendance` handlers. Remove `createChildHandler` and `updateChildHandler` (replaced by `createChildWithFullProfileHandler` and a thin per-section update). Add:
  - `POST /api/v1/children` → `CreateChildWithFullProfile`
  - `GET /api/v1/children/:child_id` → `GetChild` (returns child + room + leaving + billing profile join, no profile body)
  - `GET /api/v1/children/:child_id/profile`
  - `PATCH /api/v1/children/:child_id/profile`
  - `GET /api/v1/children/:child_id/contacts`
  - `PUT /api/v1/children/:child_id/contacts` (body bucketed by type: `{parent_carers:[...], emergency_contacts:[...], authorised_collectors:[...]}`; replaces all rows for that child in those types inside one tx)
  - `GET /api/v1/children/:child_id/health`
  - `PATCH /api/v1/children/:child_id/health`
  - `GET /api/v1/children/:child_id/safeguarding`
  - `PATCH /api/v1/children/:child_id/safeguarding`
  - `GET /api/v1/children/:child_id/consent`
  - `PUT /api/v1/children/:child_id/consent` (single row, no version)
  - `GET /api/v1/children/:child_id/funding`
  - `PATCH /api/v1/children/:child_id/funding`
  - `GET /api/v1/children/:child_id/collection-settings`
  - `PUT /api/v1/children/:child_id/collection-settings` (body `{ password?: string, over_18_collection_acknowledged?: bool }`; if password is present, hash and stamp updated_at/user/membership; if absent, leave hash as-is)
  - `GET /api/v1/children/:child_id/room-assignments`
  - `POST /api/v1/children/:child_id/room-assignments` (body `{ room_id, start_date }`; closes the current row inside the same tx)
  - `DELETE /api/v1/children/:child_id/room-assignments/:assignment_id` (sets end_date = today on that row; only valid if it is the current row)
  - `GET /api/v1/children/:child_id/billing-profile`
  - `PATCH /api/v1/children/:child_id/billing-profile`
  - `GET /api/v1/children/:child_id/leaving-record`
  - `POST /api/v1/children/:child_id/actions/mark-inactive` (kept, but now also writes a `child_leaving_records` row and updates `is_active`)

The `PATCH /api/v1/children/:child_id` route is removed (room placement no longer lives on children; identity edits go on `/profile` and the read-only `/children/:id` shows identity).

- `interfaces/http/dto.go` — split into per-resource DTO files in the same package: `child_dto.go`, `child_profile_dto.go`, `child_contacts_dto.go`, `child_health_dto.go`, `child_safeguarding_dto.go`, `child_consent_dto.go`, `child_funding_dto.go`, `child_collection_settings_dto.go`, `child_room_assignments_dto.go`, `child_billing_dto.go`, `child_leaving_dto.go`, `create_child_dto.go`. Each DTO file contains the request/response structs and the to-domain/to-response mappers for its resource.
- `domain/entities.go` → split into per-concept entity files (see §1). `domain.Child` no longer carries `PrimaryRoomID`, `CoreHourlyRateMinor`, `SiteCoreHourlyRateMinor` (replaced by `BillingProfileID` + `BillingProfile` join), `LeftAt`, `LeftReasonCode`, `LeftReasonNote`. `MissingRequirements()` and `EnrollmentComplete()` stay (they only check first_name, date_of_birth, start_date, has_guardian_link — last one is still true and stays a derived EXISTS). `AttendanceChild` keeps the same shape (no change to attendance view).
- `domain/repository.go` — becomes a composite of `ChildIdentityRepository`, `ChildProfileRepository`, `ChildContactRepository`, `ChildHealthProfileRepository`, `ChildSafeguardingProfileRepository`, `ChildConsentRepository`, `ChildFundingRepository`, `ChildCollectionSettingsRepository`, `ChildRoomAssignmentsRepository`, `ChildBillingProfileRepository`, `ChildLeavingRepository` interfaces. The postgres implementation file(s) implement them all from one `*pgxpool.Pool`.
- `domain/child_creator.go` and `domain/complete_registration.go` are deleted (no `ChildInfo` in this module; the create use case calls the repository directly).

#### Update bootstrap wiring

`api/internal/app/bootstrap/bootstrap.go`:
- Drop all `regprofileapp`/`regprofilepostgres`/`regprofilehandler`/`httpregistrationprofile` imports and var.
- Drop the `childCreatorAdapter` in `adapters.go` (no longer needed; the create use case uses the children repository directly).
- Replace the two handler instantiations with the new children handler that includes all the per-resource use cases. The route group `manager` continues to be the only registration/edit actor.
- Remove `submitHandler.RegisterRoutes(manager)`. Add the new children handler's per-resource routes to `manager` group.

`api/internal/app/bootstrap/people_routes_test.go` (and any other test that hits the registration routes): rewrite to use the new endpoints. The expected JSON shapes for `/children/:id` and the per-resource endpoints must be updated. Existing tests in that file are kept where possible and rewritten where required.

#### Update other modules that referenced `children.primary_room_id` or `core_hourly_rate_minor`

- `api/internal/modules/children/application/list_attendance.go`, `list_children.go`, `get_child.go` — drop the removed columns from the SQL projection and from the response DTO. The attendance list view in particular reads `is_active` only.
- `api/internal/modules/rooms/infrastructure/postgres/repository_test.go` (and any other reference) — `primary_room_id` lookups move to `child_room_assignments`. Update the rooms capacity query to JOIN against `child_room_assignments` where `is_current = true` instead of `children.primary_room_id`.
- `api/internal/modules/attendance/infrastructure/postgres/repository_test.go` — drop the `left_at`, `left_reason_code` UPDATE; the new mark-inactive path writes to `child_leaving_records` and updates `children.is_active`. Update the test SQL accordingly.
- `api/internal/modules/owner/interfaces/http/handler.go` — the `site_core_hourly_rate_minor` field in the children summary is still a JOIN to `branches.core_hourly_rate_minor`; the per-child `core_hourly_rate_minor` is gone. The owner handler continues to expose site-level rate only. Drop the `core_hourly_rate_minor` (per-child) field from the children summary DTO. Keep `site_core_hourly_rate_minor`.

### 3.4 Frontend

#### Web (Angular)

- `web/src/app/features/staff/data/staff-api.service.ts`:
  - Replace `getRegistrationProfile` → `getChildProfile`, `patchRegistrationProfile` → `patchChildProfile`, `setRegistrationCollectionPassword` → `setChildCollectionPassword`.
  - Replace `getRegistrationConsents` → `getChildConsent`, `createRegistrationConsent` → `updateChildConsent` (PUT instead of POST).
  - Drop `getRegistrationWorkflowStatus` and `createRegistrationCompletionAttestation`.
  - Replace `submitCompleteRegistration` → `createChildWithFullProfile` (POST `/children` instead of POST `/children/with-registration`).
  - Add: `getChildHealth`, `patchChildHealth`, `getChildSafeguarding`, `patchChildSafeguarding`, `getChildFunding`, `patchChildFunding`, `getChildContacts`, `putChildContacts`, `listChildRoomAssignments`, `createChildRoomAssignment`, `closeChildRoomAssignment`, `getChildBillingProfile`, `patchChildBillingProfile`, `getChildLeavingRecord`, `getChildCollectionSettings`, `putChildCollectionSettings`.
  - Adjust the `children` model: drop `primaryRoomId`, `coreHourlyRateMinor`, `leftAt`, `leftReasonCode`, `leftReasonNote` from `ChildRecord`/`ChildWritePayload` (these move to per-resource models).
- `web/src/app/features/staff/pages/manager-registration-intake/`:
  - Rename directory to `manager-child-edit/`. Rename selector to `app-manager-child-edit`. Routes: `/children/new` and `/children/:child_id/edit` (single component, mode inferred from URL).
  - Step 1: child identity (first_name, middle_name, last_name, date_of_birth, start_date, notes). Step 2: room placement (room_id, start_date; this also writes `child_room_assignments`). Step 3: profile (demographics, home, disability, access, routine care, GDPR declaration, registration date, section review flags). Step 4: health. Step 5: safeguarding. Step 6: contacts (parent/carers, emergency, authorised collectors). Step 7: funding. Step 8: consent. Step 9: collection settings (over-18 ack, optional collection password). Step 10: review and Create / Save.
  - On Submit: call `POST /children` (create mode) or a sequence of PATCH/PUT calls (edit mode) per §3.3 endpoints. The backend decides atomicity on create; on edit, each section is its own request and the stepper shows save status per step.
  - Delete the existing `manager-child-registration/` directory (the section editor on the child detail page).
- `web/src/app/features/staff/pages/manager-child-detail/manager-child-detail.component.ts`:
  - Read-only summary: child identity, current room, leaving record (if any), billing profile, list of all sub-records with a "View" link per section.
  - One button "Edit" that navigates to `/children/:child_id/edit`.
  - The existing "Registration" tab is gone. The existing "attendance / funding / guardians / invoices" tabs continue to live on this page.
- `web/src/app/features/staff/pages/manager-children/` and `components/child-form/`:
  - The `child-form` component is the quick "Add child" form (now thinner — it only collects identity + room and the manager is offered "Save and continue with full intake" which opens the stepper pre-filled). The standalone `POST /children` write goes away from this component; the form just collects identity + room and then navigates to the stepper.
- Models directory:
  - Rename `registration-profile.models.ts` → `child-profile.models.ts`. Update type names per §1.
  - Add new model files: `child-contacts.models.ts`, `child-health.models.ts`, `child-safeguarding.models.ts`, `child-consent.models.ts`, `child-funding.models.ts`, `child-collection-settings.models.ts`, `child-room-assignments.models.ts`, `child-billing-profile.models.ts`, `child-leaving.models.ts`. Each file owns the API DTO shape + the local form model for that step.
  - Delete `RegistrationProfile`, `RegistrationContactEntry`, `ConsentWritePayload`, `CompleteRegistrationPayload`, `CompleteRegistrationResponse`, `RegistrationWorkflowStatus`, `RegistrationProfileResponse`, `RegistrationProfileApiModel`, `CollectionPasswordPayload` from the model files (they move into the renamed/new files).
- Formatters:
  - `web/src/app/features/staff/utils/registration-profile-formatters.ts` → rename to `child-profile-formatters.ts`. The `formatCompletionStatus`/`getCompletionBadgeClass` helpers are deleted (no more "registration completion" concept). Keep any sex/religion/ethnicity mapping helpers; move them to the new model files.
- `web/src/app/features/staff/pages/manager-rooms/` (if the room assignment history is part of the rooms view): add a "Current children" list view that reads `child_room_assignments` where `is_current = true`. The existing room list view is otherwise unchanged.
- Routing in `web/src/app/features/staff/pages/.../...routes.ts` and `web/src/app/app.routes.ts`: replace the registration intake route with the new child edit route. Drop the section-editor route.

### 3.5 Tests

- `make test-api-repositories` (needs TEST_DATABASE_URL). Add new repository tests under `api/internal/modules/children/infrastructure/postgres/`:
  - `repository_test_children_create_test.go` (or extend the existing `repository_test.go`) — covers the new `Create` (without primary_room_id and core_hourly_rate_minor), `MarkInactive` (without left_reason_code/left_at/left_reason_note on the children table; assert that the new use case writes a `child_leaving_records` row).
  - New `repository_test_child_profiles.go` — round-trip a `child_profile`.
  - New `repository_test_child_contacts.go` — replace contacts for types, sort_order uniqueness.
  - New `repository_test_child_health_profiles.go`, `repository_test_child_safeguarding_profiles.go`, `repository_test_child_consents.go` (single row), `repository_test_child_funding_records.go`, `repository_test_child_collection_settings.go`, `repository_test_child_room_assignments.go` (insert + close + only-one-current), `repository_test_child_billing_profiles.go`, `repository_test_child_leaving_records.go`.
  - Update `repository_test.go` (children identity): remove assertions on `primary_room_id` and `core_hourly_rate_minor`; add assertions on the join of `child_room_assignments` for "current room" and the existence of a `child_billing_profiles` row after Create.
- Application tests under `api/internal/modules/children/application/`:
  - `create_child_with_full_profile_test.go` — covers required-field validation (first_name, date_of_birth, start_date, room_id, safeguard_ack) and the transaction path with mock repos. Existing three test cases from `submit_complete_registration_test.go` are reused, retitled, with updated expected errors.
  - `update_profile_test.go`, `replace_contacts_test.go`, `update_health_test.go`, `update_safeguarding_test.go`, `update_consent_test.go`, `update_funding_test.go`, `set_collection_password_test.go`, `create_room_assignment_test.go`, `close_room_assignment_test.go`, `update_billing_profile_test.go`, `mark_inactive_test.go` (now also writes a leaving record).
- HTTP tests:
  - `api/internal/app/bootstrap/people_routes_test.go` — rewrite `POST /children` expectations to the new full-payload response shape; rewrite the `GET /children/:id` expectations to drop removed fields; add coverage for the new per-resource endpoints.
  - Add `child_management_routes_test.go` covering the new manager-only routes (profile, contacts, health, safeguarding, consent, funding, collection-settings, room-assignments, billing-profile, leaving-record).
  - Drop any test that calls `POST /children/with-registration` or any of the old `/children/:id/registration-*` routes.
- Frontend tests:
  - `manager-child-edit.component.spec.ts` — covers both "new" and "edit" modes, the ten-step stepper navigation, and the submit/save flows.
  - `manager-child-detail.component.spec.ts` — covers the read-only view and the "Edit" navigation.
  - `child-form.component.spec.ts` — update to the new thin shape (identity + room only, then "Save and continue with full intake" navigates to the stepper).
  - `staff-api.service.spec.ts` — update the mock responses and the renamed methods; drop the registration-only methods.

### 3.6 Docs

- `CONTEXT.md`: the "Registration/Enrolment Profile (Post-MVP)" and downstream "Registration *" terms are renamed. The bounded context becomes "Child Management (Post-MVP)". A single glossary entry per new sub-record:
  - `## Child Management (Post-MVP)` — replaces the umbrella term.
  - `## Child Profile (Post-MVP)` — replaces "Registration/Enrolment Profile".
  - `## Child Contacts (Post-MVP)` — replaces "Registration Contact Entries".
  - `## Child Health Profile (Post-MVP)` — new.
  - `## Child Safeguarding Profile (Post-MVP)` — new.
  - `## Child Consent Record (Post-MVP)` — replaces "Paper-Form Consent Record", notes the collapse to single row.
  - `## Child Funding Record (Post-MVP)` — new, with the boundary against `Funding Profile (MVP)`.
  - `## Child Collection Settings (Post-MVP)` — replaces "Collection Password" entry.
  - `## Child Room Assignment (Post-MVP)` — new.
  - `## Child Billing Profile (Post-MVP)` — new.
  - `## Child Leaving Record (Post-MVP)` — new.
  - Drop the entries "Registration Profile Completeness", "Registration/Enrolment Workflow Completion", "Registration Completion Attestation", "Registration Reviewed/Complete", "Registration Reviewed/Complete Requirements", "Registration-to-Guardian Linking", "Registration Funding Support Notes", "Office-Use Enrolment Checklist", "Office-Use Checklist Completeness", "Office-Use Dates vs Enrollment Dates", "Registration Profile Partial Save", "Registration Unknown vs Confirmed None", "Guided Registration Required Answer", "Current Registration Profile", "Registration Profile Audit Redaction", "Registration Date (Post-MVP)", "Registration Contact Entries (Post-MVP)", "Registration Intake Child Creation (Post-MVP)", "Guided Registration Intake (Post-MVP)", "Manager-Assisted Registration Intake (Post-MVP)", "Registration Intake Entry Points (Post-MVP)" (these concepts are absorbed into the new model or are no longer required). Preserve the boundaries: "Manager-Assisted" is replaced with the same wording in the new "Child Management" entry ("Manager maintains the child record and surrounding profile in a guided stepper after a parent or carer has completed the physical form").
  - Add: "## Child Management Atomic Create (Post-MVP)" — manager creates a child in a single transaction that includes identity, profile, contacts, health, safeguarding, consent, funding, collection settings, room placement, and billing profile. All sections are optional except identity, room, and the safeguarding acknowledgement.
  - Add: "## Child Room Placement History (Post-MVP)" — child_room_assignments is 1:many per child; one row is current (end_date IS NULL).
  - Add: "## Child Leaving Record (Post-MVP)" — written when a child is marked inactive; reason code constrained to the existing enum; reason note required when reason_code='other'.
- `docs/POST-MVP-ROADMAP.md` and `docs/POST-MVP-FEATURE-SEQUENCE.md` (if exists): rename the "Registration and consent" lane to "Child management (identity, profile, contacts, health, safeguarding, consent, funding, collection)". The atomic-create milestone is unchanged.
- `docs/API-CONTRACT.openapi.yaml`: drop the `/children/with-registration` path; rename the `/children/:id/registration-*` paths to the new `/children/:id/{profile,contacts,health,safeguarding,consent,funding,collection-settings,room-assignments,billing-profile,leaving-record}` paths with their new request/response shapes; update `/children` and `/children/:id` to drop the removed fields; add a `POST /children` path that accepts the full create payload and returns the same shape as the existing 201 response.
- `docs/API-SCHEMA-STATE.md` and `docs/forms/*`: update the schema notes and any manager-intake form spec to point at the new model.
- No new ADR (see §2).

### 3.7 Verification commands

Run in this order before declaring the work done:

```bash
make migrate-up                                # applies 000034
cd api && go test ./internal/modules/children/... -count=1
cd api && go test ./internal/app/bootstrap/... -count=1
make test-api-repositories                     # needs TEST_DATABASE_URL
make sqlc-generate                             # regenerates sqlc
cd api && go build ./...
cd web && npm test                             # karma
```

Manual smoke: run `make run-api`, run `cd web && npm start`, log in as a manager, go through the new child add stepper end-to-end, then load the child detail page and verify it is read-only, then click Edit and verify the stepper opens pre-filled.

## 4. Acceptance criteria

- `rg "registrationprofiles|RegistrationProfile|SubmitCompleteRegistration|child_registration_|registration-profile|registration-consent|registration-workflow|registration-completion" api web` returns zero matches.
- The four old `child_registration_*` tables are gone from a fresh `make migrate-up` against an empty database.
- A single `POST /api/v1/children` call with the full payload creates the child plus its sub-records in one transaction. If any required field is missing or `safeguarding_reporting_acknowledgement` is false, the response is 400 and no rows are written.
- `GET /api/v1/children/:id` no longer exposes `primary_room_id`, `core_hourly_rate_minor`, `left_at`, `left_reason_code`, or `left_reason_note`. The current room is read from `child_room_assignments WHERE is_current`.
- The mark-inactive endpoint flips `children.is_active` to false and writes a row to `child_leaving_records` in one transaction.
- The manager child detail page is read-only; the only edit path is the guided stepper.
- All existing tests that referenced the registration module are deleted, rewritten, or moved; `go test ./...` and `npm test` pass.
- `make migrate-down` then `make migrate-up` is idempotent and reversible.
