# UK NMS — Children Data Completeness Analysis

**Date:** 2026-07-04
**Scope:** Evaluate current children dataset vs UK Nursery Management System regulatory + operational requirements.
**Stance:** Strict, compliance-focused, realistic for live UK nursery operations.

Source inventory drawn from:
- `api/internal/modules/children/domain/*.go`
- `api/db/migrations/000001_baseline.up.sql`
- `web/src/app/features/staff/models/child-profile.models.ts`

---

## 1. Coverage Assessment

| Domain | Status | Notes |
|---|---|---|
| Identity & Profile | **Covered** | Strong. Legal name, DOB, address, sex, ethnicity, language, disability all present. Missing nationality. |
| Guardians | **Mostly covered** | Parent-carer, emergency contact, authorised collector typed. Parental responsibility flag present. Missing contact priority hierarchy + per-channel permissions. |
| Medical & Safeguarding | **Partial** | Strong safeguarding. Medical is mostly free-text notes. No structured allergies, no medication consent per dose, no GP practice ID. SEN is a single boolean — no EHCP. |
| Booking & Attendance | **Mostly covered** | Booking patterns + attendance sessions/events all present. Absence tracking weak (no reason codes, no authorised/unauthorised flag). |
| Funding & Finance | **Partial** | 15/30h eligibility code + benefits captured. Missing term-based allocation, split-funding across providers, supplemental charges, discount rules. |
| Consent & Permissions | **Mostly covered** | Rich GDPR + outing + photo consent set. Missing separate EYFS observation consent, biometric consent, separate permanent/ephemeral data-sharing scope. |
| EYFS / Learning Development | **Not covered** | Zero tables, zero fields. No key person, no observations, no learning goals, no progress. |
| Operational | **Partial** | Room assignment, billing, leaving record present. No care ratios, no staff assignment, no waiting list status. |

---

## 2. Gap Analysis by Domain

### 2.1 Identity & Profile

| Missing field | Why needed |
|---|---|
| `nationality` | Census + safeguarding. |
| `legal_name_vs_known_as` distinction | Birth certificate vs daily-use. |
| `place_of_birth` | Some LA returns ask. |
| `previous_setting` | Transition continuity (EYFS 3.44). |
| `language_interpreter_needed` | Parent comms access. |

### 2.2 Guardians

| Missing field | Why needed |
|---|---|
| `contact_priority_order` | Emergency contact hierarchy currently inferred via `sort_order`. Not enforced. |
| `parental_responsibility_legal_basis` (court order ref) | PR disputes, CAO s.8. |
| `communication_channels` (sms/email/phone/app) per guardian | GDPR minimisation. |
| `lives_with_child` flag | Custody arrangements. |
| `separated_parent_scenario` fields — pickup restrictions | Safeguarding flashpoint #1 in UK nurseries. |

### 2.3 Medical & Safeguarding

| Missing field | Why needed |
|---|---|
| **Structured allergy entity** (allergen + severity + reaction + action plan) | Currently free text in `dietary_requirements_notes`. Safeguarding critical. |
| Medication entity: `name`, `dose`, `route`, `frequency`, `administering_staff_id`, `last_given_at`, `consent_expiry` | EYFS 3.44 medication policy. |
| `immunisation_schedule_due` | Public Health England schedule. |
| `gp_practice_code` (NHS Spine lookup) | GP just free text now. |
| `nhs_number` | UK-specific identifier; required by some LA funding returns. |
| `ehcp` (Education Health Care Plan) — exists, status, review date | SEND Code of Practice. |
| `sen_status` (K / K+ / EHCP) | Early Years Census return. |
| `senco_assigned` | Safeguarding + curriculum. |
| `restricted_access_order` (injunction reference) | Pickup safety. |
| `anaphylaxis_plan`, `asthma_plan`, `epi_pen_expiry` | Clinical safety. |

### 2.4 Booking & Attendance

