package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/modules/absence/domain"
	attendancedomain "nursery-management-system/api/internal/modules/attendance/domain"
	"nursery-management-system/api/internal/modules/attendance/application"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type MarkAbsentResult struct {
	Marker  domain.AbsenceMarker
	Created bool
}

type txManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type auditWriter interface {
	WriteWithTx(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, params audit.WriteParams) error
}

type MarkAbsent struct {
	repo         domain.Repository
	childChecker domain.ChildEnrollmentChecker
	txMgr        txManager
	audit        auditWriter
	clock        *application.AttendanceClock
}

func NewMarkAbsent(
	repo domain.Repository,
	childChecker domain.ChildEnrollmentChecker,
	txMgr txManager,
	auditWriter auditWriter,
	clock *application.AttendanceClock,
) *MarkAbsent {
	return &MarkAbsent{
		repo:         repo,
		childChecker: childChecker,
		txMgr:        txMgr,
		audit:        auditWriter,
		clock:        clock,
	}
}

func (uc *MarkAbsent) Execute(ctx context.Context, actor tenant.ActorContext, childIDRaw string) (MarkAbsentResult, error) {
	childID, err := uuid.Parse(childIDRaw)
	if err != nil {
		return MarkAbsentResult{}, domainerrors.Validation("Invalid child ID format.", "child_id")
	}

	var result MarkAbsentResult

	err = uc.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		now, localDate := uc.clock.Now()

		if err := uc.childChecker.CheckEnrollmentForAttendance(ctx, tx, actor.TenantID, actor.BranchID, childID, localDate); err != nil {
			return mapMarkAbsentError(err)
		}

		existing, found, err := uc.repo.FindActiveByChildDate(ctx, tx, actor.TenantID, actor.BranchID, childID, localDate)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("find active absence marker: %w", err))
		}
		if found {
			result = MarkAbsentResult{Marker: existing, Created: false}
			return nil
		}

		hasAttendance, err := uc.repo.HasAttendanceForChildDate(ctx, tx, actor.TenantID, actor.BranchID, childID, localDate)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("check attendance for child date: %w", err))
		}
		if hasAttendance {
			return domainerrors.Conflict("absence_attendance_exists", "Attendance already exists for this child today.")
		}

		marker := domain.AbsenceMarker{
			ID:                   uid.NewUUID(),
			TenantID:             actor.TenantID,
			BranchID:             actor.BranchID,
			ChildID:              childID,
			LocalDate:            localDate,
			MarkedAt:             now,
			MarkedByUserID:       actor.UserID,
			MarkedByMembershipID: actor.MembershipID,
		}

		created, err := uc.repo.Create(ctx, tx, marker)
		if err != nil {
			return domainerrors.Internal(fmt.Errorf("create absence marker: %w", err))
		}

		if err := uc.audit.WriteWithTx(ctx, tx, actor, audit.WriteParams{
			ActionType: "absence_marker_created",
			EntityType: "absence_marker",
			EntityID:   created.ID,
			Details: map[string]any{
				"child_id":   childID.String(),
				"local_date": localDate.Format("2006-01-02"),
				"marked_at":  now.Format("2006-01-02T15:04:05Z"),
			},
		}); err != nil {
			return domainerrors.Internal(fmt.Errorf("audit absence_marker_created: %w", err))
		}

		result = MarkAbsentResult{Marker: created, Created: true}
		return nil
	})

	if err != nil {
		return MarkAbsentResult{}, err
	}
	return result, nil
}

func mapMarkAbsentError(err error) error {
	if errors.Is(err, attendancedomain.ErrChildNotFound) {
		return domainerrors.NotFound("child", "Resource not found.")
	}
	if errors.Is(err, attendancedomain.ErrChildEnrollmentIncomplete) {
		return domainerrors.Conflict("child_enrollment_incomplete", "Child enrollment is not complete.")
	}
	return domainerrors.Internal(fmt.Errorf("mark absent: %w", err))
}
