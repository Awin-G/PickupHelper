package handler

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// fakeUserSvc is a stub UserService for handler unit tests.
type fakeUserSvc struct {
	getInfoFn        func(ctx context.Context, userID int64) (*models.UserInfoDTO, error)
	updateInfoFn     func(ctx context.Context, userID int64, nickname, avatar string) (*models.UserInfoDTO, error)
	applyRunnerFn    func(ctx context.Context, userID int64, req service.ApplyRunnerRequest) (int64, int8, error)
	listRunnerAppsFn func(ctx context.Context, filter service.RunnerAppListFilter) (*service.RunnerAppListResult, error)
	auditRunnerAppFn func(ctx context.Context, appID, adminID int64, action, remark string) (*service.RunnerAppDTO, error)
	setBlacklistFn   func(ctx context.Context, userID int64, isBlacklisted bool, reason string) error
}

func (f *fakeUserSvc) GetUserInfo(ctx context.Context, userID int64) (*models.UserInfoDTO, error) {
	if f.getInfoFn != nil {
		return f.getInfoFn(ctx, userID)
	}
	return &models.UserInfoDTO{ID: userID, Phone: "138****8000", Nickname: "alice"}, nil
}
func (f *fakeUserSvc) UpdateUserInfo(ctx context.Context, userID int64, nickname, avatar string) (*models.UserInfoDTO, error) {
	if f.updateInfoFn != nil {
		return f.updateInfoFn(ctx, userID, nickname, avatar)
	}
	return &models.UserInfoDTO{ID: userID, Nickname: nickname, Avatar: avatar}, nil
}
func (f *fakeUserSvc) ApplyRunner(ctx context.Context, userID int64, req service.ApplyRunnerRequest) (int64, int8, error) {
	if f.applyRunnerFn != nil {
		return f.applyRunnerFn(ctx, userID, req)
	}
	return 1, models.AppStatusPending, nil
}
func (f *fakeUserSvc) ListRunnerApps(ctx context.Context, filter service.RunnerAppListFilter) (*service.RunnerAppListResult, error) {
	if f.listRunnerAppsFn != nil {
		return f.listRunnerAppsFn(ctx, filter)
	}
	return &service.RunnerAppListResult{Items: []*service.RunnerAppDTO{{ID: 1, RealName: "bob"}}, Total: 1}, nil
}
func (f *fakeUserSvc) AuditRunnerApp(ctx context.Context, appID, adminID int64, action, remark string) (*service.RunnerAppDTO, error) {
	if f.auditRunnerAppFn != nil {
		return f.auditRunnerAppFn(ctx, appID, adminID, action, remark)
	}
	return &service.RunnerAppDTO{ID: appID, Status: models.AppStatusApproved}, nil
}
func (f *fakeUserSvc) SetBlacklist(ctx context.Context, userID int64, isBlacklisted bool, reason string) error {
	if f.setBlacklistFn != nil {
		return f.setBlacklistFn(ctx, userID, isBlacklisted, reason)
	}
	return nil
}

// newUserHandlerEngine mounts the user handler routes with the fake service.
// authAs controls the X-User-Id / X-Role headers injected for every request;
// empty userID means no auth headers (unauthenticated).
func newUserHandlerEngine(svc *fakeUserSvc, authAs struct {
	UserID int64
	Role   string
}) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set("trace_id", "test-trace")
		if authAs.UserID > 0 {
			c.Request.Header.Set("X-User-Id", strconv.FormatInt(authAs.UserID, 10))
			c.Request.Header.Set("X-Role", authAs.Role)
		}
		c.Next()
	})
	// Mount closures using the fake service.
	userG := engine.Group("/api/v1/user")
	userG.GET("/info", func(c *gin.Context) { callGetInfo(c, svc) })
	userG.PUT("/info", func(c *gin.Context) { callUpdateInfo(c, svc) })
	userG.POST("/runner/apply", func(c *gin.Context) { callApplyRunner(c, svc) })

	adminG := engine.Group("/api/v1/admin")
	adminG.GET("/user/runner/applications", func(c *gin.Context) { callListRunnerApps(c, svc) })
	adminG.PUT("/user/runner/applications/:id/audit", func(c *gin.Context) { callAuditRunnerApp(c, svc) })
	adminG.PUT("/users/:id/blacklist", func(c *gin.Context) { callSetBlacklist(c, svc) })
	return engine
}

