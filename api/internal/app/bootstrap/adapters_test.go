package bootstrap

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"nursery-management-system/api/internal/modules/billing/domain"
	billingpdf "nursery-management-system/api/internal/modules/billing/infrastructure/pdf"
	emaildomain "nursery-management-system/api/internal/modules/email/domain"
	siteprofiledomain "nursery-management-system/api/internal/modules/siteprofile/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/storage"
)

func strPtr(s string) *string { return &s }

// stubAttachmentRepo satisfies billingdomain.BillingRepository for the
// notification adapter tests. All methods except the two exercised are
// inherited from the nil embedded interface.
type stubAttachmentRepo struct {
	domain.BillingRepository
	invoice  domain.InvoiceReviewRow
	found    bool
	getErr   error
	lines    []domain.InvoiceReviewLineRow
	linesErr error
}

func (s *stubAttachmentRepo) GetInvoiceForManagerReviewTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (domain.InvoiceReviewRow, bool, error) {
	return s.invoice, s.found, s.getErr
}

func (s *stubAttachmentRepo) ListInvoiceLinesForManagerReviewTx(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) ([]domain.InvoiceReviewLineRow, error) {
	return s.lines, s.linesErr
}

type stubAttachmentParentContacts struct {
	pc  *domain.ParentContact
	err error
}

func (s *stubAttachmentParentContacts) GetForInvoice(_ context.Context, _, _, _ uuid.UUID) (*domain.ParentContact, error) {
	return s.pc, s.err
}

type stubAttachmentSiteProfiles struct {
	sp  *siteprofiledomain.SiteProfile
	err error
}

func (s *stubAttachmentSiteProfiles) GetForInvoice(_ context.Context, _, _ uuid.UUID) (*siteprofiledomain.SiteProfile, error) {
	return s.sp, s.err
}

// fakeTx is a minimal pgx.Tx used to satisfy the audit write during adapter
// tests. Only Exec is exercised (the audit insert); everything else is
// inherited from the nil embedded interface.
type fakeTx struct {
	pgx.Tx
	err error
}

func (f *fakeTx) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, f.err
}

func testInvoice() domain.InvoiceReviewRow {
	return domain.InvoiceReviewRow{
		ID:              uuid.New(),
		ChildFirstName:  "Leo",
		ChildLastName:   strPtr("Harrison"),
		InvoiceNumber:   strPtr("INV-2026-08-0001"),
		Status:          "issued",
		BillingMonth:    time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Subtotal:        domain.MustGBP(42000),
		FundedDeduction: domain.MustGBP(0),
		TotalDue:        domain.MustGBP(42000),
	}
}

func testSiteProfile() *siteprofiledomain.SiteProfile {
	return &siteprofiledomain.SiteProfile{
		NurseryName: "Sunny Days Nursery",
		Email:       "billing@sunny.example.com",
	}
}

func testParentContact() *domain.ParentContact {
	return &domain.ParentContact{
		FullName: "Jane Doe",
		Email:    "jane@example.com",
	}
}

func newTestRenderer(t *testing.T) *billingpdf.Renderer {
	t.Helper()
	r, err := billingpdf.NewRenderer()
	if err != nil {
		t.Fatalf("create pdf renderer: %v", err)
	}
	return r
}

func TestBuildInvoiceAttachment_HappyPath(t *testing.T) {
	renderer := newTestRenderer(t)
	fakeStorage := storage.NewFakeService()
	invoice := testInvoice()

	adapter := &billingNotificationAdapter{
		repo:        &stubAttachmentRepo{invoice: invoice, found: true},
		pdfRenderer: renderer,
		storage:     fakeStorage,
	}

	ref := adapter.buildInvoiceAttachment(context.Background(), &fakeTx{}, uuid.New(), uuid.New(), invoice.ID, invoice, testSiteProfile(), testParentContact())
	if ref == nil {
		t.Fatal("expected attachment ref")
	}
	if ref.Filename != "invoice.pdf" {
		t.Errorf("Filename = %q, want invoice.pdf", ref.Filename)
	}
	if ref.ContentType != "application/pdf" {
		t.Errorf("ContentType = %q, want application/pdf", ref.ContentType)
	}
	if !strings.HasPrefix(ref.S3Key, "invoices/") {
		t.Errorf("S3Key = %q, want prefix invoices/", ref.S3Key)
	}
	data, ok := fakeStorage.Objects[ref.S3Key]
	if !ok {
		t.Fatal("expected PDF uploaded to S3")
	}
	if len(data) == 0 {
		t.Error("expected non-empty PDF bytes")
	}
	if !strings.HasPrefix(string(data), "%PDF") {
		t.Errorf("expected PDF header, got %q", string(data[:min(len(data), 8)]))
	}
}

