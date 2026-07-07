//go:build integration

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pickup-helper/internal/handler"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/router"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type proxyTestEnv struct {
	env       *TestEnv
	engine    *gin.Engine
	stationID int64
	userID    int64
	userTok   string
	runnerID  int64
	runnerTok string
}

func setupProxyEngine(t *testing.T) *proxyTestEnv {
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

	stationID := SeedStation(t, env.DB)
	SeedShelf(t, env.DB, stationID, "A-01", 100)

	userID := SeedUser(t, env.DB, "13800138800")
	runner := SeedUserWithStatus(t, env.DB, "13900138801", models.UserTypeRunner, models.RunnerStatusApproved)
	runnerID := runner.ID

	userTok := signParcelToken(t, env, userID, stationID, "user")
	runnerTok := signParcelToken(t, env, runnerID, stationID, "user")

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	parcelH := handler.NewParcelHandler(parcelSvc)
	pickupH := handler.NewPickupHandler(pickupSvc)
	proxyH := handler.NewProxyHandler(proxySvc)
	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Parcel: parcelH,
		Pickup: pickupH,
		Proxy:  proxyH,
	})

	return &proxyTestEnv{
		env:       env,
		engine:    engine,
		stationID: stationID,
		userID:    userID,
		userTok:   userTok,
		runnerID:  runnerID,
		runnerTok: runnerTok,
	}
}

func (p *proxyTestEnv) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
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
	p.engine.ServeHTTP(rr, req)
	return rr
}

func proxyBodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

// Helper: scan in a parcel for our user and return parcel ID + pickup code.
func scanProxyParcel(t *testing.T, env *proxyTestEnv, trackingNo string) (int64, string) {
	t.Helper()
	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     trackingNo,
		"courier_company": "顺丰速运",
		"receiver_phone":  "13800138800",
		"station_id":      env.stationID,
	}, signParcelToken(t, env.env, env.userID, env.stationID, "admin"))
	require.Equal(t, http.StatusOK, rr.Code, "scan-in body=%s", rr.Body.String())
	b := proxyBodyMap(t, rr)
	d := b["data"].(map[string]any)
	return int64(d["parcel_id"].(float64)), d["pickup_code"].(string)
}

// Use an admin-scoped token for parsing stationID-based scan-in.
func adminScanToken(t *testing.T, env *proxyTestEnv) string {
	t.Helper()
	admin := SeedAdmin(t, env.env.DB, "proxy-admin")
	return signParcelToken(t, env.env, admin.ID, env.stationID, "admin")
}

func scanProxyParcelAdmin(t *testing.T, env *proxyTestEnv, tok, trackingNo, phone string) (int64, string) {
	t.Helper()
	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     trackingNo,
		"courier_company": "顺丰速运",
		"receiver_phone":  phone,
	}, tok)
	require.Equal(t, http.StatusOK, rr.Code, "scan-in body=%s", rr.Body.String())
	b := proxyBodyMap(t, rr)
	d := b["data"].(map[string]any)
	return int64(d["parcel_id"].(float64)), d["pickup_code"].(string)
}

// PROXY-01: Publish proxy order → returns order_id.
func TestProxy_01_Publish_Success(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-PROXY-001", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "publish body=%s", rr.Body.String())
	b := proxyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.NotEmpty(t, data["order_id"])
	assert.Equal(t, float64(models.ProxyStatusPending), data["status"])
}

// PROXY-02: Publish non-owner parcel → 409.
func TestProxy_02_Publish_NotOwner(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-PROXY-002", "13900139990")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, env.userTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b := proxyBodyMap(t, rr)
	assert.Equal(t, float64(10401), b["code"])
}

// PROXY-03: Runner accepts a published order → delivery status.
func TestProxy_03_Accept_Success(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-PROXY-003", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 15.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	rr = env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.runnerTok)
	require.Equal(t, http.StatusOK, rr.Code, "accept body=%s", rr.Body.String())
	b = proxyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(models.ProxyStatusDelivering), data["status"])
	assert.NotEmpty(t, data["temp_pickup_code"])
}

// PROXY-04: Runner who is normal user cannot accept → 403.
func TestProxy_04_Accept_NotRunner(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-PROXY-004", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	rr = env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.userTok)
	assert.Equal(t, http.StatusForbidden, rr.Code)
	b = proxyBodyMap(t, rr)
	assert.Equal(t, float64(10413), b["code"])
}

