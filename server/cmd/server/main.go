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

	"github.com/fleetml/fleetml/server/internal/api/grpc"
	"github.com/fleetml/fleetml/server/internal/api/rest"
	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/compiler"
	"github.com/fleetml/fleetml/server/internal/config"
	"github.com/fleetml/fleetml/server/internal/deploy"
	"github.com/fleetml/fleetml/server/internal/fleet"
	"github.com/fleetml/fleetml/server/internal/model"
	"github.com/fleetml/fleetml/server/internal/monitor"
	"github.com/fleetml/fleetml/server/internal/storage"
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

	// 4. Connect to MinIO/S3
	s3Store, err := storage.NewS3Store(
		cfg.Storage.Endpoint,
		cfg.Storage.AccessKey,
		cfg.Storage.SecretKey,
		cfg.Storage.Bucket,
		cfg.Storage.Region,
	)
	if err != nil {
		log.Fatalw("failed to connect to S3 storage", "error", err)
	}

	if err := s3Store.EnsureBucket(ctx); err != nil {
		log.Warnw("failed to ensure S3 bucket", "error", err)
	} else {
		log.Info("S3 storage ready")
	}

	// 5. Initialize services
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)
	fleetMgr := fleet.NewManager(pool)
	registry := model.NewRegistry(pool, s3Store)
	canaryMgr := deploy.NewCanaryManager(pool, log)
	orchestrator := deploy.NewOrchestrator(pool, fleetMgr, registry, canaryMgr, log)

	// 5b. Initialize compiler client
	var compilerClient *compiler.Client
	if cfg.Compiler.URL != "" {
		compilerClient = compiler.NewClient(cfg.Compiler.URL)
		log.Infow("compiler service configured", "url", cfg.Compiler.URL)
	}

	// 5c. Initialize monitoring
	metricsProcessor := monitor.NewMetricsProcessor(pool, log)
	alertEvaluator := monitor.NewAlertEvaluator(pool, 90*time.Second, log)
	go alertEvaluator.Start(ctx)
	log.Info("monitoring services started (metrics processor, alert evaluator)")

	// 6. Start REST API
	router := rest.NewRouter(fleetMgr, registry, orchestrator, compilerClient, jwtService, pool, log)
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

	// 7. Start gRPC server
	grpcAddr := fmt.Sprintf(":%d", cfg.Server.GRPCPort)
	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalw("failed to listen for gRPC", "address", grpcAddr, "error", err)
	}

	grpcServer := grpclib.NewServer()
	grpcHandler := grpc.NewHandler(fleetMgr, orchestrator, metricsProcessor, log)
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

	log.Info("server stopped")
}
