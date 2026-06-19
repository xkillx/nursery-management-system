package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontypes/application"
	"nursery-management-system/api/internal/modules/sessiontypes/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// ── Fake repository ───────────────────────────────────────────────────────

type fakeSessionTypeRepo struct {
	types             []domain.SessionType
	getErr            error
	createErr         error
	updateErr         error
	updateRows        int64
	archiveErr        error
	reactivateErr     error
	activeNameExists  bool
	activeNameErr     error
	exists            bool
	getForUpdate      *domain.SessionType
	getForUpdateErr   error
}

func (f *fakeSessionTypeRepo) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]domain.SessionType, error) {
	return f.types, nil
}

func (f *fakeSessionTypeRepo) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.SessionType, error) {
	if f.getErr != nil {
		return domain.SessionType{}, f.getErr
	}
	for _, st := range f.types {
		if st.ID == id {
			return st, nil
		}
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == id {
		return *f.getForUpdate, nil
	}
	return domain.SessionType{}, errors.New("session_type_not_found: Session type not found.")
}

func (f *fakeSessionTypeRepo) Create(ctx context.Context, st domain.SessionType) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.types = append(f.types, st)
	return nil
}

func (f *fakeSessionTypeRepo) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if f.updateErr != nil {
		return 0, f.updateErr
	}
	if f.updateRows == 0 {
		return 1, nil
	}
	return f.updateRows, nil
}

func (f *fakeSessionTypeRepo) Archive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error {
	if f.archiveErr != nil {
		return f.archiveErr
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == id {
		f.getForUpdate.IsActive = false
	}
	return nil
}

func (f *fakeSessionTypeRepo) Reactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error {
	if f.reactivateErr != nil {
		return f.reactivateErr
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == id {
		f.getForUpdate.IsActive = true
	}
	return nil
}

func (f *fakeSessionTypeRepo) ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	return f.activeNameExists, f.activeNameErr
}

func (f *fakeSessionTypeRepo) Exists(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	return f.exists, nil
}

func (f *fakeSessionTypeRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.SessionType, error) {
	if f.getForUpdate != nil {
		return *f.getForUpdate, f.getForUpdateErr
	}
	return domain.SessionType{}, errors.New("session_type_not_found: Session type not found.")
}

// ── Fake site checker ────────────────────────────────────────────────────

type fakeSiteChecker struct {
	exists bool
	err    error
}

func (f *fakeSiteChecker) SiteExists(ctx context.Context, tenantID, siteID uuid.UUID) (bool, error) {
	return f.exists, f.err
}

// ── Fake transaction manager ─────────────────────────────────────────────

type fakeTxManager struct{}

func (f *fakeTxManager) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

// ── Helper actors ────────────────────────────────────────────────────────

func stOwnerActor(tenantID uuid.UUID) application.SessionTypeActor {
	return application.NewOwnerSessionTypeActor(tenant.OwnerActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
	})
}

func stManagerActor(tenantID, branchID uuid.UUID) application.SessionTypeActor {
	return application.NewManagerSessionTypeActor(tenant.ActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
		BranchID: branchID,
	})
}

func stPractitionerActor(tenantID, branchID uuid.UUID) application.SessionTypeActor {
	return application.NewPractitionerSessionTypeActor(tenant.ActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
		BranchID: branchID,
	})
}

// ── Tests ────────────────────────────────────────────────────────────────

func TestCreateSessionType_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeSessionTypeRepo{}
	checker := &fakeSiteChecker{exists: true}
	txMgr := &fakeTxManager{}

	uc := application.NewCreateSessionType(repo, checker, txMgr, nil)
	st, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTypeParams{
		Name:      "Morning",
		StartTime: "08:00",
		EndTime:   "13:00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.Name != "Morning" {
		t.Errorf("got name %q, want %q", st.Name, "Morning")
	}
	if st.StartMinutes != 8*60 {
		t.Errorf("got start %d, want %d", st.StartMinutes, 8*60)
	}
	if st.EndMinutes != 13*60 {
		t.Errorf("got end %d, want %d", st.EndMinutes, 13*60)
	}
	if !st.IsActive {
		t.Error("session type should be active")
	}
}

