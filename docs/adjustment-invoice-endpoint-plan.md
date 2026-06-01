# Adjustment Invoice Endpoint Plan

## Goal and Scope

Document and, when explicitly promoted from deferred work, implement the full manager adjustment invoice endpoint for post-issue billing corrections.

Current API-25 outcome: do not implement backend code during month 1. Preserve existing schema hooks, document the deferred contract, and promote implementation only if UAT or pilot operations produce a real post-issue correction that must be handled inside the product before the next invoice/payment cycle.

Future endpoint: `POST /api/v1/invoices/:invoice_id/adjustments`.

Out of scope for the first implementation:

- Editing, deleting, refunding, regenerating, or directly offsetting the original issued invoice.
- Parent Stripe checkout for adjustment invoices.
- Parent invoice list/detail disclosure of adjustment invoices.
- Negative net credit invoices, credit carry-forward, account balances, refunds, partial offsets, and settlement allocation.
- Adjustment chains where an adjustment invoice adjusts another adjustment invoice.

## Context Alignment Notes

- `Adjustment Invoice (MVP)`: a follow-up invoice that corrects or offsets a previously issued invoice. It must link to the issued invoice it adjusts and carry a manager-provided reason.
- `Issued Invoice Edit Policy (MVP)`: issued invoices are immutable; changes require explicit adjustment rather than direct edits.
- `Adjustment Flow (MVP)`: post-issue billing changes are represented by a manager-created follow-up adjustment invoice linked to the original issued invoice, with a required reason. Month-1 operation may defer creating adjustment invoices until a validated pilot need exists.
- Parent invoice detail excludes adjustment internals by current glossary boundary.
- Invoice numbers use the invoice billing month in the `YYYYMM` segment, not the issue month.

## Decisions That Must Be Honored

- API-25 is deferred as of 2026-06-01 unless UAT or pilot operations require in-product adjustment handling.
- Do not add route, domain, repository, migration, or test code until the trigger condition is met or post-MVP prioritization explicitly selects this work.
- A pilot fallback exists before implementation: do not edit/delete the issued invoice or use ad hoc SQL; pause collection if needed, log the case, and handle the credit/debit outside the product unless that blocks the billing cycle.
- The endpoint is manager-only.
- Original invoice must be a monthly invoice in manager scope with status `issued`, `payment_failed`, `overdue`, or `paid`.
- Draft originals are rejected because existing draft regeneration is the correction path.
- Adjustment invoices cannot be adjustment targets.
- The endpoint creates an already-issued adjustment invoice in one confirmed action with `confirm: true`.
- The request requires both `reason_code` and non-empty `reason_note`.
- Reason code vocabulary: `attendance_correction`, `funding_correction`, `extra_correction`, `billing_error`, `other`.
- Adjustment lines use `line_kind = adjustment` and may have signed amounts.
- The net adjustment invoice total must be non-negative in the first implementation.
- Parent checkout remains limited to monthly invoices; adjustment invoices are not payable through Stripe.
- Parent invoice views remain monthly-invoice only.
- Manager invoice review may show adjustment invoices and their link to the original invoice.
- Reuse `invoice_runs.run_type = issue` with details `mode: "adjustment_issue"` unless a later schema decision explicitly adds a new run type.

## Step-by-Step Tasks

1. Confirm promotion trigger before coding.
   - Check UAT/pilot issue log for a real post-issue correction that cannot be handled by the fallback.
   - Record the triggering case in the implementation ticket.
   - If no trigger exists, stop and leave API-PM-01 deferred.

2. Re-read current docs and schema.
   - `CONTEXT.md`
   - `docs/API-CONTRACT-MVP.md`
   - `docs/API-SCHEMA-STATE.md`
   - `docs/MVP-30D-API-BACKEND-BACKLOG.md`
   - `api/db/migrations/000012_add_invoice_schema.up.sql`
   - `api/db/query/invoices.sql`

3. Add domain constants and DTOs.
   - Add `InvoiceKindAdjustment = "adjustment"`.
   - Add `LineKindAdjustment = "adjustment"`.
   - Add audit action such as `invoice_adjustment_issued`.
   - Add reason code constants or validation set for the adjustment reason vocabulary.
   - Add request/response DTOs matching the deferred contract.

4. Add sqlc queries.
   - Lock original invoice by tenant, branch, and invoice ID for update.
   - Insert an adjustment invoice with `invoice_kind = adjustment`, `status = issued`, `adjusts_invoice_id`, reason fields, issue fields, totals, period, and calculation details.
   - Insert adjustment invoice lines with `line_kind = adjustment`.
   - Reuse existing invoice number sequence allocation.
   - Reuse existing invoice run creation/completion where possible.
   - Run `make sqlc-generate` from `api/` or the repo's equivalent command.

5. Implement application use case.
   - Parse and validate original invoice ID.
   - Require `confirm = true`.
   - Validate reason code and non-empty reason note.
   - Require at least one line.
   - Validate line descriptions are non-empty.
   - Validate signed line amounts are non-zero unless product explicitly allows zero-value explanatory lines.
   - Sum lines and reject negative net total with `adjustment_negative_total_not_supported`.
   - Lock original invoice.
   - Reject not found, non-monthly, draft, and adjustment-chain cases.
   - Use the original invoice billing month and child ID for the adjustment invoice unless a newer decision supersedes this.
   - Create a single `invoice_runs` row with `run_type = issue`, status started, and final details containing `mode: "adjustment_issue"`, original invoice ID, adjustment invoice ID, reason code, and totals.
   - Allocate invoice number sequence using the adjustment invoice billing month.
   - Insert the already-issued adjustment invoice and lines in the same transaction.
   - Write audit event with actor, original invoice ID, adjustment invoice ID, reason, line count, total, invoice number, and request ID.
   - Complete the invoice run.

