package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateChildIdentityInput struct {
	FirstName   string
	MiddleName  string
	LastName    string
	DateOfBirth string
	StartDate   string
	EndDate     string
	Notes       string
}

type ChildProfileInput struct {
	Sex                          *string
	Religion                     *string
	EthnicOrigin                 *string
	FirstLanguage                *string
	OtherLanguages               *string
	AddressLine1                 *string
	AddressLine2                 *string
	AddressCity                  *string
	AddressPostcode              *string
	HomeTelephone                *string
	DisabilityStatus             string
	DisabilityNotes              *string
	AccessRequirements           *string
	RoutineCareNotes             *string
	GDPRDeclaredByName           *string
	GDPRDeclaredAt               *string
	GDPRDeclarationDate          *string
	RegistrationDate             *string
	DemographicsHomeReviewed     bool
	MedicalDietaryReviewed       bool
	HealthContactsReviewed       bool
	SocialDevelopmentReviewed    bool
	ParentResponsibilityReviewed bool
	EmergencyCollectionReviewed  bool
	RoutineCareReviewed          bool
}

type ChildHealthProfileInput struct {
	MedicalConditionsStatus    string
	MedicalConditionsNotes     *string
	PrescribedMedicationStatus string
	MedicationNotes            *string
	ImmunisationStatus         string
	ImmunisationCountry        *string
	IllnessDiagnosisHistory    *string
	DietaryRequirementsStatus  string
	DietaryRequirementsNotes   *string
	DietarySideEffects         *string
	DoctorName                 *string
	DoctorAddress              *string
	DoctorPhone                *string
	HealthVisitorName          *string
	HealthVisitorAddress       *string
	HealthVisitorPhone         *string
}

type ChildSafeguardingProfileInput struct {
	SocialServicesStatus      string
	SocialServicesNotes       *string
	SocialWorkerName          *string
	SocialWorkerPhone         *string
	SocialWorkerEmail         *string
	ConcernWalking            string
	ConcernSpeechLanguage     string
	ConcernHearing            string
	ConcernSight              string
	ConcernEmotionalWellbeing string
	ConcernBehaviour          string
	ProfessionalReferrals     []domain.ProfessionalReferral
	RestrictedNotes           *string
}

type ChildContactInput struct {
	ContactType               domain.ContactType
	FullName                  string
	RelationshipToChild       *string
	Address                   map[string]any
	Telephone                 *string
	Email                     *string
	WorkAddress               map[string]any
	HasParentalResponsibility *bool
}

type ChildConsentInput struct {
	UrgentMedicalTreatment               bool
	UrgentMedicalTreatmentExceptions     *string
	Plasters                             bool
	SafeguardingReportingAcknowledgement bool
	InformationSharingConsent            bool
	GDPRDataProcessingConsent            bool
	InformationTruthfulnessDeclaration   bool
	AreaSENCOLiaison                     bool
	HealthVisitorLiaison                 bool
	TransitionDocuments                  bool
	LocalOutings                         bool
	FacePainting                         bool
	ParentSuppliedSunCream               bool
	ParentSuppliedNappyCream             bool
	DevelopmentProfilePhotos             bool
	NurseryDisplayBoards                 bool
	PromotionalLiterature                bool
	NurseryWebsite                       bool
	StaffStudentCoursework               bool
	SocialMedia                          bool
	SocialMediaChannelNotes              *string
	NotesExceptions                      *string
	SignerName                           string
	SignedDate                           string
	PaperFormOnFile                      bool
}

type ChildCollectionSettingsInput struct {
	Over18CollectionAcknowledged bool
	Password                     string
	PasswordHint                 string
}

type ChildRoomAssignmentInput struct {
	RoomID    string
	StartDate string
}

type CreateChildFullInput struct {
	Child              CreateChildIdentityInput
	Profile            *ChildProfileInput
	Health             *ChildHealthProfileInput
	Safeguarding       *ChildSafeguardingProfileInput
	Contacts           []ChildContactInput
	Consent            *ChildConsentInput
	Funding            *domain.ChildFundingRecordInput
	CollectionSettings *ChildCollectionSettingsInput
	Room               *ChildRoomAssignmentInput
}

type ChildCreationResult struct {
	ChildID           uuid.UUID
	FirstName         string
	MiddleName        *string
	LastName          *string
	StartDate         string
	TermID            *uuid.UUID
	CreatedSubRecords []string
}

