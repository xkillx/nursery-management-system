package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/rooms/application"
	"nursery-management-system/api/internal/modules/rooms/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

// ── Fake repositories ──────────────────────────────────────────────────────

type fakeRoomRepo struct {
	rooms             []domain.Room
	getErr            error
	createErr         error
	updateErr         error
	archiveErr        error
	reactivateErr     error
	activeNameExists  bool
	activeNameErr     error
	countActive       int
	countActiveErr    error
	exists            bool
	existsErr         error
	getForUpdate      *domain.Room
	getForUpdateErr   error
	assignedCounts    map[uuid.UUID]int
	assignedCountsErr error
}

func (f *fakeRoomRepo) ListByBranch(ctx context.Context, tenantID, branchID uuid.UUID, includeArchived bool) ([]domain.Room, error) {
	return f.rooms, nil
}

func (f *fakeRoomRepo) GetByID(ctx context.Context, tenantID, branchID, roomID uuid.UUID) (domain.Room, error) {
	if f.getErr != nil {
		return domain.Room{}, f.getErr
	}
	for _, r := range f.rooms {
		if r.ID == roomID {
			return r, nil
		}
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == roomID {
		return *f.getForUpdate, nil
	}
	return domain.Room{}, errors.New("room_not_found: Room not found.")
}

func (f *fakeRoomRepo) Create(ctx context.Context, room domain.Room) error {
	f.rooms = append(f.rooms, room)
	return f.createErr
}

func (f *fakeRoomRepo) Update(ctx context.Context, tenantID, branchID, roomID uuid.UUID, fields map[string]any) (int64, error) {
	if f.updateErr != nil {
		return 0, f.updateErr
	}
	return 1, nil
}

func (f *fakeRoomRepo) Archive(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) error {
	if f.archiveErr != nil {
		return f.archiveErr
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == roomID {
		f.getForUpdate.IsActive = false
	}
	return nil
}

func (f *fakeRoomRepo) Reactivate(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) error {
	if f.reactivateErr != nil {
		return f.reactivateErr
	}
	if f.getForUpdate != nil && f.getForUpdate.ID == roomID {
		f.getForUpdate.IsActive = true
	}
	return nil
}

func (f *fakeRoomRepo) ActiveNameExists(ctx context.Context, tenantID, branchID uuid.UUID, name string, excludeRoomID *uuid.UUID) (bool, error) {
	return f.activeNameExists, f.activeNameErr
}

func (f *fakeRoomRepo) CountActiveChildren(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) (int, error) {
	return f.countActive, f.countActiveErr
}

func (f *fakeRoomRepo) Exists(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) (bool, error) {
	return f.exists, f.existsErr
}

func (f *fakeRoomRepo) GetByIDForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, roomID uuid.UUID) (domain.Room, error) {
	if f.getForUpdate != nil {
		return *f.getForUpdate, f.getForUpdateErr
	}
	return domain.Room{}, errors.New("room_not_found: Room not found.")
}

func (f *fakeRoomRepo) CountAssignedChildrenByBranch(ctx context.Context, tenantID, branchID uuid.UUID) (map[uuid.UUID]int, error) {
	if f.assignedCountsErr != nil {
		return nil, f.assignedCountsErr
	}
	if f.assignedCounts == nil {
		return map[uuid.UUID]int{}, nil
	}
	return f.assignedCounts, nil
}

type fakeSiteChecker struct {
	exists bool
	err    error
}

func (f *fakeSiteChecker) SiteExists(ctx context.Context, tenantID, siteID uuid.UUID) (bool, error) {
	return f.exists, f.err
}

// ── Fake transaction manager ───────────────────────────────────────────────

type fakeTxManager struct{}

func (f *fakeTxManager) ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

// ── Helper actors ───────────────────────────────────────────────────────────

func ownerActor(tenantID uuid.UUID) application.RoomActor {
	return application.NewOwnerRoomActor(tenant.OwnerActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
	})
}

func managerActor(tenantID, branchID uuid.UUID) application.RoomActor {
	return application.NewManagerRoomActor(tenant.ActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
		BranchID: branchID,
	})
}

