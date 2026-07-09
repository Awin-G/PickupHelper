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

type e2eTestEnv struct {
	env     *TestEnv
	engine  *gin.Engine
	stationID int64
	adminTok  string
	userID    int64
	userTok   string
	runnerID  int64
	runnerTok string
}

func setupE2EEngine(t *testing.T) *e2eTestEnv {
	t.Helper()
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	_ = middleware.Validator()
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	userRepo := repository.NewUserRepo()
	adminRepo := repository.NewAdminRepo()
	runnerRepo := repository.NewRunnerAppRepo()
	smsCache := repository.NewSMSCodeCache(env.Rdb)
	parcelRepo := repository.NewParcelRepo()
	shelfRepo := repository.NewShelfRepo()
	pickupRepo := repository.NewPickupLogRepo()
	proxyRepo := repository.NewProxyOrderRepo()
	notifyRepo := repository.NewNotifyRepo()

	sms := service.NewSMSProvider("test", nil)
	authSvc := service.NewAuthService(userRepo, adminRepo, smsCache, sms, env.Cfg, env.DB)
	userSvc := service.NewUserService(userRepo, runnerRepo, env.DB)
	parcelSvc := service.NewParcelService(parcelRepo, shelfRepo, userRepo, env.DB)
	pickupSvc := service.NewPickupService(parcelRepo, pickupRepo, shelfRepo, userRepo, env.DB)
	proxySvc := service.NewProxyService(proxyRepo, parcelRepo, userRepo, env.DB)
	shelfSvc := service.NewShelfService(shelfRepo, env.DB)
	notifySvc := service.NewNotifyService(notifyRepo, env.DB)
	statsSvc := service.NewStatsService(env.DB)

	stationID := SeedStation(t, env.DB)
	SeedShelf(t, env.DB, stationID, "A-01", 100)

	admin := SeedAdmin(t, env.DB, "e2e-admin")
	userID := SeedUser(t, env.DB, "13800137000")
	runner := SeedUserWithStatus(t, env.DB, "13900137001", models.UserTypeRunner, models.RunnerStatusApproved)
	runnerID := runner.ID

	adminTok := signParcelToken(t, env, admin.ID, stationID, "admin")
	userTok := signParcelToken(t, env, userID, 0, "user")
	runnerTok := signParcelToken(t, env, runnerID, 0, "user")

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	parcelH := handler.NewParcelHandler(parcelSvc)
	pickupH := handler.NewPickupHandler(pickupSvc)
	proxyH := handler.NewProxyHandler(proxySvc)
	shelfH := handler.NewShelfHandler(shelfSvc)
	notifyH := handler.NewNotifyHandler(notifySvc)
	statsH := handler.NewStatsHandler(statsSvc)

	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Auth:   authH,
		User:   userH,
		Parcel: parcelH,
		Pickup: pickupH,
		Proxy:  proxyH,
		Shelf:  shelfH,
		Notify: notifyH,
		Stats:  statsH,
	})

	return &e2eTestEnv{
		env: env, engine: engine, stationID: stationID,
		adminTok: adminTok, userID: userID, userTok: userTok,
		runnerID: runnerID, runnerTok: runnerTok,
	}
}

func (e *e2eTestEnv) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
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
	e.engine.ServeHTTP(rr, req)
	return rr
}

func e2eBodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

