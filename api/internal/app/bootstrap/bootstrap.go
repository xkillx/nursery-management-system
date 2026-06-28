package bootstrap

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	authapp "nursery-management-system/api/internal/modules/authentication/application"
	authpostgres "nursery-management-system/api/internal/modules/authentication/infrastructure/postgres"
	authtokens "nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	authhandler "nursery-management-system/api/internal/modules/authentication/interfaces/http"

	childapp "nursery-management-system/api/internal/modules/children/application"
	childpostgres "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	childhandler "nursery-management-system/api/internal/modules/children/interfaces/http"

	parentchildapp "nursery-management-system/api/internal/modules/parentchildmappings/application"
	parentchildpostgres "nursery-management-system/api/internal/modules/parentchildmappings/infrastructure/postgres"
	parentchildhandler "nursery-management-system/api/internal/modules/parentchildmappings/interfaces/http"

	attendanceapp "nursery-management-system/api/internal/modules/attendance/application"
	attendancepostgres "nursery-management-system/api/internal/modules/attendance/infrastructure/postgres"
	attendancehandler "nursery-management-system/api/internal/modules/attendance/interfaces/http"

	absenceapp "nursery-management-system/api/internal/modules/absence/application"
	absencepostgres "nursery-management-system/api/internal/modules/absence/infrastructure/postgres"
	absencehandler "nursery-management-system/api/internal/modules/absence/interfaces/http"

	inviteapp "nursery-management-system/api/internal/modules/invites/application"
	invitepostgres "nursery-management-system/api/internal/modules/invites/infrastructure/postgres"
	invitetokens "nursery-management-system/api/internal/modules/invites/infrastructure/tokens"
	invitehandler "nursery-management-system/api/internal/modules/invites/interfaces/http"

	fundingapp "nursery-management-system/api/internal/modules/funding/application"
	fundingpostgres "nursery-management-system/api/internal/modules/funding/infrastructure/postgres"
	fundinghandler "nursery-management-system/api/internal/modules/funding/interfaces/http"

	billingapp "nursery-management-system/api/internal/modules/billing/application"
	billingpostgres "nursery-management-system/api/internal/modules/billing/infrastructure/postgres"
	billinghandler "nursery-management-system/api/internal/modules/billing/interfaces/http"

	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	paymentsdomain "nursery-management-system/api/internal/modules/payments/domain"
	paymentspostgres "nursery-management-system/api/internal/modules/payments/infrastructure/postgres"
	stripeclient "nursery-management-system/api/internal/modules/payments/infrastructure/stripe"
	paymentshandler "nursery-management-system/api/internal/modules/payments/interfaces/http"

	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/email"
	"nursery-management-system/api/internal/platform/events"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/ratelimit"
	"nursery-management-system/api/internal/platform/transaction"

	resetapp "nursery-management-system/api/internal/modules/passwordreset/application"
	resetpostgres "nursery-management-system/api/internal/modules/passwordreset/infrastructure/postgres"
	resettokens "nursery-management-system/api/internal/modules/passwordreset/infrastructure/tokens"
	resethandler "nursery-management-system/api/internal/modules/passwordreset/interfaces/http"

	ownerapp "nursery-management-system/api/internal/modules/owner/application"
	ownerpostgres "nursery-management-system/api/internal/modules/owner/infrastructure/postgres"
	ownerhandler "nursery-management-system/api/internal/modules/owner/interfaces/http"

	roomsapp "nursery-management-system/api/internal/modules/rooms/application"
	roomspostgres "nursery-management-system/api/internal/modules/rooms/infrastructure/postgres"
	roomshttphandler "nursery-management-system/api/internal/modules/rooms/interfaces/http"

	sessiontypeapp "nursery-management-system/api/internal/modules/sessiontypes/application"
	sessiontypepostgres "nursery-management-system/api/internal/modules/sessiontypes/infrastructure/postgres"
	sessiontypehttphandler "nursery-management-system/api/internal/modules/sessiontypes/interfaces/http"

	sessiontemplateapp "nursery-management-system/api/internal/modules/sessiontemplates/application"
	sessiontemplatepostgres "nursery-management-system/api/internal/modules/sessiontemplates/infrastructure/postgres"
	sessiontemplatehttphandler "nursery-management-system/api/internal/modules/sessiontemplates/interfaces/http"

	billingdomain "nursery-management-system/api/internal/modules/billing/domain"
	childdomain "nursery-management-system/api/internal/modules/children/domain"
	termapp "nursery-management-system/api/internal/modules/term/application"
	termdomain "nursery-management-system/api/internal/modules/term/domain"
	termpostgres "nursery-management-system/api/internal/modules/term/infrastructure/postgres"
	termhttphandler "nursery-management-system/api/internal/modules/term/interfaces/http"
)

