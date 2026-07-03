package application_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/application"
	"nursery-management-system/api/internal/modules/children/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type fakeChildRepository struct {
	domain.Repository
	createCallCount     int
	insertPatternCalled bool
	insertPatternFrom   time.Time
	insertPatternTo     *time.Time
	insertPatternCnt    int
}

func (f *fakeChildRepository) Create(ctx context.Context, tx domain.Tx, child domain.Child, notes string, tenantID, branchID uuid.UUID) error {
	f.createCallCount++
	return nil
}
func (f *fakeChildRepository) InsertPattern(ctx context.Context, tx domain.Tx, p *domain.BookingPattern, entries []domain.BookingPatternEntry) (*domain.BookingPattern, error) {
	f.insertPatternCalled = true
	f.insertPatternFrom = p.EffectiveFrom
	f.insertPatternTo = p.EffectiveTo
	f.insertPatternCnt = len(entries)
	return p, nil
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
func (f *fakeChildRepository) GetCurrentOpenByChild(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID) (*domain.BookingPattern, bool, error) {
	return nil, false, nil
}
func (f *fakeChildRepository) CloseCurrentPattern(ctx context.Context, tx domain.Tx, tenantID, branchID, childID uuid.UUID, effectiveTo time.Time) error {
	return nil
}

type fakeCreateLookup struct {
	application.SessionTypeLookup
	activeTypes map[string]bool
}

func (f *fakeCreateLookup) GetActiveInScope(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (application.SessionTypeInfo, bool, error) {
	active, ok := f.activeTypes[sessionTypeID.String()]
	if !ok {
		return application.SessionTypeInfo{}, false, nil
	}
	return application.SessionTypeInfo{
		ID:           sessionTypeID,
		Name:         "Test Type",
		StartMinutes: 480,
		EndMinutes:   720,
		IsActive:     active,
	}, true, nil
}

type fakeCreateTxm struct {
	application.TxManager
}

func (f *fakeCreateTxm) ExecTx(ctx context.Context, fn func(pgx.Tx) error) error {
	return fn(nil)
}

type fakeTermCreator struct {
	application.EnrollmentTermCreator
	createCalled     bool
	childID          uuid.UUID
	termStartDate    time.Time
	bookingPatternID uuid.UUID
}

func (f *fakeTermCreator) CreateEnrollmentTerm(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, childID uuid.UUID, termStartDate time.Time, bookingPatternID uuid.UUID) (uuid.UUID, error) {
	f.createCalled = true
	f.childID = childID
	f.termStartDate = termStartDate
	f.bookingPatternID = bookingPatternID
	return uuid.New(), nil
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
	t.Run("SuccessfulCreateWithBookingPattern", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()
		stID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{stID.String(): true}}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, &fakeTermCreator{}, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		input := newDefaultInput()
		input.BookingPattern = &application.BookingPatternInput{
			EffectiveFrom: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			Entries: []application.BookingPatternEntryInput{
				{DayOfWeek: 1, SessionTypeID: stID},
				{DayOfWeek: 3, SessionTypeID: stID},
			},
		}

		result, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !slices.Contains(result.CreatedSubRecords, "booking_pattern") {
			t.Errorf("expected CreatedSubRecords to contain 'booking_pattern', got %v", result.CreatedSubRecords)
		}
		if !repo.insertPatternCalled {
			t.Error("expected InsertPattern to be called")
		}
		if repo.insertPatternCnt != 2 {
			t.Errorf("expected 2 pattern entries, got %d", repo.insertPatternCnt)
		}
	})

	t.Run("FailsWhenSessionTypeArchived", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()
		stID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{stID.String(): false}}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, &fakeTermCreator{}, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		input := newDefaultInput()
		input.BookingPattern = &application.BookingPatternInput{
			EffectiveFrom: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			Entries: []application.BookingPatternEntryInput{
				{DayOfWeek: 1, SessionTypeID: stID},
			},
		}

		_, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var de *domainerrors.DomainError
		if !errors.As(err, &de) {
			t.Fatalf("expected *DomainError, got %T", err)
		}
		if de.Code != "session_type_archived" {
			t.Errorf("got code %q, want session_type_archived", de.Code)
		}
		if repo.createCallCount > 0 {
			t.Error("expected Create NOT to be called when pattern resolution fails")
		}
	})

	t.Run("SuccessfulCreateWithoutBookingPattern", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{}}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, &fakeTermCreator{}, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		input := newDefaultInput()

		result, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.insertPatternCalled {
			t.Error("expected InsertPattern NOT to be called when no booking pattern")
		}
		if slices.Contains(result.CreatedSubRecords, "booking_pattern") {
			t.Errorf("expected CreatedSubRecords not to contain 'booking_pattern', got %v", result.CreatedSubRecords)
		}
	})

	t.Run("SuccessfulCreateWithBookingPatternEffectiveTo", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()
		stID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{stID.String(): true}}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, &fakeTermCreator{}, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		effectiveTo := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		input := newDefaultInput()
		input.BookingPattern = &application.BookingPatternInput{
			EffectiveFrom: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			EffectiveTo:   &effectiveTo,
			Entries: []application.BookingPatternEntryInput{
				{DayOfWeek: 1, SessionTypeID: stID},
			},
		}

		result, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !slices.Contains(result.CreatedSubRecords, "booking_pattern") {
			t.Errorf("expected CreatedSubRecords to contain 'booking_pattern', got %v", result.CreatedSubRecords)
		}
		if !repo.insertPatternCalled {
			t.Fatal("expected InsertPattern to be called")
		}
		if repo.insertPatternTo == nil {
			t.Fatal("expected EffectiveTo to be set, got nil")
		}
		if !repo.insertPatternTo.Equal(effectiveTo) {
			t.Errorf("expected EffectiveTo %s, got %s", effectiveTo.Format("2006-01-02"), repo.insertPatternTo.Format("2006-01-02"))
		}
	})

	t.Run("EffectiveToBeforeEffectiveFromRejected", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()
		stID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{stID.String(): true}}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, &fakeTermCreator{}, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		effectiveTo := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		input := newDefaultInput()
		input.BookingPattern = &application.BookingPatternInput{
			EffectiveFrom: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			EffectiveTo:   &effectiveTo,
			Entries: []application.BookingPatternEntryInput{
				{DayOfWeek: 1, SessionTypeID: stID},
			},
		}

		_, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var de *domainerrors.DomainError
		if !errors.As(err, &de) {
			t.Fatalf("expected *DomainError, got %T", err)
		}
		if de.Code != "booking_pattern_effective_to_before_from" {
			t.Errorf("got code %q, want booking_pattern_effective_to_before_from", de.Code)
		}
		if repo.insertPatternCalled {
			t.Error("expected InsertPattern NOT to be called when effective_to before effective_from")
		}
	})

	t.Run("EffectiveFromDefaultsToStartDate", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()
		stID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{stID.String(): true}}
		txm := &fakeCreateTxm{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, &fakeTermCreator{}, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		input := newDefaultInput()
		input.BookingPattern = &application.BookingPatternInput{
			Entries: []application.BookingPatternEntryInput{
				{DayOfWeek: 1, SessionTypeID: stID},
			},
		}
		startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

		result, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !repo.insertPatternCalled {
			t.Fatal("expected InsertPattern to be called")
		}
		if !repo.insertPatternFrom.Equal(startDate) {
			t.Errorf("expected effective_from %s, got %s", startDate.Format("2006-01-02"), repo.insertPatternFrom.Format("2006-01-02"))
		}
		if !slices.Contains(result.CreatedSubRecords, "booking_pattern") {
			t.Errorf("expected CreatedSubRecords to contain 'booking_pattern', got %v", result.CreatedSubRecords)
		}
	})

	t.Run("CreatesTermWhenBookingPatternProvided", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()
		stID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{stID.String(): true}}
		txm := &fakeCreateTxm{}
		termCreator := &fakeTermCreator{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, termCreator, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		input := newDefaultInput()
		input.BookingPattern = &application.BookingPatternInput{
			EffectiveFrom: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			Entries: []application.BookingPatternEntryInput{
				{DayOfWeek: 1, SessionTypeID: stID},
			},
		}

		result, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !termCreator.createCalled {
			t.Error("expected CreateEnrollmentTerm to be called")
		}
		if !slices.Contains(result.CreatedSubRecords, "term") {
			t.Errorf("expected CreatedSubRecords to contain 'term', got %v", result.CreatedSubRecords)
		}
		if result.TermID == nil {
			t.Error("expected TermID to be non-nil when term is created")
		}
	})

	t.Run("TermStartDateIsFirstOfStartDateMonth", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()
		stID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{stID.String(): true}}
		txm := &fakeCreateTxm{}
		termCreator := &fakeTermCreator{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, termCreator, func() time.Time {
			return time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		})

		input := newDefaultInput()
		input.Child.StartDate = "2026-06-24"
		input.BookingPattern = &application.BookingPatternInput{
			EffectiveFrom: time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC),
			Entries: []application.BookingPatternEntryInput{
				{DayOfWeek: 2, SessionTypeID: stID},
			},
		}

		_, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !termCreator.createCalled {
			t.Fatal("expected CreateEnrollmentTerm to be called")
		}
		expectedStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
		if !termCreator.termStartDate.Equal(expectedStart) {
			t.Errorf("expected term start date %s, got %s", expectedStart.Format("2006-01-02"), termCreator.termStartDate.Format("2006-01-02"))
		}
	})

	t.Run("SkipsTermWhenNoBookingPattern", func(t *testing.T) {
		tenantID := uuid.New()
		branchID := uuid.New()

		repo := &fakeChildRepository{}
		lookup := &fakeCreateLookup{activeTypes: map[string]bool{}}
		txm := &fakeCreateTxm{}
		termCreator := &fakeTermCreator{}
		uc := application.NewCreateChildWithFullProfile(repo, nil, txm, lookup, termCreator, func() time.Time {
			return time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		})

		input := newDefaultInput()

		result, err := uc.Execute(context.Background(), createActorContext(tenantID, branchID), input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if termCreator.createCalled {
			t.Error("expected CreateEnrollmentTerm NOT to be called when no booking pattern")
		}
		if slices.Contains(result.CreatedSubRecords, "term") {
			t.Errorf("expected CreatedSubRecords not to contain 'term', got %v", result.CreatedSubRecords)
		}
		if result.TermID != nil {
			t.Error("expected TermID to be nil when no term is created")
		}
	})
}
