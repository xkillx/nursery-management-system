package bootstrap

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/email"
	"nursery-management-system/api/internal/platform/events"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/metrics"
	"nursery-management-system/api/internal/platform/ratelimit"
	"nursery-management-system/api/internal/platform/transaction"

	absencepostgres "nursery-management-system/api/internal/modules/absence/infrastructure/postgres"
	absencehandler "nursery-management-system/api/internal/modules/absence/interfaces/http"
	attendanceapp "nursery-management-system/api/internal/modules/attendance/application"
	attendancehandler "nursery-management-system/api/internal/modules/attendance/interfaces/http"
	authtokens "nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	authhandler "nursery-management-system/api/internal/modules/authentication/interfaces/http"
	billingdomain "nursery-management-system/api/internal/modules/billing/domain"
	billinghandler "nursery-management-system/api/internal/modules/billing/interfaces/http"
	childapp "nursery-management-system/api/internal/modules/children/application"
	childdomain "nursery-management-system/api/internal/modules/children/domain"
	childpostgres "nursery-management-system/api/internal/modules/children/infrastructure/postgres"
	childhandler "nursery-management-system/api/internal/modules/children/interfaces/http"
	fundinghandler "nursery-management-system/api/internal/modules/funding/interfaces/http"
	invitetokens "nursery-management-system/api/internal/modules/invites/infrastructure/tokens"
	invitehandler "nursery-management-system/api/internal/modules/invites/interfaces/http"
	ownerpostgres "nursery-management-system/api/internal/modules/owner/infrastructure/postgres"
	ownerhandler "nursery-management-system/api/internal/modules/owner/interfaces/http"
	parentchildpostgres "nursery-management-system/api/internal/modules/parentchildmappings/infrastructure/postgres"
	parentchildhandler "nursery-management-system/api/internal/modules/parentchildmappings/interfaces/http"
	resettokens "nursery-management-system/api/internal/modules/passwordreset/infrastructure/tokens"
	resethandler "nursery-management-system/api/internal/modules/passwordreset/interfaces/http"
	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	paymentsdomain "nursery-management-system/api/internal/modules/payments/domain"
	paymentspostgres "nursery-management-system/api/internal/modules/payments/infrastructure/postgres"
	stripeclient "nursery-management-system/api/internal/modules/payments/infrastructure/stripe"
	paymentshandler "nursery-management-system/api/internal/modules/payments/interfaces/http"
	roomshttphandler "nursery-management-system/api/internal/modules/rooms/interfaces/http"
	sessiontemplatehttphandler "nursery-management-system/api/internal/modules/sessiontemplates/interfaces/http"
	sessiontypepostgres "nursery-management-system/api/internal/modules/sessiontypes/infrastructure/postgres"
	sessiontypehttphandler "nursery-management-system/api/internal/modules/sessiontypes/interfaces/http"
	termhttphandler "nursery-management-system/api/internal/modules/term/interfaces/http"
)

// ── Shared infrastructure providers ──────────────────────────────────────

func provideTxManager(pool *pgxpool.Pool) *transaction.Manager {
	return transaction.NewManager(pool)
}

func provideAuditWriter() *audit.Writer {
	return audit.NewWriter()
}

func provideEventDispatcher(txMgr *transaction.Manager) *events.EventDispatcher {
	d := events.NewEventDispatcher(txMgr)
	events.Register(d, events.TypedHandlerFunc[childdomain.ChildDeactivated](func(ctx context.Context, tx pgx.Tx, event childdomain.ChildDeactivated) error {
		return nil
	}))
	events.Register(d, events.TypedHandlerFunc[billingdomain.InvoiceIssued](func(ctx context.Context, tx pgx.Tx, event billingdomain.InvoiceIssued) error {
		return nil
	}))
	events.Register(d, events.TypedHandlerFunc[billingdomain.InvoiceMarkedOverdue](func(ctx context.Context, tx pgx.Tx, event billingdomain.InvoiceMarkedOverdue) error {
		return nil
	}))
	return d
}

// ── Conditional providers ────────────────────────────────────────────────

func provideMetricsRegistry(cfg config.Config) *prometheus.Registry {
	if !cfg.MetricsEnabled {
		return nil
	}
	return metrics.NewRegistry()
}

func provideMetricsRecorder(cfg config.Config, registry *prometheus.Registry) *metrics.Recorder {
	if !cfg.MetricsEnabled || registry == nil {
		return nil
	}
	return metrics.NewRecorder(registry)
}

func provideEmailSender(cfg config.Config) email.Sender {
	if cfg.SMTPHost != "" {
		return email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	}
	return nil
}