type BootstrapOptions struct {
	CheckoutProvider paymentsdomain.CheckoutProvider
	WebhookVerifier  paymentsdomain.WebhookVerifier
	EmailSender      email.Sender
	MetricsRegistry  *prometheus.Registry
	MetricsRecorder  *metrics.Recorder
}

func Bootstrap(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) *gin.Engine {
	return BootstrapWithOptions(cfg, logger, pool, BootstrapOptions{})
}

func BootstrapWithOptions(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool, opts BootstrapOptions) *gin.Engine {
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(httpserver.RequestIDMiddleware())

	var recorder *metrics.Recorder
	if cfg.MetricsEnabled {
		registry := opts.MetricsRegistry
		if registry == nil {
			registry = metrics.NewRegistry()
		}
		recorder = opts.MetricsRecorder
		if recorder == nil {
			recorder = metrics.NewRecorder(registry)
		}
		router.Use(httpserver.AccessLogMiddlewareWithMetrics(logger, recorder))
		router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry})))
	} else {
		router.Use(httpserver.AccessLogMiddleware(logger))
	}

	router.Use(httpserver.RecoveryMiddleware(logger))

	api := registerHealthRoutes(router, cfg.APIBasePath, pool)

	// Auth module
	tokenManager := authtokens.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLHours, cfg.JWTRefreshShortTTLHours)
	authRepo := authpostgres.NewRepository(pool)
	loginUC := authapp.NewLoginUseCase(authRepo, authRepo, tokenManager)
	refreshUC := authapp.NewRefreshUseCase(authRepo, authRepo, tokenManager)
	logoutUC := authapp.NewLogoutUseCase(authRepo, tokenManager)
	switchUC := authapp.NewSwitchMembershipUseCase(authRepo, authRepo, tokenManager)
	authHandler := authhandler.NewHandler(loginUC, refreshUC, logoutUC, switchUC, cfg, recorder, logger)
	authHandler.RegisterRoutes(api)

	// Password reset module
	resetTokenMgr := resettokens.NewManager(cfg.PasswordResetTokenSecret, cfg.PasswordResetTokenTTLMinutes)
	resetRepo := resetpostgres.NewRepository(pool)
	resetTokenGen := resetapp.NewTokenGeneratorAdapter(resetTokenMgr)
	emailSender := opts.EmailSender
	if emailSender == nil {
		emailSender = email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	}
	resetEmailAdapter := resetapp.NewEmailAdapter(emailSender)
	requestResetUC := resetapp.NewRequestResetUseCase(resetRepo, resetEmailAdapter, resetTokenGen, cfg.WebBaseURL, logger)
	setPasswordUC := resetapp.NewSetNewPasswordUseCase(resetRepo, logger)
	resetEmailLimiter := ratelimit.NewFixedWindowLimiter(5, 15*time.Minute)
	resetIPLimiter := ratelimit.NewFixedWindowLimiter(20, 15*time.Minute)
	resetHandler := resethandler.NewHandler(requestResetUC, setPasswordUC, resetEmailLimiter, resetIPLimiter, recorder, logger)
	resetHandler.RegisterRoutes(api)

	// Middleware
	tokenParser := &tokenParserAdapter{tm: tokenManager}
	protected := api.Group("")
	protected.Use(httpserver.AuthnMiddlewareWithObservability(tokenParser, logger, recorder))

	// Shared infrastructure
	txManager := transaction.NewManager(pool)
	auditWriter := audit.NewWriter()
	eventDispatcher := events.NewEventDispatcher(txManager)

	// Register no-op event handlers (ready for future side effects).
	events.Register(eventDispatcher, events.TypedHandlerFunc[childdomain.ChildDeactivated](func(ctx context.Context, tx pgx.Tx, event childdomain.ChildDeactivated) error {
		return nil
	}))
	events.Register(eventDispatcher, events.TypedHandlerFunc[billingdomain.InvoiceIssued](func(ctx context.Context, tx pgx.Tx, event billingdomain.InvoiceIssued) error {
		return nil
	}))
	events.Register(eventDispatcher, events.TypedHandlerFunc[billingdomain.InvoiceMarkedOverdue](func(ctx context.Context, tx pgx.Tx, event billingdomain.InvoiceMarkedOverdue) error {
		return nil
	}))

	// Children module
	childRepo := childpostgres.NewChildRepository(pool)

	// Session types repo (declared early so children handler can use the
	// sessionTypeLookupAdapter for booking pattern entry validation).
	sessionTypeRepo := sessiontypepostgres.NewRepository(pool)
	childConfig := childhandler.ChildrenHandlerConfig{
		Core: childhandler.CoreUseCases{
			List:           childapp.NewListChildren(childRepo),
			Get:            childapp.NewGetChild(childRepo),
			Create:         childapp.NewCreateChildWithFullProfile(childRepo, auditWriter, txManager, &sessionTypeLookupAdapter{repo: sessionTypeRepo}, func() time.Time { return time.Now().UTC() }),
			Update:         childapp.NewUpdateChild(childRepo, auditWriter, txManager),
			MarkInactive:   childapp.NewMarkInactive(childRepo, eventDispatcher, auditWriter),
			ListAttendance: childapp.NewListAttendance(childRepo, func() time.Time { return time.Now().UTC() }),
		},
		Profile: childhandler.ProfileUseCases{
			Get:    childapp.NewGetProfile(childRepo),
			Update: childapp.NewUpdateProfile(childRepo, auditWriter, txManager),
		},
		Contacts: childhandler.ContactsUseCases{
			Get:     childapp.NewGetContacts(childRepo),
			Replace: childapp.NewReplaceContacts(childRepo, auditWriter, txManager),
		},
		Health: childhandler.HealthUseCases{
			Get:    childapp.NewGetHealth(childRepo),
			Update: childapp.NewUpdateHealth(childRepo, auditWriter, txManager),
		},
		Safeguarding: childhandler.SafeguardingUseCases{
			Get:    childapp.NewGetSafeguarding(childRepo),
			Update: childapp.NewUpdateSafeguarding(childRepo, auditWriter, txManager),
		},
		Consent: childhandler.ConsentUseCases{
			Get:    childapp.NewGetConsent(childRepo),
			Update: childapp.NewUpdateConsent(childRepo, auditWriter, txManager),
		},
		Funding: childhandler.FundingUseCases{
			Get:    childapp.NewGetFunding(childRepo),
			Update: childapp.NewUpdateFunding(childRepo, auditWriter, txManager),
		},
		Collection: childhandler.CollectionUseCases{
			GetSetting:  childapp.NewGetCollectionSetting(childRepo),
			SetPassword: childapp.NewSetCollectionPassword(childRepo, auditWriter, txManager),
		},
		RoomAssignments: childhandler.RoomAssignmentUseCases{
			List:   childapp.NewListRoomAssignments(childRepo),
			Create: childapp.NewCreateRoomAssignment(childRepo, auditWriter, txManager),
			Close:  childapp.NewCloseRoomAssignment(childRepo, auditWriter, txManager),
		},
		BillingProfile: childhandler.BillingProfileUseCases{
			Get:    childapp.NewGetBillingProfile(childRepo),
			Update: childapp.NewUpdateBillingProfile(childRepo, auditWriter, txManager),
		},
		LeavingRecord: childapp.NewGetLeavingRecord(childRepo),
		BookingPatterns: childhandler.BookingPatternUseCases{
			List:       childapp.NewListBookingPatterns(childRepo),
			Get:        childapp.NewGetBookingPattern(childRepo),
			GetCurrent: childapp.NewGetCurrentBookingPattern(childRepo, func() time.Time { return time.Now().UTC() }),
			Create:     childapp.NewCreateBookingPattern(childRepo, auditWriter, txManager, &sessionTypeLookupAdapter{repo: sessionTypeRepo}, func() time.Time { return time.Now().UTC() }),
			Update:     childapp.NewUpdateBookingPattern(childRepo, auditWriter, txManager, &sessionTypeLookupAdapter{repo: sessionTypeRepo}, func() time.Time { return time.Now().UTC() }),
		},
	}
	childrenHandler := childhandler.NewHandler(childConfig, logger)

	// Parent-Child Mappings module
	mappingRepo := parentchildpostgres.NewParentChildMappingRepository(pool)
	membershipChecker := &membershipCheckerAdapter{repo: mappingRepo}
	childScopeChecker := &childScopeCheckerAdapter{repo: childRepo}
	mappingsHandler := parentchildhandler.NewHandler(
		parentchildapp.NewCreateMappingUseCase(mappingRepo, auditWriter, txManager, membershipChecker, childScopeChecker),
		parentchildapp.NewEndMappingUseCase(mappingRepo, auditWriter, txManager),
		logger,
	)

	// Attendance module
	attendanceRepo := attendancepostgres.NewAttendanceRepository(pool)
	childEnrollmentChecker := &childEnrollmentCheckerAdapter{repo: childRepo}
	childCorrectionChecker := &childCorrectionCheckerAdapter{repo: childRepo}
	attendanceClock := attendanceapp.NewAttendanceClock(attendanceapp.RealClock)

	// Absence module
	absenceRepo := absencepostgres.NewAbsenceRepository(pool)
	absenceMarkerChecker := &absenceMarkerCheckerAdapter{repo: absenceRepo}
	absenceClock := attendanceapp.NewAttendanceClock(attendanceapp.RealClock)
	markAbsentUC := absenceapp.NewMarkAbsent(absenceRepo, childEnrollmentChecker, txManager, auditWriter, absenceClock)
	clearMarkerUC := absenceapp.NewClearMarker(absenceRepo, txManager, auditWriter, absenceClock)

	attendanceHandler := attendancehandler.NewHandler(
		attendanceapp.NewCheckInChild(attendanceRepo, childEnrollmentChecker, absenceMarkerChecker, txManager, auditWriter, attendanceClock),
		attendanceapp.NewCheckOutChild(attendanceRepo, txManager, auditWriter, attendanceClock),
		attendanceapp.NewCorrectAttendance(attendanceRepo, childCorrectionChecker, txManager, auditWriter, attendanceClock),
		attendanceapp.NewListCorrectionSessions(attendanceRepo),
		attendanceapp.NewListCorrectionHistory(attendanceRepo),
		logger,
	)

	absenceHandler := absencehandler.NewHandler(markAbsentUC, clearMarkerUC, logger)

	// Register people routes
	childrenHandler.RegisterRoutes(protected)

	manager := protected.Group("")
	manager.Use(httpserver.RequireRolesWithObservability(logger, recorder, "manager"))
	mappingsHandler.RegisterRoutes(manager)

	// Register attendance routes (manager + practitioner)
	attendanceHandler.RegisterRoutes(protected)

	// Register absence routes (manager + practitioner)
	absenceHandler.RegisterRoutes(protected)

	// Funding module
	fundingRepo := fundingpostgres.NewRepository(pool)
	fundingHandler := fundinghandler.NewHandler(
		fundingapp.NewGetProfile(fundingRepo),
		fundingapp.NewUpsertProfile(fundingRepo, txManager, auditWriter),
		fundingapp.NewListOverview(fundingRepo),
		logger,
	)
	fundingHandler.RegisterRoutes(manager)

	// Billing module
	billingRepo := billingpostgres.NewRepository(pool)
	ownerRepo := ownerpostgres.NewRepository(pool)
	siteRateAdapter := &siteRateUpdateAdapter{repo: ownerRepo}
	updateSiteRateUC := billingapp.NewUpdateSiteRateUseCase(siteRateAdapter, auditWriter, txManager)
	billingCfg := billinghandler.BillingHandlerConfig{
		Drafting: billinghandler.DraftUseCases{
			Preflight:  billingapp.NewPreflightDraftInvoices(billingRepo),
			Generation: billingapp.NewGenerateDraftInvoices(billingRepo, txManager, auditWriter, logger, recorder),
		},
		Lifecycle: billinghandler.LifecycleUseCases{
			ListInvoices:          billingapp.NewListInvoices(billingRepo),
			GetInvoice:            billingapp.NewGetInvoice(billingRepo),
			IssueInvoice:          billingapp.NewIssueInvoice(billingRepo, txManager, auditWriter, eventDispatcher),
			BulkIssueInvoices:     billingapp.NewBulkIssueInvoices(billingRepo, txManager, auditWriter),
			OverrideAttendanceBlk: billingapp.NewOverrideAttendanceBlockUseCase(billingRepo, auditWriter, txManager),
		},
		Parent: billinghandler.ParentInvoiceUseCases{
			List: billingapp.NewListParentInvoices(billingRepo),
			Get:  billingapp.NewGetParentInvoice(billingRepo),
		},
		Admin: billinghandler.AdminUseCases{
			UpdateSiteRate: updateSiteRateUC,
		},
	}
	billingHandler := billinghandler.NewHandler(billingCfg, logger)
	billingHandler.RegisterRoutes(manager)

	// Parent route group
	parent := protected.Group("/parent")
	parent.Use(httpserver.RequireRolesWithObservability(logger, recorder, "parent"))
	billingHandler.RegisterParentRoutes(parent)

	// Payments module
	paymentsRepo := paymentspostgres.NewRepository(pool)
	var checkoutProvider paymentsdomain.CheckoutProvider
	stripeConfigured := cfg.StripeSecretKey != ""
	if opts.CheckoutProvider != nil {
		checkoutProvider = opts.CheckoutProvider
		stripeConfigured = true
	} else if stripeConfigured {
		checkoutProvider = stripeclient.NewClient(cfg.StripeSecretKey)
	}
	paymentsTxMgr := &txManagerAdapter{mgr: txManager}
	paymentsUC := paymentsapp.NewCreateCheckoutSession(paymentsRepo, paymentsTxMgr, checkoutProvider, cfg.WebBaseURL, stripeConfigured).WithObservability(logger, recorder)

	var webhookVerifier paymentsdomain.WebhookVerifier
	if opts.WebhookVerifier != nil {
		webhookVerifier = opts.WebhookVerifier
	} else if cfg.StripeWebhookSecret != "" {
		webhookVerifier = stripeclient.NewWebhookVerifier(cfg.StripeWebhookSecret)
	}

	var handleWebhookUC *paymentsapp.HandleStripeWebhook
	if webhookVerifier != nil {
		handleWebhookUC = paymentsapp.NewHandleStripeWebhook(
			paymentsRepo,
			webhookVerifier,
			paymentsTxMgr,
			&auditSystemWriterAdapter{w: auditWriter},
		).WithObservability(logger, recorder)
	}

	paymentsHandler := paymentshandler.NewHandler(
		paymentsUC,
		handleWebhookUC,
		paymentsapp.NewGetManagerPaymentStatus(paymentsRepo.ManagerRepo()),
		paymentsapp.NewListManagerPaymentEvents(paymentsRepo.ManagerRepo()),
		recorder,
		logger,
	)
	paymentsHandler.RegisterParentRoutes(parent)
	paymentsHandler.RegisterStripeRoutes(api)
	paymentsHandler.RegisterManagerRoutes(manager)

	// Invites module
	inviteTokenMgr := invitetokens.NewManager(cfg.InviteTokenSecret, cfg.InviteTokenTTLHours)
	inviteRepo := invitepostgres.NewRepository(pool, auditWriter)
	inviteTokenGen := inviteapp.NewTokenGeneratorAdapter(inviteTokenMgr)
	inviteEmailAdapter := inviteapp.NewInviteEmailAdapter(emailSender)
	createInviteUC := inviteapp.NewCreateInviteUseCase(inviteRepo, inviteTokenGen, inviteEmailAdapter, cfg.WebBaseURL, logger)
	listInvitesUC := inviteapp.NewListInvitesUseCase(inviteRepo)
	resendInviteUC := inviteapp.NewResendInviteUseCase(inviteRepo, inviteTokenGen, inviteEmailAdapter, cfg.WebBaseURL, logger)
	revokeInviteUC := inviteapp.NewRevokeInviteUseCase(inviteRepo, logger)
	acceptInviteUC := inviteapp.NewAcceptInviteUseCase(inviteRepo, logger)
	inviteIPLimiter := ratelimit.NewFixedWindowLimiter(10, 15*time.Minute)
	inviteHandler := invitehandler.NewHandler(createInviteUC, listInvitesUC, resendInviteUC, revokeInviteUC, acceptInviteUC, inviteTokenMgr, inviteIPLimiter, logger)
	inviteHandler.RegisterPublicRoutes(api)
	inviteHandler.RegisterManagerRoutes(manager)

	// Owner module (after invites for token infrastructure)
	// ownerRepo already declared above (before billing module)
	ownerSummariesUC := ownerapp.NewGetSiteSummariesUseCase(ownerRepo)
	ownerListAccessUC := ownerapp.NewListManagerAccessUseCase(ownerRepo)
	ownerTokenAdapter := &inviteTokenGeneratorAdapter{gen: inviteTokenMgr}
	ownerEmailAdapter := &emailSenderAdapter{sender: emailSender, baseURL: cfg.WebBaseURL}
	ownerGrantUC := ownerapp.NewGrantManagerAccessUseCase(ownerRepo, ownerTokenAdapter, ownerEmailAdapter, cfg.WebBaseURL)
	ownerDeactivateUC := ownerapp.NewDeactivateManagerAccessUseCase(ownerRepo)
	ownerReactivateUC := ownerapp.NewReactivateManagerAccessUseCase(ownerRepo)
	ownerUpdateBillingSetupUC := ownerapp.NewUpdateSiteBillingSetupUseCase(ownerRepo, auditWriter, txManager)
	ownerHandler := ownerhandler.NewHandler(ownerSummariesUC, ownerListAccessUC, ownerGrantUC, ownerDeactivateUC, ownerReactivateUC, recorder, logger).
		WithUpdateBillingSetup(ownerUpdateBillingSetupUC)
	owner := protected.Group("/owner")
	owner.Use(httpserver.RequireRolesWithObservability(logger, recorder, "owner"))
	ownerHandler.RegisterRoutes(owner)

	// Rooms module
	roomsSiteChecker := &siteExistsCheckerAdapter{repo: ownerRepo}
	roomsRepo := roomspostgres.NewRepository(pool)
	roomsCreateUC := roomsapp.NewCreateRoom(roomsRepo, roomsSiteChecker)
	roomsUpdateUC := roomsapp.NewUpdateRoom(roomsRepo, roomsSiteChecker)
	roomsListUC := roomsapp.NewListRooms(roomsRepo)
	roomsGetUC := roomsapp.NewGetRoom(roomsRepo)
	roomsArchiveUC := roomsapp.NewArchiveRoom(roomsRepo, txManager, auditWriter)
	roomsReactivateUC := roomsapp.NewReactivateRoom(roomsRepo, txManager, auditWriter)
	roomsHandler := roomshttphandler.NewHandler(roomsCreateUC, roomsUpdateUC, roomsListUC, roomsGetUC, roomsArchiveUC, roomsReactivateUC, logger)
	roomsHandler.RegisterRoutes(protected)

	// Session types module
	sessionTypesSiteChecker := &siteExistsCheckerAdapter{repo: ownerRepo}
	sessionTypesCreateUC := sessiontypeapp.NewCreateSessionType(sessionTypeRepo, sessionTypesSiteChecker, txManager, auditWriter)
	sessionTypesUpdateUC := sessiontypeapp.NewUpdateSessionType(sessionTypeRepo, sessionTypesSiteChecker, txManager, auditWriter)
	sessionTypesListUC := sessiontypeapp.NewListSessionTypes(sessionTypeRepo)
	sessionTypesGetUC := sessiontypeapp.NewGetSessionType(sessionTypeRepo)
	sessionTypesArchiveUC := sessiontypeapp.NewArchiveSessionType(sessionTypeRepo, txManager, auditWriter)
	sessionTypesReactivateUC := sessiontypeapp.NewReactivateSessionType(sessionTypeRepo, txManager, auditWriter)
	sessionTypesHandler := sessiontypehttphandler.NewHandler(sessionTypesCreateUC, sessionTypesUpdateUC, sessionTypesListUC, sessionTypesGetUC, sessionTypesArchiveUC, sessionTypesReactivateUC, logger)
	sessionTypesHandler.RegisterRoutes(protected)

	// Session templates module
	sessionTemplateRepo := sessiontemplatepostgres.NewRepository(pool)
	sessionTemplatesSiteChecker := &siteExistsCheckerAdapter{repo: ownerRepo}
	sessionTypeLookup := &sessionTemplateLookupTemplateAdapter{inner: &sessionTypeLookupAdapter{repo: sessionTypeRepo}}
	sessionTemplatesCreateUC := sessiontemplateapp.NewCreateSessionTemplate(sessionTemplateRepo, sessionTemplatesSiteChecker, sessionTypeLookup, txManager, auditWriter)
	sessionTemplatesUpdateUC := sessiontemplateapp.NewUpdateSessionTemplate(sessionTemplateRepo, sessionTemplatesSiteChecker, sessionTypeLookup, txManager, auditWriter)
	sessionTemplatesListUC := sessiontemplateapp.NewListSessionTemplates(sessionTemplateRepo)
	sessionTemplatesGetUC := sessiontemplateapp.NewGetSessionTemplate(sessionTemplateRepo)
	sessionTemplatesArchiveUC := sessiontemplateapp.NewArchiveSessionTemplate(sessionTemplateRepo, txManager, auditWriter)
	sessionTemplatesReactivateUC := sessiontemplateapp.NewReactivateSessionTemplate(sessionTemplateRepo, txManager, auditWriter)
	sessionTemplatesHandler := sessiontemplatehttphandler.NewHandler(sessionTemplatesCreateUC, sessionTemplatesUpdateUC, sessionTemplatesListUC, sessionTemplatesGetUC, sessionTemplatesArchiveUC, sessionTemplatesReactivateUC, logger)
	sessionTemplatesHandler.RegisterRoutes(protected)

	// Term module
	termRepo := termpostgres.NewTermRepository(pool)
	scheduleChangeRepo := termpostgres.NewScheduleChangeRepository(pool)
	bookingPatternLookup := &bookingPatternLookupAdapter{repo: childRepo}
	siteRateProvider := &siteRateProviderAdapter{repo: ownerRepo}

	createTermUC := termapp.NewCreateTermUseCase(termRepo, txManager, auditWriter, bookingPatternLookup, siteRateProvider)
	getTermUC := termapp.NewGetTermUseCase(termRepo)
	getCurrentTermUC := termapp.NewGetCurrentTermForChildUseCase(termRepo)
	listTermsUC := termapp.NewListTermsForChildUseCase(termRepo)
	listExpiringUC := termapp.NewListExpiringTermsUseCase(termRepo)
	requestChangeUC := termapp.NewRequestScheduleChangeUseCase(termRepo, scheduleChangeRepo, txManager, auditWriter, bookingPatternLookup)
	approveChangeUC := termapp.NewApproveScheduleChangeUseCase(scheduleChangeRepo, auditWriter, txManager)
	rejectChangeUC := termapp.NewRejectScheduleChangeUseCase(scheduleChangeRepo, auditWriter, txManager)
	terminateUC := termapp.NewTerminateTermUseCase(termRepo, txManager, auditWriter)
	expireTermsUC := termapp.NewExpireTermsUseCase(termRepo, auditWriter, txManager).
		WithDeactivator(&childDeactivatorAdapter{markInactiveUC: childapp.NewMarkInactive(childRepo, eventDispatcher, auditWriter)})
	markPendingRenewalUC := termapp.NewMarkPendingRenewalUseCase(termRepo, auditWriter, txManager)
	_ = expireTermsUC
	_ = markPendingRenewalUC
	_ = termdomain.ErrTermAlreadyExists
	termCfg := termhttphandler.TermHandlerConfig{
		Core: termhttphandler.CoreTermUseCases{
			Create:       createTermUC,
			Get:          getTermUC,
			GetCurrent:   getCurrentTermUC,
			List:         listTermsUC,
			ListExpiring: listExpiringUC,
			Terminate:    terminateUC,
		},
		Changes: termhttphandler.ScheduleChangeUseCases{
			Request: requestChangeUC,
			Approve: approveChangeUC,
			Reject:  rejectChangeUC,
		},
	}
	termHandler := termhttphandler.NewHandler(termCfg, logger)
	termHandler.RegisterManagerRoutes(manager)

	return router
}

