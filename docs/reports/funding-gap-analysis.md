# Funding Logic Gap Analysis

> Comparison of `docs/funding-logic.md` (spec) vs actual codebase implementation.

**Date:** 2026-07-06

---

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Fully implemented |
| 🟡 | Partially implemented |
| ❌ | Not implemented |

---

## 1. Funding Sources (§1)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Government Funding | 🟡 | `child_funding_records.funding_type` exists but limited to: `none`, `fifteen_hours`, `thirty_hours`, `two_year_old`, `custom`, `unknown`. No FRAS or Working Parent 9m+ distinct types. |
| Parent Payment | ✅ | Invoice system calculates parent balance after funded deduction. |
| Tax-Free Childcare | ❌ | No model, table, or field for TFC. |
| Employer Childcare Scheme | ❌ | Not modelled. |
| Childcare Voucher (Legacy) | ❌ | Not modelled. |
| Charity / Sponsor | ❌ | Not modelled. |
| Multiple funding sources per child | ❌ | `child_funding_records` has a UNIQUE constraint on `child_id` — only one funding record per child. |

**Gap:** The spec envisions a child combining government funding + TFC + parent payment. The current schema only supports one government funding record per child. No model for external payment sources.

---

## 2. Government Funding Types (§2)

| Funding Type | Status | Notes |
|-------------|--------|-------|
| Working Parent (9 months+) | ❌ | Not a distinct type. |
| FRAS (2-year-olds) | 🟡 | Mapped to `two_year_old` in CHECK constraint, but no FRAS-specific logic. |
| Universal 15 Hours | 🟡 | Mapped to `fifteen_hours`. |
| Additional 15 Hours (30h total) | 🟡 | Mapped to `thirty_hours`. |

**Gap:** No distinction between "Working Parent" and "Universal" within the 3/4-year-old types. No eligibility age validation per funding type.

---

## 3. Funding Delivery Model (§3)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Term-Time Funding (15h × 38w = 570h) | 🟡 | `funding_model` CHECK allows `term_time_only`. No automatic hour calculation from weeks. |
| Stretched Funding (570h ÷ 51w) | 🟡 | `funding_model` CHECK allows `stretched`. No automatic weekly hour derivation. |
| Configurable weeks per year | ❌ | No nursery calendar table defining open weeks. Hours stored as `funded_hours_per_week` — manual entry only. |

**Gap:** The spec describes automatic calculation (570h ÷ 51 weeks = 11.18h/week). The system stores `funded_hours_per_week` as a manually-entered decimal — no engine to compute stretched vs term-time hours.

---

## 4. Nursery Calendar (§4)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| 38/39/50/51/52 week operation | ❌ | No nursery calendar table. |
| Configurable open/closed weeks | ❌ | Not modelled. |

**Gap:** No `nursery_calendar` or equivalent table. Branch closures exist (`branch_closure_days` table from migration 000008) but only for individual dates, not week-level configuration.

---

## 5. Funding Hours Tracking (§5)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Allocated Hours | ✅ | `funding_profiles.funded_allowance_minutes` (per billing month). |
| Used Hours | 🟡 | Invoice calculation tracks `funded_deduction_minutes` per invoice. No cumulative used-hours counter. |
| Claimed Hours | ❌ | No funding claim model. |
| Remaining Hours | ❌ | Not computed or stored. |

**Gap:** Per-month allowance exists, but there's no running total of used vs remaining hours across the funding year. The spec envisions a `570 allocated / 415 used / 155 remaining` tracker.

---

## 6. Funding Allocation (§6)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Booking split between funded and private | ✅ | Invoice lines: `core_childcare` + `funded_deduction` lines. |
| Funded hours deducted from booking | ✅ | `funded_deduction_minutes` on invoice lines. |

**Status: ✅ Implemented.** The billing engine correctly splits bookings into funded deduction + private billable portions.

---

## 7. Session Funding (§7)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Morning/Afternoon/Full Day sessions | ✅ | `session_types` table with `start_time`, `end_time`. |
| Partial session funding | ✅ | Invoice engine handles partial funded deductions. |
| Multiple sessions | ✅ | `child_booking_pattern_entries` supports multiple sessions per day. |

**Status: ✅ Implemented.**

---

## 8. Holiday Funding (§8)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Government funding = 0 during holidays | ✅ | Funding profiles are per-billing-month; term-time children have no funded hours in holiday months. |
| Private bookings continue during holidays | ✅ | Ad-hoc bookings module exists. |