| Missing field | Why needed |
|---|---|
| `attendance_status` enum (present / authorised_absence / unauthorised_absence / late) | Funding claim audit. |
| `absence_reason_code` | EYFS + funding reconciliation. |
| `funded_vs_private_session_split` on each session | Funding claim accuracy. |
| `session_actual_minutes` vs `session_scheduled_minutes` for funding audit | LA over-claim clawback. |
| `grace_period` rules (15 min either side, funded) | DfE funding rules. |
| `sign_in_method` (parent app / staff / biometric) | Audit trail. |

### 2.5 Funding & Finance

| Missing field | Why needed |
|---|---|
| `funding_term_allocation` (autumn / spring / summer "stretched" vs "standard") | DfE Headcount task. |
| `split_funding_partner_provider` (URN) | DfE rules: child can split across 2 providers. |
| `supplemental_charge` consent + amount | EYPP, SENIF, DAF eligibility. |
| `eligibility_code_expiry_date` | 30h codes expire every 3 months. |
| `discretionary_discount` rules | Funding ineligible sessions discount. |
| `retainer_fee` terms | Sibling discount, termly billing terms. |
| `funding_negative_response` from eligibility checker | HMRC sync failure trail. |
| `early_years_pupil_premium` (EYPP) flag | Census. |
| `disability_access_fund` (DAF) flag | Census. |

### 2.6 Consent & Permissions

| Missing field | Why needed |
|---|---|
| `consent_given_at` per consent (currently single `signed_date` for whole block) | GDPR auditability. |
| `consent_withdrawn_at` + `withdrawn_by` | GDPR right to withdraw. |
| `eyfs_observation_consent` (separate from generic photo) | Learning journey compliance. |
| `biometric_consent` (if fingerprint signin used) | Protection of Freedoms Act 2012. |
| `data_sharing_third_parties` (LA, NHS, social services) named | UK GDPR Art 13. |
| `consent_version` snapshot (which policy version did they agree to) | Policy update legal defence. |
| `emergency_treatment_specifics` (anaesthetic, blood transfusion) | Some faith requirements. |

### 2.7 EYFS / Learning Development — **ENTIRE DOMAIN MISSING**

Nothing exists. Need:

| Required field | Why needed |
|---|---|
| `key_person_id` (staff FK) | EYFS statutory framework s.1.10. |
| `observations` (text + photo + area of learning + age-band) | EYFS assessment. |
| `learning_goals` tracking (7 areas: PSED, CL, PD, L, M, UW, EAD) | EYFS. |
| `progress_summary` + `next_steps` | Parent engagement (EYFS 2.2). |
| `two_year_progress_check` date + report | EYFS statutory s.2.3. |
| `transition_records` (in / out) | Continuity. |
| `assessment_snapshot` termly | LA moderation. |
| `interests` / `schemas` | Pedagogy. |

### 2.8 Operational

| Missing field | Why needed |
|---|---|
| **Staff:child ratio** per room per session (EYFS 3.28-3.35) | Statutory. Currently zero enforcement. |
| `room_capacity_age_range` (0-2: 1:3, 2-3: 1:5, 3+: 1:8) | Ratio calc basis. |
| `staff_assignment` per session + qualification level | Ratio qualification rules. |
| `waiting_list` entity (joined, offered, declined, position) | Occupancy pipeline. |
| `enrollment_status` enum (active / pre-booked / waiting_list / suspended / left) | Currently `is_active` boolean only. |
| `safeguarding_lead_on_duty` log | EYFS designated lead. |
| `fire_evacuation_register` source flag | Statutory register. |

---

## 3. Risk Analysis

### 3.1 Regulatory (EYFS, GDPR, SEND Code of Practice)

