package handler

import (
	"net/http"
	"strconv"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// AvatarHandler exposes avatar upload and serve endpoints.
type AvatarHandler struct {
	avatarSvc *service.AvatarService
}

func NewAvatarHandler(as *service.AvatarService) *AvatarHandler {
	return &AvatarHandler{avatarSvc: as}
}

// Upload handles POST /user/avatar (multipart form).
// Accepts field "file"; max raw size 5MB; validates square, compresses to JPEG ≤150KB.
func (h *AvatarHandler) Upload(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		Error(c, apperrors.New(apperrors.ErrInvalidParam, "缺少 file 字段"))
		return
	}

	fd, err := file.Open()
	if err != nil {
		Error(c, apperrors.New(apperrors.ErrInvalidParam, "无法打开文件"))
		return
	}
	defer fd.Close()

	if err := h.avatarSvc.UploadAvatar(c.Request.Context(), userID, fd); err != nil {
		Error(c, err)
		return
	}

	// Return avatar URL (serve endpoint) as the new avatar value.
	Success(c, gin.H{"avatar_url": "/api/v1/user/avatar"})
}

// Serve handles GET /user/avatar — returns the raw JPEG image.
func (h *AvatarHandler) Serve(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}

	data, ct, err := h.avatarSvc.GetAvatar(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	c.Writer.Header().Set("Content-Type", ct)
	c.Writer.Header().Set("Content-Length", strconv.Itoa(len(data)))
	c.Writer.Header().Set("Cache-Control", "private, max-age=3600")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(data)
}

// ServeByID handles GET /users/:id/avatar — serves any user's avatar by ID.
func (h *AvatarHandler) ServeByID(c *gin.Context) {
	userID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	data, ct, err := h.avatarSvc.GetAvatar(c.Request.Context(), userID)
	if err != nil {
		Error(c, err)
		return
	}

	c.Writer.Header().Set("Content-Type", ct)
	c.Writer.Header().Set("Content-Length", strconv.Itoa(len(data)))
	c.Writer.Header().Set("Cache-Control", "public, max-age=3600")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write(data)
}

// RegisterAvatarRoutes mounts avatar routes.
func (h *AvatarHandler) RegisterAvatarRoutes(g *gin.RouterGroup) {
	g.POST("/user/avatar", h.Upload)
	g.GET("/user/avatar", h.Serve)
	g.GET("/users/:id/avatar", h.ServeByID)
}
