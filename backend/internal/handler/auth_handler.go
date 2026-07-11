package handler

import (
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler exposes the User module's authentication endpoints.
type AuthHandler struct {
	authSvc   *service.AuthService
	wechatSvc *service.WechatService
}

func NewAuthHandler(as *service.AuthService, ws *service.WechatService) *AuthHandler {
	return &AuthHandler{authSvc: as, wechatSvc: ws}
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

// WechatLoginRequest is the body for POST /auth/wechat-login.
type WechatLoginRequest struct {
	Code      string `json:"code" binding:"required,max=128"`
	PhoneCode string `json:"phone_code" binding:"omitempty,max=128"`
	Nickname  string `json:"nickname" binding:"omitempty,max=50"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,max=500"`
}

// WechatLogin handles POST /auth/wechat-login.
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	openid, _, err := h.wechatSvc.Code2Session(c.Request.Context(), req.Code)
	if err != nil {
		Error(c, err)
		return
	}

	// Try login by openid first.
	res, err := h.authSvc.LoginByOpenID(c.Request.Context(), openid)
	if err == nil {
		Success(c, res)
		return
	}

	// New user: phone_code is required.
	if req.PhoneCode == "" {
		Error(c, apperrors.New(apperrors.ErrInvalidParam, "首次登录需要 phone_code"))
		return
	}

	phone, err := h.wechatSvc.GetPhoneNumber(c.Request.Context(), req.PhoneCode)
	if err != nil {
		Error(c, err)
		return
	}

	res, err = h.authSvc.RegisterByWechat(c.Request.Context(), openid, phone, req.Nickname, req.AvatarURL)
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
	g.POST("/wechat-login", h.WechatLogin)
}

// RegisterAdminAuthRoutes mounts /admin/auth/login on the given public
// group. Despite the /admin prefix, admin login is public — AdminOnly
// middleware must NOT be applied.
func (h *AuthHandler) RegisterAdminAuthRoutes(g *gin.RouterGroup) {
	g.POST("/admin/auth/login", h.AdminLogin)
}

// ListActiveCodesResponse is the body returned by GET /admin/auth/sms-codes.
type ListActiveCodesResponse struct {
	Codes []repository.ActiveCode `json:"codes"`
	Total int                     `json:"total"`
}

// ListActiveCodes handles GET /admin/auth/sms-codes.
func (h *AuthHandler) ListActiveCodes(c *gin.Context) {
	codes, err := h.authSvc.ListActiveCodes(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, ListActiveCodesResponse{Codes: codes, Total: len(codes)})
}

// RegisterAdminAuthManagementRoutes mounts admin-only auth management
// routes (JWT + AdminOnly required).
func (h *AuthHandler) RegisterAdminAuthManagementRoutes(g *gin.RouterGroup) {
	g.GET("/auth/sms-codes", h.ListActiveCodes)
}
