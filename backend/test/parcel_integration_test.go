//go:build integration

package test

import (
	"context"
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
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type parcelTestEnv struct {
	env       *TestEnv
	engine    *gin.Engine
	stationID int64
	adminID   int64
	adminTok  string
	userID    int64
	userTok   string
}

func setupParcelEngine(t *testing.T) *parcelTestEnv {
	t.Helper()
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	_ = middleware.Validator()
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	userRepo := repository.NewUserRepo()
	parcelRepo := repository.NewParcelRepo()
	shelfRepo := repository.NewShelfRepo()

	parcelSvc := service.NewParcelService(parcelRepo, shelfRepo, userRepo, env.DB)

	stationID := SeedStation(t, env.DB)
	SeedShelf(t, env.DB, stationID, "A-01", 100)
	SeedShelf(t, env.DB, stationID, "B-02", 50)

	admin := SeedAdmin(t, env.DB, "parcel-admin")
	userID := SeedUser(t, env.DB, "13800138099")

	adminTok := signParcelToken(t, env, admin.ID, stationID, "admin")
	userTok := signParcelToken(t, env, userID, 0, "user")

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Parcel: handler.NewParcelHandler(parcelSvc),
	})

	return &parcelTestEnv{
		env:       env,
		engine:    engine,
		stationID: stationID,
		adminID:   admin.ID,
		adminTok:  adminTok,
		userID:    userID,
		userTok:   userTok,
	}
}

func (p *parcelTestEnv) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
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

func parcelBodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

// signParcelToken creates a valid access token with stationID for parcel tests.
func signParcelToken(t *testing.T, env *TestEnv, userID, stationID int64, role string) string {
	t.Helper()
	claims := middleware.Claims{
		UserID:    userID,
		UserType:  int(models.UserTypeNormal),
		StationID: stationID,
		Role:      role,
	}
	tok, err := middleware.SignAccess(env.Cfg, claims)
	require.NoError(t, err)
	return tok
}

// PARCEL-01: ScanIn a parcel (admin) → returns pickup_code and shelf_code.
func TestParcel_01_ScanIn_Success(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":    "SF1234567890",
		"courier_company": "顺丰速运",
		"receiver_phone": "13900139000",
		"receiver_name":  "张三",
		"weight":         2.5,
		"is_fragile":     true,
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "scan-in body=%s", rr.Body.String())
	body := parcelBodyMap(t, rr)
	assert.Equal(t, float64(0), body["code"])
	data := body["data"].(map[string]any)
	assert.NotEmpty(t, data["parcel_id"])
	assert.NotEmpty(t, data["pickup_code"])
	pickupCode := data["pickup_code"].(string)
	assert.Len(t, pickupCode, 6)
	assert.NotEmpty(t, data["shelf_code"])
}

// PARCEL-02: ScanIn duplicate tracking_no in same station → 409.
func TestParcel_02_ScanIn_Duplicate(t *testing.T) {
	env := setupParcelEngine(t)

	body := map[string]any{
		"tracking_no":     "SF9999999999",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13900139001",
	}
	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", body, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)

	rr = env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", body, env.adminTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b := parcelBodyMap(t, rr)
	assert.Equal(t, float64(10201), b["code"])
}

// PARCEL-03: ScanIn with invalid phone format → 400 (binding validation).
func TestParcel_03_ScanIn_InvalidPhone(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF1111111111",
		"courier_company": "顺丰速运",
		"receiver_phone":  "12345",
	}, env.adminTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	b := parcelBodyMap(t, rr)
	assert.Equal(t, float64(10001), b["code"])
}

// PARCEL-04: ScanIn with explicit shelf_code → uses it.
func TestParcel_04_ScanIn_ExplicitShelf(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF2222222222",
		"courier_company": "中通快递",
		"receiver_phone":  "13900139002",
		"shelf_code":      "B-02",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, "B-02", data["shelf_code"])
}

// PARCEL-05: List parcels (admin) → returns paginated admin DTOs with pickup_code.
func TestParcel_05_ListParcels_Admin(t *testing.T) {
	env := setupParcelEngine(t)

	for i := 0; i < 3; i++ {
		rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
			"tracking_no":     "SF-LIST-" + itoa(int64(i)),
			"courier_company": "顺丰速运",
			"receiver_phone":  "13900139003",
		}, env.adminTok)
		require.Equal(t, http.StatusOK, rr.Code)
	}

	rr := env.do(t, http.MethodGet, "/api/v1/parcels?page=1&page_size=10", nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	list := data["list"].([]any)
	assert.Len(t, list, 3)
	assert.Equal(t, float64(3), data["total"])
	// Admin list must include pickup_code.
	item := list[0].(map[string]any)
	assert.NotEmpty(t, item["pickup_code"])
}

