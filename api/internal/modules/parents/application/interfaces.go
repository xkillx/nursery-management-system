package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/platform/email"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type UserCreator interface {
	CreateParentUser(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, email string) (uuid.UUID, error)
}

type EmailSender interface {
	SendParentPortalInvite(ctx context.Context, toEmail, acceptURL string) error
}

type ParentChildExistenceChecker interface {
	ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error)
}

var _ EmailSender = (*emailSenderAdapter)(nil)

type emailSenderAdapter struct {
	sender  email.Sender
	baseURL string
}

func NewEmailSenderAdapter(sender email.Sender, baseURL string) *emailSenderAdapter {
	return &emailSenderAdapter{sender: sender, baseURL: baseURL}
}

func (a *emailSenderAdapter) SendParentPortalInvite(ctx context.Context, toEmail, acceptURL string) error {
	msg := email.Message{
		To:      toEmail,
		Subject: "You're invited to access the parent portal",
		Text: "You have been invited to access the parent portal.\n\n" +
			"Click the link below to set up your account:\n" + acceptURL + "\n\n" +
			"This invitation expires in 7 days.",
	}
	return a.sender.Send(ctx, msg)
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
