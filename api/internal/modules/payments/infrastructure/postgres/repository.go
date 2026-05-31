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
	pool        *pgxpool.Pool
	managerRepo domain.ManagerPaymentRepository
}

// ManagerRepo returns the ManagerPaymentRepository interface for manager diagnostics.
func (r *Repository) ManagerRepo() domain.ManagerPaymentRepository {
	if r.managerRepo == nil {
		r.managerRepo = &managerQueries{pool: r.pool}
	}
	return r.managerRepo
}

type managerQueries struct {
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

func int32ToPgtypeInt4Ptr(v int32) pgtype.Int4 {
	return pgtype.Int4{Int32: v, Valid: true}
}

func (r *Repository) InsertWebhookEvent(ctx context.Context, tx pgx.Tx, event domain.StripeWebhookEvent, requestID string, processingStatus, processingReason string) (string, bool, error) {
	id, err := r.queriesTx(tx).InsertWebhookEvent(ctx, sqlc.InsertWebhookEventParams{
		ID:                uuidToPgtype(mustParseUUID(event.ID)),
		StripeEventID:     event.StripeEventID,
		EventType:         event.EventType,
		Livemode:          event.Livemode,
		ApiVersion:        strToPgtypeText(event.APIVersion),
		ProviderCreatedAt: timeToPgtypeTimestamptzPtr(event.ProviderCreatedAt),
		ProcessingStatus:  processingStatus,
		ProcessingReason:  strToPgtypeText(processingReason),
		RequestID:         strToPgtypeText(requestID),
		RawPayload:        event.RawPayload,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return pgtypeUUIDToStr(id), true, nil
}

func (r *Repository) UpdateWebhookEventStatus(ctx context.Context, tx pgx.Tx, eventID string, status, reason, errorMsg string) error {
	_, err := r.queriesTx(tx).UpdateWebhookEventStatus(ctx, sqlc.UpdateWebhookEventStatusParams{
		ID:               uuidToPgtype(mustParseUUID(eventID)),
		ProcessingStatus: status,
		ProcessingReason: strToPgtypeText(reason),
		ErrorMessage:     strToPgtypeText(errorMsg),
	})
	return err
}

func (r *Repository) GetPaymentAttemptAndInvoiceForWebhook(ctx context.Context, tx pgx.Tx, tenantID, branchID, invoiceID, attemptID, sessionID string) (*domain.WebhookAttemptInvoice, error) {
	row, err := r.queriesTx(tx).GetPaymentAttemptAndInvoiceForWebhook(ctx, sqlc.GetPaymentAttemptAndInvoiceForWebhookParams{
		TenantID:                uuidToPgtype(mustParseUUID(tenantID)),
		BranchID:                uuidToPgtype(mustParseUUID(branchID)),
		InvoiceID:               uuidToPgtype(mustParseUUID(invoiceID)),
		ID:                      uuidToPgtype(mustParseUUID(attemptID)),
		StripeCheckoutSessionID: strToPgtypeText(sessionID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	result := &domain.WebhookAttemptInvoice{
		AttemptID:              pgtypeUUIDToStr(row.AttemptID),
		AttemptStatus:          row.AttemptStatus,
		AttemptAmountMinor:     row.AttemptAmountMinor,
		AttemptCurrencyCode:    row.AttemptCurrencyCode,
		AttemptSessionID:       pgtypeTextToStr(row.AttemptSessionID),
		InvoiceID:              pgtypeUUIDToStr(row.InvoiceID),
		InvoiceStatus:          row.InvoiceStatus,
		InvoiceTotalDueMinor:   row.InvoiceTotalDueMinor,
		InvoiceAmountPaidMinor: row.InvoiceAmountPaidMinor,
		InvoiceCurrencyCode:    row.InvoiceCurrencyCode,
	}
	if row.InvoicePaidAt.Valid {
		t := row.InvoicePaidAt.Time
		result.InvoicePaidAt = &t
	}
	if row.InvoicePaymentFailedAt.Valid {
		t := row.InvoicePaymentFailedAt.Time
		result.InvoicePaymentFailedAt = &t
	}
	return result, nil
}

func (r *Repository) MarkPaymentAttemptPaid(ctx context.Context, tx pgx.Tx, tenantID, branchID, attemptID string) error {
	affected, err := r.queriesTx(tx).MarkPaymentAttemptPaid(ctx, sqlc.MarkPaymentAttemptPaidParams{
		TenantID: uuidToPgtype(mustParseUUID(tenantID)),
		BranchID: uuidToPgtype(mustParseUUID(branchID)),
		ID:       uuidToPgtype(mustParseUUID(attemptID)),
	})
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("mark payment attempt paid: expected 1 row affected, got %d", affected)
	}
	return nil
}

func (r *Repository) MarkPaymentAttemptFailed(ctx context.Context, tx pgx.Tx, tenantID, branchID, attemptID string) error {
	affected, err := r.queriesTx(tx).MarkPaymentAttemptFailed(ctx, sqlc.MarkPaymentAttemptFailedParams{
		TenantID: uuidToPgtype(mustParseUUID(tenantID)),
		BranchID: uuidToPgtype(mustParseUUID(branchID)),
		ID:       uuidToPgtype(mustParseUUID(attemptID)),
	})
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("mark payment attempt failed: expected 1 row affected, got %d", affected)
	}
	return nil
}

func (r *Repository) MarkPaymentAttemptExpired(ctx context.Context, tx pgx.Tx, tenantID, branchID, attemptID string) error {
	affected, err := r.queriesTx(tx).MarkPaymentAttemptExpired(ctx, sqlc.MarkPaymentAttemptExpiredParams{
		TenantID: uuidToPgtype(mustParseUUID(tenantID)),
		BranchID: uuidToPgtype(mustParseUUID(branchID)),
		ID:       uuidToPgtype(mustParseUUID(attemptID)),
	})
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("mark payment attempt expired: expected 1 row affected, got %d", affected)
	}
	return nil
}

func (r *Repository) MarkInvoicePaid(ctx context.Context, tx pgx.Tx, tenantID, branchID, invoiceID string) error {
	affected, err := r.queriesTx(tx).MarkInvoicePaid(ctx, sqlc.MarkInvoicePaidParams{
		TenantID: uuidToPgtype(mustParseUUID(tenantID)),
		BranchID: uuidToPgtype(mustParseUUID(branchID)),
		ID:       uuidToPgtype(mustParseUUID(invoiceID)),
	})
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("mark invoice paid: expected 1 row affected, got %d", affected)
	}
	return nil
}

func (r *Repository) MarkInvoicePaymentFailed(ctx context.Context, tx pgx.Tx, tenantID, branchID, invoiceID string) error {
	affected, err := r.queriesTx(tx).MarkInvoicePaymentFailed(ctx, sqlc.MarkInvoicePaymentFailedParams{
		TenantID: uuidToPgtype(mustParseUUID(tenantID)),
		BranchID: uuidToPgtype(mustParseUUID(branchID)),
		ID:       uuidToPgtype(mustParseUUID(invoiceID)),
	})
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("mark invoice payment_failed: expected 1 row affected, got %d", affected)
	}
	return nil
}

func (r *Repository) InsertReconciliationRecord(ctx context.Context, tx pgx.Tx, params domain.ReconciliationRecordParams) error {
	return r.queriesTx(tx).InsertReconciliationRecord(ctx, sqlc.InsertReconciliationRecordParams{
		ID:                      uuidToPgtype(mustParseUUID(params.ID)),
		TenantID:                uuidToPgtype(mustParseUUID(params.TenantID)),
		BranchID:                uuidToPgtype(mustParseUUID(params.BranchID)),
		InvoiceID:               uuidToPgtype(mustParseUUID(params.InvoiceID)),
		PaymentAttemptID:        uuidToPgtype(mustParseUUID(params.PaymentAttemptID)),
		StripeWebhookEventID:    uuidToPgtype(mustParseUUID(params.WebhookEventID)),
		StripeEventID:           params.StripeEventID,
		StripeEventType:         params.StripeEventType,
		StripeCheckoutSessionID: params.CheckoutSessionID,
		StripePaymentIntentID:   strToPgtypeText(params.PaymentIntentID),
		Outcome:                 params.Outcome,
		ReasonCode:              params.ReasonCode,
		PreviousInvoiceStatus:   strToPgtypeText(params.PreviousInvoiceStatus),
		NewInvoiceStatus:        strToPgtypeText(params.NewInvoiceStatus),
		AttemptPreviousStatus:   strToPgtypeText(params.AttemptPreviousStatus),
		AttemptNewStatus:        strToPgtypeText(params.AttemptNewStatus),
		AmountMinor:             int32ToPgtypeInt4Ptr(params.AmountMinor),
		CurrencyCode:            strToPgtypeText(params.CurrencyCode),
		Details:                 []byte(params.Details),
	})
}

func (m *managerQueries) GetManagerInvoicePaymentStatus(ctx context.Context, tenantID, branchID, invoiceID string) (domain.ManagerInvoicePaymentStatus, bool, error) {
	row, err := sqlc.New(m.pool).GetManagerInvoicePaymentStatus(ctx, sqlc.GetManagerInvoicePaymentStatusParams{
		TenantID: uuidToPgtype(mustParseUUID(tenantID)),
		BranchID: uuidToPgtype(mustParseUUID(branchID)),
		ID:       uuidToPgtype(mustParseUUID(invoiceID)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ManagerInvoicePaymentStatus{}, false, nil
		}
		return domain.ManagerInvoicePaymentStatus{}, false, err
	}

	result := domain.ManagerInvoicePaymentStatus{
		InvoiceID:            pgtypeUUIDToStr(row.InvoiceID),
		InvoiceKind:          row.InvoiceKind,
		InvoiceNumber:        pgtypeTextToStr(row.InvoiceNumber),
		InvoiceNumberDisplay: row.InvoiceNumberDisplay,
		ChildID:              pgtypeUUIDToStr(row.ChildID),
		ChildName:            row.ChildName,
		BillingMonth:         pgtypeDateToStr(row.BillingMonth),
		Status:               row.Status,
		CurrencyCode:         row.CurrencyCode,
		TotalDueMinor:        int(row.TotalDueMinor),
		AmountPaidMinor:      int(row.AmountPaidMinor),
		CreatedAt:            row.CreatedAt.Time,
		UpdatedAt:            row.UpdatedAt.Time,
	}
	if row.IssuedAt.Valid {
		result.IssuedAt = &row.IssuedAt.Time
	}
	if row.DueAt.Valid {
		result.DueAt = &row.DueAt.Time
	}
	if row.PaidAt.Valid {
		result.PaidAt = &row.PaidAt.Time
	}
	if row.PaymentFailedAt.Valid {
		result.PaymentFailedAt = &row.PaymentFailedAt.Time
	}
	if row.PaymentStatusUpdatedAt.Valid {
		result.PaymentStatusUpdatedAt = &row.PaymentStatusUpdatedAt.Time
	}
	return result, true, nil
}

func (m *managerQueries) GetLatestPaymentAttemptForInvoice(ctx context.Context, tenantID, branchID, invoiceID string) (*domain.PaymentAttemptDiagnostic, error) {
	row, err := sqlc.New(m.pool).GetLatestPaymentAttemptForInvoice(ctx, sqlc.GetLatestPaymentAttemptForInvoiceParams{
		TenantID:  uuidToPgtype(mustParseUUID(tenantID)),
		BranchID:  uuidToPgtype(mustParseUUID(branchID)),
		InvoiceID: uuidToPgtype(mustParseUUID(invoiceID)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := &domain.PaymentAttemptDiagnostic{
		PaymentAttemptID: pgtypeUUIDToStr(row.PaymentAttemptID),
		Status:           row.Status,
		AmountMinor:      int(row.AmountMinor),
		CurrencyCode:     row.CurrencyCode,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
	if row.StripeCheckoutSessionID.Valid {
		result.StripeCheckoutSessionID = &row.StripeCheckoutSessionID.String
	}
	if row.StripePaymentIntentID.Valid {
		result.StripePaymentIntentID = &row.StripePaymentIntentID.String
	}
	if row.StripeExpiresAt.Valid {
		result.StripeExpiresAt = &row.StripeExpiresAt.Time
	}
	if row.FailureReason.Valid {
		result.FailureReason = &row.FailureReason.String
	}
	if row.ProviderErrorCode.Valid {
		result.ProviderErrorCode = &row.ProviderErrorCode.String
	}
	if row.ProviderErrorMessage.Valid {
		result.ProviderErrorMessage = &row.ProviderErrorMessage.String
	}
	return result, nil
}

func (m *managerQueries) GetLatestPaymentEventForInvoice(ctx context.Context, tenantID, branchID, invoiceID string) (*domain.PaymentEventDiagnostic, error) {
	row, err := sqlc.New(m.pool).GetLatestPaymentEventForInvoice(ctx, sqlc.GetLatestPaymentEventForInvoiceParams{
		TenantID:  uuidToPgtype(mustParseUUID(tenantID)),
		BranchID:  uuidToPgtype(mustParseUUID(branchID)),
		InvoiceID: uuidToPgtype(mustParseUUID(invoiceID)),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := &domain.PaymentEventDiagnostic{
		PaymentEventID:          pgtypeUUIDToStr(row.PaymentEventID),
		PaymentAttemptID:        pgtypeUUIDToStr(row.PaymentAttemptID),
		StripeEventID:           row.StripeEventID,
		StripeEventType:         row.StripeEventType,
		StripeCheckoutSessionID: row.StripeCheckoutSessionID,
		Outcome:                 row.Outcome,
		ReasonCode:              row.ReasonCode,
		WebhookProcessingStatus: row.WebhookProcessingStatus,
		CreatedAt:               row.CreatedAt.Time,
	}
	if row.StripePaymentIntentID.Valid {
		result.StripePaymentIntentID = row.StripePaymentIntentID.String
	}
	if row.PreviousInvoiceStatus.Valid {
		result.PreviousInvoiceStatus = row.PreviousInvoiceStatus.String
	}
	if row.NewInvoiceStatus.Valid {
		result.NewInvoiceStatus = row.NewInvoiceStatus.String
	}
	if row.AttemptPreviousStatus.Valid {
		result.AttemptPreviousStatus = row.AttemptPreviousStatus.String
	}
	if row.AttemptNewStatus.Valid {
		result.AttemptNewStatus = row.AttemptNewStatus.String
	}
	if row.AmountMinor.Valid {
		result.AmountMinor = int(row.AmountMinor.Int32)
	}
	if row.CurrencyCode.Valid {
		result.CurrencyCode = row.CurrencyCode.String
	}
	if row.WebhookProcessingReason.Valid {
		result.WebhookProcessingReason = row.WebhookProcessingReason.String
	}
	if row.WebhookReceivedAt.Valid {
		result.WebhookReceivedAt = &row.WebhookReceivedAt.Time
	}
	if row.WebhookProcessedAt.Valid {
		result.WebhookProcessedAt = &row.WebhookProcessedAt.Time
	}
	return result, nil
}

func (m *managerQueries) ListPaymentEventsForInvoice(ctx context.Context, tenantID, branchID, invoiceID string, filters domain.PaymentEventFilters) ([]domain.PaymentEventDiagnostic, error) {
	rows, err := sqlc.New(m.pool).ListPaymentEventsForInvoice(ctx, sqlc.ListPaymentEventsForInvoiceParams{
		TenantID:  uuidToPgtype(mustParseUUID(tenantID)),
		BranchID:  uuidToPgtype(mustParseUUID(branchID)),
		InvoiceID: uuidToPgtype(mustParseUUID(invoiceID)),
		Limit:     int32(filters.Limit),
		Offset:    int32(filters.Offset),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.PaymentEventDiagnostic, 0, len(rows))
	for _, row := range rows {
		event := domain.PaymentEventDiagnostic{
			PaymentEventID:          pgtypeUUIDToStr(row.PaymentEventID),
			PaymentAttemptID:        pgtypeUUIDToStr(row.PaymentAttemptID),
			StripeEventID:           row.StripeEventID,
			StripeEventType:         row.StripeEventType,
			StripeCheckoutSessionID: row.StripeCheckoutSessionID,
			Outcome:                 row.Outcome,
			ReasonCode:              row.ReasonCode,
			WebhookProcessingStatus: row.WebhookProcessingStatus,
			CreatedAt:               row.CreatedAt.Time,
		}
		if row.StripePaymentIntentID.Valid {
			event.StripePaymentIntentID = row.StripePaymentIntentID.String
		}
		if row.PreviousInvoiceStatus.Valid {
			event.PreviousInvoiceStatus = row.PreviousInvoiceStatus.String
		}
		if row.NewInvoiceStatus.Valid {
			event.NewInvoiceStatus = row.NewInvoiceStatus.String
		}
		if row.AttemptPreviousStatus.Valid {
			event.AttemptPreviousStatus = row.AttemptPreviousStatus.String
		}
		if row.AttemptNewStatus.Valid {
			event.AttemptNewStatus = row.AttemptNewStatus.String
		}
		if row.AmountMinor.Valid {
			event.AmountMinor = int(row.AmountMinor.Int32)
		}
		if row.CurrencyCode.Valid {
			event.CurrencyCode = row.CurrencyCode.String
		}
		if row.WebhookProcessingReason.Valid {
			event.WebhookProcessingReason = row.WebhookProcessingReason.String
		}
		if row.WebhookReceivedAt.Valid {
			event.WebhookReceivedAt = &row.WebhookReceivedAt.Time
		}
		if row.WebhookProcessedAt.Valid {
			event.WebhookProcessedAt = &row.WebhookProcessedAt.Time
		}
		result = append(result, event)
	}
	return result, nil
}

func pgtypeDateToStr(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}
