package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	statusFilterActive   = "active"
	statusFilterInactive = "inactive"
	statusFilterAll      = "all"

	reasonCodeDuplicateRecord      = "duplicate_record"
	reasonCodeEnteredInError       = "entered_in_error"
	reasonCodeLeftNursery          = "left_nursery"
	reasonCodeSafeguardingDir      = "safeguarding_direction"
	reasonCodeContactUpdate        = "contact_update"
	reasonCodeAccessRevoked        = "access_revoked"
	reasonCodeOther                = "other"
	maxReasonNoteLen               = 500
	defaultListLimit               = 50
	maxListLimit                   = 200
	cascadeReasonGuardianDeactNote = "guardian_deactivated_cascade"
)

var lifecycleReasonCodes = map[string]struct{}{
	reasonCodeDuplicateRecord: {},
	reasonCodeEnteredInError:  {},
	reasonCodeLeftNursery:     {},
	reasonCodeSafeguardingDir: {},
	reasonCodeContactUpdate:   {},
	reasonCodeAccessRevoked:   {},
	reasonCodeOther:           {},
}

type peopleHandler struct {
	pool *pgxpool.Pool
}

func newPeopleHandler(pool *pgxpool.Pool) *peopleHandler {
	return &peopleHandler{pool: pool}
}

type childWriteRequest struct {
	FullName            string `json:"full_name"`
	DateOfBirth         string `json:"date_of_birth"`
	StartDate           string `json:"start_date"`
	EndDate             string `json:"end_date"`
	CoreHourlyRateMinor *int   `json:"core_hourly_rate_minor"`
	Notes               string `json:"notes"`
}

type guardianWriteRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Notes    string `json:"notes"`
}

type reasonRequest struct {
	ReasonCode string `json:"reason_code"`
	ReasonNote string `json:"reason_note"`
}

type guardianChildLinkCreateRequest struct {
	GuardianID string `json:"guardian_id"`
	ChildID    string `json:"child_id"`
}

type parentMappingCreateRequest struct {
	MembershipID string `json:"membership_id"`
	GuardianID   string `json:"guardian_id"`
}

