package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateConsentParams struct {
	SignerName       string  `json:"signer_name"`
	SignedDate       string  `json:"signed_date"`
	PaperFormOnFile  bool    `json:"paper_form_on_file"`

	UrgentMedicalTreatment         bool   `json:"urgent_medical_treatment"`
	UrgentMedicalTreatmentExceptions *string `json:"urgent_medical_treatment_exceptions,omitempty"`
	Plasters                       bool   `json:"plasters"`
	SafeguardingReportingAcknowledgement bool `json:"safeguarding_reporting_acknowledgement"`
	AreaSENCOLiaison               bool   `json:"area_senco_liaison"`
	HealthVisitorLiaison           bool   `json:"health_visitor_liaison"`
	TransitionDocuments            bool   `json:"transition_documents"`
	LocalOutings                   bool   `json:"local_outings"`
	FacePainting                   bool   `json:"face_painting"`
	ParentSuppliedSunCream         bool   `json:"parent_supplied_sun_cream"`
	ParentSuppliedNappyCream       bool   `json:"parent_supplied_nappy_cream"`
	DevelopmentProfilePhotos       bool   `json:"development_profile_photos"`
	NurseryDisplayBoards           bool   `json:"nursery_display_boards"`
	PromotionalLiterature          bool   `json:"promotional_literature"`
	NurseryWebsite                 bool   `json:"nursery_website"`
	StaffStudentCoursework         bool   `json:"staff_student_coursework"`
	SocialMedia                    bool   `json:"social_media"`
	SocialMediaChannelNotes        *string `json:"social_media_channel_notes,omitempty"`

	NotesExceptions *string `json:"notes_exceptions,omitempty"`
}

type CreateConsent struct {
	repo    domain.ConsentRepository
	audit   *audit.Writer
	pool    *pgxpool.Pool
	txMgr   *transaction.Manager
}

func NewCreateConsent(repo domain.ConsentRepository, auditWriter *audit.Writer, txMgr *transaction.Manager) *CreateConsent {
	return &CreateConsent{repo: repo, audit: auditWriter, pool: nil, txMgr: txMgr}
}

func (uc *CreateConsent) Execute(ctx context.Context, actor tenant.ActorContext, childID string, params CreateConsentParams) (domain.ConsentRecord, error) {
	cid, err := uuid.Parse(childID)
	if err != nil {
		return domain.ConsentRecord{}, domainerrors.Validation("Invalid request payload.", "child_id")
	}

	if strings.TrimSpace(params.SignerName) == "" {
		return domain.ConsentRecord{}, domainerrors.Validation("Invalid request payload.", "signer_name")
	}

	signedDate, err := time.Parse("2006-01-02", strings.TrimSpace(params.SignedDate))
	if err != nil {
		return domain.ConsentRecord{}, domainerrors.Validation("Invalid request payload.", "signed_date")
	}

	if !params.PaperFormOnFile {
		return domain.ConsentRecord{}, domainerrors.Validation("Paper form must be marked as on file.", "paper_form_on_file")
	}

	if !params.SafeguardingReportingAcknowledgement {
		return domain.ConsentRecord{}, domainerrors.Validation("Safeguarding/reporting acknowledgement is required.", "safeguarding_reporting_acknowledgement")
	}

	record := domain.ConsentRecord{
		ID:      uid.NewUUID(),
		TenantID: actor.TenantID,
		BranchID: actor.BranchID,
		ChildID:  cid,
		Source:   domain.ConsentSourcePaperForm,
		SignerName:       strings.TrimSpace(params.SignerName),
		SignedDate:       signedDate,
		PaperFormOnFile:  params.PaperFormOnFile,
		UrgentMedicalTreatment:         params.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions: params.UrgentMedicalTreatmentExceptions,
		Plasters:                       params.Plasters,
		SafeguardingReportingAcknowledgement: params.SafeguardingReportingAcknowledgement,
		AreaSENCOLiaison:               params.AreaSENCOLiaison,
		HealthVisitorLiaison:           params.HealthVisitorLiaison,
		TransitionDocuments:            params.TransitionDocuments,
		LocalOutings:                   params.LocalOutings,
		FacePainting:                   params.FacePainting,
		ParentSuppliedSunCream:         params.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:       params.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:       params.DevelopmentProfilePhotos,
		NurseryDisplayBoards:           params.NurseryDisplayBoards,
		PromotionalLiterature:          params.PromotionalLiterature,
		NurseryWebsite:                 params.NurseryWebsite,
		StaffStudentCoursework:         params.StaffStudentCoursework,
		SocialMedia:                    params.SocialMedia,
		SocialMediaChannelNotes:        params.SocialMediaChannelNotes,
		NotesExceptions:                params.NotesExceptions,
		EnteredByUserID:       actor.UserID,
		EnteredByMembershipID: actor.MembershipID,
	}

	err = uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		currentVersion, verErr := uc.repo.GetCurrentVersion(ctx, actor.TenantID, actor.BranchID, cid)
		if verErr != nil {
			return fmt.Errorf("get current consent version: %w", verErr)
		}
		record.Version = currentVersion + 1

		if err := uc.repo.CreateConsentRecord(ctx, tx, &record); err != nil {
			return fmt.Errorf("create consent record: %w", err)
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "registration_consent_record_created",
			EntityType: "child",
			EntityID:   cid,
			Details: map[string]any{
				"consent_record_id": record.ID.String(),
				"version":           record.Version,
			},
		}); err != nil {
			return fmt.Errorf("audit consent record created: %w", err)
		}

		return nil
	})
	if err != nil {
		return domain.ConsentRecord{}, domainerrors.Internal(err)
	}

	return record, nil
}
