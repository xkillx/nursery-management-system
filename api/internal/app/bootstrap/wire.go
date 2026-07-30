//go:build wireinject

package bootstrap

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/email"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/ratelimit"
	"nursery-management-system/api/internal/platform/transaction"

	billingapp "nursery-management-system/api/internal/modules/billing/application"

	authapp "nursery-management-system/api/internal/modules/authentication/application"
	authdomain "nursery-management-system/api/internal/modules/authentication/domain"
	authpostgres "nursery-management-system/api/internal/modules/authentication/infrastructure/postgres"
	authtokens "nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	authhandler "nursery-management-system/api/internal/modules/authentication/interfaces/http"

	childapp "nursery-management-system/api/internal/modules/children/application"
	childdomain "nursery-management-system/api/internal/modules/children/domain"
	childpostgres "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	childstorage "nursery-management-system/api/internal/modules/children/infrastructure/storage"
	childhandler "nursery-management-system/api/internal/modules/children/interfaces/http"

	parentchildapp "nursery-management-system/api/internal/modules/parentchildmappings/application"
	parentchilddomain "nursery-management-system/api/internal/modules/parentchildmappings/domain"
	parentchildpostgres "nursery-management-system/api/internal/modules/parentchildmappings/infrastructure/postgres"
	parentchildhandler "nursery-management-system/api/internal/modules/parentchildmappings/interfaces/http"

	parentsapp "nursery-management-system/api/internal/modules/parents/application"
	parentsdomain "nursery-management-system/api/internal/modules/parents/domain"
	parentspostgres "nursery-management-system/api/internal/modules/parents/infrastructure/postgres"
	parentshandler "nursery-management-system/api/internal/modules/parents/interfaces/http"

	attendanceapp "nursery-management-system/api/internal/modules/attendance/application"
	attendancedomain "nursery-management-system/api/internal/modules/attendance/domain"
	attendancepostgres "nursery-management-system/api/internal/modules/attendance/infrastructure/postgres"
	attendancehandler "nursery-management-system/api/internal/modules/attendance/interfaces/http"

	absenceapp "nursery-management-system/api/internal/modules/absence/application"
	absencedomain "nursery-management-system/api/internal/modules/absence/domain"
	absencepostgres "nursery-management-system/api/internal/modules/absence/infrastructure/postgres"
	absencehandler "nursery-management-system/api/internal/modules/absence/interfaces/http"

	fundingapp "nursery-management-system/api/internal/modules/funding/application"
	fundingdomain "nursery-management-system/api/internal/modules/funding/domain"
	fundingpostgres "nursery-management-system/api/internal/modules/funding/infrastructure/postgres"
	httpfunding "nursery-management-system/api/internal/modules/funding/interfaces/http"

	billingdomain "nursery-management-system/api/internal/modules/billing/domain"
	billingpdf "nursery-management-system/api/internal/modules/billing/infrastructure/pdf"
	billingpostgres "nursery-management-system/api/internal/modules/billing/infrastructure/postgres"
	billinghandler "nursery-management-system/api/internal/modules/billing/interfaces/http"

	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	paymentsdomain "nursery-management-system/api/internal/modules/payments/domain"
	paymentspostgres "nursery-management-system/api/internal/modules/payments/infrastructure/postgres"
	stripeclient "nursery-management-system/api/internal/modules/payments/infrastructure/stripe"

	inviteapp "nursery-management-system/api/internal/modules/invites/application"
	invitedomain "nursery-management-system/api/internal/modules/invites/domain"
	invitepostgres "nursery-management-system/api/internal/modules/invites/infrastructure/postgres"
	invitetokens "nursery-management-system/api/internal/modules/invites/infrastructure/tokens"
	invitehandler "nursery-management-system/api/internal/modules/invites/interfaces/http"

	notificationsapp "nursery-management-system/api/internal/modules/notifications/application"

	emaildomain "nursery-management-system/api/internal/modules/email/domain"

	ownerapp "nursery-management-system/api/internal/modules/owner/application"
	ownerdomain "nursery-management-system/api/internal/modules/owner/domain"
	ownerpostgres "nursery-management-system/api/internal/modules/owner/infrastructure/postgres"
	ownerhandler "nursery-management-system/api/internal/modules/owner/interfaces/http"

	roomsapp "nursery-management-system/api/internal/modules/rooms/application"
	roomsdomain "nursery-management-system/api/internal/modules/rooms/domain"
	roomspostgres "nursery-management-system/api/internal/modules/rooms/infrastructure/postgres"
	roomshttphandler "nursery-management-system/api/internal/modules/rooms/interfaces/http"

	sessiontypeapp "nursery-management-system/api/internal/modules/sessiontypes/application"
	sessiontypedomain "nursery-management-system/api/internal/modules/sessiontypes/domain"
	sessiontypepostgres "nursery-management-system/api/internal/modules/sessiontypes/infrastructure/postgres"
	sessiontypehttphandler "nursery-management-system/api/internal/modules/sessiontypes/interfaces/http"

	sessiontemplateapp "nursery-management-system/api/internal/modules/sessiontemplates/application"
	sessiontemplatedomain "nursery-management-system/api/internal/modules/sessiontemplates/domain"
	sessiontemplatepostgres "nursery-management-system/api/internal/modules/sessiontemplates/infrastructure/postgres"
	sessiontemplatehttphandler "nursery-management-system/api/internal/modules/sessiontemplates/interfaces/http"

	resetapp "nursery-management-system/api/internal/modules/passwordreset/application"
	resetdomain "nursery-management-system/api/internal/modules/passwordreset/domain"
	resetpostgres "nursery-management-system/api/internal/modules/passwordreset/infrastructure/postgres"
	resethandler "nursery-management-system/api/internal/modules/passwordreset/interfaces/http"

	termcalendarapp "nursery-management-system/api/internal/modules/term_calendar/application"
	termcalendardomain "nursery-management-system/api/internal/modules/term_calendar/domain"
	termcalendarpostgres "nursery-management-system/api/internal/modules/term_calendar/infrastructure/postgres"
	termcalendarhttphandler "nursery-management-system/api/internal/modules/term_calendar/interfaces/http"

	adhocapp "nursery-management-system/api/internal/modules/ad_hoc_bookings/application"
	adhocdomain "nursery-management-system/api/internal/modules/ad_hoc_bookings/domain"
	adhocpostgres "nursery-management-system/api/internal/modules/ad_hoc_bookings/infrastructure/postgres"
	adhochttphandler "nursery-management-system/api/internal/modules/ad_hoc_bookings/interfaces/http"

	hourlyapp "nursery-management-system/api/internal/modules/hourly_bookings/application"
	hourlydomain "nursery-management-system/api/internal/modules/hourly_bookings/domain"
	hourlypostgres "nursery-management-system/api/internal/modules/hourly_bookings/infrastructure/postgres"
	hourlyhttphandler "nursery-management-system/api/internal/modules/hourly_bookings/interfaces/http"

	bookingsapp "nursery-management-system/api/internal/modules/bookings/application"
	bookingsdomain "nursery-management-system/api/internal/modules/bookings/domain"
	bookingspostgres "nursery-management-system/api/internal/modules/bookings/infrastructure/postgres"
	bookingshttphandler "nursery-management-system/api/internal/modules/bookings/interfaces/http"

	siteprofileapp "nursery-management-system/api/internal/modules/siteprofile/application"
	siteprofiledomain "nursery-management-system/api/internal/modules/siteprofile/domain"
	siteprofilepostgres "nursery-management-system/api/internal/modules/siteprofile/infrastructure/postgres"
	siteprofilehandler "nursery-management-system/api/internal/modules/siteprofile/interfaces/http"

	branchclosureapp "nursery-management-system/api/internal/modules/branch_closures/application"
	branchclosuredomain "nursery-management-system/api/internal/modules/branch_closures/domain"
	branchclosurepostgres "nursery-management-system/api/internal/modules/branch_closures/infrastructure/postgres"
	branchclosurehandler "nursery-management-system/api/internal/modules/branch_closures/interfaces/http"

	holidayperiodsapp "nursery-management-system/api/internal/modules/holiday_periods/application"
	holidayperiodsdomain "nursery-management-system/api/internal/modules/holiday_periods/domain"
	holidayperiodspostgres "nursery-management-system/api/internal/modules/holiday_periods/infrastructure/postgres"
	holidayperiodshttphandler "nursery-management-system/api/internal/modules/holiday_periods/interfaces/http"

	nurserycalendarapp "nursery-management-system/api/internal/modules/nursery_calendar/application"
	nurserycalendardomain "nursery-management-system/api/internal/modules/nursery_calendar/domain"
	nurserycalendarhandler "nursery-management-system/api/internal/modules/nursery_calendar/interfaces/http"

	emailscheduler "nursery-management-system/api/internal/modules/email"
	emailapp "nursery-management-system/api/internal/modules/email/application"
	emailpostgres "nursery-management-system/api/internal/modules/email/infrastructure/postgres"
	emailsmtp "nursery-management-system/api/internal/modules/email/infrastructure/smtp"
	emailhandler "nursery-management-system/api/internal/modules/email/interfaces/http"
)