type childResponse struct {
	ID                  string   `json:"id"`
	FullName            string   `json:"full_name"`
	DateOfBirth         string   `json:"date_of_birth"`
	StartDate           string   `json:"start_date"`
	EndDate             *string  `json:"end_date,omitempty"`
	CoreHourlyRateMinor int      `json:"core_hourly_rate_minor"`
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

type guardianResponse struct {
	ID                     string  `json:"id"`
	FullName               string  `json:"full_name"`
	Email                  *string `json:"email,omitempty"`
	Phone                  *string `json:"phone,omitempty"`
	Notes                  *string `json:"notes,omitempty"`
	IsActive               bool    `json:"is_active"`
	DeactivatedAt          *string `json:"deactivated_at,omitempty"`
	DeactivationReasonCode *string `json:"deactivation_reason_code,omitempty"`
	DeactivationReasonNote *string `json:"deactivation_reason_note,omitempty"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

type guardianChildLinkResponse struct {
	ID              string  `json:"id"`
	GuardianID      string  `json:"guardian_id"`
	ChildID         string  `json:"child_id"`
	EndedAt         *string `json:"ended_at,omitempty"`
	EndedReasonCode *string `json:"ended_reason_code,omitempty"`
	EndedReasonNote *string `json:"ended_reason_note,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type parentMembershipGuardianResponse struct {
	ID              string  `json:"id"`
	MembershipID    string  `json:"membership_id"`
	GuardianID      string  `json:"guardian_id"`
	EndedAt         *string `json:"ended_at,omitempty"`
	EndedReasonCode *string `json:"ended_reason_code,omitempty"`
	EndedReasonNote *string `json:"ended_reason_note,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type attendanceChildResponse struct {
	ID                 string `json:"id"`
	FullName           string `json:"full_name"`
	EnrollmentComplete bool   `json:"enrollment_complete"`
}

type actorContext struct {
	UserID       uuid.UUID
	MembershipID uuid.UUID
	TenantID     uuid.UUID
	BranchID     uuid.UUID
	RequestID    string
}

type childRow struct {
	ID                  uuid.UUID
	FullName            string
	DateOfBirth         time.Time
	StartDate           time.Time
	EndDate             *time.Time
	CoreHourlyRateMinor int
	Notes               *string
	IsActive            bool
	LeftAt              *time.Time
	LeftReasonCode      *string
	LeftReasonNote      *string
	HasGuardianLink     bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type guardianRow struct {
	ID                     uuid.UUID
	FullName               string
	Email                  *string
	Phone                  *string
	Notes                  *string
	IsActive               bool
	DeactivatedAt          *time.Time
	DeactivationReasonCode *string
	DeactivationReasonNote *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type guardianChildLinkRow struct {
	ID              uuid.UUID
	GuardianID      uuid.UUID
	ChildID         uuid.UUID
	EndedAt         *time.Time
	EndedReasonCode *string
	EndedReasonNote *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type parentMembershipGuardianRow struct {
	ID              uuid.UUID
	MembershipID    uuid.UUID
	GuardianID      uuid.UUID
	EndedAt         *time.Time
	EndedReasonCode *string
	EndedReasonNote *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (h *peopleHandler) registerRoutes(protected *gin.RouterGroup) {
	protected.GET("/children/attendance", requireRoles("manager", "practitioner"), h.listAttendanceChildren)

	manager := protected.Group("")
	manager.Use(requireRoles("manager"))

	manager.GET("/children", h.listChildren)
	manager.GET("/children/:child_id", h.getChild)
	manager.POST("/children", h.createChild)
	manager.PATCH("/children/:child_id", h.updateChild)
	manager.POST("/children/:child_id/actions/mark-inactive", h.markChildInactive)

	manager.GET("/guardians", h.listGuardians)
	manager.GET("/guardians/:guardian_id", h.getGuardian)
	manager.POST("/guardians", h.createGuardian)
	manager.PATCH("/guardians/:guardian_id", h.updateGuardian)
	manager.POST("/guardians/:guardian_id/actions/deactivate", h.deactivateGuardian)
	manager.POST("/guardians/:guardian_id/actions/reactivate", h.reactivateGuardian)

	manager.POST("/guardian-child-links", h.createGuardianChildLink)
	manager.POST("/guardian-child-links/:link_id/actions/end", h.endGuardianChildLink)

	manager.POST("/parent-membership-guardian-mappings", h.createParentMapping)
	manager.POST("/parent-membership-guardian-mappings/:mapping_id/actions/end", h.endParentMapping)
}

func (h *peopleHandler) listChildren(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	statusFilter, ok := parseStatusFilter(c)
	if !ok {
		return
	}

	limit, offset, ok := parsePagination(c)
	if !ok {
		return
	}

	rows, err := h.fetchChildren(c.Request.Context(), actor.TenantID, actor.BranchID, statusFilter, limit, offset)
	if err != nil {
		writeInternalError(c)
		return
	}

	out := make([]childResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toChildResponse(row))
	}

	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *peopleHandler) getChild(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID, ok := parseUUIDParam(c, "child_id")
	if !ok {
		return
	}

	row, found, err := h.fetchChildByID(c.Request.Context(), actor.TenantID, actor.BranchID, childID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !found {
		writeError(c, http.StatusNotFound, "child_not_found", "Resource not found.", nil)
		return
	}

	c.JSON(http.StatusOK, toChildResponse(row))
}

func (h *peopleHandler) createChild(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req childWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	if req.FullName == "" {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "full_name"})
		return
	}

	dob, err := parseDate(req.DateOfBirth)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "date_of_birth"})
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "start_date"})
		return
	}

	if req.CoreHourlyRateMinor == nil || *req.CoreHourlyRateMinor < 0 {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "core_hourly_rate_minor"})
		return
	}

	var endDate *time.Time
	if strings.TrimSpace(req.EndDate) != "" {
		parsed, parseErr := parseDate(req.EndDate)
		if parseErr != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "end_date"})
			return
		}
		if parsed.Before(startDate) {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "end_date"})
			return
		}
		endDate = &parsed
	}

	notes := strings.TrimSpace(req.Notes)
	if notes == "" {
		notes = ""
	}

	childID := newUUID()
	const q = `
INSERT INTO children (
    id, tenant_id, branch_id, full_name, date_of_birth, start_date, end_date,
    core_hourly_rate_minor, notes, is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), true)`

	if _, err := h.pool.Exec(c.Request.Context(), q,
		childID,
		actor.TenantID,
		actor.BranchID,
		req.FullName,
		dob,
		startDate,
		endDate,
		*req.CoreHourlyRateMinor,
		notes,
	); err != nil {
		writeInternalError(c)
		return
	}

	row, found, err := h.fetchChildByID(c.Request.Context(), actor.TenantID, actor.BranchID, childID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := h.insertAuditLog(c.Request.Context(), actor, "child_created", "child", childID, nil, nil, map[string]any{}); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toChildResponse(row))
}

func (h *peopleHandler) updateChild(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID, ok := parseUUIDParam(c, "child_id")
	if !ok {
		return
	}

	var req childWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	fields := make([]string, 0, 6)
	args := make([]any, 0, 10)
	args = append(args, actor.TenantID, actor.BranchID, childID)
	argPos := 4

	if req.FullName != "" {
		fullName := strings.TrimSpace(req.FullName)
		if fullName == "" {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "full_name"})
			return
		}
		fields = append(fields, fmt.Sprintf("full_name = $%d", argPos))
		args = append(args, fullName)
		argPos++
	}

	if req.DateOfBirth != "" {
		dob, err := parseDate(req.DateOfBirth)
		if err != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "date_of_birth"})
			return
		}
		fields = append(fields, fmt.Sprintf("date_of_birth = $%d", argPos))
		args = append(args, dob)
		argPos++
	}

	if req.StartDate != "" {
		startDate, err := parseDate(req.StartDate)
		if err != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "start_date"})
			return
		}
		fields = append(fields, fmt.Sprintf("start_date = $%d", argPos))
		args = append(args, startDate)
		argPos++
	}

	if req.EndDate != "" {
		endDate, err := parseDate(req.EndDate)
		if err != nil {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "end_date"})
			return
		}
		fields = append(fields, fmt.Sprintf("end_date = $%d", argPos))
		args = append(args, endDate)
		argPos++
	}

	if req.CoreHourlyRateMinor != nil {
		if *req.CoreHourlyRateMinor < 0 {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "core_hourly_rate_minor"})
			return
		}
		fields = append(fields, fmt.Sprintf("core_hourly_rate_minor = $%d", argPos))
		args = append(args, *req.CoreHourlyRateMinor)
		argPos++
	}

	if req.Notes != "" {
		notes := strings.TrimSpace(req.Notes)
		fields = append(fields, fmt.Sprintf("notes = NULLIF($%d, '')", argPos))
		args = append(args, notes)
		argPos++
	}

	if len(fields) == 0 {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "body"})
		return
	}

	q := fmt.Sprintf(`
UPDATE children
SET %s, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3`, strings.Join(fields, ", "))

	ct, err := h.pool.Exec(c.Request.Context(), q, args...)
	if err != nil {
		writeInternalError(c)
		return
	}
	if ct.RowsAffected() == 0 {
		writeError(c, http.StatusNotFound, "child_not_found", "Resource not found.", nil)
		return
	}

	row, found, err := h.fetchChildByID(c.Request.Context(), actor.TenantID, actor.BranchID, childID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := h.insertAuditLog(c.Request.Context(), actor, "child_updated", "child", childID, nil, nil, map[string]any{}); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toChildResponse(row))
}

func (h *peopleHandler) markChildInactive(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	childID, ok := parseUUIDParam(c, "child_id")
	if !ok {
		return
	}

	reason, ok := parseReasonPayload(c, "child_lifecycle_reason_required")
	if !ok {
		return
	}

	tx, err := h.pool.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		writeInternalError(c)
		return
	}
	defer tx.Rollback(c.Request.Context())

	row, found, err := h.fetchChildByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !found {
		writeError(c, http.StatusNotFound, "child_not_found", "Resource not found.", nil)
		return
	}

	if row.IsActive {
		const q = `
UPDATE children
SET is_active = false,
    left_at = now(),
    left_reason_code = $1,
    left_reason_note = NULLIF($2, ''),
    updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5`
		if _, err := tx.Exec(c.Request.Context(), q, reason.Code, reason.Note, actor.TenantID, actor.BranchID, childID); err != nil {
			writeInternalError(c)
			return
		}

		if err := h.insertAuditLogTx(c.Request.Context(), tx, actor, "child_marked_inactive", "child", childID, &reason.Code, nullableReasonNote(reason.Note), map[string]any{}); err != nil {
			writeInternalError(c)
			return
		}
	}

	updated, found, err := h.fetchChildByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, childID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toChildResponse(updated))
}

func (h *peopleHandler) listGuardians(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	statusFilter, ok := parseStatusFilter(c)
	if !ok {
		return
	}

	limit, offset, ok := parsePagination(c)
	if !ok {
		return
	}

	rows, err := h.fetchGuardians(c.Request.Context(), actor.TenantID, actor.BranchID, statusFilter, limit, offset)
	if err != nil {
		writeInternalError(c)
		return
	}

	out := make([]guardianResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toGuardianResponse(row))
	}

	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *peopleHandler) getGuardian(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	guardianID, ok := parseUUIDParam(c, "guardian_id")
	if !ok {
		return
	}

	row, found, err := h.fetchGuardianByID(c.Request.Context(), actor.TenantID, actor.BranchID, guardianID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !found {
		writeError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
		return
	}

	c.JSON(http.StatusOK, toGuardianResponse(row))
}

func (h *peopleHandler) createGuardian(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req guardianWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	if req.FullName == "" {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "full_name"})
		return
	}

	guardianID := newUUID()
	const q = `
INSERT INTO guardians (id, tenant_id, branch_id, full_name, email, phone, notes, is_active)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), true)`

	if _, err := h.pool.Exec(c.Request.Context(), q,
		guardianID,
		actor.TenantID,
		actor.BranchID,
		req.FullName,
		strings.TrimSpace(req.Email),
		strings.TrimSpace(req.Phone),
		strings.TrimSpace(req.Notes),
	); err != nil {
		writeInternalError(c)
		return
	}

	row, found, err := h.fetchGuardianByID(c.Request.Context(), actor.TenantID, actor.BranchID, guardianID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := h.insertAuditLog(c.Request.Context(), actor, "guardian_created", "guardian", guardianID, nil, nil, map[string]any{}); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toGuardianResponse(row))
}

func (h *peopleHandler) updateGuardian(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	guardianID, ok := parseUUIDParam(c, "guardian_id")
	if !ok {
		return
	}

	var req guardianWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	fields := make([]string, 0, 4)
	args := make([]any, 0, 8)
	args = append(args, actor.TenantID, actor.BranchID, guardianID)
	argPos := 4

	if req.FullName != "" {
		fullName := strings.TrimSpace(req.FullName)
		if fullName == "" {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "full_name"})
			return
		}
		fields = append(fields, fmt.Sprintf("full_name = $%d", argPos))
		args = append(args, fullName)
		argPos++
	}

	if req.Email != "" {
		fields = append(fields, fmt.Sprintf("email = NULLIF($%d, '')", argPos))
		args = append(args, strings.TrimSpace(req.Email))
		argPos++
	}

	if req.Phone != "" {
		fields = append(fields, fmt.Sprintf("phone = NULLIF($%d, '')", argPos))
		args = append(args, strings.TrimSpace(req.Phone))
		argPos++
	}

	if req.Notes != "" {
		fields = append(fields, fmt.Sprintf("notes = NULLIF($%d, '')", argPos))
		args = append(args, strings.TrimSpace(req.Notes))
		argPos++
	}

	if len(fields) == 0 {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "body"})
		return
	}

	q := fmt.Sprintf(`
