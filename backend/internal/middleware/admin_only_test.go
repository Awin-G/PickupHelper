package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newAdminOnlyEngine(headers map[string]string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID())
	engine.Use(func(c *gin.Context) {
		for k, v := range headers {
			c.Request.Header.Set(k, v)
		}
		c.Next()
	})
	engine.Use(AdminOnly())
	engine.GET("/x", func(c *gin.Context) { c.Status(200) })
	return engine
}

func TestAdminOnly_AdminRole(t *testing.T) {
	engine := newAdminOnlyEngine(map[string]string{
		"X-User-Id": "1",
		"X-Role":    "admin",
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (admin should pass)", rr.Code)
	}
}

func TestAdminOnly_UserRole(t *testing.T) {
	engine := newAdminOnlyEngine(map[string]string{
		"X-User-Id": "1",
		"X-Role":    "user",
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["code"] != float64(10003) {
		t.Errorf("code = %v, want 10003 (ErrForbidden)", body["code"])
	}
}

func TestAdminOnly_NoRole(t *testing.T) {
	engine := newAdminOnlyEngine(map[string]string{})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}
