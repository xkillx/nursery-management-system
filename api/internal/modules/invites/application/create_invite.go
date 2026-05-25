package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"nursery-management-system/api/internal/modules/invites/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type TokenGenerator interface {
	Generate() (raw string, hash string, expiresAt time.Time, err error)
}

type EmailSender interface {
	SendInvite(ctx context.Context, toEmail, role, acceptURL string) error
}

type CreateInviteUseCase struct {
	repo       domain.Repository
	tokens     TokenGenerator
	email      EmailSender
	webBaseURL string
	logger     *slog.Logger
}

func NewCreateInviteUseCase(
	repo domain.Repository,
	tokens TokenGenerator,
	email EmailSender,
	webBaseURL string,
	logger *slog.Logger,
) *CreateInviteUseCase {
	return &CreateInviteUseCase{
		repo:       repo,
		tokens:     tokens,
		email:      email,
		webBaseURL: webBaseURL,
		logger:     logger,
	}
}

type CreateInviteResult struct {
	Invite domain.Invite
	IsNew  bool
}

func (uc *CreateInviteUseCase) Execute(ctx context.Context, actor tenant.ActorContext, emailAddr, role string) (CreateInviteResult, error) {
	if role != "practitioner" && role != "parent" {
		return CreateInviteResult{}, domainerrors.New("invite_role_not_allowed", "Manager role is not allowed for invitations.", "role")
	}

	emailNormalized := strings.ToLower(strings.TrimSpace(emailAddr))

	raw, hash, expiresAt, err := uc.tokens.Generate()
	if err != nil {
		return CreateInviteResult{}, domainerrors.Internal(err)
	}

	acceptURL := fmt.Sprintf("%s/invite-accept?token=%s", uc.webBaseURL, url.QueryEscape(raw))

	inv := domain.Invite{
		ID:                    uid.NewUUID(),
		TenantID:              actor.TenantID,
		BranchID:              actor.BranchID,
		Email:                 emailAddr,
		EmailNormalized:       emailNormalized,
		Role:                  role,
		TokenHash:             hash,
		ExpiresAt:             expiresAt,
		CreatedByUserID:       actor.UserID,
		CreatedByMembershipID: actor.MembershipID,
		SendCount:             1,
	}

	created, isNew, err := uc.repo.CreateInvite(ctx, actor, inv, func() error {
		return uc.email.SendInvite(ctx, emailAddr, role, acceptURL)
	})
	if err != nil {
		switch err {
		case domain.ErrEmailAlreadyRegistered:
			return CreateInviteResult{}, domainerrors.Conflict("invite_email_already_registered", "This email is already registered.")
		case domain.ErrScopeConflict:
			return CreateInviteResult{}, domainerrors.Conflict("invite_scope_conflict", "A pending invitation already exists for this email with a different role.")
		default:
			uc.logger.Error("create_invite_failed", "error", err)
			return CreateInviteResult{}, domainerrors.Internal(err)
		}
	}

	return CreateInviteResult{Invite: created, IsNew: isNew}, nil
}
