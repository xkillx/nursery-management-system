package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/payments/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) queriesTx(tx pgx.Tx) *sqlc.Queries {
	return sqlc.New(tx)
}

func (r *Repository) GetParentInvoiceForCheckoutForUpdate(ctx context.Context, tx domain.Tx, tenantID, branchID, membershipID, invoiceID string) (domain.CheckoutInvoiceCandidate, bool, error) {
	row, err := r.queriesTx(tx).GetParentInvoiceForCheckoutForUpdate(ctx, sqlc.GetParentInvoiceForCheckoutForUpdateParams{
		TenantID: uuidToPgtype(mustParseUUID(tenantID)),
		BranchID: uuidToPgtype(mustParseUUID(branchID)),
		ID:       uuidToPgtype(mustParseUUID(membershipID)),
		ID_2:     uuidToPgtype(mustParseUUID(invoiceID)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.CheckoutInvoiceCandidate{}, false, nil
		}
		return domain.CheckoutInvoiceCandidate{}, false, err
	}
	return domain.CheckoutInvoiceCandidate{
		ID:              pgtypeUUIDToStr(row.ID),
		InvoiceKind:     row.InvoiceKind,
		InvoiceNumber:   pgtypeTextToStr(row.InvoiceNumber),
		Status:          row.Status,
		CurrencyCode:    row.CurrencyCode,
		TotalDueMinor:   int(row.TotalDueMinor),
		AmountPaidMinor: int(row.AmountPaidMinor),
		ChildID:         pgtypeUUIDToStr(row.ChildID),
	}, true, nil
}

func (r *Repository) CreatePaymentAttempt(ctx context.Context, tx domain.Tx, params domain.PaymentAttemptCreateParams) error {
	return r.queriesTx(tx).CreatePaymentAttempt(ctx, sqlc.CreatePaymentAttemptParams{
		ID:                      uuidToPgtype(mustParseUUID(params.ID)),
		TenantID:                uuidToPgtype(mustParseUUID(params.TenantID)),
		BranchID:                uuidToPgtype(mustParseUUID(params.BranchID)),
		InvoiceID:               uuidToPgtype(mustParseUUID(params.InvoiceID)),
		InitiatedByUserID:       uuidToPgtype(mustParseUUID(params.InitiatedByUserID)),
		InitiatedByMembershipID: uuidToPgtype(mustParseUUID(params.InitiatedByMembershipID)),
		RequestID:               strToPgtypeText(params.RequestID),
		Status:                  params.Status,
		AmountMinor:             int32(params.AmountMinor),
		CurrencyCode:            params.CurrencyCode,
	})
}

func (r *Repository) GetInvoicePaymentState(ctx context.Context, tenantID, branchID, invoiceID string) (domain.InvoicePaymentState, bool, error) {
	row, err := sqlc.New(r.pool).GetInvoicePaymentState(ctx, sqlc.GetInvoicePaymentStateParams{
		TenantID: uuidToPgtype(mustParseUUID(tenantID)),
		BranchID: uuidToPgtype(mustParseUUID(branchID)),
		ID:       uuidToPgtype(mustParseUUID(invoiceID)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.InvoicePaymentState{}, false, nil
		}
		return domain.InvoicePaymentState{}, false, err
	}
	return domain.InvoicePaymentState{
		InvoiceKind:     row.InvoiceKind,
		Status:          row.Status,
		CurrencyCode:    row.CurrencyCode,
		TotalDueMinor:   int(row.TotalDueMinor),
		AmountPaidMinor: int(row.AmountPaidMinor),
	}, true, nil
}

func (r *Repository) MarkPaymentAttemptCheckoutCreated(ctx context.Context, params domain.PaymentAttemptCheckoutCreatedParams) error {
	affected, err := sqlc.New(r.pool).MarkPaymentAttemptCheckoutCreated(ctx, sqlc.MarkPaymentAttemptCheckoutCreatedParams{
		TenantID:                uuidToPgtype(mustParseUUID(params.TenantID)),
		BranchID:                uuidToPgtype(mustParseUUID(params.BranchID)),
		ID:                      uuidToPgtype(mustParseUUID(params.AttemptID)),
		StripeCheckoutSessionID: strToPgtypeText(params.StripeCheckoutSessionID),
		StripeCheckoutUrl:       strToPgtypeText(params.StripeCheckoutURL),
		StripePaymentIntentID:   strToPgtypeText(params.StripePaymentIntentID),
		StripeExpiresAt:         timeToPgtypeTimestamptzPtr(params.StripeExpiresAt),
	})
	if err != nil {
		return fmt.Errorf("mark payment attempt checkout created: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("mark payment attempt checkout created: expected 1 row affected, got %d", affected)
	}
	return nil
}

func (r *Repository) MarkPaymentAttemptCheckoutCreationFailed(ctx context.Context, params domain.PaymentAttemptCheckoutCreationFailedParams) error {
	affected, err := sqlc.New(r.pool).MarkPaymentAttemptCheckoutCreationFailed(ctx, sqlc.MarkPaymentAttemptCheckoutCreationFailedParams{
		TenantID:             uuidToPgtype(mustParseUUID(params.TenantID)),
		BranchID:             uuidToPgtype(mustParseUUID(params.BranchID)),
		ID:                   uuidToPgtype(mustParseUUID(params.AttemptID)),
		FailureReason:        strToPgtypeText(params.FailureReason),
		ProviderErrorCode:    strToPgtypeText(params.ProviderErrorCode),
		ProviderErrorMessage: strToPgtypeText(params.ProviderErrorMessage),
	})
	if err != nil {
		return fmt.Errorf("mark payment attempt checkout creation failed: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("mark payment attempt checkout creation failed: expected 1 row affected, got %d", affected)
	}
	return nil
}

func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("invalid uuid %q: %v", s, err))
	}
	return id
}

func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
}

func pgtypeUUIDToStr(u pgtype.UUID) string {
	return uuid.UUID(u.Bytes).String()
}

func pgtypeTextToStr(t pgtype.Text) string {
	return t.String
}

func strToPgtypeText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func timeToPgtypeTimestamptzPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
