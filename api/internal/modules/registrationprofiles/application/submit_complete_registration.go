package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type SubmitCompleteRegistration struct {
	profileRepo  domain.Repository
	consentRepo  domain.ConsentRepository
	childCreator domain.ChildCreator
	audit        *audit.Writer
	txMgr        *transaction.Manager
}

func NewSubmitCompleteRegistration(
	profileRepo domain.Repository,
	consentRepo domain.ConsentRepository,
	childCreator domain.ChildCreator,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
) *SubmitCompleteRegistration {
	return &SubmitCompleteRegistration{
		profileRepo:  profileRepo,
		consentRepo:  consentRepo,
		childCreator: childCreator,
		audit:        auditWriter,
		txMgr:        txMgr,
	}
}

type SubmitCompleteRegistrationResult struct {
	ChildID    uuid.UUID
	FirstName  string
	MiddleName *string
	LastName   *string
	StartDate  string
}

func (uc *SubmitCompleteRegistration) Execute(ctx context.Context, actor tenant.ActorContext, input domain.CompleteRegistrationInput) (*SubmitCompleteRegistrationResult, error) {
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	var result *SubmitCompleteRegistrationResult

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		childInfo := domain.ChildInfo{
			FirstName:   strings.TrimSpace(input.Child.FirstName),
			MiddleName:  strings.TrimSpace(input.Child.MiddleName),
			LastName:    strings.TrimSpace(input.Child.LastName),
			DateOfBirth: mustParseDate(input.Child.DateOfBirth),
			StartDate:   mustParseDate(input.Child.StartDate),
			Notes:       input.Child.Notes,
		}

		child, err := uc.childCreator.CreateChild(ctx, tx, childInfo, actor.TenantID, actor.BranchID)
		if err != nil {
			return fmt.Errorf("registration.submit.create_child: %w", err)
		}

		profile := input.ToProfile(actor.TenantID, actor.BranchID, child.ID)
		profile.ID = uid.NewUUID()

		createdProfile, err := uc.profileRepo.Create(ctx, tx, profile)
		if err != nil {
			return fmt.Errorf("registration.submit.create_profile: %w", err)
		}

		contacts := buildContactEntries(createdProfile.ID, child.ID, input.Profile, actor)
		if len(contacts) > 0 {
			contactTypes := []domain.ContactType{
				domain.ContactTypeParentCarer,
				domain.ContactTypeEmergencyContact,
				domain.ContactTypeAuthorisedCollector,
			}
			if err := uc.profileRepo.ReplaceContactsForTypes(ctx, tx, createdProfile.ID, contactTypes, contacts); err != nil {
				return fmt.Errorf("registration.submit.replace_contacts: %w", err)
			}
		}

		if input.CollectionPassword != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(input.CollectionPassword), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("registration.submit.hash_password: %w", err)
			}
			if err := uc.profileRepo.SetCollectionPassword(ctx, tx, actor.TenantID, actor.BranchID, child.ID, string(hash), time.Now().UTC(), actor.UserID, actor.MembershipID); err != nil {
				return fmt.Errorf("registration.submit.set_collection_password: %w", err)
			}
		}

		version, err := uc.consentRepo.GetCurrentVersion(ctx, actor.TenantID, actor.BranchID, child.ID)
		if err != nil {
			version = 0
		}

		consentRecord := buildConsentRecord(child.ID, input.Consents, version, actor)
		if err := uc.consentRepo.CreateConsentRecord(ctx, tx, consentRecord); err != nil {
			return fmt.Errorf("registration.submit.create_consent: %w", err)
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_created",
			EntityType: "child",
			EntityID:   child.ID,
			Details: map[string]any{
				"first_name":  child.FirstName,
				"middle_name": child.MiddleName,
				"last_name":   child.LastName,
			},
		}); err != nil {
			return fmt.Errorf("registration.submit.audit_child_creation: %w", err)
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "registration_completion_attested",
			EntityType: "child",
			EntityID:   child.ID,
			Details: map[string]any{
				"consent_record_id":  consentRecord.ID.String(),
				"completeness_state": "complete",
			},
		}); err != nil {
			return fmt.Errorf("registration.submit.audit_attestation: %w", err)
		}

		result = &SubmitCompleteRegistrationResult{
			ChildID:    child.ID,
			FirstName:  child.FirstName,
			MiddleName: child.MiddleName,
			LastName:   child.LastName,
			StartDate:  child.StartDate.Format("2006-01-02"),
		}
		return nil
	})
	if err != nil {
		return nil, domainerrors.Internal(err)
	}

	return result, nil
}