func callGetInfo(c *gin.Context, svc *fakeUserSvc) {
	uid, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	dto, err := svc.GetUserInfo(c.Request.Context(), uid)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

func callUpdateInfo(c *gin.Context, svc *fakeUserSvc) {
	uid, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var req UpdateProfileRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	dto, err := svc.UpdateUserInfo(c.Request.Context(), uid, req.Nickname, req.Avatar)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

func callApplyRunner(c *gin.Context, svc *fakeUserSvc) {
	uid, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing user context"))
		return
	}
	var req ApplyRunnerRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	appID, status, err := svc.ApplyRunner(c.Request.Context(), uid, service.ApplyRunnerRequest{
		RealName:    req.RealName,
		StudentID:   req.StudentID,
		IDCardImage: req.IDCardImage,
	})
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"application_id": appID, "status": status})
}

func callListRunnerApps(c *gin.Context, svc *fakeUserSvc) {
	var q ListRunnerAppsQuery
	if !middleware.BindAndValidateQuery(c, &q) {
		return
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	res, err := svc.ListRunnerApps(c.Request.Context(), service.RunnerAppListFilter{
		Status:  q.Status,
		Keyword: q.Keyword,
		Offset:  offset,
		Limit:   pageSize,
	})
	if err != nil {
		Error(c, err)
		return
	}
	SuccessPaged(c, res.Items, res.Total, page, pageSize)
}

func callAuditRunnerApp(c *gin.Context, svc *fakeUserSvc) {
	adminID, _, _, _, ok := middleware.CurrentUser(c)
	if !ok {
		Error(c, apperrors.New(apperrors.ErrUnauthenticated, "missing admin context"))
		return
	}
	appID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req AuditRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	dto, err := svc.AuditRunnerApp(c.Request.Context(), appID, adminID, req.Action, req.AuditRemark)
	if err != nil {
		Error(c, err)
		return
	}
	Success(c, dto)
}

func callSetBlacklist(c *gin.Context, svc *fakeUserSvc) {
	targetID, err := parseIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}
	var req SetBlacklistRequest
	if !middleware.BindAndValidate(c, &req) {
		return
	}
	if err := svc.SetBlacklist(c.Request.Context(), targetID, req.IsBlacklisted, req.Reason); err != nil {
		Error(c, err)
		return
	}
	Success(c, gin.H{"id": targetID})
}

func TestHandler_GetInfo_Success(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 42, Role: "user"})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/user/info", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(42), data["id"])
	assert.Equal(t, "138****8000", data["phone"])
}

func TestHandler_GetInfo_Unauthenticated(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/user/info", nil)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrUnauthenticated), body["code"])
}

func TestHandler_GetInfo_UserNotFound(t *testing.T) {
	svc := &fakeUserSvc{getInfoFn: func(_ context.Context, _ int64) (*models.UserInfoDTO, error) {
		return nil, apperrors.New(apperrors.ErrUserNotFound, "")
	}}
	engine := newUserHandlerEngine(svc, struct {
		UserID int64
		Role   string
	}{UserID: 42, Role: "user"})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/user/info", nil)
	assert.Equal(t, http.StatusNotFound, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrUserNotFound), body["code"])
}

func TestHandler_UpdateInfo_Success(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 42, Role: "user"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/user/info", UpdateProfileRequest{Nickname: "alice"})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, "alice", data["nickname"])
}

func TestHandler_UpdateInfo_NicknameTooLong(t *testing.T) {
	svc := &fakeUserSvc{updateInfoFn: func(_ context.Context, _ int64, _, _ string) (*models.UserInfoDTO, error) {
		return nil, apperrors.New(apperrors.ErrNicknameTooLong, "")
	}}
	engine := newUserHandlerEngine(svc, struct {
		UserID int64
		Role   string
	}{UserID: 42, Role: "user"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/user/info", UpdateProfileRequest{Nickname: "x"})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrNicknameTooLong), body["code"])
}

func TestHandler_ApplyRunner_Success(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 42, Role: "user"})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/user/runner/apply", ApplyRunnerRequest{
		RealName:    "Alice",
		IDCardImage: "https://example.com/id.jpg",
	})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(1), data["application_id"])
	assert.Equal(t, float64(models.AppStatusPending), data["status"])
}

