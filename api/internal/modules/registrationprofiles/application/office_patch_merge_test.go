package application

import (
	"testing"

	domain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
)

func newOfficeChecklist() *domain.OfficeUseChecklist {
	return domain.DefaultOfficeUseChecklist()
}

func TestMergeOfficeChecklistPatch_AcceptsAllFields(t *testing.T) {
	c := newOfficeChecklist()

	depositDate := "2026-05-01"
	appDate := "2026-04-15"
	contractDate := "2026-05-10"
	handbookDate := "2026-05-12"
	redBookDate := "2026-05-14"
	birthCertDate := "2026-05-16"
	proofDate := "2026-05-18"
	sessionsText := "Mon/Wed/Fri mornings"
	notesText := "All documents verified"

	changed, err := MergeOfficeChecklistPatch(c, OfficeUseChecklistPatch{
		DepositStatus:                      yes("complete"),
		DepositPaidDate:                    &depositDate,
		ApplicationDateStatus:              yes("complete"),
		ApplicationDate:                    &appDate,
		StartDateStatus:                    yes("complete"),
		SessionsDaysRequestedStatus:        yes("complete"),
		SessionsDaysRequested:              &sessionsText,
		TermTimeOnlySpaceStatus:            yes("no"),
		ContractStatus:                     yes("complete"),
		ContractDate:                       &contractDate,
		HandbookStatus:                     yes("complete"),
		HandbookDate:                       &handbookDate,
		RedBookStatus:                      yes("complete"),
		RedBookCheckedDate:                 &redBookDate,
		BirthCertificatePassportStatus:     yes("complete"),
		BirthCertificatePassportCheckedDate: &birthCertDate,
		ProofOfAddressStatus:               yes("complete"),
		ProofOfAddressCheckedDate:          &proofDate,
		Notes:                              &notesText,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(changed) != 19 {
		t.Errorf("expected 19 changed fields, got %d", len(changed))
	}
	if c.DepositStatus != domain.OfficeCheckStatusComplete {
		t.Errorf("expected deposit=complete, got %v", c.DepositStatus)
	}
	if c.TermTimeOnlySpaceStatus != domain.TermTimeOnlyStatusNo {
		t.Errorf("expected term_time_only=no, got %v", c.TermTimeOnlySpaceStatus)
	}
	if c.SessionsDaysRequested == nil || *c.SessionsDaysRequested != "Mon/Wed/Fri mornings" {
		t.Errorf("expected sessions text, got %v", c.SessionsDaysRequested)
	}
}

func TestMergeOfficeChecklistPatch_TermTimeOnlyValidValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  domain.TermTimeOnlyStatus
	}{
		{"unknown", "unknown", domain.TermTimeOnlyStatusUnknown},
		{"yes", "yes", domain.TermTimeOnlyStatusYes},
		{"no", "no", domain.TermTimeOnlyStatusNo},
		{"not_applicable", "not_applicable", domain.TermTimeOnlyStatusNotApplicable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newOfficeChecklist()
			_, err := MergeOfficeChecklistPatch(c, OfficeUseChecklistPatch{
				TermTimeOnlySpaceStatus: yes(tt.value),
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.TermTimeOnlySpaceStatus != tt.want {
				t.Errorf("expected %v, got %v", tt.want, c.TermTimeOnlySpaceStatus)
			}
		})
	}
}

func TestMergeOfficeChecklistPatch_TermTimeOnlyInvalid(t *testing.T) {
	c := newOfficeChecklist()
	_, err := MergeOfficeChecklistPatch(c, OfficeUseChecklistPatch{
		TermTimeOnlySpaceStatus: yes("maybe"),
	})
	if err == nil {
		t.Fatal("expected error for invalid term_time_only status")
	}
}

func TestMergeOfficeChecklistPatch_OfficeCheckStatusValidValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  domain.OfficeCheckStatus
	}{
		{"unknown", "unknown", domain.OfficeCheckStatusUnknown},
		{"complete", "complete", domain.OfficeCheckStatusComplete},
		{"missing", "missing", domain.OfficeCheckStatusMissing},
		{"not_applicable", "not_applicable", domain.OfficeCheckStatusNotApplicable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newOfficeChecklist()
			_, err := MergeOfficeChecklistPatch(c, OfficeUseChecklistPatch{
				DepositStatus: yes(tt.value),
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.DepositStatus != tt.want {
				t.Errorf("expected %v, got %v", tt.want, c.DepositStatus)
			}
		})
	}
}

func TestMergeOfficeChecklistPatch_InvalidDate(t *testing.T) {
	c := newOfficeChecklist()
	badDate := "not-a-date"
	_, err := MergeOfficeChecklistPatch(c, OfficeUseChecklistPatch{
		DepositPaidDate: &badDate,
	})
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestMergeOfficeChecklistPatch_ClearsDateWithEmptyString(t *testing.T) {
	c := newOfficeChecklist()
	empty := ""
	_, err := MergeOfficeChecklistPatch(c, OfficeUseChecklistPatch{
		DateLeft: &empty,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DateLeft != nil {
		t.Error("expected date_left to be nil after clearing with empty string")
	}
}

func TestMergeOfficeChecklistPatch_ClearsNotesWithEmptyString(t *testing.T) {
	c := newOfficeChecklist()
	notes := "previous note"
	c.Notes = &notes
	empty := ""
	_, err := MergeOfficeChecklistPatch(c, OfficeUseChecklistPatch{
		Notes: &empty,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Notes != nil {
		t.Error("expected notes to be nil after clearing with empty string")
	}
}
