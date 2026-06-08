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

	guardianapp "nursery-management-system/api/internal/modules/guardians/application"
	guardianpostgres "nursery-management-system/api/internal/modules/guardians/infrastructure/postgres"
	guardianhandler "nursery-management-system/api/internal/modules/guardians/interfaces/http"

	linkapp "nursery-management-system/api/internal/modules/guardianlinks/application"
	linkpostgres "nursery-management-system/api/internal/modules/guardianlinks/infrastructure/postgres"
	linkhandler "nursery-management-system/api/internal/modules/guardianlinks/interfaces/http"

	mappingapp "nursery-management-system/api/internal/modules/parentmappings/application"
	mappingpostgres "nursery-management-system/api/internal/modules/parentmappings/infrastructure/postgres"
	mappinghandler "nursery-management-system/api/internal/modules/parentmappings/interfaces/http"

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
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/ratelimit"
	"nursery-management-system/api/internal/platform/transaction"

	resetapp "nursery-management-system/api/internal/modules/passwordreset/application"
	resetpostgres "nursery-management-system/api/internal/modules/passwordreset/infrastructure/postgres"
	resettokens "nursery-management-system/api/internal/modules/passwordreset/infrastructure/tokens"
	resethandler "nursery-management-system/api/internal/modules/passwordreset/interfaces/http"
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
	tokenManager := authtokens.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLHours)
	authRepo := authpostgres.NewRepository(pool)
	loginUC := authapp.NewLoginUseCase(authRepo, authRepo, tokenManager)
	refreshUC := authapp.NewRefreshUseCase(authRepo, authRepo, tokenManager)
	logoutUC := authapp.NewLogoutUseCase(authRepo, tokenManager)
	switchUC := authapp.NewSwitchMembershipUseCase(authRepo, authRepo, tokenManager)
	authHandler := authhandler.NewHandler(loginUC, refreshUC, logoutUC, switchUC, cfg).WithObservability(logger, recorder)
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
	resetHandler := resethandler.NewHandler(requestResetUC, setPasswordUC, resetEmailLimiter, resetIPLimiter).WithObservability(logger, recorder)
	resetHandler.RegisterRoutes(api)

	// Middleware
	tokenParser := &tokenParserAdapter{tm: tokenManager}
	protected := api.Group("")
	protected.Use(httpserver.AuthnMiddlewareWithObservability(tokenParser, logger, recorder))

	// Shared infrastructure
	txManager := transaction.NewManager(pool)
	auditWriter := audit.NewWriter()

	// Children module
	childRepo := childpostgres.NewChildRepository(pool)
	childrenHandler := childhandler.NewHandler(
		childapp.NewListChildren(childRepo),
		childapp.NewGetChild(childRepo),
		childapp.NewCreateChild(childRepo, auditWriter, pool),
		childapp.NewUpdateChild(childRepo, auditWriter, pool),
		childapp.NewMarkInactive(childRepo, txManager, auditWriter),
		childapp.NewListAttendance(childRepo, func() time.Time { return time.Now().UTC() }),
	)

	// Guardians module
	guardianRepo := guardianpostgres.NewGuardianRepository(pool)
	guardiansHandler := guardianhandler.NewHandler(
		guardianapp.NewListGuardians(guardianRepo),
		guardianapp.NewGetGuardian(guardianRepo),
		guardianapp.NewCreateGuardian(guardianRepo, auditWriter, pool),
		guardianapp.NewUpdateGuardian(guardianRepo, auditWriter, pool),
		guardianapp.NewDeactivateGuardian(guardianRepo, txManager, auditWriter),
		guardianapp.NewReactivateGuardian(guardianRepo, txManager, auditWriter),
	)

	// Guardian-Child Links module
	linkRepo := linkpostgres.NewGuardianChildLinkRepository(pool)
	guardianChecker := &guardianCheckerAdapter{repo: guardianRepo}
	childChecker := &childCheckerAdapter{repo: childRepo}
	linksHandler := linkhandler.NewHandler(
		linkapp.NewCreateLinkUseCase(linkRepo, auditWriter, txManager, guardianChecker, childChecker),
		linkapp.NewEndLinkUseCase(linkRepo, auditWriter, txManager),
		linkapp.NewListChildLinksUseCase(linkRepo, txManager, childChecker),
	)

	// Parent Mappings module
	mappingRepo := mappingpostgres.NewParentMappingRepository(pool)
	membershipChecker := &membershipCheckerAdapter{repo: mappingRepo}
	mappingsHandler := mappinghandler.NewHandler(
		mappingapp.NewCreateMappingUseCase(mappingRepo, auditWriter, txManager, membershipChecker, guardianChecker),
		mappingapp.NewEndMappingUseCase(mappingRepo, auditWriter, txManager),
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
	)

	absenceHandler := absencehandler.NewHandler(markAbsentUC, clearMarkerUC)

	// Register people routes
	childrenHandler.RegisterRoutes(protected)

	manager := protected.Group("")
	manager.Use(httpserver.RequireRolesWithObservability(logger, recorder, "manager"))
	guardiansHandler.RegisterRoutes(manager)
	linksHandler.RegisterRoutes(manager)
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
	)
	fundingHandler.RegisterRoutes(manager)

	// Billing module
	billingRepo := billingpostgres.NewRepository(pool)
	billingHandler := billinghandler.NewHandler(
		billingapp.NewPreflightDraftInvoices(billingRepo),
		billingapp.NewGenerateDraftInvoices(billingRepo, txManager, auditWriter).WithObservability(logger, recorder),
		billingapp.NewListInvoices(billingRepo),
		billingapp.NewGetInvoice(billingRepo),
		billingapp.NewIssueInvoice(billingRepo, txManager, auditWriter),
		billingapp.NewBulkIssueInvoices(billingRepo, txManager, auditWriter),
		billingapp.NewListParentInvoices(billingRepo),
		billingapp.NewGetParentInvoice(billingRepo),
	)
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
	).WithObservability(logger, recorder)
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
	inviteHandler := invitehandler.NewHandler(createInviteUC, listInvitesUC, resendInviteUC, revokeInviteUC, acceptInviteUC, inviteTokenMgr, inviteIPLimiter)
	inviteHandler.RegisterPublicRoutes(api)
	inviteHandler.RegisterManagerRoutes(manager)

	return router
}

type healthPinger interface {
	Ping(context.Context) error
}

type txManagerAdapter struct {
	mgr *transaction.Manager
}

func (a *txManagerAdapter) ExecTx(ctx context.Context, fn func(tx paymentsdomain.Tx) error) error {
	return a.mgr.ExecTx(ctx, fn)
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

func (a *auditSystemWriterAdapter) WriteSystemWithTx(ctx context.Context, tx pgx.Tx, tenantID, branchID uuid.UUID, requestID string, params paymentsapp.SystemAuditParams) error {
	var reasonCode *string
	if params.ReasonCode != nil {
		reasonCode = params.ReasonCode
	}
	return a.w.WriteSystemWithTx(ctx, tx, tenantID, branchID, requestID, audit.WriteParams{
		ActionType: params.ActionType,
		EntityType: params.EntityType,
		EntityID:   params.EntityID,
		ReasonCode: reasonCode,
		Details:    params.Details,
	})
}
