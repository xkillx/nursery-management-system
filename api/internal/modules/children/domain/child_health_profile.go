package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChildHealthProfile struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	ChildID  uuid.UUID

	MedicalConditionsStatus    YesNoUnknown
	MedicalConditionsNotes     *string
	PrescribedMedicationStatus YesNoUnknown
	MedicationNotes            *string
	ImmunisationStatus         ImmunisationStatus
	ImmunisationCountry        *string
	IllnessDiagnosisHistory    *string
	DietaryRequirementsStatus  YesNoUnknown
	DietaryRequirementsNotes   *string
	DietarySideEffects         *string

	DoctorName           *string
	DoctorAddress        *string
	DoctorPhone          *string
	HealthVisitorName    *string
	HealthVisitorAddress *string
	HealthVisitorPhone   *string

	CreatedAt time.Time
	UpdatedAt time.Time
}
