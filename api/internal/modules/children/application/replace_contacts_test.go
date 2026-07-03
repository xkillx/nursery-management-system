package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"nursery-management-system/api/internal/modules/children/application"
	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type fakeReplaceContactsRepo struct {
	domain.Repository
}

func (f *fakeReplaceContactsRepo) ExistsInScope(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) (bool, error) {
	return true, nil
}

func (f *fakeReplaceContactsRepo) ReplaceContactsForTypes(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID, _ []domain.ContactType, _ []domain.ChildContact) error {
	return nil
}

func (f *fakeReplaceContactsRepo) ListContactsByChild(_ context.Context, _, _, _ uuid.UUID) ([]domain.ChildContact, error) {
	return []domain.ChildContact{}, nil
}

type fakeTx struct{}

func (f *fakeTx) Begin(_ context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakeTx) Commit(_ context.Context) error          { return nil }
func (f *fakeTx) Rollback(_ context.Context) error        { return nil }
func (f *fakeTx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakeTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects                             { return pgx.LargeObjects{} }
func (f *fakeTx) Prepare(_ context.Context, _ string, _ string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakeTx) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakeTx) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) { return nil, nil }
func (f *fakeTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row        { return nil }
func (f *fakeTx) Conn() *pgx.Conn                                               { return nil }

type fakeReplaceContactsTxm struct{}

func (f *fakeReplaceContactsTxm) ExecTx(_ context.Context, fn func(pgx.Tx) error) error {
	return fn(&fakeTx{})
}

func makeReplaceContactsInput(contactType domain.ContactType, address map[string]any) []application.ChildContactInput {
	email := "test@example.com"
	return []application.ChildContactInput{
		{
			ContactType: contactType,
			FullName:    "Test Parent",
			Email:       &email,
			Address:     address,
		},
	}
}

func TestReplaceContacts_ParentCarerAddressValidation(t *testing.T) {
	actor := tenant.ActorContext{
		TenantID: uuid.New(),
		BranchID: uuid.New(),
		UserID:   uuid.New(),
	}

	t.Run("missing parent carer address returns validation error", func(t *testing.T) {
		repo := &fakeReplaceContactsRepo{}
		txm := &fakeReplaceContactsTxm{}
		uc := application.NewReplaceContacts(repo, &audit.Writer{}, txm)

		_, err := uc.Execute(context.Background(), actor, uuid.New().String(), makeReplaceContactsInput(domain.ContactTypeParentCarer, nil))

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var de *domainerrors.DomainError
		if !assertErrorAs(t, err, &de) {
			return
		}
		if de.Code != "validation_error" {
			t.Errorf("got code %q, want validation_error", de.Code)
		}
	})

	t.Run("parent carer with empty address map returns validation error", func(t *testing.T) {
		repo := &fakeReplaceContactsRepo{}
		txm := &fakeReplaceContactsTxm{}
		uc := application.NewReplaceContacts(repo, &audit.Writer{}, txm)

		_, err := uc.Execute(context.Background(), actor, uuid.New().String(), makeReplaceContactsInput(domain.ContactTypeParentCarer, map[string]any{}))

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("parent carer with partial address (no postcode) returns validation error", func(t *testing.T) {
		repo := &fakeReplaceContactsRepo{}
		txm := &fakeReplaceContactsTxm{}
		uc := application.NewReplaceContacts(repo, &audit.Writer{}, txm)

		_, err := uc.Execute(context.Background(), actor, uuid.New().String(), makeReplaceContactsInput(domain.ContactTypeParentCarer, map[string]any{
			"street": "123 High Street",
			"city":   "London",
		}))

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("parent carer with complete address succeeds", func(t *testing.T) {
		repo := &fakeReplaceContactsRepo{}
		txm := &fakeReplaceContactsTxm{}
		uc := application.NewReplaceContacts(repo, &audit.Writer{}, txm)

		_, err := uc.Execute(context.Background(), actor, uuid.New().String(), makeReplaceContactsInput(domain.ContactTypeParentCarer, map[string]any{
			"street":   "123 High Street",
			"city":     "London",
			"postcode": "SW1A 1AA",
		}))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("second parent carer with missing address returns validation error", func(t *testing.T) {
		repo := &fakeReplaceContactsRepo{}
		txm := &fakeReplaceContactsTxm{}
		uc := application.NewReplaceContacts(repo, &audit.Writer{}, txm)

		email := "test@example.com"
		inputs := []application.ChildContactInput{
			{
				ContactType: domain.ContactTypeParentCarer,
				FullName:    "Parent One",
				Email:       &email,
				Address:     map[string]any{"street": "123 High Street", "city": "London", "postcode": "SW1A 1AA"},
			},
			{
				ContactType: domain.ContactTypeParentCarer,
				FullName:    "Parent Two",
				Email:       &email,
				Address:     map[string]any{},
			},
		}

		_, err := uc.Execute(context.Background(), actor, uuid.New().String(), inputs)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("emergency contact missing address succeeds (no validation)", func(t *testing.T) {
		repo := &fakeReplaceContactsRepo{}
		txm := &fakeReplaceContactsTxm{}
		uc := application.NewReplaceContacts(repo, &audit.Writer{}, txm)

		_, err := uc.Execute(context.Background(), actor, uuid.New().String(), makeReplaceContactsInput(domain.ContactTypeEmergencyContact, nil))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("legacy text format address fails validation", func(t *testing.T) {
		repo := &fakeReplaceContactsRepo{}
		txm := &fakeReplaceContactsTxm{}
		uc := application.NewReplaceContacts(repo, &audit.Writer{}, txm)

		_, err := uc.Execute(context.Background(), actor, uuid.New().String(), makeReplaceContactsInput(domain.ContactTypeParentCarer, map[string]any{
			"text": "123 High Street, London, SW1A 1AA",
		}))

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func assertErrorAs[T error](t *testing.T, err error, target *T) bool {
	t.Helper()
	if !errorAs(err, target) {
		t.Fatalf("expected *%T, got %T", *target, err)
		return false
	}
	return true
}

func errorAs[T error](err error, target *T) bool {
	for {
		if e, ok := err.(T); ok {
			*target = e
			return true
		}
		if e, ok := err.(interface{ Unwrap() error }); ok {
			err = e.Unwrap()
		} else {
			return false
		}
	}
}
