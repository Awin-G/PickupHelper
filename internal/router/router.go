package router

import (
	"strings"

	"pickup-helper/internal/config"
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/handler"
	"pickup-helper/internal/middleware"

	"github.com/gin-gonic/gin"
)

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
// /api/v1 group additionally mounts JWTAuth so all sub-routes require a
// valid access token. Sub-groups that allow public access (e.g. /auth/*)
// should be registered on a fresh group WITHOUT copying the middleware.
//
// engine.NoRoute enforces JWT for unmatched /api/v1/* paths so that
// `curl /api/v1/anything` without a token returns 401 (not 404).
func Register(engine *gin.Engine, cfg *config.Config, h *handler.HealthHandler) {
	engine.Use(
		middleware.Recovery(),
		middleware.TraceID(),
		middleware.Logger(),
		middleware.CORS(cfg.CORS.AllowedOrigins),
		middleware.RateLimit(cfg.RateLimit.QPS, cfg.RateLimit.Burst),
	)

	// Public health endpoints (no JWT).
	h.Register(engine)

	// Authenticated API v1 group — Phase 2+ adds business routes here
	// via NewAPIV1Group. No routes yet, so the group itself is a no-op;
	// the NoRoute handler below enforces JWT for unmatched /api/v1/* paths.
	NewAPIV1Group(engine, cfg)

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

// NewAPIV1Group creates the authenticated /api/v1 group. Phase 2+ will
// call this (or extend Register) to attach business routes.
// The returned group has JWTAuth mounted; sub-groups can opt out by
// creating a fresh group via engine.Group("/api/v1/...") without copying
// the middleware.
func NewAPIV1Group(engine *gin.Engine, cfg *config.Config) *gin.RouterGroup {
	g := engine.Group("/api/v1")
	g.Use(middleware.JWTAuth(cfg))
	// No routes yet — phase 2+ will add them here.
	return g
}
