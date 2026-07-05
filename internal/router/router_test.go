package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pickup-helper/internal/config"
	"pickup-helper/internal/handler"

	"github.com/gin-gonic/gin"
)

func mustConfig() *config.Config {
	return &config.Config{
		CORS:      config.CORSConfig{AllowedOrigins: []string{"*"}},
		RateLimit: config.RateLimitConfig{QPS: 0, Burst: 0}, // disabled
		JWT:       config.JWTConfig{AccessSecret: "test-secret"},
	}
}

// TestRegister_MiddlewareChain verifies the wired-in middleware chain
// produces a correct /health response with trace_id and unified envelope.
func TestRegister_MiddlewareChain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	hh := handler.NewHealthHandler(nil, nil)
	Register(engine, mustConfig(), hh)

	// /health should succeed with unified envelope + trace_id.
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rr.Body.String())
	}
	if body["code"] != float64(0) {
		t.Errorf("code = %v, want 0", body["code"])
	}
	if body["msg"] != "success" {
		t.Errorf("msg = %v, want success", body["msg"])
	}
	if body["trace_id"] == "" || body["trace_id"] == "-" {
		t.Errorf("trace_id should be populated, got %v", body["trace_id"])
	}
	if got := rr.Header().Get("X-Trace-Id"); got == "" {
		t.Errorf("X-Trace-Id response header should be set")
	} else if got != body["trace_id"] {
		t.Errorf("X-Trace-Id header (%q) != body.trace_id (%v)", got, body["trace_id"])
	}
}

// TestRegister_CORSHeaders verifies the CORS middleware is wired and runs
// on every response.
func TestRegister_CORSHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	hh := handler.NewHealthHandler(nil, nil)
	Register(engine, mustConfig(), hh)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Allow-Origin = %q, want *", got)
	}
}

// TestRegister_PreflightShortCircuits verifies OPTIONS requests are
// handled by the CORS middleware and never reach the handler.
func TestRegister_PreflightShortCircuits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	hh := handler.NewHealthHandler(nil, nil)
	Register(engine, mustConfig(), hh)

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (CORS preflight)", rr.Code)
	}
}

// TestRegister_APIV1RequiresJWT verifies the /api/v1 group is mounted with
// JWTAuth — any unauthenticated request should return 401 + code=10002.
// This satisfies plan 01-03's verification step:
// "curl http://localhost:8080/api/v1/anything → 401 + code=10002".
func TestRegister_APIV1RequiresJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	hh := handler.NewHealthHandler(nil, nil)
	Register(engine, mustConfig(), hh)

	// Add a dummy route inside /api/v1 so gin doesn't return 404.
	NewAPIV1Group(engine, mustConfig()).GET("/ping", func(c *gin.Context) {
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 (JWT middleware should reject)", rr.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["code"] != float64(10002) {
		t.Errorf("code = %v, want 10002", body["code"])
	}
}

// TestRegister_APIV1PreflightBypassesJWT verifies OPTIONS preflight
// requests under /api/v1 return 204 (handled by CORS before JWT).
func TestRegister_APIV1PreflightBypassesJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	hh := handler.NewHealthHandler(nil, nil)
	Register(engine, mustConfig(), hh)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/anything", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (CORS preflight should bypass JWT)", rr.Code)
	}
}
