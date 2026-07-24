package bootstrap

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	absencedomain "nursery-management-system/api/internal/modules/absence/domain"
	postgresabsence "nursery-management-system/api/internal/modules/absence/infrastructure/postgres"
	attendanceapp "nursery-management-system/api/internal/modules/attendance/application"
	attendancedomain "nursery-management-system/api/internal/modules/attendance/domain"
	billingapp "nursery-management-system/api/internal/modules/billing/application"
	billingdomain "nursery-management-system/api/internal/modules/billing/domain"
	billingpostgres "nursery-management-system/api/internal/modules/billing/infrastructure/postgres"
	bookingsapp "nursery-management-system/api/internal/modules/bookings/application"
	bookingsdomain "nursery-management-system/api/internal/modules/bookings/domain"
	branchclosurepostgres "nursery-management-system/api/internal/modules/branch_closures/infrastructure/postgres"
	childapp "nursery-management-system/api/internal/modules/children/application"
	childdomain "nursery-management-system/api/internal/modules/children/domain"
	postgreschild "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	fundingapp "nursery-management-system/api/internal/modules/funding/application"
	fundingdomain "nursery-management-system/api/internal/modules/funding/domain"
	fundingpostgres "nursery-management-system/api/internal/modules/funding/infrastructure/postgres"
	hourlypostgres "nursery-management-system/api/internal/modules/hourly_bookings/infrastructure/postgres"
	invitetokens "nursery-management-system/api/internal/modules/invites/infrastructure/tokens"
	notificationsapp "nursery-management-system/api/internal/modules/notifications/application"
	ownerdomain "nursery-management-system/api/internal/modules/owner/domain"
	ownerpostgres "nursery-management-system/api/internal/modules/owner/infrastructure/postgres"
	parentchildapp "nursery-management-system/api/internal/modules/parentchildmappings/application"
	parentchilddomain "nursery-management-system/api/internal/modules/parentchildmappings/domain"
	parentchildpostgres "nursery-management-system/api/internal/modules/parentchildmappings/infrastructure/postgres"
	parentsapp "nursery-management-system/api/internal/modules/parents/application"
	parentsdomain "nursery-management-system/api/internal/modules/parents/domain"
	roomspostgres "nursery-management-system/api/internal/modules/rooms/infrastructure/postgres"
	sessiontemplateapp "nursery-management-system/api/internal/modules/sessiontemplates/application"
	sessiontypepostgres "nursery-management-system/api/internal/modules/sessiontypes/infrastructure/postgres"
	siteprofileapp "nursery-management-system/api/internal/modules/siteprofile/application"
	siteprofiledomain "nursery-management-system/api/internal/modules/siteprofile/domain"
	termapp "nursery-management-system/api/internal/modules/term/application"
	termcalendarpostgres "nursery-management-system/api/internal/modules/term_calendar/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/db/sqlc"
	"nursery-management-system/api/internal/platform/email"
	domainerrors "nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
	"nursery-management-system/api/internal/platform/uid"

	termdomain "nursery-management-system/api/internal/modules/term/domain"
	termpostgres "nursery-management-system/api/internal/modules/term/infrastructure/postgres"
)

type membershipCheckerAdapter struct {
	repo *parentchildpostgres.ParentChildMappingRepository
}

func (a *membershipCheckerAdapter) GetForScope(ctx context.Context, tx any, tenantID, branchID, membershipID uuid.UUID) (parentchilddomain.MembershipInfo, bool, error) {
	return a.repo.GetMembershipForScope(ctx, tx, tenantID, branchID, membershipID)
}

type childScopeCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childScopeCheckerAdapter) ExistsInScope(ctx context.Context, tx pgx.Tx, tenantID, branchID, childID uuid.UUID) (bool, error) {
	return a.repo.ExistsInScope(ctx, tx, tenantID, branchID, childID)
}

var _ parentchildapp.ChildChecker = (*childScopeCheckerAdapter)(nil)

type parentsChildExistenceCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *parentsChildExistenceCheckerAdapter) ExistsInScope(ctx context.Context, tx parentsdomain.Tx, tenantID, branchID, childID uuid.UUID) (bool, error) {
	if tx != nil {
		return a.repo.ExistsInScope(ctx, tx.(pgx.Tx), tenantID, branchID, childID)
	}
	return a.repo.ExistsInScope(ctx, nil, tenantID, branchID, childID)
}

var _ parentsdomain.ChildExistenceChecker = (*parentsChildExistenceCheckerAdapter)(nil)

// parentUserCreatorAdapter creates a user account and membership for a parent portal invite.
type parentUserCreatorAdapter struct {
	pool *pgxpool.Pool
}

