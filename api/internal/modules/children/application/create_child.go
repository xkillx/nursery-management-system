package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/audit"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"
)

type CreateChildParams struct {
	FirstName      string
	MiddleName     string
	LastName       string
	DateOfBirth    string
	StartDate      string
	EndDate        string
	Notes          string
	PrimaryRoomID  string
}

type CreateChild struct {
	repo  domain.Repository
	audit *audit.Writer
	pool  *pgxpool.Pool
}

func NewCreateChild(repo domain.Repository, auditWriter *audit.Writer, pool *pgxpool.Pool) *CreateChild {
	return &CreateChild{repo: repo, audit: auditWriter, pool: pool}
}

func (uc *CreateChild) Execute(ctx context.Context, actor tenant.ActorContext, params CreateChildParams) (domain.Child, error) {
	firstName := strings.TrimSpace(params.FirstName)
	if firstName == "" {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "first_name")
	}
	middleName := strings.TrimSpace(params.MiddleName)
	lastName := strings.TrimSpace(params.LastName)

	dob, err := parseDate(params.DateOfBirth)
	if err != nil {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "date_of_birth")
	}

	startDate, err := parseDate(params.StartDate)
	if err != nil {
		return domain.Child{}, domainerrors.Validation("Invalid request payload.", "start_date")
	}

	var endDate *time.Time
	if strings.TrimSpace(params.EndDate) != "" {
		parsed, parseErr := parseDate(params.EndDate)
		if parseErr != nil {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "end_date")
		}
		if parsed.Before(startDate) {
			return domain.Child{}, domainerrors.Validation("Invalid request payload.", "end_date")
		}
		endDate = &parsed
	}

	notes := strings.TrimSpace(params.Notes)

	primaryRoomID, err := parsePrimaryRoomID(params.PrimaryRoomID)
	if err != nil {
		return domain.Child{}, err
	}

	child := domain.Child{
		ID:            uid.NewUUID(),
		FirstName:     firstName,
		DateOfBirth:   dob,
		StartDate:     startDate,
		EndDate:       endDate,
		IsActive:      true,
		PrimaryRoomID: primaryRoomID,
	}
	if middleName != "" {
		child.MiddleName = &middleName
	}
	if lastName != "" {
		child.LastName = &lastName
	}

	if err := uc.repo.Create(ctx, child, notes, actor.TenantID, actor.BranchID); err != nil {
		return domain.Child{}, domainerrors.Internal(fmt.Errorf("create child: %w", err))
	}

	created, found, err := uc.repo.GetByID(ctx, actor.TenantID, actor.BranchID, child.ID)
	if err != nil || !found {
		return domain.Child{}, domainerrors.Internal(fmt.Errorf("fetch created child: %w", err))
	}

	if err := uc.audit.Write(ctx, uc.pool, actor, audit.WriteParams{
		ActionType: "child_created",
		EntityType: "child",
		EntityID:   child.ID,
		Details:    map[string]any{},
	}); err != nil {
		return domain.Child{}, domainerrors.Internal(fmt.Errorf("audit child_created: %w", err))
	}

	return created, nil
}

func parseDate(v string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(v))
}

func parsePrimaryRoomID(v string) (*uuid.UUID, error) {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" {
		return nil, domainerrors.Validation("Invalid request payload.", "primary_room_id")
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, domainerrors.Validation("Invalid request payload.", "primary_room_id")
	}
	return &id, nil
}

// parseDatePtr is unexported; kept for potential internal reuse.
func parseDatePtr(v string) (*time.Time, error) {
	if strings.TrimSpace(v) == "" {
		return nil, nil
	}
	t, err := parseDate(v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ValidateReasonCode checks a reason code and note for lifecycle actions.
func ValidateReasonCode(code, note string) error {
	code = strings.TrimSpace(code)
	note = strings.TrimSpace(note)

	if code == "" {
		return domainerrors.New("child_lifecycle_reason_required", "Invalid request payload.", "reason_code")
	}

	if _, ok := domain.ValidReasonCodes[domain.ReasonCode(code)]; !ok {
		return domainerrors.New("lifecycle_reason_invalid", "Invalid request payload.", "reason_code")
	}

	if len(note) > maxReasonNoteLen {
		return domainerrors.Validation("Invalid request payload.", "reason_note")
	}

	if code == string(domain.ReasonOther) && note == "" {
		return domainerrors.New("reason_note_required_for_other", "Invalid request payload.", "reason_note")
	}

	return nil
}

const (
	maxReasonNoteLen = 500
	defaultListLimit = 50
	maxListLimit     = 200
)

// ValidatePagination validates and normalizes pagination parameters.
func ValidatePagination(limit, offset int) (int, int, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		return 0, 0, domainerrors.Validation("Invalid request payload.", "limit")
	}
	if offset < 0 {
		return 0, 0, domainerrors.Validation("Invalid request payload.", "offset")
	}
	return limit, offset, nil
}

// ValidateStatusFilter returns a valid StatusFilter or a validation error.
func ValidateStatusFilter(v string) (domain.StatusFilter, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return domain.StatusActive, nil
	}
	switch domain.StatusFilter(v) {
	case domain.StatusActive, domain.StatusInactive, domain.StatusAll:
		return domain.StatusFilter(v), nil
	default:
		return "", domainerrors.Validation("Invalid request payload.", "status")
	}
}

// parseUUID validates and returns a uuid.UUID from a string.
func parseUUID(v string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(v))
}
