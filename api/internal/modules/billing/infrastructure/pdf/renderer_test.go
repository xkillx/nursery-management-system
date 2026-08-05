package pdf

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"nursery-management-system/api/internal/modules/billing/domain"
)

func TestRenderer_Render_HappyPath(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	qty1 := 480
	qty2 := 600
	unit1 := 500
	unit2 := 600
	sess1 := 4
	sess2 := 5

	input := InvoicePDFInput{
		SiteProfile: InvoicePDFSiteProfile{
			NurseryName:     "Little Explorers Nursery",
			Phone:           "020 7123 4567",
			Email:           "info@littleexplorers.co.uk",
			AddressStreet:   "123 High Street",
			AddressCity:     "London",
			AddressPostcode: "SW1A 1AA",
		},
		Invoice: InvoicePDFMeta{
			InvoiceNumber: "INV-2026-001",
			BillingMonth:  time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			IssueDate:     timePtr(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)),
			DueDate:       timePtr(time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)),
			Status:        "issued",
		},
		Parent: InvoicePDFParent{
			FullName:        "Jane Smith",
			AddressLine1:    "45 Oak Avenue",
			AddressCity:     "London",
			AddressPostcode: "EC1A 1BB",
		},
		Lines: []InvoicePDFLine{
			{
				Description:     "Core sessions (480 min @ \u00a35.00/hr)",
				QuantityMinutes: &qty1,
				SessionCount:    &sess1,
				UnitAmountMinor: &unit1,
				LineAmountMinor: 4000,
			},
			{
				Description:     "Extra sessions (600 min @ \u00a36.00/hr)",
				QuantityMinutes: &qty2,
				SessionCount:    &sess2,
				UnitAmountMinor: &unit2,
				LineAmountMinor: 6000,
			},
			{
				Description:     "Funded deduction",
				LineAmountMinor: 2000,
				IsFunded:        true,
			},
		},
		SubtotalMinor:  10000,
		DeductionMinor: 2000,
		TotalMinor:     8000,
		PaymentNote:    "Please settle by the due date.",
	}

	pdfBytes, err := r.Render(context.Background(), input)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatal("expected PDF header")
	}
}

func TestRenderer_Render_EmptyLines(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	input := InvoicePDFInput{
		SiteProfile: InvoicePDFSiteProfile{
			NurseryName: "Test Nursery",
		},
		Invoice: InvoicePDFMeta{
			InvoiceNumber: "INV-001",
			BillingMonth:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:        "draft",
		},
		Parent: InvoicePDFParent{
			FullName: "Test Parent",
		},
		Lines:         []InvoicePDFLine{},
		SubtotalMinor: 0,
		TotalMinor:    0,
	}

	pdfBytes, err := r.Render(context.Background(), input)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if len(pdfBytes) == 0 || !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatal("expected valid PDF")
	}
}

func TestRenderer_Render_MoneyFormatting(t *testing.T) {
	tests := []struct {
		minor    int
		expected string
	}{
		{12500, "\u00a3125.00"},
		{0, "\u00a30.00"},
		{1, "\u00a30.01"},
		{100, "\u00a31.00"},
		{-500, "-\u00a35.00"},
	}

	for _, tt := range tests {
		got := formatMoney(tt.minor)
		if got != tt.expected {
			t.Errorf("formatMoney(%d) = %q, want %q", tt.minor, got, tt.expected)
		}
	}
}

