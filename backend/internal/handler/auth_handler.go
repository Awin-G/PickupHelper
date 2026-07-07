package handler

import (
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler exposes the User module's authentication endpoints.
// All routes are public (no JWT required) — admin login is also public
// because the caller has not yet authenticated.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler wires the handler with its AuthService dependency.
func NewAuthHandler(as *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: as}
}

// SendCodeRequest is the body for POST /auth/send-code.
type SendCodeRequest struct {
	Phone string `json:"phone" binding:"required,phone_cn"`
	Scene string `json:"scene" binding:"omitempty,oneof=login runner_apply"`
}

// SendCodeResponse is the body returned by POST /auth/send-code.
type SendCodeResponse struct {
	ExpireIn int `json:"expire_in"`
}

// LoginRequest is the body for POST /auth/login.
type LoginRequest struct {
	Phone  string `json:"phone" binding:"required,phone_cn"`
	Code   string `json:"code" binding:"required,len=6"`
	OpenID string `json:"openid" binding:"omitempty,max=64"`
}

// RefreshRequest is the body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AdminLoginRequest is the body for POST /admin/auth/login.
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// SendCode handles POST /auth/send-code.
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req SendCodeRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	expireIn, err := h.authSvc.SendCode(c.Request.Context(), req.Phone, c.ClientIP())
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, SendCodeResponse{ExpireIn: expireIn})
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.authSvc.Login(c.Request.Context(), req.Phone, req.Code, req.OpenID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.authSvc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// AdminLogin handles POST /admin/auth/login.
func (h *AuthHandler) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.authSvc.AdminLogin(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// RegisterPublicRoutes mounts /auth/send-code, /auth/login, /auth/refresh
// on the given public group (no JWT middleware).
func (h *AuthHandler) RegisterPublicRoutes(g *gin.RouterGroup) {
	g.POST("/send-code", h.SendCode)
	g.POST("/login", h.Login)
	g.POST("/refresh", h.Refresh)
}

// RegisterAdminAuthRoutes mounts /admin/auth/login on the given public
// group. Despite the /admin prefix, admin login is public — AdminOnly
// middleware must NOT be applied.
func (h *AuthHandler) RegisterAdminAuthRoutes(g *gin.RouterGroup) {
	g.POST("/admin/auth/login", h.AdminLogin)
}
