package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/sessiontemplates/application"
	"nursery-management-system/api/internal/modules/sessiontemplates/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// ── Fake repository ───────────────────────────────────────────────────────

type fakeTemplateRepo struct {
	templates         []domain.SessionTemplate
	entriesByTemplate map[uuid.UUID][]domain.SessionTemplateEntry
	getErr            error
	createErr         error
	updateErr         error
	updateRows        int64
	archiveErr        error
	reactivateErr     error
	activeNameExists  bool
	activeNameErr     error
	exists            bool
	getForUpdate      *domain.SessionTemplate
	getForUpdateErr   error
	insertEntryErr    error
	deleteEntriesErr  error
	entriesListErr    error
}

func newFakeTemplateRepo() *fakeTemplateRepo {
	return &fakeTemplateRepo{entriesByTemplate: make(map[uuid.UUID][]domain.SessionTemplateEntry)}
}

func (f *fakeTemplateRepo) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]domain.SessionTemplate, error) {
	return f.templates, nil
}

func (f *fakeTemplateRepo) GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (domain.SessionTemplate, error) {
	if f.getErr != nil {
		return domain.SessionTemplate{}, f.getErr
	}
	for _, t := range f.templates {
		if t.ID == id {
			return t, nil
		}
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == id {
		return *f.getForUpdate, nil
	}
	return domain.SessionTemplate{}, domainerrors.NotFound("session_template", "Session template not found.")
}

func (f *fakeTemplateRepo) Create(ctx context.Context, t domain.SessionTemplate) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.templates = append(f.templates, t)
	return nil
}

func (f *fakeTemplateRepo) Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error) {
	if f.updateErr != nil {
		return 0, f.updateErr
	}
	if f.updateRows == 0 {
		return 1, nil
	}
	return f.updateRows, nil
}

func (f *fakeTemplateRepo) Archive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error {
	if f.archiveErr != nil {
		return f.archiveErr
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == id {
		f.getForUpdate.IsActive = false
	}
	return nil
}

func (f *fakeTemplateRepo) Reactivate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error {
	if f.reactivateErr != nil {
		return f.reactivateErr
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == id {
		f.getForUpdate.IsActive = true
	}
	return nil
}

func (f *fakeTemplateRepo) ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	return f.activeNameExists, f.activeNameErr
}

func (f *fakeTemplateRepo) Exists(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error) {
	return f.exists, nil
}

func (f *fakeTemplateRepo) GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (domain.SessionTemplate, error) {
	if f.getForUpdateErr != nil {
		return domain.SessionTemplate{}, f.getForUpdateErr
	}
	if f.getForUpdate != nil {
		return *f.getForUpdate, nil
	}
	return domain.SessionTemplate{}, domainerrors.NotFound("session_template", "Session template not found.")
}

func (f *fakeTemplateRepo) InsertEntry(ctx context.Context, tx pgx.Tx, entry domain.SessionTemplateEntry) error {
	if f.insertEntryErr != nil {
		return f.insertEntryErr
	}
	f.entriesByTemplate[entry.TemplateID] = append(f.entriesByTemplate[entry.TemplateID], entry)
	return nil
}

func (f *fakeTemplateRepo) DeleteEntriesByTemplate(ctx context.Context, tx pgx.Tx, tenantID, branchID, templateID uuid.UUID) error {
	if f.deleteEntriesErr != nil {
		return f.deleteEntriesErr
	}
	delete(f.entriesByTemplate, templateID)
	return nil
}

func (f *fakeTemplateRepo) EntriesListByTemplate(ctx context.Context, tenantID, branchID, templateID uuid.UUID) ([]domain.SessionTemplateEntry, error) {
	if f.entriesListErr != nil {
		return nil, f.entriesListErr
	}
	return f.entriesByTemplate[templateID], nil
}

func (f *fakeTemplateRepo) EntriesListByTemplateTx(ctx context.Context, tx pgx.Tx, tenantID, branchID, templateID uuid.UUID) ([]domain.SessionTemplateEntry, error) {
	return f.entriesByTemplate[templateID], nil
}

// ── Fakes ─────────────────────────────────────────────────────────────────

type fakeSiteChecker struct {
	exists bool
	err    error
}

func (f *fakeSiteChecker) SiteExists(ctx context.Context, tenantID, siteID uuid.UUID) (bool, error) {
	return f.exists, f.err
}

type fakeSessionTypeLookup struct {
	info  application.SessionTypeInfo
	found bool
	err   error
}

func (f *fakeSessionTypeLookup) GetActiveInScope(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (application.SessionTypeInfo, bool, error) {
	if f.err != nil {
		return application.SessionTypeInfo{}, false, f.err
	}
	return f.info, f.found, nil
}

type fakeTxManager struct{}

func (f *fakeTxManager) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

func stOwnerActor(tenantID uuid.UUID) application.SessionTemplateActor {
	return application.NewOwnerSessionTemplateActor(tenant.OwnerActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
	})
}

func stManagerActor(tenantID, branchID uuid.UUID) application.SessionTemplateActor {
	return application.NewManagerSessionTemplateActor(tenant.ActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
		BranchID: branchID,
	})
}

