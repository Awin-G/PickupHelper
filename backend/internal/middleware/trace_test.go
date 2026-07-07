package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
)

func TestTraceID_WhenHeaderMissing_GeneratesNew(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID())
	engine.GET("/x", func(c *gin.Context) {
		tid := log.TraceID(c.Request.Context())
		c.String(http.StatusOK, tid)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	body := rr.Body.String()
	if len(body) == 0 {
		t.Fatalf("trace id should be non-empty, got %q", body)
	}
	// Response header should carry the same value.
	if got := rr.Header().Get("X-Trace-Id"); got != body {
		t.Errorf("response header = %q, want %q", got, body)
	}
}

func TestTraceID_WhenHeaderPresent_PreservesIt(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID())
	engine.GET("/x", func(c *gin.Context) {
		c.String(http.StatusOK, log.TraceID(c.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Trace-Id", "client-supplied")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if got := rr.Body.String(); got != "client-supplied" {
		t.Errorf("trace id in context = %q, want client-supplied", got)
	}
	if got := rr.Header().Get("X-Trace-Id"); got != "client-supplied" {
		t.Errorf("response header = %q, want client-supplied", got)
	}
}

func TestTraceID_GeneratedLength(t *testing.T) {
	// Sanity: generated IDs should be 12 hex chars (log.NewTraceID contract).
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID())
	engine.GET("/x", func(c *gin.Context) {
		c.String(http.StatusOK, log.TraceID(c.Request.Context()))
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if got := rr.Body.String(); len(got) != 12 {
		t.Errorf("generated trace id length = %d, want 12 (got %q)", len(got), got)
	}
}
