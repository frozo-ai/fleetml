package rest

import (
	"net/http"

	"github.com/fleetml/fleetml/server/internal/auth"
	"github.com/fleetml/fleetml/server/internal/deploy"
	"github.com/fleetml/fleetml/server/internal/fleet"
	"github.com/fleetml/fleetml/server/internal/model"
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
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
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
	deviceHandler := NewDeviceHandler(fleetMgr, logger)
	fleetHandler := NewFleetHandler(fleetMgr, logger)
	deployHandler := NewDeployHandler(orchestrator, logger)

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Get("/health", healthHandler.Health)
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(jwtService.AuthMiddleware)

			// Auth
			r.Get("/auth/me", authHandler.Me)

			// Models
			r.Route("/models", func(r chi.Router) {
				r.With(auth.RequirePermission("models:read")).Get("/", modelHandler.List)
				r.With(auth.RequirePermission("models:write")).Post("/", modelHandler.Upload)
				r.With(auth.RequirePermission("models:read")).Get("/{id}", modelHandler.Get)
				r.With(auth.RequirePermission("models:delete")).Delete("/{id}", modelHandler.Delete)
			})

			// Devices
			r.Route("/devices", func(r chi.Router) {
				r.With(auth.RequirePermission("devices:read")).Get("/", deviceHandler.List)
				r.With(auth.RequirePermission("devices:read")).Get("/{device_id}", deviceHandler.Get)
				r.With(auth.RequirePermission("devices:write")).Patch("/{device_id}", deviceHandler.Update)
				r.With(auth.RequirePermission("devices:delete")).Delete("/{device_id}", deviceHandler.Delete)
			})

			// Fleets
			r.Route("/fleets", func(r chi.Router) {
				r.With(auth.RequirePermission("fleets:read")).Get("/", fleetHandler.List)
				r.With(auth.RequirePermission("fleets:write")).Post("/", fleetHandler.Create)
				r.With(auth.RequirePermission("fleets:read")).Get("/{id}", fleetHandler.Get)
				r.With(auth.RequirePermission("fleets:write")).Patch("/{id}", fleetHandler.Update)
				r.With(auth.RequirePermission("fleets:delete")).Delete("/{id}", fleetHandler.Delete)
			})

			// Deployments
			r.Route("/deployments", func(r chi.Router) {
				r.With(auth.RequirePermission("deployments:read")).Get("/", deployHandler.List)
				r.With(auth.RequirePermission("deployments:write")).Post("/", deployHandler.Create)
				r.With(auth.RequirePermission("deployments:read")).Get("/{id}", deployHandler.Get)
				r.With(auth.RequirePermission("deployments:cancel")).Post("/{id}/cancel", deployHandler.Cancel)
				r.With(auth.RequirePermission("deployments:write")).Post("/{id}/rollback", deployHandler.Rollback)
			})
		})

		// Heartbeat (API key auth, not JWT)
		r.Post("/heartbeat", healthHandler.Heartbeat)
	})

	return r
}
