package handler

import (
	"context"
	"net/http"
	"testing"

	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// fakeStationSvc is a stub StationService for handler unit tests.
type fakeStationSvc struct {
	listFn   func(ctx context.Context, filter service.StationListFilter) (*service.StationListResult, error)
	getFn    func(ctx context.Context, id int64) (*models.StationDTO, error)
	createFn func(ctx context.Context, req service.CreateStationRequest) (*models.StationDTO, error)
	updateFn func(ctx context.Context, id int64, req service.UpdateStationRequest) (*models.StationDTO, error)
}

func (f *fakeStationSvc) ListStations(ctx context.Context, filter service.StationListFilter) (*service.StationListResult, error) {
	if f.listFn != nil {
		return f.listFn(ctx, filter)
	}
	return &service.StationListResult{Items: []*models.StationDTO{{ID: 1, Name: "站A"}}, Total: 1, Page: 1, Size: 20}, nil
}
func (f *fakeStationSvc) GetStation(ctx context.Context, id int64) (*models.StationDTO, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id)
	}
	return &models.StationDTO{ID: id, Name: "站B"}, nil
}
func (f *fakeStationSvc) CreateStation(ctx context.Context, req service.CreateStationRequest) (*models.StationDTO, error) {
	if f.createFn != nil {
		return f.createFn(ctx, req)
	}
	return &models.StationDTO{ID: 3, Name: req.Name}, nil
}
func (f *fakeStationSvc) UpdateStation(ctx context.Context, id int64, req service.UpdateStationRequest) (*models.StationDTO, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, id, req)
	}
	return &models.StationDTO{ID: id, Name: *req.Name}, nil
}

// newStationHandlerEngine mounts station routes with the fake service.
func newStationHandlerEngine(svc *fakeStationSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set("trace_id", "test-trace")
		c.Next()
	})
	g := engine.Group("/api/v1/stations")
	g.GET("", func(c *gin.Context) { callListStations(c, svc) })
	g.GET("/:id", func(c *gin.Context) { callGetStation(c, svc) })
	g.POST("", func(c *gin.Context) { callCreateStation(c, svc) })
	g.PUT("/:id", func(c *gin.Context) { callUpdateStation(c, svc) })
	return engine
}

func callListStations(c *gin.Context, svc *fakeStationSvc) {
	var q ListStationsQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	res, err := svc.ListStations(c.Request.Context(), service.StationListFilter{
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

func callGetStation(c *gin.Context, svc *fakeStationSvc) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	dto, err := svc.GetStation(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

func callCreateStation(c *gin.Context, svc *fakeStationSvc) {
	var req CreateStationRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	status := int8(1)
	if req.Status != nil {
		status = *req.Status
	}
	dto, err := svc.CreateStation(c.Request.Context(), service.CreateStationRequest{
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

func callUpdateStation(c *gin.Context, svc *fakeStationSvc) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req UpdateStationRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	dto, err := svc.UpdateStation(c.Request.Context(), id, service.UpdateStationRequest{
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

// --- Tests ---

func TestHandler_ListStations_Success(t *testing.T) {
	engine := newStationHandlerEngine(&fakeStationSvc{})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/stations?page=1&page_size=20", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(1), data["total"])
}

func TestHandler_GetStation_Success(t *testing.T) {
	engine := newStationHandlerEngine(&fakeStationSvc{})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/stations/5", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(5), data["id"])
}

func TestHandler_CreateStation_Success(t *testing.T) {
	engine := newStationHandlerEngine(&fakeStationSvc{})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/stations", CreateStationRequest{
		Name: "新站", Address: "某地", Latitude: 30.5, Longitude: 114.3,
	})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, "新站", data["name"])
}

func TestHandler_UpdateStation_Success(t *testing.T) {
	engine := newStationHandlerEngine(&fakeStationSvc{})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/stations/3", UpdateStationRequest{
		Name: strPtr("更新名"),
	})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, "更新名", data["name"])
}

func TestHandler_GetStation_BadID(t *testing.T) {
	engine := newStationHandlerEngine(&fakeStationSvc{})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/stations/abc", nil)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func strPtr(s string) *string { return &s }