6. Add HTTP route.
   - Register manager route `POST /api/v1/invoices/:invoice_id/adjustments`.
   - Use existing auth middleware and manager guard.
   - Map domain errors to the error codes documented in `docs/API-CONTRACT-MVP.md`.
   - Return HTTP 201 with the adjustment invoice response.

7. Keep payment behavior unchanged.
   - Do not add checkout support for adjustment invoices.
   - Keep parent checkout rejection for non-monthly invoices.
   - Add or preserve tests proving adjustment invoices are not payable.

8. Keep parent invoice disclosure unchanged.
   - Do not include adjustment invoices in parent list/detail queries.
   - Preserve omitted adjustment internals in parent detail.
   - Add tests if query changes create any risk.

9. Extend manager invoice review only if needed.
   - Existing manager review already selects adjustment fields and invoice kind.
   - Verify it can show `invoice_kind = adjustment`, `adjusts_invoice_id`, `adjustment_reason_code`, and `adjustment_reason_note`.
   - If manager list filters need invoice kind support, add it only when the UI/API consumer requires it.

10. Update docs after implementation.
   - Move `POST /api/v1/invoices/:invoice_id/adjustments` from Deferred to the live invoice section in `docs/API-CONTRACT-MVP.md`.
   - Update `docs/API-SCHEMA-STATE.md` if migrations or constraints change.
   - Mark API-PM-01 complete in `docs/MVP-30D-API-BACKEND-BACKLOG.md`.
   - Update frontend backlog only if an adjustment UI is now in scope.

## Files to Create or Change

Current API-25 documentation outcome:

- `CONTEXT.md`
- `docs/API-CONTRACT-MVP.md`
- `docs/MVP-30D-API-BACKEND-BACKLOG.md`
- `docs/adjustment-invoice-endpoint-plan.md`

Future implementation files:

- `api/db/query/invoices.sql`
- `api/internal/platform/db/sqlc/invoices.sql.go` generated by sqlc
- `api/internal/modules/billing/domain/invoice.go`
- `api/internal/modules/billing/domain/repository.go`
- `api/internal/modules/billing/infrastructure/postgres/repository.go`
- `api/internal/modules/billing/application/create_adjustment_invoice.go`
- `api/internal/modules/billing/application/create_adjustment_invoice_test.go`
- `api/internal/modules/billing/interfaces/http/dto.go`
- `api/internal/modules/billing/interfaces/http/handler.go`
- `api/internal/app/bootstrap/billing_routes_test.go`
- `api/internal/modules/billing/infrastructure/postgres/repository_test.go` if repository integration coverage is needed
- `api/internal/modules/payments/application/create_checkout_session_test.go` if non-payable adjustment coverage is not already sufficient

## Verification Steps

For the current documentation-only API-25 outcome:

- `git diff --check`
- Manual doc review that API-25 is deferred and API-PM-01 exists.

For future implementation:

- `make sqlc-generate` from `api/` or the repo's equivalent command.
- `go test ./...` from `api/`.
- Route tests:
  - unauthenticated returns `401 unauthorized`
  - practitioner/parent returns `403 forbidden_role`
  - manager wrong tenant/branch returns `404 invoice_not_found`
  - draft original returns `409 invoice_not_issued`
  - adjustment original returns `409 adjustment_chain_not_supported`
  - non-monthly original returns `409 invoice_not_monthly`
  - missing/invalid reason returns documented 400 codes
  - negative net total returns `400 adjustment_negative_total_not_supported`
  - valid request returns 201 and creates issued adjustment invoice
- Repository/application tests:
  - original invoice remains unchanged
  - adjustment invoice has `invoice_kind = adjustment`
  - adjustment invoice links `adjusts_invoice_id`
  - reason fields are persisted
  - invoice number is assigned
  - issue fields and due fields are populated
  - invoice run has `run_type = issue` and details `mode = adjustment_issue`
  - audit event is written
  - invoice lines are immutable after creation because status is issued
- Payment tests:
  - parent checkout rejects adjustment invoice as non-monthly
  - manager payment status retry reason remains `invoice_not_monthly` for adjustment invoice
- Parent visibility tests:
  - parent list excludes adjustment invoices
  - parent detail for adjustment invoice returns not found

## Explicit Assumptions

- Adjustment invoice `billing_month` is the original monthly invoice billing month because the correction belongs to that billing period and invoice numbering uses billing month.
- Adjustment invoice `period_start_date` and `period_end_date` match the original invoice period unless a later product decision introduces correction-specific periods.
- Adjustment invoice `child_id` is the original invoice child ID.
- Adjustment invoice `currency_code` remains `GBP`.
- Adjustment invoice `subtotal_minor` and `total_due_minor` are the non-negative sum of signed adjustment lines for the first implementation; `funded_deduction_minor` remains `0`.
- `calculation_details` may use a compact adjustment-specific JSON object containing original invoice ID, reason, and line summary rather than attendance/funding calculation snapshots.
- No ADR is needed for the current deferral because the decision is documented in the backlog and contract, is easy to revisit, and follows the existing month-1 scope lock.
