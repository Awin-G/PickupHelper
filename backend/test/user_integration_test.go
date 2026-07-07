//go:build integration

package test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pickup-helper/internal/handler"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/router"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// userTestEnv bundles the engine and env for user-module integration tests.
type userTestEnv struct {
	env    *TestEnv
	engine *gin.Engine
}

func setupUserEngine(t *testing.T) *userTestEnv {
	t.Helper()
	env := SetupTestEnv(t)
	// Pre-initialise the custom validator (registers phone_cn on gin's
	// binding validator) so ShouldBindJSON doesn't panic on first request.
	_ = middleware.Validator()
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	userRepo := repository.NewUserRepo()
	adminRepo := repository.NewAdminRepo()
	runnerRepo := repository.NewRunnerAppRepo()
	smsCache := repository.NewSMSCodeCache(env.Rdb)
	sms := service.NewSMSProvider("test", slog.Default())
	authSvc := service.NewAuthService(userRepo, adminRepo, smsCache, sms, env.Cfg, env.DB)
	userSvc := service.NewUserService(userRepo, runnerRepo, env.DB)

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Auth:   authH,
		User:   userH,
	})

	return &userTestEnv{env: env, engine: engine}
}

func (u *userTestEnv) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	u.engine.ServeHTTP(rr, req)
	return rr
}

func bodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

// signUserToken creates a valid access token for the given user id.
func signUserToken(t *testing.T, env *TestEnv, userID int64, role string) string {
	t.Helper()
	claims := middleware.Claims{
		UserID:   userID,
		UserType: int(models.UserTypeNormal),
		Role:     role,
	}
	tok, err := middleware.SignAccess(env.Cfg, claims)
	require.NoError(t, err)
	return tok
}

// USER-01: SendCode → Login new user → profile is masked.
func TestUser_01_SendCode_Login_NewUser(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)

	// 1. Send code.
	rr := u.do(t, http.MethodPost, "/api/v1/auth/send-code", map[string]string{"phone": "13800138000"}, "")
	require.Equal(t, http.StatusOK, rr.Code, "send-code body=%s", rr.Body.String())

	// 2. Login with stub code "123456".
	rr = u.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"phone": "13800138000", "code": "123456",
	}, "")
	require.Equal(t, http.StatusOK, rr.Code, "login body=%s", rr.Body.String())
	body := bodyMap(t, rr)
	assert.Equal(t, float64(0), body["code"])
	data := body["data"].(map[string]any)
	assert.NotEmpty(t, data["access_token"])
	assert.Equal(t, "user", data["role"])
	// Phone must be masked in the returned user DTO.
	user := data["user"].(map[string]any)
	assert.Equal(t, "138****8000", user["phone"])
}

// USER-02: Login with wrong code → 401 ErrCodeInvalid.
func TestUser_02_Login_WrongCode(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)

	u.do(t, http.MethodPost, "/api/v1/auth/send-code", map[string]string{"phone": "13800138001"}, "")
	rr := u.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"phone": "13800138001", "code": "999999",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	body := bodyMap(t, rr)
	assert.Equal(t, float64(10111), body["code"])
}

// USER-03: SendCode rate limit (60s per phone).
func TestUser_03_SendCode_RateLimit_Phone(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)

	rr1 := u.do(t, http.MethodPost, "/api/v1/auth/send-code", map[string]string{"phone": "13800138002"}, "")
	require.Equal(t, http.StatusOK, rr1.Code)
	rr2 := u.do(t, http.MethodPost, "/api/v1/auth/send-code", map[string]string{"phone": "13800138002"}, "")
	assert.Equal(t, http.StatusForbidden, rr2.Code)
	body := bodyMap(t, rr2)
	assert.Equal(t, float64(10102), body["code"])
}

// USER-04: Blacklisted user cannot request code nor login.
func TestUser_04_Login_Blacklisted(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	SeedBlackUser(t, u.env.DB, "13800138003")

	// 1. SendCode rejects blacklisted phone with ErrPhoneBlacklisted.
	rr := u.do(t, http.MethodPost, "/api/v1/auth/send-code", map[string]string{"phone": "13800138003"}, "")
	assert.Equal(t, http.StatusForbidden, rr.Code)
	body := bodyMap(t, rr)
	assert.Equal(t, float64(10104), body["code"])

	// 2. Login also rejects blacklisted users. We inject code into Redis
	// directly to bypass SendCode's blacklist check at the phone level.
	require.NoError(t, u.env.Rdb.Set(context.Background(), "sms:code:13800138003", "123456", 5*time.Minute).Err())
	rr = u.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"phone": "13800138003", "code": "123456",
	}, "")
	assert.Equal(t, http.StatusForbidden, rr.Code)
	body = bodyMap(t, rr)
	assert.Equal(t, float64(10113), body["code"])
}

