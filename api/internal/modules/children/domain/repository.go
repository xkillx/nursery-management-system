package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Tx is a transaction interface matching pgx.Tx for dependency injection.
type Tx = pgx.Tx

type StatusFilter string

const (
	StatusActive   StatusFilter = "active"
	StatusInactive StatusFilter = "inactive"
	StatusAll      StatusFilter = "all"
)

type ChildCorrectionInfo struct {
	ID        uuid.UUID
	StartDate time.Time
	EndDate   *time.Time
}

type ChildIdentityRepository interface {
	List(ctx context.Context, tenantID, branchID uuid.UUID, filter StatusFilter, limit, offset int) ([]Child, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (Child, bool, error)
	Create(ctx context.Context, child Child, notes string, tenantID, branchID uuid.UUID) error
	Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	MarkInactive(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) error
	GetByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (Child, bool, error)
	ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (bool, error)
	ListAttendance(ctx context.Context, tenantID, branchID uuid.UUID, localDate time.Time) ([]AttendanceChild, error)
	GetChildForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (ChildCorrectionInfo, bool, error)
}

type ChildProfileRepository interface {
	GetProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildProfile, error)
	GetProfileForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (*ChildProfile, error)
	InsertProfile(ctx context.Context, tx pgx.Tx, p *ChildProfile) (*ChildProfile, error)
	UpdateProfile(ctx context.Context, tx pgx.Tx, p *ChildProfile) (*ChildProfile, error)
}

type ChildContactRepository interface {
	ListContactsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]ChildContact, error)
	ReplaceContactsForTypes(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, contactTypes []ContactType, entries []ChildContact) error
}

type ChildHealthProfileRepository interface {
	GetHealthByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildHealthProfile, error)
	UpsertHealth(ctx context.Context, tx pgx.Tx, p *ChildHealthProfile) (*ChildHealthProfile, error)
}

type ChildSafeguardingProfileRepository interface {
	GetSafeguardingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildSafeguardingProfile, error)
	UpsertSafeguarding(ctx context.Context, tx pgx.Tx, p *ChildSafeguardingProfile) (*ChildSafeguardingProfile, error)
}

type ChildConsentRepository interface {
	GetConsentByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildConsent, bool, error)
	InsertConsent(ctx context.Context, tx pgx.Tx, p *ChildConsent) (*ChildConsent, error)
	UpdateConsent(ctx context.Context, tx pgx.Tx, p *ChildConsent) (*ChildConsent, error)
}

type ChildFundingRepository interface {
	GetFundingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildFundingRecord, bool, error)
	UpsertFunding(ctx context.Context, tx pgx.Tx, p *ChildFundingRecord) (*ChildFundingRecord, error)
}

type ChildCollectionSettingsRepository interface {
	GetCollectionSettingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildCollectionSetting, error)
	UpsertCollectionSetting(ctx context.Context, tx pgx.Tx, p *ChildCollectionSetting) (*ChildCollectionSetting, error)
	SetCollectionPassword(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, hash string, updatedAt time.Time, userID, membershipID uuid.UUID) error
}

type ChildRoomAssignmentsRepository interface {
	ListRoomAssignmentsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]ChildRoomAssignment, error)
	GetCurrentRoomAssignmentByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildRoomAssignment, bool, error)
	InsertRoomAssignment(ctx context.Context, tx pgx.Tx, a *ChildRoomAssignment) (*ChildRoomAssignment, error)
	CloseCurrentRoomAssignment(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, endDate time.Time) error
	GetRoomAssignmentByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID) (*ChildRoomAssignment, bool, error)
	CloseRoomAssignmentByID(ctx context.Context, tx pgx.Tx, tenantID, branchID, id uuid.UUID, endDate time.Time) (bool, error)
}

type ChildBillingProfileRepository interface {
	GetBillingProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildBillingProfile, bool, error)
	UpsertBillingProfile(ctx context.Context, tx pgx.Tx, p *ChildBillingProfile) (*ChildBillingProfile, error)
}

type ChildLeavingRepository interface {
	GetLeavingRecordByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildLeavingRecord, bool, error)
	InsertLeavingRecord(ctx context.Context, tx pgx.Tx, p *ChildLeavingRecord) error
}

// Repository composes the per-concept repositories. The postgres implementation
// file implements all of them on one struct.
type Repository interface {
	ChildIdentityRepository
	ChildProfileRepository
	ChildContactRepository
	ChildHealthProfileRepository
	ChildSafeguardingProfileRepository
	ChildConsentRepository
	ChildFundingRepository
	ChildCollectionSettingsRepository
	ChildRoomAssignmentsRepository
	ChildBillingProfileRepository
	ChildLeavingRepository
}
