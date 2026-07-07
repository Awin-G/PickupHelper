//go:build integration

package test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pickup-helper/internal/handler"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"
	"pickup-helper/internal/router"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type notifyTestEnv struct {
	env     *TestEnv
	engine  *gin.Engine
	userID  int64
	userTok string
}

func setupNotifyEngine(t *testing.T) *notifyTestEnv {
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

	_ = SeedStation(t, env.DB)
	userID := SeedUser(t, env.DB, "13800138700")
	userTok := signParcelToken(t, env, userID, 0, "user")

	healthH := handler.NewHealthHandler(env.DB, env.Rdb)
	parcelH := handler.NewParcelHandler(parcelSvc)
	pickupH := handler.NewPickupHandler(pickupSvc)
	proxyH := handler.NewProxyHandler(proxySvc)
	shelfH := handler.NewShelfHandler(shelfSvc)
	notifyH := handler.NewNotifyHandler(notifySvc)
	router.Register(engine, env.Cfg, &router.Handlers{
		Health: healthH,
		Parcel: parcelH,
		Pickup: pickupH,
		Proxy:  proxyH,
		Shelf:  shelfH,
		Notify: notifyH,
	})

	return &notifyTestEnv{env: env, engine: engine, userID: userID, userTok: userTok}
}

func (n *notifyTestEnv) do(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf strings.Builder
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, strings.NewReader(buf.String()))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	n.engine.ServeHTTP(rr, req)
	return rr
}

func notifyBodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &m), "body=%s", rr.Body.String())
	return m
}

// Seed a notification directly in DB.
func seedNotify(t *testing.T, db *sql.DB, userID int64, title string, isRead int8) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO notifications (user_id, title, content, type, is_read, send_status, channel)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, title, "test content", models.NotifyTypeSystem, isRead, models.SendStatusSent, models.ChannelSMS)
	require.NoError(t, err)
}

// NOTIFY-01: List notifications → returns items with unread count.
func TestNotify_01_List(t *testing.T) {
	env := setupNotifyEngine(t)

	seedNotify(t, env.env.DB.DB, env.userID, "通知1", 0)
	seedNotify(t, env.env.DB.DB, env.userID, "通知2", 1)
	seedNotify(t, env.env.DB.DB, env.userID, "通知3", 0)

	rr := env.do(t, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "list body=%s", rr.Body.String())
	b := notifyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	list := data["items"].([]any)
	assert.Len(t, list, 3)
	assert.Equal(t, float64(3), data["total"])
	assert.Equal(t, float64(2), data["unread"])
}

// NOTIFY-02: Mark all read → unread count becomes 0.
func TestNotify_02_MarkAllRead(t *testing.T) {
	env := setupNotifyEngine(t)

	seedNotify(t, env.env.DB.DB, env.userID, "通知1", 0)
	seedNotify(t, env.env.DB.DB, env.userID, "通知2", 0)

	rr := env.do(t, http.MethodPut, "/api/v1/notifications/read", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code, "mark-read body=%s", rr.Body.String())

	// Check unread count is 0.
	rr = env.do(t, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, env.userTok)
	b := notifyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(0), data["unread"])

	// Individual items now have is_read=true.
	item := data["items"].([]any)[0].(map[string]any)
	assert.Equal(t, true, item["is_read"])
}

// NOTIFY-03: Empty notification list → returns empty page.
func TestNotify_03_EmptyList(t *testing.T) {
	env := setupNotifyEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, env.userTok)
	require.Equal(t, http.StatusOK, rr.Code)
	b := notifyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	assert.Equal(t, float64(0), data["total"])
	assert.Equal(t, float64(0), data["unread"])
}

// NOTIFY-04: List notifications unauthenticated → 401.
func TestNotify_04_Unauthenticated(t *testing.T) {
	env := setupNotifyEngine(t)

	rr := env.do(t, http.MethodGet, "/api/v1/notifications", nil, "")
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// NOTIFY-05: Only own notifications are visible.
func TestNotify_05_Isolation(t *testing.T) {
	env := setupNotifyEngine(t)
	otherID := SeedUser(t, env.env.DB, "13900138999")

	seedNotify(t, env.env.DB.DB, env.userID, "我的通知", 0)
	seedNotify(t, env.env.DB.DB, otherID, "别人的通知", 0)

	rr := env.do(t, http.MethodGet, "/api/v1/notifications?page=1&page_size=10", nil, env.userTok)
	b := notifyBodyMap(t, rr)
	data := b["data"].(map[string]any)
	items := data["items"].([]any)
	assert.Len(t, items, 1)
}
