package application

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/billing/domain"
	siteprofiledomain "nursery-management-system/api/internal/modules/siteprofile/domain"
)

func TestInvoicePDFGenerator_Generate(t *testing.T) {
	gen := NewInvoicePDFGenerator()

	data := InvoicePDFData{
		Invoice: domain.Invoice{
			ID:            uuid.New(),
			TenantID:      uuid.New(),
			BranchID:      uuid.New(),
			InvoiceNumber: "INV-202607-0001",
			Status:        domain.InvoiceStatusIssued,
			Lines: []domain.InvoiceLine{
				{
					LineKind:        domain.LineKindCoreChildcare,
					Description:     "Core childcare",
					SortOrder:       1,
					QuantityMinutes: 120,
					UnitAmount:      domain.MustGBP(500),
					LineAmount:      domain.MustGBP(1000),
				},
				{
					LineKind:        domain.LineKindCoreChildcare,
					Description:     "Extra hours",
					SortOrder:       2,
					QuantityMinutes: 60,
					UnitAmount:      domain.MustGBP(500),
					LineAmount:      domain.MustGBP(500),
				},
			},
			TotalMinor: 1500,
			IssuedAt:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			DueDate:    time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC),
		},
		SiteProfile: siteprofiledomain.SiteProfile{
			NurseryName:     "Little Explorers Nursery",
			AddressStreet:   "123 High Street",
			AddressCity:     "London",
			AddressPostcode: "SW1A 1AA",
			Phone:           "020 1234 5678",
			Email:           "info@littleexplorers.co.uk",
		},
		ParentName:  "Jane Smith",
		ChildName:   "Tom Smith",
		CheckoutURL: "https://checkout.stripe.com/test-session",
	}

	pdfBytes, err := gen.Generate(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	if pdfBytes[0] != '%' || pdfBytes[1] != 'P' || pdfBytes[2] != 'D' || pdfBytes[3] != 'F' {
		t.Fatal("expected PDF header")
	}
}

func TestInvoicePDFGenerator_GenerateEmptyLines(t *testing.T) {
	gen := NewInvoicePDFGenerator()

	data := InvoicePDFData{
		Invoice: domain.Invoice{
			ID:            uuid.New(),
			InvoiceNumber: "INV-202607-0002",
			Status:        domain.InvoiceStatusIssued,
			Lines:         []domain.InvoiceLine{},
			TotalMinor:    0,
			IssuedAt:      time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			DueDate:       time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC),
		},
		SiteProfile: siteprofiledomain.SiteProfile{
			NurseryName: "Test Nursery",
		},
		ParentName: "Parent",
		ChildName:  "Child",
	}

	pdfBytes, err := gen.Generate(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func TestInvoicePDFGenerator_GenerateNoCheckoutURL(t *testing.T) {
	gen := NewInvoicePDFGenerator()

	data := InvoicePDFData{
		Invoice: domain.Invoice{
			ID:            uuid.New(),
			InvoiceNumber: "INV-202607-0003",
			Status:        domain.InvoiceStatusIssued,
			Lines: []domain.InvoiceLine{
				{
					LineKind:        domain.LineKindCoreChildcare,
					Description:     "Core childcare",
					QuantityMinutes: 60,
					UnitAmount:      domain.MustGBP(500),
					LineAmount:      domain.MustGBP(500),
				},
			},
			TotalMinor: 500,
			IssuedAt:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			DueDate:    time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC),
		},
		SiteProfile: siteprofiledomain.SiteProfile{
			NurseryName: "Test Nursery",
		},
	}

	pdfBytes, err := gen.Generate(context.Background(), data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func TestFormatPDFMoney(t *testing.T) {
	tests := []struct {
		minor int
		want  string
	}{
		{0, "\u00a30.00"},
		{100, "\u00a31.00"},
		{1500, "\u00a315.00"},
		{1050, "\u00a310.50"},
		{-500, "-\u00a35.00"},
	}

	for _, tt := range tests {
		got := formatPDFMoney(tt.minor)
		if got != tt.want {
			t.Errorf("formatPDFMoney(%d) = %q, want %q", tt.minor, got, tt.want)
		}
	}
}
