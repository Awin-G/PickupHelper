// Package middleware contains shared HTTP middleware used by the router.
package middleware

import (
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
)

const traceIDHeader = "X-Trace-Id"

// TraceID reads the X-Trace-Id request header (generates a new one when
// missing), injects it into the request context, and writes it back as
// a response header so clients can correlate logs with requests.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := c.GetHeader(traceIDHeader)
		if tid == "" {
			tid = log.NewTraceID()
		}
		c.Request = c.Request.WithContext(log.WithTraceID(c.Request.Context(), tid))
		c.Header(traceIDHeader, tid)
		c.Next()
	}
}
