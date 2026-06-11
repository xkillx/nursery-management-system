package httpregistrationprofile

import (
	"time"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
)

type workflowStatusResponse struct {
	Child         childSummaryResponse           `json:"child"`
	Profile       *completenessResponse           `json:"profile_completeness"`
	Office        officeCompletenessResponse      `json:"office_completeness"`
	Consent       consentCompletenessResponse     `json:"consent_completeness"`
	CurrentConsent *consentRecordResponse          `json:"current_consent_record,omitempty"`
	Attestation   *attestationResponse            `json:"latest_attestation,omitempty"`

	CanMarkComplete   bool     `json:"can_mark_complete"`
	IsReviewedComplete bool   `json:"is_reviewed_complete"`
	NeedsReview       bool     `json:"needs_review"`
	MissingGroups     []string `json:"missing_groups,omitempty"`
}

type attestationResponse struct {
	ID            string `json:"id"`
	ConsentRecordID *string `json:"consent_record_id,omitempty"`
	AttestedByUserID       string `json:"attested_by_user_id"`
	AttestedByMembershipID string `json:"attested_by_membership_id"`
	AttestedAt    string `json:"attested_at"`
}

func toWorkflowStatusResponse(status domain.WorkflowStatus) workflowStatusResponse {
	resp := workflowStatusResponse{
		Child: childSummaryResponse{
			ID:          status.ChildSummary.ID.String(),
			FullName:    status.ChildSummary.FullName,
			DateOfBirth: status.ChildSummary.DateOfBirth.Format("2006-01-02"),
		},
		Profile:           toCompletenessResponse(status.ProfileCompleteness),
		Office:            toOfficeCompletenessResponse(status.OfficeCompleteness),
		Consent:           toConsentCompletenessResponse(status.ConsentCompleteness),
		CanMarkComplete:   status.CanMarkComplete,
		IsReviewedComplete: status.IsReviewedComplete,
		NeedsReview:       status.NeedsReview,
		MissingGroups:     status.MissingGroups,
	}

	if status.CurrentConsentRecord != nil {
		cr := toConsentRecordResponse(*status.CurrentConsentRecord)
		resp.CurrentConsent = &cr
	}

	if status.LatestAttestation != nil {
		resp.Attestation = toAttestationResponse(*status.LatestAttestation)
	}

	return resp
}

func toAttestationResponse(a domain.CompletionAttestation) *attestationResponse {
	resp := &attestationResponse{
		ID:                     a.ID.String(),
		AttestedByUserID:       a.AttestedByUserID.String(),
		AttestedByMembershipID: a.AttestedByMembershipID.String(),
		AttestedAt:             a.AttestedAt.UTC().Format(time.RFC3339),
	}
	if a.ConsentRecordID != nil {
		s := a.ConsentRecordID.String()
		resp.ConsentRecordID = &s
	}
	return resp
}
