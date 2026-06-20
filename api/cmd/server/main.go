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
	"nursery-management-system/api/internal/platform/logging"
	"nursery-management-system/api/internal/platform/metrics"

	"github.com/prometheus/client_golang/prometheus"
	billingapp "nursery-management-system/api/internal/modules/billing/application"
	billingpostgres "nursery-management-system/api/internal/modules/billing/infrastructure/postgres"
	invoicerun "nursery-management-system/api/internal/modules/invoicerun"
	termpostgres "nursery-management-system/api/internal/modules/term/infrastructure/postgres"
	termapp "nursery-management-system/api/internal/modules/term/application"
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

	var registry *prometheus.Registry
	var recorder *metrics.Recorder
	if cfg.MetricsEnabled {
		registry = metrics.NewRegistry()
		recorder = metrics.NewRecorder(registry)
	}

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
		overdueUC := billingapp.NewMarkOverdueInvoices(billingRepo, txMgr, func() time.Time { return time.Now().UTC() })

		termRepo := termpostgres.NewTermRepository(pool)
		auditWriter := audit.NewWriter()
		expireUC := termapp.NewExpireTermsUseCase(termRepo, auditWriter, txMgr)
		generateUC := billingapp.NewGenerateDraftInvoices(billingRepo, txMgr, auditWriter)
		lister := invoicerun.NewSystemTenantBranchLister(pool)
		expireRunner := invoicerun.NewExpireTermsRunner(expireUC, lister)
		generateRunner := invoicerun.NewGenerateAdvanceInvoicesRunner(generateUC, lister)

		var schedErr error
		scheduler, schedErr = invoicerun.NewScheduler(logger, overdueUC, expireRunner, generateRunner, recorder)
		if schedErr != nil {
			logger.Error("failed to create scheduler", "error", schedErr)
			os.Exit(1)
		}
	}

	router := bootstrap.BootstrapWithOptions(cfg, logger, pool, bootstrap.BootstrapOptions{
		MetricsRegistry: registry,
		MetricsRecorder: recorder,
	})

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