func TestBuildInvoiceAttachment_UploadFailureReturnsNil(t *testing.T) {
	renderer := newTestRenderer(t)
	fakeStorage := storage.NewFakeService()
	fakeStorage.UploadErr = errors.New("s3 down")
	invoice := testInvoice()

	adapter := &billingNotificationAdapter{
		repo:        &stubAttachmentRepo{invoice: invoice, found: true},
		pdfRenderer: renderer,
		storage:     fakeStorage,
	}

	ref := adapter.buildInvoiceAttachment(context.Background(), &fakeTx{}, uuid.New(), uuid.New(), invoice.ID, invoice, testSiteProfile(), testParentContact())
	if ref != nil {
		t.Errorf("expected nil ref on upload failure, got %+v", ref)
	}
	if len(fakeStorage.Objects) != 0 {
		t.Errorf("expected no S3 objects, got %d", len(fakeStorage.Objects))
	}
}

func TestBuildInvoiceAttachment_LinesFetchFailureReturnsNil(t *testing.T) {
	renderer := newTestRenderer(t)
	invoice := testInvoice()

	adapter := &billingNotificationAdapter{
		repo:        &stubAttachmentRepo{invoice: invoice, found: true, linesErr: errors.New("db down")},
		pdfRenderer: renderer,
		storage:     storage.NewFakeService(),
	}

	ref := adapter.buildInvoiceAttachment(context.Background(), &fakeTx{}, uuid.New(), uuid.New(), invoice.ID, invoice, testSiteProfile(), testParentContact())
	if ref != nil {
		t.Errorf("expected nil ref on lines fetch failure, got %+v", ref)
	}
}

func TestBuildInvoiceAttachment_NoRendererOrStorageReturnsNil(t *testing.T) {
	invoice := testInvoice()

	noStorage := &billingNotificationAdapter{
		repo:        &stubAttachmentRepo{invoice: invoice, found: true},
		pdfRenderer: newTestRenderer(t),
	}
	if ref := noStorage.buildInvoiceAttachment(context.Background(), &fakeTx{}, uuid.New(), uuid.New(), invoice.ID, invoice, testSiteProfile(), testParentContact()); ref != nil {
		t.Errorf("expected nil ref without storage, got %+v", ref)
	}

	noRenderer := &billingNotificationAdapter{
		repo:    &stubAttachmentRepo{invoice: invoice, found: true},
		storage: storage.NewFakeService(),
	}
	if ref := noRenderer.buildInvoiceAttachment(context.Background(), &fakeTx{}, uuid.New(), uuid.New(), invoice.ID, invoice, testSiteProfile(), testParentContact()); ref != nil {
		t.Errorf("expected nil ref without renderer, got %+v", ref)
	}
}

