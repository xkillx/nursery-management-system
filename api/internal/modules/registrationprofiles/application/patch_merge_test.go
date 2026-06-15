package application

import (
	"testing"
	"time"

	"github.com/google/uuid"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
)

func yes(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }

func newProfile() *domain.Profile {
	id := uuid.Nil
	return &domain.Profile{
		ID:                         id,
		HomeAddress:                map[string]any{},
		DisabilityStatus:           domain.YesNoUnknownUnknown,
		MedicalConditionsStatus:    domain.YesNoUnknownUnknown,
		PrescribedMedicationStatus: domain.YesNoUnknownUnknown,
		ImmunisationStatus:         domain.ImmunisationUnknown,
		DietaryRequirementsStatus:  domain.YesNoUnknownUnknown,
		SocialServicesStatus:       domain.YesNoUnknownUnknown,
		ConcernWalking:             domain.YesNoUnknownUnknown,
		ConcernSpeechLanguage:      domain.YesNoUnknownUnknown,
		ConcernHearing:             domain.YesNoUnknownUnknown,
		ConcernSight:               domain.YesNoUnknownUnknown,
		ConcernEmotionalWellbeing:  domain.YesNoUnknownUnknown,
		ConcernBehaviour:           domain.YesNoUnknownUnknown,
		ProfessionalReferrals:      []domain.ProfessionalReferral{},
		BenefitsContributeToFees:   domain.YesNoUnknownUnknown,
		WorkingTaxCredit:           domain.YesNoUnknownUnknown,
		CollegeUniPaidToParent:     domain.YesNoUnknownUnknown,
		CollegeUniPaidToNursery:    domain.YesNoUnknownUnknown,
		Funding3yoTermTime:         domain.YesNoUnknownUnknown,
		Funding2yoTermTime:         domain.YesNoUnknownUnknown,
	}
}

