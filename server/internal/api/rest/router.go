package rest

import (
	"net/http"

	"github.com/fleetml/fleetml/server/internal/abtest"
	"github.com/fleetml/fleetml/server/internal/auth"
"github.com/fleetml/fleetml/server/internal/compiler"
	"github.com/fleetml/fleetml/server/internal/deploy"
	"github.com/fleetml/fleetml/server/internal/drift"
	"github.com/fleetml/fleetml/server/internal/fleet"
	"github.com/fleetml/fleetml/server/internal/integrations"
	mw "github.com/fleetml/fleetml/server/internal/middleware"
	"github.com/fleetml/fleetml/server/internal/model"
	"github.com/fleetml/fleetml/server/internal/policy"
	"github.com/fleetml/fleetml/server/internal/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// NewRouter creates the REST API router with all handlers.
func NewRouter(
	fleetMgr *fleet.Manager,
	registry *model.Registry,
	orchestrator *deploy.Orchestrator,
	compilerClient *compiler.Client,
	abtestMgr *abtest.Manager,
	driftDetector *drift.Detector,
	policyEngine *policy.Engine,
	integrationSvc *integrations.Service,
	jwtService *auth.JWTService,
	db *pgxpool.Pool,
	logger *zap.SugaredLogger,
) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(tracing.HTTPMiddleware)
	r.Use(mw.SecurityHeaders)

	// Rate limiting — default for all routes
	apiLimiter := mw.DefaultRateLimiter(logger)
	r.Use(apiLimiter.Middleware)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080", "https://*.fleetml.dev", "https://*.up.railway.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Handlers
	healthHandler := NewHealthHandler(db, logger)
	authHandler := NewAuthHandler(jwtService, db, logger)
	modelHandler := NewModelHandler(registry, logger)
	compileHandler := NewCompileHandler(registry, compilerClient, logger)
	deviceHandler := NewDeviceHandler(fleetMgr, logger)
	logsHandler := NewLogsHandler(db, logger)
	fleetHandler := NewFleetHandler(fleetMgr, logger)
	deployHandler := NewDeployHandler(orchestrator, logger)
	abtestHandler := NewABTestHandler(abtestMgr, logger)
	driftHandler := NewDriftHandler(driftDetector, logger)
	policyHandler := NewPolicyHandler(policyEngine, logger)
	integrationHandler := NewIntegrationHandler(integrationSvc, logger)
	apiKeyHandler := NewAPIKeyHandler(db, logger)

	// Strict rate limiter for auth endpoints
	authLimiter := mw.StrictRateLimiter(logger)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Get("/health", healthHandler.Health)
		r.With(authLimiter.Middleware, mw.RequestSizeLimit()).Post("/auth/register", authHandler.Register)
		r.With(authLimiter.Middleware, mw.RequestSizeLimit()).Post("/auth/login", authHandler.Login)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(jwtService.AuthMiddleware)

			// Auth
			r.With(mw.RequestSizeLimit()).Get("/auth/me", authHandler.Me)

			// API Keys
			r.Route("/api-keys", func(r chi.Router) {
				r.Use(mw.RequestSizeLimit())
				r.Get("/", apiKeyHandler.Get)
				r.Post("/regenerate", apiKeyHandler.Regenerate)
			})

			// Models — upload POST uses a 500MB limit; all other routes use 1MB.
			r.Route("/models", func(r chi.Router) {
				r.With(auth.RequirePermission("models:read"), mw.RequestSizeLimit()).Get("/", modelHandler.List)
				r.With(auth.RequirePermission("models:write"), mw.ModelUploadSizeLimit()).Post("/", modelHandler.Upload)
				r.With(auth.RequirePermission("models:read"), mw.RequestSizeLimit()).Get("/{id}", modelHandler.Get)
				r.With(auth.RequirePermission("models:delete"), mw.RequestSizeLimit()).Delete("/{id}", modelHandler.Delete)
				r.With(auth.RequirePermission("models:write"), mw.RequestSizeLimit()).Post("/{id}/compile", compileHandler.Compile)
			})

			// Devices
			r.Route("/devices", func(r chi.Router) {
				r.Use(mw.RequestSizeLimit())
				r.With(auth.RequirePermission("devices:read")).Get("/", deviceHandler.List)
				r.With(auth.RequirePermission("devices:read")).Get("/{device_id}", deviceHandler.Get)
				r.With(auth.RequirePermission("devices:read")).Get("/{device_id}/logs", logsHandler.GetLogs)
				r.With(auth.RequirePermission("devices:write")).Post("/{device_id}/logs", logsHandler.IngestLogs)
				r.With(auth.RequirePermission("devices:write")).Patch("/{device_id}", deviceHandler.Update)
				r.With(auth.RequirePermission("devices:delete")).Delete("/{device_id}", deviceHandler.Delete)
			})

			// Fleets
			r.Route("/fleets", func(r chi.Router) {
				r.Use(mw.RequestSizeLimit())
				r.With(auth.RequirePermission("fleets:read")).Get("/", fleetHandler.List)
				r.With(auth.RequirePermission("fleets:write")).Post("/", fleetHandler.Create)
				r.With(auth.RequirePermission("fleets:read")).Get("/{id}", fleetHandler.Get)
				r.With(auth.RequirePermission("fleets:write")).Patch("/{id}", fleetHandler.Update)
				r.With(auth.RequirePermission("fleets:delete")).Delete("/{id}", fleetHandler.Delete)
				r.With(auth.RequirePermission("fleets:read")).Get("/{id}/stats", fleetHandler.Stats)
				r.With(auth.RequirePermission("fleets:read")).Get("/{id}/devices", fleetHandler.ListDevices)
				r.With(auth.RequirePermission("fleets:write")).Post("/{id}/assign", fleetHandler.BulkAssign)
			})

			// A/B Tests
			r.Route("/ab-tests", func(r chi.Router) {
				r.Use(mw.RequestSizeLimit())
				r.With(auth.RequirePermission("deployments:read")).Get("/", abtestHandler.List)
				r.With(auth.RequirePermission("deployments:write")).Post("/", abtestHandler.Create)
				r.With(auth.RequirePermission("deployments:read")).Get("/{id}", abtestHandler.Get)
				r.With(auth.RequirePermission("deployments:write")).Post("/{id}/stop", abtestHandler.Stop)
			})

			// Drift Detection
			r.Route("/drift", func(r chi.Router) {
				r.Use(mw.RequestSizeLimit())
				r.With(auth.RequirePermission("models:write")).Post("/baselines", driftHandler.SetBaseline)
				r.With(auth.RequirePermission("models:write")).Post("/analyze", driftHandler.Analyze)
				r.With(auth.RequirePermission("models:read")).Get("/reports", driftHandler.ListReports)
			})

			// Policies
			r.Route("/policies", func(r chi.Router) {
				r.Use(mw.RequestSizeLimit())
				r.With(auth.RequirePermission("policies:read")).Get("/", policyHandler.List)
				r.With(auth.RequirePermission("policies:write")).Post("/", policyHandler.Create)
				r.With(auth.RequirePermission("policies:read")).Get("/{id}", policyHandler.Get)
				r.With(auth.RequirePermission("policies:write")).Patch("/{id}", policyHandler.Update)
				r.With(auth.RequirePermission("policies:delete")).Delete("/{id}", policyHandler.Delete)
			})

			// Integrations (MLflow, HuggingFace)
			r.Route("/integrations", func(r chi.Router) {
				r.Use(mw.RequestSizeLimit())
				r.With(auth.RequirePermission("models:write")).Post("/mlflow/import", integrationHandler.ImportMLflow)
				r.With(auth.RequirePermission("models:write")).Post("/huggingface/import", integrationHandler.ImportHuggingFace)
			})

			// Deployments
			r.Route("/deployments", func(r chi.Router) {
				r.Use(mw.RequestSizeLimit())
				r.With(auth.RequirePermission("deployments:read")).Get("/", deployHandler.List)
				r.With(auth.RequirePermission("deployments:write")).Post("/", deployHandler.Create)
				r.With(auth.RequirePermission("deployments:read")).Get("/{id}", deployHandler.Get)
				r.With(auth.RequirePermission("deployments:cancel")).Post("/{id}/cancel", deployHandler.Cancel)
				r.With(auth.RequirePermission("deployments:write")).Post("/{id}/rollback", deployHandler.Rollback)
			})
		})

		// Heartbeat (API key auth, not JWT)
		r.With(mw.RequestSizeLimit()).Post("/heartbeat", healthHandler.Heartbeat)
	})

	return r
}