// ── Auth module ─────────────────────────────────────────────────────────

var authSet = wire.NewSet(
	provideTokenManager,
	authpostgres.NewRepository,
	wire.Bind(new(authdomain.UserRepository), new(*authpostgres.Repository)),
	wire.Bind(new(authdomain.SessionRepository), new(*authpostgres.Repository)),
	wire.Bind(new(authapp.TokenGenerator), new(*authtokens.TokenManager)),
	authapp.NewLoginUseCase,
	authapp.NewRefreshUseCase,
	authapp.NewLogoutUseCase,
	authapp.NewSwitchMembershipUseCase,
	authhandler.NewHandler,
)

// Wrappers for constructors with unexported interface types or complex wiring

func provideMarkAbsent(
	repo absencedomain.Repository,
	childChecker absencedomain.ChildEnrollmentChecker,
	txMgr *transaction.Manager,
	audit *audit.Writer,
	clock *attendanceapp.AttendanceClock,
) *absenceapp.MarkAbsent {
	return absenceapp.NewMarkAbsent(repo, childChecker, txMgr, audit, clock)
}

func provideClearMarker(
	repo absencedomain.Repository,
	txMgr *transaction.Manager,
	audit *audit.Writer,
	clock *attendanceapp.AttendanceClock,
) *absenceapp.ClearMarker {
	return absenceapp.NewClearMarker(repo, txMgr, audit, clock)
}

// ── Password reset module ───────────────────────────────────────────────

var passwordResetSet = wire.NewSet(
	provideResetTokenManager,
	resetpostgres.NewRepository,
	wire.Bind(new(resetdomain.Repository), new(*resetpostgres.Repository)),
	resetapp.NewTokenGeneratorAdapter,
	wire.Bind(new(resetapp.TokenGenerator), new(*resetapp.TokenGeneratorAdapter)),
	resetapp.NewEmailAdapter,
	wire.Bind(new(resetdomain.EmailSender), new(*resetapp.EmailAdapter)),
	resetapp.NewRequestResetUseCase,
	resetapp.NewSetNewPasswordUseCase,
	provideResetHandler,
)

// ── Children module ────────────────────────────────────────────────────

func provideFileStorage() childdomain.FileStorage {
	return childstorage.NewLocalStorage(".")
}