**Status: ✅ Implemented** via billing-month-scoped funding profiles and `term_time_only` flag.

---

## 9. Nursery Closure Logic (§9)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Christmas / Bank Holidays / INSET / Emergency | ✅ | `branch_closure_days` table (migration 000008). |
| No funding claimed on closed days | ✅ | Invoice engine excludes closure days from billable calculation. |

**Status: ✅ Implemented.**

---

## 10. Child Funding Eligibility (§10)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Funding Start Date | ✅ | `child_funding_records.funding_start_date`. |
| Funding End Date | ✅ | `child_funding_records.funding_end_date`. |
| Eligibility Code | ✅ | `child_funding_records.eligibility_code`. |
| Validation Status | ✅ | `child_funding_records.eligibility_code_validated`. |
| Renewal Date | ❌ | No field for renewal date. |
| Funding Type | ✅ | `child_funding_records.funding_type`. |

**Gap:** No renewal date tracking. No automated expiry warnings.

---

## 11. Funding Status (§11)

| Status | Implemented? | Notes |
|--------|-------------|-------|
| Pending | ❌ | No status field on funding record. |
| Approved | ❌ | |
| Rejected | ❌ | |
| Expired | ❌ | |
| Cancelled | ❌ | |

**Gap:** `child_funding_records` has no status field. Funding is either enabled (`funding_enabled = true`) or not. No workflow for pending → approved → expired lifecycle.

---

## 12. Funding Claim Status (§12)

| Status | Implemented? | Notes |
|--------|-------------|-------|
| Draft | ❌ | No funding claim model at all. |
| Submitted | ❌ | |
| Accepted | ❌ | |
| Partially Accepted | ❌ | |
| Rejected | ❌ | |
| Paid | ❌ | |

**Gap:** Entire funding claim subsystem is missing. No `FundingClaim`, `FundingClaimItem` tables. The spec's claim workflow (draft → submit → accept/reject → paid) does not exist.

---

## 13. Local Authority Information (§13)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Local Authority | ❌ | No `LocalAuthority` table. |
| Claim Period | ❌ | |
| Funding Rate | ❌ | No `FundingRate` table. |
| Payment Date | ❌ | |
| Claim Reference | ❌ | |
| Payment Reference | ❌ | |

**Gap:** Entire Local Authority management is missing. No rate versioning, no LA-specific settings.

---

## 14. Funding Rate (§14)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Rate by Child Age | ❌ | |
| Rate by Funding Type | ❌ | |
| Rate by Local Authority | ❌ | |
| Rate by Financial Year | ❌ | |

**Gap:** No `FundingRate` table. The `branches.core_hourly_rate_minor` and `term.site_hourly_rate_minor` exist but are nursery-level, not LA-specific or funding-type-specific.

---

## 15. Invoice Calculation (§15)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Total Fees | ✅ | `invoices.subtotal_minor`. |
| − Government Funding | ✅ | `invoices.funded_deduction_minor`. |
| − Discounts | ❌ | No discount model. |
| + Meals / Snacks / Consumables | 🟡 | `invoice_lines` with `line_kind = 'extra'` can represent these, but no consumable catalog. |
| + Extra Hours | ✅ | Ad-hoc bookings produce extra invoice lines. |
| + Late Collection | ❌ | No late collection fee model. |
| + Registration Fee | ❌ | |
| + Deposit | ❌ | |
| = Parent Balance | ✅ | `total_due_minor = subtotal_minor - funded_deduction_minor`. |

**Gap:** No discount, late collection fee, registration fee, or deposit models. Consumables are freeform extras, not a managed catalog.

---

## 16. Consumables (§16)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Meals, Snacks, Milk, Formula, Nappies, Wipes, Sun Cream, Trips, Activities | 🟡 | Can be added as ad-hoc booking extras, but no consumable product catalog. |

**Gap:** No predefined consumable items. No automatic meal charges.

---

## 17. Additional Charges (§17)

| Charge Type | Status | Notes |
|------------|--------|-------|
| Late Pickup | ❌ | |
| Extra Hours | ✅ | Ad-hoc bookings. |
| Extra Session | ✅ | Ad-hoc bookings. |
| Holiday Club | ❌ | |
| Deposit | ❌ | |
| Registration Fee | ❌ | |
| Administration Fee | ❌ | |
| Emergency Contact Fee | ❌ | |
| Transport Fee | ❌ | |

