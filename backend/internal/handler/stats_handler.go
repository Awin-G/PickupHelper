package handler

import (
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
)

// StatsHandler exposes dashboard, trend, and proxy-finance endpoints.
type StatsHandler struct {
	statsSvc *service.StatsService
}

func NewStatsHandler(ss *service.StatsService) *StatsHandler {
	return &StatsHandler{statsSvc: ss}
}

// StatsQuery is the common query params for stats endpoints.
type StatsQuery struct {
	StationID   int64  `form:"station_id" binding:"omitempty,min=1"`
	Granularity string `form:"granularity" binding:"omitempty,oneof=day week month year"`
}

// Dashboard handles GET /stats/dashboard.
func (h *StatsHandler) Dashboard(c *gin.Context) {
	_, _, stationID, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q StatsQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	if q.StationID > 0 {
		stationID = q.StationID
	}
	res, err := h.statsSvc.Dashboard(c.Request.Context(), stationID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// Trend handles GET /stats/trend.
func (h *StatsHandler) Trend(c *gin.Context) {
	_, _, stationID, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q StatsQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	if q.StationID > 0 {
		stationID = q.StationID
	}
	gran := q.Granularity
	if gran == "" {
		gran = "day"
	}
	res, err := h.statsSvc.Trend(c.Request.Context(), stationID, gran)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// ProxyFinance handles GET /stats/proxy-finance.
func (h *StatsHandler) ProxyFinance(c *gin.Context) {
	_, _, stationID, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var q StatsQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	if q.StationID > 0 {
		stationID = q.StationID
	}
	res, err := h.statsSvc.ProxyFinance(c.Request.Context(), stationID)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, res)
}

// RegisterStatsRoutes mounts stats routes (admin-only).
func (h *StatsHandler) RegisterStatsRoutes(g *gin.RouterGroup) {
	g.GET("/dashboard", h.Dashboard)
	g.GET("/trend", h.Trend)
	g.GET("/proxy-finance", h.ProxyFinance)
}
