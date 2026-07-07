package middleware

import (
	"net/http"
	"runtime/debug"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
)

// Recovery catches panics raised inside downstream handlers, logs them with
// a stack trace, and returns a unified 500 response (code=10009).
// Without it, a panic would crash the whole process under gin.Default.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.From(c.Request.Context()).ErrorContext(
					c.Request.Context(),
					"panic recovered",
					"err", r,
					"stack", string(stack),
					"path", c.Request.URL.Path,
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    apperrors.ErrInternal,
					"msg":     apperrors.Msg(apperrors.ErrInternal),
					"trace_id": log.TraceID(c.Request.Context()),
				})
			}
		}()
		c.Next()
	}
}
