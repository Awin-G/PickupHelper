package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"pickup-helper/internal/log"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePinger implements Pinger for testing.
type fakePinger struct{ err error }

func (f fakePinger) PingContext(ctx context.Context) error { return f.err }

// fakeRedisPinger implements RedisPinger for testing.
type fakeRedisPinger struct{ err error }

func (f fakeRedisPinger) Ping(ctx context.Context) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if f.err != nil {
		cmd.SetErr(f.err)
	}
	return cmd
}

func setupRouter(h *HealthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	h.Register(engine)
	return engine
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	return body
}

func TestLive_OK(t *testing.T) {
	h := &HealthHandler{}
	engine := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req = req.WithContext(log.WithTraceID(req.Context(), "t-live"))
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeBody(t, rr)
	assert.Equal(t, 0.0, body["code"])
	assert.Equal(t, "t-live", body["trace_id"])
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "up", data["status"])
}

func TestReady_OK(t *testing.T) {
	h := &HealthHandler{db: fakePinger{}, rdb: fakeRedisPinger{}}
	engine := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeBody(t, rr)
	assert.Equal(t, 0.0, body["code"])
}

func TestReady_DBDown(t *testing.T) {
	h := &HealthHandler{
		db:  fakePinger{err: errors.New("conn refused")},
		rdb: fakeRedisPinger{},
	}
	engine := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	body := decodeBody(t, rr)
	assert.Equal(t, float64(10009), body["code"])
	assert.Equal(t, "mysql unavailable", body["msg"])
}

func TestReady_RedisDown(t *testing.T) {
	h := &HealthHandler{
		db:  fakePinger{},
		rdb: fakeRedisPinger{err: errors.New("conn refused")},
	}
	engine := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
	body := decodeBody(t, rr)
	assert.Equal(t, float64(10009), body["code"])
	assert.Equal(t, "redis unavailable", body["msg"])
}
