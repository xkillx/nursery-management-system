package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetChildSummary(ctx context.Context, tenantID, branchID, childID uuid.UUID) (domain.ChildSummary, bool, error) {
	q := sqlc.New(r.pool)
	row, err := q.RegistrationProfileChildGet(ctx, sqlc.RegistrationProfileChildGetParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ID:       uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ChildSummary{}, false, nil
	}
	if err != nil {
		return domain.ChildSummary{}, false, fmt.Errorf("get child summary: %w", err)
	}
	return domain.ChildSummary{
		ID:          childID,
		FullName:    row.FullName,
		DateOfBirth: pgtypeDateToTime(row.DateOfBirth),
	}, true, nil
}

func (r *Repository) GetByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.Profile, error) {
	q := sqlc.New(r.pool)
	row, err := q.RegistrationProfileGetByChild(ctx, sqlc.RegistrationProfileGetByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get profile by child: %w", err)
	}
	return mapProfileFromRowFields(
		row.ID, row.TenantID, row.BranchID, row.ChildID,
		row.Sex, row.Religion, row.EthnicOrigin, row.FirstLanguage, row.OtherLanguages,
		row.HomeAddress, row.HomePostcode, row.HomeTelephone,
		row.CrpDisabilityStatus, row.DisabilityNotes, row.AccessRequirements,
		row.CrpMedicalConditionsStatus, row.MedicalConditionsNotes,
		row.CrpPrescribedMedicationStatus, row.MedicationNotes,
		row.CrpImmunisationStatus, row.ImmunisationCountry, row.IllnessDiagnosisHistory,
		row.CrpDietaryRequirementsStatus, row.DietaryRequirementsNotes, row.DietarySideEffects,
		row.DoctorName, row.DoctorAddress, row.DoctorPhone,
		row.HealthVisitorName, row.HealthVisitorAddress, row.HealthVisitorPhone,
		row.CrpSocialServicesStatus, row.SocialServicesNotes, row.SocialWorkerName, row.SocialWorkerPhone, row.SocialWorkerEmail,
		row.CrpConcernWalking, row.CrpConcernSpeechLanguage, row.CrpConcernHearing,
		row.CrpConcernSight, row.CrpConcernEmotionalWellbeing, row.CrpConcernBehaviour,
		row.ProfessionalReferrals,
		row.ParentalResponsibilityNotes, row.Over18CollectionAcknowledged,
		row.CollectionPasswordIsSet, row.CollectionPasswordUpdatedAt,
		row.CollectionPasswordUpdatedByUserID, row.CollectionPasswordUpdatedByMembershipID,
		row.CrpBenefitsContributeToFees, row.CrpWorkingTaxCredit,
		row.CrpCollegeUniPaidToParent, row.CrpCollegeUniPaidToNursery,
		row.CrpFunding3yoTermTime, row.CrpFunding2yoTermTime, row.FundingSupportNotes,
		row.RoutineCareNotes,
		row.GdprDeclaredByName, row.GdprDeclaredAt, row.GdprDeclarationDate,
		row.DemographicsHomeReviewed, row.MedicalDietaryReviewed, row.HealthContactsReviewed,
		row.SocialDevelopmentReviewed, row.ParentResponsibilityReviewed,
		row.EmergencyCollectionReviewed, row.FundingSupportReviewed, row.RoutineCareReviewed,
		row.CreatedAt, row.UpdatedAt), nil
}

func (r *Repository) GetForUpdateByChild(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) (*domain.Profile, error) {
	q := sqlc.New(tx)
	row, err := q.RegistrationProfileGetForUpdateByChild(ctx, sqlc.RegistrationProfileGetForUpdateByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get profile for update: %w", err)
	}
	return mapProfileForUpdateRow(row), nil
}