type healthPinger interface {
	Ping(context.Context) error
}

type txManagerAdapter struct {
	mgr *transaction.Manager
}

func (a *txManagerAdapter) ExecTx(ctx context.Context, fn func(tx paymentsdomain.Tx) error) error {
	return a.mgr.ExecTx(ctx, func(tx pgx.Tx) error { return fn(tx) })
}

func registerHealthRoutes(router *gin.Engine, basePath string, pinger healthPinger) *gin.RouterGroup {
	router.GET("/health", healthHandler(pinger))
	api := router.Group(basePath)
	api.GET("/health", healthHandler(pinger))
	return api
}

func healthHandler(pinger healthPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := pinger.Ping(ctx); err != nil {
			httpserver.WriteError(c, http.StatusServiceUnavailable, "db_unavailable", "Database is unavailable.", nil)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": httpserver.RequestIDFromContext(c),
		})
	}
}

type auditSystemWriterAdapter struct {
	w *audit.Writer
}

func (a *auditSystemWriterAdapter) WriteSystemWithTx(ctx context.Context, tx paymentsdomain.Tx, tenantID, branchID uuid.UUID, requestID string, params paymentsapp.SystemAuditParams) error {
	var reasonCode *string
	if params.ReasonCode != nil {
		reasonCode = params.ReasonCode
	}
	return a.w.WriteSystemWithTx(ctx, tx.(pgx.Tx), tenantID, branchID, requestID, audit.WriteParams{
		ActionType: params.ActionType,
		EntityType: params.EntityType,
		EntityID:   params.EntityID,
		ReasonCode: reasonCode,
		Details:    params.Details,
	})
}
