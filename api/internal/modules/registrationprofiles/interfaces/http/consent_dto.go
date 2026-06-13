package httpregistrationprofile

import (
	"time"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
)

type consentRecordResponse struct {
	ID        string `json:"id"`
	ChildID   string `json:"child_id"`
	Version   int    `json:"version"`
	Source    string `json:"source"`

	SignerName      string `json:"signer_name"`
	SignedDate      string `json:"signed_date"`
	PaperFormOnFile bool   `json:"paper_form_on_file"`

	UrgentMedicalTreatment         bool    `json:"urgent_medical_treatment"`
	UrgentMedicalTreatmentExceptions *string `json:"urgent_medical_treatment_exceptions,omitempty"`
	Plasters                       bool    `json:"plasters"`
	SafeguardingReportingAcknowledgement bool `json:"safeguarding_reporting_acknowledgement"`
	InformationSharingConsent      bool    `json:"information_sharing_consent"`
	GDPRDataProcessingConsent      bool    `json:"gdpr_data_processing_consent"`
	AreaSENCOLiaison               bool    `json:"area_senco_liaison"`
	HealthVisitorLiaison           bool    `json:"health_visitor_liaison"`
	TransitionDocuments            bool    `json:"transition_documents"`
	LocalOutings                   bool    `json:"local_outings"`
	FacePainting                   bool    `json:"face_painting"`
	ParentSuppliedSunCream         bool    `json:"parent_supplied_sun_cream"`
	ParentSuppliedNappyCream       bool    `json:"parent_supplied_nappy_cream"`
	DevelopmentProfilePhotos       bool    `json:"development_profile_photos"`
	NurseryDisplayBoards           bool    `json:"nursery_display_boards"`
	PromotionalLiterature          bool    `json:"promotional_literature"`
	NurseryWebsite                 bool    `json:"nursery_website"`
	StaffStudentCoursework         bool    `json:"staff_student_coursework"`
	SocialMedia                    bool    `json:"social_media"`
	SocialMediaChannelNotes        *string `json:"social_media_channel_notes,omitempty"`

	NotesExceptions *string `json:"notes_exceptions,omitempty"`

	EnteredByUserID       string `json:"entered_by_user_id"`
	EnteredByMembershipID string `json:"entered_by_membership_id"`
	CreatedAt             string `json:"created_at"`
}

type consentCompletenessResponse struct {
	IsComplete      bool     `json:"is_complete"`
	MissingDecisions []string `json:"missing_decisions,omitempty"`
}

type consentsResponse struct {
	Child        childSummaryResponse         `json:"child"`
	Current      *consentRecordResponse       `json:"current"`
	History      []consentRecordResponse      `json:"history"`
	Completeness consentCompletenessResponse  `json:"completeness"`
}

func toConsentRecordResponse(r domain.ConsentRecord) consentRecordResponse {
	resp := consentRecordResponse{
		ID:        r.ID.String(),
		ChildID:   r.ChildID.String(),
		Version:   r.Version,
		Source:    string(r.Source),
		SignerName:      r.SignerName,
		SignedDate:      r.SignedDate.Format("2006-01-02"),
		PaperFormOnFile:  r.PaperFormOnFile,
		UrgentMedicalTreatment:         r.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions: r.UrgentMedicalTreatmentExceptions,
		Plasters:                       r.Plasters,
		SafeguardingReportingAcknowledgement: r.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:      r.InformationSharingConsent,
		GDPRDataProcessingConsent:      r.GDPRDataProcessingConsent,
		AreaSENCOLiaison:               r.AreaSENCOLiaison,
		HealthVisitorLiaison:           r.HealthVisitorLiaison,
		TransitionDocuments:            r.TransitionDocuments,
		LocalOutings:                   r.LocalOutings,
		FacePainting:                   r.FacePainting,
		ParentSuppliedSunCream:         r.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:       r.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:       r.DevelopmentProfilePhotos,
		NurseryDisplayBoards:           r.NurseryDisplayBoards,
		PromotionalLiterature:          r.PromotionalLiterature,
		NurseryWebsite:                 r.NurseryWebsite,
		StaffStudentCoursework:         r.StaffStudentCoursework,
		SocialMedia:                    r.SocialMedia,
		SocialMediaChannelNotes:        r.SocialMediaChannelNotes,
		NotesExceptions:                r.NotesExceptions,
		EnteredByUserID:       r.EnteredByUserID.String(),
		EnteredByMembershipID: r.EnteredByMembershipID.String(),
		CreatedAt:             r.CreatedAt.UTC().Format(time.RFC3339),
	}
	return resp
}

func toConsentCompletenessResponse(cc domain.ConsentCompleteness) consentCompletenessResponse {
	return consentCompletenessResponse{
		IsComplete:      cc.IsComplete,
		MissingDecisions: cc.MissingDecisions,
	}
}

func toConsentsResponse(child domain.ChildSummary, cwc domain.ConsentWithCompleteness) consentsResponse {
	resp := consentsResponse{
		Child: childSummaryResponse{
			ID:          child.ID.String(),
			FullName:    child.FullName,
			DateOfBirth: child.DateOfBirth.Format("2006-01-02"),
		},
		Completeness: consentCompletenessResponse{
			IsComplete:      cwc.Completeness.IsComplete,
			MissingDecisions: cwc.Completeness.MissingDecisions,
		},
	}
	if cwc.Current != nil {
		cr := toConsentRecordResponse(*cwc.Current)
		resp.Current = &cr
	}
	resp.History = make([]consentRecordResponse, len(cwc.History))
	for i, r := range cwc.History {
		resp.History[i] = toConsentRecordResponse(r)
	}
	return resp
}