var childrenSet = wire.NewSet(
	childpostgres.NewChildRepository,
	wire.Bind(new(childdomain.Repository), new(*childpostgres.ChildRepository)),
	provideFileStorage,
	provideSessionTypeLookupAdapter,
	wire.Bind(new(childapp.SessionTypeLookup), new(*sessionTypeLookupAdapter)),
	provideClock,
	provideTodayFunc,
	provideChildFundingWriterAdapter,
	wire.Bind(new(childdomain.ChildFundingWriter), new(*childFundingWriterAdapter)),
	provideChildFundingReaderAdapter,
	wire.Bind(new(childdomain.ChildFundingReader), new(*childFundingReaderAdapter)),
	childapp.NewListChildren,
	childapp.NewGetChild,
	childapp.NewCreateChildWithFullProfile,
	childapp.NewUpdateChild,
	childapp.NewMarkInactive,
	childapp.NewListAttendance,
	childapp.NewGetProfile,
	childapp.NewUpdateProfile,
	childapp.NewGetContacts,
	childapp.NewReplaceContacts,
	childapp.NewGetHealth,
	childapp.NewUpdateHealth,
	childapp.NewGetSafeguarding,
	childapp.NewUpdateSafeguarding,
	childapp.NewGetConsent,
	childapp.NewUpdateConsent,
	childapp.NewGetCollectionSetting,
	childapp.NewSetCollectionPassword,
	childapp.NewListRoomAssignments,
	childapp.NewCreateRoomAssignment,
	childapp.NewCloseRoomAssignment,
	childapp.NewGetBillingProfile,
	childapp.NewUpdateBillingProfile,
	childapp.NewGetLeavingRecord,
	childapp.NewUploadPhoto,
	childapp.NewRemovePhoto,
	childapp.NewGetChildFunding,
	wire.Struct(new(childhandler.CoreUseCases), "*"),
	wire.Struct(new(childhandler.ProfileUseCases), "*"),
	wire.Struct(new(childhandler.ContactsUseCases), "*"),
	wire.Struct(new(childhandler.HealthUseCases), "*"),
	wire.Struct(new(childhandler.SafeguardingUseCases), "*"),
	wire.Struct(new(childhandler.ConsentUseCases), "*"),
	wire.Struct(new(childhandler.CollectionUseCases), "*"),
	wire.Struct(new(childhandler.RoomAssignmentUseCases), "*"),
	wire.Struct(new(childhandler.BillingProfileUseCases), "*"),
	wire.Struct(new(childhandler.PhotoUseCases), "*"),
	wire.Struct(new(childhandler.FundingUseCases), "*"),
	wire.Struct(new(childhandler.ChildrenHandlerConfig), "*"),
	childhandler.NewHandler,
)

// ── Parent-Child Mappings module ───────────────────────────────────────

var parentChildMappingsSet = wire.NewSet(
	parentchildpostgres.NewParentChildMappingRepository,
	wire.Bind(new(parentchilddomain.Repository), new(*parentchildpostgres.ParentChildMappingRepository)),
	provideMembershipCheckerAdapter,
	wire.Bind(new(parentchilddomain.MembershipChecker), new(*membershipCheckerAdapter)),
	provideChildScopeCheckerAdapter,
	wire.Bind(new(parentchildapp.ChildChecker), new(*childScopeCheckerAdapter)),
	parentchildapp.NewCreateMappingUseCase,
	parentchildapp.NewEndMappingUseCase,
	parentchildhandler.NewHandler,
)

// ── Parents module ─────────────────────────────────────────────────────

func provideParentsChildExistenceCheckerAdapter(
	childRepo *childpostgres.ChildRepository,
) *parentsChildExistenceCheckerAdapter {
	return &parentsChildExistenceCheckerAdapter{repo: childRepo}
}

var parentsSet = wire.NewSet(
	parentspostgres.NewParentRepository,
	wire.Bind(new(parentsdomain.Repository), new(*parentspostgres.ParentRepository)),
	provideParentsChildExistenceCheckerAdapter,
	wire.Bind(new(parentsdomain.ChildExistenceChecker), new(*parentsChildExistenceCheckerAdapter)),
	provideParentUserCreatorAdapter,
	wire.Bind(new(parentsapp.UserCreator), new(*parentUserCreatorAdapter)),
	provideParentEmailSenderAdapter,
	wire.Bind(new(parentsapp.EmailSender), new(*parentEmailSenderAdapter)),
	parentsapp.NewCreateParentUseCase,
	parentsapp.NewUpdateParentUseCase,
	parentsapp.NewGetParentUseCase,
	parentsapp.NewListParentsUseCase,
	parentsapp.NewListParentsByChildUseCase,
	parentsapp.NewSoftDeleteParentUseCase,
	parentsapp.NewLinkChildUseCase,
	parentsapp.NewUnlinkChildUseCase,
	parentsapp.NewInviteToPortalUseCase,
	parentsapp.NewRevokePortalAccessUseCase,
	parentshandler.NewHandler,
)

// ── Attendance module ──────────────────────────────────────────────────

func provideParentChildLookupForAttendanceAdapter(
	parentChildRepo *parentchildpostgres.ParentChildMappingRepository,
) *parentChildLookupForAttendanceAdapter {
	return &parentChildLookupForAttendanceAdapter{repo: parentChildRepo}
}

