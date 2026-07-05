package handler

import (
	"context"
	"net/http"

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

// codeInternalError mirrors the unified error code scheme from api详细设计.md (10009).
const codeInternalError = 10009

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
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"status": "up"},
		"trace_id": log.TraceID(c.Request.Context()),
	})
}

// Ready is the readiness probe — checks MySQL and Redis connectivity.
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	traceID := log.TraceID(ctx)

	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    codeInternalError,
				"msg":     "mysql unavailable",
				"trace_id": traceID,
			})
			return
		}
	}
	if h.rdb != nil {
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    codeInternalError,
				"msg":     "redis unavailable",
				"trace_id": traceID,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"status": "ready"},
		"trace_id": traceID,
	})
}
