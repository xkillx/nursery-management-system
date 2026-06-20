# Session templates: named, per-site, reusable weeks of booked sessions

## Context

The booking pattern capture flow (`POST /children/:child_id/booking-patterns`)
already exists and the registration wizard now calls it after `POST /children`.
In practice, many new children will share a common week of booked sessions —
for example, a "Standard 3-day Full" (Mon / Wed / Fri Full Day 08:00–18:00) or
a "Standard 5-day AM" (Mon–Fri Morning 08:00–13:00). Without reusable
templates, every registration re-types the same per-day rows.

We considered two shapes for the reusable concept:

1. **A reference data row per site that the booking pattern points at** —
   `child_booking_patterns` gains a `template_id` foreign key, and the
   daily entries live in `session_template_entries`. The template is
   *referenced* by every child pattern created from it.
2. **A reference data row per site whose entries are copied into the
   booking pattern at creation time** — no `template_id` foreign key on
   `child_booking_patterns`. The template is *consulted* at create time
   and the resulting booking pattern is a standalone, independent row.

## Decision

We chose **(2)**: a child booking pattern is created by *copying* the
template's entries. Editing or archiving a template never alters
historical booking patterns. The DB enforces this by simply having no
foreign key from `child_booking_pattern_entries` (or
`child_booking_patterns`) to `session_templates`.

The new `session_templates` table is a per-site bounded context: a
named, soft-deletable row plus a 1:many `session_template_entries` table
holding `day_of_week` + `session_type_id`. A unique-name-within-site
partial index (`is_active = true`) prevents two active templates from
sharing a name; an archived template frees its name.

## Rationale

- **History integrity over live update.** A booking pattern is an
  effective-dated history (one open + N closed). If templates were
  *referenced*, editing a template would either (a) re-shape every
  historical pattern that points at it — silently rewriting
  past-as-promised — or (b) require us to version templates and choose
  which version a pattern pins to. Both are worse than the
  copy-at-instantiation model. The booking-pattern history model
  already exists (`is_current` partial unique + adjacent close on
  create); adding a template *reference* would require a parallel
  versioning system.
- **The reuse benefit is at the moment of creation.** Templates exist
  to speed up *new* registrations. After a child is created, the
  pattern is independent; subsequent edits to that pattern are made
  via the existing `PATCH /children/:child_id/booking-patterns/:id`
  endpoint and do not touch the template.
- **No new audit complexity.** Editing a template writes a
  `session_template_updated` audit row only; existing
  `child_booking_pattern_*` audits are unaffected. If templates were
  referenced, every pattern change that follows a template edit would
  have to be reasoned about — do we re-write the audit? Skip it?
  Annotate the pattern as "created from template v3"?
- **Templates can be deleted freely.** A common
  sessions-templates-management UX is "archive then delete" — but
  even hard-deleting a template (not currently exposed) cannot
  retroactively change a child's pattern. With copy semantics, the
  template's lifecycle is fully decoupled from any child.
- **Same authority model as `session_types`.** Manager + owner write;
  practitioner read. Mirrors the existing reference-data authority
  pattern in `CONTEXT.md`.

## Trade-offs

- **Editing a template does not propagate.** If a nursery changes its
  "Standard 3-day Full" definition from Full Day to Half Day, all
  children *already created* from that template keep the Full Day
  plan. To update, managers edit each child's pattern individually.
  This is the same trade-off as the existing `session_types` model
  (renaming or rescheduling a session type does not rewrite historical
  booking patterns). The decision is consistent.
- **The new bounded context adds a small surface area.** One new table
  pair, six new endpoints, one new module in the API, one new
  one-page CRUD UI. Documented and audited (`session_template_*`).
- **Templates are not billable; they are an input to the
  booking-pattern capture step, not a substitute for it.** A child
  without a pattern is still surfaced as "incomplete: no session
  pattern" on the child list and detail.

## Consequences

- A new `session_templates` bounded context is added in the API
  (`api/internal/modules/sessiontemplates/`), the DB (migration
  `000005`), the frontend (one CRUD page, one new API service), and
  the registration wizard (a new "Session pattern" step that
  optionally seeds entries from a template before submission).
- The standalone `:childId/booking-pattern` page remains reachable
  for managers who want to capture a pattern outside the wizard
  (e.g., legacy child without a pattern).
- The wizard's Review step does the two-step submit
  (`POST /children` then `POST /children/:id/booking-patterns`).
  Partial failure (pattern POST fails after child POST succeeds) shows
  a Retry button and a "Save child without pattern" fallback that
  routes to the child detail screen.
- Templates **copy** their entries at instantiation; the DB enforces
  the absence of a reverse FK by design (no `template_id` on
  `child_booking_pattern_entries`).

## Status

Accepted. Implemented in PR `feat/session-pattern-at-registration`
(2026-06-21). Future iterations may revisit this if a use case for
"update all future patterns created from a template" emerges — that
would be a separate, additive feature (e.g., a `template_version_id`
pointer or a "rebroadcast" action) rather than a breaking change.
