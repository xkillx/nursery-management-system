package application_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/term/application"
	"nursery-management-system/api/internal/modules/term/domain"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/tenant"
)

// mockTermRepo is an in-memory term repo for application-level unit tests.
type mockTermRepo struct {
	terms map[uuid.UUID]*domain.Term
}

func newMockTermRepo() *mockTermRepo {
	return &mockTermRepo{terms: map[uuid.UUID]*domain.Term{}}
}

func (m *mockTermRepo) Insert(_ context.Context, _ domain.Tx, t *domain.Term) (*domain.Term, error) {
	clone := *t
	clone.CreatedAt = time.Now().UTC()
	clone.UpdatedAt = clone.CreatedAt
	m.terms[t.ID] = &clone
	return &clone, nil
}
func (m *mockTermRepo) Terminate(_ context.Context, _ domain.Tx, _, _, id uuid.UUID, terminatedAt time.Time, code, note string) (int64, error) {
	t, ok := m.terms[id]
	if !ok {
		return 0, nil
	}
	if t.Status != domain.TermStatusActive && t.Status != domain.TermStatusPreTerm && t.Status != domain.TermStatusPendingRenewal {
		return 0, nil
	}
	t.Status = domain.TermStatusTerminated
	t.TerminatedAt = &terminatedAt
	c := code
	n := note
	t.TerminationReasonCode = &c
	t.TerminationReasonNote = &n
	t.UpdatedAt = time.Now().UTC()
	return 1, nil
}
func (m *mockTermRepo) UpdateStatus(_ context.Context, _ domain.Tx, _, _, id uuid.UUID, status domain.TermStatus) (int64, error) {
	t, ok := m.terms[id]
	if !ok {
		return 0, nil
	}
	t.Status = status
	t.UpdatedAt = time.Now().UTC()
	return 1, nil
}
func (m *mockTermRepo) GetByID(_ context.Context, _, _, id uuid.UUID) (*domain.Term, bool, error) {
	t, ok := m.terms[id]
	if !ok {
		return nil, false, nil
	}
	clone := *t
	return &clone, true, nil
}
func (m *mockTermRepo) GetActiveForChild(_ context.Context, _, _, childID uuid.UUID) (*domain.Term, bool, error) {
	for _, t := range m.terms {
		if t.ChildID == childID && (t.Status == domain.TermStatusActive || t.Status == domain.TermStatusPreTerm || t.Status == domain.TermStatusPendingRenewal) {
			clone := *t
			return &clone, true, nil
		}
	}
	return nil, false, nil
}
func (m *mockTermRepo) GetActiveForChildInTx(ctx context.Context, tx domain.Tx, t, b, c uuid.UUID) (*domain.Term, bool, error) {
	return m.GetActiveForChild(ctx, t, b, c)
}
func (m *mockTermRepo) GetByIDInTx(ctx context.Context, tx domain.Tx, t, b, id uuid.UUID) (*domain.Term, bool, error) {
	return m.GetByID(ctx, t, b, id)
}
func (m *mockTermRepo) ListByChild(_ context.Context, _, _, childID uuid.UUID) ([]domain.Term, error) {
	out := make([]domain.Term, 0)
	for _, t := range m.terms {
		if t.ChildID == childID {
			out = append(out, *t)
		}
	}
	return out, nil
}

func (m *mockTermRepo) ListByChildPaginated(_ context.Context, _, _, childID uuid.UUID, limit, offset int) ([]domain.Term, error) {
	out := make([]domain.Term, 0)
	for _, t := range m.terms {
		if t.ChildID == childID {
			out = append(out, *t)
		}
	}
	if offset >= len(out) {
		return nil, nil
	}
	end := offset + limit
	if end > len(out) {
		end = len(out)
	}
	return out[offset:end], nil
}

func (m *mockTermRepo) CountByChild(_ context.Context, _, _, childID uuid.UUID) (int, error) {
	count := 0
	for _, t := range m.terms {
		if t.ChildID == childID {
			count++
		}
	}
	return count, nil
}
func (m *mockTermRepo) ListActiveByBranch(_ context.Context, _, _ uuid.UUID) ([]domain.Term, error) {
	out := make([]domain.Term, 0)
	for _, t := range m.terms {
		if t.Status == domain.TermStatusActive || t.Status == domain.TermStatusPreTerm || t.Status == domain.TermStatusPendingRenewal {
			out = append(out, *t)
		}
	}
	return out, nil
}
func (m *mockTermRepo) ListExpiringWithin(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.Term, error) {
	return nil, nil
}

func (m *mockTermRepo) ListExpiringWithinPaginated(_ context.Context, _, _ uuid.UUID, _ time.Time, limit, offset int) ([]domain.Term, error) {
	return nil, nil
}

func (m *mockTermRepo) CountExpiringWithin(_ context.Context, _, _ uuid.UUID, _ time.Time) (int, error) {
	return 0, nil
}
func (m *mockTermRepo) ListEndingOnOrBefore(_ context.Context, _, _ uuid.UUID, _ time.Time) ([]domain.Term, error) {
	return nil, nil
}
func (m *mockTermRepo) ListActiveInBillingMonth(_ context.Context, _, _ uuid.UUID, _, _ time.Time) ([]domain.Term, error) {
	return nil, nil
}
func (m *mockTermRepo) ListActiveForChildUpdate(_ context.Context, _ domain.Tx, _, _, childID uuid.UUID) ([]domain.Term, error) {
	return m.ListByChild(context.Background(), uuid.Nil, uuid.Nil, childID)
}
func (m *mockTermRepo) SetChildCurrentTermID(_ context.Context, _ domain.Tx, _, _, _, _ uuid.UUID) error {
	return nil
}
func (m *mockTermRepo) ClearChildCurrentTermID(_ context.Context, _ domain.Tx, _, _, _ uuid.UUID) error {
	return nil
}

type mockPatternLookup struct{}

func (m *mockPatternLookup) ExistsInScope(_ context.Context, _ pgx.Tx, _, _, _ uuid.UUID) (bool, error) {
	return true, nil
}

type mockRateProvider struct {
	rate int
}

func (m *mockRateProvider) SiteHourlyRateMinor(_ context.Context, _ pgx.Tx, _, _ uuid.UUID) (int, bool, error) {
	if m.rate <= 0 {
		return 0, false, nil
	}
	return m.rate, true, nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// createTermApplicationTest is intentionally a *scaffolding* test that validates
// the wiring compiles. Real coverage of the use case requires a transaction manager
// against a real Postgres test database; see repository tests.
func TestCreateTermUseCase_CompilesAndBuilds(t *testing.T) {
	repo := newMockTermRepo()
	_ = application.NewCreateTermUseCase(repo, nil, audit.NewWriter(), &mockPatternLookup{}, &mockRateProvider{rate: 750})
}

func TestCreateTermDomainValidation_NoManagerRequired(t *testing.T) {
	// Direct domain test: the NewTerm constructor doesn't need a manager.
	t1 := uuid.New()
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	term, err := domain.NewTerm(t1, uuid.New(), uuid.New(), uuid.New(), start, uuid.New(), 750, uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if term.Status != domain.TermStatusPreTerm {
		t.Errorf("expected pre_term, got %s", term.Status)
	}
}

// Ensure the actor + tenant shim type-checks.
var _ tenant.ActorContext = tenant.ActorContext{}