var attendanceSet = wire.NewSet(
	attendancepostgres.NewAttendanceRepository,
	wire.Bind(new(attendancedomain.Repository), new(*attendancepostgres.AttendanceRepository)),
	provideChildEnrollmentCheckerAdapter,
	wire.Bind(new(attendancedomain.ChildEnrollmentChecker), new(*childEnrollmentCheckerAdapter)),
	provideChildCorrectionCheckerAdapter,
	wire.Bind(new(attendancedomain.ChildCorrectionChecker), new(*childCorrectionCheckerAdapter)),
	provideAttendanceCalendarQueryAdapter,
	wire.Bind(new(attendancedomain.CalendarQuery), new(*attendanceCalendarQueryAdapter)),
	provideAttendanceChildFundingLookupAdapter,
	wire.Bind(new(attendancedomain.ChildFundingLookup), new(*attendanceChildFundingLookupAdapter)),
	attendanceapp.NewCheckInChild,
	attendanceapp.NewCheckOutChild,
	attendanceapp.NewCorrectAttendance,
	attendanceapp.NewListCorrectionSessions,
	attendanceapp.NewListCorrectionHistory,
	attendanceapp.NewGetRegister,
	attendanceapp.NewGetRegisterSummary,
	provideParentChildLookupForAttendanceAdapter,
	wire.Bind(new(attendanceapp.ParentChildLookupForAttendance), new(*parentChildLookupForAttendanceAdapter)),
	attendanceapp.NewListParentAttendance,
	attendancehandler.NewHandler,
)

// ── Absence module ────────────────────────────────────────────────────

var absenceSet = wire.NewSet(
	absencepostgres.NewAbsenceRepository,
	wire.Bind(new(absencedomain.Repository), new(*absencepostgres.AbsenceRepository)),
	provideAbsenceMarkerCheckerAdapter,
	wire.Bind(new(attendancedomain.AbsenceMarkerChecker), new(*absenceMarkerCheckerAdapter)),
	provideMarkAbsent,
	provideClearMarker,
	absencehandler.NewHandler,
)

// ── Funding module ─────────────────────────────────────────────────────

func provideParentChildLookupForFundingAdapter(
	parentChildRepo *parentchildpostgres.ParentChildMappingRepository,
) *parentChildLookupForFundingAdapter {
	return &parentChildLookupForFundingAdapter{repo: parentChildRepo}
}

var fundingSet = wire.NewSet(
	fundingpostgres.NewRepository,
	wire.Bind(new(fundingdomain.FundingQueryRepository), new(*fundingpostgres.Repository)),
	fundingpostgres.NewHistoryRepository,
	wire.Bind(new(fundingdomain.HistoryRepository), new(*fundingpostgres.HistoryRepository)),
	fundingpostgres.NewFundingRecordRepository,
	wire.Bind(new(fundingdomain.FundingRecordRepository), new(*fundingpostgres.FundingRecordRepositoryImpl)),
	provideTermDateProviderAdapter,
	wire.Bind(new(fundingdomain.TermDateProvider), new(*termDateProviderAdapter)),
	provideConsumedMinutesProviderAdapter,
	wire.Bind(new(fundingapp.ConsumedMinutesProvider), new(*consumedMinutesProviderAdapter)),
	wire.Bind(new(fundingapp.TxManager), new(*transaction.Manager)),
	fundingapp.NewGetChildFunding,
	fundingapp.NewUpdateChildFunding,
	fundingapp.NewListOverview,
	fundingapp.NewGetEnhancedOverview,
	fundingapp.NewGetEnhancedChildDetail,
	fundingapp.NewListExpiring,
	provideParentChildLookupForFundingAdapter,
	wire.Bind(new(fundingapp.ParentChildLookupForFunding), new(*parentChildLookupForFundingAdapter)),
	fundingapp.NewGetParentFunding,
	fundingapp.NewGetParentFundingBreakdown,
	httpfunding.NewHandler,
)

// ── Billing module ─────────────────────────────────────────────────────

func provideFundingLookupAdapter(
	fundingRepo *fundingpostgres.FundingRecordRepositoryImpl,
	ownerRepo *ownerpostgres.OwnerRepository,
) *fundingLookupAdapter {
	return &fundingLookupAdapter{fundingRepo: fundingRepo, ownerRepo: ownerRepo}
}

var billingSet = wire.NewSet(
	billingpostgres.NewRepository,
	wire.Bind(new(billingdomain.BillingRepository), new(*billingpostgres.Repository)),
	wire.Bind(new(billingdomain.BranchSettingsRepository), new(*billingpostgres.Repository)),
	provideSiteProfileLookupAdapter,
	wire.Bind(new(billingapp.SiteProfileLookup), new(*siteProfileLookupAdapter)),
	provideParentContactLookupAdapter,
	wire.Bind(new(billingapp.ParentContactLookup), new(*parentContactLookupAdapter)),
	provideSiteRateUpdateAdapter,
	wire.Bind(new(billingdomain.SiteRateRepository), new(*siteRateUpdateAdapter)),
	provideTermDateLookupAdapter,
	wire.Bind(new(billingdomain.TermDateLookup), new(*termDateLookupAdapter)),
	provideAdHocBookingLookupAdapter,
	wire.Bind(new(billingdomain.AdHocBookingLookup), new(*adHocBookingLookupAdapter)),
	provideHourlyBookingLookupAdapter,
	wire.Bind(new(billingdomain.HourlyBookingLookup), new(*hourlyBookingLookupAdapter)),
	provideFundingLookupAdapter,
	wire.Bind(new(billingdomain.FundingLookup), new(*fundingLookupAdapter)),
	bookingsapp.NewBookingEntriesLookupAdapter,
	wire.Bind(new(billingdomain.BookingEntriesLookup), new(*bookingsapp.BookingEntriesLookupAdapter)),
	billingapp.NewPreflightDraftInvoices,
	billingapp.NewComputeInvoicePrefill,
	billingapp.NewCreateDraftInvoice,
	billingapp.NewCreateAndIssueInvoiceFromForm,
	billingapp.NewGenerateDraftInvoices,
	billingapp.NewListInvoices,
	billingapp.NewGetInvoice,
	billingapp.NewIssueInvoice,
	billingapp.NewBulkIssueInvoices,
	billingapp.NewOverrideAttendanceBlockUseCase,
	billingapp.NewVoidInvoice,
	billingapp.NewDeleteInvoice,
	billingapp.NewBulkDeleteInvoices,
	billingapp.NewManageInvoiceLines,
	billingapp.NewListParentInvoices,
	billingapp.NewGetParentInvoice,
	billingapp.NewUpdateSiteRateUseCase,
	billingapp.NewUpdateBranchSettingsUseCase,
	billingapp.NewExportInvoices,
	billingapp.NewInvoiceSummary,
	billingapp.NewOverdueSummary,
	provideInvoicePDFRenderer,
	wire.Bind(new(billinghandler.InvoicePDFRenderer), new(*billingpdf.Renderer)),
	wire.Struct(new(billinghandler.DraftUseCases), "*"),
	wire.Struct(new(billinghandler.LifecycleUseCases), "*"),
	wire.Struct(new(billinghandler.ParentInvoiceUseCases), "*"),
	wire.Struct(new(billinghandler.AdminUseCases), "*"),
	wire.Struct(new(billinghandler.ExportUseCases), "*"),
	wire.Struct(new(billinghandler.BillingHandlerConfig), "*"),
	billinghandler.NewHandler,
)

