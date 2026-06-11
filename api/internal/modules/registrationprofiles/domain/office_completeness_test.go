package domain

import (
	"testing"
	"time"
)

func TestComputeOfficeUseCompleteness_AllUnknownIsIncomplete(t *testing.T) {
	c := DefaultOfficeUseChecklist()
	comp := ComputeOfficeUseCompleteness(c)
	if comp.IsComplete {
		t.Error("expected all-unknown checklist to be incomplete")
	}
	if len(comp.MissingFields) != 10 {
		t.Errorf("expected 10 missing fields, got %d: %v", len(comp.MissingFields), comp.MissingFields)
	}
}

func TestComputeOfficeUseCompleteness_NotApplicableCompletesStandardItems(t *testing.T) {
	c := DefaultOfficeUseChecklist()
	c.DepositStatus = OfficeCheckStatusNotApplicable
	c.ApplicationDateStatus = OfficeCheckStatusNotApplicable
	c.StartDateStatus = OfficeCheckStatusNotApplicable
	c.SessionsDaysRequestedStatus = OfficeCheckStatusNotApplicable
	c.ContractStatus = OfficeCheckStatusNotApplicable
	c.HandbookStatus = OfficeCheckStatusNotApplicable
	c.RedBookStatus = OfficeCheckStatusNotApplicable
	c.BirthCertificatePassportStatus = OfficeCheckStatusNotApplicable
	c.ProofOfAddressStatus = OfficeCheckStatusNotApplicable
	c.TermTimeOnlySpaceStatus = TermTimeOnlyStatusNotApplicable

	comp := ComputeOfficeUseCompleteness(c)
	if !comp.IsComplete {
		t.Error("expected checklist with all not_applicable to be complete")
	}
}

func TestComputeOfficeUseCompleteness_ApplicationDateRequiresDateWhenComplete(t *testing.T) {
	c := DefaultOfficeUseChecklist()
	c.ApplicationDateStatus = OfficeCheckStatusComplete
	c.ApplicationDate = nil

	comp := ComputeOfficeUseCompleteness(c)
	item := findItem(comp.Items, OfficeItemApplicationDate)
	if item == nil {
		t.Fatal("expected application_date item")
	}
	if item.Status != StatusIncomplete {
		t.Error("expected application_date to be incomplete without date when status is complete")
	}
	if len(item.MissingFields) == 0 || item.MissingFields[0] != "application_date_required" {
		t.Errorf("expected application_date_required missing field, got %v", item.MissingFields)
	}
}

func TestComputeOfficeUseCompleteness_ApplicationDateCompleteWithDate(t *testing.T) {
	c := DefaultOfficeUseChecklist()
	now := time.Now()
	c.ApplicationDateStatus = OfficeCheckStatusComplete
	c.ApplicationDate = &now

	comp := ComputeOfficeUseCompleteness(c)
	item := findItem(comp.Items, OfficeItemApplicationDate)
	if item == nil {
		t.Fatal("expected application_date item")
	}
	if item.Status != StatusComplete {
		t.Error("expected application_date to be complete when status is complete and date provided")
	}
}

func TestComputeOfficeUseCompleteness_SessionsDaysRequiresTextWhenComplete(t *testing.T) {
	c := DefaultOfficeUseChecklist()
	c.SessionsDaysRequestedStatus = OfficeCheckStatusComplete
	c.SessionsDaysRequested = nil

	comp := ComputeOfficeUseCompleteness(c)
	item := findItem(comp.Items, OfficeItemSessionsDays)
	if item == nil {
		t.Fatal("expected sessions_days_requested item")
	}
	if item.Status != StatusIncomplete {
		t.Error("expected sessions_days to be incomplete without text when status is complete")
	}
	if len(item.MissingFields) == 0 || item.MissingFields[0] != "sessions_days_requested_required" {
		t.Errorf("expected sessions_days_requested_required missing field, got %v", item.MissingFields)
	}
}

func TestComputeOfficeUseCompleteness_SessionsDaysCompleteWithText(t *testing.T) {
	c := DefaultOfficeUseChecklist()
	text := "Mon/Wed/Fri"
	c.SessionsDaysRequestedStatus = OfficeCheckStatusComplete
	c.SessionsDaysRequested = &text

	comp := ComputeOfficeUseCompleteness(c)
	item := findItem(comp.Items, OfficeItemSessionsDays)
	if item == nil {
		t.Fatal("expected sessions_days_requested item")
	}
	if item.Status != StatusComplete {
		t.Error("expected sessions_days to be complete when status is complete and text provided")
	}
}

func TestComputeOfficeUseCompleteness_DateLeftOptional(t *testing.T) {
	c := DefaultOfficeUseChecklist()
	c.StartDateStatus = OfficeCheckStatusComplete
	c.DateLeft = nil

	item := computeOfficeItemStartDateCheck(c)
	if item.Status != StatusComplete {
		t.Error("expected start_date_check to be complete even without date_left")
	}
}

func TestComputeOfficeUseCompleteness_MissingStatus(t *testing.T) {
	c := DefaultOfficeUseChecklist()
	c.DepositStatus = OfficeCheckStatusMissing

	comp := ComputeOfficeUseCompleteness(c)
	item := findItem(comp.Items, OfficeItemDeposit)
	if item == nil {
		t.Fatal("expected deposit item")
	}
	if item.Status != StatusIncomplete {
		t.Error("expected deposit to be incomplete when status is missing")
	}
	if len(item.MissingFields) == 0 || item.MissingFields[0] != "deposit_missing" {
		t.Errorf("expected deposit_missing, got %v", item.MissingFields)
	}
}

func findItem(items []OfficeUseCompletenessItem, code OfficeChecklistItemCode) *OfficeUseCompletenessItem {
	for i := range items {
		if items[i].Code == code {
			return &items[i]
		}
	}
	return nil
}
