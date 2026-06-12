package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	absencedomain "nursery-management-system/api/internal/modules/absence/domain"
	postgresabsence "nursery-management-system/api/internal/modules/absence/infrastructure/postgres"
	attendancedomain "nursery-management-system/api/internal/modules/attendance/domain"
	childdomain "nursery-management-system/api/internal/modules/children/domain"
	postgreschild "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	postgresguardian "nursery-management-system/api/internal/modules/guardians/infrastructure/postgres"
	invitetokens "nursery-management-system/api/internal/modules/invites/infrastructure/tokens"
	ownerdomain "nursery-management-system/api/internal/modules/owner/domain"
	"nursery-management-system/api/internal/modules/parentmappings/domain"
	postgresparent "nursery-management-system/api/internal/modules/parentmappings/infrastructure/postgres"
	regprofiledomain "nursery-management-system/api/internal/modules/registrationprofiles/domain"
	"nursery-management-system/api/internal/platform/db/sqlc"
	"nursery-management-system/api/internal/platform/email"
	"nursery-management-system/api/internal/platform/uid"
)

type guardianCheckerAdapter struct {
	repo *postgresguardian.GuardianRepository
}

func (a *guardianCheckerAdapter) IsActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID) (bool, bool, error) {
	return a.repo.GetActive(ctx, tx, tenantID, branchID, guardianID)
}

type childCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childCheckerAdapter) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error) {
	return a.repo.ExistsInScope(ctx, tx, tenantID, branchID, childID)
}

type membershipCheckerAdapter struct {
	repo *postgresparent.ParentMappingRepository
}

func (a *membershipCheckerAdapter) GetForScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, membershipID uuid.UUID) (domain.MembershipInfo, bool, error) {
	return a.repo.GetMembershipForScope(ctx, tx, tenantID, branchID, membershipID)
}

type childEnrollmentCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childEnrollmentCheckerAdapter) CheckEnrollmentForAttendance(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) error {
	child, found, err := a.repo.GetForAttendanceCheck(ctx, tx, tenantID, branchID, childID)
	if err != nil {
		return fmt.Errorf("check child enrollment: %w", err)
	}
	if !found {
		return attendancedomain.ErrChildNotFound
	}
	if !child.IsActive {
		return attendancedomain.ErrChildNotFound
	}
	if !child.EnrollmentComplete() {
		return attendancedomain.ErrChildEnrollmentIncomplete
	}
	if localDate.Before(child.StartDate) {
		return attendancedomain.ErrChildEnrollmentIncomplete
	}
	if child.EndDate != nil && localDate.After(*child.EndDate) {
		return attendancedomain.ErrChildEnrollmentIncomplete
	}
	return nil
}

// Ensure adapter satisfies the interface at compile time.
var _ childdomain.Repository = (*postgreschild.ChildRepository)(nil)

type childCorrectionCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childCorrectionCheckerAdapter) GetChildForCorrection(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (attendancedomain.ChildCorrectionInfo, bool, error) {
	info, found, err := a.repo.GetChildForCorrection(ctx, tx, tenantID, branchID, childID)
	if err != nil {
		return attendancedomain.ChildCorrectionInfo{}, false, err
	}
	if !found {
		return attendancedomain.ChildCorrectionInfo{}, false, nil
	}
	return attendancedomain.ChildCorrectionInfo{
		ID:        info.ID,
		StartDate: info.StartDate,
		EndDate:   info.EndDate,
	}, true, nil
}

type absenceMarkerCheckerAdapter struct {
	repo *postgresabsence.AbsenceRepository
}

func (a *absenceMarkerCheckerAdapter) HasActiveAbsenceMarker(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID, localDate time.Time) (bool, error) {
	_, found, err := a.repo.FindActiveByChildDate(ctx, tx, tenantID, branchID, childID, localDate)
	if err != nil {
		return false, fmt.Errorf("check active absence marker: %w", err)
	}
	return found, nil
}

var _ absencedomain.Repository = (*postgresabsence.AbsenceRepository)(nil)

// ── Owner adapters ──────────────────────────────────────────────────────────

type ownerInviteTokenAdapter struct {
	gen *invitetokens.Manager
}

func (a *ownerInviteTokenAdapter) Generate() (string, string, time.Time, error) {
	tok, err := a.gen.Generate()
	if err != nil {
		return "", "", time.Time{}, err
	}
	return tok.Raw, tok.Hash, tok.ExpiresAt, nil
}

type ownerEmailSenderAdapter struct {
	sender  email.Sender
	baseURL string
}

func (a *ownerEmailSenderAdapter) SendManagerInvite(ctx context.Context, toEmail, acceptURL string) error {
	msg := email.Message{
		To:      toEmail,
		Subject: "You're invited to join as manager",
		Text: fmt.Sprintf(
			"You have been invited to join as a manager.\n\nClick the link below to accept:\n%s\n\nThis invitation expires in 7 days.",
			acceptURL,
		),
	}
	return a.sender.Send(ctx, msg)
}

type childCreatorAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childCreatorAdapter) CreateChild(ctx context.Context, tx pgx.Tx, child regprofiledomain.ChildInfo, tenantID, branchID uuid.UUID) (regprofiledomain.ChildCreationResult, error) {
	childID := uid.NewUUID()
	q := sqlc.New(tx)
	err := q.ChildrenCreate(ctx, sqlc.ChildrenCreateParams{
		ID:                  pgtype.UUID{Bytes: [16]byte(childID), Valid: true},
		TenantID:            pgtype.UUID{Bytes: [16]byte(tenantID), Valid: true},
		BranchID:            pgtype.UUID{Bytes: [16]byte(branchID), Valid: true},
		FullName:            child.FullName,
		DateOfBirth:         pgtype.Date{Time: child.DateOfBirth, Valid: true},
		StartDate:           pgtype.Date{Time: child.StartDate, Valid: true},
		CoreHourlyRateMinor: pgtype.Int4{Valid: false},
		Column9:             child.Notes,
	})
	if err != nil {
		return regprofiledomain.ChildCreationResult{}, fmt.Errorf("create child: %w", err)
	}

	return regprofiledomain.ChildCreationResult{
		ID:        childID,
		FullName:  child.FullName,
		StartDate: child.StartDate,
	}, nil
}

var (
	_ regprofiledomain.ChildCreator = (*childCreatorAdapter)(nil)

	_ ownerdomain.InviteTokenGenerator = (*ownerInviteTokenAdapter)(nil)
	_ ownerdomain.ManagerInviteSender  = (*ownerEmailSenderAdapter)(nil)
)
