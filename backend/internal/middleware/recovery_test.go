package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecovery_PanicReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID(), Recovery())
	engine.GET("/boom", func(c *gin.Context) {
		panic("something exploded")
	})

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v (body=%s)", err, rr.Body.String())
	}
	if body["code"] != float64(10009) {
		t.Errorf("code = %v, want 10009", body["code"])
	}
	if body["msg"] != "服务端内部错误" {
		t.Errorf("msg = %v, want '服务端内部错误'", body["msg"])
	}
	if _, has := body["trace_id"]; !has {
		t.Errorf("trace_id should be present in panic response")
	}
}

func TestRecovery_NoPanic_PassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Recovery())
	called := false
	engine.GET("/ok", func(c *gin.Context) {
		called = true
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if !called {
		t.Errorf("handler should have been called")
	}
	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}
}
