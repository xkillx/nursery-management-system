package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

type dbExecQuerier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func (w *Writer) Write(ctx context.Context, q dbExecQuerier, actor tenant.ActorContext, params WriteParams) error {
	if params.Details == nil {
		params.Details = map[string]any{}
	}
	detailsJSON, err := json.Marshal(params.Details)
	if err != nil {
		return fmt.Errorf("marshal audit details: %w", err)
	}

	var reasonCodeValue any
	if params.ReasonCode != nil {
		reasonCodeValue = *params.ReasonCode
	}

	var reasonNoteValue any
	if params.ReasonNote != nil {
		reasonNoteValue = *params.ReasonNote
	}

	const insertQ = `
INSERT INTO audit_logs (
    id, tenant_id, branch_id, actor_user_id, actor_membership_id,
    action_type, action_entity_type, action_entity_id,
    reason_code, reason_note, request_id, details
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11, $12::jsonb)`

	_, err = q.Exec(ctx, insertQ,
		uid.NewUUID(),
		actor.TenantID,
		actor.BranchID,
		actor.UserID,
		actor.MembershipID,
		params.ActionType,
		params.EntityType,
		params.EntityID,
		reasonCodeValue,
		reasonNoteValue,
		actor.RequestID,
		string(detailsJSON),
	)
	return err
}

func (w *Writer) WriteWithTx(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, params WriteParams) error {
	return w.Write(ctx, tx, actor, params)
}

type WriteParams struct {
	ActionType string
	EntityType string
	EntityID   uuid.UUID
	ReasonCode *string
	ReasonNote *string
	Details    map[string]any
}
