package handler

import (
	"context"
	"net/http"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// Pinger is the minimal contract HealthHandler depends on. Defined as an
// interface so unit tests can inject fakes without touching a real DB/Redis.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// RedisPinger abstracts redis.Client.Ping for testability.
type RedisPinger interface {
	Ping(ctx context.Context) *redis.StatusCmd
}

type HealthHandler struct {
	db  Pinger
	rdb RedisPinger
}

func NewHealthHandler(db *sqlx.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

// Register attaches health endpoints to the engine.
func (h *HealthHandler) Register(engine *gin.Engine) {
	engine.GET("/health", h.Live)
	engine.GET("/health/ready", h.Ready)
}

// Live is the liveness probe — always returns up when the process is alive.
func (h *HealthHandler) Live(c *gin.Context) {
	Success(c, gin.H{"status": "up"})
}

// Ready is the readiness probe — checks MySQL and Redis connectivity.
// Failures return HTTP 503 + code=10009 (matches plan 01-01 contract).
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := log.TraceID(ctx)

	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, Response{
				Code:    apperrors.ErrInternal,
				Msg:     "mysql unavailable",
				TraceID: traceID,
			})
			return
		}
	}
	if h.rdb != nil {
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, Response{
				Code:    apperrors.ErrInternal,
				Msg:     "redis unavailable",
				TraceID: traceID,
			})
			return
		}
	}

	Success(c, gin.H{"status": "ready"})
}
