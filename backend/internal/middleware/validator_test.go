package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func strReader(s string) io.Reader { return strings.NewReader(s) }

type loginReq struct {
	Phone string `json:"phone" validate:"required,phone_cn"`
	Code  string `json:"code" validate:"required,numeric,len=6"`
}

func newValidatorEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID())
	return engine
}

func TestBindAndValidate_Valid(t *testing.T) {
	engine := newValidatorEngine()
	called := false
	engine.POST("/login", func(c *gin.Context) {
		var req loginReq
		if !BindAndValidate(c, &req) {
			return
		}
		called = true
		if req.Phone != "13800138000" {
			t.Errorf("phone = %q", req.Phone)
		}
		if req.Code != "123456" {
			t.Errorf("code = %q", req.Code)
		}
		c.Status(200)
	})

	body := `{"phone":"13800138000","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200 (body=%s)", rr.Code, rr.Body.String())
	}
	if !called {
		t.Errorf("handler should have been called")
	}
}

func TestBindAndValidate_InvalidJSON(t *testing.T) {
	engine := newValidatorEngine()
	engine.POST("/login", func(c *gin.Context) {
		var req loginReq
		if !BindAndValidate(c, &req) {
			return
		}
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodPost, "/login", strReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["code"] != float64(10001) {
		t.Errorf("code = %v, want 10001", body["code"])
	}
}

func TestBindAndValidate_ValidationFail(t *testing.T) {
	engine := newValidatorEngine()
	engine.POST("/login", func(c *gin.Context) {
		var req loginReq
		if !BindAndValidate(c, &req) {
			return
		}
		c.Status(200)
	})

	cases := []struct {
		name string
		body string
	}{
		{"phone too short", `{"phone":"138","code":"123456"}`},
		{"phone wrong prefix", `{"phone":"22800138000","code":"123456"}`},
		{"code non-numeric", `{"phone":"13800138000","code":"x"}`},
		{"missing code", `{"phone":"13800138000"}`},
		{"missing both", `{}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", strReader(c.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			engine.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400 (body=%s)", rr.Code, c.body)
			}
			var resp map[string]any
			_ = json.Unmarshal(rr.Body.Bytes(), &resp)
			if resp["code"] != float64(10001) {
				t.Errorf("code = %v, want 10001", resp["code"])
			}
		})
	}
}

func TestBindAndValidateQuery_Valid(t *testing.T) {
	engine := newValidatorEngine()
	type q struct {
		Page int `form:"page" validate:"min=1"`
		Size int `form:"size" validate:"min=1,max=100"`
	}
	called := false
	engine.GET("/list", func(c *gin.Context) {
		var p q
		if !BindAndValidateQuery(c, &p) {
			return
		}
		called = true
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/list?page=1&size=10", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !called {
		t.Errorf("handler should have been called")
	}
}

func TestBindAndValidateQuery_ValidationFails(t *testing.T) {
	engine := newValidatorEngine()
	type q struct {
		Page int `form:"page" validate:"min=1"`
	}
	engine.GET("/list", func(c *gin.Context) {
		var p q
		if !BindAndValidateQuery(c, &p) {
			return
		}
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/list?page=0", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
}

func TestPhoneCNValidator(t *testing.T) {
	v := Validator()
	type s struct {
		Phone string `validate:"phone_cn"`
	}
	cases := []struct {
		phone string
		want  bool
	}{
		{"13800138000", true},  // valid
		{"15012345678", true},  // valid (15x)
		{"19900001111", true},  // valid (19x)
		{"1234", false},        // too short
		{"22800138000", false}, // wrong prefix (starts with 2)
		{"10000000000", false}, // wrong prefix (10x)
		{"1380013800", false},  // 10 digits
		{"138001380001", false}, // 12 digits
		{"", false},            // empty
	}
	for _, c := range cases {
		err := v.Struct(s{Phone: c.phone})
		got := err == nil
		if got != c.want {
			t.Errorf("phone=%q valid=%v, want %v", c.phone, got, c.want)
		}
	}
}

func TestValidator_Singleton(t *testing.T) {
	a := Validator()
	b := Validator()
	if a != b {
		t.Errorf("Validator() should return the same instance")
	}
}