func practitionerActor(tenantID, branchID uuid.UUID) application.RoomActor {
	return application.NewPractitionerRoomActor(tenant.ActorContext{
		UserID:   uuid.New(),
		TenantID: tenantID,
		BranchID: branchID,
	})
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestCreateRoom_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeRoomRepo{}
	checker := &fakeSiteChecker{exists: true}

	uc := application.NewCreateRoom(repo, checker)
	room, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, application.CreateRoomParams{
		Name:     "Baby Room",
		AgeGroup: "baby",
		Capacity: 12,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if room.Name != "Baby Room" {
		t.Errorf("got name %q, want %q", room.Name, "Baby Room")
	}
	if room.AgeGroup != "baby" {
		t.Errorf("got age_group %q, want %q", room.AgeGroup, "baby")
	}
	if room.Capacity != 12 {
		t.Errorf("got capacity %d, want %d", room.Capacity, 12)
	}
	if !room.IsActive {
		t.Error("room should be active")
	}
}

func TestCreateRoom_DuplicateName(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeRoomRepo{activeNameExists: true}
	checker := &fakeSiteChecker{exists: true}

	uc := application.NewCreateRoom(repo, checker)
	_, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, application.CreateRoomParams{
		Name:     "Baby Room",
		AgeGroup: "baby",
		Capacity: 12,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "room_name_duplicate: An active room with this name already exists in this site." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateRoom_InvalidAgeGroup(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeRoomRepo{}
	checker := &fakeSiteChecker{exists: true}

	uc := application.NewCreateRoom(repo, checker)
	_, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, application.CreateRoomParams{
		Name:     "Room",
		AgeGroup: "invalid",
		Capacity: 12,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "invalid_age_group: Invalid request payload." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateRoom_EmptyName(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeRoomRepo{}
	checker := &fakeSiteChecker{exists: true}

	uc := application.NewCreateRoom(repo, checker)
	_, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, application.CreateRoomParams{
		Name:     "",
		AgeGroup: "baby",
		Capacity: 12,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRoom_ZeroCapacity(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeRoomRepo{}
	checker := &fakeSiteChecker{exists: true}

	uc := application.NewCreateRoom(repo, checker)
	_, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, application.CreateRoomParams{
		Name:     "Room",
		AgeGroup: "baby",
		Capacity: 0,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRoom_ManagerWrongSite(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	wrongSiteID := uuid.New()
	repo := &fakeRoomRepo{}
	checker := &fakeSiteChecker{exists: true}

	uc := application.NewCreateRoom(repo, checker)
	_, err := uc.Execute(context.Background(), managerActor(tenantID, siteID), wrongSiteID, application.CreateRoomParams{
		Name:     "Room",
		AgeGroup: "baby",
		Capacity: 12,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "forbidden_site_scope: Access denied." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestArchiveRoom_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	roomID := uuid.New()
	repo := &fakeRoomRepo{
		getForUpdate: &domain.Room{ID: roomID, IsActive: true},
	}
	txMgr := &fakeTxManager{}

	uc := application.NewArchiveRoom(repo, txMgr, nil, nil)
	err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, roomID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestArchiveRoom_HasChildren(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	roomID := uuid.New()
	repo := &fakeRoomRepo{
		getForUpdate: &domain.Room{ID: roomID, IsActive: true},
		countActive:  3,
	}
	txMgr := &fakeTxManager{}

	uc := application.NewArchiveRoom(repo, txMgr, nil, nil)
	err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, roomID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "room_has_children: Room has 3 active children assigned — reassign them before archiving." {
		t.Errorf("unexpected error: %v", err)
	}
	domainErr, ok := err.(*domainerrors.DomainError)
	if !ok {
		t.Fatalf("expected *DomainError, got %T", err)
	}
	if domainErr.Details["assigned_count"] != 3 {
		t.Errorf("Details[assigned_count] = %v, want 3", domainErr.Details["assigned_count"])
	}
}

func TestArchiveRoom_AlreadyArchived(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	roomID := uuid.New()
	repo := &fakeRoomRepo{
		getForUpdate: &domain.Room{ID: roomID, IsActive: false},
	}
	txMgr := &fakeTxManager{}

	uc := application.NewArchiveRoom(repo, txMgr, nil, nil)
	err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, roomID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReactivateRoom_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	roomID := uuid.New()
	repo := &fakeRoomRepo{
		getForUpdate: &domain.Room{ID: roomID, IsActive: false},
	}
	txMgr := &fakeTxManager{}

	uc := application.NewReactivateRoom(repo, txMgr, nil, nil)
	room, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, roomID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if room.ID != roomID {
		t.Errorf("got room ID %v, want %v", room.ID, roomID)
	}
}

func TestReactivateRoom_AlreadyActive(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	roomID := uuid.New()
	repo := &fakeRoomRepo{
		getForUpdate: &domain.Room{ID: roomID, IsActive: true},
	}
	txMgr := &fakeTxManager{}

	uc := application.NewReactivateRoom(repo, txMgr, nil, nil)
	room, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, roomID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if room.ID != roomID {
		t.Errorf("got room ID %v, want %v", room.ID, roomID)
	}
}

func TestListRooms_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeRoomRepo{
		rooms: []domain.Room{
			{ID: uuid.New(), Name: "Room A"},
			{ID: uuid.New(), Name: "Room B"},
		},
	}

	uc := application.NewListRooms(repo)
	rooms, counts, err := uc.Execute(context.Background(), managerActor(tenantID, siteID), siteID, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("got %d rooms, want 2", len(rooms))
	}
	if counts != nil {
		t.Errorf("counts should be nil when includeOccupancy=false, got %v", counts)
	}
}

func TestListRooms_ManagerWrongSite(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	wrongSiteID := uuid.New()
	repo := &fakeRoomRepo{}

	uc := application.NewListRooms(repo)
	_, _, err := uc.Execute(context.Background(), managerActor(tenantID, siteID), wrongSiteID, false, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListRooms_IncludeOccupancy(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	roomA := uuid.New()
	roomB := uuid.New()
	repo := &fakeRoomRepo{
		rooms: []domain.Room{
			{ID: roomA, Name: "Room A", Capacity: 5},
			{ID: roomB, Name: "Room B", Capacity: 8},
		},
		assignedCounts: map[uuid.UUID]int{
			roomA: 6,
			roomB: 3,
		},
	}

	uc := application.NewListRooms(repo)
	rooms, counts, err := uc.Execute(context.Background(), managerActor(tenantID, siteID), siteID, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != 2 {
		t.Fatalf("got %d rooms, want 2", len(rooms))
	}
	if counts[roomA] != 6 {
		t.Errorf("room A count = %d, want 6", counts[roomA])
	}
	if counts[roomB] != 3 {
		t.Errorf("room B count = %d, want 3", counts[roomB])
	}
}

func TestGetRoom_Success(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	roomID := uuid.New()
	repo := &fakeRoomRepo{
		rooms: []domain.Room{
			{ID: roomID, Name: "Room A"},
		},
	}

	uc := application.NewGetRoom(repo)
	room, err := uc.Execute(context.Background(), managerActor(tenantID, siteID), siteID, roomID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if room.ID != roomID {
		t.Errorf("got room ID %v, want %v", room.ID, roomID)
	}
}

func TestUpdateRoom_ChangeNameToDuplicate(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	roomID := uuid.New()
	repo := &fakeRoomRepo{
		rooms: []domain.Room{
			{ID: roomID, Name: "Original Name"},
		},
		activeNameExists: true,
	}
	checker := &fakeSiteChecker{exists: true}

	uc := application.NewUpdateRoom(repo, checker)
	name := "Baby Room"
	_, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, roomID, application.UpdateRoomParams{
		Name: &name,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "room_name_duplicate: An active room with this name already exists in this site." {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPractitionerCanListRooms(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeRoomRepo{
		rooms: []domain.Room{{ID: uuid.New(), Name: "Room A"}},
	}

	uc := application.NewListRooms(repo)
	rooms, _, err := uc.Execute(context.Background(), practitionerActor(tenantID, siteID), siteID, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rooms) != 1 {
		t.Errorf("got %d rooms, want 1", len(rooms))
	}
}

func TestSiteNotFound(t *testing.T) {
	tenantID := uuid.New()
	siteID := uuid.New()
	repo := &fakeRoomRepo{}
	checker := &fakeSiteChecker{exists: false}

	uc := application.NewCreateRoom(repo, checker)
	_, err := uc.Execute(context.Background(), ownerActor(tenantID), siteID, application.CreateRoomParams{
		Name:     "Room",
		AgeGroup: "baby",
		Capacity: 12,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "site_not_found: Site not found." {
		t.Errorf("unexpected error: %v", err)
	}
}
