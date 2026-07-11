package pdf

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
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