func TestSendInvoiceIssuedEmail_EnqueuesWithAttachmentRefAndVersion2(t *testing.T) {
	renderer := newTestRenderer(t)
	fakeStorage := storage.NewFakeService()
	enqueuer := emaildomain.NewFakeEnqueuer()
	invoice := testInvoice()

	adapter := &billingNotificationAdapter{
		repo:           &stubAttachmentRepo{invoice: invoice, found: true},
		parentContacts: &stubAttachmentParentContacts{pc: testParentContact()},
		siteProfiles:   &stubAttachmentSiteProfiles{sp: testSiteProfile()},
		enqueuer:       enqueuer,
		auditWriter:    audit.NewWriter(),
		webBaseURL:     "https://app.example.com",
		pdfRenderer:    renderer,
		storage:        fakeStorage,
	}

	if err := adapter.SendInvoiceIssuedEmail(context.Background(), &fakeTx{}, invoice.ID, uuid.New(), uuid.New()); err != nil {
		t.Fatalf("send invoice issued: %v", err)
	}

	if len(enqueuer.Enqueued) != 1 {
		t.Fatalf("expected 1 enqueued message, got %d", len(enqueuer.Enqueued))
	}
	got := enqueuer.Enqueued[0]
	if got.TemplateName != "issued" {
		t.Errorf("TemplateName = %q, want issued", got.TemplateName)
	}
	if got.TemplateVersion != 2 {
		t.Errorf("TemplateVersion = %d, want 2", got.TemplateVersion)
	}
	if got.Recipient != "jane@example.com" {
		t.Errorf("Recipient = %q, want jane@example.com", got.Recipient)
	}
	if len(got.AttachmentRefs) != 1 {
		t.Fatalf("expected 1 attachment ref, got %d", len(got.AttachmentRefs))
	}
	if !strings.HasPrefix(got.AttachmentRefs[0].S3Key, "invoices/") {
		t.Errorf("S3Key = %q, want prefix invoices/", got.AttachmentRefs[0].S3Key)
	}
	if !strings.Contains(string(got.PayloadJSON), "INV-2026-08-0001") {
		t.Errorf("payload missing invoice number: %s", got.PayloadJSON)
	}
}

func TestSendInvoiceIssuedEmail_BestEffortEnqueuesWithoutAttachment(t *testing.T) {
	renderer := newTestRenderer(t)
	fakeStorage := storage.NewFakeService()
	fakeStorage.UploadErr = errors.New("s3 down")
	enqueuer := emaildomain.NewFakeEnqueuer()
	invoice := testInvoice()

	adapter := &billingNotificationAdapter{
		repo:           &stubAttachmentRepo{invoice: invoice, found: true},
		parentContacts: &stubAttachmentParentContacts{pc: testParentContact()},
		siteProfiles:   &stubAttachmentSiteProfiles{sp: testSiteProfile()},
		enqueuer:       enqueuer,
		auditWriter:    audit.NewWriter(),
		webBaseURL:     "https://app.example.com",
		pdfRenderer:    renderer,
		storage:        fakeStorage,
	}

	if err := adapter.SendInvoiceIssuedEmail(context.Background(), &fakeTx{}, invoice.ID, uuid.New(), uuid.New()); err != nil {
		t.Fatalf("send invoice issued: %v", err)
	}

	if len(enqueuer.Enqueued) != 1 {
		t.Fatalf("expected 1 enqueued message, got %d", len(enqueuer.Enqueued))
	}
	got := enqueuer.Enqueued[0]
	if got.TemplateVersion != 2 {
		t.Errorf("TemplateVersion = %d, want 2", got.TemplateVersion)
	}
	if len(got.AttachmentRefs) != 0 {
		t.Errorf("expected no attachment on best-effort failure, got %d", len(got.AttachmentRefs))
	}
}

func TestSendInvoiceIssuedEmail_SkipsWhenParentHasNoEmail(t *testing.T) {
	enqueuer := emaildomain.NewFakeEnqueuer()
	invoice := testInvoice()

	adapter := &billingNotificationAdapter{
		repo:           &stubAttachmentRepo{invoice: invoice, found: true},
		parentContacts: &stubAttachmentParentContacts{pc: &domain.ParentContact{FullName: "No Email"}},
		siteProfiles:   &stubAttachmentSiteProfiles{sp: testSiteProfile()},
		enqueuer:       enqueuer,
		auditWriter:    audit.NewWriter(),
		webBaseURL:     "https://app.example.com",
		pdfRenderer:    newTestRenderer(t),
		storage:        storage.NewFakeService(),
	}

	if err := adapter.SendInvoiceIssuedEmail(context.Background(), &fakeTx{}, invoice.ID, uuid.New(), uuid.New()); err != nil {
		t.Fatalf("send invoice issued: %v", err)
	}

	if len(enqueuer.Enqueued) != 0 {
		t.Errorf("expected no enqueued message when parent has no email, got %d", len(enqueuer.Enqueued))
	}
}