// PROXY-05: Request delivery → confirm status.
func TestProxy_05_Delivery_Success(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-PROXY-005", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 20.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.runnerTok)

	rr = env.do(t, http.MethodPost, "/api/v1/proxy/request-delivery/"+itoa(orderID), map[string]any{
		"delivery_photos": []string{"https://cdn/photo1.jpg"},
	}, env.runnerTok)
	require.Equal(t, http.StatusOK, rr.Code, "delivery body=%s", rr.Body.String())
	b = proxyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(models.ProxyStatusConfirm), data["status"])
	assert.NotEmpty(t, data["delivery_time"])
}

// PROXY-06: Confirm delivery → done.
func TestProxy_06_Confirm_Success(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-PROXY-006", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 5.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.runnerTok)
	env.do(t, http.MethodPost, "/api/v1/proxy/request-delivery/"+itoa(orderID), map[string]any{
		"delivery_photos": []string{"https://cdn/photo1.jpg"},
	}, env.runnerTok)

	rr = env.do(t, http.MethodPost, "/api/v1/proxy/confirm-delivery/"+itoa(orderID), map[string]any{
		"accepted": true,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "confirm body=%s", rr.Body.String())
	b = proxyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(models.ProxyStatusDone), data["status"])
}

// PROXY-07: Cancel order → cancelled.
func TestProxy_07_Cancel_Success(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-PROXY-007", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 8.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	rr = env.do(t, http.MethodPost, "/api/v1/proxy/cancel", map[string]any{
		"order_id":      orderID,
		"cancel_reason": "不需要了",
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "cancel body=%s", rr.Body.String())
	b = proxyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(models.ProxyStatusCancelled), data["status"])
}

// PROXY-08: My orders list → returns own orders.
func TestProxy_08_MyOrders(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-PROXY-008", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 12.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)

	rr = env.do(t, http.MethodGet, "/api/v1/proxy/my-orders?page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "my-orders body=%s", rr.Body.String())
	b := proxyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	list := data["list"].([]any)
	assert.Len(t, list, 1)
}

// PROXY-09: Task hall list → returns published tasks.
func TestProxy_09_Tasks_List(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-TASK-001", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 25.0,
		"deadline":      deadline,
	}, env.userTok)

	rr := env.do(t, http.MethodGet, "/api/v1/proxy/tasks?page=1&page_size=10", nil, env.runnerTok)
	require.Equal(t, http.StatusOK, rr.Code, "tasks body=%s", rr.Body.String())
	b := proxyBodyMap(t, rr)
	list := b["data"].(map[string]any)["list"].([]any)
	assert.Len(t, list, 1)
}

