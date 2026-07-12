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

// GetAsyncRoutes returns the dynamic route tree for the admin frontend
// (Pure Admin style). For now returns a fixed set of admin menu routes.
func (h *AuthHandler) GetAsyncRoutes(c *gin.Context) {
	routes := []gin.H{
		{
			"name":      "dashboard",
			"path":      "/dashboard",
			"component": "/dashboard/index",
			"meta": gin.H{
				"title": "看板",
				"icon":  "ant-design:dashboard-outlined",
				"rank":  1,
			},
		},
		{
			"name":      "parcel",
			"path":      "/parcel",
			"component": "/parcel/index",
			"meta": gin.H{
				"title": "包裹管理",
				"icon":  "ant-design:inbox-outlined",
				"rank":  2,
			},
			"children": []gin.H{
				{"name": "parcel-list", "path": "/parcel/list", "component": "/parcel/list", "meta": gin.H{"title": "包裹列表"}},
				{"name": "parcel-scan-in", "path": "/parcel/scan-in", "component": "/parcel/scan-in", "meta": gin.H{"title": "扫码入库"}},
			},
		},
		{
			"name":      "user",
			"path":      "/user",
			"component": "/user/index",
			"meta": gin.H{
				"title": "用户管理",
				"icon":  "ant-design:user-outlined",
				"rank":  3,
			},
			"children": []gin.H{
				{"name": "user-list", "path": "/user/list", "component": "/user/list", "meta": gin.H{"title": "用户列表"}},
				{"name": "runner-audit", "path": "/user/runner-audit", "component": "/user/runner-audit", "meta": gin.H{"title": "跑腿员审核"}},
			},
		},
		{
			"name":      "station",
			"path":      "/station",
			"component": "/station/index",
			"meta": gin.H{
				"title": "驿站管理",
				"icon":  "ant-design:shop-outlined",
				"rank":  4,
			},
			"children": []gin.H{
				{"name": "station-list", "path": "/station/list", "component": "/station/list", "meta": gin.H{"title": "驿站列表"}},
			},
		},
		{
			"name":      "shelf",
			"path":      "/shelf",
			"component": "/shelf/index",
			"meta": gin.H{
				"title": "货架管理",
				"icon":  "ant-design:table-outlined",
				"rank":  5,
			},
			"children": []gin.H{
				{"name": "shelf-list", "path": "/shelf/list", "component": "/shelf/list", "meta": gin.H{"title": "货架列表"}},
			},
		},
		{
			"name":      "stats",
			"path":      "/stats",
			"component": "/stats/index",
			"meta": gin.H{
				"title": "数据统计",
				"icon":  "ant-design:bar-chart-outlined",
				"rank":  6,
			},
			"children": []gin.H{
				{"name": "stats-dashboard", "path": "/stats/dashboard", "component": "/stats/dashboard", "meta": gin.H{"title": "统计看板"}},
				{"name": "stats-courier", "path": "/stats/courier", "component": "/stats/courier", "meta": gin.H{"title": "快递对账"}},
			},
		},
		{
			"name":      "sms",
			"path":      "/sms",
			"component": "/sms/index",
			"meta": gin.H{
				"title": "短信管理",
				"icon":  "ant-design:message-outlined",
				"rank":  7,
			},
			"children": []gin.H{
				{"name": "sms-codes", "path": "/sms/codes", "component": "/sms/codes", "meta": gin.H{"title": "验证码列表"}},
			},
		},
	}
	Success(c, gin.H{"routes": routes})
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
	g.GET("/get-async-routes", h.GetAsyncRoutes)
}
