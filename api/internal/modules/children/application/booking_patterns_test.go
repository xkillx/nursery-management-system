package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/application"
	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// ── Fakes ─────────────────────────────────────────────────────────────────

type fakeChildBPRepo struct {
	// state
	patternsByChildID  map[uuid.UUID][]domain.BookingPattern
	currentOpenByChild map[uuid.UUID]*domain.BookingPattern
	previousByChild    map[uuid.UUID]*domain.BookingPattern
	existsInScope      bool

	// errs
	getByIDErr        error
	existsErr         error
	listErr           error
	insertErr         error
	closeCurrentErr   error
	closeByIDErr      error
	replaceErr        error
	updateEffFromErr  error
}

func newFakeRepo() *fakeChildBPRepo {
	return &fakeChildBPRepo{
		patternsByChildID:  map[uuid.UUID][]domain.BookingPattern{},
		currentOpenByChild: map[uuid.UUID]*domain.BookingPattern{},
		previousByChild:    map[uuid.UUID]*domain.BookingPattern{},
		existsInScope:      true,
	}
}

func (f *fakeChildBPRepo) ListByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.BookingPattern, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.patternsByChildID[childID], nil
}

func (f *fakeChildBPRepo) GetPatternByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (*domain.BookingPattern, bool, error) {
	for _, ps := range f.patternsByChildID {
		for i := range ps {
			if ps[i].ID == id {
				cp := ps[i]
				return &cp, true, nil
			}
		}
	}
	return nil, false, nil
}

func (f *fakeChildBPRepo) GetActiveForDate(ctx context.Context, tenantID, branchID, childID uuid.UUID, date time.Time) (*domain.BookingPattern, bool, error) {
	ps := f.patternsByChildID[childID]
	for i := range ps {
		p := ps[i]
		if (p.EffectiveFrom.Equal(date) || p.EffectiveFrom.Before(date)) &&
			(p.EffectiveTo == nil || !p.EffectiveTo.Before(date)) {
			return &p, true, nil
		}
	}
	return nil, false, nil
}

