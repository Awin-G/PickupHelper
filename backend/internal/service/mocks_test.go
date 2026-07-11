package service

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/redis/go-redis/v9"
)

// --- mockUserRepo ---

type mockUserRepo struct {
	FindByPhoneFn        func(ctx context.Context, db repository.DBTX, phone string) (*models.User, error)
	FindByIDFn           func(ctx context.Context, db repository.DBTX, id int64) (*models.User, error)
	FindByOpenIDFn       func(ctx context.Context, db repository.DBTX, openid string) (*models.User, error)
	CreateFn             func(ctx context.Context, db repository.DBTX, phone, openid string) (int64, error)
	CreateAdminFn        func(ctx context.Context, db repository.DBTX, phone, nickname string, userType int8) (int64, error)
	UpdateProfileFn      func(ctx context.Context, db repository.DBTX, id int64, nickname, avatar string) error
	UpdateRunnerStatusFn func(ctx context.Context, db repository.DBTX, id int64, userType, runnerStatus int8) error
	SetBlacklistFn       func(ctx context.Context, db repository.DBTX, id int64, isBlacklisted int8) error
	UpdateOpenIDFn func(ctx context.Context, db repository.DBTX, id int64, openid string) error
	SaveAvatarFn   func(ctx context.Context, db repository.DBTX, id int64, data []byte, contentType string) error
	ListUsersFn    func(ctx context.Context, db repository.DBTX, filter repository.UserListFilter) ([]*models.User, int64, error)
	UpdateUserFn   func(ctx context.Context, db repository.DBTX, id int64, cols []string, args []any) error
	DeleteUserFn   func(ctx context.Context, db repository.DBTX, id int64) error

	// Call records for assertions.
	CreateCalls             []createCall
	UpdateProfileCalls      []updateProfileCall
	UpdateRunnerStatusCalls []updateRunnerStatusCall
	SetBlacklistCalls       []setBlacklistCall
	UpdateOpenIDCalls       []updateOpenIDCall
}

type createCall struct{ Phone, OpenID string }
type updateProfileCall struct {
	ID       int64
	Nickname string
	Avatar   string
}
type updateRunnerStatusCall struct {
	ID           int64
	UserType     int8
	RunnerStatus int8
}
type setBlacklistCall struct {
	ID            int64
	IsBlacklisted int8
}
type updateOpenIDCall struct {
	ID     int64
	OpenID string
}

