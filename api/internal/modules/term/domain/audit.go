package domain

// Audit event names. Follow the existing convention in api/internal/platform/audit
// (snake_case, past-tense verb).
const (
	AuditTermCreated                 = "term_created"
	AuditTermScheduleChangeRequested = "term_schedule_change_requested"
	AuditTermScheduleChangeApproved  = "term_schedule_change_approved"
	AuditTermScheduleChangeRejected  = "term_schedule_change_rejected"
	AuditTermTerminated              = "term_terminated"
	AuditTermRenewalCreated          = "term_renewal_created"
	AuditTermStatusTransitioned      = "term_status_transitioned"
	AuditEntityTerm                  = "term"
	AuditEntityTermScheduleChange    = "term_schedule_change"
)