func (r *Repository) Create(ctx context.Context, tx domain.Tx, profile *domain.Profile) (*domain.Profile, error) {
	q := sqlc.New(tx)
	homeAddr, _ := json.Marshal(profile.HomeAddress)
	referrals := []byte("[]")
	if profile.ProfessionalReferrals != nil {
		referrals, _ = json.Marshal(profile.ProfessionalReferrals)
	}

	row, err := q.RegistrationProfileCreate(ctx, sqlc.RegistrationProfileCreateParams{
		ID:                            uuidToPgtype(profile.ID),
		TenantID:                      uuidToPgtype(profile.TenantID),
		BranchID:                      uuidToPgtype(profile.BranchID),
		ChildID:                       uuidToPgtype(profile.ChildID),
		Sex:                           textToPgtype(profile.Sex),
		Religion:                      textToPgtype(profile.Religion),
		EthnicOrigin:                  textToPgtype(profile.EthnicOrigin),
		FirstLanguage:                 textToPgtype(profile.FirstLanguage),
		OtherLanguages:                profile.OtherLanguages,
		HomeAddress:                   homeAddr,
		HomePostcode:                  textToPgtype(profile.HomePostcode),
		HomeTelephone:                 textToPgtype(profile.HomeTelephone),
		DisabilityStatus:              sqlc.RegistrationYesNoUnknown(profile.DisabilityStatus),
		DisabilityNotes:               textToPgtype(profile.DisabilityNotes),
		AccessRequirements:            textToPgtype(profile.AccessRequirements),
		MedicalConditionsStatus:       sqlc.RegistrationYesNoUnknown(profile.MedicalConditionsStatus),
		MedicalConditionsNotes:        textToPgtype(profile.MedicalConditionsNotes),
		PrescribedMedicationStatus:    sqlc.RegistrationYesNoUnknown(profile.PrescribedMedicationStatus),
		MedicationNotes:               textToPgtype(profile.MedicationNotes),
		ImmunisationStatus:            sqlc.RegistrationImmunisationStatus(profile.ImmunisationStatus),
		ImmunisationCountry:           textToPgtype(profile.ImmunisationCountry),
		IllnessDiagnosisHistory:       textToPgtype(profile.IllnessDiagnosisHistory),
		DietaryRequirementsStatus:     sqlc.RegistrationYesNoUnknown(profile.DietaryRequirementsStatus),
		DietaryRequirementsNotes:      textToPgtype(profile.DietaryRequirementsNotes),
		DietarySideEffects:            textToPgtype(profile.DietarySideEffects),
		DoctorName:                    textToPgtype(profile.DoctorName),
		DoctorAddress:                 textToPgtype(profile.DoctorAddress),
		DoctorPhone:                   textToPgtype(profile.DoctorPhone),
		HealthVisitorName:             textToPgtype(profile.HealthVisitorName),
		HealthVisitorAddress:          textToPgtype(profile.HealthVisitorAddress),
		HealthVisitorPhone:            textToPgtype(profile.HealthVisitorPhone),
		SocialServicesStatus:          sqlc.RegistrationYesNoUnknown(profile.SocialServicesStatus),
		SocialServicesNotes:           textToPgtype(profile.SocialServicesNotes),
		SocialWorkerName:              textToPgtype(profile.SocialWorkerName),
		SocialWorkerPhone:             textToPgtype(profile.SocialWorkerPhone),
		SocialWorkerEmail:             textToPgtype(profile.SocialWorkerEmail),
		ConcernWalking:                sqlc.RegistrationYesNoUnknown(profile.ConcernWalking),
		ConcernSpeechLanguage:         sqlc.RegistrationYesNoUnknown(profile.ConcernSpeechLanguage),
		ConcernHearing:                sqlc.RegistrationYesNoUnknown(profile.ConcernHearing),
		ConcernSight:                  sqlc.RegistrationYesNoUnknown(profile.ConcernSight),
		ConcernEmotionalWellbeing:     sqlc.RegistrationYesNoUnknown(profile.ConcernEmotionalWellbeing),
		ConcernBehaviour:              sqlc.RegistrationYesNoUnknown(profile.ConcernBehaviour),
		ProfessionalReferrals:         referrals,
		ParentalResponsibilityNotes:   textToPgtype(profile.ParentalResponsibilityNotes),
		Over18CollectionAcknowledged:  profile.Over18CollectionAcknowledged,
		BenefitsContributeToFees:      sqlc.RegistrationYesNoUnknown(profile.BenefitsContributeToFees),
		WorkingTaxCredit:              sqlc.RegistrationYesNoUnknown(profile.WorkingTaxCredit),
		CollegeUniPaidToParent:        sqlc.RegistrationYesNoUnknown(profile.CollegeUniPaidToParent),
		CollegeUniPaidToNursery:       sqlc.RegistrationYesNoUnknown(profile.CollegeUniPaidToNursery),
		Funding3yoTermTime:            sqlc.RegistrationYesNoUnknown(profile.Funding3yoTermTime),
		Funding2yoTermTime:            sqlc.RegistrationYesNoUnknown(profile.Funding2yoTermTime),
		FundingSupportNotes:           textToPgtype(profile.FundingSupportNotes),
		RoutineCareNotes:              textToPgtype(profile.RoutineCareNotes),
		GdprDeclaredByName:            textToPgtype(profile.GDPRDeclaredByName),
		GdprDeclaredAt:                timestamptzToPgtype(profile.GDPRDeclaredAt),
		GdprDeclarationDate:           dateToPgtype(profile.GDPRDeclarationDate),
		DemographicsHomeReviewed:      profile.DemographicsHomeReviewed,
		MedicalDietaryReviewed:        profile.MedicalDietaryReviewed,
		HealthContactsReviewed:        profile.HealthContactsReviewed,
		SocialDevelopmentReviewed:     profile.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed:  profile.ParentResponsibilityReviewed,
		EmergencyCollectionReviewed:   profile.EmergencyCollectionReviewed,
		FundingSupportReviewed:        profile.FundingSupportReviewed,
		RoutineCareReviewed:           profile.RoutineCareReviewed,
	})
	if err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return mapCreateRow(row), nil
}

