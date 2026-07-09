package handler

import (
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// ShelfHandler exposes shelf management endpoints.
type ShelfHandler struct {
	shelfSvc *service.ShelfService
}

func NewShelfHandler(ss *service.ShelfService) *ShelfHandler {
	return &ShelfHandler{shelfSvc: ss}
}

// CreateShelfRequest is the body for POST /shelves.
type CreateShelfRequest struct {
	StationID   int64  `json:"station_id" binding:"required,min=1"`
	ShelfCode   string `json:"shelf_code" binding:"required,max=20"`
	RowNum      int    `json:"row_num" binding:"required,min=1,max=99"`
	ColNum      int    `json:"col_num" binding:"required,min=1,max=99"`
	MaxCapacity int    `json:"max_capacity" binding:"required,min=1,max=9999"`
	Remark      string `json:"remark" binding:"omitempty,max=255"`
}

// UpdateShelfRequest is the body for PUT /shelves/:id.
type UpdateShelfRequest struct {
	ShelfCode   string `json:"shelf_code" binding:"omitempty,max=20"`
	RowNum      int    `json:"row_num" binding:"omitempty,min=0,max=99"`
	ColNum      int    `json:"col_num" binding:"omitempty,min=0,max=99"`
	MaxCapacity int    `json:"max_capacity" binding:"omitempty,min=0,max=9999"`
	Remark      string `json:"remark" binding:"omitempty,max=255"`
}

// ShelfListQuery is the query for GET /shelves and GET /shelves/occupancy.
type ShelfListQuery struct {
	StationID int64 `form:"station_id" binding:"omitempty,min=1"`
	Page      int   `form:"page" binding:"omitempty,min=1"`
	PageSize  int   `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// ListShelf handles GET /shelves.
func (h *ShelfHandler) ListShelf(c *gin.Context) {
	_, _, stationID, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q ShelfListQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	if q.StationID > 0 {
		stationID = q.StationID
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
	res, err := h.shelfSvc.ListShelf(c.Request.Context(), stationID, offset, pageSize)
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, page, pageSize)
}

// CreateShelf handles POST /shelves.
func (h *ShelfHandler) CreateShelf(c *gin.Context) {
	var req CreateShelfRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.shelfSvc.CreateShelf(c.Request.Context(), req.StationID,
		req.ShelfCode, req.RowNum, req.ColNum, req.MaxCapacity, req.Remark)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// UpdateShelf handles PUT /shelves/:id.
func (h *ShelfHandler) UpdateShelf(c *gin.Context) {
	shelfID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req UpdateShelfRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.shelfSvc.UpdateShelf(c.Request.Context(), shelfID,
		req.ShelfCode, req.RowNum, req.ColNum, req.MaxCapacity, req.Remark)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// Occupancy handles GET /shelves/occupancy.
func (h *ShelfHandler) Occupancy(c *gin.Context) {
	_, _, stationID, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q ShelfListQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	if q.StationID > 0 {
		stationID = q.StationID
	}
	res, err := h.shelfSvc.Occupancy(c.Request.Context(), stationID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// RegisterShelfRoutes mounts all shelf routes (admin-only).
func (h *ShelfHandler) RegisterShelfRoutes(g *gin.RouterGroup) {
	g.GET("", h.ListShelf)
	g.POST("", h.CreateShelf)
	g.PUT("/:id", h.UpdateShelf)
	g.GET("/occupancy", h.Occupancy)
}