func TestAppendPDFLines_SessionExpansion(t *testing.T) {
	unit := domain.MustGBP(600)
	lines := appendPDFLines(nil, domain.LineKindCoreChildcare, "Core childcare", nil, &unit, 12000, nil, []domain.SessionRow{
		{OccurrenceDate: timeMustParsePdf("2026-11-02"), StartMinutes: 480, EndMinutes: 720, DurationMinutes: 240, SessionTypeName: "Morning Session", SessionAmountMinor: 6000},
		{OccurrenceDate: timeMustParsePdf("2026-11-09"), StartMinutes: 480, EndMinutes: 720, DurationMinutes: 240, SessionTypeName: "Morning Session", SessionAmountMinor: 6000},
	}, false)

	if len(lines) != 2 {
		t.Fatalf("got %d PDF lines, want 2", len(lines))
	}
	if lines[0].Description != "Mon 2 Nov · Morning Session (08:00–12:00)" {
		t.Errorf("line 0 description = %q, want 'Mon 2 Nov · Morning Session (08:00–12:00)'", lines[0].Description)
	}
	if lines[0].SubDescription != "Standard Weekly / Monthly Session Rate" {
		t.Errorf("sub-description = %q, want 'Standard Weekly / Monthly Session Rate'", lines[0].SubDescription)
	}
	if lines[0].QuantityMinutes == nil || *lines[0].QuantityMinutes != 240 {
		t.Errorf("quantity minutes = %v, want 240", lines[0].QuantityMinutes)
	}
	if lines[0].UnitAmountMinor == nil || *lines[0].UnitAmountMinor != 600 {
		t.Errorf("unit amount = %v, want 600", lines[0].UnitAmountMinor)
	}
	if lines[0].LineAmountMinor != 6000 {
		t.Errorf("line amount = %d, want 6000", lines[0].LineAmountMinor)
	}
	if lines[0].SessionDate != "Mon 2 Nov" {
		t.Errorf("session date = %q, want 'Mon 2 Nov'", lines[0].SessionDate)
	}
	if lines[1].LineAmountMinor != 6000 {
		t.Errorf("line 1 amount = %d, want 6000", lines[1].LineAmountMinor)
	}
}

func TestAppendPDFLines_SessionNameOnlyFallback(t *testing.T) {
	unit := domain.MustGBP(600)
	lines := appendPDFLines(nil, domain.LineKindCoreChildcare, "Core childcare", nil, &unit, 4000, nil, []domain.SessionRow{
		{OccurrenceDate: timeMustParsePdf("2026-11-02"), DurationMinutes: 240, SessionTypeName: "Legacy Session", SessionAmountMinor: 4000},
	}, false)

	if len(lines) != 1 {
		t.Fatalf("got %d PDF lines, want 1", len(lines))
	}
	if lines[0].Description != "Mon 2 Nov · Legacy Session" {
		t.Errorf("description = %q, want 'Mon 2 Nov · Legacy Session'", lines[0].Description)
	}
}

func TestAppendPDFLines_LegacySingleRow(t *testing.T) {
	qty := 480
	lines := appendPDFLines(nil, domain.LineKindCoreChildcare, "Core childcare", &qty, nil, 4000, nil, nil, false)
	if len(lines) != 1 {
		t.Fatalf("got %d PDF lines, want 1 (legacy aggregate)", len(lines))
	}
	if lines[0].Description != "Core childcare" {
		t.Errorf("description = %q, want 'Core childcare'", lines[0].Description)
	}
	if lines[0].QuantityMinutes == nil || *lines[0].QuantityMinutes != 480 {
		t.Errorf("quantity = %v, want 480", lines[0].QuantityMinutes)
	}
}

func TestAppendPDFLines_FundedLineUnaffected(t *testing.T) {
	lines := appendPDFLines(nil, domain.LineKindFundedDeduction, "Funded hours deduction", nil, nil, -2500, nil, nil, true)
	if len(lines) != 1 {
		t.Fatalf("got %d PDF lines, want 1", len(lines))
	}
	if !lines[0].IsFunded {
		t.Error("expected funded line to be flagged IsFunded")
	}
	if lines[0].LineAmountMinor != -2500 {
		t.Errorf("line amount = %d, want -2500", lines[0].LineAmountMinor)
	}
}

func TestFormatQuantity_SessionHours(t *testing.T) {
	mins := 240
	if got := formatQuantity(InvoicePDFLine{QuantityMinutes: &mins}); got != "4 hrs" {
		t.Errorf("formatQuantity(240) = %q, want '4 hrs'", got)
	}
	mins2 := 210
	if got := formatQuantity(InvoicePDFLine{QuantityMinutes: &mins2}); got != "3.5 hrs" {
		t.Errorf("formatQuantity(210) = %q, want '3.5 hrs'", got)
	}
}

func timeMustParsePdf(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestRenderer_Render_LongDescription(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	longDesc := strings.Repeat("This is a very long description for a line item. ", 5)

	input := InvoicePDFInput{
		SiteProfile: InvoicePDFSiteProfile{
			NurseryName: "Test Nursery",
		},
		Invoice: InvoicePDFMeta{
			InvoiceNumber: "INV-002",
			BillingMonth:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:        "issued",
		},
		Parent: InvoicePDFParent{FullName: "Test Parent"},
		Lines: []InvoicePDFLine{
			{Description: longDesc, LineAmountMinor: 5000},
		},
		SubtotalMinor: 5000,
		TotalMinor:    5000,
	}

	pdfBytes, err := r.Render(context.Background(), input)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatal("expected valid PDF")
	}
}

