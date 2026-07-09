package middleware

import (
	"net/http"
	"sync"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimit applies a per-client-IP token bucket limiter with the given
// qps and burst. Requests that exceed the bucket are rejected with 429
// (code=10006). Buckets are created lazily per IP and guarded by a mutex.
//
// Note: this is a single-instance limiter. For multi-instance deployments a
// shared store (Redis) should be used; that is deferred to a later phase.
// If qps <= 0 the limiter is disabled (pass-through).
func RateLimit(qps, burst int) gin.HandlerFunc {
	if qps <= 0 {
		return func(c *gin.Context) { c.Next() }
	}
	if burst <= 0 {
		burst = qps
	}

	var (
		mu      sync.Mutex
		buckets = make(map[string]*rate.Limiter)
		rps     = rate.Limit(qps)
	)

	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		l, ok := buckets[ip]
		if !ok {
			l = rate.NewLimiter(rps, burst)
			buckets[ip] = l
		}
		return l
	}

	return func(c *gin.Context) {
		if !getLimiter(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":     apperrors.ErrRateLimited,
				"msg":      apperrors.Msg(apperrors.ErrRateLimited),
				"trace_id": log.TraceID(c.Request.Context()),
			})
			return
		}
		c.Next()
	}
}