func (f *fakeChildBPRepo) GetCurrentOpenByChild(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (*domain.BookingPattern, bool, error) {
	p, ok := f.currentOpenByChild[childID]
	if !ok {
		return nil, false, nil
	}
	return p, true, nil
}

func (f *fakeChildBPRepo) GetPreviousClosedByChild(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (*domain.BookingPattern, bool, error) {
	p, ok := f.previousByChild[childID]
	if !ok {
		return nil, false, nil
	}
	return p, true, nil
}

func (f *fakeChildBPRepo) InsertPattern(ctx context.Context, tx pgx.Tx, p *domain.BookingPattern, entries []domain.BookingPatternEntry) (*domain.BookingPattern, error) {
	if f.insertErr != nil {
		return nil, f.insertErr
	}
	p.IsCurrent = true
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	// Build entry joined refs.
	for i := range entries {
		entries[i].PatternID = p.ID
		entries[i].TenantID = p.TenantID
		entries[i].BranchID = p.BranchID
	}
	p.Entries = entries
	f.patternsByChildID[p.ChildID] = append(f.patternsByChildID[p.ChildID], *p)
	f.currentOpenByChild[p.ChildID] = p
	return p, nil
}

func (f *fakeChildBPRepo) CloseCurrentPattern(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, effectiveTo time.Time) error {
	if f.closeCurrentErr != nil {
		return f.closeCurrentErr
	}
	if p, ok := f.currentOpenByChild[childID]; ok {
		et := effectiveTo
		p.EffectiveTo = &et
		p.IsCurrent = false
		f.previousByChild[childID] = p
		delete(f.currentOpenByChild, childID)
	}
	return nil
}

func (f *fakeChildBPRepo) ClosePatternByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID, effectiveTo time.Time) error {
	if f.closeByIDErr != nil {
		return f.closeByIDErr
	}
	for cid, ps := range f.patternsByChildID {
		for i := range ps {
			if ps[i].ID == patternID {
				et := effectiveTo
				ps[i].EffectiveTo = &et
				ps[i].IsCurrent = false
				f.previousByChild[cid] = &ps[i]
				if cur, ok := f.currentOpenByChild[cid]; ok && cur.ID == patternID {
					delete(f.currentOpenByChild, cid)
				}
			}
		}
	}
	return nil
}

func (f *fakeChildBPRepo) ReplaceEntries(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID, entries []domain.BookingPatternEntry) error {
	if f.replaceErr != nil {
		return f.replaceErr
	}
	for _, ps := range f.patternsByChildID {
		for i := range ps {
			if ps[i].ID == patternID {
				ps[i].Entries = entries
			}
		}
	}
	return nil
}

func (f *fakeChildBPRepo) UpdateEffectiveFrom(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID, effectiveFrom time.Time) error {
	if f.updateEffFromErr != nil {
		return f.updateEffFromErr
	}
	for _, ps := range f.patternsByChildID {
		for i := range ps {
			if ps[i].ID == patternID {
				ps[i].EffectiveFrom = effectiveFrom
			}
		}
	}
	return nil
}

// Stub implementations for the rest of domain.Repository.

func (f *fakeChildBPRepo) List(ctx context.Context, tenantID, branchID uuid.UUID, filter domain.StatusFilter, limit, offset int) ([]domain.Child, error) {
	return nil, nil
}
func (f *fakeChildBPRepo) Count(ctx context.Context, tenantID, branchID uuid.UUID, filter domain.StatusFilter) (int, error) {
	return 0, nil
}
func (f *fakeChildBPRepo) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	if f.getByIDErr != nil {
		return domain.Child{}, false, f.getByIDErr
	}
	return domain.Child{ID: id}, true, nil
}
func (f *fakeChildBPRepo) Create(ctx context.Context, tx pgx.Tx, child domain.Child, notes string, tenantID, branchID uuid.UUID) error {
	return nil
}
func (f *fakeChildBPRepo) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	return 0, nil
}
func (f *fakeChildBPRepo) MarkInactive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error {
	return nil
}
func (f *fakeChildBPRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.Child, bool, error) {
	return domain.Child{ID: id}, true, nil
}
func (f *fakeChildBPRepo) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	return f.existsInScope, f.existsErr
}
func (f *fakeChildBPRepo) ListAttendance(ctx context.Context, tenantID, branchID uuid.UUID, localDate time.Time) ([]domain.AttendanceChild, error) {
	return nil, nil
}
func (f *fakeChildBPRepo) GetChildForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.ChildCorrectionInfo, bool, error) {
	return domain.ChildCorrectionInfo{}, false, nil
}
func (f *fakeChildBPRepo) GetForAttendanceCheck(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (domain.Child, bool, error) {
	return domain.Child{ID: childID}, true, nil
}
func (f *fakeChildBPRepo) GetProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) GetProfileForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (*domain.ChildProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) InsertProfile(ctx context.Context, tx pgx.Tx, p *domain.ChildProfile) (*domain.ChildProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) UpdateProfile(ctx context.Context, tx pgx.Tx, p *domain.ChildProfile) (*domain.ChildProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) ListContactsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.ChildContact, error) {
	return nil, nil
}
func (f *fakeChildBPRepo) ReplaceContactsForTypes(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, contactTypes []domain.ContactType, entries []domain.ChildContact) error {
	return nil
}
func (f *fakeChildBPRepo) GetHealthByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildHealthProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) UpsertHealth(ctx context.Context, tx pgx.Tx, p *domain.ChildHealthProfile) (*domain.ChildHealthProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) GetSafeguardingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildSafeguardingProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) UpsertSafeguarding(ctx context.Context, tx pgx.Tx, p *domain.ChildSafeguardingProfile) (*domain.ChildSafeguardingProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) GetConsentByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildConsent, bool, error) {
	return nil, false, nil
}
func (f *fakeChildBPRepo) InsertConsent(ctx context.Context, tx pgx.Tx, p *domain.ChildConsent) (*domain.ChildConsent, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) UpdateConsent(ctx context.Context, tx pgx.Tx, p *domain.ChildConsent) (*domain.ChildConsent, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) GetFundingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildFundingRecord, bool, error) {
	return nil, false, nil
}
func (f *fakeChildBPRepo) UpsertFunding(ctx context.Context, tx pgx.Tx, p *domain.ChildFundingRecord) (*domain.ChildFundingRecord, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) GetCollectionSettingByChild(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (*domain.ChildCollectionSetting, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) UpsertCollectionSetting(ctx context.Context, tx pgx.Tx, p *domain.ChildCollectionSetting) (*domain.ChildCollectionSetting, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) SetCollectionPassword(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID, id uuid.UUID, password string, passwordHint string, updatedAt time.Time, userID, membershipID uuid.UUID) error {
	return nil
}
func (f *fakeChildBPRepo) ListRoomAssignmentsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]domain.ChildRoomAssignment, error) {
	return nil, nil
}
func (f *fakeChildBPRepo) GetCurrentRoomAssignmentByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildRoomAssignment, bool, error) {
	return nil, false, nil
}
func (f *fakeChildBPRepo) InsertRoomAssignment(ctx context.Context, tx pgx.Tx, a *domain.ChildRoomAssignment) (*domain.ChildRoomAssignment, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) CloseCurrentRoomAssignment(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, endDate time.Time) error {
	return nil
}
func (f *fakeChildBPRepo) GetRoomAssignmentByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (*domain.ChildRoomAssignment, bool, error) {
	return nil, false, nil
}
func (f *fakeChildBPRepo) CloseRoomAssignmentByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, endDate time.Time) (bool, error) {
	return false, nil
}
func (f *fakeChildBPRepo) GetBillingProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildBillingProfile, bool, error) {
	return nil, false, nil
}
func (f *fakeChildBPRepo) UpsertBillingProfile(ctx context.Context, tx pgx.Tx, p *domain.ChildBillingProfile) (*domain.ChildBillingProfile, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeChildBPRepo) GetLeavingRecordByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*domain.ChildLeavingRecord, bool, error) {
	return nil, false, nil
}
func (f *fakeChildBPRepo) InsertLeavingRecord(ctx context.Context, tx pgx.Tx, p *domain.ChildLeavingRecord) error {
	return nil
}

