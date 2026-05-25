package application

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/modules/invites/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/uid"
)

type AcceptInviteUseCase struct {
	repo   domain.Repository
	logger *slog.Logger
}

func NewAcceptInviteUseCase(repo domain.Repository, logger *slog.Logger) *AcceptInviteUseCase {
	return &AcceptInviteUseCase{repo: repo, logger: logger}
}

type AcceptInviteResult struct{}

func (uc *AcceptInviteUseCase) Execute(ctx context.Context, tokenHash, newPassword string) (AcceptInviteResult, error) {
	if len(newPassword) < 8 {
		return AcceptInviteResult{}, domainerrors.Validation("Password must be at least 8 characters.", "new_password")
	}

	invite, err := uc.repo.GetInviteByTokenHashForUpdate(ctx, tokenHash)
	if err != nil {
		return AcceptInviteResult{}, domainerrors.New("invite_token_invalid", "Invalid invitation token.")
	}

	if invite.AcceptedAt != nil {
		return AcceptInviteResult{}, domainerrors.New("invite_token_accepted", "Invitation has already been accepted.")
	}
	if invite.RevokedAt != nil {
		return AcceptInviteResult{}, domainerrors.New("invite_token_revoked", "Invitation has been revoked.")
	}
	if invite.ExpiresAt.Before(time.Now().UTC()) {
		return AcceptInviteResult{}, domainerrors.New("invite_token_expired", "Invitation has expired.")
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return AcceptInviteResult{}, domainerrors.Internal(err)
	}

	userID := uid.NewUUID()
	membershipID := uid.NewUUID()

	user := domain.CreatedUser{
		ID:           userID,
		Email:        invite.Email,
		PasswordHash: string(pwHash),
	}

	membership := domain.CreatedMembership{
		ID:       membershipID,
		TenantID: invite.TenantID,
		BranchID: invite.BranchID,
		UserID:   userID,
		Role:     invite.Role,
	}

	if err := uc.repo.AcceptInvite(ctx, invite, user, membership); err != nil {
		uc.logger.Error("accept_invite_failed", "error", err)
		return AcceptInviteResult{}, domainerrors.Internal(err)
	}

	return AcceptInviteResult{}, nil
}