func (m *mockUserRepo) FindByPhone(ctx context.Context, db repository.DBTX, phone string) (*models.User, error) {
	if m.FindByPhoneFn != nil {
		return m.FindByPhoneFn(ctx, db, phone)
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserRepo) FindByID(ctx context.Context, db repository.DBTX, id int64) (*models.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, db, id)
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserRepo) FindByOpenID(ctx context.Context, db repository.DBTX, openid string) (*models.User, error) {
	if m.FindByOpenIDFn != nil {
		return m.FindByOpenIDFn(ctx, db, openid)
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserRepo) Create(ctx context.Context, db repository.DBTX, phone, openid string) (int64, error) {
	m.CreateCalls = append(m.CreateCalls, createCall{Phone: phone, OpenID: openid})
	if m.CreateFn != nil {
		return m.CreateFn(ctx, db, phone, openid)
	}
	return 1, nil
}

func (m *mockUserRepo) UpdateProfile(ctx context.Context, db repository.DBTX, id int64, nickname, avatar string) error {
	m.UpdateProfileCalls = append(m.UpdateProfileCalls, updateProfileCall{ID: id, Nickname: nickname, Avatar: avatar})
	if m.UpdateProfileFn != nil {
		return m.UpdateProfileFn(ctx, db, id, nickname, avatar)
	}
	return nil
}

func (m *mockUserRepo) UpdateRunnerStatus(ctx context.Context, db repository.DBTX, id int64, userType, runnerStatus int8) error {
	m.UpdateRunnerStatusCalls = append(m.UpdateRunnerStatusCalls,
		updateRunnerStatusCall{ID: id, UserType: userType, RunnerStatus: runnerStatus})
	if m.UpdateRunnerStatusFn != nil {
		return m.UpdateRunnerStatusFn(ctx, db, id, userType, runnerStatus)
	}
	return nil
}

func (m *mockUserRepo) SetBlacklist(ctx context.Context, db repository.DBTX, id int64, isBlacklisted int8) error {
	m.SetBlacklistCalls = append(m.SetBlacklistCalls, setBlacklistCall{ID: id, IsBlacklisted: isBlacklisted})
	if m.SetBlacklistFn != nil {
		return m.SetBlacklistFn(ctx, db, id, isBlacklisted)
	}
	return nil
}

func (m *mockUserRepo) UpdateOpenID(ctx context.Context, db repository.DBTX, id int64, openid string) error {
	m.UpdateOpenIDCalls = append(m.UpdateOpenIDCalls, updateOpenIDCall{ID: id, OpenID: openid})
	if m.UpdateOpenIDFn != nil {
		return m.UpdateOpenIDFn(ctx, db, id, openid)
	}
	return nil
}

func (m *mockUserRepo) SaveAvatar(ctx context.Context, db repository.DBTX, id int64, data []byte, contentType string) error {
	if m.SaveAvatarFn != nil {
		return m.SaveAvatarFn(ctx, db, id, data, contentType)
	}
	return nil
}

func (m *mockUserRepo) CreateAdmin(ctx context.Context, db repository.DBTX, phone, nickname string, userType int8) (int64, error) {
	if m.CreateAdminFn != nil {
		return m.CreateAdminFn(ctx, db, phone, nickname, userType)
	}
	return 2, nil
}

func (m *mockUserRepo) ListUsers(ctx context.Context, db repository.DBTX, filter repository.UserListFilter) ([]*models.User, int64, error) {
	if m.ListUsersFn != nil {
		return m.ListUsersFn(ctx, db, filter)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, db repository.DBTX, id int64, cols []string, args []any) error {
	if m.UpdateUserFn != nil {
		return m.UpdateUserFn(ctx, db, id, cols, args)
	}
	return nil
}

func (m *mockUserRepo) DeleteUser(ctx context.Context, db repository.DBTX, id int64) error {
	if m.DeleteUserFn != nil {
		return m.DeleteUserFn(ctx, db, id)
	}
	return nil
}

// --- mockAdminRepo ---

type mockAdminRepo struct {
	FindByUsernameFn  func(ctx context.Context, db repository.DBTX, username string) (*models.Admin, error)
	UpdateLastLoginFn func(ctx context.Context, db repository.DBTX, id int64, lastLogin any) error

	UpdateLastLoginCalls []int64
}

func (m *mockAdminRepo) FindByUsername(ctx context.Context, db repository.DBTX, username string) (*models.Admin, error) {
	if m.FindByUsernameFn != nil {
		return m.FindByUsernameFn(ctx, db, username)
	}
	return nil, sql.ErrNoRows
}

func (m *mockAdminRepo) UpdateLastLogin(ctx context.Context, db repository.DBTX, id int64, lastLogin any) error {
	m.UpdateLastLoginCalls = append(m.UpdateLastLoginCalls, id)
	if m.UpdateLastLoginFn != nil {
		return m.UpdateLastLoginFn(ctx, db, id, lastLogin)
	}
	return nil
}

// --- mockRunnerAppRepo ---

type mockRunnerAppRepo struct {
	CreateFn       func(ctx context.Context, db repository.DBTX, app *models.RunnerApplication) (int64, error)
	FindByIDFn     func(ctx context.Context, db repository.DBTX, id int64) (*models.RunnerApplication, error)
	ListByFilterFn func(ctx context.Context, db repository.DBTX, filter repository.RunnerAppFilter) ([]*models.RunnerApplication, int64, error)
	UpdateStatusFn func(ctx context.Context, db repository.DBTX, id, auditAdminID int64, status int8, auditRemark string) error

	CreateCalls       []*models.RunnerApplication
	UpdateStatusCalls []updateStatusCall
}

type updateStatusCall struct {
	ID           int64
	AuditAdminID int64
	Status       int8
	AuditRemark  string
}

func (m *mockRunnerAppRepo) Create(ctx context.Context, db repository.DBTX, app *models.RunnerApplication) (int64, error) {
	m.CreateCalls = append(m.CreateCalls, app)
	if m.CreateFn != nil {
		return m.CreateFn(ctx, db, app)
	}
	return 1, nil
}

func (m *mockRunnerAppRepo) FindByID(ctx context.Context, db repository.DBTX, id int64) (*models.RunnerApplication, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, db, id)
	}
	return nil, sql.ErrNoRows
}

func (m *mockRunnerAppRepo) ListByFilter(ctx context.Context, db repository.DBTX, filter repository.RunnerAppFilter) ([]*models.RunnerApplication, int64, error) {
	if m.ListByFilterFn != nil {
		return m.ListByFilterFn(ctx, db, filter)
	}
	return nil, 0, nil
}

