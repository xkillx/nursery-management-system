package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/parents/domain"
	"nursery-management-system/api/internal/platform/audit"
)

type InviteToPortalUseCase struct {
	repo       domain.Repository
	audit      *audit.Writer
	txMgr      TxManager
	userCreator UserCreator
	emailSender EmailSender
	webBaseURL  string
}

func NewInviteToPortalUseCase(repo domain.Repository, auditWriter *audit.Writer, txMgr TxManager, userCreator UserCreator, emailSender EmailSender, webBaseURL string) *InviteToPortalUseCase {
	return &InviteToPortalUseCase{repo: repo, audit: auditWriter, txMgr: txMgr, userCreator: userCreator, emailSender: emailSender, webBaseURL: webBaseURL}
}

func (uc *InviteToPortalUseCase) Execute(ctx context.Context, actor ActorContext, parentID uuid.UUID) error {
	return uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		parent, found, err := uc.repo.GetByID(ctx, tx, actor.TenantID, actor.BranchID, parentID)
		if err != nil {
			return err
		}
		if !found {
			return domain.ErrParentNotFound
		}
		if !parent.IsActive {
			return domain.ErrParentInactive
		}
		if parent.UserID != nil {
			return domain.ErrUserAlreadyLinked
		}
		if parent.Email == nil || *parent.Email == "" {
			return fmt.Errorf("parent must have an email address to receive portal invite")
		}

		userID, err := uc.userCreator.CreateParentUser(ctx, tx, actor.TenantID, actor.BranchID, *parent.Email)
		if err != nil {
			return err
		}

		if err := uc.repo.SetUserID(ctx, tx, actor.TenantID, actor.BranchID, parentID, &userID); err != nil {
			return err
		}

		acceptURL := fmt.Sprintf("%s/accept-invite?user_id=%s", uc.webBaseURL, userID.String())
		if err := uc.emailSender.SendParentPortalInvite(ctx, *parent.Email, acceptURL); err != nil {
			return fmt.Errorf("send invite email: %w", err)
		}

		return uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "parent_portal_invite_sent",
			EntityType: "parent",
			EntityID:   parentID,
			Details: map[string]any{
				"user_id": userID.String(),
			},
		})
	})
}