func (r *Repository) Update(ctx context.Context, tx domain.Tx, profile *domain.Profile) (*domain.Profile, error) {
	q := sqlc.New(tx)
	homeAddr, _ := json.Marshal(profile.HomeAddress)
	referrals := []byte("[]")
	if profile.ProfessionalReferrals != nil {
		referrals, _ = json.Marshal(profile.ProfessionalReferrals)
	}

	row, err := q.RegistrationProfileUpdate(ctx, sqlc.RegistrationProfileUpdateParams{
		TenantID:                      uuidToPgtype(profile.TenantID),
		BranchID:                      uuidToPgtype(profile.BranchID),
		ChildID:                       uuidToPgtype(profile.ChildID),
		Sex:                           textToPgtype(profile.Sex),
		Religion:                      textToPgtype(profile.Religion),
		EthnicOrigin:                  textToPgtype(profile.EthnicOrigin),
		FirstLanguage:                 textToPgtype(profile.FirstLanguage),
		OtherLanguages:                profile.OtherLanguages,
		HomeAddress:                   homeAddr,
		HomePostcode:                  textToPgtype(profile.HomePostcode),
		HomeTelephone:                 textToPgtype(profile.HomeTelephone),
		DisabilityStatus:              sqlc.RegistrationYesNoUnknown(profile.DisabilityStatus),
		DisabilityNotes:               textToPgtype(profile.DisabilityNotes),
		AccessRequirements:            textToPgtype(profile.AccessRequirements),
		MedicalConditionsStatus:       sqlc.RegistrationYesNoUnknown(profile.MedicalConditionsStatus),
		MedicalConditionsNotes:        textToPgtype(profile.MedicalConditionsNotes),
		PrescribedMedicationStatus:    sqlc.RegistrationYesNoUnknown(profile.PrescribedMedicationStatus),
		MedicationNotes:               textToPgtype(profile.MedicationNotes),
		ImmunisationStatus:            sqlc.RegistrationImmunisationStatus(profile.ImmunisationStatus),
		ImmunisationCountry:           textToPgtype(profile.ImmunisationCountry),
		IllnessDiagnosisHistory:       textToPgtype(profile.IllnessDiagnosisHistory),
		DietaryRequirementsStatus:     sqlc.RegistrationYesNoUnknown(profile.DietaryRequirementsStatus),
		DietaryRequirementsNotes:      textToPgtype(profile.DietaryRequirementsNotes),
		DietarySideEffects:            textToPgtype(profile.DietarySideEffects),
		DoctorName:                    textToPgtype(profile.DoctorName),
		DoctorAddress:                 textToPgtype(profile.DoctorAddress),
		DoctorPhone:                   textToPgtype(profile.DoctorPhone),
		HealthVisitorName:             textToPgtype(profile.HealthVisitorName),
		HealthVisitorAddress:          textToPgtype(profile.HealthVisitorAddress),
		HealthVisitorPhone:            textToPgtype(profile.HealthVisitorPhone),
		SocialServicesStatus:          sqlc.RegistrationYesNoUnknown(profile.SocialServicesStatus),
		SocialServicesNotes:           textToPgtype(profile.SocialServicesNotes),
		SocialWorkerName:              textToPgtype(profile.SocialWorkerName),
		SocialWorkerPhone:             textToPgtype(profile.SocialWorkerPhone),
		SocialWorkerEmail:             textToPgtype(profile.SocialWorkerEmail),
		ConcernWalking:                sqlc.RegistrationYesNoUnknown(profile.ConcernWalking),
		ConcernSpeechLanguage:         sqlc.RegistrationYesNoUnknown(profile.ConcernSpeechLanguage),
		ConcernHearing:                sqlc.RegistrationYesNoUnknown(profile.ConcernHearing),
		ConcernSight:                  sqlc.RegistrationYesNoUnknown(profile.ConcernSight),
		ConcernEmotionalWellbeing:     sqlc.RegistrationYesNoUnknown(profile.ConcernEmotionalWellbeing),
		ConcernBehaviour:              sqlc.RegistrationYesNoUnknown(profile.ConcernBehaviour),
		ProfessionalReferrals:         referrals,
		ParentalResponsibilityNotes:   textToPgtype(profile.ParentalResponsibilityNotes),
		Over18CollectionAcknowledged:  profile.Over18CollectionAcknowledged,
		BenefitsContributeToFees:      sqlc.RegistrationYesNoUnknown(profile.BenefitsContributeToFees),
		WorkingTaxCredit:              sqlc.RegistrationYesNoUnknown(profile.WorkingTaxCredit),
		CollegeUniPaidToParent:        sqlc.RegistrationYesNoUnknown(profile.CollegeUniPaidToParent),
		CollegeUniPaidToNursery:       sqlc.RegistrationYesNoUnknown(profile.CollegeUniPaidToNursery),
		Funding3yoTermTime:            sqlc.RegistrationYesNoUnknown(profile.Funding3yoTermTime),
		Funding2yoTermTime:            sqlc.RegistrationYesNoUnknown(profile.Funding2yoTermTime),
		FundingSupportNotes:           textToPgtype(profile.FundingSupportNotes),
		RoutineCareNotes:              textToPgtype(profile.RoutineCareNotes),
		GdprDeclaredByName:            textToPgtype(profile.GDPRDeclaredByName),
		GdprDeclaredAt:                timestamptzToPgtype(profile.GDPRDeclaredAt),
		GdprDeclarationDate:           dateToPgtype(profile.GDPRDeclarationDate),
		DemographicsHomeReviewed:      profile.DemographicsHomeReviewed,
		MedicalDietaryReviewed:        profile.MedicalDietaryReviewed,
		HealthContactsReviewed:        profile.HealthContactsReviewed,
		SocialDevelopmentReviewed:     profile.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed:  profile.ParentResponsibilityReviewed,
		EmergencyCollectionReviewed:   profile.EmergencyCollectionReviewed,
		FundingSupportReviewed:        profile.FundingSupportReviewed,
		RoutineCareReviewed:           profile.RoutineCareReviewed,
	})
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return mapUpdateRow(row), nil
}

func (r *Repository) SetCollectionPassword(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, hash string, updatedAt time.Time, updatedByUserID, updatedByMembershipID uuid.UUID) error {
	q := sqlc.New(tx)
	_, err := q.RegistrationProfileSetCollectionPassword(ctx, sqlc.RegistrationProfileSetCollectionPasswordParams{
		TenantID:                      uuidToPgtype(tenantID),
		BranchID:                      uuidToPgtype(branchID),
		ChildID:                       uuidToPgtype(childID),
		CollectionPasswordHash:        pgtype.Text{String: hash, Valid: true},
		CollectionPasswordUpdatedAt:   pgtype.Timestamptz{Time: updatedAt, Valid: true},
		CollectionPasswordUpdatedByUserID:       uuidToPgtype(updatedByUserID),
		CollectionPasswordUpdatedByMembershipID: uuidToPgtype(updatedByMembershipID),
	})
	if err != nil {
		return fmt.Errorf("set collection password: %w", err)
	}
	return nil
}

