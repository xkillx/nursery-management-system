package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Tx = any

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

// Repository defines all data access methods for the children module.
// The postgres implementation struct satisfies all methods.
type Repository interface {
	// Identity
	List(ctx context.Context, tenantID, branchID uuid.UUID, filter StatusFilter, limit, offset int, roomID *uuid.UUID) ([]Child, error)
	Count(ctx context.Context, tenantID, branchID uuid.UUID, filter StatusFilter, roomID *uuid.UUID) (int, error)
	GetByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (Child, bool, error)
	Create(ctx context.Context, tx Tx, child Child, notes string, tenantID, branchID uuid.UUID) error
	Update(ctx context.Context, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	UpdateWithTx(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID, fields map[string]any) (int64, error)
	MarkInactive(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) error
	GetByIDForUpdate(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (Child, bool, error)
	ExistsInScope(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (bool, error)
	ListAttendance(ctx context.Context, tenantID, branchID uuid.UUID, localDate time.Time) ([]AttendanceChild, error)
	GetChildForCorrection(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (ChildCorrectionInfo, bool, error)

	// Profile
	GetProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildProfile, error)
	GetProfileForUpdate(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (*ChildProfile, error)
	InsertProfile(ctx context.Context, tx Tx, p *ChildProfile) (*ChildProfile, error)
	UpdateProfile(ctx context.Context, tx Tx, p *ChildProfile) (*ChildProfile, error)

	// Contacts
	ListContactsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]ChildContact, error)
	ReplaceContactsForTypes(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, contactTypes []ContactType, entries []ChildContact) error

	// Health
	GetHealthByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildHealthProfile, error)
	UpsertHealth(ctx context.Context, tx Tx, p *ChildHealthProfile) (*ChildHealthProfile, error)

	// Safeguarding
	GetSafeguardingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildSafeguardingProfile, error)
	UpsertSafeguarding(ctx context.Context, tx Tx, p *ChildSafeguardingProfile) (*ChildSafeguardingProfile, error)

	// Consent
	GetConsentByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildConsent, bool, error)
	InsertConsent(ctx context.Context, tx Tx, p *ChildConsent) (*ChildConsent, error)
	UpdateConsent(ctx context.Context, tx Tx, p *ChildConsent) (*ChildConsent, error)

	// Funding
	GetFundingByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildFundingRecord, bool, error)
	UpsertFunding(ctx context.Context, tx Tx, p *ChildFundingRecord) (*ChildFundingRecord, error)

	// Collection Settings
	GetCollectionSettingByChild(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (*ChildCollectionSetting, error)
	UpsertCollectionSetting(ctx context.Context, tx Tx, p *ChildCollectionSetting) (*ChildCollectionSetting, error)
	SetCollectionPassword(ctx context.Context, tx Tx, tenantID, branchID, childID, id uuid.UUID, password string, passwordHint string, updatedAt time.Time, userID, membershipID uuid.UUID) error

	// Room Assignments
	ListRoomAssignmentsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]ChildRoomAssignment, error)
	ListRoomAssignmentsByChildPaginated(ctx context.Context, tenantID, branchID, childID uuid.UUID, limit, offset int) ([]ChildRoomAssignment, error)
	CountRoomAssignmentsByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (int, error)
	GetCurrentRoomAssignmentByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildRoomAssignment, bool, error)
	InsertRoomAssignment(ctx context.Context, tx Tx, a *ChildRoomAssignment) (*ChildRoomAssignment, error)
	CloseCurrentRoomAssignment(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, endDate time.Time) error
	GetRoomAssignmentByID(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID) (*ChildRoomAssignment, bool, error)
	CloseRoomAssignmentByID(ctx context.Context, tx Tx, tenantID, branchID, id uuid.UUID, endDate time.Time) (bool, error)

	// Billing Profile
	GetBillingProfileByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildBillingProfile, bool, error)
	UpsertBillingProfile(ctx context.Context, tx Tx, p *ChildBillingProfile) (*ChildBillingProfile, error)

	// Leaving Records
	GetLeavingRecordByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*ChildLeavingRecord, bool, error)
	InsertLeavingRecord(ctx context.Context, tx Tx, p *ChildLeavingRecord) error

	// Booking Patterns
	ListByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]BookingPattern, error)
	ListByChildPaginated(ctx context.Context, tenantID, branchID, childID uuid.UUID, limit, offset int) ([]BookingPattern, error)
	CountByChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (int, error)
	GetPatternByID(ctx context.Context, tenantID, branchID, id uuid.UUID) (*BookingPattern, bool, error)
	GetActiveForDate(ctx context.Context, tenantID, branchID, childID uuid.UUID, date time.Time) (*BookingPattern, bool, error)
	GetCurrentOpenByChild(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (*BookingPattern, bool, error)
	GetPreviousClosedByChild(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID) (*BookingPattern, bool, error)
	InsertPattern(ctx context.Context, tx Tx, p *BookingPattern, entries []BookingPatternEntry) (*BookingPattern, error)
	CloseCurrentPattern(ctx context.Context, tx Tx, tenantID, branchID, childID uuid.UUID, effectiveTo time.Time) error
	ClosePatternByID(ctx context.Context, tx Tx, tenantID, branchID, patternID uuid.UUID, effectiveTo time.Time) error
	ReplaceEntries(ctx context.Context, tx Tx, tenantID, branchID, patternID uuid.UUID, entries []BookingPatternEntry) error
	UpdateEffectiveFrom(ctx context.Context, tx Tx, tenantID, branchID, patternID uuid.UUID, effectiveFrom time.Time) error
	UpdateTermTimeOnly(ctx context.Context, tx Tx, tenantID, branchID, patternID uuid.UUID, termTimeOnly bool) error
}
