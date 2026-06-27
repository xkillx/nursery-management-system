package httpchild

import (
	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/modules/children/application"
	"nursery-management-system/api/internal/modules/children/domain"
)

type childProfileRequest struct {
	Sex                          *string        `json:"sex"`
	Religion                     *string        `json:"religion"`
	EthnicOrigin                 *string        `json:"ethnic_origin"`
	FirstLanguage                *string        `json:"first_language"`
	OtherLanguages               *string        `json:"other_languages"`
	HomeAddress                  map[string]any `json:"home_address"`
	HomePostcode                 *string        `json:"home_postcode"`
	HomeTelephone                *string        `json:"home_telephone"`
	DisabilityStatus             string         `json:"disability_status"`
	DisabilityNotes              *string        `json:"disability_notes"`
	AccessRequirements           *string        `json:"access_requirements"`
	RoutineCareNotes             *string        `json:"routine_care_notes"`
	GDPRDeclaredByName           *string        `json:"gdpr_declared_by_name"`
	GDPRDeclaredAt               *string        `json:"gdpr_declared_at"`
	GDPRDeclarationDate          *string        `json:"gdpr_declaration_date"`
	RegistrationDate             *string        `json:"registration_date"`
	DemographicsHomeReviewed     bool           `json:"demographics_home_reviewed"`
	MedicalDietaryReviewed       bool           `json:"medical_dietary_reviewed"`
	HealthContactsReviewed       bool           `json:"health_contacts_reviewed"`
	SocialDevelopmentReviewed    bool           `json:"social_development_reviewed"`
	ParentResponsibilityReviewed bool           `json:"parent_responsibility_reviewed"`
	EmergencyCollectionReviewed  bool           `json:"emergency_collection_reviewed"`
	RoutineCareReviewed          bool           `json:"routine_care_reviewed"`
}

type childProfileResponse struct {
	ID                  string         `json:"id"`
	ChildID             string         `json:"child_id"`
	Sex                 *string        `json:"sex,omitempty"`
	Religion            *string        `json:"religion,omitempty"`
	EthnicOrigin        *string        `json:"ethnic_origin,omitempty"`
	FirstLanguage       *string        `json:"first_language,omitempty"`
	OtherLanguages      *string        `json:"other_languages,omitempty"`
	HomeAddress         map[string]any `json:"home_address"`
	HomePostcode        *string        `json:"home_postcode,omitempty"`
	HomeTelephone       *string        `json:"home_telephone,omitempty"`
	DisabilityStatus    string         `json:"disability_status"`
	DisabilityNotes     *string        `json:"disability_notes,omitempty"`
	AccessRequirements  *string        `json:"access_requirements,omitempty"`
	RoutineCareNotes    *string        `json:"routine_care_notes,omitempty"`
	GDPRDeclaredByName  *string        `json:"gdpr_declared_by_name,omitempty"`
	GDPRDeclaredAt      *string        `json:"gdpr_declared_at,omitempty"`
	GDPRDeclarationDate *string        `json:"gdpr_declaration_date,omitempty"`
	RegistrationDate    *string        `json:"registration_date,omitempty"`

	DemographicsHomeReviewed     bool `json:"demographics_home_reviewed"`
	MedicalDietaryReviewed       bool `json:"medical_dietary_reviewed"`
	HealthContactsReviewed       bool `json:"health_contacts_reviewed"`
	SocialDevelopmentReviewed    bool `json:"social_development_reviewed"`
	ParentResponsibilityReviewed bool `json:"parent_responsibility_reviewed"`
	EmergencyCollectionReviewed  bool `json:"emergency_collection_reviewed"`
	RoutineCareReviewed          bool `json:"routine_care_reviewed"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func mapChildProfileRequest(req childProfileRequest) *application.ChildProfileInput {
	return &application.ChildProfileInput{
		Sex:                          req.Sex,
		Religion:                     req.Religion,
		EthnicOrigin:                 req.EthnicOrigin,
		FirstLanguage:                req.FirstLanguage,
		OtherLanguages:               req.OtherLanguages,
		HomeAddress:                  req.HomeAddress,
		HomePostcode:                 req.HomePostcode,
		HomeTelephone:                req.HomeTelephone,
		DisabilityStatus:             req.DisabilityStatus,
		DisabilityNotes:              req.DisabilityNotes,
		AccessRequirements:           req.AccessRequirements,
		RoutineCareNotes:             req.RoutineCareNotes,
		GDPRDeclaredByName:           req.GDPRDeclaredByName,
		GDPRDeclaredAt:               req.GDPRDeclaredAt,
		GDPRDeclarationDate:          req.GDPRDeclarationDate,
		RegistrationDate:             req.RegistrationDate,
		DemographicsHomeReviewed:     req.DemographicsHomeReviewed,
		MedicalDietaryReviewed:       req.MedicalDietaryReviewed,
		HealthContactsReviewed:       req.HealthContactsReviewed,
		SocialDevelopmentReviewed:    req.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed: req.ParentResponsibilityReviewed,
		EmergencyCollectionReviewed:  req.EmergencyCollectionReviewed,
		RoutineCareReviewed:          req.RoutineCareReviewed,
	}
}

func toChildProfileResponse(p *domain.ChildProfile) gin.H {
	if p == nil {
		return gin.H{"profile": nil}
	}
	return gin.H{"profile": childProfileResponse{
		ID:                           p.ID.String(),
		ChildID:                      p.ChildID.String(),
		Sex:                          p.Sex,
		Religion:                     p.Religion,
		EthnicOrigin:                 p.EthnicOrigin,
		FirstLanguage:                p.FirstLanguage,
		OtherLanguages:               p.OtherLanguages,
		HomeAddress:                  p.HomeAddress,
		HomePostcode:                 p.HomePostcode,
		HomeTelephone:                p.HomeTelephone,
		DisabilityStatus:             string(p.DisabilityStatus),
		DisabilityNotes:              p.DisabilityNotes,
		AccessRequirements:           p.AccessRequirements,
		RoutineCareNotes:             p.RoutineCareNotes,
		GDPRDeclaredByName:           p.GDPRDeclaredByName,
		GDPRDeclaredAt:               timeStringPtr(p.GDPRDeclaredAt),
		GDPRDeclarationDate:          dateStringPtr(p.GDPRDeclarationDate),
		RegistrationDate:             dateStringPtr(p.RegistrationDate),
		DemographicsHomeReviewed:     p.DemographicsHomeReviewed,
		MedicalDietaryReviewed:       p.MedicalDietaryReviewed,
		HealthContactsReviewed:       p.HealthContactsReviewed,
		SocialDevelopmentReviewed:    p.SocialDevelopmentReviewed,
		ParentResponsibilityReviewed: p.ParentResponsibilityReviewed,
		EmergencyCollectionReviewed:  p.EmergencyCollectionReviewed,
		RoutineCareReviewed:          p.RoutineCareReviewed,
		CreatedAt:                    p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:                    p.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}}
}