func (r *Repository) ReplaceContactsForTypes(ctx context.Context, tx domain.Tx, profileID uuid.UUID, contactTypes []domain.ContactType, entries []domain.ContactEntry) error {
	q := sqlc.New(tx)

	sqlcTypes := make([]string, len(contactTypes))
	for i, ct := range contactTypes {
		sqlcTypes[i] = string(ct)
	}

	if len(entries) > 0 {
		err := q.RegistrationProfileDeleteContactsByTypes(ctx, sqlc.RegistrationProfileDeleteContactsByTypesParams{
			TenantID:     uuidToPgtype(entries[0].TenantID),
			BranchID:     uuidToPgtype(entries[0].BranchID),
			ProfileID:    uuidToPgtype(profileID),
			ContactTypes: sqlcTypes,
		})
		if err != nil {
			return fmt.Errorf("delete contacts for types: %w", err)
		}
	}

	for _, e := range entries {
		addr, _ := json.Marshal(e.Address)
		workAddr, _ := json.Marshal(e.WorkAddress)
		_, err := q.RegistrationProfileCreateContact(ctx, sqlc.RegistrationProfileCreateContactParams{
			ID:                       uuidToPgtype(e.ID),
			TenantID:                 uuidToPgtype(e.TenantID),
			BranchID:                 uuidToPgtype(e.BranchID),
			ProfileID:                uuidToPgtype(e.ProfileID),
			ChildID:                  uuidToPgtype(e.ChildID),
			ContactType:              sqlc.RegistrationContactType(e.ContactType),
			SortOrder:                int32(e.SortOrder),
			FullName:                 e.FullName,
			RelationshipToChild:      textToPgtype(e.RelationshipToChild),
			Address:                  addr,
			Telephone:                textToPgtype(e.Telephone),
			Email:                    textToPgtype(e.Email),
			WorkAddress:              workAddr,
			HasParentalResponsibility: boolToPgtype(e.HasParentalResponsibility),
		})
		if err != nil {
			return fmt.Errorf("create contact: %w", err)
		}
	}

	return nil
}

func (r *Repository) ListContactsByProfile(ctx context.Context, tenantID, branchID, profileID uuid.UUID) ([]domain.ContactEntry, error) {
	q := sqlc.New(r.pool)
	rows, err := q.RegistrationProfileListContacts(ctx, sqlc.RegistrationProfileListContactsParams{
		TenantID:  uuidToPgtype(tenantID),
		BranchID:  uuidToPgtype(branchID),
		ProfileID: uuidToPgtype(profileID),
	})
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	entries := make([]domain.ContactEntry, 0, len(rows))
	for _, r := range rows {
		entries = append(entries, mapContactRow(r))
	}
	return entries, nil
}

