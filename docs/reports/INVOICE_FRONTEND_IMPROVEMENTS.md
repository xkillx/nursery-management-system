# Invoice Frontend Improvements

10 suggested improvements for the Angular invoice frontend, prioritised by impact.

---

## 1. Draft Invoice Edit Page with Inline Line Item Editing

**Problem:** The create page (`manager-invoice-create`) allows adding lines, but once a draft exists there's no edit page. Managers must delete and recreate drafts to fix mistakes.

**Proposal:**
- Create `manager-invoice-edit` component reusing the create form's line item editor
- Route: `/staff/invoices/:invoiceId/edit` (guard: draft status only)
- Pre-populate form from `GET /invoices/:id`
- Support inline add/edit/remove for `extra` and `ad_hoc` line kinds
- System-generated lines (`core_childcare`, `funded_deduction`, `hourly`) shown as read-only with a "Regenerate" button
- Auto-save draft changes (debounced PATCH requests) or explicit Save button
- Recalculate totals in real-time as lines change

**Impact:** Eliminates the most common workflow friction — correcting draft invoices.

---

## 2. Invoice PDF Download Button

**Problem:** No way to download or print a formatted invoice. Parents and managers need PDFs for records, accountants, and expense claims.

**Proposal:**
- Add "Download PDF" button to:
  - Manager detail page header actions
  - Parent detail page header actions
  - Parent list page (per-row action)
- Call `GET /invoices/:id/pdf` (new API endpoint)
- Trigger browser download with filename `INV-202607-0001.pdf`
- Show loading spinner while PDF generates
- Handle errors gracefully (toast notification)

**Impact:** High-value feature for both parent convenience and manager record-keeping.

---

## 3. Overdue Invoice Dashboard Widget

**Problem:** The manager invoice list has metric cards but no dedicated view for overdue invoices that need attention. Managers must manually filter by "overdue" status.

**Proposal:**
- Create an "Invoice Collections" dashboard widget on the manager home page
- Show:
  - Total overdue amount (prominent, red)
  - Count of overdue invoices
  - Top 5 oldest overdue invoices with child name, amount, days overdue
  - Quick action: "View all overdue" → navigates to filtered list
- Auto-refresh on page load (no polling needed)
- Use the existing `invoice-metrics` component pattern

**Impact:** Surfaces critical collection information at a glance without extra clicks.

---

## 4. Bulk Actions Toolbar on Invoice List

**Problem:** Managers can only issue invoices one at a time via the detail page. For end-of-month workflows with 50+ drafts, this is extremely tedious.

**Proposal:**
- Add checkboxes to the invoice table rows (draft status only)
- "Select all" checkbox in table header
- Floating action bar at bottom when items selected:
  - "Issue X invoices" button with confirmation dialog
  - "Export selected" button
- Show selection count and total amount
- Optimistic UI: mark selected invoices as "issuing..." with spinner
- Handle partial failures gracefully (show which invoices failed)

**Impact:** Transforms end-of-month invoice runs from a painful chore into a two-click operation.

---

## 5. Invoice Creation Wizard with Step-by-Step Flow

**Problem:** The current create page is a single long form. For complex invoices with multiple line types, funding considerations, and ad-hoc bookings, the form can be overwhelming.

**Proposal:**
- Convert to a stepper/wizard (matching the existing registration stepper pattern):
  1. **Child & Month** — select child, billing month, view prefill summary
  2. **Review Lines** — show auto-generated lines (core, funded, hourly), allow editing
  3. **Add Extras** — add extra/ad-hoc lines with description and amount
  4. **Summary & Confirm** — show final total, line breakdown, due date
- Each step validates before allowing progression
- Back button preserves state
- "Save as Draft" available at any step
- "Create & Issue" only on final step

**Impact:** Reduces errors in manual invoice creation and makes the process more approachable for new staff.

---

## 6. Real-Time Invoice Status Timeline

**Problem:** The detail page shows an audit trail but it's a static list. There's no visual timeline showing the invoice lifecycle progression.

**Proposal:**
- Replace the current audit trail list with a visual timeline component:
  - Vertical timeline with nodes for each status transition
  - Colour-coded: green (positive: issued, paid), red (negative: overdue, payment_failed), grey (draft, void)
  - Each node shows: timestamp, who triggered it, details (e.g., "Paid via Stripe, £450.00")
  - Pending states shown as dashed outlines (e.g., "Due in 5 days")
- Responsive: horizontal on desktop, vertical on mobile
- Reusable component in both manager and parent views

**Impact:** Makes invoice history scannable at a glance and more professional-looking.

---

## 7. Parent Invoice Comparison / Statement View

**Problem:** Parents can only view one invoice at a time. There's no way to see a monthly statement or compare charges across months.

**Proposal:**
- Add a "Statement" tab to the parent invoices page
- Show a table: rows = billing months, columns = child (if multiple children)
- Each cell shows: total amount, status badge, link to detail
- Footer row: totals
- Date range selector (last 3, 6, 12 months)
- Optional: simple bar chart showing monthly charge trend
- "Download statement" as CSV

**Impact:** Gives parents financial clarity and reduces "what was this charge for?" support queries.

---

## 8. Inline Payment Status on Manager List

**Problem:** The manager list shows invoice status but not payment status. Managers must click into each invoice to see if a payment attempt was made.

**Proposal:**
- Add a "Payment" column to the invoice table:
  - `paid` → green checkmark + paid date
  - `payment_failed` → red X + "Retry available" badge if applicable
  - `issued` (unpaid) → grey dash + "Awaiting payment"
  - `overdue` → red warning + days overdue
- Add a small payment status chip next to the main status badge
- On hover/tap: show tooltip with last payment attempt date and method

**Impact:** Eliminates the need to click into each invoice to check payment status.

---

## 9. Invoice Filtering by Date Range + Multi-Status

**Problem:** The current filter uses date range presets and single-status tabs. Managers can't combine "overdue + payment_failed" or filter by custom date ranges easily.

**Proposal:**
- Replace status tabs with multi-select status filter (checkboxes):
  - ☑ Draft ☑ Issued ☑ Payment Failed ☑ Overdue ☑ Paid ☑ Void
- Improve date range picker:
  - Keep presets (This month, Last 3 months, etc.)
  - Add custom range with proper date picker (from/to)
  - Remember last used filter in localStorage
- Add child name autocomplete search (server-side, not client-side)
- "Clear all filters" button
- URL state sync: filters reflected in query params for bookmarkable links

**Impact:** Enables power users to find exactly the invoices they need without multiple clicks.

---

## 10. Keyboard Shortcuts & Quick Actions

**Problem:** Common invoice workflows (issue, navigate, search) require multiple mouse clicks. Power users processing many invoices would benefit from keyboard shortcuts.

**Proposal:**
- Invoice list page:
  - `j` / `k` — navigate up/down through list
  - `Enter` — open selected invoice
  - `s` — focus search box
  - `f` — open filter panel
- Invoice detail page:
  - `i` — issue invoice (with confirmation)
  - `e` — edit invoice (if draft)
  - `p` — download PDF
  - `←` / `→` — navigate to prev/next invoice in list
  - `Esc` — back to list
- Global:
  - `?` — show keyboard shortcut overlay
  - `Ctrl+K` / `Cmd+K` — command palette (search invoices, navigate)
- Add shortcut hints in tooltips and the help overlay
- Use Angular `@HostListener` for keyboard events with proper focus management

**Impact:** Dramatically speeds up invoice processing for managers handling high volumes.