**Gap:** Only extra hours/sessions via ad-hoc bookings. No fee catalog for other charge types.

---

## 18. Funding Audit Trail (§18)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Track User, Date, Old Value, New Value, Reason | ✅ | `audit_logs` table with `actor_user_id`, `action_type`, `details` JSONB. Funding upserts write audit entries. |

**Status: ✅ Implemented.** Audit log captures funding changes with actor, timestamp, and details.

---

## 19. Funding Adjustments (§19)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Backdated Funding | ❌ | |
| Manual Adjustment | 🟡 | Can update `funded_allowance_minutes` per month, but no formal adjustment workflow. |
| Refund | ❌ | |
| Credit Note | ❌ | |
| Additional Funded Hours | 🟡 | Can increase allowance via upsert. |
| Reduced Funded Hours | 🟡 | Can decrease allowance via upsert. |

**Gap:** No formal adjustment, refund, or credit note system. Manual allowance changes are possible but not structured as adjustments.

---

## 20. Funding Reports (§20)

| Report | Status | Notes |
|--------|--------|-------|
| Funded / Used / Remaining / Claimed Hours | ❌ | |
| Parent / Government Contribution | ❌ | |
| Funding by Child / Room / Age / LA | ❌ | |
| Rejected Claims / Expired Eligibility | ❌ | |
| Forecast Funding Income | ❌ | |

**Gap:** No funding reports module. The `funding` module has an overview page but it's a per-child profile list, not a reporting engine.

---

## 21. Payment Allocation (§21)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Split Payments (Gov + TFC + Card) | ❌ | Payment is Stripe-only, one attempt per invoice. |
| Partial Payments | ❌ | Invoice must be paid in full. |
| Overpayments | ❌ | |
| Credit Balance | ❌ | |

**Gap:** Payment model is single-provider (Stripe), full-amount-only. No multi-source payment allocation.

---

## 22. Booking vs Attendance (§22)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Booked Hours | ✅ | `child_booking_patterns` + entries. |
| Attended Hours | ✅ | `attendance_sessions` with check-in/out times. |
| Absent / Holiday / Sick / No Show | 🟡 | `absence_markers` exist. No distinction between holiday/sick/no-show. |

**Gap:** Absence tracking exists but lacks granular absence reason types (holiday vs sick vs no-show).

---

## 23. Funding Validation Rules (§23)

| Rule | Status | Notes |
|------|--------|-------|
| Child eligible? | ✅ | `funding_enabled` check. |
| Funding dates valid? | ✅ | `funding_start_date` / `funding_end_date` checked. |
| Eligibility code valid? | 🟡 | `eligibility_code_validated` boolean exists. No external validation. |
| Funding hours available? | ✅ | `funded_allowance_minutes` checked during invoice generation. |
| Session claimable? | ❌ | |
| Nursery open? | ✅ | `branch_closure_days` checked. |
| Claim period open? | ❌ | No claim period model. |
| Duplicate claim? | ❌ | |
| Age eligible? | ❌ | No age-based validation. |

**Gap:** Core validation works. Missing: claim period checks, duplicate claim prevention, age eligibility validation.

---

## 24. Funding Priority Rules (§24)

| Priority | Status | Notes |
|----------|--------|-------|
| 1. Government Funding | ✅ | Funded deduction applied first. |
| 2. Tax-Free Childcare | ❌ | |
| 3. Childcare Voucher | ❌ | |
| 4. Employer Contribution | ❌ | |
| 5. Parent Payment | ✅ | Remainder is parent balance. |

**Gap:** Only government-first → parent-remainder flow. No multi-source priority engine.

---

## 25. Financial Year Support (§25)

| Spec Requirement | Status | Notes |
|-----------------|--------|-------|
| Financial Year | ❌ | No financial year entity. |
| Effective Date | ❌ | |
| Funding Rate Version | ❌ | |

**Gap:** No financial year versioning or rate history.

---

## 26. Suggested Database Tables (§26)

