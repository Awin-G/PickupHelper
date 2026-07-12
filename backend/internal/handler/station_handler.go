package handler

import (
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// StationHandler exposes station CRUD endpoints.
type StationHandler struct {
	stationSvc *service.StationService
}

func NewStationHandler(ss *service.StationService) *StationHandler {
	return &StationHandler{stationSvc: ss}
}

// ListStationsQuery is the query for GET /stations.
type ListStationsQuery struct {
	Keyword  string `form:"keyword" binding:"omitempty,max=50"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// CreateStationRequest is the body for POST /stations.
type CreateStationRequest struct {
	Name          string  `json:"name" binding:"required,max=100"`
	Address       string  `json:"address" binding:"required,max=255"`
	Latitude      float64 `json:"latitude" binding:"required"`
	Longitude     float64 `json:"longitude" binding:"required"`
	BusinessHours string  `json:"business_hours" binding:"omitempty,max=100"`
	Status        *int8   `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateStationRequest is the body for PUT /stations/:id.
type UpdateStationRequest struct {
	Name          *string  `json:"name,omitempty" binding:"omitempty,max=100"`
	Address       *string  `json:"address,omitempty" binding:"omitempty,max=255"`
	Latitude      *float64 `json:"latitude,omitempty"`
	Longitude     *float64 `json:"longitude,omitempty"`
	BusinessHours *string  `json:"business_hours,omitempty" binding:"omitempty,max=100"`
	Status        *int8    `json:"status,omitempty" binding:"omitempty,oneof=0 1"`
}

// ListStations handles GET /stations.
func (h *StationHandler) ListStations(c *gin.Context) {
	var q ListStationsQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	res, err := h.stationSvc.ListStations(c.Request.Context(), service.StationListFilter{
		Keyword:  q.Keyword,
		Status:   q.Status,
		Page:     q.Page,
		PageSize: q.PageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, res.Page, res.Size)
}

// GetStation handles GET /stations/:id.
func (h *StationHandler) GetStation(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	dto, err := h.stationSvc.GetStation(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// CreateStation handles POST /stations.
func (h *StationHandler) CreateStation(c *gin.Context) {
	var req CreateStationRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	dto, err := h.stationSvc.CreateStation(c.Request.Context(), service.CreateStationRequest{
		Name:          req.Name,
		Address:       req.Address,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		BusinessHours: req.BusinessHours,
		Status:        status,
	})
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// UpdateStation handles PUT /stations/:id.
func (h *StationHandler) UpdateStation(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req UpdateStationRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	dto, err := h.stationSvc.UpdateStation(c.Request.Context(), id, service.UpdateStationRequest{
		Name:          req.Name,
		Address:       req.Address,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		BusinessHours: req.BusinessHours,
		Status:        req.Status,
	})
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// RegisterStationRoutes mounts station routes (admin-only).
func (h *StationHandler) RegisterStationRoutes(g *gin.RouterGroup) {
	g.GET("/stations", h.ListStations)
	g.GET("/stations/:id", h.GetStation)
	g.POST("/stations", h.CreateStation)
	g.PUT("/stations/:id", h.UpdateStation)
}
