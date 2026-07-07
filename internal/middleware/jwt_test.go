package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pickup-helper/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func mustConfig(secret string) *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:  secret,
			RefreshSecret: secret + "-refresh",
			AccessTTL:     time.Hour,
			RefreshTTL:    24 * time.Hour,
			Issuer:        "pickup-helper-test",
		},
	}
}

func makeToken(t *testing.T, secret string, claims Claims) string {
	t.Helper()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

func newJWTEngine(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID())
	engine.Use(JWTAuth(cfg))
	return engine
}

func TestSignAccess_ParseAccess_RoundTrip(t *testing.T) {
	cfg := mustConfig("top-secret")
	in := Claims{UserID: 42, UserType: 2, StationID: 7, Role: "admin"}
	tok, err := SignAccess(cfg, in)
	if err != nil {
		t.Fatalf("SignAccess: %v", err)
	}
	out, err := ParseAccess(cfg, tok)
	if err != nil {
		t.Fatalf("ParseAccess: %v", err)
	}
	if out.UserID != 42 || out.UserType != 2 || out.StationID != 7 || out.Role != "admin" {
		t.Errorf("claims mismatch: %+v", out)
	}
}

func TestParseAccess_Expired(t *testing.T) {
	cfg := mustConfig("top-secret")
	claims := Claims{UserID: 1}
	claims.RegisteredClaims = jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWT.AccessSecret))
	if _, err := ParseAccess(cfg, tok); err == nil {
		t.Errorf("ParseAccess should fail on expired token")
	}
}

func TestParseAccess_WrongSecret(t *testing.T) {
	cfg := mustConfig("correct-secret")
	tok := makeToken(t, "wrong-secret", Claims{UserID: 1})
	if _, err := ParseAccess(cfg, tok); err == nil {
		t.Errorf("ParseAccess should fail on wrong-secret token")
	}
}

func TestJWTAuth_NoHeader(t *testing.T) {
	engine := newJWTEngine(mustConfig("secret"))
	engine.GET("/x", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["code"] != float64(10002) {
		t.Errorf("code = %v, want 10002", body["code"])
	}
}

func TestJWTAuth_InvalidScheme(t *testing.T) {
	engine := newJWTEngine(mustConfig("secret"))
	engine.GET("/x", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Basic abc")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	cfg := mustConfig("top-secret")
	engine := newJWTEngine(cfg)
	var (
		hdrUserID    string
		hdrStationID string
		hdrRole      string
	)
	engine.GET("/x", func(c *gin.Context) {
		hdrUserID = c.GetHeader("X-User-Id")
		hdrStationID = c.GetHeader("X-Station-Id")
		hdrRole = c.GetHeader("X-Role")
		c.Status(200)
	})

	tok := makeToken(t, cfg.JWT.AccessSecret, Claims{UserID: 42, UserType: 2, StationID: 7, Role: "admin"})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", rr.Code, rr.Body.String())
	}
	if hdrUserID != "42" {
		t.Errorf("X-User-Id = %q, want 42", hdrUserID)
	}
	if hdrStationID != "7" {
		t.Errorf("X-Station-Id = %q, want 7", hdrStationID)
	}
	if hdrRole != "admin" {
		t.Errorf("X-Role = %q, want admin", hdrRole)
	}
}

func TestJWTAuth_OverridesClientHeader(t *testing.T) {
	cfg := mustConfig("top-secret")
	engine := newJWTEngine(cfg)
	var hdrUserID string
	engine.GET("/x", func(c *gin.Context) {
		hdrUserID = c.GetHeader("X-User-Id")
		c.Status(200)
	})

	tok := makeToken(t, cfg.JWT.AccessSecret, Claims{UserID: 99})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	// Client tries to forge identity.
	req.Header.Set("X-User-Id", "1")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if hdrUserID != "99" {
		t.Errorf("X-User-Id = %q, want 99 (token should override forged value)", hdrUserID)
	}
}

func TestJWTAuth_Expired(t *testing.T) {
	cfg := mustConfig("top-secret")
	engine := newJWTEngine(cfg)
	engine.GET("/x", func(c *gin.Context) { c.Status(200) })

	claims := Claims{UserID: 1}
	claims.RegisteredClaims = jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWT.AccessSecret))

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &body)
	if body["msg"] == nil || !contains(body["msg"].(string), "expired") && !contains(body["msg"].(string), "invalid") {
		t.Errorf("msg should mention expired/invalid, got %v", body["msg"])
	}
}

func TestJWTAuthOptional_NoHeaderPasses(t *testing.T) {
	cfg := mustConfig("top-secret")
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(JWTAuthOptional(cfg))
	called := false
	engine.GET("/x", func(c *gin.Context) {
		called = true
		c.Status(200)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200 (no header should pass)", rr.Code)
	}
	if !called {
		t.Errorf("handler should have been called")
	}
}

func TestJWTAuthOptional_InvalidTokenRejects(t *testing.T) {
	cfg := mustConfig("top-secret")
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(JWTAuthOptional(cfg))
	engine.GET("/x", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(s[0:len(sub)] == sub) || contains(s[1:], sub))
}

// --- Phase 2: ParseRefresh + CurrentUser ---

func TestParseRefresh_Success(t *testing.T) {
	cfg := mustConfig("top-secret")
	tok, err := SignRefresh(cfg, Claims{UserID: 7, Role: "user"})
	if err != nil {
		t.Fatalf("SignRefresh: %v", err)
	}
	out, err := ParseRefresh(cfg, tok)
	if err != nil {
		t.Fatalf("ParseRefresh: %v", err)
	}
	if out.UserID != 7 || out.Role != "user" {
		t.Errorf("claims mismatch: %+v", out)
	}
}

func TestParseRefresh_Invalid(t *testing.T) {
	cfg := mustConfig("top-secret")
	if _, err := ParseRefresh(cfg, "not-a-real-jwt"); err == nil {
		t.Errorf("ParseRefresh should fail on garbage input")
	}
}

func TestParseRefresh_Expired(t *testing.T) {
	cfg := mustConfig("top-secret")
	claims := Claims{UserID: 1}
	claims.RegisteredClaims = jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}
	tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWT.RefreshSecret))
	if _, err := ParseRefresh(cfg, tok); err == nil {
		t.Errorf("ParseRefresh should fail on expired token")
	}
}

func TestCurrentUser_FullHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Id", "42")
	req.Header.Set("X-User-Type", "2")
	req.Header.Set("X-Station-Id", "7")
	req.Header.Set("X-Role", "admin")
	c.Request = req

	uid, ut, sid, role, ok := CurrentUser(c)
	if !ok {
		t.Fatalf("ok = false, want true")
	}
	if uid != 42 {
		t.Errorf("userID = %d, want 42", uid)
	}
	if ut != 2 {
		t.Errorf("userType = %d, want 2", ut)
	}
	if sid != 7 {
		t.Errorf("stationID = %d, want 7", sid)
	}
	if role != "admin" {
		t.Errorf("role = %q, want admin", role)
	}
}

func TestCurrentUser_NoHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	_, _, _, _, ok := CurrentUser(c)
	if ok {
		t.Errorf("ok = true, want false when X-User-Id missing")
	}
}

func TestCurrentUser_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-Id", "not-a-number")
	c.Request = req

	_, _, _, _, ok := CurrentUser(c)
	if ok {
		t.Errorf("ok = true, want false when X-User-Id is non-numeric")
	}
}
