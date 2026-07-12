package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAuthSvc is a stub AuthService for handler unit tests. Each method
// is controlled by a function field; when nil, the method returns a
// sensible zero-value default.
type fakeAuthSvc struct {
	sendCodeFn       func(ctx context.Context, phone, ip string) (int, error)
	loginFn          func(ctx context.Context, phone, code, openid string) (*service.LoginResult, error)
	refreshFn        func(ctx context.Context, refreshToken string) (*service.RefreshResult, error)
	adminLoginFn     func(ctx context.Context, username, password string) (*service.LoginResult, error)
	listActiveCodesFn func(ctx context.Context) ([]repository.ActiveCode, error)
}

func (f *fakeAuthSvc) SendCode(ctx context.Context, phone, ip string) (int, error) {
	if f.sendCodeFn != nil {
		return f.sendCodeFn(ctx, phone, ip)
	}
	return 300, nil
}
func (f *fakeAuthSvc) Login(ctx context.Context, phone, code, openid string) (*service.LoginResult, error) {
	if f.loginFn != nil {
		return f.loginFn(ctx, phone, code, openid)
	}
	return &service.LoginResult{AccessToken: "tok", RefreshToken: "ref", ExpiresIn: 7200, Role: "user"}, nil
}
func (f *fakeAuthSvc) RefreshToken(ctx context.Context, rt string) (*service.RefreshResult, error) {
	if f.refreshFn != nil {
		return f.refreshFn(ctx, rt)
	}
	return &service.RefreshResult{AccessToken: "new-tok", ExpiresIn: 7200}, nil
}
func (f *fakeAuthSvc) AdminLogin(ctx context.Context, u, p string) (*service.LoginResult, error) {
	if f.adminLoginFn != nil {
		return f.adminLoginFn(ctx, u, p)
	}
	return &service.LoginResult{AccessToken: "admin-tok", ExpiresIn: 7200, Role: "admin"}, nil
}
func (f *fakeAuthSvc) ListActiveCodes(ctx context.Context) ([]repository.ActiveCode, error) {
	if f.listActiveCodesFn != nil {
		return f.listActiveCodesFn(ctx)
	}
	return []repository.ActiveCode{{Phone: "13800138000", Code: "123456", ExpireIn: 120}}, nil
}

// newAuthHandlerEngine builds a gin engine with the TraceID middleware and
// the auth handler's public + admin-auth routes mounted, delegating to the
// fake service. We mount closures (rather than the real AuthHandler) so
// we can inject a fake without changing AuthHandler's concrete dependency.
func newAuthHandlerEngine(svc *fakeAuthSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set("trace_id", "test-trace")
		c.Next()
	})
	public := engine.Group("/api/v1/auth")
	public.POST("/send-code", func(c *gin.Context) {
		var req SendCodeRequest
		if !middleware.BindAndValidate(c, &req) {
			return
		}
		exp, err := svc.SendCode(c.Request.Context(), req.Phone, c.ClientIP())
		if err != nil {
			Error(c, err)
			return
		}
		Success(c, SendCodeResponse{ExpireIn: exp})
	})
	public.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if !middleware.BindAndValidate(c, &req) {
			return
		}
		res, err := svc.Login(c.Request.Context(), req.Phone, req.Code, req.OpenID)
		if err != nil {
			Error(c, err)
			return
		}
		Success(c, res)
	})
	public.POST("/refresh", func(c *gin.Context) {
		var req RefreshRequest
		if !middleware.BindAndValidate(c, &req) {
			return
		}
		res, err := svc.RefreshToken(c.Request.Context(), req.RefreshToken)
		if err != nil {
			Error(c, err)
			return
		}
		Success(c, res)
	})
	engine.POST("/api/v1/admin/auth/login", func(c *gin.Context) {
		var req AdminLoginRequest
		if !middleware.BindAndValidate(c, &req) {
			return
		}
		res, err := svc.AdminLogin(c.Request.Context(), req.Username, req.Password)
		if err != nil {
			Error(c, err)
			return
		}
		Success(c, res)
	})
	adminMG := engine.Group("/api/v1/admin")
	adminMG.GET("/auth/sms-codes", func(c *gin.Context) {
		codes, err := svc.ListActiveCodes(c.Request.Context())
		if err != nil {
			Error(c, err)
			return
		}
		Success(c, ListActiveCodesResponse{Codes: codes, Total: len(codes)})
	})
	adminMG.GET("/get-async-routes", func(c *gin.Context) {
		h := &AuthHandler{}
		h.GetAsyncRoutes(c)
	})
	return engine
}