| Table | Status | Notes |
|-------|--------|-------|
| Child | ✅ | `children` |
| Parent | ✅ | `memberships` (role=parent) + `parent_membership_children` |
| Booking | ✅ | `child_booking_patterns` + entries |
| BookingSession | ✅ | `child_booking_pattern_entries` |
| Attendance | ✅ | `attendance_sessions` |
| FundingEligibility | 🟡 | `child_funding_records` (partial) |
| FundingAllocation | ✅ | `funding_profiles` |
| FundingRate | ❌ | |
| FundingRule | ❌ | |
| FundingClaim | ❌ | |
| FundingClaimItem | ❌ | |
| FundingAdjustment | ❌ | |
| FundingAudit | ✅ | `audit_logs` |
| FundingCalendar | ❌ | |
| Holiday | ❌ | |
| Closure | ✅ | `branch_closure_days` |
| Invoice | ✅ | `invoices` |
| InvoiceLine | ✅ | `invoice_lines` |
| Payment | ✅ | `payment_attempts` + `payment_reconciliation_records` |
| PaymentAllocation | ❌ | |
| LocalAuthority | ❌ | |

---

## 27. Funding Engine Processing Flow (§27)

| Step | Status | Notes |
|------|--------|-------|
| Child Booking | ✅ | |
| Validate Eligibility | ✅ | |
| Determine Funding Type | 🟡 | Type stored but no rule engine. |
| Determine Term-Time / Stretched | 🟡 | Flag stored, no calculation logic. |
| Calculate Available Hours | ✅ | Per-month allowance. |
| Allocate Funded Hours | ✅ | |
| Calculate Private Hours | ✅ | |
| Add Consumables | 🟡 | Manual extras only. |
| Add Extra Charges | 🟡 | Ad-hoc bookings only. |
| Generate Invoice | ✅ | |
| Allocate Payments | 🟡 | Stripe-only, full amount. |
| Generate Funding Claim | ❌ | |
| Create Audit Log | ✅ | |

---

## Summary

| Category | Spec Items | Implemented | Partial | Missing |
|----------|-----------|-------------|---------|---------|
| Funding Sources (§1) | 7 | 1 | 1 | 5 |
| Government Types (§2) | 4 | 0 | 4 | 0 |
| Delivery Model (§3) | 3 | 0 | 2 | 1 |
| Nursery Calendar (§4) | 2 | 0 | 0 | 2 |
| Hours Tracking (§5) | 4 | 1 | 1 | 2 |
| Funding Allocation (§6) | 2 | 2 | 0 | 0 |
| Session Funding (§7) | 3 | 3 | 0 | 0 |
| Holiday Funding (§8) | 2 | 2 | 0 | 0 |
| Closure Logic (§9) | 2 | 2 | 0 | 0 |
| Eligibility (§10) | 6 | 5 | 0 | 1 |
| Funding Status (§11) | 5 | 0 | 0 | 5 |
| Claim Status (§12) | 6 | 0 | 0 | 6 |
| Local Authority (§13) | 6 | 0 | 0 | 6 |
| Funding Rate (§14) | 4 | 0 | 0 | 4 |
| Invoice Calc (§15) | 9 | 4 | 1 | 4 |
| Consumables (§16) | 9 | 0 | 1 | 8 |
| Additional Charges (§17) | 9 | 2 | 0 | 7 |
| Audit Trail (§18) | 5 | 5 | 0 | 0 |
| Adjustments (§19) | 6 | 0 | 3 | 3 |
| Reports (§20) | 9 | 0 | 0 | 9 |
| Payment Allocation (§21) | 4 | 0 | 0 | 4 |
| Booking vs Attendance (§22) | 5 | 2 | 1 | 2 |
| Validation Rules (§23) | 9 | 4 | 1 | 4 |
| Priority Rules (§24) | 5 | 2 | 0 | 3 |
| Financial Year (§25) | 3 | 0 | 0 | 3 |
| Tables (§26) | 21 | 12 | 1 | 8 |
| Engine Flow (§27) | 12 | 6 | 4 | 2 |

### By Severity

**Critical Gaps (block production use for UK nurseries):**
1. Funding claims subsystem (§12) — required for LA reimbursement
2. Local Authority management (§13) — rates, contacts, claim periods
3. Funding rate versioning (§14, §25) — rates change annually
4. Multiple funding sources per child (§1) — most children combine Gov + TFC + parent
5. Tax-Free Childcare / Employer schemes (§1) — common payment methods

**Important Gaps (affect operational completeness):**
6. Funding status lifecycle (§11) — pending/approved/expired workflow
7. Funding reports (§20) — operational visibility
8. Renewal date tracking (§10) — expiry management
9. Age eligibility validation (§23)
10. Consumable catalog (§16)

**Nice-to-Have Gaps:**
11. Nursery calendar (§4)
12. Discount model (§15)
13. Late collection fees (§17)
14. Multi-source payment allocation (§21)
15. Absence reason granularity (§22)
