//go:build integration

package test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"testing"

	"pickup-helper/internal/handler"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/router"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type avatarTestEnv struct {
	env     *TestEnv
	engine  *gin.Engine
	userID  int64
	userTok string
}

func setupAvatarEngine(t *testing.T) *avatarTestEnv {
	t.Helper()
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	_ = middleware.Validator()
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	userRepo := repository.NewUserRepo()
	parcelRepo := repository.NewParcelRepo()
	shelfRepo := repository.NewShelfRepo()
	pickupRepo := repository.NewPickupLogRepo()
	proxyRepo := repository.NewProxyOrderRepo()
	notifyRepo := repository.NewNotifyRepo()

	parcelSvc := service.NewParcelService(parcelRepo, shelfRepo, userRepo, env.DB)
	pickupSvc := service.NewPickupService(parcelRepo, pickupRepo, shelfRepo, userRepo, env.DB)
	proxySvc := service.NewProxyService(proxyRepo, parcelRepo, userRepo, env.DB)
	shelfSvc := service.NewShelfService(shelfRepo, env.DB)
	notifySvc := service.NewNotifyService(notifyRepo, env.DB)
	statsSvc := service.NewStatsService(env.DB)
	avatarSvc := service.NewAvatarService(userRepo, env.DB)

	_ = SeedStation(t, env.DB)
	userID := SeedUser(t, env.DB, "13800136000")
	userTok := signParcelToken(t, env, userID, 0, "user")

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	parcelH := handler.NewParcelHandler(parcelSvc)
	pickupH := handler.NewPickupHandler(pickupSvc)
	proxyH := handler.NewProxyHandler(proxySvc)
	shelfH := handler.NewShelfHandler(shelfSvc)
	notifyH := handler.NewNotifyHandler(notifySvc)
	statsH := handler.NewStatsHandler(statsSvc)
	avatarH := handler.NewAvatarHandler(avatarSvc)
	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Parcel: parcelH,
		Pickup: pickupH,
		Proxy:  proxyH,
		Shelf:  shelfH,
		Notify: notifyH,
		Stats:  statsH,
		Avatar: avatarH,
	})

	return &avatarTestEnv{env: env, engine: engine, userID: userID, userTok: userTok}
}

func (a *avatarTestEnv) doMultipart(t *testing.T, path string, fileData []byte, fileName string, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	buf.WriteString("--boundary\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"file\"; filename=\"" + fileName + "\"\r\n")
	buf.WriteString("Content-Type: image/jpeg\r\n\r\n")
	buf.Write(fileData)
	buf.WriteString("\r\n--boundary--\r\n")

	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	a.engine.ServeHTTP(rr, req)
	return rr
}

func makeJPEG(t *testing.T, size int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}))
	return buf.Bytes()
}

func makeNonSquareJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, color.RGBA{100, 200, 50, 255})
		}
	}
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}))
	return buf.Bytes()
}

// AVATAR-01: Upload square JPEG → 200, compressed ≤150KB in DB.
func TestAvatar_01_Upload_Success(t *testing.T) {
	env := setupAvatarEngine(t)
	jpg := makeJPEG(t, 200)

	rr := env.doMultipart(t, "/api/v1/user/avatar", jpg, "avatar.jpg", env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "upload body=%s", rr.Body.String())

	var data []byte
	require.NoError(t, env.env.DB.Get(&data, "SELECT avatar_data FROM users WHERE id=?", env.userID))
	assert.NotEmpty(t, data)
	assert.LessOrEqual(t, len(data), 150*1024)
}

// AVATAR-02: Upload non-square → 400.
func TestAvatar_02_Upload_NotSquare(t *testing.T) {
	env := setupAvatarEngine(t)
	jpg := makeNonSquareJPEG(t, 200, 300)

	rr := env.doMultipart(t, "/api/v1/user/avatar", jpg, "rect.jpg", env.userTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// AVATAR-03: Upload too small (<64px) → 400.
func TestAvatar_03_Upload_TooSmall(t *testing.T) {
	env := setupAvatarEngine(t)
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}))

	rr := env.doMultipart(t, "/api/v1/user/avatar", buf.Bytes(), "tiny.jpg", env.userTok)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

// AVATAR-04: Serve own avatar → image/jpeg with data.
func TestAvatar_04_Serve_Success(t *testing.T) {
	env := setupAvatarEngine(t)
	jpg := makeJPEG(t, 200)
	env.doMultipart(t, "/api/v1/user/avatar", jpg, "avatar.jpg", env.userTok)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+env.userTok)
	rr := httptest.NewRecorder()
	env.engine.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "image/jpeg", rr.Header().Get("Content-Type"))
	assert.NotEmpty(t, rr.Body.Bytes())
}

// AVATAR-05: Serve without upload → 404.
func TestAvatar_05_Serve_NotFound(t *testing.T) {
	env := setupAvatarEngine(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+env.userTok)
	rr := httptest.NewRecorder()
	env.engine.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// AVATAR-06: Upload unauthenticated → 401.
func TestAvatar_06_Unauthenticated(t *testing.T) {
	env := setupAvatarEngine(t)
	jpg := makeJPEG(t, 200)

	rr := env.doMultipart(t, "/api/v1/user/avatar", jpg, "avatar.jpg", "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// AVATAR-07: Serve by user ID → image/jpeg.
func TestAvatar_07_ServeByID(t *testing.T) {
	env := setupAvatarEngine(t)
	jpg := makeJPEG(t, 200)
	env.doMultipart(t, "/api/v1/user/avatar", jpg, "avatar.jpg", env.userTok)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+itoa(env.userID)+"/avatar", nil)
	req.Header.Set("Authorization", "Bearer "+env.userTok)
	rr := httptest.NewRecorder()
	env.engine.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "image/jpeg", rr.Header().Get("Content-Type"))
}

// AVATAR-08: Upload large image → resized and compressed to ≤150KB.
func TestAvatar_08_Upload_LargeImage(t *testing.T) {
	env := setupAvatarEngine(t)
	jpg := makeJPEG(t, 1024)

	rr := env.doMultipart(t, "/api/v1/user/avatar", jpg, "big.jpg", env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)

	var data []byte
	require.NoError(t, env.env.DB.Get(&data, "SELECT avatar_data FROM users WHERE id=?", env.userID))
	assert.LessOrEqual(t, len(data), 150*1024)
}