// USER-05: GetInfo / UpdateInfo with JWT.
func TestUser_05_GetInfo_UpdateInfo(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	uid := SeedUser(t, u.env.DB, "13800138004")
	tok := signUserToken(t, u.env, uid, "user")

	// GET /user/info
	rr := u.do(t, http.MethodGet, "/api/v1/user/info", nil, tok)
	require.Equal(t, http.StatusOK, rr.Code)
	body := bodyMap(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, "138****8004", data["phone"])

	// PUT /user/info
	rr = u.do(t, http.MethodPut, "/api/v1/user/info", map[string]string{
		"nickname": "alice", "avatar": "https://cdn/avatar.png",
	}, tok)
	require.Equal(t, http.StatusOK, rr.Code)
	body = bodyMap(t, rr)
	data = body["data"].(map[string]any)
	assert.Equal(t, "alice", data["nickname"])
	assert.Equal(t, "https://cdn/avatar.png", data["avatar"])
}

// USER-06: UpdateInfo unauthenticated → 401.
func TestUser_06_GetInfo_Unauthenticated(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)

	rr := u.do(t, http.MethodGet, "/api/v1/user/info", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// USER-07: Apply runner → audit approve → user_type=2.
func TestUser_07_ApplyRunner_AuditApprove(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	uid := SeedUser(t, u.env.DB, "13800138005")
	userTok := signUserToken(t, u.env, uid, "user")

	// 1. Apply.
	rr := u.do(t, http.MethodPost, "/api/v1/user/runner/apply", map[string]string{
		"real_name":     "Bob",
		"id_card_image": "https://cdn/bob-id.jpg",
	}, userTok)
	require.Equal(t, http.StatusOK, rr.Code, "apply body=%s", rr.Body.String())
	body := bodyMap(t, rr)
	data := body["data"].(map[string]any)
	appID := int64(data["application_id"].(float64))
	assert.Equal(t, float64(models.AppStatusPending), data["status"])

	// 2. Admin audit approve. Need admin token.
	admin := SeedAdmin(t, u.env.DB, "auditor")
	adminTok := signUserToken(t, u.env, admin.ID, "admin")

	rr = u.do(t, http.MethodPut,
		"/api/v1/admin/user/runner/applications/"+itoa(appID)+"/audit",
		map[string]string{"action": "approve", "audit_remark": "ok"},
		adminTok)
	require.Equal(t, http.StatusOK, rr.Code, "audit body=%s", rr.Body.String())
	body = bodyMap(t, rr)
	data = body["data"].(map[string]any)
	assert.Equal(t, float64(models.AppStatusApproved), data["status"])

	// 3. Verify user_type flipped to 2 in DB.
	var ut int8
	require.NoError(t, u.env.DB.Get(&ut, "SELECT user_type FROM users WHERE id=?", uid))
	assert.Equal(t, int8(models.UserTypeRunner), ut)
}

// USER-08: Duplicate runner application → 409.
func TestUser_08_ApplyRunner_Duplicate(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	// Seed a user already in pending state.
	SeedUserWithStatus(t, u.env.DB, "13800138006", models.UserTypeNormal, models.RunnerStatusPending)
	// We need the user's id for the token; query it.
	var uid int64
	require.NoError(t, u.env.DB.Get(&uid, "SELECT id FROM users WHERE phone=?", "13800138006"))
	tok := signUserToken(t, u.env, uid, "user")

	rr := u.do(t, http.MethodPost, "/api/v1/user/runner/apply", map[string]string{
		"real_name":     "Carol",
		"id_card_image": "https://cdn/carol-id.jpg",
	}, tok)
	assert.Equal(t, http.StatusConflict, rr.Code)
	body := bodyMap(t, rr)
	assert.Equal(t, float64(10141), body["code"])
}

// USER-09: Audit non-existent application → 404.
func TestUser_09_Audit_NotFound(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	admin := SeedAdmin(t, u.env.DB, "auditor2")
	tok := signUserToken(t, u.env, admin.ID, "admin")

	rr := u.do(t, http.MethodPut, "/api/v1/admin/user/runner/applications/99999/audit",
		map[string]string{"action": "approve"}, tok)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	body := bodyMap(t, rr)
	assert.Equal(t, float64(10151), body["code"])
}

// USER-10: SetBlacklist + admin route protected by AdminOnly.
func TestUser_10_SetBlacklist_AdminGuard(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	uid := SeedUser(t, u.env.DB, "13800138007")

	// 1. Non-admin user token → 403.
	userTok := signUserToken(t, u.env, uid, "user")
	rr := u.do(t, http.MethodPut, "/api/v1/admin/users/"+itoa(uid)+"/blacklist",
		map[string]any{"is_blacklisted": true}, userTok)
	assert.Equal(t, http.StatusForbidden, rr.Code)
	body := bodyMap(t, rr)
	assert.Equal(t, float64(10003), body["code"])

	// 2. Admin token → 200.
	admin := SeedAdmin(t, u.env.DB, "superadmin")
	adminTok := signUserToken(t, u.env, admin.ID, "admin")
	rr = u.do(t, http.MethodPut, "/api/v1/admin/users/"+itoa(uid)+"/blacklist",
		map[string]any{"is_blacklisted": true}, adminTok)
	assert.Equal(t, http.StatusOK, rr.Code)

	// 3. Verify flag persisted.
	var flag int8
	require.NoError(t, u.env.DB.Get(&flag, "SELECT is_blacklisted FROM users WHERE id=?", uid))
	assert.Equal(t, int8(1), flag)
}

// USER-11 (bonus): Admin login flow end-to-end.
func TestUser_11_AdminLogin_Success(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	SeedAdminWithPassword(t, u.env.DB, "root", "passw0rd")

	rr := u.do(t, http.MethodPost, "/api/v1/admin/auth/login", map[string]string{
		"username": "root", "password": "passw0rd",
	}, "")
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	body := bodyMap(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, "admin", data["role"])
	assert.NotEmpty(t, data["access_token"])
}

func TestUser_12_AdminLogin_Disabled(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	SeedDisabledAdmin(t, u.env.DB, "ghost")

	rr := u.do(t, http.MethodPost, "/api/v1/admin/auth/login", map[string]string{
		"username": "ghost", "password": "test-password-123",
	}, "")
	assert.Equal(t, http.StatusForbidden, rr.Code)
	body := bodyMap(t, rr)
	assert.Equal(t, float64(10162), body["code"])
}

// USER-13: Refresh token flow.
func TestUser_13_RefreshToken(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)
	// Trigger login to obtain refresh token.
	u.do(t, http.MethodPost, "/api/v1/auth/send-code", map[string]string{"phone": "13800138008"}, "")
	rr := u.do(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"phone": "13800138008", "code": "123456",
	}, "")
	require.Equal(t, http.StatusOK, rr.Code)
	body := bodyMap(t, rr)
	data := body["data"].(map[string]any)
	refresh := data["refresh_token"].(string)
	require.NotEmpty(t, refresh)

	// Refresh.
	rr = u.do(t, http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": refresh,
	}, "")
	require.Equal(t, http.StatusOK, rr.Code, "refresh body=%s", rr.Body.String())
	body = bodyMap(t, rr)
	data = body["data"].(map[string]any)
	assert.NotEmpty(t, data["access_token"])
}

// USER-14: Refresh with expired token → 401 ErrRefreshExpired.
func TestUser_14_RefreshToken_Expired(t *testing.T) {
	u := setupUserEngine(t)
	TruncateAll(t, u.env.DB)

	claims := middleware.Claims{
		UserID:   1,
		UserType: int(models.UserTypeNormal),
		Role:     "user",
	}
	claims.RegisteredClaims.IssuedAt = jwt.NewNumericDate(time.Now().Add(-2 * time.Hour))
	claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(u.env.Cfg.JWT.RefreshSecret))
	require.NoError(t, err)

	rr := u.do(t, http.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": tok,
	}, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	body := bodyMap(t, rr)
	assert.Equal(t, float64(10122), body["code"])
}

// itoa is a small helper to avoid strconv import noise.
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// guard against unused imports when variants don't exercise them.
var (
	_ = sql.ErrNoRows
	_ = strings.TrimSpace
)
