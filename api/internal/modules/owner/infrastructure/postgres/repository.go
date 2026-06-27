package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/owner/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type OwnerRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *OwnerRepository {
	return &OwnerRepository{pool: pool}
}

func (r *OwnerRepository) GetActiveSites(ctx context.Context, tenantID uuid.UUID) ([]domain.Site, error) {
	q := sqlc.New(r.pool)
	rows, err := q.OwnerGetActiveSites(ctx, uuidToPgtype(tenantID))
	if err != nil {
		return nil, err
	}
	sites := make([]domain.Site, 0, len(rows))
	for _, row := range rows {
		sites = append(sites, domain.Site{
			ID:                  pgtypeUUIDToUUID(row.ID),
			Name:                row.Name,
			CoreHourlyRateMinor: pgtypeInt4ToPtr(row.CoreHourlyRateMinor),
		})
	}
	return sites, nil
}

func (r *OwnerRepository) GetActiveSite(ctx context.Context, tenantID, siteID uuid.UUID) (domain.Site, error) {
	q := sqlc.New(r.pool)
	row, err := q.OwnerGetActiveSite(ctx, sqlc.OwnerGetActiveSiteParams{
		TenantID: uuidToPgtype(tenantID),
		ID:       uuidToPgtype(siteID),
	})
	if err != nil {
		if isNoRows(err) {
			return domain.Site{}, domain.ErrSiteNotFound
		}
		return domain.Site{}, err
	}
	return domain.Site{ID: pgtypeUUIDToUUID(row.ID), Name: row.Name, CoreHourlyRateMinor: pgtypeInt4ToPtr(row.CoreHourlyRateMinor)}, nil
}

func (r *OwnerRepository) CountActiveManagers(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(branchIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	q := sqlc.New(r.pool)
	rows, err := q.OwnerCountActiveManagersByBranches(ctx, sqlc.OwnerCountActiveManagersByBranchesParams{
		TenantID: uuidToPgtype(tenantID),
		Column2:  uuidsToPgtypeArray(branchIDs),
	})
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		result[pgtypeUUIDToUUID(row.BranchID)] = int(row.Count)
	}
	return result, nil
}

func (r *OwnerRepository) CountPendingManagerInvites(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(branchIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	q := sqlc.New(r.pool)
	rows, err := q.OwnerCountPendingManagerInvitesByBranches(ctx, sqlc.OwnerCountPendingManagerInvitesByBranchesParams{
		TenantID: uuidToPgtype(tenantID),
		Column2:  uuidsToPgtypeArray(branchIDs),
	})
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		result[pgtypeUUIDToUUID(row.BranchID)] = int(row.Count)
	}
	return result, nil
}

func (r *OwnerRepository) CountActiveChildren(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(branchIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	q := sqlc.New(r.pool)
	rows, err := q.OwnerCountActiveChildrenByBranches(ctx, sqlc.OwnerCountActiveChildrenByBranchesParams{
		TenantID: uuidToPgtype(tenantID),
		Column2:  uuidsToPgtypeArray(branchIDs),
	})
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		result[pgtypeUUIDToUUID(row.BranchID)] = int(row.Count)
	}
	return result, nil
}

func (r *OwnerRepository) CountAttendanceToday(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID, localDate time.Time) (map[uuid.UUID]int, error) {
	if len(branchIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	q := sqlc.New(r.pool)
	rows, err := q.OwnerCountAttendanceTodayByBranches(ctx, sqlc.OwnerCountAttendanceTodayByBranchesParams{
		TenantID:         uuidToPgtype(tenantID),
		Column2:          uuidsToPgtypeArray(branchIDs),
		CheckInLocalDate: timeToPgtypeDate(localDate),
	})
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		result[pgtypeUUIDToUUID(row.BranchID)] = int(row.CheckedInCount)
	}
	return result, nil
}

func (r *OwnerRepository) CountIncompleteAttendance(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID, periodStart, periodEnd time.Time) (map[uuid.UUID]int, error) {
	if len(branchIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}
	q := sqlc.New(r.pool)
	rows, err := q.OwnerCountIncompleteAttendanceByBranches(ctx, sqlc.OwnerCountIncompleteAttendanceByBranchesParams{
		TenantID:           uuidToPgtype(tenantID),
		Column2:            uuidsToPgtypeArray(branchIDs),
		CheckInLocalDate:   timeToPgtypeDate(periodStart),
		CheckInLocalDate_2: timeToPgtypeDate(periodEnd),
	})
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		result[pgtypeUUIDToUUID(row.BranchID)] = int(row.Count)
	}
	return result, nil
}

