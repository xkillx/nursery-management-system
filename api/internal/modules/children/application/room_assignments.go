package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/transaction"
	"nursery-management-system/api/internal/platform/uid"
)

type ListRoomAssignments struct {
	repo domain.Repository
}

func NewListRoomAssignments(repo domain.Repository) *ListRoomAssignments {
	return &ListRoomAssignments{repo: repo}
}

func (uc *ListRoomAssignments) Execute(ctx context.Context, actor tenant.ActorContext, childID string) ([]domain.ChildRoomAssignment, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	_, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("check child exists: %w", err))
	}
	if !found {
		return nil, domainerrors.NotFound("child", "Resource not found.")
	}
	assignments, err := uc.repo.ListRoomAssignmentsByChild(ctx, actor.TenantID, actor.BranchID, id)
	if err != nil {
		return nil, domainerrors.Internal(fmt.Errorf("list child room assignments: %w", err))
	}
	return assignments, nil
}

type CreateRoomAssignment struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewCreateRoomAssignment(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *CreateRoomAssignment {
	return &CreateRoomAssignment{repo: repo, audit: auditWriter, txm: txm}
}

type CreateRoomAssignmentInput struct {
	RoomID    string
	StartDate string
}

func (uc *CreateRoomAssignment) Execute(ctx context.Context, actor tenant.ActorContext, childID string, in CreateRoomAssignmentInput) (*domain.ChildRoomAssignment, error) {
	id, err := parseUUID(childID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "child_id")
	}
	roomID, err := uuid.Parse(in.RoomID)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "room_id")
	}
	startDate, err := time.Parse("2006-01-02", in.StartDate)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "start_date")
	}

	var result *domain.ChildRoomAssignment
	err = uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, id)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}

		// close the current assignment (if any)
		if err := uc.repo.CloseCurrentRoomAssignment(ctx, tx, actor.TenantID, actor.BranchID, id, startDate); err != nil {
			return domainerrors.Internal(fmt.Errorf("close current room assignment: %w", err))
		}

		assignment := &domain.ChildRoomAssignment{
			ID:        uid.NewUUID(),
			TenantID:  actor.TenantID,
			BranchID:  actor.BranchID,
			ChildID:   id,
			RoomID:    roomID,
			StartDate: startDate,
		}
		saved, eerr := uc.repo.InsertRoomAssignment(ctx, tx, assignment)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("insert room assignment: %w", eerr))
		}

		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_room_assigned",
			EntityType: "child",
			EntityID:   id,
			Details: map[string]any{
				"room_id":    roomID.String(),
				"start_date": startDate.Format("2006-01-02"),
			},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_room_assigned: %w", aerr))
		}
		result = saved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type CloseRoomAssignment struct {
	repo  domain.Repository
	audit *audit.Writer
	txm   *transaction.Manager
}

func NewCloseRoomAssignment(repo domain.Repository, auditWriter *audit.Writer, txm *transaction.Manager) *CloseRoomAssignment {
	return &CloseRoomAssignment{repo: repo, audit: auditWriter, txm: txm}
}

func (uc *CloseRoomAssignment) Execute(ctx context.Context, actor tenant.ActorContext, childID, assignmentID string) error {
	cid, err := parseUUID(childID)
	if err != nil {
		return domainerrors.Validation("Invalid request payload.", "child_id")
	}
	aid, err := parseUUID(assignmentID)
	if err != nil {
		return domainerrors.Validation("Invalid request payload.", "assignment_id")
	}

	return uc.txm.ExecTx(ctx, func(tx pgx.Tx) error {
		exists, eerr := uc.repo.ExistsInScope(ctx, tx, actor.TenantID, actor.BranchID, cid)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("check child exists: %w", eerr))
		}
		if !exists {
			return domainerrors.NotFound("child", "Resource not found.")
		}
		assignment, found, eerr := uc.repo.GetRoomAssignmentByID(ctx, tx, actor.TenantID, actor.BranchID, aid)
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("get room assignment: %w", eerr))
		}
		if !found || assignment.ChildID != cid {
			return domainerrors.NotFound("room_assignment", "Resource not found.")
		}
		if !assignment.IsCurrent {
			return domainerrors.New("room_assignment_already_closed", "Room assignment is already closed.", "assignment_id")
		}
		closed, eerr := uc.repo.CloseRoomAssignmentByID(ctx, tx, actor.TenantID, actor.BranchID, aid, time.Now().UTC())
		if eerr != nil {
			return domainerrors.Internal(fmt.Errorf("close room assignment: %w", eerr))
		}
		if !closed {
			return domainerrors.New("room_assignment_already_closed", "Room assignment is already closed.", "assignment_id")
		}
		if aerr := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "child_room_unassigned",
			EntityType: "child",
			EntityID:   cid,
			Details: map[string]any{
				"assignment_id": aid.String(),
			},
		}); aerr != nil {
			return domainerrors.Internal(fmt.Errorf("audit child_room_unassigned: %w", aerr))
		}
		return nil
	})
}