// PROXY-10: Task hall filtered by station_id.
func TestProxy_10_Tasks_ByStation(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-TASK-002", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 30.0,
		"deadline":      deadline,
	}, env.userTok)

	// Filter by correct station.
	rr := env.do(t, http.MethodGet, "/api/v1/proxy/tasks?station_id="+itoa(env.stationID)+"&page=1&page_size=10", nil, env.runnerTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	list := b["data"].(map[string]any)["list"].([]any)
	assert.Len(t, list, 1)

	// Filter by wrong station → empty.
	rr = env.do(t, http.MethodGet, "/api/v1/proxy/tasks?station_id=99999&page=1&page_size=10", nil, env.runnerTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = proxyBodyMap(t, rr)
	list = b["data"].(map[string]any)["list"].([]any)
	assert.Len(t, list, 0)
}

// PROXY-11: Publish with invalid deadline → 400.
func TestProxy_11_Publish_InvalidDeadline(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-DL-001", "13800138800")

	// Deadline less than 30 min in future.
	deadline := time.Now().Add(10 * time.Minute).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, env.userTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	b := proxyBodyMap(t, rr)
	assert.Equal(t, float64(10405), b["code"])
}

// PROXY-12: Duplicate publish on same parcel → 409.
func TestProxy_12_Publish_Duplicate(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-DUP-001", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	body := map[string]any{"parcel_id": parcelID, "reward_amount": 10.0, "deadline": deadline}
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", body, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)

	rr = env.do(t, http.MethodPost, "/api/v1/proxy/publish", body, env.userTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b := proxyBodyMap(t, rr)
	assert.Equal(t, float64(10403), b["code"])
}

// PROXY-13: Publisher cannot accept own order → 409.
func TestProxy_13_Accept_SelfAccept(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)

	runnerUser := SeedUserWithStatus(t, env.env.DB, "13800138888", models.UserTypeRunner, models.RunnerStatusApproved)
	pubRunnerTok := signParcelToken(t, env.env, runnerUser.ID, env.stationID, "user")

	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-SELF-001", "13800138888")
	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, pubRunnerTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	rr = env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, pubRunnerTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b = proxyBodyMap(t, rr)
	assert.Equal(t, float64(10414), b["code"])
}

// PROXY-14: Cancel order that is already completed → 409.
func TestProxy_14_Cancel_Completed(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-CCOMP-001", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 5.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.runnerTok)
	env.do(t, http.MethodPost, "/api/v1/proxy/request-delivery/"+itoa(orderID), map[string]any{
		"delivery_photos": []string{"https://cdn/p.jpg"},
	}, env.runnerTok)
	env.do(t, http.MethodPost, "/api/v1/proxy/confirm-delivery/"+itoa(orderID), map[string]any{
		"accepted": true,
	}, env.userTok)

	// Try cancelling completed order.
	rr = env.do(t, http.MethodPost, "/api/v1/proxy/cancel", map[string]any{
		"order_id":      orderID,
		"cancel_reason": "late",
	}, env.userTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b = proxyBodyMap(t, rr)
	assert.Equal(t, float64(10443), b["code"])
}

// PROXY-15: Non-taker cannot request delivery → 409.
func TestProxy_15_Delivery_NotTaker(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-NTAKER-001", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.runnerTok)

	// Publisher (not runner) tries to request delivery.
	rr = env.do(t, http.MethodPost, "/api/v1/proxy/request-delivery/"+itoa(orderID), map[string]any{
		"delivery_photos": []string{"https://cdn/p.jpg"},
	}, env.userTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b = proxyBodyMap(t, rr)
	assert.Equal(t, float64(10421), b["code"])
}

// PROXY-16: Confirm delivery before delivery → 409.
func TestProxy_16_Confirm_BeforeDelivery(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-CONF-001", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.runnerTok)

	// Confirm before delivery is requested.
	rr = env.do(t, http.MethodPost, "/api/v1/proxy/confirm-delivery/"+itoa(orderID), map[string]any{
		"accepted": true,
	}, env.userTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b = proxyBodyMap(t, rr)
	assert.Equal(t, float64(10432), b["code"])
}

// PROXY-17: Reject delivery without reason → 400.
func TestProxy_17_Confirm_RejectNoReason(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-REJ-001", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.runnerTok)
	env.do(t, http.MethodPost, "/api/v1/proxy/request-delivery/"+itoa(orderID), map[string]any{
		"delivery_photos": []string{"https://cdn/p.jpg"},
	}, env.runnerTok)

	rr = env.do(t, http.MethodPost, "/api/v1/proxy/confirm-delivery/"+itoa(orderID), map[string]any{
		"accepted": false,
		"reason":   "",
	}, env.userTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	b = proxyBodyMap(t, rr)
	assert.Equal(t, float64(10433), b["code"])
}

// PROXY-18: My-orders filtered by status.
func TestProxy_18_MyOrders_ByStatus(t *testing.T) {
	env := setupProxyEngine(t)
	adminTok := adminScanToken(t, env)
	parcelID, _ := scanProxyParcelAdmin(t, env, adminTok, "SF-MYSTAT-001", "13800138800")

	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr := env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 10.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)

	// Filter by status=1 (pending).
	rr = env.do(t, http.MethodGet, "/api/v1/proxy/my-orders?status=1&page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := proxyBodyMap(t, rr)
	list := b["data"].(map[string]any)["list"].([]any)
	assert.Len(t, list, 1)

	// Filter by status=2 (delivering) → empty.
	rr = env.do(t, http.MethodGet, "/api/v1/proxy/my-orders?status=2&page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = proxyBodyMap(t, rr)
	list = b["data"].(map[string]any)["list"].([]any)
	assert.Len(t, list, 0)
}