func (r *OwnerRepository) GetFundingReadiness(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID, billingMonth time.Time) (map[uuid.UUID]domain.FundingReadiness, error) {
	if len(branchIDs) == 0 {
		return map[uuid.UUID]domain.FundingReadiness{}, nil
	}
	q := sqlc.New(r.pool)
	rows, err := q.OwnerGetFundingReadinessByBranches(ctx, sqlc.OwnerGetFundingReadinessByBranchesParams{
		TenantID:     uuidToPgtype(tenantID),
		Column2:      uuidsToPgtypeArray(branchIDs),
		BillingMonth: timeToPgtypeDate(billingMonth),
	})
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]domain.FundingReadiness, len(rows))
	for _, row := range rows {
		result[pgtypeUUIDToUUID(row.BranchID)] = domain.FundingReadiness{
			IncludedChildCount:  int(row.IncludedChildCount),
			MissingProfileCount: int(row.MissingProfileCount),
			ExplicitZeroCount:   int(row.ExplicitZeroCount),
			UnderOneHourCount:   int(row.UnderOneHourCount),
			Above160HoursCount:  int(row.Above160HoursCount),
		}
	}
	return result, nil
}

func (r *OwnerRepository) GetInvoicePaymentHealth(ctx context.Context, tenantID uuid.UUID, branchIDs []uuid.UUID, billingMonth time.Time) (map[uuid.UUID]domain.InvoicePaymentHealth, error) {
	if len(branchIDs) == 0 {
		return map[uuid.UUID]domain.InvoicePaymentHealth{}, nil
	}
	q := sqlc.New(r.pool)
	rows, err := q.OwnerGetInvoicePaymentHealthByBranches(ctx, sqlc.OwnerGetInvoicePaymentHealthByBranchesParams{
		TenantID:     uuidToPgtype(tenantID),
		Column2:      uuidsToPgtypeArray(branchIDs),
		BillingMonth: timeToPgtypeDate(billingMonth),
	})
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]domain.InvoicePaymentHealth, len(rows))
	for _, row := range rows {
		result[pgtypeUUIDToUUID(row.BranchID)] = domain.InvoicePaymentHealth{
			CurrencyCode:            row.CurrencyCode,
			DraftCount:              int(row.DraftCount),
			IssuedCount:             int(row.IssuedCount),
			OverdueCount:            int(row.OverdueCount),
			PaymentFailedCount:      int(row.PaymentFailedCount),
			PaidCount:               int(row.PaidCount),
			TotalIssuedMinor:        row.TotalIssuedMinor,
			TotalPaidMinor:          row.TotalPaidMinor,
			OutstandingMinor:        row.OutstandingMinor,
			OverdueOutstandingMinor: row.OverdueOutstandingMinor,
		}
	}
	return result, nil
}

func isNoRows(err error) bool {
	return err == pgx.ErrNoRows
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return uuid.UUID(u.Bytes)
}

func uuidsToPgtypeArray(ids []uuid.UUID) []pgtype.UUID {
	out := make([]pgtype.UUID, 0, len(ids))
	for _, id := range ids {
		out = append(out, uuidToPgtype(id))
	}
	return out
}

func timeToPgtypeDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

func timeToPgtypeTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func pgtypeInt4ToPtr(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int32)
	return &i
}

// ── ManagerAccessRepository ──────────────────────────────────────────────────

func (r *OwnerRepository) UpdateSiteCoreHourlyRate(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, siteID uuid.UUID, coreHourlyRateMinor int) (previous *int, current int, err error) {
	q := sqlc.New(tx)
	row, err := q.OwnerGetActiveSite(ctx, sqlc.OwnerGetActiveSiteParams{
		TenantID: uuidToPgtype(tenantID),
		ID:       uuidToPgtype(siteID),
	})
	if err != nil {
		if isNoRows(err) {
			return nil, 0, domain.ErrSiteNotFound
		}
		return nil, 0, err
	}
	prev := pgtypeInt4ToPtr(row.CoreHourlyRateMinor)

	const updateQ = `UPDATE branches SET core_hourly_rate_minor = $1 WHERE tenant_id = $2 AND id = $3 AND is_active = true`
	tag, err := tx.Exec(ctx, updateQ, coreHourlyRateMinor, tenantID, siteID)
	if err != nil {
		return prev, 0, err
	}
	if tag.RowsAffected() == 0 {
		return prev, 0, domain.ErrSiteNotFound
	}
	return prev, coreHourlyRateMinor, nil
}

// ── ManagerAccessRepository ──────────────────────────────────────────────────

func (r *OwnerRepository) FindActiveUserByEmail(ctx context.Context, emailNormalized string) (*uuid.UUID, error) {
	q := sqlc.New(r.pool)
	row, err := q.OwnerFindActiveUserByEmail(ctx, emailNormalized)
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	id := pgtypeUUIDToUUID(row.ID)
	return &id, nil
}

func (r *OwnerRepository) FindManagerMembership(ctx context.Context, tenantID, branchID, userID uuid.UUID) (*domain.ManagerMembership, error) {
	q := sqlc.New(r.pool)
	row, err := q.OwnerFindManagerMembershipForUser(ctx, sqlc.OwnerFindManagerMembershipForUserParams{
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		UserID:   uuidToPgtype(userID),
	})
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	m := &domain.ManagerMembership{
		ID:       pgtypeUUIDToUUID(row.ID),
		TenantID: pgtypeUUIDToUUID(row.TenantID),
		BranchID: pgtypeUUIDToUUID(row.BranchID),
		UserID:   pgtypeUUIDToUUID(row.UserID),
		IsActive: row.IsActive,
	}
	if row.EndedAt.Valid {
		t := row.EndedAt.Time
		m.EndedAt = &t
	}
	return m, nil
}

