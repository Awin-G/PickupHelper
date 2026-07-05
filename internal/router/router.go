package router

import (
	"pickup-helper/internal/config"
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
// should be registered on a fresh group WITHOUT the JWT middleware.
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

	// Authenticated API v1 group — every sub-route requires a valid token.
	NewAPIV1Group(engine, cfg)
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
