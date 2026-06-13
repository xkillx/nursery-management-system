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
	profileRepo domain.Repository
	consentRepo domain.ConsentRepository
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
	ChildID   uuid.UUID
	FullName  string
	StartDate string
}

func (uc *SubmitCompleteRegistration) Execute(ctx context.Context, actor tenant.ActorContext, input domain.CompleteRegistrationInput) (*SubmitCompleteRegistrationResult, error) {
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	var result *SubmitCompleteRegistrationResult

	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		childInfo := domain.ChildInfo{
			FullName:    input.Child.FullName,
			DateOfBirth: mustParseDate(input.Child.DateOfBirth),
			StartDate:   mustParseDate(input.Child.StartDate),
			Notes:       input.Child.Notes,
		}

		child, err := uc.childCreator.CreateChild(ctx, tx, childInfo, actor.TenantID, actor.BranchID)
		if err != nil {
			return fmt.Errorf("create child: %w", err)
		}

		profile := input.ToProfile(actor.TenantID, actor.BranchID, child.ID)
		profile.ID = uid.NewUUID()

		createdProfile, err := uc.profileRepo.Create(ctx, tx, profile)
		if err != nil {
			return fmt.Errorf("create profile: %w", err)
		}

		contacts := buildContactEntries(createdProfile.ID, child.ID, input.Profile, actor)
		if len(contacts) > 0 {
			contactTypes := []domain.ContactType{
				domain.ContactTypeParentCarer,
				domain.ContactTypeEmergencyContact,
				domain.ContactTypeAuthorisedCollector,
			}
			if err := uc.profileRepo.ReplaceContactsForTypes(ctx, tx, createdProfile.ID, contactTypes, contacts); err != nil {
				return fmt.Errorf("replace contacts: %w", err)
			}
		}

		if input.CollectionPassword != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(input.CollectionPassword), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("hash password: %w", err)
			}
			if err := uc.profileRepo.SetCollectionPassword(ctx, tx, actor.TenantID, actor.BranchID, child.ID, string(hash), time.Now().UTC(), actor.UserID, actor.MembershipID); err != nil {
				return fmt.Errorf("set collection password: %w", err)
			}
		}

		version, err := uc.consentRepo.GetCurrentVersion(ctx, actor.TenantID, actor.BranchID, child.ID)
		if err != nil {
			version = 0
		}

		consentRecord := buildConsentRecord(child.ID, input.Consents, version, actor)
		if err := uc.consentRepo.CreateConsentRecord(ctx, tx, consentRecord); err != nil {
			return fmt.Errorf("create consent: %w", err)
		}

		officeChecklist := buildOfficeChecklist(child.ID, input.OfficeChecklist, actor)
		if _, err := uc.profileRepo.CreateOfficeChecklist(ctx, tx, officeChecklist); err != nil {
			return fmt.Errorf("create office checklist: %w", err)
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_created",
			EntityType: "child",
			EntityID:   child.ID,
			Details: map[string]any{
				"full_name": child.FullName,
			},
		}); err != nil {
			return fmt.Errorf("audit child creation: %w", err)
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
			return fmt.Errorf("audit attestation: %w", err)
		}

		result = &SubmitCompleteRegistrationResult{
			ChildID:   child.ID,
			FullName:  child.FullName,
			StartDate: child.StartDate.Format("2006-01-02"),
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

	if strings.TrimSpace(input.Child.FullName) == "" {
		missing = append(missing, "full_name")
	}
	if strings.TrimSpace(input.Child.DateOfBirth) == "" {
		missing = append(missing, "date_of_birth")
	}
	if strings.TrimSpace(input.Child.StartDate) == "" {
		missing = append(missing, "start_date")
	}
	if strings.TrimSpace(input.Consents.SignerName) == "" {
		missing = append(missing, "consents.signer_name")
	}
	if strings.TrimSpace(input.Consents.SignedDate) == "" {
		missing = append(missing, "consents.signed_date")
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
	signedDate, _ := time.Parse("2006-01-02", ci.SignedDate)

	return &domain.ConsentRecord{
		ID:      uid.NewUUID(),
		TenantID: actor.TenantID,
		BranchID: actor.BranchID,
		ChildID:  childID,
		Version:  version + 1,
		Source:   domain.ConsentSourcePaperForm,

		SignerName:       ci.SignerName,
		SignedDate:       signedDate,
		PaperFormOnFile:  ci.PaperFormOnFile,

		UrgentMedicalTreatment:         ci.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions: ci.UrgentMedicalTreatmentExceptions,
		Plasters:                       ci.Plasters,
			SafeguardingReportingAcknowledgement: ci.SafeguardingReportingAcknowledgement,
			InformationSharingConsent:      ci.InformationSharingConsent,
			GDPRDataProcessingConsent:      ci.GDPRDataProcessingConsent,
		AreaSENCOLiaison:               ci.AreaSENCOLiaison,
		HealthVisitorLiaison:           ci.HealthVisitorLiaison,
		TransitionDocuments:            ci.TransitionDocuments,
		LocalOutings:                   ci.LocalOutings,
		FacePainting:                   ci.FacePainting,
		ParentSuppliedSunCream:         ci.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:       ci.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:       ci.DevelopmentProfilePhotos,
		NurseryDisplayBoards:           ci.NurseryDisplayBoards,
		PromotionalLiterature:          ci.PromotionalLiterature,
		NurseryWebsite:                 ci.NurseryWebsite,
		StaffStudentCoursework:         ci.StaffStudentCoursework,
		SocialMedia:                    ci.SocialMedia,
		SocialMediaChannelNotes:        ci.SocialMediaChannelNotes,

		NotesExceptions: ci.NotesExceptions,

		EnteredByUserID:       actor.UserID,
		EnteredByMembershipID: actor.MembershipID,
	}
}

func buildOfficeChecklist(childID uuid.UUID, oi domain.OfficeUseChecklistInput, actor tenant.ActorContext) *domain.OfficeUseChecklist {
	return &domain.OfficeUseChecklist{
		ID:                                uid.NewUUID(),
		TenantID:                          actor.TenantID,
		BranchID:                          actor.BranchID,
		ChildID:                           childID,

		DepositStatus:                     domain.OfficeCheckStatus(optStr(oi.DepositStatus, "unknown")),
		DepositPaidDate:                   parseOptDate(oi.DepositPaidDate),
		ApplicationDateStatus:             domain.OfficeCheckStatus(optStr(oi.ApplicationDateStatus, "unknown")),
		ApplicationDate:                   parseOptDate(oi.ApplicationDate),
		StartDateStatus:                   domain.OfficeCheckStatus(optStr(oi.StartDateStatus, "unknown")),
		DateLeft:                          parseOptDate(oi.DateLeft),
		SessionsDaysRequestedStatus:       domain.OfficeCheckStatus(optStr(oi.SessionsDaysRequestedStatus, "unknown")),
		SessionsDaysRequested:             oi.SessionsDaysRequested,
		TermTimeOnlySpaceStatus:           domain.TermTimeOnlyStatus(optStr(oi.TermTimeOnlySpaceStatus, "unknown")),
		ContractStatus:                    domain.OfficeCheckStatus(optStr(oi.ContractStatus, "unknown")),
		ContractDate:                      parseOptDate(oi.ContractDate),
		HandbookStatus:                    domain.OfficeCheckStatus(optStr(oi.HandbookStatus, "unknown")),
		HandbookDate:                      parseOptDate(oi.HandbookDate),
		RedBookStatus:                     domain.OfficeCheckStatus(optStr(oi.RedBookStatus, "unknown")),
		RedBookCheckedDate:                parseOptDate(oi.RedBookCheckedDate),
		BirthCertificatePassportStatus:    domain.OfficeCheckStatus(optStr(oi.BirthCertificatePassportStatus, "unknown")),
		BirthCertificatePassportCheckedDate: parseOptDate(oi.BirthCertificatePassportCheckedDate),
		ProofOfAddressStatus:              domain.OfficeCheckStatus(optStr(oi.ProofOfAddressStatus, "unknown")),
		ProofOfAddressCheckedDate:         parseOptDate(oi.ProofOfAddressCheckedDate),
		Notes:                             oi.Notes,
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
		ID:                       uid.NewUUID(),
		TenantID:                 actor.TenantID,
		BranchID:                 actor.BranchID,
		ProfileID:                profileID,
		ChildID:                  childID,
		ContactType:              contactType,
		SortOrder:                order,
		FullName:                 src.FullName,
		RelationshipToChild:      src.RelationshipToChild,
		Telephone:                src.Telephone,
		Email:                    src.Email,
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

func optStr(s *string, fallback string) string {
	if s != nil && *s != "" {
		return *s
	}
	return fallback
}

func parseOptDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}
