package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fleetml/fleetml/server/internal/abtest"
	"github.com/fleetml/fleetml/server/internal/api/grpc"
	"github.com/fleetml/fleetml/server/internal/api/rest"
	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/billing"
	"github.com/fleetml/fleetml/server/internal/compiler"
	"github.com/fleetml/fleetml/server/internal/config"
	"github.com/fleetml/fleetml/server/internal/deploy"
	"github.com/fleetml/fleetml/server/internal/drift"
	"github.com/fleetml/fleetml/server/internal/fleet"
	"github.com/fleetml/fleetml/server/internal/integrations"
	"github.com/fleetml/fleetml/server/internal/messaging"
	"github.com/fleetml/fleetml/server/internal/model"
	"github.com/fleetml/fleetml/server/internal/monitor"
	"github.com/fleetml/fleetml/server/internal/policy"
	"github.com/fleetml/fleetml/server/internal/storage"
	"github.com/fleetml/fleetml/server/internal/tracing"
	"go.uber.org/zap"
	grpclib "google.golang.org/grpc"
)

var (
	version   = "0.1.0"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	var (
		configPath  string
		showVersion bool
	)

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("fleetml-server %s (commit: %s, built: %s)\n", version, gitCommit, buildDate)
		os.Exit(0)
	}

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	log.Infow("starting fleetml-server",
		"version", version,
		"commit", gitCommit,
	)

	// 1. Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalw("failed to load config", "error", err)
	}

	log.Infow("configuration loaded",
		"rest_port", cfg.Server.RESTPort,
		"grpc_port", cfg.Server.GRPCPort,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1b. Initialize tracing
	tracingProvider, err := tracing.Init(ctx, tracing.Config{
		Enabled:    cfg.Tracing.Enabled,
		Endpoint:   cfg.Tracing.Endpoint,
		SampleRate: cfg.Tracing.SampleRate,
	}, version, log)
	if err != nil {
		log.Warnw("failed to initialize tracing", "error", err)
	} else {
		defer tracingProvider.Shutdown(ctx)
	}

	// 2. Connect to PostgreSQL
	pool, err := storage.NewPostgresPool(ctx, cfg.Database.URL, cfg.Database.MaxConnections)
	if err != nil {
		log.Fatalw("failed to connect to database", "error", err)
	}
	defer pool.Close()
	log.Info("connected to PostgreSQL")

	// 3. Run database migrations
	// Try Docker path first, fall back to local dev path
	migrationsPath := "/migrations"
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		migrationsPath = "server/migrations"
	}
	if err := storage.RunMigrations(pool, migrationsPath); err != nil {
		log.Warnw("migrations failed", "error", err)
	} else {
		log.Info("database migrations applied")
	}

	// 4. Connect to MinIO/S3 (optional — server starts without it)
	var s3Store *storage.S3Store
	if cfg.Storage.Endpoint != "" && cfg.Storage.AccessKey != "" {
		var err error
		s3Store, err = storage.NewS3Store(
			cfg.Storage.Endpoint,
			cfg.Storage.AccessKey,
			cfg.Storage.SecretKey,
			cfg.Storage.Bucket,
			cfg.Storage.Region,
		)
		if err != nil {
			log.Warnw("failed to connect to S3 storage (model uploads disabled)", "error", err)
		} else if err := s3Store.EnsureBucket(ctx); err != nil {
			log.Warnw("failed to ensure S3 bucket", "error", err)
		} else {
			log.Info("S3 storage ready")
		}
	} else {
		log.Warn("S3 storage not configured — model uploads/downloads disabled")
	}

	// 5. Initialize services
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)
	fleetMgr := fleet.NewManager(pool)
	registry := model.NewRegistry(pool, s3Store)
	canaryMgr := deploy.NewCanaryManager(pool, log)
	orchestrator := deploy.NewOrchestrator(pool, fleetMgr, registry, canaryMgr, log)
	abtestMgr := abtest.NewManager(pool, log)
	driftDetector := drift.NewDetector(pool, log)
	policyEngine := policy.NewEngine(pool, log)

	// 5a2. Initialize integrations service
	integrationSvc := integrations.NewService(registry, log)
	if cfg.Integrations.MLflowURL != "" {
		integrationSvc.SetMLflowClient(integrations.NewMLflowClient(cfg.Integrations.MLflowURL))
		log.Infow("MLflow integration configured", "url", cfg.Integrations.MLflowURL)
	}
	integrationSvc.SetHuggingFaceClient(integrations.NewHuggingFaceClient(cfg.Integrations.HFToken))
	if cfg.Integrations.HFToken != "" {
		log.Info("HuggingFace integration configured (with auth token)")
	} else {
		log.Info("HuggingFace integration configured (public repos only)")
	}

	// 5b. Initialize compiler client
	var compilerClient *compiler.Client
	if cfg.Compiler.URL != "" {
		compilerClient = compiler.NewClient(cfg.Compiler.URL)
		log.Infow("compiler service configured", "url", cfg.Compiler.URL)
	}

	// 5b2. Initialize NATS message bus
	var natsClient *messaging.NATSClient
	if cfg.NATS.URL != "" {
		var err error
		natsClient, err = messaging.NewNATSClient(cfg.NATS.URL, log)
		if err != nil {
			log.Warnw("failed to connect to NATS (commands will use heartbeat piggybacking only)", "error", err)
		} else {
			defer natsClient.Close()
			log.Infow("NATS message bus connected", "url", cfg.NATS.URL)
		}
	}
	_ = natsClient // Will be used for command publishing in orchestrator

	// 5c. Initialize monitoring
	metricsProcessor := monitor.NewMetricsProcessor(pool, log)
	alertEvaluator := monitor.NewAlertEvaluator(pool, 90*time.Second, log)
	go alertEvaluator.Start(ctx)

	// 5d. Initialize Prometheus exporter
	promExporter := monitor.NewPrometheusExporter(pool, log)
	go promExporter.Start(ctx, 15*time.Second)
	log.Info("monitoring services started (metrics processor, alert evaluator, prometheus)")

	// 5e. Initialize billing client
	billingClient := billing.NewClient(cfg.Billing, pool, log)
	if cfg.Billing.DodoAPIKey != "" {
		log.Infow("Dodo Payments billing configured", "environment", cfg.Billing.DodoEnvironment)
	}

	// 6. Start REST API
	router := rest.NewRouter(fleetMgr, registry, orchestrator, compilerClient, abtestMgr, driftDetector, policyEngine, integrationSvc, billingClient, jwtService, pool, log)
	restAddr := fmt.Sprintf(":%d", cfg.Server.RESTPort)
	httpServer := &http.Server{
		Addr:         restAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Infow("REST API listening", "address", restAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("REST server error", "error", err)
		}
	}()

	// 6b. Start Prometheus metrics server on port 9090
	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/metrics", promExporter.Handler())
	metricsMux.HandleFunc("/metrics/json", promExporter.JSONHandler())
	metricsServer := &http.Server{
		Addr:         ":9090",
		Handler:      metricsMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		log.Infow("Prometheus metrics listening", "address", ":9090")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("metrics server error", "error", err)
		}
	}()

	// 7. Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalw("failed to listen for gRPC", "address", grpcAddr, "error", err)
	}

	grpcServer := grpclib.NewServer()
	grpcHandler := grpc.NewHandler(fleetMgr, orchestrator, registry, s3Store, metricsProcessor, log)
	grpcHandler.RegisterService(grpcServer)

	go func() {
		log.Infow("gRPC server listening", "address", grpcAddr)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalw("gRPC server error", "error", err)
		}
	}()

	log.Infow("server ready",
		"rest", restAddr,
		"grpc", grpcAddr,
	)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Infow("received signal, shutting down", "signal", sig)
		cancel()
	case <-ctx.Done():
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()
	httpServer.Shutdown(shutdownCtx)
	metricsServer.Shutdown(shutdownCtx)

	log.Info("server stopped")
}