| Risk | Likelihood | Impact | Mitigation gap |
|---|---|---|---|
| **Missing key person assignment** | Certain | High | EYFS statutory framework s.1.10 breach. |
| **No EYFS learning records** | Certain | Critical | EYFS assessment requirements breach. OFSTED inadequate judgement. |
| **No two-year progress check** | Certain | High | EYFS s.2.3 — must complete by child's 3rd birthday. |
| **Single GDPR block consent** | Probable | High | UK GDPR Art 7 — granular consent required. |
| **No consent versioning** | Probable | Medium | Policy update creates legal exposure. |
| **SEN tracked as boolean** | Probable | High | SEND Code of Practice breach on EHCP tracking. |
| **No medication entity** | Probable | Critical | EYFS s.3.44-3.47 breach. OFSTED immediate action. |

### 3.2 Financial (Funding, Billing)

| Risk | Likelihood | Impact | Mitigation gap |
|---|---|---|---|
| **No 30h code expiry tracking** | Certain | High | Parent loses entitlement, nursery absorbs cost. |
| **No split funding support** | Probable | Medium | LA claim rejected. |
| **No term allocation** | Certain | High | Headcount census rejected by LA. |
| **Absence not classified** | Probable | High | Funding over-claim → clawback + penalty. |
| **No EYPP/DAF flag** | Probable | Medium | Missed supplementary funding. |
| **No structured discount rules** | Probable | Medium | Billing errors, parent disputes. |

### 3.3 Safeguarding (Child Safety)

| Risk | Likelihood | Impact | Mitigation gap |
|---|---|---|---|
| **Allergies in free text** | Certain | Critical | Anaphylaxis risk. Death. (Multiple UK coroner cases.) |
| **No structured medication consent** | Probable | Critical | Wrong-dose administration. |
| **No restricted-access order field** | Probable | Critical | Abusive parent collects. |
| **Pickup restrictions not modelled** | Probable | Critical | Court order breach. |
| **No epi-pen / asthma expiry** | Probable | High | Stale medication. |
| **No sign-out audit for custody disputes** | Probable | High | Evidentiary gap. |
| **No NHS number** | Probable | Medium | Delayed medical intervention if emergency. |

### 3.4 Operational

| Risk | Likelihood | Impact | Mitigation gap |
|---|---|---|---|
| **No care ratios enforced** | Certain | Critical | EYFS 3.28 breach. OFSTED immediate action. |
| **No staff qualification tracking for ratios** | Certain | High | Level 3 / EYTS rule breach. |
| **`is_active` boolean for enrollment** | Certain | Medium | No waiting list pipeline. Revenue lost. |
| **No fire register source** | Probable | High | Statutory register non-compliance. |
| **No absence reason codes** | Probable | Medium | Inability to identify patterns of concern. |

---

## 4. Ideal Data Model (Normalized)

```
Child (1) ──── (N) GuardianContact
   │
   ├─ (1)─(1) ChildProfile
   ├─ (1)─(1) MedicalProfile ─── (N) Allergy
   │                          ├─ (N) Medication
   │                          └─ (1) GPPractice (FK or embedded)
   ├─ (1)─(1) SafeguardingProfile ─ (N) ProfessionalReferral
   ├─ (1)─(1) SENProfile (EHCP, SENDCO, status)
   ├─ (1)─(N) BookingPattern ─ (N) BookingSession
   ├─ (1)─(N) AttendanceSession ─ (N) AttendanceEvent
   ├─ (1)─(N) FundingProfile ─ (N) FundingTermAllocation
   │                      └─ (N) SplitFundingPartner
   ├─ (1)─(N) ConsentRecord  (one row per consent type, versioned)
   ├─ (1)─(1) EYFSProfile (key_person_id FK)
   │       ├─ (N) Observation
   │       ├─ (N) LearningGoalAssessment
   │       ├─ (1)─(0..1) TwoYearProgressCheck
   │       └─ (N) AssessmentSnapshot
   ├─ (1)─(N) RoomAssignment
   ├─ (1)─(N) StaffAssignment (key person, buddy)
   ├─ (1)─(N) Invoice ─ (N) InvoiceLine
   └─ (1)─(0..1) LeavingRecord
```

