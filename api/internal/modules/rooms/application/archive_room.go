package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/rooms/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type ArchiveRoom struct {
	repo  domain.Repository
	txMgr TxManager
	audit *audit.Writer
}

func NewArchiveRoom(repo domain.Repository, txMgr TxManager, auditWriter *audit.Writer) *ArchiveRoom {
	return &ArchiveRoom{repo: repo, txMgr: txMgr, audit: auditWriter}
}

func (uc *ArchiveRoom) Execute(ctx context.Context, actor RoomActor, siteID, roomID uuid.UUID) error {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return err
	}

	return uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		existing, err := uc.repo.GetByIDForUpdate(ctx, tx, actor.TenantID(), siteID, roomID)
		if err != nil {
			return err
		}

		if !existing.IsActive {
			return domainerrors.Conflict("room_not_active", "Room is already archived.")
		}

		count, err := uc.repo.CountActiveChildren(ctx, tx, actor.TenantID(), siteID, roomID)
		if err != nil {
			return internalError(err)
		}
		if count > 0 {
			return domainerrors.ConflictWithDetails(
				"room_has_children",
				fmt.Sprintf("Room has %d active children assigned — reassign them before archiving.", count),
				map[string]any{"assigned_count": count},
			)
		}

		if err := uc.repo.Archive(ctx, tx, actor.TenantID(), siteID, roomID); err != nil {
			return internalError(err)
		}

		if uc.audit != nil {
			if err := uc.audit.WriteSystemWithTx(ctx, tx, actor.TenantID(), siteID, actor.RequestID(), audit.WriteParams{
				ActionType: "room_archived",
				EntityType: "room",
				EntityID:   roomID,
				Details:    map[string]any{},
			}); err != nil {
				return internalError(err)
			}
		}

		return nil
	})
}