func (r *OwnerRepository) CreateManagerMembership(ctx context.Context, id, tenantID, branchID, userID uuid.UUID) error {
	q := sqlc.New(r.pool)
	return q.OwnerCreateManagerMembership(ctx, sqlc.OwnerCreateManagerMembershipParams{
		ID:       uuidToPgtype(id),
		TenantID: uuidToPgtype(tenantID),
		BranchID: uuidToPgtype(branchID),
		UserID:   uuidToPgtype(userID),
	})
}

func (r *OwnerRepository) ReactivateManagerMembership(ctx context.Context, id, tenantID uuid.UUID) error {
	q := sqlc.New(r.pool)
	return q.OwnerReactivateManagerMembership(ctx, sqlc.OwnerReactivateManagerMembershipParams{
		ID:       uuidToPgtype(id),
		TenantID: uuidToPgtype(tenantID),
	})
}

func (r *OwnerRepository) DeactivateManagerMembership(ctx context.Context, id, tenantID uuid.UUID) (int64, error) {
	q := sqlc.New(r.pool)
	return q.OwnerDeactivateManagerMembership(ctx, sqlc.OwnerDeactivateManagerMembershipParams{
		ID:       uuidToPgtype(id),
		TenantID: uuidToPgtype(tenantID),
	})
}

func (r *OwnerRepository) RevokeRefreshTokensByMembership(ctx context.Context, membershipID uuid.UUID) error {
	q := sqlc.New(r.pool)
	return q.OwnerRevokeRefreshTokensByMembership(ctx, uuidToPgtype(membershipID))
}

func (r *OwnerRepository) FindPendingManagerInvite(ctx context.Context, tenantID, branchID uuid.UUID, emailNormalized string) (*domain.PendingManagerInvite, error) {
	q := sqlc.New(r.pool)
	row, err := q.OwnerFindPendingManagerInvite(ctx, sqlc.OwnerFindPendingManagerInviteParams{
		TenantID:        uuidToPgtype(tenantID),
		BranchID:        uuidToPgtype(branchID),
		EmailNormalized: emailNormalized,
	})
	if err != nil {
		if isNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.PendingManagerInvite{
		ID:              pgtypeUUIDToUUID(row.ID),
		Email:           row.Email,
		EmailNormalized: row.EmailNormalized,
		ExpiresAt:       row.ExpiresAt.Time,
		SendCount:       int(row.SendCount),
		CreatedAt:       row.CreatedAt.Time,
	}, nil
}

func (r *OwnerRepository) CreateManagerInvite(ctx context.Context, id, tenantID, branchID uuid.UUID, email, emailNormalized, tokenHash string, expiresAt time.Time, createdByUserID, createdByMembershipID uuid.UUID) error {
	q := sqlc.New(r.pool)
	return q.OwnerCreateManagerInvite(ctx, sqlc.OwnerCreateManagerInviteParams{
		ID:                    uuidToPgtype(id),
		TenantID:              uuidToPgtype(tenantID),
		BranchID:              uuidToPgtype(branchID),
		Email:                 email,
		EmailNormalized:       emailNormalized,
		TokenHash:             tokenHash,
		ExpiresAt:             timeToPgtypeTimestamptz(expiresAt),
		CreatedByUserID:       uuidToPgtype(createdByUserID),
		CreatedByMembershipID: uuidToPgtype(createdByMembershipID),
	})
}

func (r *OwnerRepository) RefreshManagerInvite(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time, resentByUserID, resentByMembershipID uuid.UUID) error {
	q := sqlc.New(r.pool)
	return q.OwnerRefreshManagerInvite(ctx, sqlc.OwnerRefreshManagerInviteParams{
		ID:                   uuidToPgtype(id),
		TokenHash:            tokenHash,
		ExpiresAt:            timeToPgtypeTimestamptz(expiresAt),
		ResentByUserID:       uuidToPgtype(resentByUserID),
		ResentByMembershipID: uuidToPgtype(resentByMembershipID),
	})
}

func (r *OwnerRepository) ListManagerAccess(ctx context.Context, tenantID, branchID uuid.UUID, statusFilter string) ([]domain.ManagerAccessEntry, error) {
	q := sqlc.New(r.pool)
	rows, err := q.OwnerListManagerAccess(ctx, sqlc.OwnerListManagerAccessParams{
		TenantID:     uuidToPgtype(tenantID),
		BranchID:     uuidToPgtype(branchID),
		StatusFilter: statusFilter,
	})
	if err != nil {
		return nil, err
	}
	entries := make([]domain.ManagerAccessEntry, 0, len(rows))
	for _, row := range rows {
		e := domain.ManagerAccessEntry{
			MembershipID: pgtypeUUIDToUUID(row.MembershipID),
			UserID:       pgtypeUUIDToUUID(row.UserID),
			Email:        row.Email,
			IsActive:     row.IsActive,
		}
		if row.EndedAt.Valid {
			t := row.EndedAt.Time
			e.EndedAt = &t
		}
		entries = append(entries, e)
	}
	return entries, nil
}
