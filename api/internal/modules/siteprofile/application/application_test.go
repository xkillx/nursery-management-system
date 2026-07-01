package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/siteprofile/application"
	"nursery-management-system/api/internal/modules/siteprofile/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type fakeRepo struct {
	profile   *domain.SiteProfile
	getErr    error
	upsertErr error
}

func (f *fakeRepo) GetByBranch(ctx context.Context, tenantID, branchID uuid.UUID) (domain.SiteProfile, error) {
	if f.getErr != nil {
		return domain.SiteProfile{}, f.getErr
	}
	if f.profile == nil {
		return domain.SiteProfile{}, domainerrors.NotFound("site_profile", "Site profile not found.")
	}
	return *f.profile, nil
}

func (f *fakeRepo) Upsert(ctx context.Context, tx domain.Tx, tenantID, branchID uuid.UUID, profile domain.SiteProfile) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.profile = &profile
	return nil
}

type fakeTxManager struct{}

func (f *fakeTxManager) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

type fakeAuditWriter struct {
	written bool
}

func (f *fakeAuditWriter) WriteWithTx(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, params audit.WriteParams) error {
	f.written = true
	return nil
}

func makeActor() tenant.ActorContext {
	return tenant.ActorContext{
		UserID:       uuid.New(),
		MembershipID: uuid.New(),
		TenantID:     uuid.New(),
		BranchID:     uuid.New(),
	}
}

