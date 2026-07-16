# Invoice API Improvements

10 suggested improvements for the invoice API, prioritised by impact.

---

## 1. Wire Void/Cancel Endpoint

**Problem:** `Invoice.Void()` exists in the domain layer (`billing/domain/invoice.go`) but no HTTP handler exposes it. Managers cannot cancel draft invoices without direct database access.

**Proposal:**
- Add `POST /invoices/:invoice_id/void` handler in `billing/interfaces/http/handler.go`
- Accept `{ "reason": string }` body
- Guard: only `draft` invoices can be voided
- Emit a `InvoiceVoided` domain event
- Update frontend detail page with a "Void" action button (with confirmation dialog)

**Impact:** Closes a functional gap — managers can clean up erroneous drafts without DB access.

---

## 2. Register Missing Bulk Handlers in Routes

**Problem:** `generateDraftsHandler` and `bulkIssueInvoicesHandler` are defined in `handler.go` (lines ~305+) but never wired into `RegisterRoutes()`. This means bulk generation and bulk issue are only available through the scheduler, not via on-demand API calls.

**Proposal:**
- Wire `POST /invoices/drafts/generate` → `generateDraftsHandler`
- Wire `POST /invoices/drafts/bulk-issue` → `bulkIssueInvoicesHandler`
- Add proper request DTOs with `{ "billing_month": "YYYY-MM", "child_ids"?: uuid[] }`
- Add RBAC guard (manager role only)

**Impact:** Enables on-demand invoice runs without waiting for the scheduler cycle.

---

## 3. Server-Side Invoice Search

**Problem:** The manager invoice list (`GET /invoices`) only supports filter by `billing_month`, `status`, and `child_id`. Search by invoice number or child name is client-side only, which doesn't scale.

**Proposal:**
- Add `?q=<search_term>` query parameter to `GET /invoices`
- Implement server-side search using PostgreSQL `ILIKE` or `pg_trgm` on:
  - `invoice_number`
  - `child_first_name || ' ' || child_last_name`
- Add a GIN index on `invoices.invoice_number` for performance
- Return results in the existing paginated response

**Impact:** Enables efficient search for large invoice datasets without loading everything client-side.

---

## 4. Invoice PDF Generation Endpoint

**Problem:** Invoices are only viewable as HTML in the browser. No PDF export exists for archiving, emailing, or printing.

**Proposal:**
- Add `GET /invoices/:invoice_id/pdf` endpoint
- Use a Go PDF library (e.g., `jung-kurt/gofpdf` or `signintech/gopdf`)
- Render the same data as the detail endpoint but formatted as an A4 invoice:
  - Site profile header (nursery name, address)
  - Invoice number, billing month, due date
  - Line items table
  - Calculation summary
  - Payment instructions
- Return as `Content-Type: application/pdf` with `Content-Disposition: attachment`
- Add a "Download PDF" button on both manager and parent frontend detail pages

**Impact:** Essential for record-keeping, accountant handoff, and parent convenience.

---

## 5. Email Notifications on Invoice Events

**Problem:** Domain events `InvoiceIssued` and `InvoiceMarkedOverdue` are emitted but no subscriber sends emails to parents. Parents must manually check the portal.

**Proposal:**
- Create a `notifications` module with an email subscriber
- Subscribe to `InvoiceIssued` → send "New invoice" email with:
  - Invoice number, amount, due date
  - Link to parent portal invoice detail page
- Subscribe to `InvoiceMarkedOverdue` → send "Invoice overdue" reminder
- Use an email provider abstraction (e.g., Resend, SendGrid, or SMTP)
- Add email templates (HTML + plain text)
- Store notification log in `invoice_events` table for audit

**Impact:** Reduces late payments and improves parent experience by proactively communicating charges.

---

## 6. Credit Note / Adjustment Workflow

**Problem:** The schema has `adjusts_invoice_id` and `adjustment_reason_code/note` columns, suggesting credit notes were planned but never implemented. Managers have no way to issue partial refunds or corrections.

**Proposal:**
- Define a new `InvoiceKind` constant: `credit_note`
- Domain entity: `CreditNote` references an original `Invoice` via `adjusts_invoice_id`
- Status lifecycle: `draft → issued → applied` (simpler than invoice)
- Lines: negative amounts that reduce the original invoice balance
- API endpoints:
  - `POST /invoices/:invoice_id/credit-notes` — create credit note draft
  - `POST /invoices/:invoice_id/credit-notes/:id/issue` — issue credit note
- Effect on original invoice: reduce `amount_paid` or mark as partially credited
- Frontend: "Issue Credit Note" button on manager detail page

**Impact:** Critical for handling billing disputes, overcharges, and mid-term departures.

---

## 7. Configurable Overdue Grace Period & Reminders

**Problem:** The overdue transition job runs daily at 02:00 London time with no grace period. An invoice due on Monday is marked overdue at 00:00 Tuesday with no warning.

**Proposal:**
- Add `overdue_grace_days` setting to branch configuration (default: 3 days)
- Modify `MarkIssuedInvoicesOverdue` query: `due_at + grace_days < now()`
- Add a pre-overdue reminder event:
  - Emit `InvoiceDueSoon` 3 days before due date
  - Emit `InvoiceReminderOnDueDate` on the due date itself
- Email subscribers for both events with escalating tone

**Impact:** Reduces false "overdue" markings and gives parents a fair chance to pay on time.

---

## 8. Invoice Line Item CRUD on Drafts

**Problem:** Once a draft invoice is created, there's no API to add, update, or remove individual line items. The only option is to delete and recreate the entire invoice.

**Proposal:**
- `POST /invoices/:invoice_id/lines` — add a line (extra or ad_hoc kind only)
- `PUT /invoices/:invoice_id/lines/:line_id` — update description/amount on a line
- `DELETE /invoices/:invoice_id/lines/:line_id` — remove a line
- Guard: only `draft` status, only `extra`/`ad_hoc` line kinds (system-generated lines are immutable)
- Recalculate `total_minor` after each mutation
- Frontend: editable line items table on draft detail page

**Impact:** Gives managers fine-grained control over draft invoices without the destructive delete-and-recreate cycle.

---

## 9. Stripe Payment Link Sharing

**Problem:** Parents can only pay via the portal's "Pay now" button which creates a Stripe checkout session. Managers cannot generate a shareable payment link for ad-hoc payment collection (e.g., over the phone).

**Proposal:**
- Add `POST /invoices/:invoice_id/payment-link` endpoint
- Create a Stripe Payment Link (not Checkout Session) with the invoice amount
- Return the URL in the response
- Manager can copy and share via email, SMS, or WhatsApp
- Link expires after the invoice due date

**Impact:** Enables flexible payment collection for parents who struggle with the portal.

---

## 10. Invoice Export / Reporting Endpoint

**Problem:** No bulk export or reporting capability exists. Accountants and managers cannot download invoice data for financial reporting, VAT filing, or reconciliation.

**Proposal:**
- Add `GET /invoices/export` endpoint
- Accept filters: `billing_month_from`, `billing_month_to`, `status[]`, `format`
- Formats:
  - `csv` — one row per invoice with columns: invoice_number, child_name, billing_month, status, total, issued_at, due_date, paid_at
  - `csv-detail` — one row per line item (flattened)
- Add `GET /invoices/summary` endpoint returning aggregated metrics:
  - Total invoiced, collected, outstanding, overdue by month
  - Breakdown by line kind
- Both endpoints restricted to manager role

**Impact:** Essential for month-end financial reporting and accountant handoff.