type CreateChildWithFullProfile struct {
	repo          domain.Repository
	audit         *audit.Writer
	txm           TxManager
	sessionLookup SessionTypeLookup
	termCreator   EnrollmentTermCreator
	fundingWriter domain.ChildFundingWriter
	clock         TodayFunc
}

func NewCreateChildWithFullProfile(repo domain.Repository, auditWriter *audit.Writer, txm TxManager, lookup SessionTypeLookup, termCreator EnrollmentTermCreator, fundingWriter domain.ChildFundingWriter, clock TodayFunc) *CreateChildWithFullProfile {
	if clock == nil {
		clock = func() time.Time { return time.Now().UTC() }
	}
	return &CreateChildWithFullProfile{repo: repo, audit: auditWriter, txm: txm, sessionLookup: lookup, termCreator: termCreator, fundingWriter: fundingWriter, clock: clock}
}

func (uc *CreateChildWithFullProfile) Execute(ctx context.Context, actor tenant.ActorContext, input CreateChildFullInput) (*ChildCreationResult, error) {
	if err := uc.validateInput(input); err != nil {
		return nil, err
	}

	firstName := strings.TrimSpace(input.Child.FirstName)
	middleName := strings.TrimSpace(input.Child.MiddleName)
	lastName := strings.TrimSpace(input.Child.LastName)
	dob, _ := time.Parse("2006-01-02", strings.TrimSpace(input.Child.DateOfBirth))
	startDate, _ := time.Parse("2006-01-02", strings.TrimSpace(input.Child.StartDate))

	var endDate *time.Time
	if s := strings.TrimSpace(input.Child.EndDate); s != "" {
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			return nil, domainerrors.Validation("Invalid request payload.", "end_date")
		}
		endDate = &t
	}

	roomID, err := uuid.Parse(strings.TrimSpace(input.Room.RoomID))
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "room_id")
	}
	roomStart, err := time.Parse("2006-01-02", strings.TrimSpace(input.Room.StartDate))
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "room_start_date")
	}

	consent := input.Consent
	signedDate, err := time.Parse("2006-01-02", strings.TrimSpace(consent.SignedDate))
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "consents.signed_date")
	}

	var collectionPassword, collectionPasswordHint string
	if input.CollectionSettings != nil {
		collectionPassword = strings.TrimSpace(input.CollectionSettings.Password)
		collectionPasswordHint = strings.TrimSpace(input.CollectionSettings.PasswordHint)
	}

	var result ChildCreationResult
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		childID := uid.NewUUID()
		child := domain.Child{
			ID:          childID,
			FirstName:   firstName,
			DateOfBirth: dob,
			StartDate:   startDate,
			EndDate:     endDate,
			IsActive:    true,
		}
		if middleName != "" {
			child.MiddleName = &middleName
		}
		if lastName != "" {
			child.LastName = &lastName
		}

		notes := strings.TrimSpace(input.Child.Notes)
		if err := uc.repo.Create(ctx, tx, child, notes, actor.TenantID, actor.BranchID); err != nil {
			return fmt.Errorf("create child: %w", err)
		}

		created := []string{"identity"}

		// collection settings row (always)
		cs := &domain.ChildCollectionSetting{
			ID:       uid.NewUUID(),
			TenantID: actor.TenantID,
			BranchID: actor.BranchID,
			ChildID:  childID,
		}
		if input.CollectionSettings != nil {
			cs.Over18CollectionAcknowledged = input.CollectionSettings.Over18CollectionAcknowledged
		}
		if _, err := uc.repo.UpsertCollectionSetting(ctx, tx, cs); err != nil {
			return fmt.Errorf("create collection settings: %w", err)
		}
		created = append(created, "collection_settings")

		if collectionPassword != "" || collectionPasswordHint != "" {
			if err := uc.repo.SetCollectionPassword(ctx, tx, actor.TenantID, actor.BranchID, childID, cs.ID, collectionPassword, collectionPasswordHint, time.Now().UTC(), actor.UserID, actor.MembershipID); err != nil {
				return fmt.Errorf("set collection password: %w", err)
			}
		}

		// room assignment (required)
		room := &domain.ChildRoomAssignment{
			ID:        uid.NewUUID(),
			TenantID:  actor.TenantID,
			BranchID:  actor.BranchID,
			ChildID:   childID,
			RoomID:    roomID,
			StartDate: roomStart,
		}
		if _, err := uc.repo.InsertRoomAssignment(ctx, tx, room); err != nil {
			return fmt.Errorf("create room assignment: %w", err)
		}
		created = append(created, "room_assignment")

		// optional profile
		if input.Profile != nil {
			p := buildChildProfileFromInput(actor.TenantID, actor.BranchID, childID, input.Profile)
			if _, err := uc.repo.InsertProfile(ctx, tx, p); err != nil {
				return fmt.Errorf("create child profile: %w", err)
			}
			created = append(created, "profile")
		}

		// optional health
		if input.Health != nil {
			h := buildChildHealthFromInput(actor.TenantID, actor.BranchID, childID, input.Health)
			if _, err := uc.repo.UpsertHealth(ctx, tx, h); err != nil {
				return fmt.Errorf("create child health: %w", err)
			}
			created = append(created, "health")
		}

		// optional safeguarding
		if input.Safeguarding != nil {
			s := buildChildSafeguardingFromInput(actor.TenantID, actor.BranchID, childID, input.Safeguarding)
			if _, err := uc.repo.UpsertSafeguarding(ctx, tx, s); err != nil {
				return fmt.Errorf("create child safeguarding: %w", err)
			}
			created = append(created, "safeguarding")
		}

		// optional contacts
		if len(input.Contacts) > 0 {
			contactTypes := []domain.ContactType{
				domain.ContactTypeEmergencyContact,
				domain.ContactTypeAuthorisedCollector,
			}
			entries := buildChildContactEntries(actor.TenantID, actor.BranchID, childID, input.Contacts)
			if err := uc.repo.ReplaceContactsForTypes(ctx, tx, actor.TenantID, actor.BranchID, childID, contactTypes, entries); err != nil {
				return fmt.Errorf("create child contacts: %w", err)
			}
			created = append(created, "contacts")
		}

		// required consent
		consentRecord := &domain.ChildConsent{
			ID:                                   uid.NewUUID(),
			TenantID:                             actor.TenantID,
			BranchID:                             actor.BranchID,
			ChildID:                              childID,
			UrgentMedicalTreatment:               consent.UrgentMedicalTreatment,
			UrgentMedicalTreatmentExceptions:     consent.UrgentMedicalTreatmentExceptions,
			Plasters:                             consent.Plasters,
			SafeguardingReportingAcknowledgement: consent.SafeguardingReportingAcknowledgement,
			InformationSharingConsent:            consent.InformationSharingConsent,
			GDPRDataProcessingConsent:            consent.GDPRDataProcessingConsent,
			InformationTruthfulnessDeclaration:   consent.InformationTruthfulnessDeclaration,
			AreaSENCOLiaison:                     consent.AreaSENCOLiaison,
			HealthVisitorLiaison:                 consent.HealthVisitorLiaison,
			TransitionDocuments:                  consent.TransitionDocuments,
			LocalOutings:                         consent.LocalOutings,
			FacePainting:                         consent.FacePainting,
			ParentSuppliedSunCream:               consent.ParentSuppliedSunCream,
			ParentSuppliedNappyCream:             consent.ParentSuppliedNappyCream,
			DevelopmentProfilePhotos:             consent.DevelopmentProfilePhotos,
			NurseryDisplayBoards:                 consent.NurseryDisplayBoards,
			PromotionalLiterature:                consent.PromotionalLiterature,
			NurseryWebsite:                       consent.NurseryWebsite,
			StaffStudentCoursework:               consent.StaffStudentCoursework,
			SocialMedia:                          consent.SocialMedia,
			SocialMediaChannelNotes:              consent.SocialMediaChannelNotes,
			NotesExceptions:                      consent.NotesExceptions,
			SignerName:                           strings.TrimSpace(consent.SignerName),
			SignedDate:                           signedDate,
			PaperFormOnFile:                      consent.PaperFormOnFile,
			EnteredByUserID:                      actor.UserID,
			EnteredByMembershipID:                actor.MembershipID,
		}
		if _, err := uc.repo.InsertConsent(ctx, tx, consentRecord); err != nil {
			return fmt.Errorf("create child consent: %w", err)
		}
		created = append(created, "consent")

		// optional funding
		if input.Funding != nil {
			if err := uc.fundingWriter.SaveFunding(ctx, tx, actor.TenantID, actor.BranchID, childID, input.Funding); err != nil {
				return fmt.Errorf("create child funding: %w", err)
			}
			if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "child_funding_created",
				EntityType: "child",
				EntityID:   childID,
				Details: map[string]any{
					"funding_enabled": input.Funding.FundingEnabled,
					"funding_type":    input.Funding.FundingType,
				},
			}); aerr != nil {
				return fmt.Errorf("audit child_funding_created: %w", aerr)
			}
			created = append(created, "funding")
		}

		// billing profile (default: site_rate, no custom)
		bp := &domain.ChildBillingProfile{
			ID:            uid.NewUUID(),
			TenantID:      actor.TenantID,
			BranchID:      actor.BranchID,
			ChildID:       childID,
			BillingBasis:  domain.BillingBasisSiteRate,
			EffectiveFrom: time.Now().UTC(),
		}
		if _, err := uc.repo.UpsertBillingProfile(ctx, tx, bp); err != nil {
			return fmt.Errorf("create child billing profile: %w", err)
		}
		created = append(created, "billing_profile")

		if uc.audit != nil {
			if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
				ActionType: "child_created",
				EntityType: "child",
				EntityID:   childID,
				Details: map[string]any{
					"first_name":          firstName,
					"created_sub_records": created,
				},
			}); err != nil {
				return fmt.Errorf("audit child_created: %w", err)
			}
		}

		result = ChildCreationResult{
			ChildID:           childID,
			FirstName:         firstName,
			MiddleName:        child.MiddleName,
			LastName:          child.LastName,
			StartDate:         startDate.Format("2006-01-02"),
			TermID:            result.TermID,
			CreatedSubRecords: created,
		}
		return nil
	})
	if err != nil {
		return nil, mapExecTxError(err)
	}

	return &result, nil
}