UPDATE guardians
SET %s, updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3`, strings.Join(fields, ", "))

	ct, err := h.pool.Exec(c.Request.Context(), q, args...)
	if err != nil {
		writeInternalError(c)
		return
	}
	if ct.RowsAffected() == 0 {
		writeError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
		return
	}

	row, found, err := h.fetchGuardianByID(c.Request.Context(), actor.TenantID, actor.BranchID, guardianID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := h.insertAuditLog(c.Request.Context(), actor, "guardian_updated", "guardian", guardianID, nil, nil, map[string]any{}); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toGuardianResponse(row))
}

func (h *peopleHandler) deactivateGuardian(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	guardianID, ok := parseUUIDParam(c, "guardian_id")
	if !ok {
		return
	}

	reason, ok := parseReasonPayload(c, "guardian_deactivation_reason_required")
	if !ok {
		return
	}

	tx, err := h.pool.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		writeInternalError(c)
		return
	}
	defer tx.Rollback(c.Request.Context())

	guardian, found, err := h.fetchGuardianByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, guardianID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !found {
		writeError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
		return
	}

	if guardian.IsActive {
		const deactivateQ = `
UPDATE guardians
SET is_active = false,
    deactivated_at = now(),
    deactivation_reason_code = $1,
    deactivation_reason_note = NULLIF($2, ''),
    updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5`
		if _, err := tx.Exec(c.Request.Context(), deactivateQ, reason.Code, reason.Note, actor.TenantID, actor.BranchID, guardianID); err != nil {
			writeInternalError(c)
			return
		}

		const cascadeLinksQ = `
UPDATE guardian_child_links
SET ended_at = now(),
    ended_reason_code = $1,
    ended_reason_note = $2,
    updated_at = now()
WHERE tenant_id = $3
  AND branch_id = $4
  AND guardian_id = $5
  AND ended_at IS NULL`
		if _, err := tx.Exec(c.Request.Context(), cascadeLinksQ, reasonCodeAccessRevoked, cascadeReasonGuardianDeactNote, actor.TenantID, actor.BranchID, guardianID); err != nil {
			writeInternalError(c)
			return
		}

		const cascadeMappingsQ = `
UPDATE parent_membership_guardians
SET ended_at = now(),
    ended_reason_code = $1,
    ended_reason_note = $2,
    updated_at = now()
WHERE tenant_id = $3
  AND branch_id = $4
  AND guardian_id = $5
  AND ended_at IS NULL`
		if _, err := tx.Exec(c.Request.Context(), cascadeMappingsQ, reasonCodeAccessRevoked, cascadeReasonGuardianDeactNote, actor.TenantID, actor.BranchID, guardianID); err != nil {
			writeInternalError(c)
			return
		}

		if err := h.insertAuditLogTx(c.Request.Context(), tx, actor, "guardian_deactivated", "guardian", guardianID, &reason.Code, nullableReasonNote(reason.Note), map[string]any{}); err != nil {
			writeInternalError(c)
			return
		}
	}

	updated, found, err := h.fetchGuardianByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, guardianID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toGuardianResponse(updated))
}

func (h *peopleHandler) reactivateGuardian(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	guardianID, ok := parseUUIDParam(c, "guardian_id")
	if !ok {
		return
	}

	tx, err := h.pool.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		writeInternalError(c)
		return
	}
	defer tx.Rollback(c.Request.Context())

	guardian, found, err := h.fetchGuardianByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, guardianID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !found {
		writeError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
		return
	}

	if !guardian.IsActive {
		const q = `