func mapProfileFromRowFields(
	id, tenantID, branchID, childID pgtype.UUID,
	sex, religion, ethnicOrigin, firstLanguage pgtype.Text,
	otherLanguages []string,
	homeAddress []byte,
	homePostcode, homeTelephone pgtype.Text,
	disabilityStatus string, disabilityNotes, accessRequirements pgtype.Text,
	medicalConditionsStatus string, medicalConditionsNotes pgtype.Text,
	prescribedMedicationStatus string, medicationNotes pgtype.Text,
	immunisationStatus string, immunisationCountry pgtype.Text,
	illnessDiagnosisHistory pgtype.Text,
	dietaryRequirementsStatus string, dietaryRequirementsNotes pgtype.Text,
	dietarySideEffects pgtype.Text,
	doctorName, doctorAddress, doctorPhone pgtype.Text,
	healthVisitorName, healthVisitorAddress, healthVisitorPhone pgtype.Text,
	socialServicesStatus string, socialServicesNotes, socialWorkerName, socialWorkerPhone, socialWorkerEmail pgtype.Text,
	concernWalking, concernSpeechLanguage, concernHearing, concernSight, concernEmotionalWellbeing, concernBehaviour string,
	professionalReferrals []byte,
	parentalResponsibilityNotes pgtype.Text,
	over18CollectionAcknowledged bool,
	collectionPasswordIsSet pgtype.Bool,
	collectionPasswordUpdatedAt pgtype.Timestamptz,
	collectionPasswordUpdatedByUserID, collectionPasswordUpdatedByMembershipID pgtype.UUID,
	benefitsContributeToFees, workingTaxCredit, collegeUniPaidToParent, collegeUniPaidToNursery, funding3yoTermTime, funding2yoTermTime string,
	fundingSupportNotes pgtype.Text,
	routineCareNotes pgtype.Text,
	gdprDeclaredByName pgtype.Text,
	gdprDeclaredAt pgtype.Timestamptz,
	gdprDeclarationDate pgtype.Date,
	demographicsHomeReviewed, medicalDietaryReviewed, healthContactsReviewed,
	socialDevelopmentReviewed, parentResponsibilityReviewed,
	emergencyCollectionReviewed, fundingSupportReviewed, routineCareReviewed bool,
	createdAt, updatedAt pgtype.Timestamptz,
) *domain.Profile {
	p := &domain.Profile{
		ID:                           pgtypeUUIDToUUID(id),
		TenantID:                     pgtypeUUIDToUUID(tenantID),
		BranchID:                     pgtypeUUIDToUUID(branchID),
		ChildID:                      pgtypeUUIDToUUID(childID),
		Sex:                          pgtypeTextToPtr(sex),
		Religion:                     pgtypeTextToPtr(religion),
		EthnicOrigin:                 pgtypeTextToPtr(ethnicOrigin),
		FirstLanguage:                pgtypeTextToPtr(firstLanguage),
		OtherLanguages:               otherLanguages,
		HomePostcode:                 pgtypeTextToPtr(homePostcode),
		HomeTelephone:                pgtypeTextToPtr(homeTelephone),
		DisabilityStatus:             domain.YesNoUnknown(disabilityStatus),
		DisabilityNotes:              pgtypeTextToPtr(disabilityNotes),
		AccessRequirements:           pgtypeTextToPtr(accessRequirements),
		MedicalConditionsStatus:      domain.YesNoUnknown(medicalConditionsStatus),
		MedicalConditionsNotes:       pgtypeTextToPtr(medicalConditionsNotes),
		PrescribedMedicationStatus:   domain.YesNoUnknown(prescribedMedicationStatus),
		MedicationNotes:              pgtypeTextToPtr(medicationNotes),
		ImmunisationStatus:           domain.ImmunisationStatus(immunisationStatus),
		ImmunisationCountry:          pgtypeTextToPtr(immunisationCountry),
		IllnessDiagnosisHistory:      pgtypeTextToPtr(illnessDiagnosisHistory),
		DietaryRequirementsStatus:    domain.YesNoUnknown(dietaryRequirementsStatus),
		DietaryRequirementsNotes:     pgtypeTextToPtr(dietaryRequirementsNotes),
		DietarySideEffects:           pgtypeTextToPtr(dietarySideEffects),
		DoctorName:                   pgtypeTextToPtr(doctorName),
		DoctorAddress:                pgtypeTextToPtr(doctorAddress),
		DoctorPhone:                  pgtypeTextToPtr(doctorPhone),
		HealthVisitorName:            pgtypeTextToPtr(healthVisitorName),
		HealthVisitorAddress:         pgtypeTextToPtr(healthVisitorAddress),
		HealthVisitorPhone:           pgtypeTextToPtr(healthVisitorPhone),
		SocialServicesStatus:         domain.YesNoUnknown(socialServicesStatus),
		SocialServicesNotes:          pgtypeTextToPtr(socialServicesNotes),
		SocialWorkerName:             pgtypeTextToPtr(socialWorkerName),
		SocialWorkerPhone:            pgtypeTextToPtr(socialWorkerPhone),
		SocialWorkerEmail:            pgtypeTextToPtr(socialWorkerEmail),
		ConcernWalking:               domain.YesNoUnknown(concernWalking),
		ConcernSpeechLanguage:        domain.YesNoUnknown(concernSpeechLanguage),
		ConcernHearing:               domain.YesNoUnknown(concernHearing),
		ConcernSight:                 domain.YesNoUnknown(concernSight),
		ConcernEmotionalWellbeing:    domain.YesNoUnknown(concernEmotionalWellbeing),
		ConcernBehaviour:             domain.YesNoUnknown(concernBehaviour),
		ParentalResponsibilityNotes:  pgtypeTextToPtr(parentalResponsibilityNotes),
		Over18CollectionAcknowledged: over18CollectionAcknowledged,
		BenefitsContributeToFees:     domain.YesNoUnknown(benefitsContributeToFees),
		WorkingTaxCredit:             domain.YesNoUnknown(workingTaxCredit),
		CollegeUniPaidToParent:       domain.YesNoUnknown(collegeUniPaidToParent),
		CollegeUniPaidToNursery:      domain.YesNoUnknown(collegeUniPaidToNursery),
		Funding3yoTermTime:           domain.YesNoUnknown(funding3yoTermTime),
		Funding2yoTermTime:           domain.YesNoUnknown(funding2yoTermTime),
		FundingSupportNotes:          pgtypeTextToPtr(fundingSupportNotes),
		RoutineCareNotes:             pgtypeTextToPtr(routineCareNotes),
		GDPRDeclaredByName:           pgtypeTextToPtr(gdprDeclaredByName),
		GDPRDeclaredAt:               pgtypeTimestamptzToPtr(gdprDeclaredAt),
		GDPRDeclarationDate:          pgtypeDateToPtr(gdprDeclarationDate),
		DemographicsHomeReviewed:      demographicsHomeReviewed,
		MedicalDietaryReviewed:        medicalDietaryReviewed,
		HealthContactsReviewed:        healthContactsReviewed,
		SocialDevelopmentReviewed:     socialDevelopmentReviewed,
		ParentResponsibilityReviewed:  parentResponsibilityReviewed,
		EmergencyCollectionReviewed:   emergencyCollectionReviewed,
		FundingSupportReviewed:        fundingSupportReviewed,
		RoutineCareReviewed:           routineCareReviewed,
		CreatedAt:                     pgtypeTimestamptzToTime(createdAt),
		UpdatedAt:                     pgtypeTimestamptzToTime(updatedAt),
	}
	if homeAddress != nil {
		json.Unmarshal(homeAddress, &p.HomeAddress)
	}
	if p.HomeAddress == nil {
		p.HomeAddress = map[string]any{}
	}
	if professionalReferrals != nil {
		json.Unmarshal(professionalReferrals, &p.ProfessionalReferrals)
	}
	if p.ProfessionalReferrals == nil {
		p.ProfessionalReferrals = []domain.ProfessionalReferral{}
	}
	if otherLanguages == nil {
		p.OtherLanguages = []string{}
	}
	return p
}

