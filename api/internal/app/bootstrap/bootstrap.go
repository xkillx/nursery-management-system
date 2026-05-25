package bootstrap

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

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

	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/email"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/ratelimit"
	"nursery-management-system/api/internal/platform/transaction"

	resetapp "nursery-management-system/api/internal/modules/passwordreset/application"
	resetpostgres "nursery-management-system/api/internal/modules/passwordreset/infrastructure/postgres"
	resettokens "nursery-management-system/api/internal/modules/passwordreset/infrastructure/tokens"
	resethandler "nursery-management-system/api/internal/modules/passwordreset/interfaces/http"
)

func Bootstrap(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) *gin.Engine {
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(httpserver.RequestIDMiddleware())
	router.Use(httpserver.AccessLogMiddleware(logger))
	router.Use(httpserver.RecoveryMiddleware(logger))

	api := registerHealthRoutes(router, cfg.APIBasePath, pool)

	// Auth module
	tokenManager := authtokens.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTLMin, cfg.JWTRefreshTTLHours)
	authRepo := authpostgres.NewRepository(pool)
	loginUC := authapp.NewLoginUseCase(authRepo, authRepo, tokenManager)
	refreshUC := authapp.NewRefreshUseCase(authRepo, authRepo, tokenManager)
	logoutUC := authapp.NewLogoutUseCase(authRepo, tokenManager)
	switchUC := authapp.NewSwitchMembershipUseCase(authRepo, authRepo, tokenManager)
	authHandler := authhandler.NewHandler(loginUC, refreshUC, logoutUC, switchUC, cfg)
	authHandler.RegisterRoutes(api)

	// Password reset module
	resetTokenMgr := resettokens.NewManager(cfg.PasswordResetTokenSecret, cfg.PasswordResetTokenTTLMinutes)
	resetRepo := resetpostgres.NewRepository(pool)
	resetTokenGen := resetapp.NewTokenGeneratorAdapter(resetTokenMgr)
	smtpSender := email.NewSMTPSender(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	resetEmailAdapter := resetapp.NewEmailAdapter(smtpSender)
	requestResetUC := resetapp.NewRequestResetUseCase(resetRepo, resetEmailAdapter, resetTokenGen, cfg.WebBaseURL, logger)
	setPasswordUC := resetapp.NewSetNewPasswordUseCase(resetRepo, logger)
	resetEmailLimiter := ratelimit.NewFixedWindowLimiter(5, 15*time.Minute)
	resetIPLimiter := ratelimit.NewFixedWindowLimiter(20, 15*time.Minute)
	resetHandler := resethandler.NewHandler(requestResetUC, setPasswordUC, resetEmailLimiter, resetIPLimiter)
	resetHandler.RegisterRoutes(api)

	// Middleware
	tokenParser := &tokenParserAdapter{tm: tokenManager}
	protected := api.Group("")
	protected.Use(httpserver.AuthnMiddleware(tokenParser))
	protected.GET("/me", httpserver.RequireRoles("manager", "practitioner", "parent"), meHandler())
	protected.GET("/authz/probe/manager", httpserver.RequireRoles("manager"), meHandler())
	protected.GET("/authz/probe/practitioner", httpserver.RequireRoles("practitioner"), meHandler())
	protected.GET("/authz/probe/parent", httpserver.RequireRoles("parent"), meHandler())
	protected.GET("/authz/probe/scope/:tenant_id/:branch_id", httpserver.RequireRoles("manager", "practitioner", "parent"), scopeProbeHandler())
	protected.GET("/authz/probe/parent-link/:child_id", httpserver.RequireRoles("parent"), parentLinkProbeHandler())

	// Shared infrastructure
	txManager := transaction.NewManager(pool)
	auditWriter := audit.NewWriter()

	// Children module
	childRepo := childpostgres.NewChildRepository(pool)
	childrenHandler := childhandler.NewHandler(
		childapp.NewListChildren(childRepo),
		childapp.NewGetChild(childRepo),
		childapp.NewCreateChild(childRepo, auditWriter),
		childapp.NewUpdateChild(childRepo, auditWriter),
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
	attendanceHandler := attendancehandler.NewHandler(
		attendanceapp.NewCheckInChild(attendanceRepo, childEnrollmentChecker, txManager, auditWriter, attendanceClock),
		attendanceapp.NewCheckOutChild(attendanceRepo, txManager, auditWriter, attendanceClock),
		attendanceapp.NewCorrectAttendance(attendanceRepo, childCorrectionChecker, txManager, auditWriter, attendanceClock),
	)

	// Register people routes
	childrenHandler.RegisterRoutes(protected)

	manager := protected.Group("")
	manager.Use(httpserver.RequireRoles("manager"))
	guardiansHandler.RegisterRoutes(manager)
	linksHandler.RegisterRoutes(manager)
	mappingsHandler.RegisterRoutes(manager)

	// Register attendance routes (manager + practitioner)
	attendanceHandler.RegisterRoutes(protected)

	return router
}

type healthPinger interface {
	Ping(context.Context) error
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

func meHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, ok := httpserver.AuthContextFromContext(c)
		if !ok {
			httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}
		c.JSON(http.StatusOK, gin.H{"auth": authCtx})
	}
}

func scopeProbeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, ok := httpserver.AuthContextFromContext(c)
		if !ok {
			httpserver.WriteError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}
		if c.Param("tenant_id") != authCtx.TenantID || c.Param("branch_id") != authCtx.BranchID {
			httpserver.WriteError(c, http.StatusForbidden, "forbidden_scope", "Access denied.", nil)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func parentLinkProbeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		linkedChildID := c.Query("linked_child_id")
		if linkedChildID == "" || linkedChildID != c.Param("child_id") {
			httpserver.WriteError(c, http.StatusForbidden, "forbidden_parent_child_link", "Access denied.", nil)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
