package middleware

import (
	"net/http"
	"strings"

	"pickup-helper/internal/config"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowOrigin      = "Access-Control-Allow-Origin"
	corsAllowMethods     = "Access-Control-Allow-Methods"
	corsAllowHeaders     = "Access-Control-Allow-Headers"
	corsAllowCredentials = "Access-Control-Allow-Credentials"
	corsMaxAge           = "Access-Control-Max-Age"
	corsExposeHeaders    = "Access-Control-Expose-Headers"
)

// CORS handles preflight OPTIONS requests and sets the appropriate
// Access-Control-* headers on actual responses. allowedOrigins is the
// configured whitelist; the special value "*" allows any origin (without
// credentials). When the request Origin matches a configured entry, that
// origin is echoed back and credentials are allowed.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowAny := false
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o == "*" {
			allowAny = true
		}
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		switch {
		case allowAny:
			c.Header(corsAllowOrigin, "*")
		case origin != "":
			if _, ok := allowed[origin]; ok {
				c.Header(corsAllowOrigin, origin)
				c.Header(corsAllowCredentials, "true")
				c.Header("Vary", "Origin")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.Header(corsAllowMethods, "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header(corsAllowHeaders, "Content-Type, Authorization, X-Client-Type, X-Client-Version, X-Device-Id, X-Geo-Lat, X-Geo-Lng, Idempotency-Key, X-Trace-Id")
			c.Header(corsMaxAge, "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Expose trace id so clients can correlate logs.
		c.Header(corsExposeHeaders, "X-Trace-Id")
		c.Next()
	}
}

// CORSFromConfig is a convenience wrapper that pulls the allowed origins
// from a CORSConfig (used by router.Register).
func CORSFromConfig(cfg config.CORSConfig) gin.HandlerFunc {
	return CORS(cfg.AllowedOrigins)
}
