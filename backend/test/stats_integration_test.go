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

type statsTestEnv struct {
	env       *TestEnv
	engine    *gin.Engine
	stationID int64
	adminTok  string
	userTok   string
}

func setupStatsEngine(t *testing.T) *statsTestEnv {
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
	notifyRepo := repository.NewNotifyRepo()

	parcelSvc := service.NewParcelService(parcelRepo, shelfRepo, userRepo, env.DB)
	pickupSvc := service.NewPickupService(parcelRepo, pickupRepo, shelfRepo, userRepo, env.DB)
	proxySvc := service.NewProxyService(proxyRepo, parcelRepo, userRepo, env.DB)
	shelfSvc := service.NewShelfService(shelfRepo, env.DB)
	notifySvc := service.NewNotifyService(notifyRepo, env.DB)
	statsSvc := service.NewStatsService(env.DB)

	stationID := SeedStation(t, env.DB)
	SeedShelf(t, env.DB, stationID, "A-01", 100)

	admin := SeedAdmin(t, env.DB, "stats-admin")
	userID := SeedUser(t, env.DB, "13800138600")
	adminTok := signParcelToken(t, env, admin.ID, stationID, "admin")
	userTok := signParcelToken(t, env, userID, 0, "user")

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	parcelH := handler.NewParcelHandler(parcelSvc)
	pickupH := handler.NewPickupHandler(pickupSvc)
	proxyH := handler.NewProxyHandler(proxySvc)
	shelfH := handler.NewShelfHandler(shelfSvc)
	notifyH := handler.NewNotifyHandler(notifySvc)
	statsH := handler.NewStatsHandler(statsSvc)
	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Parcel: parcelH,
		Pickup: pickupH,
		Proxy:  proxyH,
		Shelf:  shelfH,
		Notify: notifyH,
		Stats:  statsH,
	})

	return &statsTestEnv{env: env, engine: engine, stationID: stationID, adminTok: adminTok, userTok: userTok}
}

func (s *statsTestEnv) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
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

func statsBodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

// STATS-01: Dashboard returns count data.
func TestStats_01_Dashboard(t *testing.T) {
	env := setupStatsEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/stats/dashboard?station_id="+itoa(env.stationID), nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "dashboard body=%s", rr.Body.String())
	b := statsBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(0), data["today_inbound"])
	assert.Equal(t, float64(0), data["pending_count"])
	assert.NotNil(t, data["shelf_usage_rate"])
}

// STATS-02: Dashboard non-admin → 403.
func TestStats_02_Dashboard_NonAdmin(t *testing.T) {
	env := setupStatsEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/stats/dashboard?station_id="+itoa(env.stationID), nil, env.userTok)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// STATS-03: Trend returns empty points when no data.
func TestStats_03_Trend_Empty(t *testing.T) {
	env := setupStatsEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/stats/trend?station_id="+itoa(env.stationID)+"&granularity=day", nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "trend body=%s", rr.Body.String())
	b := statsBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, "day", data["granularity"])
}

// STATS-04: Proxy finance returns zeros when no orders.
func TestStats_04_ProxyFinance_Empty(t *testing.T) {
	env := setupStatsEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/stats/proxy-finance?station_id="+itoa(env.stationID), nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "finance body=%s", rr.Body.String())
	b := statsBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(0), data["total_orders"])
	assert.Equal(t, float64(0), data["total_amount"])
}

// STATS-05: Dashboard unauthenticated → 401.
func TestStats_05_Unauthenticated(t *testing.T) {
	env := setupStatsEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/stats/dashboard", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// STATS-06: Trend with different granularity → works.
func TestStats_06_Trend_Month(t *testing.T) {
	env := setupStatsEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/stats/trend?station_id="+itoa(env.stationID)+"&granularity=month", nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := statsBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, "month", data["granularity"])
}
