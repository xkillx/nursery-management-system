package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	absencedomain "nursery-management-system/api/internal/modules/absence/domain"
	postgresabsence "nursery-management-system/api/internal/modules/absence/infrastructure/postgres"
	attendancedomain "nursery-management-system/api/internal/modules/attendance/domain"
	childapp "nursery-management-system/api/internal/modules/children/application"
	childdomain "nursery-management-system/api/internal/modules/children/domain"
	postgreschild "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	postgresguardian "nursery-management-system/api/internal/modules/guardians/infrastructure/postgres"
	invitetokens "nursery-management-system/api/internal/modules/invites/infrastructure/tokens"
	ownerdomain "nursery-management-system/api/internal/modules/owner/domain"
	ownerpostgres "nursery-management-system/api/internal/modules/owner/infrastructure/postgres"
	"nursery-management-system/api/internal/modules/parentmappings/domain"
	postgresparent "nursery-management-system/api/internal/modules/parentmappings/infrastructure/postgres"
	sessiontypepostgres "nursery-management-system/api/internal/modules/sessiontypes/infrastructure/postgres"
	termapp "nursery-management-system/api/internal/modules/term/application"
	"nursery-management-system/api/internal/platform/email"
	"nursery-management-system/api/internal/platform/tenant"
)

type guardianCheckerAdapter struct {
	repo *postgresguardian.GuardianRepository
}

func (a *guardianCheckerAdapter) IsActive(ctx context.Context, tx pgx.Tx, tenantID, branchID, guardianID uuid.UUID) (bool, bool, error) {
	return a.repo.GetActive(ctx, tx, tenantID, branchID, guardianID)
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

type childCreatorAdapter struct{}

var _ = (*childCreatorAdapter)(nil)

// ── Rooms adapters ──────────────────────────────────────────────────────────

type siteExistsCheckerAdapter struct {
	repo *ownerpostgres.OwnerRepository
}

func (a *siteExistsCheckerAdapter) SiteExists(ctx context.Context, tenantID, siteID uuid.UUID) (bool, error) {
	_, err := a.repo.GetActiveSite(ctx, tenantID, siteID)
	if err != nil {
		if err == ownerdomain.ErrSiteNotFound {
			return false, nil
		}
		return false, fmt.Errorf("check site exists: %w", err)
	}
	return true, nil
}

var (
	_ ownerdomain.InviteTokenGenerator = (*ownerInviteTokenAdapter)(nil)
	_ ownerdomain.ManagerInviteSender  = (*ownerEmailSenderAdapter)(nil)
)

// ── Session types adapters ───────────────────────────────────────────────

type sessionTypeLookupAdapter struct {
	repo *sessiontypepostgres.SessionTypeRepository
}

// GetActiveInScope delegates to repo.GetByID. Active/inactive is reported via
// the IsActive flag on SessionTypeInfo; the application layer enforces the
// "must be active" rule.
func (a *sessionTypeLookupAdapter) GetActiveInScope(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (childapp.SessionTypeInfo, bool, error) {
	st, err := a.repo.GetByID(ctx, tenantID, branchID, sessionTypeID)
	if err != nil {
		return childapp.SessionTypeInfo{}, false, fmt.Errorf("sessiontype lookup: %w", err)
	}
	return childapp.SessionTypeInfo{
		ID:           st.ID,
		Name:         st.Name,
		StartMinutes: st.StartMinutes,
		EndMinutes:   st.EndMinutes,
		IsActive:     st.IsActive,
	}, true, nil
}

var _ childapp.SessionTypeLookup = (*sessionTypeLookupAdapter)(nil)

// ── Term module adapters ──────────────────────────────────────────────────

// bookingPatternLookupAdapter satisfies termapp.BookingPatternLookup by delegating
// to the children module's child_booking_patterns lookup.
type bookingPatternLookupAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *bookingPatternLookupAdapter) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, patternID uuid.UUID) (bool, error) {
	_, found, err := a.repo.GetPatternByID(ctx, tenantID, branchID, patternID)
	if err != nil {
		return false, fmt.Errorf("booking pattern lookup: %w", err)
	}
	return found, nil
}

var _ termapp.BookingPatternLookup = (*bookingPatternLookupAdapter)(nil)

// siteRateProviderAdapter returns the branch's core_hourly_rate_minor (snapshotted
// at term creation).
type siteRateProviderAdapter struct {
	repo *ownerpostgres.OwnerRepository
}

func (a *siteRateProviderAdapter) SiteHourlyRateMinor(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID) (int, bool, error) {
	site, err := a.repo.GetActiveSite(ctx, tenantID, branchID)
	if err != nil {
		if err == ownerdomain.ErrSiteNotFound {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("site rate lookup: %w", err)
	}
	if site.CoreHourlyRateMinor == nil || *site.CoreHourlyRateMinor <= 0 {
		return 0, false, nil
	}
	return *site.CoreHourlyRateMinor, true, nil
}

var _ termapp.SiteRateProvider = (*siteRateProviderAdapter)(nil)

// childDeactivatorAdapter satisfies termapp.ChildDeactivator by delegating
// to the children.MarkInactive use case. The reason code/note come from
// the term module (e.g., "term ended without renewal").
type childDeactivatorAdapter struct {
	markInactiveUC *childapp.MarkInactive
}

func (a *childDeactivatorAdapter) MarkChildInactive(ctx context.Context, tenantID, branchID, childID uuid.UUID, reasonCode, reasonNote string) error {
	actor := tenant.ActorContext{
		TenantID:      tenantID,
		BranchID:      branchID,
		RequestID:     "system:expire_terms",
		CorrelationID: "system:expire_terms",
	}
	_, err := a.markInactiveUC.Execute(ctx, actor, childID.String(), childapp.MarkInactiveParams{
		ReasonCode: reasonCode,
		ReasonNote: reasonNote,
	})
	return err
}

var _ termapp.ChildDeactivator = (*childDeactivatorAdapter)(nil)
