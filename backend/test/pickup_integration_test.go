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
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/router"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pickupTestEnv struct {
	env       *TestEnv
	engine    *gin.Engine
	stationID int64
	adminID   int64
	adminTok  string
	userID    int64
	userTok   string
}

func setupPickupEngine(t *testing.T) *pickupTestEnv {
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

	parcelSvc := service.NewParcelService(parcelRepo, shelfRepo, userRepo, env.DB)
	pickupSvc := service.NewPickupService(parcelRepo, pickupRepo, shelfRepo, userRepo, env.DB)

	stationID := SeedStation(t, env.DB)
	SeedShelf(t, env.DB, stationID, "A-01", 100)

	admin := SeedAdmin(t, env.DB, "pickup-admin")
	userID := SeedUser(t, env.DB, "13800138900")

	adminTok := signParcelToken(t, env, admin.ID, stationID, "admin")
	userTok := signParcelToken(t, env, userID, 0, "user")

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	parcelH := handler.NewParcelHandler(parcelSvc)
	pickupH := handler.NewPickupHandler(pickupSvc)
	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Parcel: parcelH,
		Pickup: pickupH,
	})

	return &pickupTestEnv{
		env:       env,
		engine:    engine,
		stationID: stationID,
		adminID:   admin.ID,
		adminTok:  adminTok,
		userID:    userID,
		userTok:   userTok,
	}
}

func (p *pickupTestEnv) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
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

func pickupBodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

// Helper: scan in a parcel and return pickup code.
func scanParcel(t *testing.T, env *pickupTestEnv, trackingNo, phone string) string {
	t.Helper()
	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     trackingNo,
		"courier_company": "顺丰速运",
		"receiver_phone":  phone,
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "scan-in body=%s", rr.Body.String())
	b := pickupBodyMap(t, rr)
	return b["data"].(map[string]any)["pickup_code"].(string)
}

// PICKUP-01: Admin verifies a parcel → status changes to 已取.
func TestPickup_01_Verify_Admin_Success(t *testing.T) {
	env := setupPickupEngine(t)
	code := scanParcel(t, env, "SF-VERIFY-001", "13900139000")

	rr := env.do(t, http.MethodPost, "/api/v1/pickup/verify", map[string]any{
		"pickup_code":         code,
		"verification_method": 1,
		"station_id":          env.stationID,
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "verify body=%s", rr.Body.String())
	b := pickupBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.NotEmpty(t, data["parcel_id"])
	assert.NotEmpty(t, data["pickup_time"])
	assert.Equal(t, float64(models.OpTypeAdmin), data["operator_type"])
}

// PICKUP-02: Verify with invalid pickup code → 400.
func TestPickup_02_Verify_InvalidCode(t *testing.T) {
	env := setupPickupEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/pickup/verify", map[string]any{
		"pickup_code":         "999999",
		"verification_method": 1,
		"station_id":          env.stationID,
	}, env.adminTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	b := pickupBodyMap(t, rr)
	assert.Equal(t, float64(10301), b["code"])
}

// PICKUP-03: Verify already-picked parcel → 403.
func TestPickup_03_Verify_AlreadyPicked(t *testing.T) {
	env := setupPickupEngine(t)
	code := scanParcel(t, env, "SF-VERIFY-002", "13900139001")

	// First verify succeeds.
	env.do(t, http.MethodPost, "/api/v1/pickup/verify", map[string]any{
		"pickup_code":         code,
		"verification_method": 1,
		"station_id":          env.stationID,
	}, env.adminTok)

	// Second verify on same code fails.
	rr := env.do(t, http.MethodPost, "/api/v1/pickup/verify", map[string]any{
		"pickup_code":         code,
		"verification_method": 1,
		"station_id":          env.stationID,
	}, env.adminTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	b := pickupBodyMap(t, rr)
	assert.Equal(t, float64(10301), b["code"])
}

// PICKUP-04: Self-checkout as owner → success.
func TestPickup_04_SelfCheckout_Success(t *testing.T) {
	env := setupPickupEngine(t)
	code := scanParcel(t, env, "SF-SELF-001", "13800138900")

	rr := env.do(t, http.MethodPost, "/api/v1/pickup/self-checkout", map[string]any{
		"pickup_code": code,
		"station_id":  env.stationID,
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "self-checkout body=%s", rr.Body.String())
	b := pickupBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.NotEmpty(t, data["parcel_id"])
	assert.NotEmpty(t, data["pickup_time"])
}

// PICKUP-05: Self-checkout non-owner → 400.
func TestPickup_05_SelfCheckout_NotOwner(t *testing.T) {
	env := setupPickupEngine(t)
	code := scanParcel(t, env, "SF-SELF-002", "13900139990")

	rr := env.do(t, http.MethodPost, "/api/v1/pickup/self-checkout", map[string]any{
		"pickup_code": code,
		"station_id":  env.stationID,
	}, env.userTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	b := pickupBodyMap(t, rr)
	assert.Equal(t, float64(10311), b["code"])
}

// PICKUP-06: Scan-station bulk checkout → some succeed, some fail.
func TestPickup_06_ScanStation_Bulk(t *testing.T) {
	env := setupPickupEngine(t)
	code1 := scanParcel(t, env, "SF-BULK-001", "13800138900")
	code2 := scanParcel(t, env, "SF-BULK-002", "13800138900")
	codeNotMine := scanParcel(t, env, "SF-BULK-003", "13900139991")

	rr := env.do(t, http.MethodPost, "/api/v1/pickup/scan-station", map[string]any{
		"station_qr":   "station:" + itoa(env.stationID),
		"pickup_codes": []string{code1, code2, codeNotMine},
	}, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "scan-station body=%s", rr.Body.String())
	b := pickupBodyMap(t, rr)
	data := b["data"].(map[string]any)
	success := data["success"].([]any)
	failed := data["failed"].([]any)
	assert.Len(t, success, 2)
	assert.Len(t, failed, 1)
}

// PICKUP-07: Pickup logs list → returns paginated logs.
func TestPickup_07_Logs_List(t *testing.T) {
	env := setupPickupEngine(t)
	code := scanParcel(t, env, "SF-LOG-001", "13900139002")

	// Do a verify to create a log entry.
	env.do(t, http.MethodPost, "/api/v1/pickup/verify", map[string]any{
		"pickup_code":         code,
		"verification_method": 1,
		"station_id":          env.stationID,
	}, env.adminTok)

	rr := env.do(t, http.MethodGet, "/api/v1/pickup/logs?page=1&page_size=10", nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "logs body=%s", rr.Body.String())
	b := pickupBodyMap(t, rr)
	data := b["data"].(map[string]any)
	list := data["list"].([]any)
	assert.Len(t, list, 1)
	total := data["total"].(float64)
	assert.Equal(t, float64(1), total)

	item := list[0].(map[string]any)
	assert.NotEmpty(t, item["tracking_no"])
}

// PICKUP-08: Self-checkout unauthenticated → 401.
func TestPickup_08_SelfCheckout_Unauthenticated(t *testing.T) {
	env := setupPickupEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/pickup/self-checkout", map[string]any{
		"pickup_code": "123456",
		"station_id":  env.stationID,
	}, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