// ── Payments module ─────────────────────────────────────────────────────

func provideCreateCheckoutSession(
	repo *paymentspostgres.Repository,
	paymentsTxMgr *txManagerAdapter,
	checkoutProvider paymentsdomain.CheckoutProvider,
	cfg config.Config,
	logger *slog.Logger,
	recorder *metrics.Recorder,
) *paymentsapp.CreateCheckoutSession {
	stripeConfigured := checkoutProvider != nil
	uc := paymentsapp.NewCreateCheckoutSession(repo, paymentsTxMgr, checkoutProvider, cfg.WebBaseURL, stripeConfigured)
	return uc.WithObservability(logger, recorder)
}

func provideCreatePaymentLink(
	repo *paymentspostgres.Repository,
	paymentLinkProvider paymentsdomain.PaymentLinkProvider,
	logger *slog.Logger,
	recorder *metrics.Recorder,
) *paymentsapp.CreatePaymentLink {
	stripeConfigured := paymentLinkProvider != nil
	uc := paymentsapp.NewCreatePaymentLink(repo.ManagerRepo(), paymentLinkProvider, repo.PaymentLinkRepo(), stripeConfigured)
	return uc.WithObservability(logger, recorder)
}

func provideHandleStripeWebhook(
	repo *paymentspostgres.Repository,
	webhookVerifier paymentsdomain.WebhookVerifier,
	paymentsTxMgr *txManagerAdapter,
	auditWriter *auditSystemWriterAdapter,
	logger *slog.Logger,
	recorder *metrics.Recorder,
) *paymentsapp.HandleStripeWebhook {
	if webhookVerifier == nil {
		return nil
	}
	uc := paymentsapp.NewHandleStripeWebhook(repo, webhookVerifier, paymentsTxMgr, auditWriter)
	return uc.WithObservability(logger, recorder)
}

var paymentsSet = wire.NewSet(
	paymentspostgres.NewRepository,
	provideTxManagerAdapter,
	provideAuditSystemWriterAdapter,
	provideStripeClient,
	wire.Bind(new(paymentsdomain.CheckoutProvider), new(*stripeclient.Client)),
	wire.Bind(new(paymentsdomain.PaymentLinkProvider), new(*stripeclient.Client)),
	provideWebhookVerifier,
	wire.Bind(new(paymentsdomain.WebhookVerifier), new(*stripeclient.WebhookVerifier)),
	provideManagerPaymentRepo,
	provideCreateCheckoutSession,
	provideCreatePaymentLink,
	provideHandleStripeWebhook,
	paymentsapp.NewGetManagerPaymentStatus,
	paymentsapp.NewListManagerPaymentEvents,
	providePaymentsHandler,
)

// ── Invites module ──────────────────────────────────────────────────────

var invitesSet = wire.NewSet(
	provideInviteTokenManager,
	invitepostgres.NewRepository,
	wire.Bind(new(invitedomain.Repository), new(*invitepostgres.Repository)),
	inviteapp.NewTokenGeneratorAdapter,
	wire.Bind(new(inviteapp.TokenGenerator), new(*inviteapp.TokenGeneratorAdapter)),
	inviteapp.NewInviteEmailAdapter,
	wire.Bind(new(inviteapp.EmailSender), new(*inviteapp.InviteEmailAdapter)),
	wire.Bind(new(inviteapp.TokenValidator), new(*invitetokens.Manager)),
	inviteapp.NewCreateInviteUseCase,
	inviteapp.NewListInvitesUseCase,
	inviteapp.NewResendInviteUseCase,
	inviteapp.NewRevokeInviteUseCase,
	inviteapp.NewAcceptInviteUseCase,
	provideInviteIPLimiter,
	invitehandler.NewHandler,
)

func provideResetHandler(
	requestReset *resetapp.RequestResetUseCase,
	setPassword *resetapp.SetNewPasswordUseCase,
	recorder *metrics.Recorder,
	logger *slog.Logger,
) *resethandler.Handler {
	emailLimiter := ratelimit.NewFixedWindowLimiter(5, 15*time.Minute)
	ipLimiter := ratelimit.NewFixedWindowLimiter(20, 15*time.Minute)
	return resethandler.NewHandler(requestReset, setPassword, emailLimiter, ipLimiter, recorder, logger)
}

// ── Owner module ───────────────────────────────────────────────────────