func (a *parentUserCreatorAdapter) CreateParentUser(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, emailAddr string) (uuid.UUID, error) {
	q := sqlc.New(tx)
	userID := uid.NewUUID()
	emailNormalized := strings.ToLower(strings.TrimSpace(emailAddr))

	_, err := q.InviteCreateUser(ctx, sqlc.InviteCreateUserParams{
		ID:              pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		Email:           emailAddr,
		EmailNormalized: emailNormalized,
		PasswordHash:    "",
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create parent user: %w", err)
	}

	err = q.InviteCreateMembership(ctx, sqlc.InviteCreateMembershipParams{
		ID:       pgtype.UUID{Bytes: [16]byte(uid.NewUUID()), Valid: true},
		TenantID: pgtype.UUID{Bytes: [16]byte(tenantID), Valid: true},
		BranchID: pgtype.UUID{Bytes: [16]byte(branchID), Valid: true},
		UserID:   pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		Role:     "parent",
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create parent membership: %w", err)
	}

	return userID, nil
}

var _ parentsapp.UserCreator = (*parentUserCreatorAdapter)(nil)

// parentEmailSenderAdapter sends portal invite emails to parents.
type parentEmailSenderAdapter struct {
	sender  email.Sender
	baseURL string
}

func (a *parentEmailSenderAdapter) SendParentPortalInvite(ctx context.Context, toEmail, acceptURL string) error {
	msg := email.Message{
		To:      toEmail,
		Subject: "You're invited to access the parent portal",
		Text: "You have been invited to access the parent portal.\n\n" +
			"Click the link below to set up your account:\n" + acceptURL + "\n\n" +
			"This invitation expires in 7 days.",
	}
	return a.sender.Send(ctx, msg)
}

var _ parentsapp.EmailSender = (*parentEmailSenderAdapter)(nil)

type childEnrollmentCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childEnrollmentCheckerAdapter) CheckEnrollmentForAttendance(ctx context.Context, tx any, tenantID, branchID, childID uuid.UUID, localDate time.Time) error {
	child, found, err := a.repo.GetForAttendanceCheck(ctx, tx, tenantID, branchID, childID)
	if err != nil {
		return fmt.Errorf("check child enrollment: %w", err)
	}
	if !found {
		return attendancedomain.ErrChildNotFound
	}
	if !child.EnrollmentComplete() {
		return attendancedomain.ErrChildEnrollmentIncomplete
	}
	if !child.IsEligibleForAttendance(localDate) {
		return attendancedomain.ErrChildNotFound
	}
	return nil
}

// Ensure adapter satisfies the interface at compile time.
var _ childdomain.Repository = (*postgreschild.ChildRepository)(nil)

type childCorrectionCheckerAdapter struct {
	repo *postgreschild.ChildRepository
}

func (a *childCorrectionCheckerAdapter) GetChildForCorrection(ctx context.Context, tx any, tenantID, branchID, childID uuid.UUID) (attendancedomain.ChildCorrectionInfo, bool, error) {
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

func (a *absenceMarkerCheckerAdapter) HasActiveAbsenceMarker(ctx context.Context, tx any, tenantID, branchID, childID uuid.UUID, localDate time.Time) (bool, error) {
	_, found, err := a.repo.FindActiveByChildDate(ctx, tx, tenantID, branchID, childID, localDate)
	if err != nil {
		return false, fmt.Errorf("check active absence marker: %w", err)
	}
	return found, nil
}

var _ absencedomain.Repository = (*postgresabsence.AbsenceRepository)(nil)

// siteRateUpdateAdapter wraps ownerpostgres.OwnerRepository as billingdomain.SiteRateRepository.
// This avoids duplicating the SQL update in the billing postgres repo (KTD-2).
type siteRateUpdateAdapter struct {
	repo *ownerpostgres.OwnerRepository
}

func (a *siteRateUpdateAdapter) GetCoreHourlyRate(ctx context.Context, tenantID, branchID uuid.UUID) (int, bool, error) {
	site, err := a.repo.GetActiveSite(ctx, tenantID, branchID)
	if err != nil {
		if err == ownerdomain.ErrSiteNotFound {
			return 0, false, nil
		}
		return 0, false, err
	}
	if site.CoreHourlyRateMinor == nil {
		return 0, false, nil
	}
	return *site.CoreHourlyRateMinor, true, nil
}

func (a *siteRateUpdateAdapter) UpdateCoreHourlyRate(ctx context.Context, tx billingdomain.Tx, tenantID, branchID uuid.UUID, rateMinor int) error {
	prev, _, err := a.repo.UpdateSiteCoreHourlyRate(ctx, tx, tenantID, branchID, rateMinor)
	if err != nil {
		return err
	}
	_ = prev
	return nil
}

var _ billingdomain.SiteRateRepository = (*siteRateUpdateAdapter)(nil)

// ── Parent Contact adapter (reads from parents table via parent_children) ──

type parentContactLookupAdapter struct {
	pool *pgxpool.Pool
}

func (a *parentContactLookupAdapter) GetForInvoice(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*billingdomain.ParentContact, error) {
	var fullName, email string
	var phone pgtype.Text
	var addrLine1, addrLine2, addrCity, addrPostcode pgtype.Text

	err := a.pool.QueryRow(ctx, `
		SELECT
			p.first_name || COALESCE(' ' || p.last_name, ''),
			COALESCE(p.email, ''),
			p.phone,
			p.address_line1,
			p.address_line2,
			p.address_city,
			p.address_postcode
		FROM parent_children pc
		JOIN parents p ON p.id = pc.parent_id
		WHERE pc.tenant_id = $1
		  AND pc.branch_id = $2
		  AND pc.child_id = $3
		  AND pc.ended_at IS NULL
		  AND p.is_active = true
		ORDER BY pc.created_at ASC
		LIMIT 1
	`, tenantID, branchID, childID).Scan(&fullName, &email, &phone, &addrLine1, &addrLine2, &addrCity, &addrPostcode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query parent contact: %w", err)
	}

	pc := &billingdomain.ParentContact{
		FullName:  fullName,
		Email:     email,
		Telephone: pgtypeTextVal(phone),
	}
	if addrLine1.Valid {
		pc.AddressLine1 = addrLine1.String
	}
	if addrLine2.Valid {
		pc.AddressLine2 = addrLine2.String
	}
	if addrCity.Valid {
		pc.AddressCity = addrCity.String
	}
	if addrPostcode.Valid {
		pc.AddressPostcode = addrPostcode.String
	}

	return pc, nil
}

func pgtypeTextVal(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

var _ billingapp.ParentContactLookup = (*parentContactLookupAdapter)(nil)

// ── Parent Child Lookup adapter ─────────────────────────────────────────

// parentChildLookupAdapter satisfies bookingsapp.ParentChildLookup by
// delegating to the parentchildmappings module's repository.
type parentChildLookupAdapter struct {
	repo *parentchildpostgres.ParentChildMappingRepository
}

func (a *parentChildLookupAdapter) ListChildIDsForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID) ([]uuid.UUID, error) {
	mappings, err := a.repo.ListActiveByMembership(ctx, nil, tenantID, branchID, membershipID)
	if err != nil {
		return nil, fmt.Errorf("parent child lookup: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(mappings))
	for _, m := range mappings {
		ids = append(ids, m.ChildID)
	}
	return ids, nil
}

var _ bookingsapp.ParentChildLookup = (*parentChildLookupAdapter)(nil)

// parentChildLookupForFundingAdapter satisfies fundingapp.ParentChildLookupForFunding.
type parentChildLookupForFundingAdapter struct {
	repo *parentchildpostgres.ParentChildMappingRepository
}

func (a *parentChildLookupForFundingAdapter) ListChildIDsForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID) ([]uuid.UUID, error) {
	mappings, err := a.repo.ListActiveByMembership(ctx, nil, tenantID, branchID, membershipID)
	if err != nil {
		return nil, fmt.Errorf("parent child lookup: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(mappings))
	for _, m := range mappings {
		ids = append(ids, m.ChildID)
	}
	return ids, nil
}

var _ fundingapp.ParentChildLookupForFunding = (*parentChildLookupForFundingAdapter)(nil)

// parentChildLookupForAttendanceAdapter satisfies attendanceapp.ParentChildLookupForAttendance.
type parentChildLookupForAttendanceAdapter struct {
	repo *parentchildpostgres.ParentChildMappingRepository
}

func (a *parentChildLookupForAttendanceAdapter) ListChildIDsForParent(ctx context.Context, tenantID, branchID, membershipID uuid.UUID) ([]uuid.UUID, error) {
	mappings, err := a.repo.ListActiveByMembership(ctx, nil, tenantID, branchID, membershipID)
	if err != nil {
		return nil, fmt.Errorf("parent child lookup: %w", err)
	}
	ids := make([]uuid.UUID, 0, len(mappings))
	for _, m := range mappings {
		ids = append(ids, m.ChildID)
	}
	return ids, nil
}

var _ attendanceapp.ParentChildLookupForAttendance = (*parentChildLookupForAttendanceAdapter)(nil)

// ── Site Profile adapter ──────────────────────────────────────────────────

type siteProfileLookupAdapter struct {
	getUC *siteprofileapp.GetSiteProfileUseCase
}

func (a *siteProfileLookupAdapter) GetForInvoice(ctx context.Context, tenantID, branchID uuid.UUID) (*siteprofiledomain.SiteProfile, error) {
	profile, err := a.getUC.Execute(ctx, tenant.ActorContext{
		TenantID: tenantID,
		BranchID: branchID,
	})
	if err != nil {
		if domainerrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return profile, nil
}

// ── Owner adapters ──────────────────────────────────────────────────────────

type inviteTokenGeneratorAdapter struct {
	gen *invitetokens.Manager
}

func (a *inviteTokenGeneratorAdapter) Generate() (string, string, time.Time, error) {
	tok, err := a.gen.Generate()
	if err != nil {
		return "", "", time.Time{}, err
	}
	return tok.Raw, tok.Hash, tok.ExpiresAt, nil
}

type emailSenderAdapter struct {
	sender  email.Sender
	baseURL string
}

func (a *emailSenderAdapter) SendManagerInvite(ctx context.Context, toEmail, acceptURL string) error {
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
	_ ownerdomain.InviteTokenGenerator = (*inviteTokenGeneratorAdapter)(nil)
	_ ownerdomain.ManagerInviteSender  = (*emailSenderAdapter)(nil)
)

// ── Session types adapters ───────────────────────────────────────────────

type sessionTypeLookupAdapter struct {
	repo *sessiontypepostgres.SessionTypeRepository
}

// GetActiveInScope delegates to repo.GetByID. Active/inactive is reported via
// the IsActive flag on SessionTypeInfo; the application layer enforces the
// "must be active" rule.
func (a *sessionTypeLookupAdapter) GetActiveInScope(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (childapp.SessionTypeInfo, bool, error) {
	info, found, err := a.lookup(ctx, tenantID, branchID, sessionTypeID)
	if err != nil {
		return childapp.SessionTypeInfo{}, false, err
	}
	if !found {
		return childapp.SessionTypeInfo{}, false, nil
	}
	return childapp.SessionTypeInfo{
		ID:           info.ID,
		Name:         info.Name,
		StartMinutes: info.StartMinutes,
		EndMinutes:   info.EndMinutes,
		IsActive:     info.IsActive,
	}, true, nil
}

// GetActiveInScopeForTemplates satisfies the sessiontemplates-package lookup
// interface. Both packages need the same shape of projection, so the work is
// done in `lookup` and we project the result here.
func (a *sessionTypeLookupAdapter) GetActiveInScopeForTemplates(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (sessiontemplateapp.SessionTypeInfo, bool, error) {
	info, found, err := a.lookup(ctx, tenantID, branchID, sessionTypeID)
	if err != nil {
		return sessiontemplateapp.SessionTypeInfo{}, false, err
	}
	if !found {
		return sessiontemplateapp.SessionTypeInfo{}, false, nil
	}
	return sessiontemplateapp.SessionTypeInfo{
		ID:           info.ID,
		Name:         info.Name,
		StartMinutes: info.StartMinutes,
		EndMinutes:   info.EndMinutes,
		IsActive:     info.IsActive,
	}, true, nil
}

func (a *sessionTypeLookupAdapter) lookup(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (struct {
	ID           uuid.UUID
	Name         string
	StartMinutes int
	EndMinutes   int
	IsActive     bool
}, bool, error) {
	st, err := a.repo.GetByID(ctx, tenantID, branchID, sessionTypeID)
	if err != nil {
		var de *domainerrors.DomainError
		if errors.As(err, &de) {
			if len(de.Code) > 10 && de.Code[len(de.Code)-10:] == "_not_found" {
				return struct {
					ID           uuid.UUID
					Name         string
					StartMinutes int
					EndMinutes   int
					IsActive     bool
				}{}, false, nil
			}
		}
		return struct {
			ID           uuid.UUID
			Name         string
			StartMinutes int
			EndMinutes   int
			IsActive     bool
		}{}, false, fmt.Errorf("sessiontype lookup: %w", err)
	}
	return struct {
		ID           uuid.UUID
		Name         string
		StartMinutes int
		EndMinutes   int
		IsActive     bool
	}{
		ID:           st.ID,
		Name:         st.Name,
		StartMinutes: st.StartMinutes,
		EndMinutes:   st.EndMinutes,
		IsActive:     st.IsActive,
	}, true, nil
}

var _ childapp.SessionTypeLookup = (*sessionTypeLookupAdapter)(nil)
var _ sessiontemplateapp.SessionTypeLookup = (*sessionTemplateLookupTemplateAdapter)(nil)

// sessionTemplateLookupTemplateAdapter wraps the parent adapter to expose only
// the template-package lookup signature. This keeps the two interfaces
// (children + sessiontemplates) decoupled at the type level while sharing
// the underlying repository.
type sessionTemplateLookupTemplateAdapter struct {
	inner *sessionTypeLookupAdapter
}

func (a *sessionTemplateLookupTemplateAdapter) GetActiveInScope(ctx context.Context, tenantID, branchID, sessionTypeID uuid.UUID) (sessiontemplateapp.SessionTypeInfo, bool, error) {
	return a.inner.GetActiveInScopeForTemplates(ctx, tenantID, branchID, sessionTypeID)
}

// ── Term module adapters ──────────────────────────────────────────────────

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

// enrollmentTermCreatorAdapter implements childapp.EnrollmentTermCreator
// by creating a 12-month term for the child in the same transaction as enrollment.
type enrollmentTermCreatorAdapter struct {
	termRepo     *termpostgres.TermRepository
	rateProvider termapp.SiteRateProvider
	auditWriter  *audit.Writer
}

func (a *enrollmentTermCreatorAdapter) CreateEnrollmentTerm(ctx context.Context, tx pgx.Tx, actor tenant.ActorContext, childID uuid.UUID, termStartDate time.Time, bookingPatternID uuid.UUID) (uuid.UUID, error) {
	// 1. Idempotent: skip if child already has an active term.
	_, found, err := a.termRepo.GetActiveForChildInTx(ctx, tx, actor.TenantID, actor.BranchID, childID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("check existing term: %w", err)
	}
	if found {
		return uuid.Nil, nil
	}

	// 2. Snapshot the site hourly rate.
	rate, rateFound, err := a.rateProvider.SiteHourlyRateMinor(ctx, tx, actor.TenantID, actor.BranchID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("lookup site rate: %w", err)
	}
	if !rateFound || rate <= 0 {
		return uuid.Nil, domainerrors.New("site_rate_missing", "A site hourly rate must be configured before enrolling a child.", "site_hourly_rate")
	}

	// 3. Build the Term.
	termID := uid.NewUUID()
	term, err := termdomain.NewTerm(
		termID,
		actor.TenantID,
		actor.BranchID,
		childID,
		termStartDate,
		bookingPatternID,
		rate,
		actor.MembershipID,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("build term: %w", err)
	}

	// 4. Persist.
	saved, err := a.termRepo.Insert(ctx, tx, term)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert term: %w", err)
	}

	// 5. Update child denormalisation.
	if err := a.termRepo.SetChildCurrentTermID(ctx, tx, actor.TenantID, actor.BranchID, childID, saved.ID); err != nil {
		return uuid.Nil, fmt.Errorf("set child current term: %w", err)
	}

	// 6. Audit.
	if err := a.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
		ActionType: termdomain.AuditTermCreated,
		EntityType: termdomain.AuditEntityTerm,
		EntityID:   saved.ID,
		Details: map[string]any{
			"child_id":               saved.ChildID.String(),
			"term_start_date":        saved.TermStartDate.Format("2006-01-02"),
			"term_end_date":          saved.TermEndDate.Format("2006-01-02"),
			"booking_pattern_id":     saved.BookingPatternID.String(),
			"site_hourly_rate_minor": saved.SiteHourlyRateMinor,
			"status":                 string(saved.Status),
		},
	}); err != nil {
		return uuid.Nil, fmt.Errorf("audit term_created: %w", err)
	}

	return saved.ID, nil
}

var _ childapp.EnrollmentTermCreator = (*enrollmentTermCreatorAdapter)(nil)

// ── Billing pipeline adapters ──────────────────────────────────────────

// termDateLookupAdapter satisfies billingdomain.TermDateLookup by delegating
// to the term_calendar module's academic term repository.
type termDateLookupAdapter struct {
	repo *termcalendarpostgres.AcademicTermRepository
}

func (a *termDateLookupAdapter) GetTermDateRangesForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, month time.Time) ([]billingdomain.TermDateRange, error) {
	from := month
	to := month.AddDate(0, 1, 0).AddDate(0, 0, -1)
	ranges, err := a.repo.ListActiveDateRanges(ctx, tenantID, branchID, from, to)
	if err != nil {
		return nil, fmt.Errorf("term date lookup: %w", err)
	}
	out := make([]billingdomain.TermDateRange, 0, len(ranges))
	for _, r := range ranges {
		out = append(out, billingdomain.TermDateRange{
			StartDate: r.StartDate,
			EndDate:   r.EndDate,
		})
	}
	return out, nil
}

var _ billingdomain.TermDateLookup = (*termDateLookupAdapter)(nil)

// adHocBookingLookupAdapter satisfies billingdomain.AdHocBookingLookup by
// delegating to the billing repository's ad-hoc booking query.
type adHocBookingLookupAdapter struct {
	repo *billingpostgres.Repository
}

func (a *adHocBookingLookupAdapter) ListActiveBookingsForChildInMonth(ctx context.Context, tenantID, branchID, childID uuid.UUID, month time.Time) ([]billingdomain.AdHocBookingRow, error) {
	from := month
	to := month.AddDate(0, 1, 0).AddDate(0, 0, -1)
	rows, err := a.repo.ListActiveAdHocBookingsForChildInMonth(ctx, nil, tenantID, branchID, childID, from, to)
	if err != nil {
		return nil, fmt.Errorf("ad-hoc booking lookup: %w", err)
	}
	return rows, nil
}

var _ billingdomain.AdHocBookingLookup = (*adHocBookingLookupAdapter)(nil)

// closureDateLookupAdapter satisfies billingdomain.ClosureDateLookup by delegating
// to the branch_closures module's repository.
type closureDateLookupAdapter struct {
	repo *branchclosurepostgres.Repository
}

func (a *closureDateLookupAdapter) GetClosureDatesForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, month time.Time) ([]time.Time, error) {
	dates, err := a.repo.ListClosureDatesForBranchAndMonth(ctx, tenantID, branchID, month)
	if err != nil {
		return nil, fmt.Errorf("closure date lookup: %w", err)
	}
	return dates, nil
}

var _ billingdomain.ClosureDateLookup = (*closureDateLookupAdapter)(nil)

// hourlyBookingLookupAdapter satisfies billingdomain.HourlyBookingLookup by
// delegating to the hourly_bookings module's repository.
type hourlyBookingLookupAdapter struct {
	repo *hourlypostgres.HourlyBookingRepository
}

func (a *hourlyBookingLookupAdapter) ListActiveByChildAndMonth(ctx context.Context, tenantID, branchID, childID uuid.UUID, monthStart, monthEnd time.Time) ([]billingdomain.HourlyBookingRow, error) {
	rows, err := a.repo.ListActiveByChildAndDateRange(ctx, tenantID, branchID, childID, monthStart, monthEnd)
	if err != nil {
		return nil, fmt.Errorf("hourly booking lookup: %w", err)
	}
	out := make([]billingdomain.HourlyBookingRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, billingdomain.HourlyBookingRow{
			ID:               r.ID,
			ChildID:          r.ChildID,
			CalendarDate:     r.CalendarDate,
			StartTimeMinutes: r.StartTimeMinutes,
			DurationMinutes:  r.DurationMinutes,
		})
	}
	return out, nil
}

var _ billingdomain.HourlyBookingLookup = (*hourlyBookingLookupAdapter)(nil)

// ── Room Capacity Lookup adapter ───────────────────────────────────────────

// roomCapacityLookupAdapter satisfies bookingsapp.RoomCapacityLookup by
// delegating to the rooms module's repository.
type roomCapacityLookupAdapter struct {
	repo      *roomspostgres.RoomRepository
	childRepo *postgreschild.ChildRepository
}

func (a *roomCapacityLookupAdapter) ListActiveRooms(ctx context.Context, tenantID, branchID uuid.UUID) ([]bookingsapp.RoomInfo, error) {
	rooms, err := a.repo.ListByBranch(ctx, tenantID, branchID, false)
	if err != nil {
		return nil, fmt.Errorf("room capacity lookup: %w", err)
	}
	out := make([]bookingsapp.RoomInfo, 0, len(rooms))
	for _, r := range rooms {
		if r.IsActive {
			out = append(out, bookingsapp.RoomInfo{
				RoomID:   r.ID,
				RoomName: r.Name,
				Capacity: r.Capacity,
			})
		}
	}
	return out, nil
}

func (a *roomCapacityLookupAdapter) ListRoomAssignmentsForChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) ([]bookingsapp.ChildRoomAssignmentInfo, error) {
	assignments, err := a.childRepo.ListRoomAssignmentsByChild(ctx, tenantID, branchID, childID)
	if err != nil {
		return nil, fmt.Errorf("room assignments lookup: %w", err)
	}
	out := make([]bookingsapp.ChildRoomAssignmentInfo, 0, len(assignments))
	for _, a := range assignments {
		out = append(out, bookingsapp.ChildRoomAssignmentInfo{
			RoomID:    a.RoomID,
			StartDate: a.StartDate,
			EndDate:   a.EndDate,
		})
	}
	return out, nil
}

var _ bookingsapp.RoomCapacityLookup = (*roomCapacityLookupAdapter)(nil)

// ── Billing Notification adapter ──────────────────────────────────────────

//go:embed templates/*.html templates/*.txt
var notificationTemplatesFS embed.FS

type notificationTemplateData struct {
	NurseryName   string
	ChildName     string
	InvoiceNumber string
	BillingMonth  string
	TotalDue      string
	DueDate       string
	PortalLink    string
}

type billingNotificationAdapter struct {
	repo           billingdomain.BillingRepository
	parentContacts billingapp.ParentContactLookup
	siteProfiles   billingapp.SiteProfileLookup
	sender         email.Sender
	auditWriter    *audit.Writer
	webBaseURL     string
}

func (a *billingNotificationAdapter) SendInvoiceIssuedEmail(ctx context.Context, tx pgx.Tx, invoiceID, tenantID, branchID uuid.UUID) error {
	invoice, found, err := a.repo.GetInvoiceForManagerReview(ctx, tenantID, branchID, invoiceID)
	if err != nil {
		return fmt.Errorf("get invoice: %w", err)
	}
	if !found {
		return fmt.Errorf("invoice %s not found", invoiceID)
	}

	parent, err := a.parentContacts.GetForInvoice(ctx, tenantID, branchID, invoice.ChildID)
	if err != nil {
		return fmt.Errorf("get parent contact: %w", err)
	}
	if parent == nil || parent.Email == "" {
		return nil
	}

	site, err := a.siteProfiles.GetForInvoice(ctx, tenantID, branchID)
	if err != nil {
		return fmt.Errorf("get site profile: %w", err)
	}
	if site == nil {
		return fmt.Errorf("site profile not found")
	}

	childName := invoice.ChildFirstName
	if invoice.ChildLastName != nil {
		childName += " " + *invoice.ChildLastName
	}

	invoiceNumber := ""
	if invoice.InvoiceNumber != nil {
		invoiceNumber = *invoice.InvoiceNumber
	}

	billingMonth := invoice.BillingMonth.Format("January 2006")
	totalDue := formatMoney(invoice.TotalDue)
	dueDate := ""
	if invoice.DueAt != nil {
		dueDate = invoice.DueAt.Format("2 January 2006")
	}

	portalLink := fmt.Sprintf("%s/parent/billing/%s", a.webBaseURL, invoiceID)

	data := notificationTemplateData{
		NurseryName:   site.NurseryName,
		ChildName:     childName,
		InvoiceNumber: invoiceNumber,
		BillingMonth:  billingMonth,
		TotalDue:      totalDue,
		DueDate:       dueDate,
		PortalLink:    portalLink,
	}

	subject := fmt.Sprintf("New Invoice %s - %s", invoiceNumber, site.NurseryName)
	htmlBody, textBody, err := renderTemplates("issued", data)
	if err != nil {
		return fmt.Errorf("render templates: %w", err)
	}

	msg := email.Message{
		To:      parent.Email,
		Subject: subject,
		Text:    textBody,
		HTML:    htmlBody,
	}

	if err := a.sender.Send(ctx, msg); err != nil {
		a.writeAudit(ctx, tx, tenantID, branchID, invoiceID, parent.Email, notificationsapp.AuditNotificationInvoiceIssuedFailed, err)
		return fmt.Errorf("send email: %w", err)
	}

	a.writeAudit(ctx, tx, tenantID, branchID, invoiceID, parent.Email, notificationsapp.AuditNotificationInvoiceIssuedSent, nil)
	return nil
}

func (a *billingNotificationAdapter) SendInvoiceOverdueEmail(ctx context.Context, tx pgx.Tx, invoiceID, tenantID, branchID uuid.UUID) error {
	invoice, found, err := a.repo.GetInvoiceForManagerReview(ctx, tenantID, branchID, invoiceID)
	if err != nil {
		return fmt.Errorf("get invoice: %w", err)
	}
	if !found {
		return fmt.Errorf("invoice %s not found", invoiceID)
	}

	parent, err := a.parentContacts.GetForInvoice(ctx, tenantID, branchID, invoice.ChildID)
	if err != nil {
		return fmt.Errorf("get parent contact: %w", err)
	}
	if parent == nil || parent.Email == "" {
		return nil
	}

	site, err := a.siteProfiles.GetForInvoice(ctx, tenantID, branchID)
	if err != nil {
		return fmt.Errorf("get site profile: %w", err)
	}
	if site == nil {
		return fmt.Errorf("site profile not found")
	}

	childName := invoice.ChildFirstName
	if invoice.ChildLastName != nil {
		childName += " " + *invoice.ChildLastName
	}

	invoiceNumber := ""
	if invoice.InvoiceNumber != nil {
		invoiceNumber = *invoice.InvoiceNumber
	}

	totalDue := formatMoney(invoice.TotalDue)
	portalLink := fmt.Sprintf("%s/parent/billing/%s", a.webBaseURL, invoiceID)

	data := notificationTemplateData{
		NurseryName:   site.NurseryName,
		ChildName:     childName,
		InvoiceNumber: invoiceNumber,
		TotalDue:      totalDue,
		PortalLink:    portalLink,
	}

	subject := fmt.Sprintf("Invoice Overdue %s - %s", invoiceNumber, site.NurseryName)
	htmlBody, textBody, err := renderTemplates("overdue", data)
	if err != nil {
		return fmt.Errorf("render templates: %w", err)
	}

	msg := email.Message{
		To:      parent.Email,
		Subject: subject,
		Text:    textBody,
		HTML:    htmlBody,
	}

	if err := a.sender.Send(ctx, msg); err != nil {
		a.writeAudit(ctx, tx, tenantID, branchID, invoiceID, parent.Email, notificationsapp.AuditNotificationInvoiceOverdueFailed, err)
		return fmt.Errorf("send email: %w", err)
	}

	a.writeAudit(ctx, tx, tenantID, branchID, invoiceID, parent.Email, notificationsapp.AuditNotificationInvoiceOverdueSent, nil)
	return nil
}

func (a *billingNotificationAdapter) SendInvoiceDueSoonEmail(ctx context.Context, tx pgx.Tx, invoiceID, tenantID, branchID uuid.UUID) error {
	invoice, found, err := a.repo.GetInvoiceForManagerReview(ctx, tenantID, branchID, invoiceID)
	if err != nil {
		return fmt.Errorf("get invoice: %w", err)
	}
	if !found {
		return fmt.Errorf("invoice %s not found", invoiceID)
	}

	parent, err := a.parentContacts.GetForInvoice(ctx, tenantID, branchID, invoice.ChildID)
	if err != nil {
		return fmt.Errorf("get parent contact: %w", err)
	}
	if parent == nil || parent.Email == "" {
		return nil
	}

	site, err := a.siteProfiles.GetForInvoice(ctx, tenantID, branchID)
	if err != nil {
		return fmt.Errorf("get site profile: %w", err)
	}
	if site == nil {
		return fmt.Errorf("site profile not found")
	}

	childName := invoice.ChildFirstName
	if invoice.ChildLastName != nil {
		childName += " " + *invoice.ChildLastName
	}

	invoiceNumber := ""
	if invoice.InvoiceNumber != nil {
		invoiceNumber = *invoice.InvoiceNumber
	}

	totalDue := formatMoney(invoice.TotalDue)
	dueDate := ""
	if invoice.DueAt != nil {
		dueDate = invoice.DueAt.Format("2 January 2006")
	}

	portalLink := fmt.Sprintf("%s/parent/billing/%s", a.webBaseURL, invoiceID)

	data := notificationTemplateData{
		NurseryName:   site.NurseryName,
		ChildName:     childName,
		InvoiceNumber: invoiceNumber,
		TotalDue:      totalDue,
		DueDate:       dueDate,
		PortalLink:    portalLink,
	}

	subject := fmt.Sprintf("Payment Reminder: Invoice %s Due Soon - %s", invoiceNumber, site.NurseryName)
	htmlBody, textBody, err := renderTemplates("due-soon", data)
	if err != nil {
		return fmt.Errorf("render templates: %w", err)
	}

	msg := email.Message{
		To:      parent.Email,
		Subject: subject,
		Text:    textBody,
		HTML:    htmlBody,
	}

	if err := a.sender.Send(ctx, msg); err != nil {
		a.writeAudit(ctx, tx, tenantID, branchID, invoiceID, parent.Email, notificationsapp.AuditNotificationInvoiceDueSoonFailed, err)
		return fmt.Errorf("send email: %w", err)
	}

	a.writeAudit(ctx, tx, tenantID, branchID, invoiceID, parent.Email, notificationsapp.AuditNotificationInvoiceDueSoonSent, nil)
	return nil
}

func (a *billingNotificationAdapter) SendInvoiceDueReminderEmail(ctx context.Context, tx pgx.Tx, invoiceID, tenantID, branchID uuid.UUID) error {
	invoice, found, err := a.repo.GetInvoiceForManagerReview(ctx, tenantID, branchID, invoiceID)
	if err != nil {
		return fmt.Errorf("get invoice: %w", err)
	}
	if !found {
		return fmt.Errorf("invoice %s not found", invoiceID)
	}

	parent, err := a.parentContacts.GetForInvoice(ctx, tenantID, branchID, invoice.ChildID)
	if err != nil {
		return fmt.Errorf("get parent contact: %w", err)
	}
	if parent == nil || parent.Email == "" {
		return nil
	}

	site, err := a.siteProfiles.GetForInvoice(ctx, tenantID, branchID)
	if err != nil {
		return fmt.Errorf("get site profile: %w", err)
	}
	if site == nil {
		return fmt.Errorf("site profile not found")
	}

	childName := invoice.ChildFirstName
	if invoice.ChildLastName != nil {
		childName += " " + *invoice.ChildLastName
	}

	invoiceNumber := ""
	if invoice.InvoiceNumber != nil {
		invoiceNumber = *invoice.InvoiceNumber
	}

	totalDue := formatMoney(invoice.TotalDue)
	dueDate := ""
	if invoice.DueAt != nil {
		dueDate = invoice.DueAt.Format("2 January 2006")
	}

	portalLink := fmt.Sprintf("%s/parent/billing/%s", a.webBaseURL, invoiceID)

	data := notificationTemplateData{
		NurseryName:   site.NurseryName,
		ChildName:     childName,
		InvoiceNumber: invoiceNumber,
		TotalDue:      totalDue,
		DueDate:       dueDate,
		PortalLink:    portalLink,
	}

	subject := fmt.Sprintf("Payment Due Today: Invoice %s - %s", invoiceNumber, site.NurseryName)
	htmlBody, textBody, err := renderTemplates("due-reminder", data)
	if err != nil {
		return fmt.Errorf("render templates: %w", err)
	}

	msg := email.Message{
		To:      parent.Email,
		Subject: subject,
		Text:    textBody,
		HTML:    htmlBody,
	}

	if err := a.sender.Send(ctx, msg); err != nil {
		a.writeAudit(ctx, tx, tenantID, branchID, invoiceID, parent.Email, notificationsapp.AuditNotificationInvoiceDueReminderFailed, err)
		return fmt.Errorf("send email: %w", err)
	}

	a.writeAudit(ctx, tx, tenantID, branchID, invoiceID, parent.Email, notificationsapp.AuditNotificationInvoiceDueReminderSent, nil)
	return nil
}

func renderTemplates(name string, data notificationTemplateData) (htmlBody, textBody string, err error) {
	htmlTmpl, err := template.ParseFS(notificationTemplatesFS, "templates/"+name+".html")
	if err != nil {
		return "", "", fmt.Errorf("parse html template: %w", err)
	}

	var htmlBuf strings.Builder
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", fmt.Errorf("execute html template: %w", err)
	}

	textTmpl, err := template.ParseFS(notificationTemplatesFS, "templates/"+name+".txt")
	if err != nil {
		return "", "", fmt.Errorf("parse text template: %w", err)
	}

	var textBuf strings.Builder
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", fmt.Errorf("execute text template: %w", err)
	}

	return htmlBuf.String(), textBuf.String(), nil
}