func mapProfileForUpdateRow(row sqlc.RegistrationProfileGetForUpdateByChildRow) *domain.Profile {
	return mapProfileFromRowFields(
		row.ID, row.TenantID, row.BranchID, row.ChildID,
		row.Sex, row.Religion, row.EthnicOrigin, row.FirstLanguage, row.OtherLanguages,
		row.HomeAddress, row.HomePostcode, row.HomeTelephone,
		row.CrpDisabilityStatus, row.DisabilityNotes, row.AccessRequirements,
		row.CrpMedicalConditionsStatus, row.MedicalConditionsNotes,
		row.CrpPrescribedMedicationStatus, row.MedicationNotes,
		row.CrpImmunisationStatus, row.ImmunisationCountry, row.IllnessDiagnosisHistory,
		row.CrpDietaryRequirementsStatus, row.DietaryRequirementsNotes, row.DietarySideEffects,
		row.DoctorName, row.DoctorAddress, row.DoctorPhone,
		row.HealthVisitorName, row.HealthVisitorAddress, row.HealthVisitorPhone,
		row.CrpSocialServicesStatus, row.SocialServicesNotes, row.SocialWorkerName, row.SocialWorkerPhone, row.SocialWorkerEmail,
		row.CrpConcernWalking, row.CrpConcernSpeechLanguage, row.CrpConcernHearing,
		row.CrpConcernSight, row.CrpConcernEmotionalWellbeing, row.CrpConcernBehaviour,
		row.ProfessionalReferrals,
		row.ParentalResponsibilityNotes, row.Over18CollectionAcknowledged,
		row.CollectionPasswordIsSet, row.CollectionPasswordUpdatedAt,
		row.CollectionPasswordUpdatedByUserID, row.CollectionPasswordUpdatedByMembershipID,
		row.CrpBenefitsContributeToFees, row.CrpWorkingTaxCredit,
		row.CrpCollegeUniPaidToParent, row.CrpCollegeUniPaidToNursery,
		row.CrpFunding3yoTermTime, row.CrpFunding2yoTermTime, row.FundingSupportNotes,
		row.RoutineCareNotes,
		row.GdprDeclaredByName, row.GdprDeclaredAt, row.GdprDeclarationDate,
		row.DemographicsHomeReviewed, row.MedicalDietaryReviewed, row.HealthContactsReviewed,
		row.SocialDevelopmentReviewed, row.ParentResponsibilityReviewed,
		row.EmergencyCollectionReviewed, row.FundingSupportReviewed, row.RoutineCareReviewed,
		row.CreatedAt, row.UpdatedAt)
}

func mapCreateRow(row sqlc.RegistrationProfileCreateRow) *domain.Profile {
	return mapProfileFromRowFields(
		row.ID, row.TenantID, row.BranchID, row.ChildID,
		row.Sex, row.Religion, row.EthnicOrigin, row.FirstLanguage, row.OtherLanguages,
		row.HomeAddress, row.HomePostcode, row.HomeTelephone,
		row.DisabilityStatus, row.DisabilityNotes, row.AccessRequirements,
		row.MedicalConditionsStatus, row.MedicalConditionsNotes,
		row.PrescribedMedicationStatus, row.MedicationNotes,
		row.ImmunisationStatus, row.ImmunisationCountry, row.IllnessDiagnosisHistory,
		row.DietaryRequirementsStatus, row.DietaryRequirementsNotes, row.DietarySideEffects,
		row.DoctorName, row.DoctorAddress, row.DoctorPhone,
		row.HealthVisitorName, row.HealthVisitorAddress, row.HealthVisitorPhone,
		row.SocialServicesStatus, row.SocialServicesNotes, row.SocialWorkerName, row.SocialWorkerPhone, row.SocialWorkerEmail,
		row.ConcernWalking, row.ConcernSpeechLanguage, row.ConcernHearing,
		row.ConcernSight, row.ConcernEmotionalWellbeing, row.ConcernBehaviour,
		row.ProfessionalReferrals,
		row.ParentalResponsibilityNotes, row.Over18CollectionAcknowledged,
		pgtype.Bool{Bool: row.CollectionPasswordIsSet, Valid: true}, row.CollectionPasswordUpdatedAt,
		row.CollectionPasswordUpdatedByUserID, row.CollectionPasswordUpdatedByMembershipID,
		row.BenefitsContributeToFees, row.WorkingTaxCredit,
		row.CollegeUniPaidToParent, row.CollegeUniPaidToNursery,
		row.Funding3yoTermTime, row.Funding2yoTermTime, row.FundingSupportNotes,
		row.RoutineCareNotes,
		row.GdprDeclaredByName, row.GdprDeclaredAt, row.GdprDeclarationDate,
		row.DemographicsHomeReviewed, row.MedicalDietaryReviewed, row.HealthContactsReviewed,
		row.SocialDevelopmentReviewed, row.ParentResponsibilityReviewed,
		row.EmergencyCollectionReviewed, row.FundingSupportReviewed, row.RoutineCareReviewed,
		row.CreatedAt, row.UpdatedAt)
}

