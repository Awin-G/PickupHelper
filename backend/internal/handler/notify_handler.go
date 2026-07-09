package handler

import (
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// NotifyHandler exposes notification endpoints.
type NotifyHandler struct {
	notifySvc *service.NotifyService
}

func NewNotifyHandler(ns *service.NotifyService) *NotifyHandler {
	return &NotifyHandler{notifySvc: ns}
}

// NotifyListQuery is the query for GET /notifications.
type NotifyListQuery struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// ListNotifications handles GET /notifications.
func (h *NotifyHandler) ListNotifications(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q NotifyListQuery
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
	res, err := h.notifySvc.ListNotifications(c.Request.Context(), userID, offset, pageSize)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// MarkAllRead handles PUT /notifications/read.
func (h *NotifyHandler) MarkAllRead(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	if err := h.notifySvc.MarkAllRead(c.Request.Context(), userID); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"ok": true})
}

// RegisterNotifyRoutes mounts notification routes.
func (h *NotifyHandler) RegisterNotifyRoutes(g *gin.RouterGroup) {
	g.GET("", h.ListNotifications)
	g.PUT("/read", h.MarkAllRead)
}