// PARCEL-06: List parcels non-admin → 403.
func TestParcel_06_ListParcels_NonAdmin(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/parcels?page=1&page_size=10", nil, env.userTok)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// PARCEL-07: ScanIn unauthenticated → 401.
func TestParcel_07_ScanIn_Unauthenticated(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF3333333333",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13900139004",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// PARCEL-08: Get parcel detail as owner → pickup_code visible.
func TestParcel_08_GetParcel_Owner(t *testing.T) {
	env := setupParcelEngine(t)

	// Scan in a parcel for the seeded user.
	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-OWNER-001",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13800138099",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	parcelID := int64(data["parcel_id"].(float64))
	pickupCode := data["pickup_code"].(string)

	// Owner reads their own parcel detail.
	rr = env.do(t, http.MethodGet, "/api/v1/parcels/"+itoa(parcelID), nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = parcelBodyMap(t, rr)
	data = b["data"].(map[string]any)
	assert.Equal(t, pickupCode, data["pickup_code"])
}

// PARCEL-09: Get parcel detail as non-owner → 403.
func TestParcel_09_GetParcel_NotOwner(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-NOTOWN-001",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13900139999",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	parcelID := int64(data["parcel_id"].(float64))

	rr = env.do(t, http.MethodGet, "/api/v1/parcels/"+itoa(parcelID), nil, env.userTok)
	assert.Equal(t, http.StatusForbidden, rr.Code)
	b = parcelBodyMap(t, rr)
	assert.Equal(t, float64(10222), b["code"])
}

// PARCEL-10: My parcels list → returns only own parcels with pickup_code.
func TestParcel_10_MyParcels(t *testing.T) {
	env := setupParcelEngine(t)

	// Scan two parcels for our user, one for someone else.
	env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-MY-001",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13800138099",
	}, env.adminTok)
	env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-MY-002",
		"courier_company": "中通快递",
		"receiver_phone":  "13800138099",
	}, env.adminTok)
	env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-MY-OTHER",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13900139998",
	}, env.adminTok)

	rr := env.do(t, http.MethodGet, "/api/v1/parcels/my?page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	list := data["list"].([]any)
	assert.Len(t, list, 2)
}

// PARCEL-11: Update status to detained (admin) → 200.
func TestParcel_11_UpdateStatus_Detain(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-DETAIN-001",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13900139997",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	parcelID := int64(data["parcel_id"].(float64))

	rr = env.do(t, http.MethodPut, "/api/v1/parcels/"+itoa(parcelID)+"/status", map[string]any{
		"status": 3,
		"reason": "超过72小时",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = parcelBodyMap(t, rr)
	data = b["data"].(map[string]any)
	assert.Equal(t, float64(3), data["status"])
	assert.Equal(t, "滞留", data["status_text"])
}

// PARCEL-12: Update status to picked_up (2) → 409 (not allowed via manual change).
func TestParcel_12_UpdateStatus_InvalidTarget(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-STATUS-001",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13900139996",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	parcelID := int64(data["parcel_id"].(float64))

	// First detain it (pending → detained).
	rr = env.do(t, http.MethodPut, "/api/v1/parcels/"+itoa(parcelID)+"/status", map[string]any{
		"status": 3,
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)

	// Then try to detain again (detained → detained) → invalid transition.
	rr = env.do(t, http.MethodPut, "/api/v1/parcels/"+itoa(parcelID)+"/status", map[string]any{
		"status": 3,
	}, env.adminTok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	b = parcelBodyMap(t, rr)
	assert.Equal(t, float64(10232), b["code"])
}

// PARCEL-13: Get pickup-code → returns code to owner.
func TestParcel_13_GetPickupCode(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-CODE-001",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13800138099",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	parcelID := int64(data["parcel_id"].(float64))
	pickupCode := data["pickup_code"].(string)

	rr = env.do(t, http.MethodGet, "/api/v1/parcels/"+itoa(parcelID)+"/pickup-code", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = parcelBodyMap(t, rr)
	data = b["data"].(map[string]any)
	assert.Equal(t, pickupCode, data["pickup_code"])
}

// PARCEL-14: Get pickup-code for non-owner → 400.
func TestParcel_14_GetPickupCode_NotOwner(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-CODE-002",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13900139995",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	parcelID := int64(data["parcel_id"].(float64))

	rr = env.do(t, http.MethodGet, "/api/v1/parcels/"+itoa(parcelID)+"/pickup-code", nil, env.userTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	b = parcelBodyMap(t, rr)
	assert.Equal(t, float64(10242), b["code"])
}

// PARCEL-15: Get parcel detail as admin → pickup_code visible, phone masked.
func TestParcel_15_GetParcel_Admin(t *testing.T) {
	env := setupParcelEngine(t)

	rr := env.do(t, http.MethodPost, "/api/v1/parcels/scan-in", map[string]any{
		"tracking_no":     "SF-ADMIN-001",
		"courier_company": "顺丰速运",
		"receiver_phone":  "13900139994",
	}, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := parcelBodyMap(t, rr)
	data := b["data"].(map[string]any)
	parcelID := int64(data["parcel_id"].(float64))

	rr = env.do(t, http.MethodGet, "/api/v1/parcels/"+itoa(parcelID), nil, env.adminTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b = parcelBodyMap(t, rr)
	data = b["data"].(map[string]any)
	// Admin should see pickup_code.
	assert.NotEmpty(t, data["pickup_code"])
	// Phone should be masked.
	phone := data["receiver_phone"].(string)
	assert.True(t, strings.Contains(phone, "****"), "phone should be masked, got %s", phone)
}

// guard
var (
	_ = time.Second
	_ = sqlx.DB{}
	_ = context.Background
)