UPDATE guardians
SET is_active = true,
    deactivated_at = NULL,
    deactivation_reason_code = NULL,
    deactivation_reason_note = NULL,
    updated_at = now()
WHERE tenant_id = $1 AND branch_id = $2 AND id = $3`
		if _, err := tx.Exec(c.Request.Context(), q, actor.TenantID, actor.BranchID, guardianID); err != nil {
			writeInternalError(c)
			return
		}

		if err := h.insertAuditLogTx(c.Request.Context(), tx, actor, "guardian_reactivated", "guardian", guardianID, nil, nil, map[string]any{}); err != nil {
			writeInternalError(c)
			return
		}
	}

	updated, found, err := h.fetchGuardianByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, guardianID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toGuardianResponse(updated))
}

func (h *peopleHandler) createGuardianChildLink(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req guardianChildLinkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	guardianID, err := uuid.Parse(strings.TrimSpace(req.GuardianID))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "guardian_id"})
		return
	}
	childID, err := uuid.Parse(strings.TrimSpace(req.ChildID))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "child_id"})
		return
	}

	tx, err := h.pool.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		writeInternalError(c)
		return
	}
	defer tx.Rollback(c.Request.Context())

	guardianActive, exists, err := h.fetchGuardianActive(c.Request.Context(), tx, actor.TenantID, actor.BranchID, guardianID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !exists {
		writeError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
		return
	}
	if !guardianActive {
		writeError(c, http.StatusBadRequest, "guardian_not_active", "Invalid request payload.", nil)
		return
	}

	if exists, err := h.childExistsInScope(c.Request.Context(), tx, actor.TenantID, actor.BranchID, childID); err != nil {
		writeInternalError(c)
		return
	} else if !exists {
		writeError(c, http.StatusNotFound, "child_not_found", "Resource not found.", nil)
		return
	}

	existing, exists, err := h.fetchActiveGuardianChildLinkForPair(c.Request.Context(), tx, actor.TenantID, actor.BranchID, guardianID, childID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if exists {
		if err := tx.Commit(c.Request.Context()); err != nil {
			writeInternalError(c)
			return
		}
		c.JSON(http.StatusOK, toGuardianChildLinkResponse(existing))
		return
	}

	createdID := newUUID()
	const insertQ = `
INSERT INTO guardian_child_links (id, tenant_id, branch_id, guardian_id, child_id)
VALUES ($1, $2, $3, $4, $5)`

	if _, err := tx.Exec(c.Request.Context(), insertQ, createdID, actor.TenantID, actor.BranchID, guardianID, childID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			retryRow, retryFound, retryErr := h.fetchActiveGuardianChildLinkForPair(c.Request.Context(), tx, actor.TenantID, actor.BranchID, guardianID, childID)
			if retryErr != nil {
				writeInternalError(c)
				return
			}
			if retryFound {
				if err := tx.Commit(c.Request.Context()); err != nil {
					writeInternalError(c)
					return
				}
				c.JSON(http.StatusOK, toGuardianChildLinkResponse(retryRow))
				return
			}
		}
		writeInternalError(c)
		return
	}

	created, found, err := h.fetchGuardianChildLinkByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, createdID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := h.insertAuditLogTx(c.Request.Context(), tx, actor, "guardian_child_link_created", "guardian_child_link", created.ID, nil, nil, map[string]any{
		"guardian_id": guardianID.String(),
		"child_id":    childID.String(),
	}); err != nil {
		writeInternalError(c)
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toGuardianChildLinkResponse(created))
}

func (h *peopleHandler) endGuardianChildLink(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	linkID, ok := parseUUIDParam(c, "link_id")
	if !ok {
		return
	}

	reason, ok := parseReasonPayload(c, "relationship_reason_required")
	if !ok {
		return
	}

	tx, err := h.pool.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		writeInternalError(c)
		return
	}
	defer tx.Rollback(c.Request.Context())

	row, found, err := h.fetchGuardianChildLinkByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, linkID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !found {
		writeError(c, http.StatusNotFound, "guardian_child_link_not_found", "Resource not found.", nil)
		return
	}

	if row.EndedAt == nil {
		const q = `
UPDATE guardian_child_links
SET ended_at = now(),
    ended_reason_code = $1,
    ended_reason_note = NULLIF($2, ''),
    updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5`
		if _, err := tx.Exec(c.Request.Context(), q, reason.Code, reason.Note, actor.TenantID, actor.BranchID, linkID); err != nil {
			writeInternalError(c)
			return
		}

		if err := h.insertAuditLogTx(c.Request.Context(), tx, actor, "guardian_child_link_ended", "guardian_child_link", linkID, &reason.Code, nullableReasonNote(reason.Note), map[string]any{}); err != nil {
			writeInternalError(c)
			return
		}
	}

	updated, found, err := h.fetchGuardianChildLinkByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, linkID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toGuardianChildLinkResponse(updated))
}

func (h *peopleHandler) createParentMapping(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	var req parentMappingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return
	}

	membershipID, err := uuid.Parse(strings.TrimSpace(req.MembershipID))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "membership_id"})
		return
	}
	guardianID, err := uuid.Parse(strings.TrimSpace(req.GuardianID))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "guardian_id"})
		return
	}

	tx, err := h.pool.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		writeInternalError(c)
		return
	}
	defer tx.Rollback(c.Request.Context())

	membership, mFound, err := h.fetchMembershipByIDForScope(c.Request.Context(), tx, actor.TenantID, actor.BranchID, membershipID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !mFound {
		writeError(c, http.StatusNotFound, "membership_not_found", "Resource not found.", nil)
		return
	}
	if membership.Role != "parent" {
		writeError(c, http.StatusBadRequest, "membership_not_parent", "Invalid request payload.", nil)
		return
	}
	if !membership.IsActive {
		writeError(c, http.StatusBadRequest, "membership_not_active", "Invalid request payload.", nil)
		return
	}

	guardianActive, gFound, err := h.fetchGuardianActive(c.Request.Context(), tx, actor.TenantID, actor.BranchID, guardianID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !gFound {
		writeError(c, http.StatusNotFound, "guardian_not_found", "Resource not found.", nil)
		return
	}
	if !guardianActive {
		writeError(c, http.StatusBadRequest, "guardian_not_active", "Invalid request payload.", nil)
		return
	}

	existingByMembership, hasMembershipMapping, err := h.fetchActiveParentMappingByMembership(c.Request.Context(), tx, actor.TenantID, actor.BranchID, membershipID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if hasMembershipMapping {
		if existingByMembership.GuardianID == guardianID {
			if err := tx.Commit(c.Request.Context()); err != nil {
				writeInternalError(c)
				return
			}
			c.JSON(http.StatusOK, toParentMembershipGuardianResponse(existingByMembership))
			return
		}
		writeError(c, http.StatusConflict, "parent_mapping_active_conflict", "Invalid request payload.", nil)
		return
	}

	createdID := newUUID()
	const insertQ = `
