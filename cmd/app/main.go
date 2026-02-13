package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"palantir/config"
	"palantir/controllers"
	"palantir/database"
	"palantir/internal/server"
	"palantir/internal/storage"
	"palantir/queue"
	"palantir/router"
	"palantir/router/middleware"
	"palantir/telemetry"
	"palantir/queue/workers"
	"riverqueue.com/riverui"
	"palantir/clients/email"

	"github.com/a-h/templ"
)

var appVersion string

// setupControllers initializes and registers all controllers with the router.
// The comment: 'andurel:controller-registration-point' is used as a marker for automated code generation tools.
// If you remove it or change it, the generator tool may not function correctly. 
// If you decide to move it, ensure it remains in a logical place within a function that passes the correct dependencies to the controllers.
func setupControllers(
	cfg config.Config,
	db storage.Pool,
	insertOnly queue.InsertOnly,
	r *router.Router,
	riverHandler *riverui.Handler,
	mw middleware.Middleware,
) error {
	pagesCache, err := controllers.NewCacheBuilder[templ.Component]().Build()
	if err != nil {
		return err
	}

	assetsCache, err := controllers.NewCacheBuilder[string]().WithSize(2).Build()
	if err != nil {
		return err
	}
	assets := controllers.NewAssets(assetsCache)
	api := controllers.NewAPI(db)
	pages := controllers.NewPages(db, insertOnly, pagesCache)
	sessions := controllers.NewSessions(db, cfg)
	registrations := controllers.NewRegistrations(db, insertOnly, cfg)
	confirmations := controllers.NewConfirmations(db, cfg)
	resetPasswords := controllers.NewResetPasswords(db, insertOnly, cfg)

	if err := r.RegisterAPIRoutes(api); err != nil {
		return err
	}

	if err := r.RegisterAssetsRoutes(assets); err != nil {
		return err
	}

	if err := r.RegisterConfirmationsRoutes(confirmations); err != nil {
		return err
	}

	if err := r.RegisterPagesRoutes(pages); err != nil {
		return err
	}

	if err := r.RegisterResetPasswordsRoutes(resetPasswords); err != nil {
		return err
	}

	if err := r.RegisterSessionsRoutes(sessions); err != nil {
		return err
	}

	if err := r.RegisterRegistrationsRoutes(registrations); err != nil {
		return err
	}

	// andurel:controller-registration-point

	r.RegisterCustomRoutes(
		riverHandler,
		pages.NotFound,
	)

	return nil
}

func setupRouter(
	cfg config.Config,
	tel *telemetry.Telemetry,
	mw middleware.Middleware,
) (*router.Router, error) {
	authKey, err := hex.DecodeString(cfg.App.SessionKey)
	if err != nil {
		return nil, err
	}
	encKey, err := hex.DecodeString(cfg.App.SessionEncryptionKey)
	if err != nil {
		return nil, err
	}

	globalMiddleware, err := router.SetupGlobalMiddleware(cfg, tel, authKey, encKey, mw, "_csrf")
	if err != nil {
		return nil, err
	}

	r, err := router.New(
		true,
		globalMiddleware,
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func parseHeaders(headersStr string) map[string]string {
	headers := make(map[string]string)
	if headersStr == "" {
		return headers
	}

	pairs := strings.SplitSeq(headersStr, ",")
	for pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return headers
}

func buildTelemetry(ctx context.Context, cfg config.Config) (*telemetry.Telemetry, error) {
	opts := []telemetry.Option{
		telemetry.WithService(cfg.Telemetry.ServiceName, cfg.Telemetry.ServiceVersion),
		telemetry.WithBatchConfig(cfg.Telemetry.BatchSize, cfg.Telemetry.BatchTimeoutMs, 2048),
		telemetry.WithTraceSampleRate(cfg.Telemetry.TraceSampleRate),
	}

	opts = append(opts, telemetry.WithLogExporters(telemetry.NewStdoutExporter()))

	if cfg.Telemetry.OtlpMetricsEndpoint != "" {
		opts = append(opts, telemetry.WithMetricExporters(
			telemetry.NewOtlpMetricExporter(cfg.Telemetry.OtlpMetricsEndpoint, parseHeaders(cfg.Telemetry.OtlpHeaders))))
	}

	if cfg.Telemetry.OtlpTracesEndpoint != "" {
		opts = append(opts, telemetry.WithTraceExporters(
			telemetry.NewOtlpTraceExporter(cfg.Telemetry.OtlpTracesEndpoint, parseHeaders(cfg.Telemetry.OtlpHeaders))))
	} else {
		opts = append(opts, telemetry.WithTraceExporters(telemetry.NewNoopTraceExporter()))
	}

	return telemetry.New(ctx, opts...)
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cfg := config.NewConfig()

	tel, err := buildTelemetry(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tel.Shutdown(shutdownCtx); err != nil {
			slog.Error("telemetry shutdown error", "error", err)
		}
	}()

	if err := tel.HealthCheck(ctx); err != nil {
		slog.Warn("telemetry health check failed", "error", err)
	}


	db, err := database.NewPostgres(ctx, cfg.DB.GetDatabaseURL())
	if err != nil {
		return err
	}
	emailClient := mailclients.NewMailpit(cfg.Email.MailpitHost, cfg.Email.MailpitPort)

	wrks, err := workers.Register(emailClient, emailClient)
	if err != nil {
		return err
	}

	insertOnly, err := queue.NewInsertOnly(
		db,
		wrks,
	)
	if err != nil {
		return err
	}

	processor, err := queue.NewProcessor(
		ctx,
		db,
		wrks,
	)
	if err != nil {
		return err
	}

	go func() {
		if err := processor.Start(ctx); err != nil {
			slog.Error("queue processor error", "error", err)
			cancel()
		}
	}()

	mw := middleware.New(db)

	endpoints := riverui.NewEndpoints(processor.Client, nil)
	opts := &riverui.HandlerOpts{
		Endpoints: endpoints,
		Logger:    slog.Default(),
		Prefix:    "/riverui", // mount the UI and its APIs under /riverui or another path
	}
	riverHandler, err := riverui.NewHandler(opts)
	if err != nil {
		return err
	}

	riverHandler.Start(ctx)

	r, err := setupRouter(cfg, tel, mw)
	if err != nil {
		return err
	}

	err = setupControllers(
		cfg,
		db,
		insertOnly,
		r,
		riverHandler,
		mw,
	)
	if err != nil {
		return err
	}

	server := server.New(
		ctx,
		cfg.App.Host,
		cfg.App.Port,
		config.Env,
		r.Handler,
		[]server.Shutdowner{processor},
	)

	slog.InfoContext(ctx, "starting server", "host", cfg.App.Host, "port", cfg.App.Port)
	return server.Start(ctx, config.Env)
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