func formatMoney(m billingdomain.Money) string {
	minor := m.Minor()
	pounds := minor / 100
	pence := minor % 100
	if pence < 0 {
		pence = -pence
	}
	return fmt.Sprintf("£%d.%02d", pounds, pence)
}

func (a *billingNotificationAdapter) writeAudit(ctx context.Context, tx pgx.Tx, tenantID, branchID, invoiceID uuid.UUID, emailAddr, actionType string, sendErr error) {
	actor := tenant.ActorContext{
		TenantID: tenantID,
		BranchID: branchID,
	}
	details := map[string]any{
		"invoice_id":        invoiceID.String(),
		"notification_type": actionType,
	}
	if emailAddr != "" {
		parts := strings.SplitN(emailAddr, "@", 2)
		if len(parts) == 2 {
			details["parent_email_domain"] = parts[1]
		}
	}
	if sendErr != nil {
		details["error"] = sendErr.Error()
	}
	_ = a.auditWriter.WriteWithTx(ctx, tx, actor, audit.WriteParams{
		ActionType: actionType,
		EntityType: "invoice",
		EntityID:   invoiceID,
		Details:    details,
	})
}

var _ billingdomain.HourlyBookingLookup = (*hourlyBookingLookupAdapter)(nil)

type consumedMinutesProviderAdapter struct {
	pool *pgxpool.Pool
}