func provideStripeClient(cfg config.Config) *stripeclient.Client {
	if cfg.StripeSecretKey == "" {
		return nil
	}
	return stripeclient.NewClient(cfg.StripeSecretKey)
}

func provideWebhookVerifier(cfg config.Config) *stripeclient.WebhookVerifier {
	if cfg.StripeWebhookSecret == "" {
		return nil
	}
	return stripeclient.NewWebhookVerifier(cfg.StripeWebhookSecret)
}

// ── Clock & time providers ───────────────────────────────────────────────

func provideClock() func() time.Time {
	return func() time.Time { return time.Now().UTC() }
}

func provideTodayFunc() childapp.TodayFunc {
	return func() time.Time { return time.Now().UTC() }
}

func provideAttendanceClock() *attendanceapp.AttendanceClock {
	return attendanceapp.NewAttendanceClock(attendanceapp.RealClock)
}

func provideInviteIPLimiter() *ratelimit.FixedWindowLimiter {
	return ratelimit.NewFixedWindowLimiter(10, 15*time.Minute)
}

// ── Adapter providers ────────────────────────────────────────────────────

func provideMembershipCheckerAdapter(repo *parentchildpostgres.ParentChildMappingRepository) *membershipCheckerAdapter {
	return &membershipCheckerAdapter{repo: repo}
}

func provideChildScopeCheckerAdapter(repo *childpostgres.ChildRepository) *childScopeCheckerAdapter {
	return &childScopeCheckerAdapter{repo: repo}
}

func provideChildEnrollmentCheckerAdapter(repo *childpostgres.ChildRepository) *childEnrollmentCheckerAdapter {
	return &childEnrollmentCheckerAdapter{repo: repo}
}

func provideChildCorrectionCheckerAdapter(repo *childpostgres.ChildRepository) *childCorrectionCheckerAdapter {
	return &childCorrectionCheckerAdapter{repo: repo}
}

func provideAbsenceMarkerCheckerAdapter(repo *absencepostgres.AbsenceRepository) *absenceMarkerCheckerAdapter {
	return &absenceMarkerCheckerAdapter{repo: repo}
}

func provideSiteRateUpdateAdapter(repo *ownerpostgres.OwnerRepository) *siteRateUpdateAdapter {
	return &siteRateUpdateAdapter{repo: repo}
}

func provideSiteExistsCheckerAdapter(repo *ownerpostgres.OwnerRepository) *siteExistsCheckerAdapter {
	return &siteExistsCheckerAdapter{repo: repo}
}

func provideSessionTypeLookupAdapter(repo *sessiontypepostgres.SessionTypeRepository) *sessionTypeLookupAdapter {
	return &sessionTypeLookupAdapter{repo: repo}
}

func provideSessionTemplateLookupTemplateAdapter(inner *sessionTypeLookupAdapter) *sessionTemplateLookupTemplateAdapter {
	return &sessionTemplateLookupTemplateAdapter{inner: inner}
}

func provideBookingPatternLookupAdapter(repo *childpostgres.ChildRepository) *bookingPatternLookupAdapter {
	return &bookingPatternLookupAdapter{repo: repo}
}

func provideSiteRateProviderAdapter(repo *ownerpostgres.OwnerRepository) *siteRateProviderAdapter {
	return &siteRateProviderAdapter{repo: repo}
}

func provideChildDeactivatorAdapter(markInactive *childapp.MarkInactive) *childDeactivatorAdapter {
	return &childDeactivatorAdapter{markInactiveUC: markInactive}
}

func provideTxManagerAdapter(mgr *transaction.Manager) *txManagerAdapter {
	return &txManagerAdapter{mgr: mgr}
}

func provideAuditSystemWriterAdapter(w *audit.Writer) *auditSystemWriterAdapter {
	return &auditSystemWriterAdapter{w: w}
}

func provideManagerPaymentRepo(repo *paymentspostgres.Repository) paymentsdomain.ManagerPaymentRepository {
	return repo.ManagerRepo()
}

func providePaymentsHandler(
	checkoutSession *paymentsapp.CreateCheckoutSession,
	handleWebhook *paymentsapp.HandleStripeWebhook,
	getStatus *paymentsapp.GetManagerPaymentStatus,
	listEvents *paymentsapp.ListManagerPaymentEvents,
	recorder *metrics.Recorder,
	logger *slog.Logger,
) *paymentshandler.Handler {
	return paymentshandler.NewHandler(checkoutSession, handleWebhook, getStatus, listEvents, recorder, logger)
}

// ── Token parser adapter ───────────────────────────────────────────────

func provideTokenParserAdapter(tm *authtokens.TokenManager) httpserver.TokenParser {
	return &tokenParserAdapter{tm: tm}
}