// ── Session type lookup ───────────────────────────────────────────────────

type fakeSessionTypeLookup struct {
	byID map[uuid.UUID]application.SessionTypeInfo
}

func (f *fakeSessionTypeLookup) GetActiveInScope(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (application.SessionTypeInfo, bool, error) {
	if info, ok := f.byID[sessionTypeID]; ok {
		return info, true, nil
	}
	return application.SessionTypeInfo{}, false, nil
}

// ── TxManager / Audit (using real impls, not tx-mocked) ───────────────────

// Override the fake's ExecTx usage — the children app uses *transaction.Manager.
// We bypass the real txm by injecting a wrapper that doesn't talk to a pool.
type noopTxm struct{}

func (n *noopTxm) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

// ── Helpers ───────────────────────────────────────────────────────────────

func managerActorContext(tenantID, branchID uuid.UUID) tenant.ActorContext {
	return tenant.ActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
		BranchID: branchID,
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────

func TestCreateBookingPattern_Success(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()

	repo := newFakeRepo()
	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "Morning", StartMinutes: 480, EndMinutes: 720, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	result, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsCurrent {
		t.Error("expected is_current true")
	}
	if result.EffectiveFrom.Format("2006-01-02") != "2026-06-20" {
		t.Errorf("got effective_from %s, want 2026-06-20", result.EffectiveFrom.Format("2006-01-02"))
	}
}

func TestCreateBookingPattern_Backdated(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()

	repo := newFakeRepo()
	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "X", StartMinutes: 480, EndMinutes: 720, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "booking_pattern_backdated: Invalid request payload." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateBookingPattern_OverlapRejected(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()
	openID := uuid.New()

	repo := newFakeRepo()
	repo.currentOpenByChild[childID] = &domain.BookingPattern{
		ID:            openID,
		TenantID:      tenantID,
		BranchID:      branchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC),
		IsCurrent:     true,
	}
	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "X", StartMinutes: 480, EndMinutes: 720, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "booking_pattern_overlap: Invalid request payload." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateBookingPattern_ClosesPreviousAdjacently(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()
	openID := uuid.New()

	repo := newFakeRepo()
	repo.currentOpenByChild[childID] = &domain.BookingPattern{
		ID:            openID,
		TenantID:      tenantID,
		BranchID:      branchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC),
		IsCurrent:     true,
	}
	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "X", StartMinutes: 480, EndMinutes: 720, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify the previous was closed with effective_to = new_from - 1.
	prev := repo.previousByChild[childID]
	if prev == nil {
		t.Fatal("expected previous pattern to be set")
	}
	if prev.EffectiveTo == nil {
		t.Fatal("expected previous pattern to be closed")
	}
	if prev.EffectiveTo.Format("2006-01-02") != "2026-06-24" {
		t.Errorf("expected previous closed to 2026-06-24, got %s", prev.EffectiveTo.Format("2006-01-02"))
	}
}

func TestCreateBookingPattern_SessionTypeNotInBranch(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()

	repo := newFakeRepo()
	lookup := &fakeSessionTypeLookup{} // empty
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "session_type_not_in_branch: Invalid request payload." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateBookingPattern_SessionTypeArchived(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()

	repo := newFakeRepo()
	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "X", StartMinutes: 480, EndMinutes: 720, IsActive: false},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "session_type_archived: Invalid request payload." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateBookingPattern_DuplicateDayAndSessionRejected(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()

	repo := newFakeRepo()
	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "X", StartMinutes: 480, EndMinutes: 720, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID},
			{DayOfWeek: 1, SessionTypeID: stID},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "booking_pattern_duplicate_entry: Invalid request payload." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateBookingPattern_MultipleSessionsSameDayAllowed(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID1 := uuid.New()
	stID2 := uuid.New()

	repo := newFakeRepo()
	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID1: {ID: stID1, Name: "Morning", StartMinutes: 480, EndMinutes: 720, IsActive: true},
		stID2: {ID: stID2, Name: "Afternoon", StartMinutes: 780, EndMinutes: 1020, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	result, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID1},
			{DayOfWeek: 1, SessionTypeID: stID2},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Entries) != 2 {
		t.Errorf("got %d entries, want 2", len(result.Entries))
	}
}

func TestCreateBookingPattern_EmptyEntries(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()

	repo := newFakeRepo()
	lookup := &fakeSessionTypeLookup{}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries:       nil,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateBookingPattern_InvalidDayOfWeek(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()

	repo := newFakeRepo()
	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "X", StartMinutes: 480, EndMinutes: 720, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewCreateBookingPattern(repo, nil, txm, lookup, clock)

	effective := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), application.CreateBookingPatternInput{
		EffectiveFrom: effective,
		Entries: []application.BookingPatternEntryInput{
			{DayOfWeek: 8, SessionTypeID: stID},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateBookingPattern_Success(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()
	patternID := uuid.New()

	repo := newFakeRepo()
	effective := time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC)
	pattern := domain.BookingPattern{
		ID:            patternID,
		TenantID:      tenantID,
		BranchID:      branchID,
		ChildID:       childID,
		EffectiveFrom: effective,
		IsCurrent:     true,
	}
	repo.patternsByChildID[childID] = []domain.BookingPattern{pattern}
	repo.currentOpenByChild[childID] = &pattern

	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "X", StartMinutes: 480, EndMinutes: 720, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewUpdateBookingPattern(repo, nil, txm, lookup, clock)

	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), patternID.String(), application.UpdateBookingPatternInput{
		Entries: &[]application.BookingPatternEntryInput{
			{DayOfWeek: 2, SessionTypeID: stID},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := repo.patternsByChildID[childID][0]
	if len(got.Entries) != 1 {
		t.Errorf("got %d entries, want 1", len(got.Entries))
	}
	if got.Entries[0].DayOfWeek != 2 {
		t.Errorf("got day_of_week %d, want 2", got.Entries[0].DayOfWeek)
	}
}

func TestUpdateBookingPattern_StartedRejected(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	patternID := uuid.New()

	repo := newFakeRepo()
	pattern := domain.BookingPattern{
		ID:            patternID,
		TenantID:      tenantID,
		BranchID:      branchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		IsCurrent:     true,
	}
	repo.patternsByChildID[childID] = []domain.BookingPattern{pattern}
	repo.currentOpenByChild[childID] = &pattern

	lookup := &fakeSessionTypeLookup{}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewUpdateBookingPattern(repo, nil, txm, lookup, clock)

	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), patternID.String(), application.UpdateBookingPatternInput{
		Entries: &[]application.BookingPatternEntryInput{},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "booking_pattern_not_editable: Resource not editable." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateBookingPattern_ReclosesPreviousOnDateChange(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	stID := uuid.New()
	prevID := uuid.New()
	patternID := uuid.New()

	repo := newFakeRepo()
	prevClosed := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	prevPattern := domain.BookingPattern{
		ID:            prevID,
		TenantID:      tenantID,
		BranchID:      branchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		EffectiveTo:   &prevClosed,
		IsCurrent:     false,
	}
	repo.previousByChild[childID] = &prevPattern
	pattern := domain.BookingPattern{
		ID:            patternID,
		TenantID:      tenantID,
		BranchID:      branchID,
		ChildID:       childID,
		EffectiveFrom: time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC),
		IsCurrent:     true,
	}
	repo.patternsByChildID[childID] = []domain.BookingPattern{prevPattern, pattern}
	repo.currentOpenByChild[childID] = &pattern

	lookup := &fakeSessionTypeLookup{byID: map[uuid.UUID]application.SessionTypeInfo{
		stID: {ID: stID, Name: "X", StartMinutes: 480, EndMinutes: 720, IsActive: true},
	}}
	txm := &noopTxm{}
	clock := func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) }
	uc := application.NewUpdateBookingPattern(repo, nil, txm, lookup, clock)

	newFrom := time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)
	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), patternID.String(), application.UpdateBookingPatternInput{
		EffectiveFrom: &newFrom,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Previous should be re-closed to newFrom - 1.
	prev := repo.previousByChild[childID]
	if prev == nil || prev.EffectiveTo == nil {
		t.Fatal("expected previous to be re-closed")
	}
	if prev.EffectiveTo.Format("2006-01-02") != "2026-06-21" {
		t.Errorf("expected previous closed to 2026-06-21, got %s", prev.EffectiveTo.Format("2006-01-02"))
	}
	// Current should be moved to newFrom.
	var currentFound bool
	for _, p := range repo.patternsByChildID[childID] {
		if p.ID == patternID {
			if p.EffectiveFrom.Format("2006-01-02") != "2026-06-22" {
				t.Errorf("expected current effective_from 2026-06-22, got %s", p.EffectiveFrom.Format("2006-01-02"))
			}
			currentFound = true
		}
	}
	if !currentFound {
		t.Error("expected to find current pattern")
	}
}

func TestGetCurrentBookingPattern_DefaultToday(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	patternID := uuid.New()

	repo := newFakeRepo()
	repo.patternsByChildID[childID] = []domain.BookingPattern{
		{
			ID:            patternID,
			TenantID:      tenantID,
			BranchID:      branchID,
			ChildID:       childID,
			EffectiveFrom: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			IsCurrent:     true,
		},
	}
	uc := application.NewGetCurrentBookingPattern(repo, func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) })

	p, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != patternID {
		t.Errorf("got pattern %s, want %s", p.ID, patternID)
	}
}

