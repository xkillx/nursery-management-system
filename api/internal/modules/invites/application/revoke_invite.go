package application

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/invites/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type RevokeInviteUseCase struct {
	repo   domain.Repository
	logger *slog.Logger
}

func NewRevokeInviteUseCase(repo domain.Repository, logger *slog.Logger) *RevokeInviteUseCase {
	return &RevokeInviteUseCase{repo: repo, logger: logger}
}

type RevokeInviteResult struct {
	Invite domain.Invite
}

func (uc *RevokeInviteUseCase) Execute(ctx context.Context, actor tenant.ActorContext, inviteID uuid.UUID) (RevokeInviteResult, error) {
	invite, err := uc.repo.GetInviteForUpdate(ctx, actor.TenantID, actor.BranchID, inviteID)
	if err != nil {
		return RevokeInviteResult{}, domainerrors.NotFound("invite", "Invitation not found.")
	}

	if invite.AcceptedAt != nil {
		return RevokeInviteResult{}, domainerrors.Conflict("invite_already_accepted", "Invitation has already been accepted.")
	}

	if invite.RevokedAt != nil {
		return RevokeInviteResult{Invite: invite}, nil
	}

	revoked, err := uc.repo.RevokeInvite(ctx, actor, invite.ID)
	if err != nil {
		uc.logger.Error("revoke_invite_failed", "error", err)
		return RevokeInviteResult{}, domainerrors.Internal(err)
	}

	return RevokeInviteResult{Invite: revoked}, nil
}