func provideOwnerHandler(
	summaries *ownerapp.GetSiteSummariesUseCase,
	listAccess *ownerapp.ListManagerAccessUseCase,
	grant *ownerapp.GrantManagerAccessUseCase,
	deactivate *ownerapp.DeactivateManagerAccessUseCase,
	reactivate *ownerapp.ReactivateManagerAccessUseCase,
	updateBilling *ownerapp.UpdateSiteBillingSetupUseCase,
	recorder *metrics.Recorder,
	logger *slog.Logger,
) *ownerhandler.Handler {
	return ownerhandler.NewHandler(summaries, listAccess, grant, deactivate, reactivate, recorder, logger).
		WithUpdateBillingSetup(updateBilling)
}

var ownerSet = wire.NewSet(
	ownerpostgres.NewRepository,
	wire.Bind(new(ownerdomain.SummaryRepository), new(*ownerpostgres.OwnerRepository)),
	wire.Bind(new(ownerdomain.ManagerAccessRepository), new(*ownerpostgres.OwnerRepository)),
	ownerapp.NewGetSiteSummariesUseCase,
	ownerapp.NewListManagerAccessUseCase,
	provideInviteTokenGeneratorAdapter,
	wire.Bind(new(ownerdomain.InviteTokenGenerator), new(*inviteTokenGeneratorAdapter)),
	provideEmailSenderAdapter,
	wire.Bind(new(ownerdomain.ManagerInviteSender), new(*emailSenderAdapter)),
	ownerapp.NewGrantManagerAccessUseCase,
	ownerapp.NewDeactivateManagerAccessUseCase,
	ownerapp.NewReactivateManagerAccessUseCase,
	ownerapp.NewUpdateSiteBillingSetupUseCase,
	provideOwnerHandler,
)

// ── Rooms module ────────────────────────────────────────────────────────

var roomsSet = wire.NewSet(
	roomspostgres.NewRepository,
	wire.Bind(new(roomsdomain.Repository), new(*roomspostgres.RoomRepository)),
	roomsapp.NewCreateRoom,
	roomsapp.NewUpdateRoom,
	roomsapp.NewListRooms,
	roomsapp.NewGetRoom,
	roomsapp.NewArchiveRoom,
	roomsapp.NewReactivateRoom,
	roomshttphandler.NewHandler,
)

// ── Session types module ───────────────────────────────────────────────

var sessionTypesSet = wire.NewSet(
	sessiontypeapp.NewCreateSessionType,
	sessiontypeapp.NewUpdateSessionType,
	sessiontypeapp.NewListSessionTypes,
	sessiontypeapp.NewGetSessionType,
	sessiontypeapp.NewArchiveSessionType,
	sessiontypeapp.NewReactivateSessionType,
	sessiontypehttphandler.NewHandler,
)

// ── Session templates module ──────────────────────────────────────────

var sessionTemplatesSet = wire.NewSet(
	sessiontemplatepostgres.NewRepository,
	wire.Bind(new(sessiontemplatedomain.Repository), new(*sessiontemplatepostgres.SessionTemplateRepository)),
	provideSessionTemplateLookupTemplateAdapter,
	wire.Bind(new(sessiontemplateapp.SessionTypeLookup), new(*sessionTemplateLookupTemplateAdapter)),
	sessiontemplateapp.NewCreateSessionTemplate,
	sessiontemplateapp.NewUpdateSessionTemplate,
	sessiontemplateapp.NewListSessionTemplates,
	sessiontemplateapp.NewGetSessionTemplate,
	sessiontemplateapp.NewArchiveSessionTemplate,
	sessiontemplateapp.NewReactivateSessionTemplate,
	sessiontemplatehttphandler.NewHandler,
)

// ── Site Profile module ────────────────────────────────────────────────

func provideSiteProfileHandlerSet(
	repo siteprofiledomain.Repository,
	auditWriter *audit.Writer,
	txMgr *transaction.Manager,
	logger *slog.Logger,
) *siteprofilehandler.Handler {
	getUC := siteprofileapp.NewGetSiteProfileUseCase(repo)
	updateUC := siteprofileapp.NewUpdateSiteProfileUseCase(repo, auditWriter, txMgr)
	return siteprofilehandler.NewHandler(getUC, updateUC, logger)
}

var siteProfileSet = wire.NewSet(
	siteprofilepostgres.NewRepository,
	wire.Bind(new(siteprofiledomain.Repository), new(*siteprofilepostgres.SiteProfileRepository)),
	siteprofileapp.NewGetSiteProfileUseCase,
	siteprofileapp.NewUpdateSiteProfileUseCase,
	provideSiteProfileHandlerSet,
)

// ── Academic Term Calendar module ─────────────────────────────────────

var termCalendarSet = wire.NewSet(
	termcalendarpostgres.NewRepository,
	wire.Bind(new(termcalendardomain.Repository), new(*termcalendarpostgres.AcademicTermRepository)),
	termcalendarapp.NewCreateTerm,
	termcalendarapp.NewListTerms,
	termcalendarapp.NewUpdateTerm,
	termcalendarapp.NewArchiveTerm,
	termcalendarhttphandler.NewHandler,
)

// ── Ad-Hoc Bookings module ───────────────────────────────────────────

var adHocBookingsSet = wire.NewSet(
	adhocpostgres.NewRepository,
	wire.Bind(new(adhocdomain.Repository), new(*adhocpostgres.AdHocBookingRepository)),
	adhocapp.NewCreateAdHocBooking,
	provideAdhocCalendarQueryAdapter,
	wire.Bind(new(adhocdomain.CalendarQuery), new(*adhocCalendarQueryAdapter)),
	provideAdhocChildFundingLookupAdapter,
	wire.Bind(new(adhocdomain.ChildFundingLookup), new(*adhocChildFundingLookupAdapter)),
	adhocapp.NewListAdHocBookings,
	adhocapp.NewCancelAdHocBooking,
	adhocapp.NewUpdateAdHocBooking,
	adhochttphandler.NewHandler,
)

// ── Hourly Bookings module ───────────────────────────────────────────