func mapUpdateRow(row sqlc.RegistrationProfileUpdateRow) *domain.Profile {
	return mapProfileFromRowFields(
		row.ID, row.TenantID, row.BranchID, row.ChildID,
		row.Sex, row.Religion, row.EthnicOrigin, row.FirstLanguage, row.OtherLanguages,
		row.HomeAddress, row.HomePostcode, row.HomeTelephone,
		row.DisabilityStatus, row.DisabilityNotes, row.AccessRequirements,
		row.MedicalConditionsStatus, row.MedicalConditionsNotes,
		row.PrescribedMedicationStatus, row.MedicationNotes,
		row.ImmunisationStatus, row.ImmunisationCountry, row.IllnessDiagnosisHistory,
		row.DietaryRequirementsStatus, row.DietaryRequirementsNotes, row.DietarySideEffects,
		row.DoctorName, row.DoctorAddress, row.DoctorPhone,
		row.HealthVisitorName, row.HealthVisitorAddress, row.HealthVisitorPhone,
		row.SocialServicesStatus, row.SocialServicesNotes, row.SocialWorkerName, row.SocialWorkerPhone, row.SocialWorkerEmail,
		row.ConcernWalking, row.ConcernSpeechLanguage, row.ConcernHearing,
		row.ConcernSight, row.ConcernEmotionalWellbeing, row.ConcernBehaviour,
		row.ProfessionalReferrals,
		row.ParentalResponsibilityNotes, row.Over18CollectionAcknowledged,
		row.CollectionPasswordIsSet, row.CollectionPasswordUpdatedAt,
		row.CollectionPasswordUpdatedByUserID, row.CollectionPasswordUpdatedByMembershipID,
		row.BenefitsContributeToFees, row.WorkingTaxCredit,
		row.CollegeUniPaidToParent, row.CollegeUniPaidToNursery,
		row.Funding3yoTermTime, row.Funding2yoTermTime, row.FundingSupportNotes,
		row.RoutineCareNotes,
		row.GdprDeclaredByName, row.GdprDeclaredAt, row.GdprDeclarationDate,
		row.DemographicsHomeReviewed, row.MedicalDietaryReviewed, row.HealthContactsReviewed,
		row.SocialDevelopmentReviewed, row.ParentResponsibilityReviewed,
		row.EmergencyCollectionReviewed, row.FundingSupportReviewed, row.RoutineCareReviewed,
		row.CreatedAt, row.UpdatedAt)
}