func TestRenderer_Render_ManyLines(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	lines := make([]InvoicePDFLine, 25)
	for i := range lines {
		lines[i] = InvoicePDFLine{
			Description:     "Line item",
			LineAmountMinor: 1000,
		}
	}

	input := InvoicePDFInput{
		SiteProfile: InvoicePDFSiteProfile{NurseryName: "Test"},
		Invoice: InvoicePDFMeta{
			InvoiceNumber: "INV-003",
			BillingMonth:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:        "issued",
		},
		Parent:        InvoicePDFParent{FullName: "Parent"},
		Lines:         lines,
		SubtotalMinor: 25000,
		TotalMinor:    25000,
	}

	pdfBytes, err := r.Render(context.Background(), input)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatal("expected valid PDF")
	}
}

func TestRenderer_Render_StatusValues(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	statuses := []string{"draft", "issued", "paid", "overdue", "void"}
	for _, status := range statuses {
		input := InvoicePDFInput{
			SiteProfile: InvoicePDFSiteProfile{NurseryName: "Test"},
			Invoice: InvoicePDFMeta{
				InvoiceNumber: "INV-001",
				BillingMonth:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Status:        status,
			},
			Parent:     InvoicePDFParent{FullName: "Parent"},
			Lines:      []InvoicePDFLine{},
			TotalMinor: 0,
		}

		pdfBytes, err := r.Render(context.Background(), input)
		if err != nil {
			t.Errorf("status %q: Render: %v", status, err)
			continue
		}
		if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
			t.Errorf("status %q: expected valid PDF", status)
		}
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestRenderer_Render_NurseryProDesign(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	unit1 := 22500
	unit2 := 1500
	unit3 := 1050
	unit4 := 3750
	qty1 := 9600
	qty2 := 720
	qty3 := 3600

	input := InvoicePDFInput{
		SiteProfile: InvoicePDFSiteProfile{
			NurseryName:     "NurseryPro Central",
			Phone:           "020 8822 0033",
			Email:           "admin@nurserypro.co.uk",
			AddressStreet:   "88 Education Square",
			AddressCity:     "Bloomsbury",
			AddressPostcode: "WC1N 1EX",
			RegNumber:       "12993844",
			VATNumber:       "GB 992 1122 33",
			BankName:        "Barclays UK",
			SortCode:        "20-44-55",
			AccountNumber:   "88220033",
		},
		Invoice: InvoicePDFMeta{
			InvoiceNumber: "INV-2024-0892",
			BillingMonth:  time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC),
			IssueDate:     timePtr(time.Date(2023, 10, 12, 0, 0, 0, 0, time.UTC)),
			DueDate:       timePtr(time.Date(2023, 10, 26, 0, 0, 0, 0, time.UTC)),
			Status:        "PAID",
		},
		Parent: InvoicePDFParent{
			FullName:        "Mr. David Harrison",
			AddressLine1:    "42 Nightingale Lane",
			AddressCity:     "London",
			AddressPostcode: "SW12 8TH",
			Email:           "david.h@example.com",
		},
		Child: InvoicePDFChild{
			ChildName: "Leo Harrison",
			RoomName:  "Pre-School Room",
		},
		Lines: []InvoicePDFLine{
			{
				Description:     "Full-Time Childcare",
				SubDescription:  "Standard Weekly Rate - October Week 1-4",
				QuantityMinutes: &qty1,
				UnitAmountMinor: &unit1,
				LineAmountMinor: 90000,
			},
			{
				Description:     "After-School Club",
				SubDescription:  "Extended hours: 17:30 - 18:30",
				QuantityMinutes: &qty2,
				UnitAmountMinor: &unit2,
				LineAmountMinor: 18000,
			},
			{
				Description:     "Universal 15 Hours Funding",
				SubDescription:  "Early Years Funding Offset",
				QuantityMinutes: &qty3,
				UnitAmountMinor: &unit3,
				LineAmountMinor: 63000,
				IsFunded:        true,
			},
			{
				Description:     "Sibling Discount",
				SubDescription:  "5% applied to care fees",
				UnitAmountMinor: &unit4,
				LineAmountMinor: 3750,
				IsDiscount:      true,
			},
		},
		SubtotalMinor:   108000,
		DeductionMinor:  66750,
		TotalMinor:      41250,
		BalanceDueMinor: 41250,
	}

	pdfBytes, err := r.Render(context.Background(), input)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatal("expected valid PDF header")
	}
}