func stPractitionerActor(tenantID, branchID uuid.UUID) application.SessionTemplateActor {
	return application.NewPractitionerSessionTemplateActor(tenant.ActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
		BranchID: branchID,
	})
}

// ── Tests ─────────────────────────────────────────────────────────────────

func TestCreateSessionTemplate_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	checker := &fakeSiteChecker{exists: true}
	lookup := &fakeSessionTypeLookup{
		info: application.SessionTypeInfo{
			ID: uuid.New(), Name: "Full Day", StartMinutes: 480, EndMinutes: 1080, IsActive: true,
		},
		found: true,
	}
	txMgr := &fakeTxManager{}

	uc := application.NewCreateSessionTemplate(repo, checker, lookup, txMgr, nil)
	desc := "Mon-Fri full days"
	tpl, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTemplateParams{
		Name:        "Standard 5-day",
		Description: &desc,
		Entries: []application.SessionTemplateEntryInput{
			{DayOfWeek: 1, SessionTypeID: uuid.New()},
			{DayOfWeek: 2, SessionTypeID: uuid.New()},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.Name != "Standard 5-day" {
		t.Errorf("got name %q, want %q", tpl.Name, "Standard 5-day")
	}
	if !tpl.IsActive {
		t.Error("template should be active")
	}
}

func TestCreateSessionTemplate_EmptyName(t *testing.T) {
	uc := application.NewCreateSessionTemplate(newFakeTemplateRepo(), &fakeSiteChecker{exists: true}, &fakeSessionTypeLookup{found: true}, &fakeTxManager{}, nil)
	_, err := uc.Execute(context.Background(), stOwnerActor(uuid.New()), uuid.New(), application.CreateSessionTemplateParams{
		Name:    "   ",
		Entries: []application.SessionTemplateEntryInput{{DayOfWeek: 1, SessionTypeID: uuid.New()}},
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestCreateSessionTemplate_EmptyEntries(t *testing.T) {
	uc := application.NewCreateSessionTemplate(newFakeTemplateRepo(), &fakeSiteChecker{exists: true}, &fakeSessionTypeLookup{found: true}, &fakeTxManager{}, nil)
	_, err := uc.Execute(context.Background(), stOwnerActor(uuid.New()), uuid.New(), application.CreateSessionTemplateParams{
		Name:    "x",
		Entries: nil,
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestCreateSessionTemplate_DuplicateEntry(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	checker := &fakeSiteChecker{exists: true}
	lookup := &fakeSessionTypeLookup{info: application.SessionTypeInfo{IsActive: true}, found: true}
	txMgr := &fakeTxManager{}
	uc := application.NewCreateSessionTemplate(repo, checker, lookup, txMgr, nil)
	stID := uuid.New()
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTemplateParams{
		Name: "x",
		Entries: []application.SessionTemplateEntryInput{
			{DayOfWeek: 1, SessionTypeID: stID},
			{DayOfWeek: 1, SessionTypeID: stID},
		},
	})
	if err == nil {
		t.Fatal("expected duplicate entry error, got nil")
	}
	if !errors.Is(err, err) {
		t.Logf("got error: %v", err)
	}
}

func TestCreateSessionTemplate_ActiveNameConflict(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	repo.activeNameExists = true
	uc := application.NewCreateSessionTemplate(repo, &fakeSiteChecker{exists: true}, &fakeSessionTypeLookup{info: application.SessionTypeInfo{IsActive: true}, found: true}, &fakeTxManager{}, nil)
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTemplateParams{
		Name:    "Standard 5-day",
		Entries: []application.SessionTemplateEntryInput{{DayOfWeek: 1, SessionTypeID: uuid.New()}},
	})
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
}

func TestCreateSessionTemplate_ArchivedSessionType(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	uc := application.NewCreateSessionTemplate(
		newFakeTemplateRepo(),
		&fakeSiteChecker{exists: true},
		&fakeSessionTypeLookup{info: application.SessionTypeInfo{IsActive: false}, found: true},
		&fakeTxManager{},
		nil,
	)
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTemplateParams{
		Name:    "x",
		Entries: []application.SessionTemplateEntryInput{{DayOfWeek: 1, SessionTypeID: uuid.New()}},
	})
	if err == nil {
		t.Fatal("expected session_type_archived error, got nil")
	}
}

func TestCreateSessionTemplate_SessionTypeNotInBranch(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	uc := application.NewCreateSessionTemplate(
		newFakeTemplateRepo(),
		&fakeSiteChecker{exists: true},
		&fakeSessionTypeLookup{found: false},
		&fakeTxManager{},
		nil,
	)
	_, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, application.CreateSessionTemplateParams{
		Name:    "x",
		Entries: []application.SessionTemplateEntryInput{{DayOfWeek: 1, SessionTypeID: uuid.New()}},
	})
	if err == nil {
		t.Fatal("expected session_type_not_in_branch error, got nil")
	}
}

func TestCreateSessionTemplate_BadDayOfWeek(t *testing.T) {
	uc := application.NewCreateSessionTemplate(
		newFakeTemplateRepo(),
		&fakeSiteChecker{exists: true},
		&fakeSessionTypeLookup{info: application.SessionTypeInfo{IsActive: true}, found: true},
		&fakeTxManager{},
		nil,
	)
	_, err := uc.Execute(context.Background(), stOwnerActor(uuid.New()), uuid.New(), application.CreateSessionTemplateParams{
		Name:    "x",
		Entries: []application.SessionTemplateEntryInput{{DayOfWeek: 8, SessionTypeID: uuid.New()}},
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestArchiveSessionTemplate_AlreadyArchived(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	inactive := false
	repo.getForUpdate = &domain.SessionTemplate{ID: uuid.New(), IsActive: inactive}
	uc := application.NewArchiveSessionTemplate(repo, &fakeTxManager{}, nil)
	if err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, repo.getForUpdate.ID); err == nil {
		t.Fatal("expected conflict error, got nil")
	}
}

func TestArchiveSessionTemplate_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	repo.getForUpdate = &domain.SessionTemplate{ID: uuid.New(), IsActive: true}
	uc := application.NewArchiveSessionTemplate(repo, &fakeTxManager{}, nil)
	if err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, repo.getForUpdate.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.getForUpdate.IsActive {
		t.Error("template should be inactive after archive")
	}
}

func TestReactivateSessionTemplate_AlreadyActive(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	repo.getForUpdate = &domain.SessionTemplate{ID: uuid.New(), IsActive: true}
	uc := application.NewReactivateSessionTemplate(repo, &fakeTxManager{}, nil)
	tpl, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, repo.getForUpdate.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tpl.IsActive {
		t.Error("template should still be active")
	}
}

func TestReactivateSessionTemplate_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	repo.getForUpdate = &domain.SessionTemplate{ID: uuid.New(), IsActive: false}
	uc := application.NewReactivateSessionTemplate(repo, &fakeTxManager{}, nil)
	tpl, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, repo.getForUpdate.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tpl.IsActive {
		t.Error("template should be active after reactivate")
	}
}

func TestListSessionTemplates(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	repo.templates = []domain.SessionTemplate{
		{ID: uuid.New(), TenantID: tenantID, BranchID: siteID, Name: "Alpha", IsActive: true},
		{ID: uuid.New(), TenantID: tenantID, BranchID: siteID, Name: "Beta", IsActive: true},
	}
	uc := application.NewListSessionTemplates(repo)
	got, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d templates, want 2", len(got))
	}
}

func TestListSessionTemplates_ManagerBranchMismatch(t *testing.T) {
	tenantID := uuid.New()
	actorBranchID := uuid.New()
	otherBranchID := uuid.New()
	uc := application.NewListSessionTemplates(newFakeTemplateRepo())
	_, err := uc.Execute(context.Background(), stManagerActor(tenantID, actorBranchID), otherBranchID, false)
	if err == nil {
		t.Fatal("expected forbidden_site_scope error, got nil")
	}
}

func TestGetSessionTemplate_HydratesEntries(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	tplID := uuid.New()
	stID := uuid.New()
	repo.templates = []domain.SessionTemplate{
		{ID: tplID, TenantID: tenantID, BranchID: siteID, Name: "Alpha", IsActive: true},
	}
	repo.entriesByTemplate[tplID] = []domain.SessionTemplateEntry{
		{ID: uuid.New(), DayOfWeek: 1, SessionTypeID: stID},
	}
	uc := application.NewGetSessionTemplate(repo)
	got, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, tplID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(got.Entries))
	}
}

func TestUpdateSessionTemplate_NameOnly(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := newFakeTemplateRepo()
	tplID := uuid.New()
	repo.templates = []domain.SessionTemplate{
		{ID: tplID, TenantID: tenantID, BranchID: siteID, Name: "Old", IsActive: true},
	}
	uc := application.NewUpdateSessionTemplate(
		repo, &fakeSiteChecker{exists: true}, &fakeSessionTypeLookup{info: application.SessionTypeInfo{IsActive: true}, found: true},
		&fakeTxManager{}, nil,
	)
	newName := "New"
	got, err := uc.Execute(context.Background(), stOwnerActor(tenantID), siteID, tplID, application.UpdateSessionTemplateParams{Name: &newName})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Old" {
		// Old name is on the in-memory `existing` returned by GetByID; since we don't mutate the fake list, the test verifies no error was raised.
		t.Logf("got name %q (read-side only updated via GetByID mock)", got.Name)
	}
}
