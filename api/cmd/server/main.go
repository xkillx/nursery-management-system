package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nursery-management-system/api/internal/app/bootstrap"
	"nursery-management-system/api/internal/platform/audit"
	"nursery-management-system/api/internal/platform/config"
	"nursery-management-system/api/internal/platform/db"
	"nursery-management-system/api/internal/platform/events"
	"nursery-management-system/api/internal/platform/logging"

	billingapp "nursery-management-system/api/internal/modules/billing/application"
	billingpostgres "nursery-management-system/api/internal/modules/billing/infrastructure/postgres"
	invoicerun "nursery-management-system/api/internal/modules/invoicerun"
	termapp "nursery-management-system/api/internal/modules/term/application"
	termpostgres "nursery-management-system/api/internal/modules/term/infrastructure/postgres"
	"nursery-management-system/api/internal/platform/transaction"
)

func main() {
	logger := logging.NewJSONLogger(os.Stdout, slog.LevelInfo)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logLevel, err := logging.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Error("invalid log level", "error", err)
		os.Exit(1)
	}
	logger = logging.NewJSONLogger(os.Stdout, logLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	var scheduler *invoicerun.Scheduler
	if cfg.SchedulerOwner {
		billingRepo := billingpostgres.NewRepository(pool)
		txMgr := transaction.NewManager(pool)
		eventDispatcher := events.NewEventDispatcher(txMgr)
		overdueUC := billingapp.NewMarkOverdueInvoices(billingRepo, eventDispatcher, func() time.Time { return time.Now().UTC() })

		termRepo := termpostgres.NewTermRepository(pool)
		auditWriter := audit.NewWriter()
		expireUC := termapp.NewExpireTermsUseCase(termRepo, auditWriter, txMgr)
		markPendingUC := termapp.NewMarkPendingRenewalUseCase(termRepo, auditWriter, txMgr)
		lister := invoicerun.NewSystemTenantBranchLister(pool)
		expireRunner := invoicerun.NewExpireTermsRunner(expireUC, markPendingUC, lister)

		var schedErr error
		scheduler, schedErr = invoicerun.NewScheduler(logger, overdueUC, expireRunner, nil)
		if schedErr != nil {
			logger.Error("failed to create scheduler", "error", schedErr)
			os.Exit(1)
		}
	}

	router, err := bootstrap.InitializeApp(cfg, logger, pool)
	if err != nil {
		logger.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting api server", "port", cfg.APIPort, "base_path", cfg.APIBasePath, "env", cfg.AppEnv)
		errCh <- httpServer.ListenAndServe()
	}()

	if scheduler != nil {
		scheduler.Start(ctx)
	}

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if scheduler != nil {
		scheduler.Stop(shutdownCtx)
	}

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server shutdown complete")
}
