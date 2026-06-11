package httpchild

import (
	"time"

	"nursery-management-system/api/internal/modules/children/domain"
)

type childResponse struct {
	ID                  string   `json:"id"`
	FullName            string   `json:"full_name"`
	DateOfBirth         string   `json:"date_of_birth"`
	StartDate           string   `json:"start_date"`
	EndDate             *string  `json:"end_date,omitempty"`
	CoreHourlyRateMinor *int     `json:"core_hourly_rate_minor"`
	Notes               *string  `json:"notes,omitempty"`
	IsActive            bool     `json:"is_active"`
	LeftAt              *string  `json:"left_at,omitempty"`
	LeftReasonCode      *string  `json:"left_reason_code,omitempty"`
	LeftReasonNote      *string  `json:"left_reason_note,omitempty"`
	EnrollmentComplete  bool     `json:"enrollment_complete"`
	MissingRequirements []string `json:"missing_requirements,omitempty"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

type attendanceChildResponse struct {
	ID                   string  `json:"id"`
	FullName             string  `json:"full_name"`
	EnrollmentComplete   bool    `json:"enrollment_complete"`
	AttendanceState      string  `json:"attendance_state"`
	OpenSessionID        *string `json:"open_session_id,omitempty"`
	CheckedInAt          *string `json:"checked_in_at,omitempty"`
	HasIncompleteSession bool    `json:"has_incomplete_session"`
	AbsenceMarkerID      *string `json:"absence_marker_id,omitempty"`
	AbsenceMarkedAt      *string `json:"absence_marked_at,omitempty"`
}

func toChildResponse(child domain.Child) childResponse {
	resp := childResponse{
		ID:                  child.ID.String(),
		FullName:            child.FullName,
		DateOfBirth:         child.DateOfBirth.Format("2006-01-02"),
		StartDate:           child.StartDate.Format("2006-01-02"),
		CoreHourlyRateMinor: child.CoreHourlyRateMinor,
		Notes:               child.Notes,
		IsActive:            child.IsActive,
		LeftReasonCode:      child.LeftReasonCode,
		LeftReasonNote:      child.LeftReasonNote,
		EnrollmentComplete:  child.EnrollmentComplete(),
		MissingRequirements: child.MissingRequirements(),
		CreatedAt:           child.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           child.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if child.EndDate != nil {
		endDate := child.EndDate.Format("2006-01-02")
		resp.EndDate = &endDate
	}
	if child.LeftAt != nil {
		leftAt := child.LeftAt.UTC().Format(time.RFC3339)
		resp.LeftAt = &leftAt
	}

	return resp
}

func toAttendanceResponse(child domain.AttendanceChild) attendanceChildResponse {
	resp := attendanceChildResponse{
		ID:                   child.ID.String(),
		FullName:             child.FullName,
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
