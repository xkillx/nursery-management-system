package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type GetConsent struct {
	repo domain.Repository
}

func NewGetConsent(repo domain.Repository) *GetConsent {
	return &GetConsent{repo: repo}
}

func (uc *GetConsent) Execute(ctx context.Context, actor tenant.ActorContext, childID string) (*domain.ChildConsent, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	_, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("check child exists: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}
	consent, found, err := uc.repo.GetConsentByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("get child consent: %w", err))
	}
	if !found {
		return nil, nil
	}
	return consent, nil
}

type UpdateConsent struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewUpdateConsent(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *UpdateConsent {
	return &UpdateConsent{repo: repo, audit: auditWriter, txm: txm}
}

func (uc *UpdateConsent) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in *ChildConsentInput) (*domain.ChildConsent, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	if in == nil {
		return nil, domainerrors.Validation("Invalid request payload.", "body")
	}
	signedDate, err := time.Parse("2006-01-02", in.SignedDate)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "consents.signed_date")
	}

	var result *domain.ChildConsent
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}
		existing, _, eerr := uc.repo.GetConsentByChild(ctx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("get child consent: %w", eerr))
		}
		consent := &domain.ChildConsent{
			TenantID:                             actor.TenantID,
			BranchID:                             actor.BranchID,
			ChildID:                              id,
			UrgentMedicalTreatment:               in.UrgentMedicalTreatment,
			UrgentMedicalTreatmentExceptions:     in.UrgentMedicalTreatmentExceptions,
			Plasters:                             in.Plasters,
			SafeguardingReportingAcknowledgement: in.SafeguardingReportingAcknowledgement,
			InformationSharingConsent:            in.InformationSharingConsent,
			GDPRDataProcessingConsent:            in.GDPRDataProcessingConsent,
			InformationTruthfulnessDeclaration:   in.InformationTruthfulnessDeclaration,
			AreaSENCOLiaison:                     in.AreaSENCOLiaison,
			HealthVisitorLiaison:                 in.HealthVisitorLiaison,
			TransitionDocuments:                  in.TransitionDocuments,
			LocalOutings:                         in.LocalOutings,
			FacePainting:                         in.FacePainting,
			ParentSuppliedSunCream:               in.ParentSuppliedSunCream,
			ParentSuppliedNappyCream:             in.ParentSuppliedNappyCream,
			DevelopmentProfilePhotos:             in.DevelopmentProfilePhotos,
			NurseryDisplayBoards:                 in.NurseryDisplayBoards,
			PromotionalLiterature:                in.PromotionalLiterature,
			NurseryWebsite:                       in.NurseryWebsite,
			StaffStudentCoursework:               in.StaffStudentCoursework,
			SocialMedia:                          in.SocialMedia,
			SocialMediaChannelNotes:              in.SocialMediaChannelNotes,
			NotesExceptions:                      in.NotesExceptions,
			SignerName:                           in.SignerName,
			SignedDate:                           signedDate,
			PaperFormOnFile:                      in.PaperFormOnFile,
			EnteredByUserID:                      actor.UserID,
			EnteredByMembershipID:                actor.MembershipID,
		}
		var saved *domain.ChildConsent
		if existing == nil {
			consent.ID = uid.NewUUID()
			saved, eerr = uc.repo.InsertConsent(ctx, tx, consent)
		} else {
			consent.ID = existing.ID
			saved, eerr = uc.repo.UpdateConsent(ctx, tx, consent)
		}
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("save child consent: %w", eerr))
		}
		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_consent_updated",
			EntityType: "child",
			EntityID:   id,
			Details:    map[string]any{},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_consent_updated: %w", aerr))
		}
		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