func (uc *SubmitCompleteRegistration) validateInput(input domain.CompleteRegistrationInput) error {
	var missing []string

	if strings.TrimSpace(input.Child.FirstName) == "" {
		missing = append(missing, "first_name")
	}
	if strings.TrimSpace(input.Child.DateOfBirth) == "" {
		missing = append(missing, "date_of_birth")
	}
	if strings.TrimSpace(input.Child.StartDate) == "" {
		missing = append(missing, "start_date")
	}
	if !input.Consents.PaperFormOnFile {
		missing = append(missing, "consents.paper_form_on_file")
	}
	if !input.Consents.SafeguardingReportingAcknowledgement {
		missing = append(missing, "consents.safeguarding_reporting_acknowledgement")
	}

	if len(missing) > 0 {
		return domainerrors.Validation(
			fmt.Sprintf("Missing required fields: %s", strings.Join(missing, ", ")),
			strings.Join(missing, ","),
		)
	}
	return nil
}

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}

func buildConsentRecord(childID uuid.UUID, ci domain.ConsentInput, version int, actor tenant.ActorContext) *domain.ConsentRecord {
	return &domain.ConsentRecord{
		ID:       uid.NewUUID(),
		TenantID: actor.TenantID,
		BranchID: actor.BranchID,
		ChildID:  childID,
		Version:  version + 1,
		Source:   domain.ConsentSourcePaperForm,

		PaperFormOnFile: ci.PaperFormOnFile,

		UrgentMedicalTreatment:               ci.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions:     ci.UrgentMedicalTreatmentExceptions,
		Plasters:                             ci.Plasters,
		SafeguardingReportingAcknowledgement: ci.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:            ci.InformationSharingConsent,
		GDPRDataProcessingConsent:            ci.GDPRDataProcessingConsent,
		AreaSENCOLiaison:                     ci.AreaSENCOLiaison,
		HealthVisitorLiaison:                 ci.HealthVisitorLiaison,
		TransitionDocuments:                  ci.TransitionDocuments,
		LocalOutings:                         ci.LocalOutings,
		FacePainting:                         ci.FacePainting,
		ParentSuppliedSunCream:               ci.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:             ci.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:             ci.DevelopmentProfilePhotos,
		NurseryDisplayBoards:                 ci.NurseryDisplayBoards,
		PromotionalLiterature:                ci.PromotionalLiterature,
		NurseryWebsite:                       ci.NurseryWebsite,
		StaffStudentCoursework:               ci.StaffStudentCoursework,
		SocialMedia:                          ci.SocialMedia,
		SocialMediaChannelNotes:              ci.SocialMediaChannelNotes,

		NotesExceptions: ci.NotesExceptions,

		EnteredByUserID:       actor.UserID,
		EnteredByMembershipID: actor.MembershipID,
	}
}

func buildContactEntries(profileID, childID uuid.UUID, psi domain.ProfileSectionsInput, actor tenant.ActorContext) []domain.ContactEntry {
	var entries []domain.ContactEntry
	order := 0

	for _, src := range psi.ParentCarers {
		entries = append(entries, makeContactEntry(profileID, childID, domain.ContactTypeParentCarer, order, src, actor))
		order++
	}
	for _, src := range psi.EmergencyContacts {
		entries = append(entries, makeContactEntry(profileID, childID, domain.ContactTypeEmergencyContact, order, src, actor))
		order++
	}
	for _, src := range psi.AuthorisedCollectors {
		entries = append(entries, makeContactEntry(profileID, childID, domain.ContactTypeAuthorisedCollector, order, src, actor))
		order++
	}

	return entries
}

func makeContactEntry(profileID, childID uuid.UUID, contactType domain.ContactType, order int, src domain.ContactEntryInput, actor tenant.ActorContext) domain.ContactEntry {
	e := domain.ContactEntry{
		ID:                        uid.NewUUID(),
		TenantID:                  actor.TenantID,
		BranchID:                  actor.BranchID,
		ProfileID:                 profileID,
		ChildID:                   childID,
		ContactType:               contactType,
		SortOrder:                 order,
		FullName:                  src.FullName,
		RelationshipToChild:       src.RelationshipToChild,
		Telephone:                 src.Telephone,
		Email:                     src.Email,
		HasParentalResponsibility: src.HasParentalResponsibility,
	}
	if src.Address != nil {
		e.Address = src.Address
	} else {
		e.Address = map[string]any{}
	}
	if src.WorkAddress != nil {
		e.WorkAddress = src.WorkAddress
	} else {
		e.WorkAddress = map[string]any{}
	}
	return e
}