INSERT INTO parent_membership_guardians (id, tenant_id, branch_id, membership_id, guardian_id)
VALUES ($1, $2, $3, $4, $5)`

	if _, err := tx.Exec(c.Request.Context(), insertQ, createdID, actor.TenantID, actor.BranchID, membershipID, guardianID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			retryRow, retryFound, retryErr := h.fetchActiveParentMappingByMembership(c.Request.Context(), tx, actor.TenantID, actor.BranchID, membershipID)
			if retryErr != nil {
				writeInternalError(c)
				return
			}
			if retryFound {
				if retryRow.GuardianID == guardianID {
					if err := tx.Commit(c.Request.Context()); err != nil {
						writeInternalError(c)
						return
					}
					c.JSON(http.StatusOK, toParentMembershipGuardianResponse(retryRow))
					return
				}
				writeError(c, http.StatusConflict, "parent_mapping_active_conflict", "Invalid request payload.", nil)
				return
			}
		}
		writeInternalError(c)
		return
	}

	created, found, err := h.fetchParentMappingByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, createdID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := h.insertAuditLogTx(c.Request.Context(), tx, actor, "parent_mapping_created", "parent_membership_guardian_mapping", createdID, nil, nil, map[string]any{
		"membership_id": membershipID.String(),
		"guardian_id":   guardianID.String(),
	}); err != nil {
		writeInternalError(c)
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toParentMembershipGuardianResponse(created))
}

func (h *peopleHandler) endParentMapping(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	mappingID, ok := parseUUIDParam(c, "mapping_id")
	if !ok {
		return
	}

	reason, ok := parseReasonPayload(c, "relationship_reason_required")
	if !ok {
		return
	}

	tx, err := h.pool.BeginTx(c.Request.Context(), pgx.TxOptions{})
	if err != nil {
		writeInternalError(c)
		return
	}
	defer tx.Rollback(c.Request.Context())

	row, found, err := h.fetchParentMappingByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, mappingID)
	if err != nil {
		writeInternalError(c)
		return
	}
	if !found {
		writeError(c, http.StatusNotFound, "parent_mapping_not_found", "Resource not found.", nil)
		return
	}

	if row.EndedAt == nil {
		const q = `
UPDATE parent_membership_guardians
SET ended_at = now(),
    ended_reason_code = $1,
    ended_reason_note = NULLIF($2, ''),
    updated_at = now()
WHERE tenant_id = $3 AND branch_id = $4 AND id = $5`
		if _, err := tx.Exec(c.Request.Context(), q, reason.Code, reason.Note, actor.TenantID, actor.BranchID, mappingID); err != nil {
			writeInternalError(c)
			return
		}

		if err := h.insertAuditLogTx(c.Request.Context(), tx, actor, "parent_mapping_ended", "parent_membership_guardian_mapping", mappingID, &reason.Code, nullableReasonNote(reason.Note), map[string]any{}); err != nil {
			writeInternalError(c)
			return
		}
	}

	updated, found, err := h.fetchParentMappingByIDForUpdate(c.Request.Context(), tx, actor.TenantID, actor.BranchID, mappingID)
	if err != nil || !found {
		writeInternalError(c)
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, toParentMembershipGuardianResponse(updated))
}

func (h *peopleHandler) listAttendanceChildren(c *gin.Context) {
	actor, ok := actorFromRequest(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
		return
	}

	const q = `
SELECT c.id,
       c.full_name,
       EXISTS (
           SELECT 1
           FROM guardian_child_links gcl
           WHERE gcl.tenant_id = c.tenant_id
             AND gcl.branch_id = c.branch_id
             AND gcl.child_id = c.id
             AND gcl.ended_at IS NULL
       ) AS has_guardian_link
FROM children c
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.is_active = true
ORDER BY c.updated_at DESC`

	rows, err := h.pool.Query(c.Request.Context(), q, actor.TenantID, actor.BranchID)
	if err != nil {
		writeInternalError(c)
		return
	}
	defer rows.Close()

	out := make([]attendanceChildResponse, 0)
	for rows.Next() {
		var id uuid.UUID
		var fullName string
		var hasGuardian bool
		if err := rows.Scan(&id, &fullName, &hasGuardian); err != nil {
			writeInternalError(c)
			return
		}
		out = append(out, attendanceChildResponse{
			ID:                 id.String(),
			FullName:           fullName,
			EnrollmentComplete: hasGuardian,
		})
	}
	if err := rows.Err(); err != nil {
		writeInternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": out})
}

func (h *peopleHandler) fetchChildren(ctx context.Context, tenantID, branchID uuid.UUID, statusFilter string, limit, offset int) ([]childRow, error) {
	statusClause := "AND c.is_active = true"
	if statusFilter == statusFilterInactive {
		statusClause = "AND c.is_active = false"
	}
	if statusFilter == statusFilterAll {
		statusClause = ""
	}

	q := fmt.Sprintf(`
SELECT c.id,
       c.full_name,
       c.date_of_birth,
       c.start_date,
       c.end_date,
       c.core_hourly_rate_minor,
       c.notes,
       c.is_active,
       c.left_at,
       c.left_reason_code::text,
       c.left_reason_note,
       EXISTS (
           SELECT 1
           FROM guardian_child_links gcl
           WHERE gcl.tenant_id = c.tenant_id
             AND gcl.branch_id = c.branch_id
             AND gcl.child_id = c.id
             AND gcl.ended_at IS NULL
       ) AS has_guardian_link,
       c.created_at,
       c.updated_at
FROM children c
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  %s
ORDER BY c.updated_at DESC
LIMIT $3 OFFSET $4`, statusClause)

	rows, err := h.pool.Query(ctx, q, tenantID, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]childRow, 0)
	for rows.Next() {
		var row childRow
		if err := rows.Scan(
			&row.ID,
			&row.FullName,
			&row.DateOfBirth,
			&row.StartDate,
			&row.EndDate,
			&row.CoreHourlyRateMinor,
			&row.Notes,
			&row.IsActive,
			&row.LeftAt,
			&row.LeftReasonCode,
			&row.LeftReasonNote,
			&row.HasGuardianLink,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	return out, rows.Err()
}

func (h *peopleHandler) fetchChildByID(ctx context.Context, tenantID, branchID, childID uuid.UUID) (childRow, bool, error) {
	const q = `