func TestMergePatch_DemographicsHomeWithDisabilityAccess(t *testing.T) {
	p := newProfile()
	homeAddr := map[string]any{"text": "123 Main St, London"}

	_, err := MergePatch(p, PatchSection{
		DemographicsHome: &DemographicsHomePatch{
			Sex:                      yes("female"),
			Religion:                 yes("Christian"),
			EthnicOrigin:             yes("White British"),
			FirstLanguage:            yes("English"),
			OtherLanguages:           yes("French"),
			HomeAddress:              &homeAddr,
			HomePostcode:             yes("SW1A 1AA"),
			HomeTelephone:            yes("020 1234 5678"),
			DisabilityStatus:         yes("no"),
			DisabilityNotes:          yes("None"),
			AccessRequirements:       yes("None"),
			DemographicsHomeReviewed: boolPtr(true),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Sex == nil || *p.Sex != "female" {
		t.Errorf("expected sex=female, got %v", p.Sex)
	}
	if p.Religion == nil || *p.Religion != "Christian" {
		t.Errorf("expected religion=Christian, got %v", p.Religion)
	}
	if p.DisabilityStatus != domain.YesNoUnknownNo {
		t.Errorf("expected disability_status=no, got %v", p.DisabilityStatus)
	}
	if p.DisabilityNotes == nil || *p.DisabilityNotes != "None" {
		t.Errorf("expected disability_notes=None, got %v", p.DisabilityNotes)
	}
	if p.AccessRequirements == nil || *p.AccessRequirements != "None" {
		t.Errorf("expected access_requirements=None, got %v", p.AccessRequirements)
	}
	if p.DemographicsHomeReviewed != true {
		t.Errorf("expected demographics_home_reviewed=true, got %v", p.DemographicsHomeReviewed)
	}
}

func TestMergePatch_DisabilityStatusInvalid(t *testing.T) {
	p := newProfile()
	_, err := MergePatch(p, PatchSection{
		DemographicsHome: &DemographicsHomePatch{
			DisabilityStatus: yes("invalid"),
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid disability_status")
	}
}

func TestMergePatch_DisabilityStatusYes(t *testing.T) {
	p := newProfile()
	_, err := MergePatch(p, PatchSection{
		DemographicsHome: &DemographicsHomePatch{
			DisabilityStatus: yes("yes"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DisabilityStatus != domain.YesNoUnknownYes {
		t.Errorf("expected disability_status=yes, got %v", p.DisabilityStatus)
	}
}

func TestMergePatch_DisabilityStatusUnknown(t *testing.T) {
	p := newProfile()
	_, err := MergePatch(p, PatchSection{
		DemographicsHome: &DemographicsHomePatch{
			DisabilityStatus: yes("unknown"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DisabilityStatus != domain.YesNoUnknownUnknown {
		t.Errorf("expected disability_status=unknown, got %v", p.DisabilityStatus)
	}
}

func TestMergePatch_MedicalDietaryEnums(t *testing.T) {
	p := newProfile()

	_, err := MergePatch(p, PatchSection{
		MedicalDietary: &MedicalDietaryPatch{
			MedicalConditionsStatus:    yes("yes"),
			PrescribedMedicationStatus: yes("no"),
			DietaryRequirementsStatus:  yes("unknown"),
			ImmunisationStatus:         yes("up_to_date"),
			MedicalDietaryReviewed:     boolPtr(true),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MedicalConditionsStatus != domain.YesNoUnknownYes {
		t.Errorf("expected medical_conditions=yes, got %v", p.MedicalConditionsStatus)
	}
	if p.PrescribedMedicationStatus != domain.YesNoUnknownNo {
		t.Errorf("expected prescribed_medication=no, got %v", p.PrescribedMedicationStatus)
	}
	if p.ImmunisationStatus != domain.ImmunisationUpToDate {
		t.Errorf("expected immunisation=up_to_date, got %v", p.ImmunisationStatus)
	}
	if p.MedicalDietaryReviewed != true {
		t.Errorf("expected medical_dietary_reviewed=true, got %v", p.MedicalDietaryReviewed)
	}
}

func TestMergePatch_ImmunisationStatusInvalid(t *testing.T) {
	p := newProfile()
	_, err := MergePatch(p, PatchSection{
		MedicalDietary: &MedicalDietaryPatch{
			ImmunisationStatus: yes("bogus"),
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid immunisation_status")
	}
}

func TestMergePatch_ContactEntryTrimsFullName(t *testing.T) {
	entries := []ContactEntryPatch{
		{FullName: "  Sarah Thompson  ", RelationshipToChild: yes("Mother"), Telephone: yes("+44 7700 900001")},
	}
	built := BuildContactEntries(domain.ContactTypeParentCarer, entries, uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil)
	if len(built) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(built))
	}
	if built[0].FullName != "Sarah Thompson" {
		t.Errorf("expected trimmed name 'Sarah Thompson', got '%s'", built[0].FullName)
	}
	if built[0].RelationshipToChild == nil || *built[0].RelationshipToChild != "Mother" {
		t.Errorf("expected relationship 'Mother', got %v", built[0].RelationshipToChild)
	}
	if built[0].Telephone == nil || *built[0].Telephone != "+44 7700 900001" {
		t.Errorf("expected telephone, got %v", built[0].Telephone)
	}
}

func TestMergePatch_ContactEntryPreservesOptionalFields(t *testing.T) {
	entries := []ContactEntryPatch{
		{FullName: "Jane Doe", RelationshipToChild: yes("Aunt"), Email: yes("jane@example.com"), WorkAddress: &map[string]any{"text": "Office"}},
	}
	built := BuildContactEntries(domain.ContactTypeEmergencyContact, entries, uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil)
	if len(built) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(built))
	}
	if built[0].Email == nil || *built[0].Email != "jane@example.com" {
		t.Errorf("expected email, got %v", built[0].Email)
	}
	if built[0].HasParentalResponsibility != nil {
		t.Errorf("expected nil parental responsibility, got %v", built[0].HasParentalResponsibility)
	}
}

func TestMergePatch_YesNoUnknownValidation(t *testing.T) {
	p := newProfile()
	_, err := MergePatch(p, PatchSection{
		DemographicsHome: &DemographicsHomePatch{
			DisabilityStatus: yes("nope"),
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid yes_no_unknown value")
	}
}

func TestMergePatch_MultipleSections(t *testing.T) {
	p := newProfile()

	changed, err := MergePatch(p, PatchSection{
		DemographicsHome: &DemographicsHomePatch{
			Sex:                      yes("male"),
			DisabilityStatus:         yes("no"),
			DemographicsHomeReviewed: boolPtr(true),
		},
		MedicalDietary: &MedicalDietaryPatch{
			MedicalConditionsStatus: yes("no"),
			MedicalDietaryReviewed:  boolPtr(true),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Sex == nil || *p.Sex != "male" {
		t.Errorf("expected sex=male, got %v", p.Sex)
	}
	if p.DisabilityStatus != domain.YesNoUnknownNo {
		t.Errorf("expected disability_status=no, got %v", p.DisabilityStatus)
	}
	if p.MedicalConditionsStatus != domain.YesNoUnknownNo {
		t.Errorf("expected medical_conditions=no, got %v", p.MedicalConditionsStatus)
	}

	foundDemo := false
	foundMed := false
	for _, s := range changed {
		if s == domain.SectionDemographicsHome {
			foundDemo = true
		}
		if s == domain.SectionMedicalDietary {
			foundMed = true
		}
	}
	if !foundDemo {
		t.Error("expected demographics_home in changed sections")
	}
	if !foundMed {
		t.Error("expected medical_dietary in changed sections")
	}
}

func TestMergePatch_CollectionPatch(t *testing.T) {
	p := newProfile()
	_, err := MergePatch(p, PatchSection{
		Collection: &CollectionPatch{
			Over18CollectionAcknowledged: boolPtr(true),
			EmergencyCollectionReviewed:  boolPtr(true),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Over18CollectionAcknowledged != true {
		t.Error("expected over18_collection_acknowledged=true")
	}
	if p.EmergencyCollectionReviewed != true {
		t.Error("expected emergency_collection_reviewed=true")
	}
}

func TestMergePatch_RoutineCare(t *testing.T) {
	p := newProfile()
	notes := "Child needs quiet time after lunch"
	_, err := MergePatch(p, PatchSection{
		RoutineCare: &RoutineCarePatch{
			RoutineCareNotes:    &notes,
			RoutineCareReviewed: boolPtr(true),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.RoutineCareNotes == nil || *p.RoutineCareNotes != "Child needs quiet time after lunch" {
		t.Errorf("expected routine care notes, got %v", p.RoutineCareNotes)
	}
}

func TestMergePatch_GDPRDeclarationDate(t *testing.T) {
	p := newProfile()
	dateStr := "2026-06-01"
	nameStr := "Manager Name"
	_, err := MergePatch(p, PatchSection{
		GDPRDeclaration: &GDPRDeclarationPatch{
			GDPRDeclaredByName:  &nameStr,
			GDPRDeclarationDate: &dateStr,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.GDPRDeclaredByName == nil || *p.GDPRDeclaredByName != "Manager Name" {
		t.Errorf("expected gdpr name, got %v", p.GDPRDeclaredByName)
	}
	if p.GDPRDeclarationDate == nil {
		t.Fatal("expected gdpr declaration date to be set")
	}
	expectedDate, _ := time.Parse("2006-01-02", "2026-06-01")
	if !p.GDPRDeclarationDate.Equal(expectedDate) {
		t.Errorf("expected %v, got %v", expectedDate, p.GDPRDeclarationDate)
	}
}

func TestMergePatch_GDPRDeclarationDateInvalid(t *testing.T) {
	p := newProfile()
	dateStr := "not-a-date"
	_, err := MergePatch(p, PatchSection{
		GDPRDeclaration: &GDPRDeclarationPatch{
			GDPRDeclarationDate: &dateStr,
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid GDPR date")
	}
}