// E2E-01: Full lifecycle — Login → Parcel intake → Pickup verify → Stats.
func TestE2E_01_Login_Intake_Verify_Stats(t *testing.T) {
	env := setupE2EEngine(t)

	// Step 1: Admin login via JWT.
	rr := env.do(t, http.MethodPost, "/api/v1/admin/auth/login", map[string]string{
		"username": "e2e-admin", "password": "test-password-123",
	}, "")
	require.Equal(t, http.StatusOK, rr.Code, "admin login failed: %s", rr.Body.String())
	b := e2eBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, "admin", data["role"])
	assert.NotEmpty(t, data["access_token"])

	// Step 2: Admin scans in a parcel for the user.
	rr = env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "E2E-TEST-001",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13800137000",
		"receiver_name":   "E2E用户",
		"weight":          1.5,
		"is_fragile":      false,
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "scan-in failed: %s", rr.Body.String())
	b = e2eBodyMap(t, rr)
	pData := b["data"].(map[string]any)
	parcelID := int64(pData["parcel_id"].(float64))
	pickupCode := pData["pickup_code"].(string)
	assert.NotEmpty(t, pickupCode)
	assert.NotEmpty(t, pData["shelf_code"])

	// Step 3: User views their parcels → should find 1 pending.
	rr = env.do(t, http.MethodGet, "/api/v1/parcels/my?page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	list := b["data"].(map[string]any)["list"].([]any)
	assert.Len(t, list, 1)
	assert.Equal(t, pickupCode, list[0].(map[string]any)["pickup_code"])

	// Step 4: Admin verifies the parcel (pickup).
	rr = env.do(t, http.MethodPost, "/api/v1/pickup/verify", map[string]any{
		"pickup_code":         pickupCode,
		"verification_method": 1,
		"station_id":          env.stationID,
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "verify failed: %s", rr.Body.String())
	b = e2eBodyMap(t, rr)
	vData := b["data"].(map[string]any)
	assert.Equal(t, float64(models.OpTypeAdmin), vData["operator_type"])

	// Step 5: Verify parcel status is now "已取" in DB.
	var status int8
	require.NoError(t, env.env.DB.Get(&status, "SELECT status FROM parcels WHERE id=?", parcelID))
	assert.Equal(t, int8(models.ParcelStatusPickedUp), status)

	// Step 6: Dashboard shows 1 outbound today.
	rr = env.do(t, http.MethodGet, "/api/v1/stats/dashboard?station_id="+itoa(env.stationID), nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	dashData := b["data"].(map[string]any)
	assert.Equal(t, float64(1), dashData["today_outbound"])
	assert.Equal(t, float64(0), dashData["pending_count"])

	// Step 7: Pickup logs show 1 entry.
	rr = env.do(t, http.MethodGet, "/api/v1/pickup/logs?page=1&page_size=10", nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	logList := b["data"].(map[string]any)["list"].([]any)
	assert.Len(t, logList, 1)
}

// E2E-02: Full proxy lifecycle — Publish → Accept → Delivery → Confirm.
func TestE2E_02_Proxy_Lifecycle(t *testing.T) {
	env := setupE2EEngine(t)

	// 1. Admin scans in a parcel for the user.
	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "E2E-PROXY-001",
		"courier_company": "中通快递",
		"receiver_phone":  "13800137000",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := e2eBodyMap(t, rr)
	parcelID := int64(b["data"].(map[string]any)["parcel_id"].(float64))
	_ = parcelID

	// 2. User publishes a proxy order.
	deadline := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	rr = env.do(t, http.MethodPost, "/api/v1/proxy/publish", map[string]any{
		"parcel_id":     parcelID,
		"reward_amount": 20.0,
		"deadline":      deadline,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "publish failed: %s", rr.Body.String())
	b = e2eBodyMap(t, rr)
	orderID := int64(b["data"].(map[string]any)["order_id"].(float64))

	// 3. Runner views task hall → should see 1 task.
	rr = env.do(t, http.MethodGet, "/api/v1/proxy/tasks?page=1&page_size=10", nil, env.runnerTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	taskList := b["data"].(map[string]any)["list"].([]any)
	assert.Len(t, taskList, 1)

	// 4. Runner accepts the order.
	rr = env.do(t, http.MethodPost, "/api/v1/proxy/accept/"+itoa(orderID), nil, env.runnerTok)
	require.Equal(t, http.StatusOK, rr.Code, "accept failed: %s", rr.Body.String())
	b = e2eBodyMap(t, rr)
	aData := b["data"].(map[string]any)
	assert.Equal(t, float64(models.ProxyStatusDelivering), aData["status"])
	assert.NotEmpty(t, aData["temp_pickup_code"])

	// 5. Runner requests delivery.
	rr = env.do(t, http.MethodPost, "/api/v1/proxy/request-delivery/"+itoa(orderID), map[string]any{
		"delivery_photos": []string{"https://cdn/e2e-delivery.jpg"},
	}, env.runnerTok)
	require.Equal(t, http.StatusOK, rr.Code, "delivery failed: %s", rr.Body.String())
	b = e2eBodyMap(t, rr)
	dData := b["data"].(map[string]any)
	assert.Equal(t, float64(models.ProxyStatusConfirm), dData["status"])

	// 6. User confirms receipt.
	rr = env.do(t, http.MethodPost, "/api/v1/proxy/confirm-delivery/"+itoa(orderID), map[string]any{
		"accepted": true,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "confirm failed: %s", rr.Body.String())
	b = e2eBodyMap(t, rr)
	cData := b["data"].(map[string]any)
	assert.Equal(t, float64(models.ProxyStatusDone), cData["status"])

	// 7. Verify order status in DB.
	var orderStatus int8
	require.NoError(t, env.env.DB.Get(&orderStatus, "SELECT status FROM proxy_orders WHERE id=?", orderID))
	assert.Equal(t, int8(models.ProxyStatusDone), orderStatus)
}

// E2E-03: Shelf management and occupancy.
func TestE2E_03_Shelf_Management(t *testing.T) {
	env := setupE2EEngine(t)

	// 1. Create shelf.
	rr := env.do(t, http.MethodPost, "/api/v1/shelves", map[string]any{
		"station_id":   env.stationID,
		"shelf_code":   "E2E-Z01",
		"row_num":      3,
		"col_num":      5,
		"max_capacity": 30,
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := e2eBodyMap(t, rr)
	data := b["data"].(map[string]any)
	shelfID := int64(data["id"].(float64))
	assert.Equal(t, "E2E-Z01", data["shelf_code"])

	// 2. Update shelf capacity.
	rr = env.do(t, http.MethodPut, "/api/v1/shelves/"+itoa(shelfID), map[string]any{
		"max_capacity": 50,
		"remark":       "e2e updated",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	data = b["data"].(map[string]any)
	assert.Equal(t, float64(50), data["max_capacity"])

	// 3. List shelves.
	rr = env.do(t, http.MethodGet, "/api/v1/shelves?page=1&page_size=10", nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	sList := b["data"].(map[string]any)["list"].([]any)
	assert.True(t, len(sList) >= 1)

	// 4. Occupancy heatmap.
	rr = env.do(t, http.MethodGet, "/api/v1/shelves/occupancy?station_id="+itoa(env.stationID), nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	occData := b["data"].(map[string]any)
	assert.NotEmpty(t, occData["shelves"])
	assert.GreaterOrEqual(t, occData["total_max"].(float64), float64(50))
}

// E2E-04: Notification flow.
func TestE2E_04_Notification_Read(t *testing.T) {
	env := setupE2EEngine(t)

	// Seed notifications directly.
	for i := 1; i <= 5; i++ {
		_, err := env.env.DB.Exec(
			`INSERT INTO notifications (user_id, title, content, type, is_read, send_status, channel)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			env.userID, "通知"+itoa(int64(i)), "内容"+itoa(int64(i)),
			models.NotifyTypeSystem, 0, models.SendStatusSent, models.ChannelSMS)
		require.NoError(t, err)
	}

	// List → 5 unread.
	rr := env.do(t, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := e2eBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(5), data["total"])
	assert.Equal(t, float64(5), data["unread"])

	// Mark all read.
	env.do(t, http.MethodPut, "/api/v1/notifications/read", nil, env.userTok)

	// List again → 0 unread.
	rr = env.do(t, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	data = b["data"].(map[string]any)
	assert.Equal(t, float64(0), data["unread"])
}

// E2E-05: Parcel status transitions.
func TestE2E_05_Parcel_StatusTransitions(t *testing.T) {
	env := setupE2EEngine(t)

	// Scan in parcel.
	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "E2E-STATUS-001",
		"courier_company": "圆通速递",
		"receiver_phone":  "13800137000",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := e2eBodyMap(t, rr)
	parcelID := int64(b["data"].(map[string]any)["parcel_id"].(float64))

	// Status: 1(待取) → 3(滞留).
	rr = env.do(t, http.MethodPut, "/api/v1/parcels/"+itoa(parcelID)+"/status", map[string]any{
		"status": 3, "reason": "超过72小时",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	assert.Equal(t, float64(3), b["data"].(map[string]any)["status"])

	// Status: 3(滞留) → 4(已退件).
	rr = env.do(t, http.MethodPut, "/api/v1/parcels/"+itoa(parcelID)+"/status", map[string]any{
		"status": 4, "reason": "超期退件",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	assert.Equal(t, float64(4), b["data"].(map[string]any)["status"])

	// Status: 4(已退件) → 5(异常).
	rr = env.do(t, http.MethodPut, "/api/v1/parcels/"+itoa(parcelID)+"/status", map[string]any{
		"status": 5, "reason": "退件异常",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = e2eBodyMap(t, rr)
	assert.Equal(t, float64(5), b["data"].(map[string]any)["status"])

	// Verify final status in DB.
	var status int8
	require.NoError(t, env.env.DB.Get(&status, "SELECT status FROM parcels WHERE id=?", parcelID))
	assert.Equal(t, int8(models.ParcelStatusAbnormal), status)
}
