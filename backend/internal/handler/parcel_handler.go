package handler

import (
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// ParcelHandler exposes parcel intake, listing, detail, status-change,
// my-parcels, and pickup-code endpoints.
type ParcelHandler struct {
	parcelSvc *service.ParcelService
}

func NewParcelHandler(ps *service.ParcelService) *ParcelHandler {
	return &ParcelHandler{parcelSvc: ps}
}

// ScanInRequest is the body for POST /parcels/scan-in.
type ScanInRequest struct {
	TrackingNo    string  `json:"tracking_no" binding:"required,max=64"`
	CourierCompany string `json:"courier_company" binding:"required,max=50"`
	ReceiverPhone string  `json:"receiver_phone" binding:"required,phone_cn"`
	ReceiverName  string  `json:"receiver_name" binding:"omitempty,max=50"`
	ShelfCode     string  `json:"shelf_code" binding:"omitempty,max=20"`
	Weight        float64 `json:"weight" binding:"omitempty,min=0"`
	IsFragile     bool    `json:"is_fragile"`
	Remarks       string  `json:"remarks" binding:"omitempty,max=255"`
}

// ParcelListQuery is the query string for GET /parcels.
type ParcelListQuery struct {
	TrackingNo    string `form:"tracking_no" binding:"omitempty,max=64"`
	ReceiverPhone string `form:"receiver_phone" binding:"omitempty,max=11"`
	Status        *int8  `form:"status" binding:"omitempty,oneof=1 2 3 4 5"`
	CourierCompany string `form:"courier_company" binding:"omitempty,max=50"`
	ShelfCode     string `form:"shelf_code" binding:"omitempty,max=20"`
	StorageStart  string `form:"storage_start" binding:"omitempty"`
	StorageEnd    string `form:"storage_end" binding:"omitempty"`
	Page          int    `form:"page" binding:"omitempty,min=1"`
	PageSize      int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// MyParcelListQuery is the query string for GET /parcels/my.
type MyParcelListQuery struct {
	Status   *int8  `form:"status" binding:"omitempty,oneof=1 2 3 4"`
	Keyword  string `form:"keyword" binding:"omitempty,max=64"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// UpdateStatusRequest is the body for PUT /parcels/:id/status.
type UpdateStatusRequest struct {
	Status int8   `json:"status" binding:"required,min=1,max=5"`
	Reason string `json:"reason" binding:"omitempty,max=255"`
}

// ScanIn handles POST /parcels/scan-in.
func (h *ParcelHandler) ScanIn(c *gin.Context) {
	userID, _, stationID, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var req ScanInRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.parcelSvc.ScanIn(c.Request.Context(), service.ScanInRequest{
		StationID:      stationID,
		TrackingNo:     req.TrackingNo,
		CourierCompany: req.CourierCompany,
		ReceiverPhone:  req.ReceiverPhone,
		ReceiverName:   req.ReceiverName,
		ShelfCode:       req.ShelfCode,
		Weight:          req.Weight,
		IsFragile:       req.IsFragile,
		Remarks:          req.Remarks,
		OperatorID:       userID,
	})
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// ListParcels handles GET /parcels (admin view).
func (h *ParcelHandler) ListParcels(c *gin.Context) {
	_, _, stationID, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q ParcelListQuery
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
	res, err := h.parcelSvc.ListParcels(c.Request.Context(), service.ParcelListFilter{
		StationID:      stationID,
		TrackingNo:     q.TrackingNo,
		ReceiverPhone:  q.ReceiverPhone,
		Status:         q.Status,
		CourierCompany: q.CourierCompany,
		ShelfCode:       q.ShelfCode,
		StorageStart:   q.StorageStart,
		StorageEnd:     q.StorageEnd,
		Offset:         offset,
		Limit:          pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, page, pageSize)
}

// GetParcel handles GET /parcels/:id.
func (h *ParcelHandler) GetParcel(c *gin.Context) {
	userID, userType, stationID, role, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	parcelID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	isAdmin := role == "admin"
	_ = stationID
	_ = userType
	dto, err := h.parcelSvc.GetParcel(c.Request.Context(), parcelID, userID, isAdmin)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// UpdateStatus handles PUT /parcels/:id/status.
func (h *ParcelHandler) UpdateStatus(c *gin.Context) {
	parcelID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req UpdateStatusRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	dto, err := h.parcelSvc.UpdateStatus(c.Request.Context(), parcelID, service.UpdateStatusRequest{
		Status: req.Status,
		Reason: req.Reason,
	})
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// ListMyParcels handles GET /parcels/my.
func (h *ParcelHandler) ListMyParcels(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q MyParcelListQuery
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
	res, err := h.parcelSvc.ListMyParcels(c.Request.Context(), userID, q.Status, q.Keyword, offset, pageSize)
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, page, pageSize)
}

// GetPickupCode handles GET /parcels/:id/pickup-code.
func (h *ParcelHandler) GetPickupCode(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	parcelID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	dto, err := h.parcelSvc.GetPickupCode(c.Request.Context(), parcelID, userID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

// RegisterParcelAdminRoutes mounts the admin-only parcel CRUD routes.
// Routes:
//   - POST   /parcels/scan-in
//   - GET    /parcels
//   - PUT    /parcels/:id/status
func (h *ParcelHandler) RegisterParcelAdminRoutes(g *gin.RouterGroup) {
	g.POST("/scan-in", h.ScanIn)
	g.GET("", h.ListParcels)
	g.PUT("/:id/status", h.UpdateStatus)
}

// RegisterParcelUserRoutes mounts the user-facing parcel routes.
// Routes:
//   - GET  /parcels/my
//   - GET  /parcels/:id
//   - GET  /parcels/:id/pickup-code
func (h *ParcelHandler) RegisterParcelUserRoutes(g *gin.RouterGroup) {
	g.GET("/my", h.ListMyParcels)
	g.GET("/:id", h.GetParcel)
	g.GET("/:id/pickup-code", h.GetPickupCode)
}
