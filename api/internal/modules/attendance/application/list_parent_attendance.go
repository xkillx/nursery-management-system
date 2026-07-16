package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/attendance/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type ParentAttendanceEntry struct {
	ChildID             uuid.UUID
	ChildFirstName      string
	ChildLastName       *string
	SessionTemplateID   uuid.UUID
	SessionTemplateName string
	BookingType         string
	AttendanceID        *uuid.UUID
	AttendanceStatus    *string
	CheckInAt           *time.Time
	CheckOutAt          *time.Time
}

type ParentChildLookupForAttendance interface {
	ListChildIDsForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID) ([]uuid.UUID, error)
}

type ListParentAttendance struct {
	repo      domain.Repository
	childLook ParentChildLookupForAttendance
}

func NewListParentAttendance(repo domain.Repository, childLook ParentChildLookupForAttendance) *ListParentAttendance {
	return &ListParentAttendance{repo: repo, childLook: childLook}
}

func (uc *ListParentAttendance) Execute(ctx context.Context, actor tenant.ActorContext, registerDate time.Time) ([]ParentAttendanceEntry, error) {
	childIDs, err := uc.childLook.ListChildIDsForParent(ctx, actor.TenantID, actor.BranchID, actor.MembershipID)
	if err != nil {
		return nil, domainerrors.Internal(err)
	}
	if len(childIDs) == 0 {
		return nil, nil
	}

	dow := []int32{int32(registerDate.Weekday())}
	allEntries, err := uc.repo.GetRegister(ctx, actor.TenantID, actor.BranchID, registerDate, dow)
	if err != nil {
		return nil, domainerrors.Internal(err)
	}

	childSet := make(map[uuid.UUID]bool, len(childIDs))
	for _, cid := range childIDs {
		childSet[cid] = true
	}

	var results []ParentAttendanceEntry
	for _, e := range allEntries {
		if !childSet[e.ChildID] {
			continue
		}
		results = append(results, ParentAttendanceEntry{
			ChildID:             e.ChildID,
			ChildFirstName:      e.ChildFirstName,
			ChildLastName:       e.ChildLastName,
			SessionTemplateID:   e.SessionTemplateID,
			SessionTemplateName: e.SessionTemplateName,
			BookingType:         e.BookingType,
			AttendanceID:        e.AttendanceID,
			AttendanceStatus:    e.AttendanceStatus,
			CheckInAt:           e.CheckInAt,
			CheckOutAt:          e.CheckOutAt,
		})
	}

	return results, nil
}
