package handler

import (
	"strings"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// PickupHandler exposes pickup verification, self-checkout, scan-station,
// and pickup log endpoints.
type PickupHandler struct {
	pickupSvc *service.PickupService
}

func NewPickupHandler(ps *service.PickupService) *PickupHandler {
	return &PickupHandler{pickupSvc: ps}
}

// VerifyRequest is the body for POST /pickup/verify.
type VerifyRequest struct {
	PickupCode         string `json:"pickup_code" binding:"required,max=10"`
	VerificationMethod int8   `json:"verification_method" binding:"required,oneof=1 2 3"`
	StationID          int64  `json:"station_id" binding:"required,min=1"`
}

// SelfCheckoutRequest is the body for POST /pickup/self-checkout.
type SelfCheckoutRequest struct {
	PickupCode string `json:"pickup_code" binding:"required,max=10"`
	StationID  int64  `json:"station_id" binding:"required,min=1"`
}

// ScanStationRequest is the body for POST /pickup/scan-station.
type ScanStationRequest struct {
	StationQR   string   `json:"station_qr" binding:"required,max=255"`
	PickupCodes []string `json:"pickup_codes" binding:"required,min=1,max=10,dive,max=10"`
}

// PickupLogQuery is the query string for GET /pickup/logs.
type PickupLogQuery struct {
	ParcelID     *int64 `form:"parcel_id" binding:"omitempty,min=1"`
	OperatorID   *int64 `form:"operator_id" binding:"omitempty,min=1"`
	OperatorType *int8  `form:"operator_type" binding:"omitempty,oneof=1 2 3 4"`
	Start        string `form:"start" binding:"omitempty"`
	End          string `form:"end" binding:"omitempty"`
	Page         int    `form:"page" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// Verify handles POST /pickup/verify.
func (h *PickupHandler) Verify(c *gin.Context) {
	userID, userType, stationID, role, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	_ = stationID
	_ = role

	var req VerifyRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	code, err := parsePickupCode(req.PickupCode)
	if err != nil {
		Error(c, err)
		return
	}

	opType := int8(models.OpTypeAdmin)
	if userType == int(models.UserTypeRunner) {
		opType = int8(models.OpTypeRunner)
	}

	res, err := h.pickupSvc.Verify(
		c.Request.Context(),
		code, req.StationID, req.VerificationMethod,
		userID, opType,
		0, 0, c.ClientIP(), c.Request.UserAgent(),
	)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// SelfCheckout handles POST /pickup/self-checkout.
func (h *PickupHandler) SelfCheckout(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}

	var req SelfCheckoutRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	code, err := parsePickupCode(req.PickupCode)
	if err != nil {
		Error(c, err)
		return
	}

	res, err := h.pickupSvc.SelfCheckout(
		c.Request.Context(),
		code, req.StationID, userID,
		0, 0, c.ClientIP(), c.Request.UserAgent(),
	)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// ScanStation handles POST /pickup/scan-station.
func (h *PickupHandler) ScanStation(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}

	var req ScanStationRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}

	result := h.pickupSvc.ScanStation(
		c.Request.Context(),
		userID, req.PickupCodes,
		0, 0, c.ClientIP(), c.Request.UserAgent(),
	)
	Success(c, result)
}

// ListLogs handles GET /pickup/logs.
func (h *PickupHandler) ListLogs(c *gin.Context) {
	var q PickupLogQuery
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

	res, err := h.pickupSvc.ListLogs(c.Request.Context(), service.PickupLogFilter{
		ParcelID:     q.ParcelID,
		OperatorID:   q.OperatorID,
		OperatorType: q.OperatorType,
		Start:        q.Start,
		End:          q.End,
		Offset:       offset,
		Limit:        pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, page, pageSize)
}

// RegisterPickupRoutes mounts all pickup routes.
func (h *PickupHandler) RegisterPickupRoutes(authedGroup *gin.RouterGroup) {
	// Admin/runner: verify + logs.
	authedGroup.POST("/verify", h.Verify)
	authedGroup.GET("/logs", h.ListLogs)
	// User: self-checkout + scan-station.
	authedGroup.POST("/self-checkout", h.SelfCheckout)
	authedGroup.POST("/scan-station", h.ScanStation)
}

// parsePickupCode converts a string pickup code to int64.
func parsePickupCode(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 10 {
		return 0, apperrors.New(apperrors.ErrInvalidParam, "pickup_code 非法")
	}
	var code int64
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, apperrors.New(apperrors.ErrInvalidParam, "pickup_code 必须为数字")
		}
		code = code*10 + int64(ch-'0')
	}
	return code, nil
}