SELECT c.id,
       c.full_name,
       c.date_of_birth,
       c.start_date,
       c.end_date,
       c.core_hourly_rate_minor,
       c.notes,
       c.is_active,
       c.left_at,
       c.left_reason_code::text,
       c.left_reason_note,
       EXISTS (
           SELECT 1
           FROM guardian_child_links gcl
           WHERE gcl.tenant_id = c.tenant_id
             AND gcl.branch_id = c.branch_id
             AND gcl.child_id = c.id
             AND gcl.ended_at IS NULL
       ) AS has_guardian_link,
       c.created_at,
       c.updated_at
FROM children c
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.id = $3`

	var row childRow
	err := h.pool.QueryRow(ctx, q, tenantID, branchID, childID).Scan(
		&row.ID,
		&row.FullName,
		&row.DateOfBirth,
		&row.StartDate,
		&row.EndDate,
		&row.CoreHourlyRateMinor,
		&row.Notes,
		&row.IsActive,
		&row.LeftAt,
		&row.LeftReasonCode,
		&row.LeftReasonNote,
		&row.HasGuardianLink,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return childRow{}, false, nil
	}
	if err != nil {
		return childRow{}, false, err
	}

	return row, true, nil
}

func (h *peopleHandler) fetchChildByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (childRow, bool, error) {
	const q = `
SELECT c.id,
       c.full_name,
       c.date_of_birth,
       c.start_date,
       c.end_date,
       c.core_hourly_rate_minor,
       c.notes,
       c.is_active,
       c.left_at,
       c.left_reason_code::text,
       c.left_reason_note,
       EXISTS (
           SELECT 1
           FROM guardian_child_links gcl
           WHERE gcl.tenant_id = c.tenant_id
             AND gcl.branch_id = c.branch_id
             AND gcl.child_id = c.id
             AND gcl.ended_at IS NULL
       ) AS has_guardian_link,
       c.created_at,
       c.updated_at
FROM children c
WHERE c.tenant_id = $1
  AND c.branch_id = $2
  AND c.id = $3
FOR UPDATE`

	var row childRow
	err := tx.QueryRow(ctx, q, tenantID, branchID, childID).Scan(
		&row.ID,
		&row.FullName,
		&row.DateOfBirth,
		&row.StartDate,
		&row.EndDate,
		&row.CoreHourlyRateMinor,
		&row.Notes,
		&row.IsActive,
		&row.LeftAt,
		&row.LeftReasonCode,
		&row.LeftReasonNote,
		&row.HasGuardianLink,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return childRow{}, false, nil
	}
	if err != nil {
		return childRow{}, false, err
	}

	return row, true, nil
}

func (h *peopleHandler) fetchGuardians(ctx context.Context, tenantID, branchID uuid.UUID, statusFilter string, limit, offset int) ([]guardianRow, error) {
	statusClause := "AND g.is_active = true"
	if statusFilter == statusFilterInactive {
		statusClause = "AND g.is_active = false"
	}
	if statusFilter == statusFilterAll {
		statusClause = ""
	}

	q := fmt.Sprintf(`
SELECT g.id,
       g.full_name,
       g.email,
       g.phone,
       g.notes,
       g.is_active,
       g.deactivated_at,
       g.deactivation_reason_code::text,
       g.deactivation_reason_note,
       g.created_at,
       g.updated_at
FROM guardians g
WHERE g.tenant_id = $1
  AND g.branch_id = $2
  %s
ORDER BY g.updated_at DESC
LIMIT $3 OFFSET $4`, statusClause)

	rows, err := h.pool.Query(ctx, q, tenantID, branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]guardianRow, 0)
	for rows.Next() {
		var row guardianRow
		if err := rows.Scan(
			&row.ID,
			&row.FullName,
			&row.Email,
			&row.Phone,
			&row.Notes,
			&row.IsActive,
			&row.DeactivatedAt,
			&row.DeactivationReasonCode,
			&row.DeactivationReasonNote,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	return out, rows.Err()
}

func (h *peopleHandler) fetchGuardianByID(ctx context.Context, tenantID, branchID, guardianID uuid.UUID) (guardianRow, bool, error) {
	const q = `
SELECT g.id,
       g.full_name,
       g.email,
       g.phone,
       g.notes,
       g.is_active,
       g.deactivated_at,
       g.deactivation_reason_code::text,
       g.deactivation_reason_note,
       g.created_at,
       g.updated_at
FROM guardians g
WHERE g.tenant_id = $1
  AND g.branch_id = $2
  AND g.id = $3`

	var row guardianRow
	err := h.pool.QueryRow(ctx, q, tenantID, branchID, guardianID).Scan(
		&row.ID,
		&row.FullName,
		&row.Email,
		&row.Phone,
		&row.Notes,
		&row.IsActive,
		&row.DeactivatedAt,
		&row.DeactivationReasonCode,
		&row.DeactivationReasonNote,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return guardianRow{}, false, nil
	}
	if err != nil {
		return guardianRow{}, false, err
	}

	return row, true, nil
}

func (h *peopleHandler) fetchGuardianByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID) (guardianRow, bool, error) {
	const q = `
SELECT g.id,
       g.full_name,
       g.email,
       g.phone,
       g.notes,
       g.is_active,
       g.deactivated_at,
       g.deactivation_reason_code::text,
       g.deactivation_reason_note,
       g.created_at,
       g.updated_at
FROM guardians g
WHERE g.tenant_id = $1
  AND g.branch_id = $2
  AND g.id = $3
FOR UPDATE`

	var row guardianRow
	err := tx.QueryRow(ctx, q, tenantID, branchID, guardianID).Scan(
		&row.ID,
		&row.FullName,
		&row.Email,
		&row.Phone,
		&row.Notes,
		&row.IsActive,
		&row.DeactivatedAt,
		&row.DeactivationReasonCode,
		&row.DeactivationReasonNote,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return guardianRow{}, false, nil
	}
	if err != nil {
		return guardianRow{}, false, err
	}

	return row, true, nil
}

func (h *peopleHandler) fetchGuardianActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID) (bool, bool, error) {
	const q = `
SELECT is_active
FROM guardians
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3`
	var isActive bool
	err := tx.QueryRow(ctx, q, tenantID, branchID, guardianID).Scan(&isActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return isActive, true, nil
}

func (h *peopleHandler) childExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1 FROM children WHERE tenant_id = $1 AND branch_id = $2 AND id = $3
)`
	var exists bool
	err := tx.QueryRow(ctx, q, tenantID, branchID, childID).Scan(&exists)
	return exists, err
}

