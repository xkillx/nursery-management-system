package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/invites/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ResendInviteUseCase struct {
	repo       domain.Repository
	tokens     TokenGenerator
	email      EmailSender
	webBaseURL string
	logger     *slog.Logger
}

func NewResendInviteUseCase(
	repo domain.Repository,
	tokens TokenGenerator,
	email EmailSender,
	webBaseURL string,
	logger *slog.Logger,
) *ResendInviteUseCase {
	return &ResendInviteUseCase{
		repo:       repo,
		tokens:     tokens,
		email:      email,
		webBaseURL: webBaseURL,
		logger:     logger,
	}
}

type ResendInviteResult struct {
	Invite domain.Invite
}

func (uc *ResendInviteUseCase) Execute(ctx context.Context, actor tenant.ActorContext, inviteID uuid.UUID) (ResendInviteResult, error) {
	invite, err := uc.repo.GetInviteForUpdate(ctx, actor.TenantID, actor.BranchID, inviteID)
	if err != nil {
		return ResendInviteResult{}, domainerrors.NotFound("invite", "Invitation not found.")
	}

	if !invite.IsLivePending() {
		return ResendInviteResult{}, domainerrors.Conflict("invite_not_pending", "Invitation is not pending.")
	}

	raw, hash, expiresAt, err := uc.tokens.Generate()
	if err != nil {
		return ResendInviteResult{}, domainerrors.Internal(err)
	}

	acceptURL := fmt.Sprintf("%s/invite-accept?token=%s", uc.webBaseURL, url.QueryEscape(raw))

	updated, err := uc.repo.ResendInvite(ctx, actor, invite.ID, hash, expiresAt, func() error {
		return uc.email.SendInvite(ctx, invite.Email, invite.Role, acceptURL)
	})
	if err != nil {
		uc.logger.Error("resend_invite_failed", "error", err)
		return ResendInviteResult{}, domainerrors.Internal(err)
	}

	return ResendInviteResult{Invite: updated}, nil
}
