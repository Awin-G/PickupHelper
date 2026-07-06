package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"pickup-helper/internal/config"
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload shared between auth (Phase 2) and downstream
// handlers. UserID > 0 means the request is authenticated.
type Claims struct {
	UserID    int64  `json:"user_id"`
	UserType  int    `json:"user_type,omitempty"`
	StationID int64  `json:"station_id,omitempty"`
	Role      string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// SignAccess signs an access token using cfg.JWT.AccessSecret.
func SignAccess(cfg *config.Config, claims Claims) (string, error) {
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(cfg.JWT.AccessTTL))
	}
	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(time.Now())
	}
	if claims.Issuer == "" {
		claims.Issuer = cfg.JWT.Issuer
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(cfg.JWT.AccessSecret))
}

// SignRefresh signs a refresh token using cfg.JWT.RefreshSecret.
func SignRefresh(cfg *config.Config, claims Claims) (string, error) {
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(cfg.JWT.RefreshTTL))
	}
	if claims.IssuedAt == nil {
		claims.IssuedAt = jwt.NewNumericDate(time.Now())
	}
	if claims.Issuer == "" {
		claims.Issuer = cfg.JWT.Issuer
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(cfg.JWT.RefreshSecret))
}

// ParseAccess parses and validates an access token against cfg.JWT.AccessSecret.
func ParseAccess(cfg *config.Config, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(cfg.JWT.AccessSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// ParseRefresh parses and validates a refresh token against
// cfg.JWT.RefreshSecret. Returns the wrapped error so callers can
// distinguish ErrTokenExpired (→ ErrRefreshExpired) from other failures
// (→ ErrRefreshInvalid).
func ParseRefresh(cfg *config.Config, tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(cfg.JWT.RefreshSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// CurrentUser extracts the authenticated user's ID from the gin context.
// Returns (0, false) when no authenticated user is present.
func CurrentUser(c *gin.Context) (int64, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	if !ok || id <= 0 {
		return 0, false
	}
	return id, true
}

// JWTAuth validates Bearer tokens in the Authorization header. On success
// it stores claims in the gin context and overrides X-User-Id / X-User-Type /
// X-Station-Id / X-Role headers so downstream services cannot be tricked by
// client-forged values. Requests without a Bearer token are rejected with
// 401 (code=10002). Use JWTAuthOptional for routes that allow public access.
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !requireBearer(c, cfg) {
			return
		}
		c.Next()
	}
}

// JWTAuthOptional allows requests without an Authorization header to pass
// through; if a header is present it must be a valid Bearer token.
func JWTAuthOptional(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		if !requireBearer(c, cfg) {
			return
		}
		c.Next()
	}
}

func requireBearer(c *gin.Context, cfg *config.Config) bool {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		jwtReject(c, "missing Authorization header")
		return false
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		jwtReject(c, "Authorization header must be Bearer <token>")
		return false
	}
	tokenStr := strings.TrimPrefix(authHeader, prefix)

	claims, err := ParseAccess(cfg, tokenStr)
	if err != nil {
		jwtReject(c, "token invalid or expired")
		return false
	}
	if claims.UserID <= 0 {
		jwtReject(c, "token missing user_id")
		return false
	}

	// Force header override — clients cannot forge identity headers.
	c.Request.Header.Set("X-User-Id", strconv.FormatInt(claims.UserID, 10))
	if claims.UserType > 0 {
		c.Request.Header.Set("X-User-Type", strconv.Itoa(claims.UserType))
	}
	if claims.StationID > 0 {
		c.Request.Header.Set("X-Station-Id", strconv.FormatInt(claims.StationID, 10))
	}
	if claims.Role != "" {
		c.Request.Header.Set("X-Role", claims.Role)
	}

	c.Set("user_id", claims.UserID)
	c.Set("user_type", claims.UserType)
	c.Set("station_id", claims.StationID)
	c.Set("role", claims.Role)
	return true
}

func jwtReject(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":     apperrors.ErrUnauthenticated,
		"msg":      msg,
		"trace_id": log.TraceID(c.Request.Context()),
	})
}
