# Two-hop parent access authorization model for MVP

For month 1, parent invoice visibility is authorized through a two-hop model: an active parent membership must map to a guardian record, and that guardian must have an active guardian-child link to the child being accessed. We chose this over email-matching or direct user-child links because it keeps authorization membership-scoped, preserves guardian contact/login separation, supports immediate access revocation via link end-dating, and avoids ambiguous identity coupling across tenants and branches.