func (uc *CreateChildWithFullProfile) validateInput(input CreateChildFullInput) error {
	var missing []string
	var fieldErrors []domainerrors.FieldError

	if strings.TrimSpace(input.Child.FirstName) == "" {
		missing = append(missing, "first_name")
	}
	if strings.TrimSpace(input.Child.DateOfBirth) == "" {
		missing = append(missing, "date_of_birth")
	} else if _, err := time.Parse("2006-01-02", strings.TrimSpace(input.Child.DateOfBirth)); err != nil {
		fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "date_of_birth", Message: "Invalid request payload."})
	}
	if strings.TrimSpace(input.Child.StartDate) == "" {
		missing = append(missing, "start_date")
	} else if _, err := time.Parse("2006-01-02", strings.TrimSpace(input.Child.StartDate)); err != nil {
		fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "start_date", Message: "Invalid request payload."})
	}
	if input.Consent == nil {
		missing = append(missing, "consents")
	} else if !input.Consent.SafeguardingReportingAcknowledgement {
		missing = append(missing, "consents.safeguarding_reporting_acknowledgement")
	}
	if input.Room == nil || strings.TrimSpace(input.Room.RoomID) == "" {
		fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "room_id", Message: "Pick a room."})
	} else if _, err := uuid.Parse(strings.TrimSpace(input.Room.RoomID)); err != nil {
		fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "room_id", Message: "Pick a room."})
	} else if input.Room != nil {
		if s := strings.TrimSpace(input.Room.StartDate); s == "" {
			fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "room_start_date", Message: "Pick a start date."})
		} else if _, err := time.Parse("2006-01-02", s); err != nil {
			fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "room_start_date", Message: "Pick a start date."})
		}
	}
	if input.Consent != nil {
		if strings.TrimSpace(input.Consent.SignerName) == "" {
			fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "consents.signer_name", Message: "Enter the signer name."})
		}
		if s := strings.TrimSpace(input.Consent.SignedDate); s == "" {
			fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "consents.signed_date", Message: "Enter the signed date."})
		} else if _, err := time.Parse("2006-01-02", s); err != nil {
			fieldErrors = append(fieldErrors, domainerrors.FieldError{Field: "consents.signed_date", Message: "Enter the signed date."})
		}
	}

	if input.Funding != nil && input.Funding.FundingEnabled {
		fieldErrors = append(fieldErrors, validateFundingInput(input.Funding, "funding.")...)
	}

	if len(missing) > 0 {
		return domainerrors.Validation(
			fmt.Sprintf("Missing required fields: %s", strings.Join(missing, ", ")),
			strings.Join(missing, ","),
		)
	}
	if len(fieldErrors) > 0 {
		return domainerrors.ValidationWithFields("Some fields did not validate.", fieldErrors)
	}
	return nil
}
