# One-hop parent access authorization model

For month 1, parent invoice visibility is authorized through a one-hop
model: an active parent membership must have an active
`parent_membership_children` row to the child being accessed. The
intermediate `guardians` and `guardian_child_links` tables no longer
exist; the parent's relationship to a child is recorded directly on the
parent-membership row.

## Rationale

The one-hop model replaces the two-hop model documented in the
(now-superseded) ADR-0002. We chose to collapse the chain because:

- **Simpler authorization queries.** Parent invoice reads and
  payment-attempt writes join `parent_membership_children pmc` once
  against `invoices.child_id`, instead of joining `parent_membership_guardians
  pmg` and then `guardian_child_links gcl`. The two-hop join was the
  source of the original read-cost concerns; the one-hop model makes
  the read path a single indexed join.
- **Direct manager control.** Ending a parent's access to a child is a
  single `end` action on a `parent_membership_children` row, with the
  same `lifecycle_reason_code` vocabulary the rest of the system uses
  (e.g., `access_revoked`, `safeguarding_direction`, `other`). There is
  no separate guardian-deactivation cascade to reason about.
- **Identity is a property of the parent membership.** The parent user
  account is created by manager invitation only; the
  `parent_membership` row carries role and active state; the
  `parent_membership_children` row is the only thing that authorises
  that membership to act on a child. There is no separate
  `guardians` entity whose deactivation status would have to be
  cross-referenced at request time.
- **No ambiguous identity coupling.** The original two-hop model
  conflated "this person is a contact for a child" with "this person
  has parent-portal access for a child." The one-hop model keeps the
  first concern in `child_contacts` (a `parent_carer` row) and the
  second in `parent_membership_children`. Conflating them had
  repeatedly produced UX confusion: contacts visible on the child
  record were not the same set as users with portal access, but staff
  treated them as interchangeable.

## Trade-offs

- **Loss of the "guardian" concept.** A child may have many contacts
  in `child_contacts` (e.g., grandparents, neighbours, emergency
  contacts) but only the rows of type `parent_carer` are required for
  enrollment-completeness. There is no longer a separate record that
  names "the people who are responsible for this child at this
  nursery"; the parent-carer contact set is the source of truth.
- **A parent may map to many children.** A single parent membership
  may have any number of active `parent_membership_children` rows;
  siblings share one parent account. The idempotent create endpoint
  returns the existing active row for a duplicate
  `(membership_id, child_id)` pair.
- **Cascades are still system-driven.** When a `memberships` row is
  ended, all active `parent_membership_children` rows for that
  membership are ended with `ended_reason_code = 'system_cascade'` by
  a database trigger (`cascade_parent_membership_child_end`). This
  mirrors the old behaviour, but the cascade target is now a single
  table.

## Why this is hard to reverse

Every parent-facing query (parent invoice list, parent invoice detail,
parent invoice line items, parent payment attempt authorisation) joins
`parent_membership_children pmc` directly against `invoices.child_id`.
Reversing to a two-hop model would require reintroducing the
`guardians` and `guardian_child_links` tables, backfilling the
mid-table, and rewriting every parent-facing read path. The cost of
that rewrite is real, and the value (separating contact identity
from portal access) is already captured in `child_contacts` under the
new model.

## Surprising without context

A future reader will reasonably ask: "Why doesn't parent access go
through a `guardians` table like every other parent-child relationship
in the system?" The answer is the trade-off above: the
`parent_carer` contact in `child_contacts` is the user-facing
identity, and `parent_membership_children` is the authorization
mechanism. They are decoupled by design.

## Result of a real trade-off

The original two-hop model preserved the option to support a
non-parent guardian (e.g., a social worker) with portal access without
creating a `parent` membership. The one-hop model gives up that
option: anyone with portal access must have a `parent` role
membership. The pilot has no such use case, and the option was not
deemed worth the cost of the two-hop read path.