func doJSON(t *testing.T, engine *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	// Ensure phone_cn validator is registered on gin's binding validator
	// before any ShouldBindJSON call panics on the unregistered tag.
	_ = middleware.Validator()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)
	return rr
}

func decodeResponse(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

func TestHandler_SendCode_Success(t *testing.T) {
	engine := newAuthHandlerEngine(&fakeAuthSvc{})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/auth/send-code", SendCodeRequest{Phone: "13800138000"})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(0), body["code"])
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(300), data["expire_in"])
}

func TestHandler_SendCode_InvalidPhone(t *testing.T) {
	engine := newAuthHandlerEngine(&fakeAuthSvc{})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/auth/send-code", map[string]string{"phone": "123"})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrInvalidParam), body["code"])
}

func TestHandler_SendCode_RateLimited(t *testing.T) {
	svc := &fakeAuthSvc{sendCodeFn: func(_ context.Context, _, _ string) (int, error) {
		return 0, apperrors.New(apperrors.ErrSMSTooFrequent, "")
	}}
	engine := newAuthHandlerEngine(svc)
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/auth/send-code", SendCodeRequest{Phone: "13800138000"})
	assert.Equal(t, http.StatusForbidden, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrSMSTooFrequent), body["code"])
}

func TestHandler_Login_Success(t *testing.T) {
	engine := newAuthHandlerEngine(&fakeAuthSvc{})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/auth/login", LoginRequest{Phone: "13800138000", Code: "123456"})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.NotEmpty(t, data["access_token"])
}

func TestHandler_Login_Blacklisted(t *testing.T) {
	svc := &fakeAuthSvc{loginFn: func(_ context.Context, _, _, _ string) (*service.LoginResult, error) {
		return nil, apperrors.New(apperrors.ErrUserBlacklisted, "")
	}}
	engine := newAuthHandlerEngine(svc)
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/auth/login", LoginRequest{Phone: "13800138000", Code: "123456"})
	assert.Equal(t, http.StatusForbidden, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrUserBlacklisted), body["code"])
}

func TestHandler_Refresh_Success(t *testing.T) {
	engine := newAuthHandlerEngine(&fakeAuthSvc{})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/auth/refresh", RefreshRequest{RefreshToken: "old"})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, "new-tok", data["access_token"])
}

func TestHandler_Refresh_Expired(t *testing.T) {
	svc := &fakeAuthSvc{refreshFn: func(_ context.Context, _ string) (*service.RefreshResult, error) {
		return nil, apperrors.New(apperrors.ErrRefreshExpired, "")
	}}
	engine := newAuthHandlerEngine(svc)
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/auth/refresh", RefreshRequest{RefreshToken: "old"})
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrRefreshExpired), body["code"])
}

func TestHandler_AdminLogin_Success(t *testing.T) {
	engine := newAuthHandlerEngine(&fakeAuthSvc{})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/admin/auth/login", AdminLoginRequest{Username: "admin", Password: "secret123"})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, "admin-tok", data["access_token"])
	assert.Equal(t, "admin", data["role"])
}

func TestHandler_AdminLogin_Disabled(t *testing.T) {
	svc := &fakeAuthSvc{adminLoginFn: func(_ context.Context, _, _ string) (*service.LoginResult, error) {
		return nil, apperrors.New(apperrors.ErrAdminDisabled, "")
	}}
	engine := newAuthHandlerEngine(svc)
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/admin/auth/login", AdminLoginRequest{Username: "admin", Password: "secret123"})
	assert.Equal(t, http.StatusForbidden, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrAdminDisabled), body["code"])
}

func TestHandler_ListActiveCodes_Success(t *testing.T) {
	engine := newAuthHandlerEngine(&fakeAuthSvc{})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/admin/auth/sms-codes", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	codes := data["codes"].([]any)
	assert.Len(t, codes, 1)
	assert.Equal(t, float64(1), data["total"])
}

func TestHandler_ListActiveCodes_Empty(t *testing.T) {
	svc := &fakeAuthSvc{listActiveCodesFn: func(_ context.Context) ([]repository.ActiveCode, error) {
		return []repository.ActiveCode{}, nil
	}}
	engine := newAuthHandlerEngine(svc)
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/admin/auth/sms-codes", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(0), data["total"])
}

func TestHandler_GetAsyncRoutes_Success(t *testing.T) {
	engine := newAuthHandlerEngine(&fakeAuthSvc{})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/admin/get-async-routes", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	routes := data["routes"].([]any)
	assert.Greater(t, len(routes), 0)
}

// guard against unused imports when no test exercises them.
var _ = errors.New
var _ = models.MaskPhone