func TestGetSiteProfile_NotFound(t *testing.T) {
	repo := &fakeRepo{}
	uc := application.NewGetSiteProfileUseCase(repo)

	_, err := uc.Execute(context.Background(), makeActor())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSiteProfile_Found(t *testing.T) {
	actor := makeActor()
	repo := &fakeRepo{
		profile: &domain.SiteProfile{
			NurseryName: "Little Stars Nursery",
		},
	}
	uc := application.NewGetSiteProfileUseCase(repo)

	profile, err := uc.Execute(context.Background(), actor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.NurseryName != "Little Stars Nursery" {
		t.Errorf("got nursery_name %q, want %q", profile.NurseryName, "Little Stars Nursery")
	}
}

func TestUpdateSiteProfile_HappyPath(t *testing.T) {
	actor := makeActor()
	repo := &fakeRepo{}
	audit := &fakeAuditWriter{}
	txMgr := &fakeTxManager{}
	uc := application.NewUpdateSiteProfileUseCase(repo, audit, txMgr)

	in := application.UpdateSiteProfileInput{
		NurseryName:     "Little Stars Nursery",
		Description:     "A warm and caring nursery",
		Phone:           "+44 161 555 0100",
		Email:           "hello@littlestars.example",
		Website:         "https://littlestars.example",
		AddressStreet:   "12 Acacia Ave",
		AddressCity:     "Manchester",
		AddressPostcode: "M1 4BT",
	}

	saved, err := uc.Execute(context.Background(), actor, in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved.NurseryName != "Little Stars Nursery" {
		t.Errorf("got nursery_name %q, want %q", saved.NurseryName, "Little Stars Nursery")
	}
	if !audit.written {
		t.Error("expected audit log to be written")
	}
}

func TestUpdateSiteProfile_AllEmptyFields(t *testing.T) {
	actor := makeActor()
	repo := &fakeRepo{}
	audit := &fakeAuditWriter{}
	txMgr := &fakeTxManager{}
	uc := application.NewUpdateSiteProfileUseCase(repo, audit, txMgr)

	in := application.UpdateSiteProfileInput{}

	_, err := uc.Execute(context.Background(), actor, in)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var domainErr *domainerrors.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if domainErr.Code != "validation_error" {
		t.Errorf("got code %q, want %q", domainErr.Code, "validation_error")
	}
	fields, ok := domainErr.Details["field_errors"].([]domainerrors.FieldError)
	if !ok {
		t.Fatal("expected field_errors in details")
	}
	if len(fields) != 7 {
		t.Errorf("expected 7 field errors (all required except description), got %d", len(fields))
	}
}

func TestUpdateSiteProfile_WhitespaceName(t *testing.T) {
	actor := makeActor()
	repo := &fakeRepo{}
	audit := &fakeAuditWriter{}
	txMgr := &fakeTxManager{}
	uc := application.NewUpdateSiteProfileUseCase(repo, audit, txMgr)

	in := application.UpdateSiteProfileInput{
		NurseryName:     "   ",
		Description:     "",
		Phone:           "+44 161 555 0100",
		Email:           "hello@littlestars.example",
		Website:         "https://littlestars.example",
		AddressStreet:   "12 Acacia Ave",
		AddressCity:     "Manchester",
		AddressPostcode: "M1 4BT",
	}

	_, err := uc.Execute(context.Background(), actor, in)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var domainErr *domainerrors.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	fields, ok := domainErr.Details["field_errors"].([]domainerrors.FieldError)
	if !ok {
		t.Fatal("expected field_errors in details")
	}
	found := false
	for _, f := range fields {
		if f.Field == "nursery_name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected nursery_name field error after trim")
	}
}

func TestUpdateSiteProfile_NameMaxLength(t *testing.T) {
	actor := makeActor()
	repo := &fakeRepo{}
	audit := &fakeAuditWriter{}
	txMgr := &fakeTxManager{}
	uc := application.NewUpdateSiteProfileUseCase(repo, audit, txMgr)

	in := application.UpdateSiteProfileInput{
		NurseryName:     string(make([]byte, 120)),
		Description:     "",
		Phone:           "+44 161 555 0100",
		Email:           "hello@littlestars.example",
		Website:         "https://littlestars.example",
		AddressStreet:   "12 Acacia Ave",
		AddressCity:     "Manchester",
		AddressPostcode: "M1 4BT",
	}

	for i := range in.NurseryName {
		in.NurseryName = in.NurseryName[:i] + "a" + in.NurseryName[i+1:]
	}

	_, err := uc.Execute(context.Background(), actor, in)
	if err != nil {
		t.Fatalf("expected success with 120-char name, got %v", err)
	}
}

func TestUpdateSiteProfile_NameTooLong(t *testing.T) {
	actor := makeActor()
	repo := &fakeRepo{}
	audit := &fakeAuditWriter{}
	txMgr := &fakeTxManager{}
	uc := application.NewUpdateSiteProfileUseCase(repo, audit, txMgr)

	in := application.UpdateSiteProfileInput{
		NurseryName:     string(make([]byte, 121)),
		Description:     "",
		Phone:           "+44 161 555 0100",
		Email:           "hello@littlestars.example",
		Website:         "https://littlestars.example",
		AddressStreet:   "12 Acacia Ave",
		AddressCity:     "Manchester",
		AddressPostcode: "M1 4BT",
	}

	for i := range in.NurseryName {
		in.NurseryName = in.NurseryName[:i] + "a" + in.NurseryName[i+1:]
	}

	_, err := uc.Execute(context.Background(), actor, in)
	if err == nil {
		t.Fatal("expected error for 121-char name, got nil")
	}
	var domainErr *domainerrors.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	if domainErr.Code != "validation_error" {
		t.Errorf("got code %q, want %q", domainErr.Code, "validation_error")
	}
}

func TestUpdateSiteProfile_InvalidEmail(t *testing.T) {
	actor := makeActor()
	repo := &fakeRepo{}
	audit := &fakeAuditWriter{}
	txMgr := &fakeTxManager{}
	uc := application.NewUpdateSiteProfileUseCase(repo, audit, txMgr)

	in := application.UpdateSiteProfileInput{
		NurseryName:     "Little Stars Nursery",
		Description:     "",
		Phone:           "+44 161 555 0100",
		Email:           "not-an-email",
		Website:         "https://littlestars.example",
		AddressStreet:   "12 Acacia Ave",
		AddressCity:     "Manchester",
		AddressPostcode: "M1 4BT",
	}

	_, err := uc.Execute(context.Background(), actor, in)
	if err == nil {
		t.Fatal("expected error for invalid email, got nil")
	}
	var domainErr *domainerrors.DomainError
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected DomainError, got %T", err)
	}
	fields, ok := domainErr.Details["field_errors"].([]domainerrors.FieldError)
	if !ok {
		t.Fatal("expected field_errors in details")
	}
	found := false
	for _, f := range fields {
		if f.Field == "email" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected email field error")
	}
}

func TestUpdateSiteProfile_ArchivedBranch_gate(t *testing.T) {
	t.Skip("archived branch gate not yet implemented")
}

func TestUpdateSiteProfile_TxError(t *testing.T) {
	repo := &fakeRepo{upsertErr: errors.New("db error")}
	audit := &fakeAuditWriter{}
	txMgr := &fakeTxManager{}
	uc := application.NewUpdateSiteProfileUseCase(repo, audit, txMgr)

	actor := makeActor()
	in := application.UpdateSiteProfileInput{
		NurseryName:     "Little Stars Nursery",
		Description:     "",
		Phone:           "+44 161 555 0100",
		Email:           "hello@littlestars.example",
		Website:         "https://littlestars.example",
		AddressStreet:   "12 Acacia Ave",
		AddressCity:     "Manchester",
		AddressPostcode: "M1 4BT",
	}

	_, err := uc.Execute(context.Background(), actor, in)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
