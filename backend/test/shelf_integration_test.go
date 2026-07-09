//go:build integration

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pickup-helper/internal/handler"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/router"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type shelfTestEnv struct {
	env       *TestEnv
	engine    *gin.Engine
	stationID int64
	adminTok  string
}

func setupShelfEngine(t *testing.T) *shelfTestEnv {
	t.Helper()
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	_ = middleware.Validator()
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	userRepo := repository.NewUserRepo()
	parcelRepo := repository.NewParcelRepo()
	shelfRepo := repository.NewShelfRepo()
	pickupRepo := repository.NewPickupLogRepo()
	proxyRepo := repository.NewProxyOrderRepo()

	parcelSvc := service.NewParcelService(parcelRepo, shelfRepo, userRepo, env.DB)
	pickupSvc := service.NewPickupService(parcelRepo, pickupRepo, shelfRepo, userRepo, env.DB)
	proxySvc := service.NewProxyService(proxyRepo, parcelRepo, userRepo, env.DB)
	shelfSvc := service.NewShelfService(shelfRepo, env.DB)

	stationID := SeedStation(t, env.DB)
	admin := SeedAdmin(t, env.DB, "shelf-admin")
	adminTok := signParcelToken(t, env, admin.ID, stationID, "admin")

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	parcelH := handler.NewParcelHandler(parcelSvc)
	pickupH := handler.NewPickupHandler(pickupSvc)
	proxyH := handler.NewProxyHandler(proxySvc)
	shelfH := handler.NewShelfHandler(shelfSvc)
	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Parcel: parcelH,
		Pickup: pickupH,
		Proxy:  proxyH,
		Shelf:  shelfH,
	})

	return &shelfTestEnv{env: env, engine: engine, stationID: stationID, adminTok: adminTok}
}

func (s *shelfTestEnv) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf strings.Builder
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, strings.NewReader(buf.String()))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	s.engine.ServeHTTP(rr, req)
	return rr
}

func shelfBodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

// SHELF-01: Create shelf → returns shelf DTO.
func TestShelf_01_Create_Success(t *testing.T) {
	env := setupShelfEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/shelves", map[string]any{
		"station_id":   env.stationID,
		"shelf_code":   "A-01",
		"row_num":      5,
		"col_num":      10,
		"max_capacity": 50,
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "create body=%s", rr.Body.String())
	b := shelfBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, "A-01", data["shelf_code"])
	assert.Equal(t, float64(50), data["max_capacity"])
}

// SHELF-02: Create duplicate shelf code → 409.
func TestShelf_02_Create_Duplicate(t *testing.T) {
	env := setupShelfEngine(t)

	body := map[string]any{"station_id": env.stationID, "shelf_code": "B-01", "row_num": 3, "col_num": 8, "max_capacity": 30}
	env.do(t, http.MethodPost, "/api/v1/shelves", body, env.adminTok)
	rr := env.do(t, http.MethodPost, "/api/v1/shelves", body, env.adminTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b := shelfBodyMap(t, rr)
	assert.Equal(t, float64(10501), b["code"])
}

// SHELF-03: List shelves → returns paginated list.
func TestShelf_03_List(t *testing.T) {
	env := setupShelfEngine(t)
	SeedShelf(t, env.env.DB, env.stationID, "A-01", 100)
	SeedShelf(t, env.env.DB, env.stationID, "B-02", 50)

	rr := env.do(t, http.MethodGet, "/api/v1/shelves?page=1&page_size=10", nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "list body=%s", rr.Body.String())
	b := shelfBodyMap(t, rr)
	data := b["data"].(map[string]any)
	list := data["list"].([]any)
	assert.Len(t, list, 2)
	assert.Equal(t, float64(2), data["total"])
}

// SHELF-04: Update shelf → returns updated DTO.
func TestShelf_04_Update(t *testing.T) {
	env := setupShelfEngine(t)
	shelfID := SeedShelf(t, env.env.DB, env.stationID, "A-01", 100)

	rr := env.do(t, http.MethodPut, "/api/v1/shelves/"+itoa(shelfID), map[string]any{
		"max_capacity": 200,
		"remark":       "expanded",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "update body=%s", rr.Body.String())
	b := shelfBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(200), data["max_capacity"])
}

// SHELF-05: Update non-existent shelf → 404.
func TestShelf_05_Update_NotFound(t *testing.T) {
	env := setupShelfEngine(t)

	rr := env.do(t, http.MethodPut, "/api/v1/shelves/99999", map[string]any{
		"max_capacity": 100,
	}, env.adminTok)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	b := shelfBodyMap(t, rr)
	assert.Equal(t, float64(10511), b["code"])
}

// SHELF-06: Occupancy heatmap → returns shelf list with heat levels.
func TestShelf_06_Occupancy(t *testing.T) {
	env := setupShelfEngine(t)
	SeedShelf(t, env.env.DB, env.stationID, "A-01", 100)
	SeedShelf(t, env.env.DB, env.stationID, "B-02", 50)

	rr := env.do(t, http.MethodGet, "/api/v1/shelves/occupancy?station_id="+itoa(env.stationID), nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "occupancy body=%s", rr.Body.String())
	b := shelfBodyMap(t, rr)
	data := b["data"].(map[string]any)
	shelves := data["shelves"].([]any)
	assert.Len(t, shelves, 2)
	assert.Equal(t, float64(0), data["total_used"])
	assert.Equal(t, float64(150), data["total_max"])

	item := shelves[0].(map[string]any)
	assert.Equal(t, float64(0), item["heat_level"])
}

// SHELF-07: Update shelf to change code → uniqueness verified.
func TestShelf_07_Update_CodeConflict(t *testing.T) {
	env := setupShelfEngine(t)
	SeedShelf(t, env.env.DB, env.stationID, "A-01", 100)
	shelfID := SeedShelf(t, env.env.DB, env.stationID, "B-02", 50)

	rr := env.do(t, http.MethodPut, "/api/v1/shelves/"+itoa(shelfID), map[string]any{
		"shelf_code": "A-01",
	}, env.adminTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b := shelfBodyMap(t, rr)
	assert.Equal(t, float64(10513), b["code"])
}

// SHELF-08: Update max_capacity below current → 409.
func TestShelf_08_Update_CapacityBelow(t *testing.T) {
	env := setupShelfEngine(t)
	// Seed a shelf and directly set current_capacity to 30.
	_ = SeedShelf(t, env.env.DB, env.stationID, "A-01", 100)
	_, err := env.env.DB.Exec("UPDATE shelf_layout SET current_capacity = 30 WHERE shelf_code = 'A-01'")
	require.NoError(t, err)

	// The shelfID might have changed - re-query.
	type shelfRow struct{ ID int64 }
	var row shelfRow
	require.NoError(t, env.env.DB.Get(&row, "SELECT id FROM shelf_layout WHERE shelf_code = 'A-01'"))

	rr := env.do(t, http.MethodPut, "/api/v1/shelves/"+itoa(row.ID), map[string]any{
		"max_capacity": 20,
	}, env.adminTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b := shelfBodyMap(t, rr)
	assert.Equal(t, float64(10512), b["code"])
}

// SHELF-09: List shelves unauthenticated → 401.
func TestShelf_09_List_Unauthenticated(t *testing.T) {
	env := setupShelfEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/shelves", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