func TestGetCurrentBookingPattern_ByDate(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()
	patternID := uuid.New()

	repo := newFakeRepo()
	repo.patternsByChildID[childID] = []domain.BookingPattern{
		{
			ID:            patternID,
			TenantID:      tenantID,
			BranchID:      branchID,
			ChildID:       childID,
			EffectiveFrom: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			IsCurrent:     true,
		},
	}
	uc := application.NewGetCurrentBookingPattern(repo, func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) })

	p, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), "2026-06-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != patternID {
		t.Errorf("got pattern %s, want %s", p.ID, patternID)
	}
}

func TestGetCurrentBookingPattern_NotFound(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()

	repo := newFakeRepo()
	uc := application.NewGetCurrentBookingPattern(repo, func() time.Time { return time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC) })

	_, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	domainErr, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected *DomainError, got %T", err)
	}
	if domainErr.Code != "booking_pattern_not_found" {
		t.Errorf("got code %q, want booking_pattern_not_found", domainErr.Code)
	}
}

func TestListBookingPatterns_Empty(t *testing.T) {
	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()

	repo := newFakeRepo()
	uc := application.NewListBookingPatterns(repo)
	ps, err := uc.Execute(context.Background(), managerActorContext(tenantID, branchID), childID.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ps) != 0 {
		t.Errorf("got %d patterns, want 0", len(ps))
	}
}

// Ensure imports are used (audit referenced for type completeness)
var _ = audit.NewWriter