### Suggested table skeletons

```sql
-- Allergy entity (structured)
CREATE TABLE child_allergies (
  id UUID PRIMARY KEY,
  child_id UUID NOT NULL REFERENCES children(id),
  allergen TEXT NOT NULL,              -- milk, peanuts, latex
  severity TEXT NOT NULL,              -- mild | moderate | severe | anaphylactic
  reaction_description TEXT,
  action_plan_url TEXT,                -- stored PDF
  action_plan_reviewed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Medication entity
CREATE TABLE child_medications (
  id UUID PRIMARY KEY,
  child_id UUID NOT NULL REFERENCES children(id),
  name TEXT NOT NULL,
  dose TEXT NOT NULL,
  route TEXT,                          -- oral | inhalation | topical | injection
  frequency TEXT,
  administering_staff_id UUID REFERENCES staff(id),
  last_administered_at TIMESTAMPTZ,
  consent_expires_at TIMESTAMPTZ,
  prescriber TEXT
);

-- SEN profile (EHCP)
CREATE TABLE child_sen_profiles (
  child_id UUID PRIMARY KEY REFERENCES children(id),
  sen_status TEXT NOT NULL,            -- none | k | k_plus | ehcp
  ehcp_date DATE,
  ehcp_review_due DATE,
  sendco_staff_id UUID REFERENCES staff(id),
  sen_notes TEXT,
  funding_hours_per_week NUMERIC(4,1)
);

-- Consent record (per type, versioned, withdrawable)
CREATE TABLE child_consent_records (
  id UUID PRIMARY KEY,
  child_id UUID NOT NULL REFERENCES children(id),
  consent_type TEXT NOT NULL,          -- photo | outings | emergency_treatment | eyfs_observation | data_sharing | biometric
  granted_at TIMESTAMPTZ,
  granted_by_guardian_id UUID REFERENCES child_contacts(id),
  withdrawn_at TIMESTAMPTZ,
  withdrawn_by_user_id UUID,
  policy_version TEXT NOT NULL,
  notes TEXT
);

-- EYFS — key person + observation + goals
CREATE TABLE child_eyfs_profiles (
  child_id UUID PRIMARY KEY REFERENCES children(id),
  key_person_staff_id UUID NOT NULL REFERENCES staff(id),
  buddy_staff_id UUID REFERENCES staff(id),
  started_at DATE NOT NULL,
  two_year_check_date DATE,
  two_year_check_summary TEXT
);

CREATE TABLE child_eyfs_observations (
  id UUID PRIMARY KEY,
  child_id UUID NOT NULL REFERENCES children(id),
  observed_at TIMESTAMPTZ NOT NULL,
  observed_by_staff_id UUID REFERENCES staff(id),
  area_of_learning TEXT NOT NULL,      -- PSED | CL | PD | L | M | UW | EAD
  age_band TEXT,
  observation_text TEXT,
  photo_url TEXT,
  next_steps TEXT
);

-- Funding allocation per term
CREATE TABLE child_funding_term_allocations (
  id UUID PRIMARY KEY,
  child_id UUID NOT NULL REFERENCES children(id),
  academic_year INT NOT NULL,
  term TEXT NOT NULL,                  -- autumn | spring | summer
  stretched BOOLEAN NOT NULL DEFAULT FALSE,
  funded_hours_claimed NUMERIC(5,2) NOT NULL,
  headcount_submitted_at TIMESTAMPTZ,
  partner_provider_urn TEXT,           -- split funding
  CONSTRAINT one_per_child_term UNIQUE (child_id, academic_year, term)
);

-- Care ratio per room/session
CREATE TABLE room_staff_ratios (
  id UUID PRIMARY KEY,
  room_id UUID NOT NULL REFERENCES rooms(id),
  age_band TEXT NOT NULL,              -- 0_2 | 2_3 | 3_plus
  ratio_required TEXT NOT NULL,        -- 1:3 | 1:5 | 1:8
  effective_from DATE,
  effective_to DATE,
  qualification_required TEXT          -- level_3 | eyts | qts
);

-- Waiting list
CREATE TABLE children_waiting_list (
  child_id UUID PRIMARY KEY REFERENCES children(id),
  joined_at DATE NOT NULL,
  position INT NOT NULL,
  status TEXT NOT NULL,                -- waiting | offered | accepted | declined
  desired_start_date DATE,
  notes TEXT
);
```