func TestHandler_ApplyRunner_Duplicate(t *testing.T) {
	svc := &fakeUserSvc{applyRunnerFn: func(_ context.Context, _ int64, _ service.ApplyRunnerRequest) (int64, int8, error) {
		return 0, 0, apperrors.New(apperrors.ErrRunnerDuplicate, "")
	}}
	engine := newUserHandlerEngine(svc, struct {
		UserID int64
		Role   string
	}{UserID: 42, Role: "user"})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/user/runner/apply", ApplyRunnerRequest{
		RealName:    "Alice",
		IDCardImage: "https://example.com/id.jpg",
	})
	assert.Equal(t, http.StatusConflict, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrRunnerDuplicate), body["code"])
}

func TestHandler_ApplyRunner_MissingIDCard(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 42, Role: "user"})
	rr := doJSON(t, engine, http.MethodPost, "/api/v1/user/runner/apply", map[string]string{"real_name": "Alice"})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrInvalidParam), body["code"])
}

func TestHandler_ListRunnerApps_Success(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 1, Role: "admin"})
	rr := doJSON(t, engine, http.MethodGet, "/api/v1/admin/user/runner/applications?page=1&page_size=20", nil)
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(0), body["code"])
	data := body["data"].(map[string]any)
	items := data["list"].([]any)
	assert.Len(t, items, 1)
	assert.Equal(t, float64(1), data["total"])
}

func TestHandler_AuditRunnerApp_Approve(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 1, Role: "admin"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/admin/user/runner/applications/5/audit", AuditRequest{Action: "approve"})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(5), data["id"])
	assert.Equal(t, float64(models.AppStatusApproved), data["status"])
}

func TestHandler_AuditRunnerApp_NotFound(t *testing.T) {
	svc := &fakeUserSvc{auditRunnerAppFn: func(_ context.Context, _, _ int64, _, _ string) (*service.RunnerAppDTO, error) {
		return nil, apperrors.New(apperrors.ErrAppNotFound, "")
	}}
	engine := newUserHandlerEngine(svc, struct {
		UserID int64
		Role   string
	}{UserID: 1, Role: "admin"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/admin/user/runner/applications/5/audit", AuditRequest{Action: "approve"})
	assert.Equal(t, http.StatusNotFound, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrAppNotFound), body["code"])
}

func TestHandler_AuditRunnerApp_NotPending(t *testing.T) {
	svc := &fakeUserSvc{auditRunnerAppFn: func(_ context.Context, _, _ int64, _, _ string) (*service.RunnerAppDTO, error) {
		return nil, apperrors.New(apperrors.ErrAppNotPending, "")
	}}
	engine := newUserHandlerEngine(svc, struct {
		UserID int64
		Role   string
	}{UserID: 1, Role: "admin"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/admin/user/runner/applications/5/audit", AuditRequest{Action: "approve"})
	assert.Equal(t, http.StatusConflict, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrAppNotPending), body["code"])
}

func TestHandler_AuditRunnerApp_InvalidAction(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 1, Role: "admin"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/admin/user/runner/applications/5/audit", AuditRequest{Action: "frob"})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrInvalidParam), body["code"])
}

func TestHandler_AuditRunnerApp_InvalidID(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 1, Role: "admin"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/admin/user/runner/applications/abc/audit", AuditRequest{Action: "approve"})
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrInvalidParam), body["code"])
}

func TestHandler_SetBlacklist_Success(t *testing.T) {
	engine := newUserHandlerEngine(&fakeUserSvc{}, struct {
		UserID int64
		Role   string
	}{UserID: 1, Role: "admin"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/admin/users/42/blacklist", SetBlacklistRequest{IsBlacklisted: true})
	assert.Equal(t, http.StatusOK, rr.Code)
	body := decodeResponse(t, rr)
	data := body["data"].(map[string]any)
	assert.Equal(t, float64(42), data["id"])
}

func TestHandler_SetBlacklist_UserNotFound(t *testing.T) {
	svc := &fakeUserSvc{setBlacklistFn: func(_ context.Context, _ int64, _ bool, _ string) error {
		return apperrors.New(apperrors.ErrUserNotFound, "")
	}}
	engine := newUserHandlerEngine(svc, struct {
		UserID int64
		Role   string
	}{UserID: 1, Role: "admin"})
	rr := doJSON(t, engine, http.MethodPut, "/api/v1/admin/users/42/blacklist", SetBlacklistRequest{IsBlacklisted: true})
	assert.Equal(t, http.StatusNotFound, rr.Code)
	body := decodeResponse(t, rr)
	assert.Equal(t, float64(apperrors.ErrUserNotFound), body["code"])
}
