package httpchild

import (
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/children/domain"
)

type childResponse struct {
	ID                      string   `json:"id"`
	FirstName               string   `json:"first_name"`
	MiddleName              *string  `json:"middle_name"`
	LastName                *string  `json:"last_name"`
	DateOfBirth             string   `json:"date_of_birth"`
	StartDate               string   `json:"start_date"`
	EndDate                 *string  `json:"end_date,omitempty"`
	SiteCoreHourlyRateMinor *int     `json:"site_core_hourly_rate_minor"`
	Notes                   *string  `json:"notes,omitempty"`
	IsActive                bool     `json:"is_active"`
	PrimaryRoomID           *string  `json:"primary_room_id,omitempty"`
	HasCurrentRoom          bool     `json:"has_current_room"`
	HasBookingPattern       bool     `json:"has_booking_pattern"`
	EnrollmentComplete      bool     `json:"enrollment_complete"`
	MissingRequirements     []string `json:"missing_requirements,omitempty"`
	CreatedAt               string   `json:"created_at"`
	UpdatedAt               string   `json:"updated_at"`
}

type attendanceChildResponse struct {
	ID                   string  `json:"id"`
	FirstName            string  `json:"first_name"`
	MiddleName           *string `json:"middle_name"`
	LastName             *string `json:"last_name"`
	EnrollmentComplete   bool    `json:"enrollment_complete"`
	AttendanceState      string  `json:"attendance_state"`
	OpenSessionID        *string `json:"open_session_id,omitempty"`
	CheckedInAt          *string `json:"checked_in_at,omitempty"`
	HasIncompleteSession bool    `json:"has_incomplete_session"`
	AbsenceMarkerID      *string `json:"absence_marker_id,omitempty"`
	AbsenceMarkedAt      *string `json:"absence_marked_at,omitempty"`
}

type childWriteRequest struct {
	FirstName   string  `json:"first_name"`
	MiddleName  *string `json:"middle_name"`
	LastName    *string `json:"last_name"`
	DateOfBirth string  `json:"date_of_birth"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	Notes       string  `json:"notes"`
}

type reasonRequest struct {
	ReasonCode string `json:"reason_code"`
	ReasonNote string `json:"reason_note"`
}

func toChildResponse(child domain.Child) childResponse {
	return childResponse{
		ID:                      child.ID.String(),
		FirstName:               child.FirstName,
		MiddleName:              child.MiddleName,
		LastName:                child.LastName,
		DateOfBirth:             child.DateOfBirth.Format("2006-01-02"),
		StartDate:               child.StartDate.Format("2006-01-02"),
		EndDate:                 formatDatePtr(child.EndDate),
		SiteCoreHourlyRateMinor: child.SiteCoreHourlyRateMinor,
		Notes:                   child.Notes,
		IsActive:                child.IsActive,
		PrimaryRoomID:           uuidPtrToStringPtr(child.PrimaryRoomID),
		HasCurrentRoom:          child.HasCurrentRoom,
		HasBookingPattern:       child.HasBookingPattern,
		EnrollmentComplete:      child.EnrollmentComplete(),
		MissingRequirements:     child.MissingRequirements(),
		CreatedAt:               child.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:               child.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func toAttendanceResponse(child domain.AttendanceChild) attendanceChildResponse {
	resp := attendanceChildResponse{
		ID:                   child.ID.String(),
		FirstName:            child.FirstName,
		MiddleName:           child.MiddleName,
		LastName:             child.LastName,
		EnrollmentComplete:   child.EnrollmentComplete,
		AttendanceState:      child.AttendanceState,
		HasIncompleteSession: child.HasIncompleteSession,
	}
	if child.OpenSessionID != nil {
		id := child.OpenSessionID.String()
		resp.OpenSessionID = &id
	}
	if child.CheckedInAt != nil {
		at := child.CheckedInAt.UTC().Format(time.RFC3339)
		resp.CheckedInAt = &at
	}
	if child.AbsenceMarkerID != nil {
		id := child.AbsenceMarkerID.String()
		resp.AbsenceMarkerID = &id
	}
	if child.AbsenceMarkedAt != nil {
		at := child.AbsenceMarkedAt.UTC().Format(time.RFC3339)
		resp.AbsenceMarkedAt = &at
	}
	return resp
}

func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func uuidPtrToStringPtr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}