func (m *mockRunnerAppRepo) UpdateStatus(ctx context.Context, db repository.DBTX, id, auditAdminID int64, status int8, auditRemark string) error {
	m.UpdateStatusCalls = append(m.UpdateStatusCalls, updateStatusCall{
		ID: id, AuditAdminID: auditAdminID, Status: status, AuditRemark: auditRemark,
	})
	if m.UpdateStatusFn != nil {
		return m.UpdateStatusFn(ctx, db, id, auditAdminID, status, auditRemark)
	}
	return nil
}

// --- mockSMSCodeCache ---

type mockSMSCodeCache struct {
	mu sync.Mutex
	// codes holds phone → code. Empty string means "deleted".
	codes map[string]string
	// rate counters
	phoneRate map[string]int
	ipRate    map[string]int

	// Function overrides (when set, take precedence over the in-memory defaults).
	SetCodeFn               func(ctx context.Context, phone, code string, ttl time.Duration) error
	GetCodeFn               func(ctx context.Context, phone string) (string, error)
	DelCodeFn               func(ctx context.Context, phone string) error
	CheckAndIncrPhoneRateFn func(ctx context.Context, phone string) (int, error)
	CheckAndIncrIPRateFn    func(ctx context.Context, ip string) (int, error)
	ListCodesFn             func(ctx context.Context) ([]repository.ActiveCode, error)

	DelCodeCalls []string
}

func newMockSMSCodeCache() *mockSMSCodeCache {
	return &mockSMSCodeCache{
		codes:     map[string]string{},
		phoneRate: map[string]int{},
		ipRate:    map[string]int{},
	}
}

func (m *mockSMSCodeCache) SetCode(ctx context.Context, phone, code string, ttl time.Duration) error {
	if m.SetCodeFn != nil {
		return m.SetCodeFn(ctx, phone, code, ttl)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[phone] = code
	return nil
}

func (m *mockSMSCodeCache) GetCode(ctx context.Context, phone string) (string, error) {
	if m.GetCodeFn != nil {
		return m.GetCodeFn(ctx, phone)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	code, ok := m.codes[phone]
	if !ok || code == "" {
		return "", redis.Nil
	}
	return code, nil
}

func (m *mockSMSCodeCache) DelCode(ctx context.Context, phone string) error {
	m.DelCodeCalls = append(m.DelCodeCalls, phone)
	if m.DelCodeFn != nil {
		return m.DelCodeFn(ctx, phone)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.codes, phone)
	return nil
}

func (m *mockSMSCodeCache) CheckAndIncrPhoneRate(ctx context.Context, phone string) (int, error) {
	if m.CheckAndIncrPhoneRateFn != nil {
		return m.CheckAndIncrPhoneRateFn(ctx, phone)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.phoneRate[phone]++
	return m.phoneRate[phone], nil
}

func (m *mockSMSCodeCache) CheckAndIncrIPRate(ctx context.Context, ip string) (int, error) {
	if m.CheckAndIncrIPRateFn != nil {
		return m.CheckAndIncrIPRateFn(ctx, ip)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ipRate[ip]++
	return m.ipRate[ip], nil
}

func (m *mockSMSCodeCache) ListCodes(ctx context.Context) ([]repository.ActiveCode, error) {
	if m.ListCodesFn != nil {
		return m.ListCodesFn(ctx)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]repository.ActiveCode, 0, len(m.codes))
	for phone, code := range m.codes {
		if code == "" {
			continue
		}
		out = append(out, repository.ActiveCode{
			Phone:    phone,
			Code:     code,
			ExpireIn: 300,
		})
	}
	return out, nil
}

// --- mockSMSProvider ---

type mockSMSProvider struct {
	SendFn         func(ctx context.Context, phone, code string) error
	GenerateCodeFn func() string

	SendCalls []sendCall
}

type sendCall struct{ Phone, Code string }

func (m *mockSMSProvider) Send(ctx context.Context, phone, code string) error {
	m.SendCalls = append(m.SendCalls, sendCall{Phone: phone, Code: code})
	if m.SendFn != nil {
		return m.SendFn(ctx, phone, code)
	}
	return nil
}

func (m *mockSMSProvider) GenerateCode() string {
	if m.GenerateCodeFn != nil {
		return m.GenerateCodeFn()
	}
	return "123456"
}

// Compile-time assertions that mocks satisfy the interfaces.
var (
	_ repository.UserRepo      = (*mockUserRepo)(nil)
	_ repository.AdminRepo     = (*mockAdminRepo)(nil)
	_ repository.RunnerAppRepo = (*mockRunnerAppRepo)(nil)
	_ repository.SMSCodeCache  = (*mockSMSCodeCache)(nil)
	_ SMSProvider              = (*mockSMSProvider)(nil)
)