---

## 5. Completeness Score

Weighted by domain per UK NMS criticality.

| Domain | Weight | Score (0–100) | Weighted |
|---|---:|---:|---:|
| Identity & Profile | 10% | 82 | 8.2 |
| Guardians | 15% | 70 | 10.5 |
| Medical & Safeguarding | 20% | 45 | 9.0 |
| Booking & Attendance | 15% | 72 | 10.8 |
| Funding & Finance | 15% | 55 | 8.25 |
| Consent & Permissions | 10% | 75 | 7.5 |
| EYFS / Learning Development | 10% | 0 | 0.0 |
| Operational | 5% | 45 | 2.25 |

### **Total: 56.5 / 100 — Non-compliant for production UK nursery use.**

### Grade justification per domain

- **Identity (82):** Strong core. Missing only nationality + transition context.
- **Guardians (70):** Good typing. No priority enforcement, no custody scenario fields.
- **Medical/Safeguarding (45):** Safeguarding strong but medical dangerously thin. **Allergies-as-free-text is the single biggest child-safety exposure.**
- **Booking/Attendance (72):** Tables exist. Absence classification + funding audit fields absent.
- **Funding (55):** Core eligibility captured but no term allocation, no split funding, no expiry tracking. **Cannot run a valid LA headcount return.**
- **Consent (75):** Rich. Loses points for un-versioned, single-block consent.
- **EYFS (0):** Zero. **Hard blocker for OFSTED registration.**
- **Operational (45):** No ratios. No staff qualification for ratio calc. Boolean enrollment blocks waiting list.

---

## 6. Recommended Priority Order

| # | Action | Driver |
|---|---|---|
| 1 | Build structured Allergy + Medication entities | Safeguarding. Lives at stake. |
| 2 | Implement EYFS module (key person + observations + 2yr check) | Regulatory blocker. |
| 3 | Add care-ratio engine + staff qualification tracking | Regulatory + operational. |
| 4 | Restructure consent to per-type, versioned, withdrawable | GDPR. |
| 5 | Add SEN profile (EHCP + SENDCO) | SEND Code of Practice. |
| 6 | Funding term allocation + 30h code expiry | Financial. LA rejection. |
| 7 | Absence classification + funded session split | Financial clawback. |
| 8 | Waiting list entity + enrollment_status enum | Revenue pipeline. |
| 9 | Restricted-access / custody orders on guardians | Safeguarding. |
| 10 | NHS number + structured GP practice | Medical audit. |

---

## 7. Conclusion

Current dataset scores **56.5 / 100** against UK NMS requirements.

**Production-ready for:** Identity capture, basic guardian management, booking patterns, basic safeguarding flagging, attendance logging, basic invoicing.

**Not production-ready for:**
- OFSTED-registered EYFS delivery (no learning records, no key person)
- Safe medication administration (free-text only)
- Statutory ratio compliance (not modelled)
- DfE / LA funding claims (no term allocation, no split funding)
- Full GDPR consent management (single-block, un-versioned)
- SEND Code of Practice (boolean only)

**Bottom line:** Strong back-office admin baseline. **Cannot legally operate a UK nursery today.** Prioritise allergy/medication (child safety), EYFS module (regulatory blocker), ratios (statutory), then funding + SEN to reach compliance.
