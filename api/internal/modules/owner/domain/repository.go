package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SummaryRepository interface {
	GetActiveSites(ctx context.Context, tenantID uuid.UUID) ([]Site, error)
	GetActiveSite(ctx context.Context, tenantID, siteID uuid.UUID) (Site, error)
	UpdateSiteCoreHourlyRate(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, siteID uuid.UUID, coreHourlyRateMinor int) (previous *int, current int, err error)

	CountActiveManagers(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID) (map[uuid.UUID]int, error)
	CountPendingManagerInvites(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID) (map[uuid.UUID]int, error)
	CountActiveChildren(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID) (map[uuid.UUID]int, error)
	CountAttendanceToday(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID, localDate time.Time) (map[uuid.UUID]int, error)
	CountIncompleteAttendance(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID, periodStart, periodEnd time.Time) (map[uuid.UUID]int, error)
	GetFundingReadiness(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID, billingMonth time.Time) (map[uuid.UUID]FundingReadiness, error)
	GetInvoicePaymentHealth(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID, billingMonth time.Time) (map[uuid.UUID]InvoicePaymentHealth, error)
}

type ManagerAccessEntry struct {
	MembershipID uuid.UUID
	UserID       uuid.UUID
	Email        string
	IsActive     bool
	EndedAt      *time.Time
}

type ManagerMembership struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	BranchID uuid.UUID
	UserID   uuid.UUID
	IsActive bool
	EndedAt  *time.Time
}

type PendingManagerInvite struct {
	ID              uuid.UUID
	Email           string
	EmailNormalized string
	ExpiresAt       time.Time
	SendCount       int
	CreatedAt       time.Time
}

type ManagerAccessRepository interface {
	GetActiveSite(ctx context.Context, tenantID, siteID uuid.UUID) (Site, error)

	FindActiveUserByEmail(ctx context.Context, emailNormalized string) (*uuid.UUID, error)
	FindManagerMembership(ctx context.Context, tenantID, branchID, userID uuid.UUID) (*ManagerMembership, error)
	CreateManagerMembership(ctx context.Context, id, tenantID, branchID, userID uuid.UUID) error
	ReactivateManagerMembership(ctx context.Context, id, tenantID uuid.UUID) error
	DeactivateManagerMembership(ctx context.Context, id, tenantID uuid.UUID) (int64, error)
	RevokeRefreshTokensByMembership(ctx context.Context, membershipID uuid.UUID) error

	FindPendingManagerInvite(ctx context.Context, tenantID, branchID uuid.UUID, emailNormalized string) (*PendingManagerInvite, error)
	CreateManagerInvite(ctx context.Context, id, tenantID, branchID uuid.UUID, email, emailNormalized, tokenHash string, expiresAt time.Time, createdByUserID, createdByMembershipID uuid.UUID) error
	RefreshManagerInvite(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time, resentByUserID, resentByMembershipID uuid.UUID) error

	ListManagerAccess(ctx context.Context, tenantID, branchID uuid.UUID, statusFilter string) ([]ManagerAccessEntry, error)
}

type InviteTokenGenerator interface {
	Generate() (raw string, hash string, expiresAt time.Time, err error)
}

type ManagerInviteSender interface {
	SendManagerInvite(ctx context.Context, toEmail, acceptURL string) error
}
