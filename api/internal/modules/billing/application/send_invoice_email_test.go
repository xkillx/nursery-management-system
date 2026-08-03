package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/billing/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

func sendTestInvoice(status string) domain.InvoiceReviewRow {
	return domain.InvoiceReviewRow{
		ID:              uuid.New(),
		ChildID:         uuid.New(),
		Status:          status,
		InvoiceNumber:   sendTestStrPtr("INV-2026-0001"),
		ChildFirstName:  "Leo",
		ChildLastName:   sendTestStrPtr("Harrison"),
		BillingMonth:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		Subtotal:        domain.MustGBP(42000),
		FundedDeduction: domain.MustGBP(0),
		TotalDue:        domain.MustGBP(42000),
	}
}

func sendTestStrPtr(s string) *string { return &s }

func sendTestActor() tenant.ActorContext {
	return tenant.ActorContext{
		UserID:       uuid.New(),
		MembershipID: uuid.New(),
		TenantID:     uuid.New(),
		BranchID:     uuid.New(),
	}
}

type sendEmailRepoStub struct {
	domain.BillingRepository
	invoice          domain.InvoiceReviewRow
	found            bool
	getErr           error
	recentResends    int
	recentResendsErr error
	lockCalls        int
	lockResult       bool
	lockErr          error
}

func (s *sendEmailRepoStub) GetInvoiceForManagerReviewTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	return s.invoice, s.found, s.getErr
}

func (s *sendEmailRepoStub) CountRecentInvoiceResendsTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ time.Time) (int, error) {
	return s.recentResends, s.recentResendsErr
}

func (s *sendEmailRepoStub) LockInvoiceForResendTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (bool, error) {
	s.lockCalls++
	return s.lockResult, s.lockErr
}

type sendEmailParentStub struct {
	pc  *domain.ParentContact
	err error
}

func (s *sendEmailParentStub) GetForInvoice(_ context.Context, _, _, _ uuid.UUID) (*domain.ParentContact, error) {
	return s.pc, s.err
}

type sendEmailSenderStub struct {
	calls int
	err   error
}

func (s *sendEmailSenderStub) SendInvoiceResendEmail(_ context.Context, _ pgx.Tx, _, _, _ uuid.UUID) error {
	s.calls++
	return s.err
}

// sendEmailTxMgr runs the closure inline, substituting for the real
// transaction.Manager in unit tests.
type sendEmailTxMgr struct{}

func (m *sendEmailTxMgr) ExecTx(_ context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

func newSendEmailUC(repo *sendEmailRepoStub, parent *sendEmailParentStub, sender *sendEmailSenderStub) *SendInvoiceEmail {
	return &SendInvoiceEmail{repo: repo, parentLookup: parent, sender: sender, txMgr: &sendEmailTxMgr{}}
}

func TestSendInvoiceEmail_HappyPathQueues(t *testing.T) {
	repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: true, lockResult: true}
	parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
	sender := &sendEmailSenderStub{}
	uc := newSendEmailUC(repo, parent, sender)

	result, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "queued" {
		t.Fatalf("status = %q, want queued", result.Status)
	}
	if sender.calls != 1 {
		t.Fatalf("sender calls = %d, want 1", sender.calls)
	}
}

func TestSendInvoiceEmail_RejectsNonResendableStatus(t *testing.T) {
	for _, status := range []string{domain.InvoiceStatusDraft, domain.InvoiceStatusVoid} {
		repo := &sendEmailRepoStub{invoice: sendTestInvoice(status), found: true, lockResult: true}
		parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
		sender := &sendEmailSenderStub{}
		uc := newSendEmailUC(repo, parent, sender)

		_, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
		var de *domainerrors.DomainError
		if !errors.As(err, &de) {
			t.Fatalf("status %s: expected DomainError, got %v", status, err)
		}
		if de.Code != "invoice_not_payable" {
			t.Fatalf("status %s: code = %q, want invoice_not_payable", status, de.Code)
		}
		if sender.calls != 0 {
			t.Fatalf("status %s: sender calls = %d, want 0", status, sender.calls)
		}
	}
}

func TestSendInvoiceEmail_InvoiceNotFound(t *testing.T) {
	repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: false, lockResult: true}
	parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
	sender := &sendEmailSenderStub{}
	uc := newSendEmailUC(repo, parent, sender)

	_, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
	var de *domainerrors.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %v", err)
	}
	if de.Code != "invoice_not_found" {
		t.Fatalf("code = %q, want invoice_not_found", de.Code)
	}
	if sender.calls != 0 {
		t.Fatalf("sender calls = %d, want 0", sender.calls)
	}
}

