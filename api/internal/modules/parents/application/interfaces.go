package application

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	emaildomain "nursery-management-system/api/internal/modules/email/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type UserCreator interface {
	CreateParentUser(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, email string) (uuid.UUID, error)
}

type EmailSender interface {
	SendParentPortalInvite(ctx context.Context, tenantID, branchID uuid.UUID, toEmail, acceptURL string) error
}

type ParentChildExistenceChecker interface {
	ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error)
}

var _ EmailSender = (*emailSenderAdapter)(nil)

type emailSenderAdapter struct {
	enqueuer emaildomain.EmailEnqueuer
}

func NewEmailSenderAdapter(enqueuer emaildomain.EmailEnqueuer) *emailSenderAdapter {
	return &emailSenderAdapter{enqueuer: enqueuer}
}

func (a *emailSenderAdapter) SendParentPortalInvite(ctx context.Context, tenantID, branchID uuid.UUID, toEmail, acceptURL string) error {
	payloadJSON, err := json.Marshal(map[string]string{
		"accept_url": acceptURL,
	})
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	_, err = a.enqueuer.Enqueue(ctx, tenantID, branchID, emaildomain.EnqueueParams{
		EventType:       "portal_invite",
		Recipient:       toEmail,
		Subject:         "You're invited to access the parent portal",
		TemplateName:    "portal_invite",
		TemplateVersion: 1,
		PayloadJSON:     payloadJSON,
		EntityID:        toEmail,
	})
	return err
}

var _ ParentChildExistenceChecker = (*childExistenceCheckerAdapter)(nil)

type childExistenceCheckerAdapter struct {
	checker interface {
		ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error)
	}
}

func wrapChildChecker(checker interface {
	ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error)
}) *childExistenceCheckerAdapter {
	return &childExistenceCheckerAdapter{checker: checker}
}

func (a *childExistenceCheckerAdapter) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error) {
	return a.checker.ExistsInScope(ctx, tx, tenantID, branchID, childID)
}

func parseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, domainerrors.Validation("Invalid request payload.", "id")
	}
	return id, nil
}