var hourlyBookingsSet = wire.NewSet(
	hourlypostgres.NewRepository,
	wire.Bind(new(hourlydomain.Repository), new(*hourlypostgres.HourlyBookingRepository)),
	hourlyapp.NewCreateHourlyBooking,
	provideHourlyCalendarQueryAdapter,
	wire.Bind(new(hourlydomain.CalendarQuery), new(*hourlyCalendarQueryAdapter)),
	provideHourlyChildFundingLookupAdapter,
	wire.Bind(new(hourlydomain.ChildFundingLookup), new(*hourlyChildFundingLookupAdapter)),
	hourlyapp.NewListHourlyBookings,
	hourlyapp.NewCancelHourlyBooking,
	hourlyapp.NewUpdateHourlyBooking,
	hourlyhttphandler.NewHandler,
)

// ── Bookings module ────────────────────────────────────────────────────

func provideRoomCapacityLookupAdapter(
	roomsRepo *roomspostgres.RoomRepository,
	childRepo *childpostgres.ChildRepository,
) *roomCapacityLookupAdapter {
	return &roomCapacityLookupAdapter{repo: roomsRepo, childRepo: childRepo}
}

func provideParentChildLookupAdapter(
	parentChildRepo *parentchildpostgres.ParentChildMappingRepository,
) *parentChildLookupAdapter {
	return &parentChildLookupAdapter{repo: parentChildRepo}
}

func provideBookingsFundingLookupAdapterForWire(
	fundingRepo *fundingpostgres.FundingRecordRepositoryImpl,
) *bookingsFundingLookupAdapter {
	return provideBookingsFundingLookupAdapter(fundingRepo)
}

var bookingsSet = wire.NewSet(
	bookingspostgres.NewRepository,
	wire.Bind(new(bookingsdomain.Repository), new(*bookingspostgres.BookingRepository)),
	bookingsapp.NewCreateBooking,
	provideBookingsFundingLookupAdapterForWire,
	wire.Bind(new(bookingsdomain.FundingLookup), new(*bookingsFundingLookupAdapter)),
	provideBookingsCalendarQueryAdapter,
	wire.Bind(new(bookingsdomain.CalendarQuery), new(*bookingsCalendarQueryAdapter)),
	bookingsapp.NewGetBooking,
	bookingsapp.NewListBookings,
	bookingsapp.NewUpdateBooking,
	bookingsapp.NewCancelBooking,
	bookingsapp.NewPauseBooking,
	bookingsapp.NewCloneBooking,
	bookingsapp.NewListCapacity,
	provideRoomCapacityLookupAdapter,
	wire.Bind(new(bookingsapp.RoomCapacityLookup), new(*roomCapacityLookupAdapter)),
	provideParentChildLookupAdapter,
	wire.Bind(new(bookingsapp.ParentChildLookup), new(*parentChildLookupAdapter)),
	bookingsapp.NewListParentBookings,
	bookingsapp.NewCreateBookingRequest,
	bookingsapp.NewCancelParentBooking,
	bookingshttphandler.NewHandler,
)

// ── Branch Closures module ──────────────────────────────────────────

var branchClosuresSet = wire.NewSet(
	branchclosurepostgres.NewRepository,
	wire.Bind(new(branchclosuredomain.Repository), new(*branchclosurepostgres.Repository)),
	branchclosureapp.NewCreateClosureDay,
	branchclosureapp.NewListClosureDays,
	branchclosureapp.NewDeleteClosureDay,
	branchclosurehandler.NewHandler,
	provideClosureDateLookupAdapter,
	wire.Bind(new(billingdomain.ClosureDateLookup), new(*closureDateLookupAdapter)),
)

// ── Holiday Periods module ──────────────────────────────────────────

var holidayPeriodsSet = wire.NewSet(
	holidayperiodspostgres.NewRepository,
	wire.Bind(new(holidayperiodsdomain.Repository), new(*holidayperiodspostgres.Repository)),
	holidayperiodsapp.NewCreateHolidayPeriod,
	holidayperiodsapp.NewUpdateHolidayPeriod,
	holidayperiodsapp.NewDeleteHolidayPeriod,
	holidayperiodsapp.NewListHolidayPeriods,
	holidayperiodshttphandler.NewHandler,
	provideHolidayPeriodLookupAdapter,
	wire.Bind(new(billingdomain.HolidayPeriodLookup), new(*holidayPeriodLookupAdapter)),
)

// ── Nursery Calendar module ────────────────────────────────────────────

var nurseryCalendarSet = wire.NewSet(
	provideNurseryCalendarClosureAdapter,
	wire.Bind(new(nurserycalendardomain.ClosureDayLookup), new(*nurseryCalendarClosureAdapter)),
	provideNurseryCalendarHolidayAdapter,
	wire.Bind(new(nurserycalendardomain.HolidayPeriodLookup), new(*nurseryCalendarHolidayAdapter)),
	nurserycalendarapp.NewQueryCalendarDay,
	nurserycalendarapp.NewQueryDateRange,
	nurserycalendarhandler.NewHandler,
)

// ── Notifications module ──────────────────────────────────────────────

var notificationsSet = wire.NewSet(
	provideBillingNotificationAdapter,
	wire.Bind(new(notificationsapp.InvoiceNotificationSender), new(*billingNotificationAdapter)),
	notificationsapp.NewInvoiceIssuedHandler,
	notificationsapp.NewInvoiceOverdueHandler,
)

// ── Email module ──────────────────────────────────────────────────────

func provideEmailOutboxRepo(pool *pgxpool.Pool) *emailpostgres.OutboxRepository {
	return emailpostgres.NewOutboxRepository(pool)
}

func provideEmailProvider(sender email.Sender) *emailsmtp.Provider {
	return emailsmtp.NewProvider(sender)
}

func provideEmailRenderer() *emailapp.Renderer {
	return emailapp.NewRenderer()
}

