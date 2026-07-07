package handler

import (
	"strings"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// ProxyHandler exposes proxy order endpoints.
type ProxyHandler struct {
	proxySvc *service.ProxyService
}

func NewProxyHandler(ps *service.ProxyService) *ProxyHandler {
	return &ProxyHandler{proxySvc: ps}
}

// PublishRequest is the body for POST /proxy/publish.
type PublishRequest struct {
	ParcelID     int64   `json:"parcel_id" binding:"required,min=1"`
	RewardAmount float64 `json:"reward_amount" binding:"required,min=0.01,max=500"`
	Deadline     string  `json:"deadline" binding:"required"`
	Remark       string  `json:"remark" binding:"omitempty,max=255"`
}

// AcceptBody is not used (accept uses path param), define for API doc.
// Request delivery body.
type RequestDeliveryBody struct {
	DeliveryPhotos []string `json:"delivery_photos" binding:"required,min=1,max=5"`
	Remark         string   `json:"remark" binding:"omitempty,max=255"`
}

// ConfirmDeliveryBody is the body for POST /proxy/confirm-delivery/:id.
type ConfirmDeliveryBody struct {
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason" binding:"omitempty,max=255"`
}

// CancelBody is the body for POST /proxy/cancel.
type CancelBody struct {
	OrderID      int64  `json:"order_id" binding:"required,min=1"`
	CancelReason string `json:"cancel_reason" binding:"required,max=255"`
}

// TaskListQuery is the query for GET /proxy/tasks.
type TaskListQuery struct {
	StationID int64   `form:"station_id" binding:"omitempty,min=1"`
	MinReward float64 `form:"min_reward" binding:"omitempty,min=0"`
	Sort      string  `form:"sort" binding:"omitempty"`
	Page      int     `form:"page" binding:"omitempty,min=1"`
	PageSize  int     `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// MyOrdersQuery is the query for GET /proxy/my-orders.
type MyOrdersQuery struct {
	Role     *int8  `form:"role" binding:"omitempty,oneof=1 2 3 4 5 6"` // actually status filter
	Status   *int8  `form:"status" binding:"omitempty,oneof=1 2 3 4 5 6"`
	RoleType string  `form:"role_type" binding:"omitempty,oneof=publisher taker"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// Publish handles POST /proxy/publish.
func (h *ProxyHandler) Publish(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var req PublishRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.proxySvc.Publish(c.Request.Context(), userID, req.ParcelID,
		req.RewardAmount, req.Deadline, req.Remark)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// ListTasks handles GET /proxy/tasks.
func (h *ProxyHandler) ListTasks(c *gin.Context) {
	var q TaskListQuery
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
	res, err := h.proxySvc.ListTasks(c.Request.Context(), q.StationID, q.MinReward, q.Sort, offset, pageSize)
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, page, pageSize)
}

// Accept handles POST /proxy/accept/:id.
func (h *ProxyHandler) Accept(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	orderID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	res, err := h.proxySvc.Accept(c.Request.Context(), orderID, userID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// RequestDelivery handles POST /proxy/request-delivery/:id.
func (h *ProxyHandler) RequestDelivery(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	orderID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req RequestDeliveryBody
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.proxySvc.RequestDelivery(c.Request.Context(), orderID, userID, req.DeliveryPhotos, req.Remark)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// ConfirmDelivery handles POST /proxy/confirm-delivery/:id.
func (h *ProxyHandler) ConfirmDelivery(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	orderID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req ConfirmDeliveryBody
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.proxySvc.ConfirmDelivery(c.Request.Context(), orderID, userID, req.Accepted, req.Reason)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// CancelOrder handles POST /proxy/cancel.
func (h *ProxyHandler) CancelOrder(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var req CancelBody
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	res, err := h.proxySvc.CancelOrder(c.Request.Context(), req.OrderID, userID, req.CancelReason)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// ListMyOrders handles GET /proxy/my-orders.
func (h *ProxyHandler) ListMyOrders(c *gin.Context) {
	userID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q MyOrdersQuery
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
	res, err := h.proxySvc.ListMyOrders(c.Request.Context(), service.ProxyMyOrderFilter{
		UserID: userID,
		Role:   strings.TrimSpace(q.RoleType),
		Status: q.Status,
		Offset: offset,
		Limit:  pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, page, pageSize)
}

// RegisterProxyRoutes mounts all proxy routes (JWT required).
func (h *ProxyHandler) RegisterProxyRoutes(g *gin.RouterGroup) {
	g.POST("/publish", h.Publish)
	g.GET("/tasks", h.ListTasks)
	g.POST("/accept/:id", h.Accept)
	g.POST("/request-delivery/:id", h.RequestDelivery)
	g.POST("/confirm-delivery/:id", h.ConfirmDelivery)
	g.POST("/cancel", h.CancelOrder)
	g.GET("/my-orders", h.ListMyOrders)
}
