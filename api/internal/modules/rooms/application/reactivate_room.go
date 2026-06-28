package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/rooms/domain"
	"nursery-management-system/api/internal/platform/audit"
)

type ReactivateRoom struct {
	repo  domain.Repository
	txMgr TxManager
	audit *audit.Writer
}

func NewReactivateRoom(repo domain.Repository, txMgr TxManager, auditWriter *audit.Writer) *ReactivateRoom {
	return &ReactivateRoom{repo: repo, txMgr: txMgr, audit: auditWriter}
}

func (uc *ReactivateRoom) Execute(ctx context.Context, actor RoomActor, siteID, roomID uuid.UUID) (domain.Room, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return domain.Room{}, err
	}

	var room domain.Room
	err := uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		existing, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, roomID)
		if err != nil {
			return err
		}

		if existing.IsActive {
			room = existing
			return nil
		}

		if err := uc.repo.Reactivate(ctx, tx, actor.TenantID(), siteID, roomID); err != nil {
			return internalError(err)
		}

		if uc.audit != nil {
			if err := uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "room_reactivated",
				EntityType: "room",
				EntityID:   roomID,
				Details:    map[string]any{},
			}); err != nil {
				return internalError(err)
			}
		}

		room, err = uc.repo.GetByID(ctx, actor.TenantID(), siteID, roomID)
		return err
	})
	if err != nil {
		return domain.Room{}, err
	}

	return room, nil
}