func mapContactRow(row sqlc.RegistrationProfileListContactsRow) domain.ContactEntry {
	e := domain.ContactEntry{
		ID:                       pgtypeUUIDToUUID(row.ID),
		TenantID:                 pgtypeUUIDToUUID(row.TenantID),
		BranchID:                 pgtypeUUIDToUUID(row.BranchID),
		ProfileID:                pgtypeUUIDToUUID(row.ProfileID),
		ChildID:                  pgtypeUUIDToUUID(row.ChildID),
		ContactType:              domain.ContactType(row.CrcContactType),
		SortOrder:                int(row.SortOrder),
		FullName:                 row.FullName,
		RelationshipToChild:      pgtypeTextToPtr(row.RelationshipToChild),
		Telephone:                pgtypeTextToPtr(row.Telephone),
		Email:                    pgtypeTextToPtr(row.Email),
		HasParentalResponsibility: pgtypeBoolToPtr(row.HasParentalResponsibility),
		CreatedAt:                pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:                pgtypeTimestamptzToTime(row.UpdatedAt),
	}
	if row.Address != nil {
		json.Unmarshal(row.Address, &e.Address)
	}
	if e.Address == nil {
		e.Address = map[string]any{}
	}
	if row.WorkAddress != nil {
		json.Unmarshal(row.WorkAddress, &e.WorkAddress)
	}
	if e.WorkAddress == nil {
		e.WorkAddress = map[string]any{}
	}
	return e
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func textToPgtype(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgtypeTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func timestamptzToPgtype(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func pgtypeTimestamptzToTime(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func pgtypeTimestamptzToPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func dateToPgtype(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func pgtypeDateToPtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

func pgtypeDateToTime(d pgtype.Date) time.Time {
	return d.Time
}

func boolToPgtype(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func pgtypeBoolToPtr(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}
	return &b.Bool
}

// --- Consent Repository ---

func (r *Repository) GetLatestByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ConsentRecord, error) {
	q := sqlc.New(r.pool)
	row, err := q.ConsentGetLatestByChild(ctx, sqlc.ConsentGetLatestByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest consent: %w", err)
	}
	return mapConsentRow(row), nil
}

func (r *Repository) ListByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.ConsentRecord, error) {
	q := sqlc.New(r.pool)
	rows, err := q.ConsentListByChild(ctx, sqlc.ConsentListByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		return nil, fmt.Errorf("list consents: %w", err)
	}
	out := make([]domain.ConsentRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapConsentRow(row))
	}
	return out, nil
}

func (r *Repository) GetCurrentVersion(ctx context.Context, tenantID, branchID, childID uuid.UUID) (int, error) {
	q := sqlc.New(r.pool)
	v, err := q.ConsentGetCurrentVersion(ctx, sqlc.ConsentGetCurrentVersionParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if err != nil {
		return 0, fmt.Errorf("get current consent version: %w", err)
	}
	vi, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected type for current_version: %T", v)
	}
	return int(vi), nil
}

func (r *Repository) CreateConsentRecord(ctx context.Context, tx domain.Tx, record *domain.ConsentRecord) error {
	q := sqlc.New(tx)
	return q.ConsentCreate(ctx, sqlc.ConsentCreateParams{
		ID:                               uuidToPgtype(record.ID),
		TenantID:                         uuidToPgtype(record.TenantID),
		BranchID:                         uuidToPgtype(record.BranchID),
		ChildID:                          uuidToPgtype(record.ChildID),
		Version:                          int32(record.Version),
		Source:                           string(record.Source),
		PaperFormOnFile:                  record.PaperFormOnFile,
		UrgentMedicalTreatment:           record.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions: textToPgtype(record.UrgentMedicalTreatmentExceptions),
		Plasters:                         record.Plasters,
		SafeguardingReportingAcknowledgement: record.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:        record.InformationSharingConsent,
		GdprDataProcessingConsent:        record.GDPRDataProcessingConsent,
		AreaSencoLiaison:                 record.AreaSENCOLiaison,
		HealthVisitorLiaison:             record.HealthVisitorLiaison,
		TransitionDocuments:              record.TransitionDocuments,
		LocalOutings:                     record.LocalOutings,
		FacePainting:                     record.FacePainting,
		ParentSuppliedSunCream:           record.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:         record.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:         record.DevelopmentProfilePhotos,
		NurseryDisplayBoards:             record.NurseryDisplayBoards,
		PromotionalLiterature:            record.PromotionalLiterature,
		NurseryWebsite:                   record.NurseryWebsite,
		StaffStudentCoursework:           record.StaffStudentCoursework,
		SocialMedia:                      record.SocialMedia,
		SocialMediaChannelNotes:          textToPgtype(record.SocialMediaChannelNotes),
		NotesExceptions:                  textToPgtype(record.NotesExceptions),
		EnteredByUserID:                  uuidToPgtype(record.EnteredByUserID),
		EnteredByMembershipID:            uuidToPgtype(record.EnteredByMembershipID),
	})
}

func mapConsentRow(row sqlc.ChildRegistrationConsentRecord) *domain.ConsentRecord {
	return &domain.ConsentRecord{
		ID:      pgtypeUUIDToUUID(row.ID),
		TenantID: pgtypeUUIDToUUID(row.TenantID),
		BranchID: pgtypeUUIDToUUID(row.BranchID),
		ChildID:  pgtypeUUIDToUUID(row.ChildID),
		Version:  int(row.Version),
		Source:   domain.ConsentSource(row.Source),
		PaperFormOnFile:  row.PaperFormOnFile,
		UrgentMedicalTreatment:         row.UrgentMedicalTreatment,
		UrgentMedicalTreatmentExceptions: pgtypeTextToPtr(row.UrgentMedicalTreatmentExceptions),
		Plasters:                       row.Plasters,
		SafeguardingReportingAcknowledgement: row.SafeguardingReportingAcknowledgement,
		InformationSharingConsent:      row.InformationSharingConsent,
		GDPRDataProcessingConsent:      row.GdprDataProcessingConsent,
		AreaSENCOLiaison:               row.AreaSencoLiaison,
		HealthVisitorLiaison:           row.HealthVisitorLiaison,
		TransitionDocuments:            row.TransitionDocuments,
		LocalOutings:                   row.LocalOutings,
		FacePainting:                   row.FacePainting,
		ParentSuppliedSunCream:         row.ParentSuppliedSunCream,
		ParentSuppliedNappyCream:       row.ParentSuppliedNappyCream,
		DevelopmentProfilePhotos:       row.DevelopmentProfilePhotos,
		NurseryDisplayBoards:           row.NurseryDisplayBoards,
		PromotionalLiterature:          row.PromotionalLiterature,
		NurseryWebsite:                 row.NurseryWebsite,
		StaffStudentCoursework:         row.StaffStudentCoursework,
		SocialMedia:                    row.SocialMedia,
		SocialMediaChannelNotes:        pgtypeTextToPtr(row.SocialMediaChannelNotes),
		NotesExceptions:                pgtypeTextToPtr(row.NotesExceptions),
		EnteredByUserID:       pgtypeUUIDToUUID(row.EnteredByUserID),
		EnteredByMembershipID: pgtypeUUIDToUUID(row.EnteredByMembershipID),
		CreatedAt:             pgtypeTimestamptzToTime(row.CreatedAt),
	}
}

// --- Attestation Repository ---

func (r *Repository) GetLatestAttestationByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.CompletionAttestation, error) {
	q := sqlc.New(r.pool)
	row, err := q.AttestationGetLatestByChild(ctx, sqlc.AttestationGetLatestByChildParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		ChildID:  uuidToPgtype(childID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest attestation: %w", err)
	}
	return mapAttestationRow(row), nil
}

func (r *Repository) CreateAttestation(ctx context.Context, tx domain.Tx, a *domain.CompletionAttestation) error {
	q := sqlc.New(tx)
	return q.AttestationCreate(ctx, sqlc.AttestationCreateParams{
		ID:                       uuidToPgtype(a.ID),
		TenantID:                 uuidToPgtype(a.TenantID),
		BranchID:                 uuidToPgtype(a.BranchID),
		ChildID:                  uuidToPgtype(a.ChildID),
		ConsentRecordID:          uuidToPgtypePtr(a.ConsentRecordID),
		ProfileUpdatedAt:         pgtype.Timestamptz{Time: a.ProfileUpdatedAt, Valid: true},
		AttestedByUserID:         uuidToPgtype(a.AttestedByUserID),
		AttestedByMembershipID:   uuidToPgtype(a.AttestedByMembershipID),
		AttestedAt:               pgtype.Timestamptz{Time: a.AttestedAt, Valid: true},
		RequestID:                textToPgtype(a.RequestID),
	})
}

func mapAttestationRow(row sqlc.ChildRegistrationCompletionAttestation) *domain.CompletionAttestation {
	return &domain.CompletionAttestation{
		ID:                     pgtypeUUIDToUUID(row.ID),
		TenantID:               pgtypeUUIDToUUID(row.TenantID),
		BranchID:               pgtypeUUIDToUUID(row.BranchID),
		ChildID:                pgtypeUUIDToUUID(row.ChildID),
		ConsentRecordID:        pgtypeUUIDToUUIDPtr(row.ConsentRecordID),
		ProfileUpdatedAt:       pgtypeTimestamptzToTime(row.ProfileUpdatedAt),
		AttestedByUserID:       pgtypeUUIDToUUID(row.AttestedByUserID),
		AttestedByMembershipID: pgtypeUUIDToUUID(row.AttestedByMembershipID),
		AttestedAt:             pgtypeTimestamptzToTime(row.AttestedAt),
		RequestID:              pgtypeTextToPtr(row.RequestID),
		CreatedAt:              pgtypeTimestamptzToTime(row.CreatedAt),
	}
}

func uuidToPgtypePtr(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(*u), Valid: true}
}

func pgtypeUUIDToUUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

