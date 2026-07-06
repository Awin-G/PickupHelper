//go:build integration

package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup-helper/internal/handler"
	"pickup-helper/internal/log"
	"pickup-helper/internal/router"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEngine(env *TestEnv) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h := handler.NewHealthHandler(env.DB, env.Rdb)
	router.Register(engine, env.Cfg, h)
	return engine
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	return body
}

func TestHealth_Live_Integration(t *testing.T) {
	env := SetupTestEnv(t)
	engine := setupEngine(env)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req = req.WithContext(log.WithTraceID(req.Context(), "int-live"))
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeBody(t, rr)
	assert.Equal(t, 0.0, body["code"])
	assert.Equal(t, "int-live", body["trace_id"])
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "up", data["status"])
}

func TestHealth_Ready_OK_Integration(t *testing.T) {
	env := SetupTestEnv(t)
	engine := setupEngine(env)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeBody(t, rr)
	assert.Equal(t, 0.0, body["code"])
}

func TestHealth_Ready_DBDown_Integration(t *testing.T) {
	env := SetupTestEnv(t)
	engine := setupEngine(env)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	require.NoError(t, env.MySQLContainerStop(ctx), "stop mysql container")

	// Give the container a moment to actually stop.
	time.Sleep(2 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	body := decodeBody(t, rr)
	assert.Equal(t, float64(10009), body["code"])
	assert.Equal(t, "mysql unavailable", body["msg"])
}

func TestHealth_Ready_RedisDown_Integration(t *testing.T) {
	env := SetupTestEnv(t)
	engine := setupEngine(env)

	// Stop MySQL first so it doesn't mask the redis failure (handler checks
	// DB before Redis). Then we restart a fresh env… no — instead, just
	// verify the redis-down case using a stand-alone handler with a working
	// DB and a stopped redis.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	require.NoError(t, env.RedisContainerStop(ctx), "stop redis container")

	time.Sleep(2 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	body := decodeBody(t, rr)
	assert.Equal(t, float64(10009), body["code"])
	assert.Equal(t, "redis unavailable", body["msg"])
}

func TestMigrations_AllTables_Integration(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()

	var tableNames []string
	err := env.DB.SelectContext(ctx, &tableNames,
		`SELECT table_name FROM information_schema.tables
		 WHERE table_schema = 'pickup_helper_test'
		 ORDER BY table_name`)
	require.NoError(t, err)

	expected := []string{
		"admins", "courier_companies", "goose_db_version", "notifications",
		"operation_logs", "parcels", "pickup_logs", "proxy_orders",
		"runner_applications", "shelf_layout", "stations", "users",
	}
	assert.Equal(t, expected, tableNames)

	// parcels should have 20 columns per DDL (id, station_id, tracking_no,
	// courier_company, shelf_code, pickup_code, receiver_phone,
	// receiver_user_id, receiver_name, weight, is_fragile, remarks, status,
	// storage_time, pickup_time, return_time, last_notify_time, notify_count,
	// operator_id, updated_at).
	var cols []string
	err = env.DB.SelectContext(ctx, &cols,
		`SELECT column_name FROM information_schema.columns
		 WHERE table_schema = 'pickup_helper_test' AND table_name = 'parcels'
		 ORDER BY ordinal_position`)
	require.NoError(t, err)
	assert.Len(t, cols, 20, "parcels column count")
}
