package middleware

import (
	"net/http"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
)

// AdminOnly rejects requests whose authenticated role is not "admin".
// It must run after JWTAuth, which populates the X-Role header. Requests
// without a valid admin role are aborted with 403 (code=10003).
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, _, role, ok := CurrentUser(c)
		if !ok || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":     apperrors.ErrForbidden,
				"msg":      apperrors.Msg(apperrors.ErrForbidden),
				"trace_id": log.TraceID(c.Request.Context()),
			})
			return
		}
		c.Next()
	}
}