func (a *consumedMinutesProviderAdapter) GetConsumedMinutes(ctx context.Context, tenantID, branchID uuid.UUID, childIDs []uuid.UUID, billingMonth time.Time) (map[uuid.UUID]int, error) {
	if len(childIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	q := sqlc.New(a.pool)
	pgChildIDs := make([]pgtype.UUID, len(childIDs))
	for i, id := range childIDs {
		pgChildIDs[i] = pgtype.UUID{Bytes: [16]byte(id), Valid: true}
	}

	rows, err := q.GetConsumedMinutesByChildren(ctx, sqlc.GetConsumedMinutesByChildrenParams{
		TenantID:     pgtype.UUID{Bytes: [16]byte(tenantID), Valid: true},
		BranchID:     pgtype.UUID{Bytes: [16]byte(branchID), Valid: true},
		BillingMonth: pgtype.Date{Time: billingMonth, Valid: true},
		Column4:      pgChildIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("get consumed minutes: %w", err)
	}

	result := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		childID := uuid.UUID(row.ChildID.Bytes)
		consumed := 0
		if v, ok := row.ConsumedMinutes.(int64); ok {
			consumed = int(v)
		} else if v, ok := row.ConsumedMinutes.(int32); ok {
			consumed = int(v)
		}
		result[childID] = consumed
	}
	return result, nil
}

// termDateProviderAdapter satisfies fundingdomain.TermDateProvider by delegating
// to the academic term repository.
type termDateProviderAdapter struct {
	repo *termcalendarpostgres.AcademicTermRepository
}

func (a *termDateProviderAdapter) GetTermDatesForBranchAndMonth(ctx context.Context, tenantID, branchID uuid.UUID, month time.Time) ([]fundingdomain.TermDateRange, error) {
	from := month
	to := month.AddDate(0, 1, 0).AddDate(0, 0, -1)
	ranges, err := a.repo.ListActiveDateRanges(ctx, tenantID, branchID, from, to)
	if err != nil {
		return nil, fmt.Errorf("term date provider: %w", err)
	}
	out := make([]fundingdomain.TermDateRange, 0, len(ranges))
	for _, r := range ranges {
		out = append(out, fundingdomain.TermDateRange{
			StartDate: r.StartDate,
			EndDate:   r.EndDate,
		})
	}
	return out, nil
}

var _ fundingdomain.TermDateProvider = (*termDateProviderAdapter)(nil)

func provideTermDateProviderAdapter(
	repo *termcalendarpostgres.AcademicTermRepository,
) *termDateProviderAdapter {
	return &termDateProviderAdapter{repo: repo}
}

// childFundingWriterAdapter satisfies childdomain.ChildFundingWriter by mapping
// ChildFundingRecordInput to the funding module's FundingRecord and upserting
// via the funding module's repository.
type childFundingWriterAdapter struct {
	repo *fundingpostgres.FundingRecordRepositoryImpl
}

func (a *childFundingWriterAdapter) SaveFunding(ctx context.Context, tx any, tenantID, branchID, childID uuid.UUID, input *childdomain.ChildFundingRecordInput) error {
	ft := fundingdomain.FundingType(input.FundingType)
	if !ft.Valid() {
		ft = fundingdomain.FundingTypeUnknown
	}
	fm := fundingdomain.FundingModel(input.FundingModel)
	if !fm.Valid() {
		fm = fundingdomain.FundingModelUnknown
	}

	var startDate, endDate *time.Time
	if input.FundingStartDate != nil && *input.FundingStartDate != "" {
		if t, err := time.Parse("2006-01-02", *input.FundingStartDate); err == nil {
			startDate = &t
		}
	}
	if input.FundingEndDate != nil && *input.FundingEndDate != "" {
		if t, err := time.Parse("2006-01-02", *input.FundingEndDate); err == nil {
			endDate = &t
		}
	}

	record := fundingdomain.FundingRecord{
		ID:                       uid.NewUUID(),
		TenantID:                 tenantID,
		BranchID:                 branchID,
		ChildID:                  childID,
		FundingEnabled:           input.FundingEnabled,
		FundingType:              ft,
		FundingModel:             fm,
		FundedHoursPerWeek:       input.FundedHoursPerWeek,
		FundingStartDate:         startDate,
		FundingEndDate:           endDate,
		EligibilityCode:          input.EligibilityCode,
		EligibilityCodeValidated: input.EligibilityCodeValidated,
		EvidenceReceived:         input.EvidenceReceived,
	}

	_, err := a.repo.UpsertFundingRecord(ctx, tx, record)
	if err != nil {
		return fmt.Errorf("adapter upsert funding record: %w", err)
	}
	return nil
}

var _ childdomain.ChildFundingWriter = (*childFundingWriterAdapter)(nil)

func provideChildFundingWriterAdapter(
	repo *fundingpostgres.FundingRecordRepositoryImpl,
) *childFundingWriterAdapter {
	return &childFundingWriterAdapter{repo: repo}
}

// childFundingReaderAdapter satisfies childdomain.ChildFundingReader by reading
// from the funding module's repository and mapping to the children domain struct.
type childFundingReaderAdapter struct {
	repo *fundingpostgres.FundingRecordRepositoryImpl
}

func (a *childFundingReaderAdapter) GetFundingForChild(ctx context.Context, tenantID, branchID, childID uuid.UUID) (*childdomain.ChildFundingData, bool, error) {
	record, found, err := a.repo.GetFundingRecord(ctx, tenantID, branchID, childID)
	if err != nil {
		return nil, false, fmt.Errorf("adapter get funding record: %w", err)
	}
	if !found {
		return nil, false, nil
	}

	var startDate, endDate *time.Time
	if record.FundingStartDate != nil {
		startDate = record.FundingStartDate
	}
	if record.FundingEndDate != nil {
		endDate = record.FundingEndDate
	}

	return &childdomain.ChildFundingData{
		ID:                       record.ID,
		ChildID:                  record.ChildID,
		FundingEnabled:           record.FundingEnabled,
		FundingType:              string(record.FundingType),
		FundingModel:             string(record.FundingModel),
		FundedHoursPerWeek:       record.FundedHoursPerWeek,
		FundingStartDate:         startDate,
		FundingEndDate:           endDate,
		EligibilityCode:          record.EligibilityCode,
		EligibilityCodeValidated: record.EligibilityCodeValidated,
		EvidenceReceived:         record.EvidenceReceived,
		BenefitsStatus:           "unknown",
		CreatedAt:                record.CreatedAt,
		UpdatedAt:                record.UpdatedAt,
	}, true, nil
}

var _ childdomain.ChildFundingReader = (*childFundingReaderAdapter)(nil)

func provideChildFundingReaderAdapter(
	repo *fundingpostgres.FundingRecordRepositoryImpl,
) *childFundingReaderAdapter {
	return &childFundingReaderAdapter{repo: repo}
}

// fundingLookupAdapter satisfies billingdomain.FundingLookup by loading
// FundingRecord from the funding module's repository and computing allowance on the fly.
type fundingLookupAdapter struct {
	fundingRepo *fundingpostgres.FundingRecordRepositoryImpl
	ownerRepo   *ownerpostgres.OwnerRepository
}

func (a *fundingLookupAdapter) GetChildFunding(ctx context.Context, tenantID, branchID, childID uuid.UUID, billingMonth time.Time) (billingdomain.FundedChildInfo, error) {
	record, found, err := a.fundingRepo.GetFundingRecord(ctx, tenantID, branchID, childID)
	if err != nil {
		return billingdomain.FundedChildInfo{}, fmt.Errorf("get child funding record: %w", err)
	}
	if !found || !record.FundingEnabled {
		return billingdomain.FundedChildInfo{HasFunding: false}, nil
	}

	site, err := a.ownerRepo.GetActiveSite(ctx, tenantID, branchID)
	if err != nil {
		return billingdomain.FundedChildInfo{}, fmt.Errorf("get site for funded rate: %w", err)
	}
	fundedRateMinor := 0
	if site.FundedHourlyRateMinor != nil {
		fundedRateMinor = *site.FundedHourlyRateMinor
	}

	hoursPerWeek := 0.0
	if record.FundedHoursPerWeek != nil {
		hoursPerWeek = *record.FundedHoursPerWeek
	}

	fundingModel := record.FundingModel
	termDates, _ := a.getTermDates(ctx, tenantID, branchID, billingMonth)
	allowance, _ := fundingdomain.ComputeAllowanceMinutes(hoursPerWeek, fundingModel, termDates, billingMonth, nil, record.FundingStartDate, record.FundingEndDate)

	return billingdomain.FundedChildInfo{
		HasFunding:             true,
		FundingType:            string(record.FundingType),
		FundedAllowanceMinutes: allowance,
		FundedHourlyRateMinor:  fundedRateMinor,
		FundedHoursPerWeek:     hoursPerWeek,
	}, nil
}

func (a *fundingLookupAdapter) getTermDates(ctx context.Context, tenantID, branchID uuid.UUID, billingMonth time.Time) ([]fundingdomain.TermDateRange, error) {
	// This is a simplified implementation; the real one would query the term_calendar module.
	// For now, return nil to use the fallback computation.
	return nil, nil
}

// ── Bookings Funding Lookup adapter ─────────────────────────────────────

// bookingsFundingLookupAdapter satisfies bookingsdomain.FundingLookup by
// loading FundingRecord from the funding module's repository.
type bookingsFundingLookupAdapter struct {
	fundingRepo *fundingpostgres.FundingRecordRepositoryImpl
}

func (a *bookingsFundingLookupAdapter) GetChildFunding(ctx context.Context, tenantID, branchID, childID uuid.UUID) (bookingsdomain.FundingInfo, error) {
	record, found, err := a.fundingRepo.GetFundingRecord(ctx, tenantID, branchID, childID)
	if err != nil {
		return bookingsdomain.FundingInfo{}, fmt.Errorf("get child funding record: %w", err)
	}
	if !found || !record.FundingEnabled {
		return bookingsdomain.FundingInfo{HasFunding: false}, nil
	}

	laRef := ""
	if record.EligibilityCode != nil {
		laRef = *record.EligibilityCode
	}

	termTimeOnly := record.FundingModel == fundingdomain.FundingModelTermTimeOnly

	return bookingsdomain.FundingInfo{
		HasFunding:         true,
		FundingType:        string(record.FundingType),
		FundedHoursPerWeek: record.FundedHoursPerWeek,
		LaReference:        &laRef,
		TermTimeOnly:       termTimeOnly,
	}, nil
}

func provideBookingsFundingLookupAdapter(
	fundingRepo *fundingpostgres.FundingRecordRepositoryImpl,
) *bookingsFundingLookupAdapter {
	return &bookingsFundingLookupAdapter{fundingRepo: fundingRepo}
}

var _ bookingsdomain.FundingLookup = (*bookingsFundingLookupAdapter)(nil)