func (h *peopleHandler) fetchActiveGuardianChildLinkForPair(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID, childID uuid.UUID) (guardianChildLinkRow, bool, error) {
	const q = `
SELECT id,
       guardian_id,
       child_id,
       ended_at,
       ended_reason_code::text,
       ended_reason_note,
       created_at,
       updated_at
FROM guardian_child_links
WHERE tenant_id = $1
  AND branch_id = $2
  AND guardian_id = $3
  AND child_id = $4
  AND ended_at IS NULL
LIMIT 1`

	var row guardianChildLinkRow
	err := tx.QueryRow(ctx, q, tenantID, branchID, guardianID, childID).Scan(
		&row.ID,
		&row.GuardianID,
		&row.ChildID,
		&row.EndedAt,
		&row.EndedReasonCode,
		&row.EndedReasonNote,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return guardianChildLinkRow{}, false, nil
	}
	if err != nil {
		return guardianChildLinkRow{}, false, err
	}

	return row, true, nil
}

func (h *peopleHandler) fetchGuardianChildLinkByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, linkID uuid.UUID) (guardianChildLinkRow, bool, error) {
	const q = `
SELECT id,
       guardian_id,
       child_id,
       ended_at,
       ended_reason_code::text,
       ended_reason_note,
       created_at,
       updated_at
FROM guardian_child_links
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE`

	var row guardianChildLinkRow
	err := tx.QueryRow(ctx, q, tenantID, branchID, linkID).Scan(
		&row.ID,
		&row.GuardianID,
		&row.ChildID,
		&row.EndedAt,
		&row.EndedReasonCode,
		&row.EndedReasonNote,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return guardianChildLinkRow{}, false, nil
	}
	if err != nil {
		return guardianChildLinkRow{}, false, err
	}

	return row, true, nil
}

type membershipRow struct {
	ID       uuid.UUID
	Role     string
	IsActive bool
}

func (h *peopleHandler) fetchMembershipByIDForScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (membershipRow, bool, error) {
	const q = `
SELECT id, role, is_active
FROM memberships
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3`

	var row membershipRow
	err := tx.QueryRow(ctx, q, tenantID, branchID, membershipID).Scan(&row.ID, &row.Role, &row.IsActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return membershipRow{}, false, nil
	}
	if err != nil {
		return membershipRow{}, false, err
	}

	return row, true, nil
}

func (h *peopleHandler) fetchActiveParentMappingByMembership(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (parentMembershipGuardianRow, bool, error) {
	const q = `
SELECT id,
       membership_id,
       guardian_id,
       ended_at,
       ended_reason_code::text,
       ended_reason_note,
       created_at,
       updated_at
FROM parent_membership_guardians
WHERE tenant_id = $1
  AND branch_id = $2
  AND membership_id = $3
  AND ended_at IS NULL
LIMIT 1`

	var row parentMembershipGuardianRow
	err := tx.QueryRow(ctx, q, tenantID, branchID, membershipID).Scan(
		&row.ID,
		&row.MembershipID,
		&row.GuardianID,
		&row.EndedAt,
		&row.EndedReasonCode,
		&row.EndedReasonNote,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return parentMembershipGuardianRow{}, false, nil
	}
	if err != nil {
		return parentMembershipGuardianRow{}, false, err
	}

	return row, true, nil
}

func (h *peopleHandler) fetchParentMappingByIDForUpdate(ctx context.Context, tx pgx.Tx, tenantID, branchID, mappingID uuid.UUID) (parentMembershipGuardianRow, bool, error) {
	const q = `
SELECT id,
       membership_id,
       guardian_id,
       ended_at,
       ended_reason_code::text,
       ended_reason_note,
       created_at,
       updated_at
FROM parent_membership_guardians
WHERE tenant_id = $1
  AND branch_id = $2
  AND id = $3
FOR UPDATE`

	var row parentMembershipGuardianRow
	err := tx.QueryRow(ctx, q, tenantID, branchID, mappingID).Scan(
		&row.ID,
		&row.MembershipID,
		&row.GuardianID,
		&row.EndedAt,
		&row.EndedReasonCode,
		&row.EndedReasonNote,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return parentMembershipGuardianRow{}, false, nil
	}
	if err != nil {
		return parentMembershipGuardianRow{}, false, err
	}

	return row, true, nil
}

func (h *peopleHandler) insertAuditLog(ctx context.Context, actor actorContext, actionType, entityType string, entityID uuid.UUID, reasonCode, reasonNote *string, details map[string]any) error {
	return h.insertAuditLogWithQuerier(ctx, h.pool, actor, actionType, entityType, entityID, reasonCode, reasonNote, details)
}

func (h *peopleHandler) insertAuditLogTx(ctx context.Context, tx pgx.Tx, actor actorContext, actionType, entityType string, entityID uuid.UUID, reasonCode, reasonNote *string, details map[string]any) error {
	return h.insertAuditLogWithQuerier(ctx, tx, actor, actionType, entityType, entityID, reasonCode, reasonNote, details)
}

type dbExecQuerier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func (h *peopleHandler) insertAuditLogWithQuerier(ctx context.Context, q dbExecQuerier, actor actorContext, actionType, entityType string, entityID uuid.UUID, reasonCode, reasonNote *string, details map[string]any) error {
	if details == nil {
		details = map[string]any{}
	}
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	const insertQ = `
INSERT INTO audit_logs (
    id,
    tenant_id,
    branch_id,
    actor_user_id,
    actor_membership_id,
    action_type,
    action_entity_type,
    action_entity_id,
    reason_code,
    reason_note,
    request_id,
    details
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11, $12::jsonb)`

	var reasonCodeValue any
	if reasonCode != nil {
		reasonCodeValue = *reasonCode
	}

	var reasonNoteValue any
	if reasonNote != nil {
		reasonNoteValue = *reasonNote
	}

	_, err = q.Exec(ctx, insertQ,
		newUUID(),
		actor.TenantID,
		actor.BranchID,
		actor.UserID,
		actor.MembershipID,
		actionType,
		entityType,
		entityID,
		reasonCodeValue,
		reasonNoteValue,
		actor.RequestID,
		string(detailsJSON),
	)
	return err
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(strings.TrimSpace(c.Param(name)))
	if err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": name})
		return uuid.UUID{}, false
	}
	return id, true
}

func parseDate(v string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(v))
}

