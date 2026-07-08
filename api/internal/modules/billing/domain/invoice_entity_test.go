package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newDraftInvoice() Invoice {
	return Invoice{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		BranchID: uuid.New(),
		ChildID:  uuid.New(),
		TermID:   uuid.New(),
		Status:   InvoiceStatusDraft,
	}
}

func TestInvoiceIssue(t *testing.T) {
	t.Run("draft invoice can be issued", func(t *testing.T) {
		inv := newDraftInvoice()
		issuedAt := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		err := inv.Issue("INV-001", issuedAt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv.Status != InvoiceStatusIssued {
			t.Errorf("status = %q, want %q", inv.Status, InvoiceStatusIssued)
		}
		if inv.InvoiceNumber != "INV-001" {
			t.Errorf("InvoiceNumber = %q, want %q", inv.InvoiceNumber, "INV-001")
		}
		if !inv.IssuedAt.Equal(issuedAt) {
			t.Errorf("IssuedAt = %v, want %v", inv.IssuedAt, issuedAt)
		}
	})

	t.Run("non-draft invoice cannot be issued", func(t *testing.T) {
		inv := newDraftInvoice()
		inv.Status = InvoiceStatusIssued
		err := inv.Issue("INV-002", time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty invoice number returns error", func(t *testing.T) {
		inv := newDraftInvoice()
		err := inv.Issue("", time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestInvoiceMarkOverdue(t *testing.T) {
	t.Run("issued invoice can be marked overdue", func(t *testing.T) {
		inv := newDraftInvoice()
		inv.Status = InvoiceStatusIssued
		overdueAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		err := inv.MarkOverdue(overdueAt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv.Status != InvoiceStatusOverdue {
			t.Errorf("status = %q, want %q", inv.Status, InvoiceStatusOverdue)
		}
		if inv.OverdueAt == nil || !inv.OverdueAt.Equal(overdueAt) {
			t.Errorf("OverdueAt = %v, want %v", inv.OverdueAt, overdueAt)
		}
	})

	t.Run("draft invoice cannot be marked overdue", func(t *testing.T) {
		inv := newDraftInvoice()
		err := inv.MarkOverdue(time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestInvoiceVoid(t *testing.T) {
	t.Run("draft invoice can be voided", func(t *testing.T) {
		inv := newDraftInvoice()
		voidedAt := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		err := inv.Void("cancelled by parent", voidedAt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv.Status != InvoiceStatusVoid {
			t.Errorf("status = %q, want %q", inv.Status, InvoiceStatusVoid)
		}
		if inv.VoidReason != "cancelled by parent" {
			t.Errorf("VoidReason = %q, want %q", inv.VoidReason, "cancelled by parent")
		}
		if inv.VoidedAt == nil || !inv.VoidedAt.Equal(voidedAt) {
			t.Errorf("VoidedAt = %v, want %v", inv.VoidedAt, voidedAt)
		}
	})

	t.Run("issued invoice cannot be voided", func(t *testing.T) {
		inv := newDraftInvoice()
		inv.Status = InvoiceStatusIssued
		err := inv.Void("reason", time.Now())
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestInvoiceAddLine(t *testing.T) {
	line := InvoiceLine{
		LineKind:        LineKindCoreChildcare,
		Description:     "Core childcare",
		SortOrder:       1,
		QuantityMinutes: 120,
		UnitAmount:      MustGBP(500),
		LineAmount:      MustGBP(1000),
	}

	t.Run("draft invoice accepts line", func(t *testing.T) {
		inv := newDraftInvoice()
		err := inv.AddLine(line)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(inv.Lines) != 1 {
			t.Errorf("lines count = %d, want 1", len(inv.Lines))
		}
		if inv.TotalMinor != 1000 {
			t.Errorf("TotalMinor = %d, want 1000", inv.TotalMinor)
		}
	})

	t.Run("multiple lines accumulate total", func(t *testing.T) {
		inv := newDraftInvoice()
		inv.AddLine(line)
		inv.AddLine(InvoiceLine{
			LineKind:        LineKindFundedDeduction,
			Description:     "Funded deduction",
			SortOrder:       2,
			QuantityMinutes: 60,
			LineAmount:      MustGBP(-300),
		})
		if inv.TotalMinor != 700 {
			t.Errorf("TotalMinor = %d, want 700", inv.TotalMinor)
		}
	})

	t.Run("issued invoice rejects line", func(t *testing.T) {
		inv := newDraftInvoice()
		inv.Status = InvoiceStatusIssued
		err := inv.AddLine(line)
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("negative quantity minutes returns error", func(t *testing.T) {
		inv := newDraftInvoice()
		badLine := line
		badLine.QuantityMinutes = -1
		err := inv.AddLine(badLine)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestInvoiceIsDraft(t *testing.T) {
	inv := newDraftInvoice()
	if !inv.IsDraft() {
		t.Error("new invoice should be draft")
	}
	inv.Status = InvoiceStatusIssued
	if inv.IsDraft() {
		t.Error("issued invoice should not be draft")
	}
}

func TestInvoiceIsIssued(t *testing.T) {
	inv := newDraftInvoice()
	if inv.IsIssued() {
		t.Error("draft invoice should not be issued")
	}
	inv.Status = InvoiceStatusIssued
	if !inv.IsIssued() {
		t.Error("issued invoice should be issued")
	}
}
