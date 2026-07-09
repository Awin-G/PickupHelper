package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRateLimit_Allow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RateLimit(10, 10))
	engine.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
}

func TestRateLimit_DisabledWhenQPSZero(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := 0
	engine := gin.New()
	engine.Use(RateLimit(0, 0))
	engine.GET("/x", func(c *gin.Context) {
		called++
		c.Status(http.StatusOK)
	})

	for i := 0; i < 50; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		rr := httptest.NewRecorder()
		engine.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, rr.Code)
		}
	}
	if called != 50 {
		t.Errorf("handler called %d times, want 50", called)
	}
}

func TestRateLimit_Deny(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RateLimit(1, 1))
	engine.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	// First request consumes the token (200). Subsequent should be 429.
	allowed, rejected := 0, 0
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		rr := httptest.NewRecorder()
		engine.ServeHTTP(rr, req)
		switch rr.Code {
		case http.StatusOK:
			allowed++
		case http.StatusTooManyRequests:
			rejected++
		default:
			t.Fatalf("request %d: unexpected status %d", i, rr.Code)
		}
	}
	if allowed != 1 {
		t.Errorf("allowed = %d, want 1 (burst)", allowed)
	}
	if rejected != 4 {
		t.Errorf("rejected = %d, want 4", rejected)
	}
}

func TestRateLimit_RejectedBodyHasCode10006(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID(), RateLimit(1, 1))
	engine.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	// First request consumes the token.
	req1 := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr1 := httptest.NewRecorder()
	engine.ServeHTTP(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want 200", rr1.Code)
	}

	// Second request should be rejected.
	req2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr2 := httptest.NewRecorder()
	engine.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rr2.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr2.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rr2.Body.String())
	}
	if body["code"] != float64(10006) {
		t.Errorf("code = %v, want 10006", body["code"])
	}
	if body["trace_id"] == "" || body["trace_id"] == "-" {
		t.Errorf("trace_id should be populated, got %v", body["trace_id"])
	}
}
