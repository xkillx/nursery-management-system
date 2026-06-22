# Atomic child-with-pattern wizard submit

## Context

The registration wizard at `/staff/manager/children/new` was shortened from
six steps to four in `docs/plans/shorten-registration-wizard-plan.md`, which
removed the in-wizard Session Pattern step and the Review step. After that
change a child created via the wizard intentionally had no booking pattern;
the manager was routed to the child detail page with an "Incomplete: no
session pattern" badge and a CTA to the standalone
`/staff/manager/children/:childId/booking-pattern` page. That plan deviated
from `CONTEXT.md`'s "Booking Pattern Enrollment Independence" entry (which
still required a non-empty pattern at wizard creation) but never updated the
glossary.

Product feedback is that managers forget to add the pattern after create,
which leaves children in the "incomplete: no session pattern" state and
blocks downstream billing/attendance readiness. The product decision is to
re-introduce Session Pattern as a mandatory step 5 of 5 in the wizard, with
no "save without pattern" fallback. A strictly mandatory step raises the
stakes on the failure path: the previous wizard used a two-POST submit
(`POST /children` then `POST /children/:id/booking-patterns`) with a Retry
button and a "Save child without pattern" fallback if the pattern POST
failed after the child POST succeeded. With strict mandatory semantics,
that fallback is gone, so the two-POST path would silently produce an
incomplete child on a transient pattern-POST failure while the manager
believes registration failed — the worst of both worlds.

We considered three failure-path shapes:

1. **Two-POST with in-step retry, no escape.** Child is created first; if
   the pattern POST fails the wizard stays on step 5 showing "retry". The
   child already exists in the DB but the manager is not told that. A
   manager who closes the tab leaves an orphan incomplete child.
2. **Two-POST, retry-then-hard-error to standalone page.** Child is
   created; on persistent pattern-POST failure, route the manager to the
   standalone pattern page. This is a fallback by another name and
   produces the exact "incomplete child from wizard" state the strict
   mandatory decision rejects.
3. **Atomic single-transaction create.** A new backend path inserts the
   child, contacts, and booking-pattern entries in one database
   transaction. Either the whole registration succeeds or nothing is
   created; the manager sees one error and clicks "Create child" again.

## Decision

We chose **(3): atomic single-transaction create**. The wizard's submit
gesture calls `POST /children` with an optional `booking_pattern` payload.
When the payload is present, the server commits the child identity,
profile, health, safeguarding, contacts, consent, funding, collection
settings, room assignment, billing profile, and booking pattern in one
`ExecTx`. A domain error anywhere in the transaction rolls back the entire
registration and surfaces a single error to the manager, who stays on
step 5 and retries. No incomplete child is ever produced from the wizard.

The implementation extends the existing
`CreateChildWithFullProfile.Execute` use case (which already wraps
everything in one `ExecTx`) with an optional `BookingPattern` input field.
The booking-pattern validation and insertion logic is extracted from
`CreateBookingPattern.Execute` into tx-aware helpers
(`resolveBookingPatternEntries` and `createBookingPatternInTx`) so the
standalone `POST /children/:child_id/booking-patterns` endpoint continues
to work unchanged and shares the same validation code path. The
`SessionTypeLookup` adapter is injected into
`CreateChildWithFullProfile` so session-type scope/active checks happen
inside the same use case.

The `effective_from` of the wizard-created pattern defaults to the child's
`start_date` (so planned attendance and the enrollment window agree) and
may be overridden by the manager subject to the existing "today or later"
rule. The standalone pattern page and its per-child endpoints are
unchanged; they remain the path for adding or changing a pattern after
the child exists (including for legacy children created before this
change).

## Rationale

- **The product promise must hold in the failure case too.** "Strictly
  mandatory, no fallback" is only true if a transient API failure cannot
  produce an incomplete child. Options (1) and (2) both produce incomplete
  children on failure; (1) hides them from the manager and (2) routes the
  manager away from them. Only (3) makes the strict-mandatory promise
  actually hold.
