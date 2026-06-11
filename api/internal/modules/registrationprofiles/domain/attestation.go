package domain

import (
	"time"

	"github.com/google/uuid"
)

type CompletionAttestation struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	BranchID      uuid.UUID
	ChildID       uuid.UUID

	ConsentRecordID        *uuid.UUID
	ProfileUpdatedAt       time.Time
	OfficeChecklistUpdatedAt time.Time

	AttestedByUserID       uuid.UUID
	AttestedByMembershipID uuid.UUID
	AttestedAt             time.Time

	RequestID *string

	CreatedAt time.Time
}

type WorkflowStatus struct {
	ChildSummary   ChildSummary `json:"child_summary"`

	ProfileCompleteness    Completeness         `json:"profile_completeness"`
	OfficeCompleteness     OfficeUseCompleteness `json:"office_completeness"`
	ConsentCompleteness    ConsentCompleteness   `json:"consent_completeness"`

	CurrentConsentRecord *ConsentRecord        `json:"current_consent_record"`
	LatestAttestation    *CompletionAttestation `json:"latest_attestation"`

	CanMarkComplete   bool `json:"can_mark_complete"`
	IsReviewedComplete bool `json:"is_reviewed_complete"`
	NeedsReview       bool `json:"needs_review"`

	MissingGroups []string `json:"missing_groups,omitempty"`
}

func ComputeWorkflowStatus(
	child ChildSummary,
	profileCompleteness Completeness,
	officeCompleteness OfficeUseCompleteness,
	consentCompleteness ConsentCompleteness,
	currentConsent *ConsentRecord,
	latestAttestation *CompletionAttestation,
	profileUpdatedAt *time.Time,
	officeUpdatedAt *time.Time,
) WorkflowStatus {
	status := WorkflowStatus{
		ChildSummary:        child,
		ProfileCompleteness:  profileCompleteness,
		OfficeCompleteness:   officeCompleteness,
		ConsentCompleteness:  consentCompleteness,
		CurrentConsentRecord: currentConsent,
		LatestAttestation:    latestAttestation,
	}

	missing := make([]string, 0)
	if !profileCompleteness.IsComplete {
		missing = append(missing, "registration_profile")
	}
	if !officeCompleteness.IsComplete {
		missing = append(missing, "office_use_checklist")
	}
	if !consentCompleteness.IsComplete {
		missing = append(missing, "consent_records")
	}

	status.CanMarkComplete = len(missing) == 0
	status.MissingGroups = missing

	if status.CanMarkComplete && latestAttestation != nil && currentConsent != nil {
		snapshotsMatch := latestAttestation.ConsentRecordID != nil &&
			*latestAttestation.ConsentRecordID == currentConsent.ID &&
			profileUpdatedAt != nil && latestAttestation.ProfileUpdatedAt.Equal(*profileUpdatedAt) &&
			officeUpdatedAt != nil && latestAttestation.OfficeChecklistUpdatedAt.Equal(*officeUpdatedAt)

		status.IsReviewedComplete = snapshotsMatch
	}

	status.NeedsReview = latestAttestation != nil && !status.IsReviewedComplete

	return status
}
