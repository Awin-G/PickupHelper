package router

import (
	"strings"

	"pickup-helper/internal/config"
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/handler"
	"pickup-helper/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Handlers bundles all module handlers that the router mounts. Phase 1
// only populates Health; Phase 2 adds Auth and User. Future phases extend
// this struct (parcel / pickup / station / etc.).
type Handlers struct {
	Health *handler.HealthHandler
	Auth   *handler.AuthHandler
	User   *handler.UserHandler
	Parcel *handler.ParcelHandler
	Pickup *handler.PickupHandler
	Proxy  *handler.ProxyHandler
	Shelf  *handler.ShelfHandler
	Notify *handler.NotifyHandler
	Stats  *handler.StatsHandler
}

// Register wires the global middleware chain and application routes onto
// the engine.
//
// Middleware order (outermost → innermost):
//  1. Recovery  — catch panics, never crash the process
//  2. TraceID   — inject trace_id into context for all subsequent logs
//  3. Logger    — one structured log line per request (uses defer)
//  4. CORS      — set Access-Control-* headers, short-circuit OPTIONS
//  5. RateLimit — per-IP token bucket
//
// Route groups under /api/v1:
//   - /auth/*                     public  (no JWT)
//   - /admin/auth/login           public  (no JWT, admin login)
//   - /user/*                     user    (JWT required)
//   - /admin/*                    admin   (JWT + AdminOnly)
//
// engine.NoRoute enforces JWT for unmatched /api/v1/* paths so that
// `curl /api/v1/anything` without a token returns 401 (not 404).
func Register(engine *gin.Engine, cfg *config.Config, h *Handlers) {
	engine.Use(
		middleware.Recovery(),
		middleware.TraceID(),
		middleware.Logger(),
		middleware.CORS(cfg.CORS.AllowedOrigins),
		middleware.RateLimit(cfg.RateLimit.QPS, cfg.RateLimit.Burst),
	)

	// Public health endpoints (no JWT).
	h.Health.Register(engine)

	// --- Public API v1 routes (no JWT) ---
	// We deliberately use engine.Group (not NewAPIV1Group) so the JWT
	// middleware is NOT inherited. Each public sub-group is responsible
	// for its own protection (none — these are intentionally open).
	publicV1 := engine.Group("/api/v1")
	if h.Auth != nil {
		h.Auth.RegisterPublicRoutes(publicV1.Group("/auth"))
	}
	// Admin login is public despite the /admin prefix — AdminOnly must
	// NOT be applied to this route.
	if h.Auth != nil {
		h.Auth.RegisterAdminAuthRoutes(publicV1)
	}

	// --- Authenticated API v1 routes (JWT required) ---
	jwtGroup := NewAPIV1Group(engine, cfg)

	// User-facing routes (JWT only).
	if h.User != nil {
		h.User.RegisterUserRoutes(jwtGroup.Group("/user"))
	}

	// Admin-only routes (JWT + AdminOnly).
	if h.User != nil {
		adminGroup := jwtGroup.Group("/admin")
		adminGroup.Use(middleware.AdminOnly())
		h.User.RegisterAdminRoutes(adminGroup)
	}

	// Parcel admin routes (JWT + AdminOnly).
	if h.Parcel != nil {
		parcelAdminGroup := jwtGroup.Group("/parcels")
		parcelAdminGroup.Use(middleware.AdminOnly())
		h.Parcel.RegisterParcelAdminRoutes(parcelAdminGroup)
	}

	// Parcel user routes (JWT only).
	if h.Parcel != nil {
		h.Parcel.RegisterParcelUserRoutes(jwtGroup.Group("/parcels"))
	}

	// Pickup routes (JWT only, handler does internal permission checks).
	if h.Pickup != nil {
		h.Pickup.RegisterPickupRoutes(jwtGroup.Group("/pickup"))
	}

	if h.Proxy != nil {
		h.Proxy.RegisterProxyRoutes(jwtGroup.Group("/proxy"))
	}

	if h.Shelf != nil {
		shelfGroup := jwtGroup.Group("/shelves")
		shelfGroup.Use(middleware.AdminOnly())
		h.Shelf.RegisterShelfRoutes(shelfGroup)
	}

	if h.Notify != nil {
		h.Notify.RegisterNotifyRoutes(jwtGroup.Group("/notifications"))
	}

	if h.Stats != nil {
		statsGroup := jwtGroup.Group("/stats")
		statsGroup.Use(middleware.AdminOnly())
		h.Stats.RegisterStatsRoutes(statsGroup)
	}

	// NoRoute: enforce JWT for unmatched /api/v1/* paths.
	// Without this, `GET /api/v1/nonexistent` would return 404 (bypassing
	// the group's JWT middleware) and leak the existence of routes. By
	// re-running JWTAuth here, unauthenticated requests get 401, and
	// authenticated requests get a proper 404 envelope.
	jwtAuth := middleware.JWTAuth(cfg)
	engine.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/v1/") {
			handler.Error(c, apperrors.New(apperrors.ErrNotFound, "route not found"))
			return
		}
		// Re-run JWT validation for unmatched API routes.
		jwtAuth(c)
		if c.IsAborted() {
			return // JWT rejected — 401 already written.
		}
		// Token valid but route doesn't exist → 404.
		handler.Error(c, apperrors.New(apperrors.ErrNotFound, "route not found"))
	})
}

// NewAPIV1Group creates the authenticated /api/v1 group with JWTAuth.
// Sub-groups can opt out of JWT by registering directly on the engine
// via engine.Group("/api/v1/...") without copying the middleware.
func NewAPIV1Group(engine *gin.Engine, cfg *config.Config) *gin.RouterGroup {
	g := engine.Group("/api/v1")
	g.Use(middleware.JWTAuth(cfg))
	return g
}
