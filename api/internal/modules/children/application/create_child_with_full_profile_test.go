package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/application"
	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/tenant"
)

type fakeChildRepository struct {
	domain.Repository
	createCallCount int
}

func (f *fakeChildRepository) Create(ctx context.Context, tx domain.Tx, child domain.Child, notes string, tenantID, branchID uuid.UUID) error {
	f.createCallCount++
	return nil
}
func (f *fakeChildRepository) UpsertCollectionSetting(ctx context.Context, tx domain.Tx, p *domain.ChildCollectionSetting) (*domain.ChildCollectionSetting, error) {
	return p, nil
}
func (f *fakeChildRepository) SetCollectionPassword(ctx context.Context, tx domain.Tx, tenantID, branchID, childID, id uuid.UUID, password string, passwordHint string, updatedAt time.Time, userID, membershipID uuid.UUID) error {
	return nil
}
func (f *fakeChildRepository) InsertRoomAssignment(ctx context.Context, tx domain.Tx, a *domain.ChildRoomAssignment) (*domain.ChildRoomAssignment, error) {
	return a, nil
}
func (f *fakeChildRepository) InsertConsent(ctx context.Context, tx domain.Tx, p *domain.ChildConsent) (*domain.ChildConsent, error) {
	return p, nil
}
func (f *fakeChildRepository) UpsertBillingProfile(ctx context.Context, tx domain.Tx, p *domain.ChildBillingProfile) (*domain.ChildBillingProfile, error) {
	return p, nil
}
func (f *fakeChildRepository) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	return domain.Child{}, false, nil
}
func (f *fakeChildRepository) ExistsInScope(ctx context.Context, tx domain.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	return true, nil
}

type fakeCreateTxm struct {
	application.TxManager
}

func (f *fakeCreateTxm) ExecTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return fn(nil)
}

func createActorContext(tenantID, branchID uuid.UUID) tenant.ActorContext {
	return tenant.ActorContext{
		UserID:       uuid.New(),
		MembershipID: uuid.New(),
		TenantID:     tenantID,
		BranchID:     branchID,
	}
}

func newDefaultInput() application.CreateChildFullInput {
	return application.CreateChildFullInput{
		Child: application.CreateChildIdentityInput{
			FirstName:   "Alice",
			DateOfBirth: "2022-06-01",
			StartDate:   "2026-09-01",
		},
		Consent: &application.ChildConsentInput{
			SafeguardingReportingAcknowledgement: true,
			SignerName:                           "Jane Manager",
			SignedDate:                           "2026-09-01",
		},
		Room: &application.ChildRoomAssignmentInput{
			RoomID:    uuid.New().String(),
			StartDate: "2026-09-01",
		},
	}
}

func TestCreateChildWithFullProfile(t *testing.T) {
	t.Run("SuccessfulCreate", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()

		repo := &fakeChildRepository{}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, nil, nil, nil, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		input := newDefaultInput()

		result, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ChildID == uuid.Nil {
			t.Error("expected ChildID to be set")
		}
		if result.FirstName != "Alice" {
			t.Errorf("expected FirstName 'Alice', got %q", result.FirstName)
		}
	})

	t.Run("MissingFirstName", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()

		repo := &fakeChildRepository{}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, nil, nil, nil, nil)

		input := newDefaultInput()
		input.Child.FirstName = ""

		_, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("MissingConsent", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()

		repo := &fakeChildRepository{}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, nil, nil, nil, nil)

		input := newDefaultInput()
		input.Consent = nil

		_, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("MissingRoom", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()

		repo := &fakeChildRepository{}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, nil, nil, nil, nil)

		input := newDefaultInput()
		input.Room = nil

		_, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
