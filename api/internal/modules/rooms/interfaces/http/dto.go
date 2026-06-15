package httprooms

import (
	"time"

	"nursery-management-system/api/internal/modules/rooms/domain"
)

type roomResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	AgeGroup    string  `json:"age_group"`
	Capacity    int     `json:"capacity"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
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

func toRoomListResponse(rooms []domain.Room) []roomResponse {
	out := make([]roomResponse, 0, len(rooms))
	for _, r := range rooms {
		out = append(out, toRoomResponse(r))
	}
	return out
}