func provideEmailScheduler(
	logger *slog.Logger,
	sendPending *emailapp.SendPendingEmails,
	recorder *metrics.Recorder,
	cfg config.Config,
) *emailscheduler.Scheduler {
	return emailscheduler.NewScheduler(logger, sendPending, recorder, cfg.Email.PollIntervalSeconds)
}

func provideEmailHandler(
	logger *slog.Logger,
	listEmails *emailapp.ListEmails,
	getEmail *emailapp.GetEmail,
	retryEmail *emailapp.RetryEmail,
	getStats *emailapp.GetEmailStats,
	repo *emailpostgres.OutboxRepository,
) *emailhandler.Handler {
	return emailhandler.NewHandler(logger, listEmails, getEmail, retryEmail, getStats, repo)
}

var emailSet = wire.NewSet(
	provideEmailOutboxRepo,
	provideEmailProvider,
	provideEmailRenderer,
	emailapp.NewEnqueueEmail,
	wire.Bind(new(emaildomain.EmailEnqueuer), new(*emailapp.EnqueueEmail)),
	emailapp.NewSendPendingEmails,
	emailapp.NewListEmails,
	emailapp.NewGetEmail,
	emailapp.NewRetryEmail,
	emailapp.NewGetEmailStats,
	provideEmailScheduler,
	provideEmailHandler,
)

// ── Injector ────────────────────────────────────────────────────────────

func InitializeApp(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) (*gin.Engine, *emailscheduler.Scheduler, error) {
	wire.Build(
		provideTxManager,
		wire.Bind(new(childapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(roomsapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(sessiontypeapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(sessiontemplateapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(termcalendarapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(adhocapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(hourlyapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(bookingsapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(parentchildapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(parentsapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(billingapp.PrefillTxManager), new(*transaction.Manager)),
		wire.Bind(new(billingapp.DraftInvoiceTxManager), new(*transaction.Manager)),
		wire.Bind(new(absencedomain.ChildEnrollmentChecker), new(*childEnrollmentCheckerAdapter)),
		provideAuditWriter,
		provideEventDispatcher,
		provideMetricsRegistry,
		provideMetricsRecorder,
		provideEmailSender,
		provideTokenParserAdapter,
		sessiontypepostgres.NewRepository,
		wire.Bind(new(sessiontypedomain.Repository), new(*sessiontypepostgres.SessionTypeRepository)),
		provideSiteExistsCheckerAdapter,
		wire.Bind(new(roomsapp.SiteExistsChecker), new(*siteExistsCheckerAdapter)),
		wire.Bind(new(sessiontypeapp.SiteExistsChecker), new(*siteExistsCheckerAdapter)),
		wire.Bind(new(sessiontemplateapp.SiteExistsChecker), new(*siteExistsCheckerAdapter)),

		provideAttendanceClock,
		provideWebBaseURL,
		authSet,
		passwordResetSet,
		childrenSet,
		parentChildMappingsSet,
		parentsSet,
		attendanceSet,
		absenceSet,
		fundingSet,
		billingSet,
		paymentsSet,
		invitesSet,
		ownerSet,
		roomsSet,
		sessionTypesSet,
		sessionTemplatesSet,
		termCalendarSet,
		adHocBookingsSet,
		hourlyBookingsSet,
		bookingsSet,
		siteProfileSet,
		branchClosuresSet,
		holidayPeriodsSet,
		nurseryCalendarSet,
		notificationsSet,
		emailSet,
		wire.Struct(new(appComponents), "*"),
		buildGinEngine,
	)
	return nil, nil, nil
}

func InitializeTestApp(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) (*gin.Engine, *emailscheduler.Scheduler, error) {
	wire.Build(
		provideTxManager,
		wire.Bind(new(childapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(roomsapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(sessiontypeapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(sessiontemplateapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(termcalendarapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(adhocapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(hourlyapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(bookingsapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(parentchildapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(parentsapp.TxManager), new(*transaction.Manager)),
		wire.Bind(new(billingapp.PrefillTxManager), new(*transaction.Manager)),
		wire.Bind(new(billingapp.DraftInvoiceTxManager), new(*transaction.Manager)),
		wire.Bind(new(absencedomain.ChildEnrollmentChecker), new(*childEnrollmentCheckerAdapter)),
		provideAuditWriter,
		provideEventDispatcher,
		provideMetricsRegistry,
		provideMetricsRecorder,
		provideEmailSender,
		provideTokenParserAdapter,
		sessiontypepostgres.NewRepository,
		wire.Bind(new(sessiontypedomain.Repository), new(*sessiontypepostgres.SessionTypeRepository)),
		provideSiteExistsCheckerAdapter,
		wire.Bind(new(roomsapp.SiteExistsChecker), new(*siteExistsCheckerAdapter)),
		wire.Bind(new(sessiontypeapp.SiteExistsChecker), new(*siteExistsCheckerAdapter)),
		wire.Bind(new(sessiontemplateapp.SiteExistsChecker), new(*siteExistsCheckerAdapter)),

		provideAttendanceClock,
		provideWebBaseURL,
		authSet,
		passwordResetSet,
		childrenSet,
		parentChildMappingsSet,
		parentsSet,
		attendanceSet,
		absenceSet,
		fundingSet,
		billingSet,
		paymentsSet,
		invitesSet,
		ownerSet,
		roomsSet,
		sessionTypesSet,
		sessionTemplatesSet,
		termCalendarSet,
		adHocBookingsSet,
		hourlyBookingsSet,
		bookingsSet,
		siteProfileSet,
		branchClosuresSet,
		holidayPeriodsSet,
		nurseryCalendarSet,
		notificationsSet,
		emailSet,
		wire.Struct(new(appComponents), "*"),
		buildGinEngine,
	)
	return nil, nil, nil
}
