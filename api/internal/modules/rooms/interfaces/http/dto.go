package httprooms

import (
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/rooms/domain"
)

type roomResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	AgeGroup       string  `json:"age_group"`
	Capacity       int     `json:"capacity"`
	IsActive       bool    `json:"is_active"`
	AssignedCount  *int    `json:"assigned_count,omitempty"`
	IsOverCapacity *bool   `json:"is_over_capacity,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type createRoomRequest struct {
	Name        string `json:"name" binding:"required"`
	AgeGroup    string `json:"age_group" binding:"required"`
	Capacity    int    `json:"capacity" binding:"required"`
	Description string `json:"description"`
}

type updateRoomRequest struct {
	Name        *string `json:"name"`
	AgeGroup    *string `json:"age_group"`
	Capacity    *int    `json:"capacity"`
	Description *string `json:"description"`
}

func toRoomResponse(room domain.Room) roomResponse {
	return roomResponse{
		ID:          room.ID.String(),
		Name:        room.Name,
		Description: room.Description,
		AgeGroup:    room.AgeGroup,
		Capacity:    room.Capacity,
		IsActive:    room.IsActive,
		CreatedAt:   room.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   room.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toRoomListResponse(rooms []domain.Room, counts map[uuid.UUID]int) []roomResponse {
	out := make([]roomResponse, 0, len(rooms))
	for _, r := range rooms {
		out = append(out, toRoomListItem(r, counts))
	}
	return out
}

func toRoomListItem(room domain.Room, counts map[uuid.UUID]int) roomResponse {
	resp := toRoomResponse(room)
	if counts == nil {
		return resp
	}
	assigned, ok := counts[room.ID]
	if !ok {
		return resp
	}
	count := assigned
	over := room.Capacity > 0 && count > room.Capacity
	resp.AssignedCount = &count
	resp.IsOverCapacity = &over
	return resp
}
