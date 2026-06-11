package application

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/owner/domain"
)

type GrantOutcome string

const (
	GrantOutcomeGranted       GrantOutcome = "manager_membership_granted"
	GrantOutcomeReactivated   GrantOutcome = "manager_membership_reactivated"
	GrantOutcomeAlreadyActive GrantOutcome = "manager_membership_already_active"
	GrantOutcomeInvitePending GrantOutcome = "manager_invite_pending"
)

type GrantManagerAccessResult struct {
	Outcome       GrantOutcome
	MembershipID  *uuid.UUID
	InviteDetails *InviteDetails
}

type InviteDetails struct {
	Email     string
	ExpiresAt time.Time
	SendCount int
}

type GrantManagerAccessUseCase struct {
	repo          domain.ManagerAccessRepository
	tokenGen      domain.InviteTokenGenerator
	emailSender   domain.ManagerInviteSender
	webBaseURL    string
}

func NewGrantManagerAccessUseCase(
	repo domain.ManagerAccessRepository,
	tokenGen domain.InviteTokenGenerator,
	emailSender domain.ManagerInviteSender,
	webBaseURL string,
) *GrantManagerAccessUseCase {
	return &GrantManagerAccessUseCase{
		repo:        repo,
		tokenGen:    tokenGen,
		emailSender: emailSender,
		webBaseURL:  webBaseURL,
	}
}

func (uc *GrantManagerAccessUseCase) Execute(ctx context.Context, actor domain.OwnerActor, siteID uuid.UUID, emailAddr string) (GrantManagerAccessResult, error) {
	if _, err := uc.repo.GetActiveSite(ctx, actor.TenantID, siteID); err != nil {
		return GrantManagerAccessResult{}, err
	}

	emailNormalized := strings.ToLower(strings.TrimSpace(emailAddr))

	userID, err := uc.repo.FindActiveUserByEmail(ctx, emailNormalized)
	if err != nil {
		return GrantManagerAccessResult{}, fmt.Errorf("lookup user: %w", err)
	}

	if userID != nil {
		return uc.handleExistingUser(ctx, actor, siteID, *userID)
	}

	return uc.handleNewUser(ctx, actor, siteID, emailAddr, emailNormalized)
}

func (uc *GrantManagerAccessUseCase) handleExistingUser(ctx context.Context, actor domain.OwnerActor, siteID, userID uuid.UUID) (GrantManagerAccessResult, error) {
	existing, err := uc.repo.FindManagerMembership(ctx, actor.TenantID, siteID, userID)
	if err != nil {
		return GrantManagerAccessResult{}, fmt.Errorf("find membership: %w", err)
	}

	if existing != nil && existing.IsActive {
		return GrantManagerAccessResult{
			Outcome:      GrantOutcomeAlreadyActive,
			MembershipID: &existing.ID,
		}, nil
	}

	if existing != nil && !existing.IsActive {
		if err := uc.repo.ReactivateManagerMembership(ctx, existing.ID, actor.TenantID); err != nil {
			return GrantManagerAccessResult{}, fmt.Errorf("reactivate: %w", err)
		}
		return GrantManagerAccessResult{
			Outcome:      GrantOutcomeReactivated,
			MembershipID: &existing.ID,
		}, nil
	}

	id := uuid.New()
	if err := uc.repo.CreateManagerMembership(ctx, id, actor.TenantID, siteID, userID); err != nil {
		return GrantManagerAccessResult{}, fmt.Errorf("create membership: %w", err)
	}
	return GrantManagerAccessResult{
		Outcome:      GrantOutcomeGranted,
		MembershipID: &id,
	}, nil
}

func (uc *GrantManagerAccessUseCase) handleNewUser(ctx context.Context, actor domain.OwnerActor, siteID uuid.UUID, emailAddr, emailNormalized string) (GrantManagerAccessResult, error) {
	raw, hash, expiresAt, err := uc.tokenGen.Generate()
	if err != nil {
		return GrantManagerAccessResult{}, fmt.Errorf("generate token: %w", err)
	}

	existing, err := uc.repo.FindPendingManagerInvite(ctx, actor.TenantID, siteID, emailNormalized)
	if err != nil {
		return GrantManagerAccessResult{}, fmt.Errorf("find pending invite: %w", err)
	}

	if existing != nil {
		if err := uc.repo.RefreshManagerInvite(ctx, existing.ID, hash, expiresAt, actor.UserID, actor.MembershipID); err != nil {
			return GrantManagerAccessResult{}, fmt.Errorf("refresh invite: %w", err)
		}
	} else {
		inviteID := uuid.New()
		if err := uc.repo.CreateManagerInvite(ctx, inviteID, actor.TenantID, siteID, emailAddr, emailNormalized, hash, expiresAt, actor.UserID, actor.MembershipID); err != nil {
			return GrantManagerAccessResult{}, fmt.Errorf("create invite: %w", err)
		}
	}

	acceptURL := fmt.Sprintf("%s/invite-accept?token=%s", uc.webBaseURL, url.QueryEscape(raw))
	_ = uc.emailSender.SendManagerInvite(ctx, emailAddr, acceptURL)

	return GrantManagerAccessResult{
		Outcome: GrantOutcomeInvitePending,
		InviteDetails: &InviteDetails{
			Email:     emailAddr,
			ExpiresAt: expiresAt,
			SendCount: 1,
		},
	}, nil
}