// ── Config-derived value providers ─────────────────────────────────────

func provideTokenManager(cfg config.Config) *authtokens.TokenManager {
	return authtokens.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLHours, cfg.JWTRefreshShortTTLHours)
}

func provideResetTokenManager(cfg config.Config) *resettokens.Manager {
	return resettokens.NewManager(cfg.PasswordResetTokenSecret, cfg.PasswordResetTokenTTLMinutes)
}

func provideInviteTokenManager(cfg config.Config) *invitetokens.Manager {
	return invitetokens.NewManager(cfg.InviteTokenSecret, cfg.InviteTokenTTLHours)
}

func provideWebBaseURL(cfg config.Config) string {
	return cfg.WebBaseURL
}

func provideInviteTokenGeneratorAdapter(gen *invitetokens.Manager) *inviteTokenGeneratorAdapter {
	return &inviteTokenGeneratorAdapter{gen: gen}
}

func provideEmailSenderAdapter(sender email.Sender, cfg config.Config) *emailSenderAdapter {
	return &emailSenderAdapter{sender: sender, baseURL: cfg.WebBaseURL}
}

// ── App assembly ───────────────────────────────────────────────────────

type appComponents struct {
	Logger      *slog.Logger
	Config      config.Config
	Pool        *pgxpool.Pool
	Recorder    *metrics.Recorder
	Registry    *prometheus.Registry
	TokenParser httpserver.TokenParser

	AuthHandler             *authhandler.Handler
	ResetHandler            *resethandler.Handler
	ChildrenHandler         *childhandler.Handler
	MappingsHandler         *parentchildhandler.Handler
	AttendanceHandler       *attendancehandler.Handler
	AbsenceHandler          *absencehandler.Handler
	FundingHandler          *fundinghandler.Handler
	BillingHandler          *billinghandler.Handler
	PaymentsHandler         *paymentshandler.Handler
	InviteHandler           *invitehandler.Handler
	OwnerHandler            *ownerhandler.Handler
	RoomsHandler            *roomshttphandler.Handler
	SessionTypesHandler     *sessiontypehttphandler.Handler
	SessionTemplatesHandler *sessiontemplatehttphandler.Handler
	TermHandler             *termhttphandler.Handler
}

func buildGinEngine(c appComponents) *gin.Engine {
	if c.Config.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(httpserver.RequestIDMiddleware())

	if c.Recorder != nil {
		router.Use(httpserver.AccessLogMiddlewareWithMetrics(c.Logger, c.Recorder))
		if c.Registry != nil {
			router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(c.Registry, promhttp.HandlerOpts{Registry: c.Registry})))
		}
	} else {
		router.Use(httpserver.AccessLogMiddleware(c.Logger))
	}

	router.Use(httpserver.RecoveryMiddleware(c.Logger))

	api := registerHealthRoutes(router, c.Config.APIBasePath, c.Pool)

	c.AuthHandler.RegisterRoutes(api)
	c.ResetHandler.RegisterRoutes(api)

	protected := api.Group("")
	protected.Use(httpserver.AuthnMiddlewareWithObservability(c.TokenParser, c.Logger, c.Recorder))

	c.ChildrenHandler.RegisterRoutes(protected)

	manager := protected.Group("")
	manager.Use(httpserver.RequireRolesWithObservability(c.Logger, c.Recorder, "manager"))
	c.MappingsHandler.RegisterRoutes(manager)

	c.AttendanceHandler.RegisterRoutes(protected)
	c.AbsenceHandler.RegisterRoutes(protected)
	c.FundingHandler.RegisterRoutes(manager)
	c.BillingHandler.RegisterRoutes(manager)

	parent := protected.Group("/parent")
	parent.Use(httpserver.RequireRolesWithObservability(c.Logger, c.Recorder, "parent"))
	c.BillingHandler.RegisterParentRoutes(parent)

	c.PaymentsHandler.RegisterParentRoutes(parent)
	c.PaymentsHandler.RegisterStripeRoutes(api)
	c.PaymentsHandler.RegisterManagerRoutes(manager)

	c.InviteHandler.RegisterPublicRoutes(api)
	c.InviteHandler.RegisterManagerRoutes(manager)

	ownerGroup := protected.Group("/owner")
	ownerGroup.Use(httpserver.RequireRolesWithObservability(c.Logger, c.Recorder, "owner"))
	c.OwnerHandler.RegisterRoutes(ownerGroup)

	c.RoomsHandler.RegisterRoutes(protected)
	c.SessionTypesHandler.RegisterRoutes(protected)
	c.SessionTemplatesHandler.RegisterRoutes(protected)
	c.TermHandler.RegisterManagerRoutes(manager)

	return router
}