- **The existing use case is already transactional.**
  `CreateChildWithFullProfile.Execute` already wraps child + contacts +
  consent + funding + room + billing profile in one `ExecTx`; adding the
  booking pattern to the same transaction is a small, local change. The
  repo methods the booking-pattern path needs (`InsertPattern`,
  `GetCurrentOpenByChild`, `CloseCurrentPattern`) already accept a
  `pgx.Tx` argument.
- **No new DB migration.** The `child_booking_patterns` and
  `child_booking_pattern_entries` tables already exist from ADR-0009. The
  atomic path inserts into them via the existing repo methods; no schema
  change is required.
- **No new public API surface.** `POST /children` already exists; the
  atomic path is an optional field on the existing request body. The
  standalone `POST /children/:id/booking-patterns` endpoint is preserved
  unchanged so the standalone page and any other callers are unaffected.
- **Validation reuse.** Extracting the booking-pattern validation into
  shared helpers means the wizard and the standalone page enforce the
  same `booking_pattern_backdated`, `booking_pattern_duplicate_entry`,
  `booking_pattern_overlap`, and `session_type_not_in_branch` /
  `session_type_archived` rules from a single code path. No second
  implementation of the rules is introduced.

## Trade-offs

- **`CONTEXT.md`'s "Booking Pattern Creation Endpoint" entry is
  rewritten.** The previous two-POST + Retry + "Save child without
  pattern" semantics are removed and replaced with the atomic
  single-POST semantics. The "no single-transaction" note is removed.
  This is a documented supersession, not a silent contradiction.
- **`docs/plans/shorten-registration-wizard-plan.md` is superseded.** A
  future reader of that plan will need this ADR to understand why the
  wizard is back to five steps. The ADR is linked from the plan's
  replacement.
- **The wizard's `CreateChildWithFullProfile` use case now depends on
  `SessionTypeLookup`.** This adds one new constructor argument to the
  use case and one new wiring line in `bootstrap.go`. The adapter already
  exists (`sessionTypeLookupAdapter` in `adapters.go`).
- **The atomic transaction holds a DB transaction open across child +
  contacts + pattern inserts.** In practice this is the same duration as
  the existing child + contacts + consent + funding + room + billing
  transaction plus one pattern insert (a handful of rows). No measurable
  latency impact is expected; if one appears, the fallback is to move
  only the pattern insert out of the transaction (back to two-POST), but
  that reintroduces the incomplete-child failure mode and is rejected for
  this release.
- **Stale drafts in `localStorage`.** A manager who started a
  registration under the four-step wizard may have a draft with no
  `step5` data. On restore, step 5 is shown empty and the manager fills
  it in before "Create child" is enabled. No data loss; the step 1-4
  draft data is preserved.

## Consequences

- The wizard is five steps: Child Details, Medical & Health, Contacts &
  Security, Permissions & Consents, Session Pattern. "Create child" is
  the primary button at the bottom of step 5.
- `POST /children` accepts an optional `booking_pattern` field. When
  present and valid, the created child has `hasBookingPattern === true`
  on the children directory and child detail pages. When absent, the
  endpoint behaves exactly as before (used by the edit flow and any
  non-wizard callers).
- The "Incomplete: no session pattern" badge remains on the children
  directory and child detail for children created before this change or
  via paths that do not set a pattern; it is no longer produced by the
  wizard.
- The standalone `/staff/manager/children/:childId/booking-pattern` page
  is unchanged and remains the canonical path for editing a pattern
  after creation.
- `CONTEXT.md` is updated inline: "Session Pattern (user-facing label)",
  "Booking Pattern Enrollment Independence", and "Booking Pattern
  Creation Endpoint" are rewritten, and a new "Atomic Child-with-Pattern
  Creation" entry is added.

## Status

Accepted. Supersedes the wizard-shape decision in
`docs/plans/shorten-registration-wizard-plan.md` (which is preserved as a
historical record). ADR-0009 (session templates — copy-at-instantiation)
is unchanged; the wizard step 5 still copies template entries into the
pattern at creation time.