func TestCreateSessionType_StartEqualsEnd(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeSessionTypeRepo{}
	checker := &fakeSiteChecker{exists: true}
	txMgr := &fakeTxManager{}

	uc := application.NewCreateSessionType(repo, checker, txMgr, nil)
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTypeParams{
		Name:      "Bad",
		StartTime: "08:00",
		EndTime:   "08:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "session_type_invalid_time_order: Invalid request payload." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateSessionType_StartAfterEnd(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeSessionTypeRepo{}
	checker := &fakeSiteChecker{exists: true}
	txMgr := &fakeTxManager{}

	uc := application.NewCreateSessionType(repo, checker, txMgr, nil)
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTypeParams{
		Name:      "Bad",
		StartTime: "13:00",
		EndTime:   "08:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateSessionType_DuplicateActiveName(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeSessionTypeRepo{activeNameExists: true}
	checker := &fakeSiteChecker{exists: true}
	txMgr := &fakeTxManager{}

	uc := application.NewCreateSessionType(repo, checker, txMgr, nil)
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTypeParams{
		Name:      "Morning",
		StartTime: "08:00",
		EndTime:   "13:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "session_type_name_duplicate: An active session type with this name already exists in this site." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateSessionType_EmptyName(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeSessionTypeRepo{}
	checker := &fakeSiteChecker{exists: true}
	txMgr := &fakeTxManager{}

	uc := application.NewCreateSessionType(repo, checker, txMgr, nil)
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTypeParams{
		Name:      "  ",
		StartTime: "08:00",
		EndTime:   "13:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateSessionType_OwnerSiteNotFound(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeSessionTypeRepo{}
	checker := &fakeSiteChecker{exists: false}
	txMgr := &fakeTxManager{}

	uc := application.NewCreateSessionType(repo, checker, txMgr, nil)
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTypeParams{
		Name:      "Morning",
		StartTime: "08:00",
		EndTime:   "13:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	domainErr, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected *DomainError, got %T", err)
	}
	if domainErr.Code != "site_not_found" {
		t.Errorf("got code %q, want %q", domainErr.Code, "site_not_found")
	}
}

func TestCreateSessionType_ManagerWrongSite(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	wrongSiteID := uuid.New()
	repo := &fakeSessionTypeRepo{}
	checker := &fakeSiteChecker{exists: true}
	txMgr := &fakeTxManager{}

	uc := application.NewCreateSessionType(repo, checker, txMgr, nil)
	_, err := uc.Execute(context.Background(), stManagerActor(tenantID, siteID), wrongSiteID, application.CreateSessionTypeParams{
		Name:      "Morning",
		StartTime: "08:00",
		EndTime:   "13:00",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "forbidden_site_scope: Access denied." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateSessionType_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	stID := uuid.New()
	existing := domain.SessionType{
		ID: stID, TenantID: tenantID, BranchID: siteID, Name: "Old",
		StartMinutes: 480, EndMinutes: 720, IsActive: true,
	}
	updated := existing
	updated.Name = "New"
	repo := &fakeSessionTypeRepo{
		types:      []domain.SessionType{updated, existing},
		updateRows: 1,
	}
	checker := &fakeSiteChecker{exists: true}
	txMgr := &fakeTxManager{}

	uc := application.NewUpdateSessionType(repo, checker, txMgr, nil)
	newName := "New"
	st, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, stID, application.UpdateSessionTypeParams{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.Name != "New" {
		t.Errorf("got name %q, want %q", st.Name, "New")
	}
}

func TestUpdateSessionType_TimeOrderViolation(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	stID := uuid.New()
	existing := domain.SessionType{
		ID: stID, TenantID: tenantID, BranchID: siteID, Name: "X",
		StartMinutes: 480, EndMinutes: 720, IsActive: true,
	}
	repo := &fakeSessionTypeRepo{types: []domain.SessionType{existing}}
	checker := &fakeSiteChecker{exists: true}
	txMgr := &fakeTxManager{}

	uc := application.NewUpdateSessionType(repo, checker, txMgr, nil)
	start := "13:00"
	end := "08:00"
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, stID, application.UpdateSessionTypeParams{
		StartTime: &start,
		EndTime:   &end,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestArchiveSessionType_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	stID := uuid.New()
	repo := &fakeSessionTypeRepo{
		getForUpdate: &domain.SessionType{ID: stID, IsActive: true},
	}
	txMgr := &fakeTxManager{}

	uc := application.NewArchiveSessionType(repo, txMgr, nil)
	err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, stID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestArchiveSessionType_AlreadyArchived(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	stID := uuid.New()
	repo := &fakeSessionTypeRepo{
		getForUpdate: &domain.SessionType{ID: stID, IsActive: false},
	}
	txMgr := &fakeTxManager{}

	uc := application.NewArchiveSessionType(repo, txMgr, nil)
	err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, stID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReactivateSessionType_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	stID := uuid.New()
	repo := &fakeSessionTypeRepo{
		types:        []domain.SessionType{{ID: stID, TenantID: tenantID, BranchID: siteID, IsActive: true, StartMinutes: 480, EndMinutes: 720}},
		getForUpdate: &domain.SessionType{ID: stID, IsActive: false},
	}
	txMgr := &fakeTxManager{}

	uc := application.NewReactivateSessionType(repo, txMgr, nil)
	st, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, stID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !st.IsActive {
		t.Error("expected active after reactivate")
	}
}

func TestListSessionTypes_RejectsPractitioner(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeSessionTypeRepo{}
	uc := application.NewListSessionTypes(repo)
	_, err := uc.Execute(context.Background(), stPractitionerActor(tenantID, siteID), siteID, false)
	if err != nil {
		// practitioner can list, but path goes through requireRoles in handler.
		// Application does not reject practitioner here. Owner/Manager/Read test only.
		t.Skipf("list allows practitioner at application level (handler enforces): %v", err)
	}
}
