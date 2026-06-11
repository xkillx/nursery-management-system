package httpregistrationprofile

import (
	"time"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
)

type officeChildSummaryResponse struct {
	ID          string  `json:"id"`
	FullName    string  `json:"full_name"`
	DateOfBirth string  `json:"date_of_birth"`
	StartDate   *string `json:"start_date,omitempty"`
	EndDate     *string `json:"end_date,omitempty"`
}

type officeChecklistMetadataResponse struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type officeUseChecklistResponse struct {
	DepositStatus                      *string `json:"deposit_status,omitempty"`
	DepositPaidDate                    *string `json:"deposit_paid_date,omitempty"`
	ApplicationDateStatus              *string `json:"application_date_status,omitempty"`
	ApplicationDate                    *string `json:"application_date,omitempty"`
	StartDateStatus                    *string `json:"start_date_status,omitempty"`
	DateLeft                           *string `json:"date_left,omitempty"`
	SessionsDaysRequestedStatus        *string `json:"sessions_days_requested_status,omitempty"`
	SessionsDaysRequested              *string `json:"sessions_days_requested,omitempty"`
	TermTimeOnlySpaceStatus            *string `json:"term_time_only_space_status,omitempty"`
	ContractStatus                     *string `json:"contract_status,omitempty"`
	ContractDate                       *string `json:"contract_date,omitempty"`
	HandbookStatus                     *string `json:"handbook_status,omitempty"`
	HandbookDate                       *string `json:"handbook_date,omitempty"`
	RedBookStatus                      *string `json:"red_book_status,omitempty"`
	RedBookCheckedDate                 *string `json:"red_book_checked_date,omitempty"`
	BirthCertificatePassportStatus     *string `json:"birth_certificate_passport_status,omitempty"`
	BirthCertificatePassportCheckedDate *string `json:"birth_certificate_passport_checked_date,omitempty"`
	ProofOfAddressStatus               *string `json:"proof_of_address_status,omitempty"`
	ProofOfAddressCheckedDate          *string `json:"proof_of_address_checked_date,omitempty"`
	Notes                              *string `json:"notes,omitempty"`
}

type officeCompletenessItemResponse struct {
	Code          string   `json:"code"`
	Status        string   `json:"status"`
	Label         string   `json:"label"`
	MissingFields []string `json:"missing_fields,omitempty"`
}

type officeCompletenessResponse struct {
	IsComplete    bool                           `json:"is_complete"`
	MissingFields []string                       `json:"missing_fields,omitempty"`
	Items         []officeCompletenessItemResponse `json:"items"`
}

type registrationOfficeUseChecklistResponse struct {
	Child            officeChildSummaryResponse       `json:"child"`
	ChecklistExists  bool                             `json:"checklist_exists"`
	Checklist        *officeChecklistMetadataResponse `json:"checklist,omitempty"`
	OfficeUseChecklist *officeUseChecklistResponse    `json:"office_use_checklist"`
	Completeness     officeCompletenessResponse       `json:"completeness"`
}

func toOfficeChildSummary(c domain.OfficeChildSummary) officeChildSummaryResponse {
	r := officeChildSummaryResponse{
		ID:          c.ID.String(),
		FullName:    c.FullName,
		DateOfBirth: c.DateOfBirth.Format("2006-01-02"),
	}
	if c.StartDate != nil {
		s := c.StartDate.Format("2006-01-02")
		r.StartDate = &s
	}
	if c.EndDate != nil {
		s := c.EndDate.Format("2006-01-02")
		r.EndDate = &s
	}
	return r
}

func toOfficeChecklistResponse(owc domain.OfficeUseChecklistWithChild, comp domain.OfficeUseCompleteness) *registrationOfficeUseChecklistResponse {
	resp := &registrationOfficeUseChecklistResponse{
		Child:           toOfficeChildSummary(owc.Child),
		ChecklistExists: owc.ChecklistExists,
		Completeness:    toOfficeCompletenessResponse(comp),
	}

	if !owc.ChecklistExists || owc.Checklist == nil {
		resp.OfficeUseChecklist = &officeUseChecklistResponse{}
		return resp
	}

	c := owc.Checklist

	resp.Checklist = &officeChecklistMetadataResponse{
		ID:        c.ID.String(),
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	}

	resp.OfficeUseChecklist = &officeUseChecklistResponse{
		DepositStatus:                      statusPtr(string(c.DepositStatus)),
		DepositPaidDate:                    formatDatePtr(c.DepositPaidDate),
		ApplicationDateStatus:              statusPtr(string(c.ApplicationDateStatus)),
		ApplicationDate:                    formatDatePtr(c.ApplicationDate),
		StartDateStatus:                    statusPtr(string(c.StartDateStatus)),
		DateLeft:                           formatDatePtr(c.DateLeft),
		SessionsDaysRequestedStatus:        statusPtr(string(c.SessionsDaysRequestedStatus)),
		SessionsDaysRequested:              c.SessionsDaysRequested,
		TermTimeOnlySpaceStatus:            statusPtr(string(c.TermTimeOnlySpaceStatus)),
		ContractStatus:                     statusPtr(string(c.ContractStatus)),
		ContractDate:                       formatDatePtr(c.ContractDate),
		HandbookStatus:                     statusPtr(string(c.HandbookStatus)),
		HandbookDate:                       formatDatePtr(c.HandbookDate),
		RedBookStatus:                      statusPtr(string(c.RedBookStatus)),
		RedBookCheckedDate:                 formatDatePtr(c.RedBookCheckedDate),
		BirthCertificatePassportStatus:     statusPtr(string(c.BirthCertificatePassportStatus)),
		BirthCertificatePassportCheckedDate: formatDatePtr(c.BirthCertificatePassportCheckedDate),
		ProofOfAddressStatus:               statusPtr(string(c.ProofOfAddressStatus)),
		ProofOfAddressCheckedDate:          formatDatePtr(c.ProofOfAddressCheckedDate),
		Notes:                              c.Notes,
	}

	return resp
}

func toOfficeCompletenessResponse(c domain.OfficeUseCompleteness) officeCompletenessResponse {
	items := make([]officeCompletenessItemResponse, len(c.Items))
	for i, item := range c.Items {
		items[i] = officeCompletenessItemResponse{
			Code:          string(item.Code),
			Status:        string(item.Status),
			Label:         item.Label,
			MissingFields: item.MissingFields,
		}
	}
	missingStr := make([]string, len(c.MissingFields))
	for i, mf := range c.MissingFields {
		missingStr[i] = string(mf)
	}
	return officeCompletenessResponse{
		IsComplete:    c.IsComplete,
		MissingFields: missingStr,
		Items:         items,
	}
}

func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
