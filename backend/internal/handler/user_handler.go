package handler

import (
	"path"
	"strconv"
	"strings"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// UserHandler exposes the user-profile, runner-application, audit, and
// blacklist endpoints of the User module.
type UserHandler struct {
	userSvc *service.UserService
}

// NewUserHandler wires the handler with its UserService dependency.
func NewUserHandler(us *service.UserService) *UserHandler {
	return &UserHandler{userSvc: us}
}

// UpdateProfileRequest is the body for PUT /user/info.
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,max=50"`
	Avatar   string `json:"avatar" binding:"omitempty,url"`
}

// ApplyRunnerRequest is the body for POST /user/runner/apply.
type ApplyRunnerRequest struct {
	RealName    string `json:"real_name" binding:"required,max=50"`
	StudentID   string `json:"student_id" binding:"omitempty,max=50"`
	IDCardImage string `json:"id_card_image" binding:"required,url"`
}

// AuditRequest is the body for PUT /admin/user/runner/applications/:id/audit.
type AuditRequest struct {
	Action      string `json:"action" binding:"required,oneof=approve reject"`
	AuditRemark string `json:"audit_remark" binding:"omitempty,max=255"`
}

// SetBlacklistRequest is the body for PUT /admin/users/:id/blacklist.
type SetBlacklistRequest struct {
	IsBlacklisted bool   `json:"is_blacklisted" binding:"required"`
	Reason        string `json:"reason" binding:"omitempty,max=255"`
}

// ListRunnerAppsQuery is the query string for GET /admin/user/runner/applications.
type ListRunnerAppsQuery struct {
	Status   *int8  `form:"status" binding:"omitempty,oneof=1 2 3"`
	Keyword  string `form:"keyword" binding:"omitempty,max=50"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// GetInfo handles GET /user/info.
func (h *UserHandler) GetInfo(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	dto, err := h.userSvc.GetUserInfo(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// UpdateInfo handles PUT /user/info.
func (h *UserHandler) UpdateInfo(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var req UpdateProfileRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	dto, err := h.userSvc.UpdateUserInfo(c.Request.Context(), userID, req.Nickname, req.Avatar)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// ApplyRunner handles POST /user/runner/apply.
func (h *UserHandler) ApplyRunner(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var req ApplyRunnerRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	appID, status, err := h.userSvc.ApplyRunner(c.Request.Context(), userID, service.ApplyRunnerRequest{
		RealName:    req.RealName,
		StudentID:   req.StudentID,
		IDCardImage: req.IDCardImage,
	})
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"application_id": appID, "status": status})
}

// ListRunnerApps handles GET /admin/user/runner/applications.
func (h *UserHandler) ListRunnerApps(c *gin.Context) {
	var q ListRunnerAppsQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	res, err := h.userSvc.ListRunnerApps(c.Request.Context(), service.RunnerAppListFilter{
		Status:  q.Status,
		Keyword: q.Keyword,
		Offset:  offset,
		Limit:   pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, page, pageSize)
}

// AuditRunnerApp handles PUT /admin/user/runner/applications/:id/audit.
func (h *UserHandler) AuditRunnerApp(c *gin.Context) {
	adminID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing admin context"))
		return
	}
	appID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req AuditRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	dto, err := h.userSvc.AuditRunnerApp(c.Request.Context(), appID, adminID, req.Action, req.AuditRemark)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// SetBlacklist handles PUT /admin/users/:id/blacklist.
func (h *UserHandler) SetBlacklist(c *gin.Context) {
	targetID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req SetBlacklistRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	if err := h.userSvc.SetBlacklist(c.Request.Context(), targetID, req.IsBlacklisted, req.Reason); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"id": targetID})
}

// RegisterUserRoutes mounts /user/info (GET/PUT) and /user/runner/apply (POST)
// on the given user group (JWT required).
func (h *UserHandler) RegisterUserRoutes(g *gin.RouterGroup) {
	g.GET("/info", h.GetInfo)
	g.PUT("/info", h.UpdateInfo)
	g.POST("/runner/apply", h.ApplyRunner)
}

// RegisterAdminRoutes mounts the admin-only User-module routes on the
// given admin group (JWT + AdminOnly). Routes:
//   - GET    /user/runner/applications
//   - PUT    /user/runner/applications/:id/audit
//   - PUT    /users/:id/blacklist
//
// The admin group is conventionally mounted under /api/v1/admin, so the
// fully-qualified paths are /api/v1/admin/user/runner/applications etc.
func (h *UserHandler) RegisterAdminRoutes(g *gin.RouterGroup) {
	g.GET("/user/runner/applications", h.ListRunnerApps)
	g.PUT("/user/runner/applications/:id/audit", h.AuditRunnerApp)
	g.PUT("/users/:id/blacklist", h.SetBlacklist)
}

// parseIDParam extracts a positive int64 path parameter. Returns an
// ErrInvalidParam AppError when the param is missing or not a positive integer.
func parseIDParam(c *gin.Context, name string) (int64, error) {
	raw := path.Base(c.Param(name))
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == name {
		// gin sets the param to "" when missing; path.Base("") returns "."
		return 0, apperrors.New(apperrors.ErrInvalidParam, name+" required")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperrors.New(apperrors.ErrInvalidParam, name+" must be a positive integer")
	}
	return id, nil
}
