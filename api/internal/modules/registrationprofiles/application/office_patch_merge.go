package application

import (
	"strings"
	"time"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type OfficeUseChecklistPatch struct {
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

func MergeOfficeChecklistPatch(c *domain.OfficeUseChecklist, patch OfficeUseChecklistPatch) ([]string, error) {
	changed := make([]string, 0)

	if patch.DepositStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.DepositStatus)
		if err != nil {
			return nil, err
		}
		c.DepositStatus = v
		changed = append(changed, "deposit_status")
	}
	if patch.DepositPaidDate != nil {
		if *patch.DepositPaidDate == "" {
			c.DepositPaidDate = nil
		} else {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.DepositPaidDate))
			if err != nil {
				return nil, domainerrors.Validation("Invalid request payload.", "deposit_paid_date")
			}
			c.DepositPaidDate = &t
		}
		changed = append(changed, "deposit_paid_date")
	}
	if patch.ApplicationDateStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.ApplicationDateStatus)
		if err != nil {
			return nil, err
		}
		c.ApplicationDateStatus = v
		changed = append(changed, "application_date_status")
	}
	if patch.ApplicationDate != nil {
		if *patch.ApplicationDate == "" {
			c.ApplicationDate = nil
		} else {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.ApplicationDate))
			if err != nil {
				return nil, domainerrors.Validation("Invalid request payload.", "application_date")
			}
			c.ApplicationDate = &t
		}
		changed = append(changed, "application_date")
	}
	if patch.StartDateStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.StartDateStatus)
		if err != nil {
			return nil, err
		}
		c.StartDateStatus = v
		changed = append(changed, "start_date_status")
	}
	if patch.DateLeft != nil {
		if *patch.DateLeft == "" {
			c.DateLeft = nil
		} else {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.DateLeft))
			if err != nil {
				return nil, domainerrors.Validation("Invalid request payload.", "date_left")
			}
			c.DateLeft = &t
		}
		changed = append(changed, "date_left")
	}
	if patch.SessionsDaysRequestedStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.SessionsDaysRequestedStatus)
		if err != nil {
			return nil, err
		}
		c.SessionsDaysRequestedStatus = v
		changed = append(changed, "sessions_days_requested_status")
	}
	if patch.SessionsDaysRequested != nil {
		v := strings.TrimSpace(*patch.SessionsDaysRequested)
		if v == "" {
			c.SessionsDaysRequested = nil
		} else {
			c.SessionsDaysRequested = &v
		}
		changed = append(changed, "sessions_days_requested")
	}
	if patch.TermTimeOnlySpaceStatus != nil {
		v, err := parseTermTimeOnlyStatus(*patch.TermTimeOnlySpaceStatus)
		if err != nil {
			return nil, err
		}
		c.TermTimeOnlySpaceStatus = v
		changed = append(changed, "term_time_only_space_status")
	}
	if patch.ContractStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.ContractStatus)
		if err != nil {
			return nil, err
		}
		c.ContractStatus = v
		changed = append(changed, "contract_status")
	}
	if patch.ContractDate != nil {
		if *patch.ContractDate == "" {
			c.ContractDate = nil
		} else {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.ContractDate))
			if err != nil {
				return nil, domainerrors.Validation("Invalid request payload.", "contract_date")
			}
			c.ContractDate = &t
		}
		changed = append(changed, "contract_date")
	}
	if patch.HandbookStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.HandbookStatus)
		if err != nil {
			return nil, err
		}
		c.HandbookStatus = v
		changed = append(changed, "handbook_status")
	}
	if patch.HandbookDate != nil {
		if *patch.HandbookDate == "" {
			c.HandbookDate = nil
		} else {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.HandbookDate))
			if err != nil {
				return nil, domainerrors.Validation("Invalid request payload.", "handbook_date")
			}
			c.HandbookDate = &t
		}
		changed = append(changed, "handbook_date")
	}
	if patch.RedBookStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.RedBookStatus)
		if err != nil {
			return nil, err
		}
		c.RedBookStatus = v
		changed = append(changed, "red_book_status")
	}
	if patch.RedBookCheckedDate != nil {
		if *patch.RedBookCheckedDate == "" {
			c.RedBookCheckedDate = nil
		} else {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.RedBookCheckedDate))
			if err != nil {
				return nil, domainerrors.Validation("Invalid request payload.", "red_book_checked_date")
			}
			c.RedBookCheckedDate = &t
		}
		changed = append(changed, "red_book_checked_date")
	}
	if patch.BirthCertificatePassportStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.BirthCertificatePassportStatus)
		if err != nil {
			return nil, err
		}
		c.BirthCertificatePassportStatus = v
		changed = append(changed, "birth_certificate_passport_status")
	}
	if patch.BirthCertificatePassportCheckedDate != nil {
		if *patch.BirthCertificatePassportCheckedDate == "" {
			c.BirthCertificatePassportCheckedDate = nil
		} else {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.BirthCertificatePassportCheckedDate))
			if err != nil {
				return nil, domainerrors.Validation("Invalid request payload.", "birth_certificate_passport_checked_date")
			}
			c.BirthCertificatePassportCheckedDate = &t
		}
		changed = append(changed, "birth_certificate_passport_checked_date")
	}
	if patch.ProofOfAddressStatus != nil {
		v, err := parseOfficeCheckStatus(*patch.ProofOfAddressStatus)
		if err != nil {
			return nil, err
		}
		c.ProofOfAddressStatus = v
		changed = append(changed, "proof_of_address_status")
	}
	if patch.ProofOfAddressCheckedDate != nil {
		if *patch.ProofOfAddressCheckedDate == "" {
			c.ProofOfAddressCheckedDate = nil
		} else {
			t, err := time.Parse("2006-01-02", strings.TrimSpace(*patch.ProofOfAddressCheckedDate))
			if err != nil {
				return nil, domainerrors.Validation("Invalid request payload.", "proof_of_address_checked_date")
			}
			c.ProofOfAddressCheckedDate = &t
		}
		changed = append(changed, "proof_of_address_checked_date")
	}
	if patch.Notes != nil {
		v := strings.TrimSpace(*patch.Notes)
		if v == "" {
			c.Notes = nil
		} else {
			c.Notes = &v
		}
		changed = append(changed, "notes")
	}

	return changed, nil
}

func parseOfficeCheckStatus(v string) (domain.OfficeCheckStatus, error) {
	switch strings.TrimSpace(v) {
	case "unknown":
		return domain.OfficeCheckStatusUnknown, nil
	case "complete":
		return domain.OfficeCheckStatusComplete, nil
	case "missing":
		return domain.OfficeCheckStatusMissing, nil
	case "not_applicable":
		return domain.OfficeCheckStatusNotApplicable, nil
	default:
		return "", domainerrors.Validation("Invalid request payload.", "office_check_status")
	}
}

func parseTermTimeOnlyStatus(v string) (domain.TermTimeOnlyStatus, error) {
	switch strings.TrimSpace(v) {
	case "unknown":
		return domain.TermTimeOnlyStatusUnknown, nil
	case "yes":
		return domain.TermTimeOnlyStatusYes, nil
	case "no":
		return domain.TermTimeOnlyStatusNo, nil
	case "not_applicable":
		return domain.TermTimeOnlyStatusNotApplicable, nil
	default:
		return "", domainerrors.Validation("Invalid request payload.", "term_time_only_status")
	}
}
