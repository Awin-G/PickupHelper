package middleware

import (
	"time"

	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
)

// Logger records one structured log line per HTTP request: method, path,
// status, latency, client IP, and trace_id (injected by the log handler).
// It writes at INFO level for 2xx/3xx, WARN for 4xx, ERROR for 5xx.
// Uses defer so the line is emitted even when a downstream handler panics
// (Recovery will catch the panic and write a 500; Logger still logs it).
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		defer func() {
			latency := time.Since(start)
			status := c.Writer.Status()

			attrs := []any{
				"method", c.Request.Method,
				"path", path,
				"query", query,
				"status", status,
				"latency_ms", latency.Milliseconds(),
				"client_ip", c.ClientIP(),
				"bytes", c.Writer.Size(),
			}
			if len(c.Errors) > 0 {
				attrs = append(attrs, "errors", c.Errors.String())
			}

			ctx := c.Request.Context()
			switch {
			case status >= 500:
				log.From(ctx).ErrorContext(ctx, "http", attrs...)
			case status >= 400:
				log.From(ctx).WarnContext(ctx, "http", attrs...)
			default:
				log.From(ctx).InfoContext(ctx, "http", attrs...)
			}
		}()

		c.Next()
	}
}
