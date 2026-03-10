package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fleetml/fleetml/agent/internal/communication"
	"github.com/fleetml/fleetml/agent/internal/config"
	agentdeploy "github.com/fleetml/fleetml/agent/internal/deploy"
	"github.com/fleetml/fleetml/agent/internal/device"
	"github.com/fleetml/fleetml/agent/internal/health"
	"github.com/fleetml/fleetml/agent/internal/heartbeat"
	"github.com/fleetml/fleetml/agent/internal/model"
	"github.com/fleetml/fleetml/agent/internal/offline"
	"github.com/fleetml/fleetml/agent/pkg/version"
	"go.uber.org/zap"
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
		fmt.Printf("fleetml-agent %s (commit: %s, built: %s)\n",
			version.Version, version.GitCommit, version.BuildDate)
		os.Exit(0)
	}

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	log.Infow("starting fleetml-agent",
		"version", version.Version,
		"commit", version.GitCommit,
	)

	// 1. Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalw("failed to load config", "error", err)
	}

	// 2. Detect hardware
	info, err := device.Fingerprint(cfg.DeviceID)
	if err != nil {
		log.Fatalw("failed to fingerprint device", "error", err)
	}
	log.Infow("device fingerprint",
		"device_id", info.DeviceID,
		"arch", info.Arch,
		"gpu", info.GPUType,
		"runtime", info.Runtime,
		"ram_mb", info.RAMMB,
		"disk_gb", info.DiskGB,
		"os", info.OS,
	)

	// 3. Initialize model hot-swapper and loader
	swapper := model.NewHotSwapper()
	loader := model.NewLoader(cfg.ModelStorageDir, cfg.MaxModelVersions)
	rollbackMgr := agentdeploy.NewRollbackManager(cfg.ModelStorageDir, cfg.MaxModelVersions)

	// 4. Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5. Initialize offline store (SQLite)
	os.MkdirAll(cfg.ModelStorageDir, 0755)
	dbPath := filepath.Join(cfg.ModelStorageDir, "agent.db")
	offlineStore, err := offline.NewSQLiteStore(dbPath)
	if err != nil {
		log.Warnw("failed to initialize offline store, continuing without buffering", "error", err)
	} else {
		defer offlineStore.Close()
		log.Infow("offline store initialized", "path", dbPath)
	}

	// 6. Connect to control plane via gRPC
	grpcClient, err := communication.NewGRPCClient(cfg.ServerAddress, log)
	if err != nil {
		log.Fatalw("failed to create gRPC client", "error", err)
	}
	defer grpcClient.Close()

	if cfg.APIKey != "" {
		grpcClient.SetAPIKey(cfg.APIKey)
	}

	if err := grpcClient.Connect(ctx); err != nil {
		log.Fatalw("failed to connect to control plane", "address", cfg.ServerAddress, "error", err)
	}

	// 7. Wrap with store-forward manager for offline resilience
	var communicator communication.Communicator = grpcClient
	if offlineStore != nil {
		communicator = communication.NewStoreForwardManager(grpcClient, offlineStore, log)
	}

	// 8. Register device with control plane
	agentID, heartbeatIntervalSec, err := communicator.Register(ctx, info)
	if err != nil {
		log.Fatalw("failed to register device", "error", err)
	}
	log.Infow("device registered",
		"agent_id", agentID,
		"heartbeat_interval_sec", heartbeatIntervalSec,
	)

	// 9. Initialize deploy manager
	deployMgr := agentdeploy.NewManager(info.DeviceID, loader, swapper, rollbackMgr, communicator, log)
	_ = deployMgr

	// 10. Initialize health reporter and heartbeat scheduler
	hbInterval := time.Duration(heartbeatIntervalSec) * time.Second
	if hbInterval == 0 {
		hbInterval = cfg.HeartbeatInterval
	}
	healthReporter := health.NewReporter(hbInterval)
	hbScheduler := heartbeat.NewScheduler(info.DeviceID, hbInterval, communicator, healthReporter, log)

	// 10. Start heartbeat loop in background
	go hbScheduler.Start(ctx)

	log.Infow("agent ready",
		"device_id", cfg.DeviceID,
		"agent_id", agentID,
		"server", cfg.ServerAddress,
		"heartbeat_interval", hbInterval,
	)

	// 11. Main event loop: handle commands + shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case cmd := <-hbScheduler.Commands():
			log.Infow("received command", "id", cmd.ID, "type", cmd.Type)
			go deployMgr.HandleCommand(ctx, cmd)

		case sig := <-sigCh:
			log.Infow("received signal, shutting down", "signal", sig)
			cancel()
			log.Info("agent stopped")
			return

		case <-ctx.Done():
			log.Info("agent stopped")
			return
		}
	}
}