func parseStatusFilter(c *gin.Context) (string, bool) {
	v := strings.TrimSpace(c.Query("status"))
	if v == "" {
		return statusFilterActive, true
	}
	switch v {
	case statusFilterActive, statusFilterInactive, statusFilterAll:
		return v, true
	default:
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "status"})
		return "", false
	}
}

func parsePagination(c *gin.Context) (int, int, bool) {
	limit := defaultListLimit
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > maxListLimit {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "limit"})
			return 0, 0, false
		}
		limit = parsed
	}

	offset := 0
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "offset"})
			return 0, 0, false
		}
		offset = parsed
	}

	return limit, offset, true
}

type parsedReason struct {
	Code string
	Note string
}

func parseReasonPayload(c *gin.Context, missingCode string) (parsedReason, bool) {
	var req reasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", err.Error())
		return parsedReason{}, false
	}

	req.ReasonCode = strings.TrimSpace(req.ReasonCode)
	req.ReasonNote = strings.TrimSpace(req.ReasonNote)

	if req.ReasonCode == "" {
		writeError(c, http.StatusBadRequest, missingCode, "Invalid request payload.", gin.H{"field": "reason_code"})
		return parsedReason{}, false
	}
	if _, ok := lifecycleReasonCodes[req.ReasonCode]; !ok {
		writeError(c, http.StatusBadRequest, "lifecycle_reason_invalid", "Invalid request payload.", gin.H{"field": "reason_code"})
		return parsedReason{}, false
	}
	if len(req.ReasonNote) > maxReasonNoteLen {
		writeError(c, http.StatusBadRequest, "validation_error", "Invalid request payload.", gin.H{"field": "reason_note"})
		return parsedReason{}, false
	}
	if req.ReasonCode == reasonCodeOther && req.ReasonNote == "" {
		writeError(c, http.StatusBadRequest, "reason_note_required_for_other", "Invalid request payload.", gin.H{"field": "reason_note"})
		return parsedReason{}, false
	}

	return parsedReason{Code: req.ReasonCode, Note: req.ReasonNote}, true
}

func nullableReasonNote(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func actorFromRequest(c *gin.Context) (actorContext, bool) {
	authCtx, ok := authContextFromContext(c)
	if !ok {
		return actorContext{}, false
	}

	userID, err := uuid.Parse(authCtx.UserID)
	if err != nil {
		return actorContext{}, false
	}
	membershipID, err := uuid.Parse(authCtx.MembershipID)
	if err != nil {
		return actorContext{}, false
	}
	tenantID, err := uuid.Parse(authCtx.TenantID)
	if err != nil {
		return actorContext{}, false
	}
	branchID, err := uuid.Parse(authCtx.BranchID)
	if err != nil {
		return actorContext{}, false
	}

	return actorContext{
		UserID:       userID,
		MembershipID: membershipID,
		TenantID:     tenantID,
		BranchID:     branchID,
		RequestID:    authCtx.RequestID,
	}, true
}

func toChildResponse(row childRow) childResponse {
	resp := childResponse{
		ID:                  row.ID.String(),
		FullName:            row.FullName,
		DateOfBirth:         formatDate(row.DateOfBirth),
		StartDate:           formatDate(row.StartDate),
		CoreHourlyRateMinor: row.CoreHourlyRateMinor,
		Notes:               row.Notes,
		IsActive:            row.IsActive,
		LeftReasonCode:      row.LeftReasonCode,
		LeftReasonNote:      row.LeftReasonNote,
		EnrollmentComplete:  row.HasGuardianLink,
		CreatedAt:           row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:           row.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if row.EndDate != nil {
		endDate := formatDate(*row.EndDate)
		resp.EndDate = &endDate
	}
	if row.LeftAt != nil {
		leftAt := row.LeftAt.UTC().Format(time.RFC3339)
		resp.LeftAt = &leftAt
	}

	missing := childMissingRequirements(row)
	resp.MissingRequirements = missing
	resp.EnrollmentComplete = len(missing) == 0

	return resp
}

func childMissingRequirements(row childRow) []string {
	missing := make([]string, 0)
	if strings.TrimSpace(row.FullName) == "" {
		missing = append(missing, "full_name")
	}
	if row.DateOfBirth.IsZero() {
		missing = append(missing, "date_of_birth")
	}
	if row.StartDate.IsZero() {
		missing = append(missing, "start_date")
	}
	if row.CoreHourlyRateMinor < 0 {
		missing = append(missing, "billing_rate")
	}
	if !row.HasGuardianLink {
		missing = append(missing, "guardian_link")
	}
	return missing
}

func toGuardianResponse(row guardianRow) guardianResponse {
	resp := guardianResponse{
		ID:                     row.ID.String(),
		FullName:               row.FullName,
		Email:                  row.Email,
		Phone:                  row.Phone,
		Notes:                  row.Notes,
		IsActive:               row.IsActive,
		DeactivationReasonCode: row.DeactivationReasonCode,
		DeactivationReasonNote: row.DeactivationReasonNote,
		CreatedAt:              row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:              row.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if row.DeactivatedAt != nil {
		v := row.DeactivatedAt.UTC().Format(time.RFC3339)
		resp.DeactivatedAt = &v
	}
	return resp
}

func toGuardianChildLinkResponse(row guardianChildLinkRow) guardianChildLinkResponse {
	resp := guardianChildLinkResponse{
		ID:              row.ID.String(),
		GuardianID:      row.GuardianID.String(),
		ChildID:         row.ChildID.String(),
		EndedReasonCode: row.EndedReasonCode,
		EndedReasonNote: row.EndedReasonNote,
		CreatedAt:       row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       row.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if row.EndedAt != nil {
		v := row.EndedAt.UTC().Format(time.RFC3339)
		resp.EndedAt = &v
	}
	return resp
}

func toParentMembershipGuardianResponse(row parentMembershipGuardianRow) parentMembershipGuardianResponse {
	resp := parentMembershipGuardianResponse{
		ID:              row.ID.String(),
		MembershipID:    row.MembershipID.String(),
		GuardianID:      row.GuardianID.String(),
		EndedReasonCode: row.EndedReasonCode,
		EndedReasonNote: row.EndedReasonNote,
		CreatedAt:       row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       row.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if row.EndedAt != nil {
		v := row.EndedAt.UTC().Format(time.RFC3339)
		resp.EndedAt = &v
	}
	return resp
}

func formatDate(v time.Time) string {
	return v.Format("2006-01-02")
}

func newUUID() uuid.UUID {
	if id, err := uuid.NewV7(); err == nil {
		return id
	}
	return uuid.New()
}
