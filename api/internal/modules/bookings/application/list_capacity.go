package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/bookings/domain"
)

type RoomInfo struct {
	RoomID   uuid.UUID
	RoomName string
	Capacity int
}

type RoomCapacityEntry struct {
	Date        time.Time
	RoomID      uuid.UUID
	RoomName    string
	Capacity    int
	BookedCount int
}

type RoomCapacityLookup interface {
	ListActiveRooms(ctx context.Context, tenantID, branchID uuid.UUID) ([]RoomInfo, error)
}

type ListCapacity struct {
	repo       domain.Repository
	roomLookup RoomCapacityLookup
}

func NewListCapacity(repo domain.Repository, roomLookup RoomCapacityLookup) *ListCapacity {
	return &ListCapacity{repo: repo, roomLookup: roomLookup}
}

func (uc *ListCapacity) Execute(ctx context.Context, actor BookingActor, siteID uuid.UUID, from, to time.Time) ([]RoomCapacityEntry, error) {
	if err := actor.ValidateSiteAccess(ctx, siteID); err != nil {
		return nil, err
	}

	rooms, err := uc.roomLookup.ListActiveRooms(ctx, actor.TenantID(), siteID)
	if err != nil {
		return nil, internalError(err)
	}

	bookings, err := uc.repo.ListByChildAndDateRange(ctx, actor.TenantID(), siteID, uuid.Nil, from, to)
	if err != nil {
		return nil, internalError(err)
	}

	roomMap := make(map[uuid.UUID]RoomInfo)
	for _, r := range rooms {
		roomMap[r.RoomID] = r
	}

	counts := make(map[string]int)
	for _, b := range bookings {
		if b.RoomID == uuid.Nil {
			continue
		}
		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			dayOfWeek := int32(d.Weekday())
			if dayOfWeek == 0 {
				dayOfWeek = 7
			}
			if containsDay(b.DaysOfWeek, dayOfWeek) &&
				!d.Before(b.EffectiveStartDate) &&
				(b.EffectiveEndDate == nil || !d.After(*b.EffectiveEndDate)) {
				key := b.RoomID.String() + "|" + d.Format("2006-01-02")
				counts[key]++
			}
		}
	}

	var entries []RoomCapacityEntry
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		for _, r := range rooms {
			key := r.RoomID.String() + "|" + d.Format("2006-01-02")
			entries = append(entries, RoomCapacityEntry{
				Date:        d,
				RoomID:      r.RoomID,
				RoomName:    r.RoomName,
				Capacity:    r.Capacity,
				BookedCount: counts[key],
			})
		}
	}

	return entries, nil
}

func containsDay(days []int32, day int32) bool {
	for _, d := range days {
		if d == day {
			return true
		}
	}
	return false
}
