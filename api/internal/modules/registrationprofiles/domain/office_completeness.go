package domain

type OfficeChecklistItemCode string

const (
	OfficeItemDeposit             OfficeChecklistItemCode = "deposit"
	OfficeItemApplicationDate     OfficeChecklistItemCode = "application_date"
	OfficeItemStartDateCheck      OfficeChecklistItemCode = "start_date_check"
	OfficeItemSessionsDays        OfficeChecklistItemCode = "sessions_days_requested"
	OfficeItemTermTimeOnly        OfficeChecklistItemCode = "term_time_only_space"
	OfficeItemContract            OfficeChecklistItemCode = "contract"
	OfficeItemHandbook            OfficeChecklistItemCode = "handbook"
	OfficeItemRedBook             OfficeChecklistItemCode = "red_book"
	OfficeItemBirthCertPassport   OfficeChecklistItemCode = "birth_certificate_passport"
	OfficeItemProofOfAddress      OfficeChecklistItemCode = "proof_of_address"
)

type OfficeUseCompletenessItem struct {
	Code          OfficeChecklistItemCode `json:"code"`
	Status        CompletenessStatus      `json:"status"`
	Label         string                   `json:"label"`
	MissingFields []string                 `json:"missing_fields,omitempty"`
}

type OfficeUseCompleteness struct {
	IsComplete    bool                        `json:"is_complete"`
	MissingFields []OfficeChecklistItemCode   `json:"missing_fields,omitempty"`
	Items         []OfficeUseCompletenessItem `json:"items"`
}

func ComputeOfficeUseCompleteness(c *OfficeUseChecklist) OfficeUseCompleteness {
	items := []OfficeUseCompletenessItem{
		computeOfficeItemDeposit(c),
		computeOfficeItemApplicationDate(c),
		computeOfficeItemStartDateCheck(c),
		computeOfficeItemSessionsDays(c),
		computeOfficeItemTermTimeOnly(c),
		computeOfficeItemContract(c),
		computeOfficeItemHandbook(c),
		computeOfficeItemRedBook(c),
		computeOfficeItemBirthCertPassport(c),
		computeOfficeItemProofOfAddress(c),
	}

	missing := make([]OfficeChecklistItemCode, 0)
	for _, item := range items {
		if item.Status == StatusIncomplete {
			missing = append(missing, item.Code)
		}
	}

	return OfficeUseCompleteness{
		IsComplete:    len(missing) == 0,
		MissingFields: missing,
		Items:         items,
	}
}

func computeOfficeItem(c OfficeCheckStatus, code OfficeChecklistItemCode, label string, extraMissing ...string) OfficeUseCompletenessItem {
	item := OfficeUseCompletenessItem{
		Code:  code,
		Label: label,
	}
	if c == OfficeCheckStatusComplete || c == OfficeCheckStatusNotApplicable {
		item.Status = StatusComplete
	} else {
		item.Status = StatusIncomplete
		if c == OfficeCheckStatusUnknown {
			item.MissingFields = append(item.MissingFields, string(code)+"_unknown")
		}
		if c == OfficeCheckStatusMissing {
			item.MissingFields = append(item.MissingFields, string(code)+"_missing")
		}
		item.MissingFields = append(item.MissingFields, extraMissing...)
	}
	return item
}

func computeOfficeItemDeposit(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	return computeOfficeItem(c.DepositStatus, OfficeItemDeposit, "Deposit")
}

func computeOfficeItemApplicationDate(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	item := OfficeUseCompletenessItem{
		Code:  OfficeItemApplicationDate,
		Label: "Application date",
	}
	if c.ApplicationDateStatus == OfficeCheckStatusComplete {
		if c.ApplicationDate == nil {
			item.Status = StatusIncomplete
			item.MissingFields = append(item.MissingFields, "application_date_required")
		} else {
			item.Status = StatusComplete
		}
	} else if c.ApplicationDateStatus == OfficeCheckStatusNotApplicable {
		item.Status = StatusComplete
	} else {
		item.Status = StatusIncomplete
		if c.ApplicationDateStatus == OfficeCheckStatusUnknown {
			item.MissingFields = append(item.MissingFields, "application_date_unknown")
		}
		if c.ApplicationDateStatus == OfficeCheckStatusMissing {
			item.MissingFields = append(item.MissingFields, "application_date_missing")
		}
	}
	return item
}

func computeOfficeItemStartDateCheck(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	return computeOfficeItem(c.StartDateStatus, OfficeItemStartDateCheck, "Start date check")
}

func computeOfficeItemSessionsDays(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	item := OfficeUseCompletenessItem{
		Code:  OfficeItemSessionsDays,
		Label: "Sessions/days requested",
	}
	if c.SessionsDaysRequestedStatus == OfficeCheckStatusComplete {
		if c.SessionsDaysRequested == nil || *c.SessionsDaysRequested == "" {
			item.Status = StatusIncomplete
			item.MissingFields = append(item.MissingFields, "sessions_days_requested_required")
		} else {
			item.Status = StatusComplete
		}
	} else if c.SessionsDaysRequestedStatus == OfficeCheckStatusNotApplicable {
		item.Status = StatusComplete
	} else {
		item.Status = StatusIncomplete
		if c.SessionsDaysRequestedStatus == OfficeCheckStatusUnknown {
			item.MissingFields = append(item.MissingFields, "sessions_days_requested_unknown")
		}
		if c.SessionsDaysRequestedStatus == OfficeCheckStatusMissing {
			item.MissingFields = append(item.MissingFields, "sessions_days_requested_missing")
		}
	}
	return item
}

func computeOfficeItemTermTimeOnly(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	item := OfficeUseCompletenessItem{
		Code:  OfficeItemTermTimeOnly,
		Label: "Term-time-only space",
	}
	if c.TermTimeOnlySpaceStatus == TermTimeOnlyStatusUnknown {
		item.Status = StatusIncomplete
		item.MissingFields = []string{"term_time_only_space_unknown"}
	} else {
		item.Status = StatusComplete
	}
	return item
}

func computeOfficeItemContract(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	return computeOfficeItem(c.ContractStatus, OfficeItemContract, "Contract/signature")
}

func computeOfficeItemHandbook(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	return computeOfficeItem(c.HandbookStatus, OfficeItemHandbook, "Handbook")
}

func computeOfficeItemRedBook(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	return computeOfficeItem(c.RedBookStatus, OfficeItemRedBook, "Red Book")
}

func computeOfficeItemBirthCertPassport(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	return computeOfficeItem(c.BirthCertificatePassportStatus, OfficeItemBirthCertPassport, "Birth certificate/passport")
}

func computeOfficeItemProofOfAddress(c *OfficeUseChecklist) OfficeUseCompletenessItem {
	return computeOfficeItem(c.ProofOfAddressStatus, OfficeItemProofOfAddress, "Proof of address")
}
