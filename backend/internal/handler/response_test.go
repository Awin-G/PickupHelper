package handler

import (
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/log"

	"github.com/gin-gonic/gin"
)

func newTestEngineWithTrace(traceID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(log.WithTraceID(c.Request.Context(), traceID))
		c.Next()
	})
	return engine
}

func TestSuccess(t *testing.T) {
	engine := newTestEngineWithTrace("abc")
	engine.GET("/x", func(c *gin.Context) {
		Success(c, gin.H{"foo": "bar"})
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := decodeBody(t, rr)
	if body["code"] != float64(0) {
		t.Errorf("code = %v, want 0", body["code"])
	}
	if body["msg"] != "success" {
		t.Errorf("msg = %v, want success", body["msg"])
	}
	if body["trace_id"] != "abc" {
		t.Errorf("trace_id = %v, want abc", body["trace_id"])
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data should be a map, got %T", body["data"])
	}
	if data["foo"] != "bar" {
		t.Errorf("data.foo = %v, want bar", data["foo"])
	}
}

func TestSuccessPaged(t *testing.T) {
	engine := newTestEngineWithTrace("tid-1")
	engine.GET("/x", func(c *gin.Context) {
		SuccessPaged(c, []string{"a", "b"}, 100, 2, 10)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	body := decodeBody(t, rr)
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data should be a map, got %T", body["data"])
	}
	if data["total"] != float64(100) {
		t.Errorf("total = %v, want 100", data["total"])
	}
	if data["page"] != float64(2) {
		t.Errorf("page = %v, want 2", data["page"])
	}
	if data["page_size"] != float64(10) {
		t.Errorf("page_size = %v, want 10", data["page_size"])
	}
	list, ok := data["list"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("list should be 2-element array, got %T (%v)", data["list"], data["list"])
	}
}

func TestError_AppError(t *testing.T) {
	engine := newTestEngineWithTrace("tid-2")
	engine.GET("/x", func(c *gin.Context) {
		Error(c, apperrors.InvalidParam("missing phone"))
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	body := decodeBody(t, rr)
	if body["code"] != float64(apperrors.ErrInvalidParam) {
		t.Errorf("code = %v, want %d", body["code"], apperrors.ErrInvalidParam)
	}
	if body["msg"] != "missing phone" {
		t.Errorf("msg = %v, want 'missing phone'", body["msg"])
	}
	if body["trace_id"] != "tid-2" {
		t.Errorf("trace_id = %v, want tid-2", body["trace_id"])
	}
	// data should be omitted (nil + omitempty).
	if _, has := body["data"]; has {
		t.Errorf("data should be omitted, got %v", body["data"])
	}
}

func TestError_GenericError(t *testing.T) {
	engine := newTestEngineWithTrace("tid-3")
	engine.GET("/x", func(c *gin.Context) {
		Error(c, stderrors.New("boom"))
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
	body := decodeBody(t, rr)
	if body["code"] != float64(apperrors.ErrInternal) {
		t.Errorf("code = %v, want %d", body["code"], apperrors.ErrInternal)
	}
	if body["msg"] != "boom" {
		t.Errorf("msg = %v, want 'boom'", body["msg"])
	}
}

func TestError_WrappedAppError(t *testing.T) {
	engine := newTestEngineWithTrace("tid-4")
	engine.GET("/x", func(c *gin.Context) {
		cause := stderrors.New("connection refused")
		err := apperrors.Wrap(cause, apperrors.ErrInternal, "query failed")
		Error(c, err)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
	body := decodeBody(t, rr)
	if body["code"] != float64(apperrors.ErrInternal) {
		t.Errorf("code = %v, want %d", body["code"], apperrors.ErrInternal)
	}
	// Error() should surface the user-facing msg, not the cause.
	if body["msg"] != "query failed" {
		t.Errorf("msg = %v, want 'query failed'", body["msg"])
	}
}
