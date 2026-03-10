package grpc

import (
	"fmt"
	"net"

	"github.com/fleetml/fleetml/server/internal/deploy"
	"github.com/fleetml/fleetml/server/internal/fleet"
	servermodel "github.com/fleetml/fleetml/server/internal/model"
	"github.com/fleetml/fleetml/server/internal/monitor"
	"github.com/fleetml/fleetml/server/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server wraps the gRPC server.
type Server struct {
	server *grpc.Server
	fleet  *fleet.Manager
	logger *zap.SugaredLogger
	port   int
}

// NewServer creates a new gRPC server.
func NewServer(fleetMgr *fleet.Manager, orchestrator *deploy.Orchestrator, registry *servermodel.Registry, store storage.ObjectStore, metrics *monitor.MetricsProcessor, db *pgxpool.Pool, logger *zap.SugaredLogger, port int) *Server {
	s := &Server{
		server: grpc.NewServer(),
		fleet:  fleetMgr,
		logger: logger,
		port:   port,
	}

	handler := NewHandler(fleetMgr, orchestrator, registry, store, metrics, db, logger)
	handler.RegisterService(s.server)

	return s
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w", s.port, err)
	}

	s.logger.Infow("gRPC server starting", "port", s.port)
	return s.server.Serve(lis)
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	s.server.GracefulStop()
}