func TestSendInvoiceEmail_ParentNoEmail(t *testing.T) {
	for _, pc := range []*domain.ParentContact{
		{FullName: "Jane Doe"},
		nil,
	} {
		repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: true, lockResult: true}
		parent := &sendEmailParentStub{pc: pc}
		sender := &sendEmailSenderStub{}
		uc := newSendEmailUC(repo, parent, sender)

		_, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
		var de *domainerrors.DomainError
		if !errors.As(err, &de) {
			t.Fatalf("expected DomainError, got %v", err)
		}
		if de.Code != "parent_no_email" {
			t.Fatalf("code = %q, want parent_no_email", de.Code)
		}
		if sender.calls != 0 {
			t.Fatalf("sender calls = %d, want 0", sender.calls)
		}
	}
}

func TestSendInvoiceEmail_ThrottledWithinCooldown(t *testing.T) {
	repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: true, recentResends: 1, lockResult: true}
	parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
	sender := &sendEmailSenderStub{}
	uc := newSendEmailUC(repo, parent, sender)

	_, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
	var de *domainerrors.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %v", err)
	}
	if de.Code != "invoice_resend_throttled" {
		t.Fatalf("code = %q, want invoice_resend_throttled", de.Code)
	}
	if sender.calls != 0 {
		t.Fatalf("sender calls = %d, want 0", sender.calls)
	}
}

func TestSendInvoiceEmail_AllowsAfterCooldown(t *testing.T) {
	repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: true, recentResends: 0, lockResult: true}
	parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
	sender := &sendEmailSenderStub{}
	uc := newSendEmailUC(repo, parent, sender)

	result, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "queued" {
		t.Fatalf("status = %q, want queued", result.Status)
	}
	if sender.calls != 1 {
		t.Fatalf("sender calls = %d, want 1", sender.calls)
	}
}

func TestSendInvoiceEmail_SenderFailurePropagates(t *testing.T) {
	repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: true, lockResult: true}
	parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
	senderErr := errors.New("enqueue failed")
	sender := &sendEmailSenderStub{err: senderErr}
	uc := newSendEmailUC(repo, parent, sender)

	_, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
	if err == nil {
		t.Fatal("expected error from sender failure")
	}
	if !errors.Is(err, senderErr) {
		t.Fatalf("expected sender error to propagate, got %v", err)
	}
	if sender.calls != 1 {
		t.Fatalf("sender calls = %d, want 1", sender.calls)
	}
}

func TestSendInvoiceEmail_InvalidID(t *testing.T) {
	repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: true, lockResult: true}
	parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
	sender := &sendEmailSenderStub{}
	uc := newSendEmailUC(repo, parent, sender)

	_, err := uc.Execute(context.Background(), sendTestActor(), "not-a-uuid")
	var de *domainerrors.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %v", err)
	}
	if de.Code != "validation_error" {
		t.Fatalf("code = %q, want validation_error", de.Code)
	}
	if sender.calls != 0 {
		t.Fatalf("sender calls = %d, want 0", sender.calls)
	}
}

func TestSendInvoiceEmail_LockSerializesBeforeCooldownCheck(t *testing.T) {
	repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: true, lockResult: true}
	parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
	sender := &sendEmailSenderStub{}
	uc := newSendEmailUC(repo, parent, sender)

	_, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.lockCalls != 1 {
		t.Fatalf("lock calls = %d, want 1", repo.lockCalls)
	}
}

func TestSendInvoiceEmail_InvoiceNotFoundWhenLockMisses(t *testing.T) {
	repo := &sendEmailRepoStub{invoice: sendTestInvoice(domain.InvoiceStatusIssued), found: true, lockResult: false}
	parent := &sendEmailParentStub{pc: &domain.ParentContact{FullName: "Jane Doe", Email: "jane@example.com"}}
	sender := &sendEmailSenderStub{}
	uc := newSendEmailUC(repo, parent, sender)

	_, err := uc.Execute(context.Background(), sendTestActor(), repo.invoice.ID.String())
	var de *domainerrors.DomainError
	if !errors.As(err, &de) {
		t.Fatalf("expected DomainError, got %v", err)
	}
	if de.Code != "invoice_not_found" {
		t.Fatalf("code = %q, want invoice_not_found", de.Code)
	}
	if sender.calls != 0 {
		t.Fatalf("sender calls = %d, want 0", sender.calls)
	}
}
