package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"pickup-helper/internal/config"
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// testCfg returns a Config with non-zero JWT secrets/TTLs for service tests.
func testCfg() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:  "test-access-secret",
			RefreshSecret: "test-refresh-secret",
			AccessTTL:     2 * time.Hour,
			RefreshTTL:    7 * 24 * time.Hour,
			Issuer:        "pickup-helper-test",
		},
	}
}

func TestSendCode_Success(t *testing.T) {
	ur := &mockUserRepo{FindByPhoneFn: func(_ context.Context, _ repository.DBTX, _ string) (*models.User, error) {
		return nil, sql.ErrNoRows
	}}
	cache := newMockSMSCodeCache()
	sms := &mockSMSProvider{}
	svc := NewAuthService(ur, &mockAdminRepo{}, cache, sms, testCfg(), nil)

	exp, err := svc.SendCode(context.Background(), "13800138000", "10.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, 300, exp)
	require.Len(t, sms.SendCalls, 1)
	assert.Equal(t, "13800138000", sms.SendCalls[0].Phone)
	assert.Equal(t, "123456", sms.SendCalls[0].Code)

	got, err := cache.GetCode(context.Background(), "13800138000")
	require.NoError(t, err)
	assert.Equal(t, "123456", got)
}

func TestSendCode_InvalidPhone(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{}, &mockAdminRepo{}, newMockSMSCodeCache(), &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.SendCode(context.Background(), "12345", "10.0.0.1")
	requireAppErr(t, err, apperrors.ErrPhoneFormat)
}

func TestSendCode_PhoneRateLimit(t *testing.T) {
	cache := newMockSMSCodeCache()
	cache.CheckAndIncrPhoneRateFn = func(_ context.Context, _ string) (int, error) {
		return 2, nil // second call within window
	}
	ur := &mockUserRepo{FindByPhoneFn: func(_ context.Context, _ repository.DBTX, _ string) (*models.User, error) {
		return nil, sql.ErrNoRows
	}}
	svc := NewAuthService(ur, &mockAdminRepo{}, cache, &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.SendCode(context.Background(), "13800138000", "10.0.0.1")
	requireAppErr(t, err, apperrors.ErrSMSTooFrequent)
}

func TestSendCode_IPRateLimit(t *testing.T) {
	cache := newMockSMSCodeCache()
	cache.CheckAndIncrIPRateFn = func(_ context.Context, _ string) (int, error) {
		return 11, nil // 11th from same IP
	}
	ur := &mockUserRepo{FindByPhoneFn: func(_ context.Context, _ repository.DBTX, _ string) (*models.User, error) {
		return nil, sql.ErrNoRows
	}}
	svc := NewAuthService(ur, &mockAdminRepo{}, cache, &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.SendCode(context.Background(), "13800138000", "10.0.0.1")
	requireAppErr(t, err, apperrors.ErrSMSTooFrequent)
}

func TestSendCode_Blacklisted(t *testing.T) {
	ur := &mockUserRepo{FindByPhoneFn: func(_ context.Context, _ repository.DBTX, _ string) (*models.User, error) {
		return &models.User{ID: 1, Phone: "13800138000", IsBlacklisted: 1}, nil
	}}
	svc := NewAuthService(ur, &mockAdminRepo{}, newMockSMSCodeCache(), &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.SendCode(context.Background(), "13800138000", "10.0.0.1")
	requireAppErr(t, err, apperrors.ErrPhoneBlacklisted)
}

func TestSendCode_SMSChannelFail(t *testing.T) {
	ur := &mockUserRepo{FindByPhoneFn: func(_ context.Context, _ repository.DBTX, _ string) (*models.User, error) {
		return nil, sql.ErrNoRows
	}}
	sms := &mockSMSProvider{SendFn: func(_ context.Context, _, _ string) error {
		return errors.New("sms gateway down")
	}}
	svc := NewAuthService(ur, &mockAdminRepo{}, newMockSMSCodeCache(), sms, testCfg(), nil)
	_, err := svc.SendCode(context.Background(), "13800138000", "10.0.0.1")
	requireAppErr(t, err, apperrors.ErrSMSChannelFail)
}

func TestLogin_Success_NewUser(t *testing.T) {
	cache := newMockSMSCodeCache()
	require.NoError(t, cache.SetCode(context.Background(), "13800138000", "123456", 60*time.Second))

	ur := &mockUserRepo{
		CreateFn: func(_ context.Context, _ repository.DBTX, _, _ string) (int64, error) {
			return 42, nil
		},
		FindByPhoneFn: func(_ context.Context, _ repository.DBTX, phone string) (*models.User, error) {
			return &models.User{ID: 42, Phone: phone, Nickname: "u", UserType: models.UserTypeNormal, CreditScore: 100}, nil
		},
	}
	svc := NewAuthService(ur, &mockAdminRepo{}, cache, &mockSMSProvider{}, testCfg(), nil)

	res, err := svc.Login(context.Background(), "13800138000", "123456", "")
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	assert.Equal(t, 7200, res.ExpiresIn)
	require.NotNil(t, res.User)
	assert.Equal(t, int64(42), res.User.ID)
	assert.Equal(t, "138****8000", res.User.Phone)
	assert.Equal(t, "user", res.Role)

	// Code should be consumed.
	_, err = cache.GetCode(context.Background(), "13800138000")
	assert.ErrorIs(t, err, redis.Nil)
	require.Len(t, cache.DelCodeCalls, 1)
}

func TestLogin_InvalidPhone(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{}, &mockAdminRepo{}, newMockSMSCodeCache(), &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.Login(context.Background(), "abc", "123456", "")
	requireAppErr(t, err, apperrors.ErrPhoneFormat)
}

func TestLogin_CodeInvalid_NotFound(t *testing.T) {
	cache := newMockSMSCodeCache() // empty
	svc := NewAuthService(&mockUserRepo{}, &mockAdminRepo{}, cache, &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.Login(context.Background(), "13800138000", "123456", "")
	requireAppErr(t, err, apperrors.ErrCodeInvalid)
}

func TestLogin_CodeInvalid_Mismatch(t *testing.T) {
	cache := newMockSMSCodeCache()
	require.NoError(t, cache.SetCode(context.Background(), "13800138000", "999999", 60*time.Second))
	svc := NewAuthService(&mockUserRepo{}, &mockAdminRepo{}, cache, &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.Login(context.Background(), "13800138000", "123456", "")
	requireAppErr(t, err, apperrors.ErrCodeInvalid)
}

func TestLogin_Blacklisted(t *testing.T) {
	cache := newMockSMSCodeCache()
	require.NoError(t, cache.SetCode(context.Background(), "13800138000", "123456", 60*time.Second))

	ur := &mockUserRepo{
		CreateFn: func(_ context.Context, _ repository.DBTX, _, _ string) (int64, error) { return 5, nil },
		FindByPhoneFn: func(_ context.Context, _ repository.DBTX, phone string) (*models.User, error) {
			return &models.User{ID: 5, Phone: phone, IsBlacklisted: 1}, nil
		},
	}
	svc := NewAuthService(ur, &mockAdminRepo{}, cache, &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.Login(context.Background(), "13800138000", "123456", "")
	requireAppErr(t, err, apperrors.ErrUserBlacklisted)
	// Code is still burned.
	_, err = cache.GetCode(context.Background(), "13800138000")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestLogin_OpenIDAssociate(t *testing.T) {
	cache := newMockSMSCodeCache()
	require.NoError(t, cache.SetCode(context.Background(), "13800138000", "123456", 60*time.Second))

	ur := &mockUserRepo{
		CreateFn: func(_ context.Context, _ repository.DBTX, phone, openid string) (int64, error) {
			assert.Equal(t, "wx_oid", openid)
			return 7, nil
		},
		FindByPhoneFn: func(_ context.Context, _ repository.DBTX, phone string) (*models.User, error) {
			return &models.User{ID: 7, Phone: phone, UserType: models.UserTypeNormal}, nil
		},
	}
	svc := NewAuthService(ur, &mockAdminRepo{}, cache, &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.Login(context.Background(), "13800138000", "123456", "wx_oid")
	require.NoError(t, err)
	require.Len(t, ur.UpdateOpenIDCalls, 1)
	assert.Equal(t, int64(7), ur.UpdateOpenIDCalls[0].ID)
	assert.Equal(t, "wx_oid", ur.UpdateOpenIDCalls[0].OpenID)
}

func TestLogin_OpenIDSkipped_WhenAlreadyBound(t *testing.T) {
	cache := newMockSMSCodeCache()
	require.NoError(t, cache.SetCode(context.Background(), "13800138000", "123456", 60*time.Second))

	ur := &mockUserRepo{
		FindByPhoneFn: func(_ context.Context, _ repository.DBTX, phone string) (*models.User, error) {
			return &models.User{
				ID:     7,
				Phone:  phone,
				OpenID: sql.NullString{String: "existing", Valid: true},
			}, nil
		},
	}
	svc := NewAuthService(ur, &mockAdminRepo{}, cache, &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.Login(context.Background(), "13800138000", "123456", "new_oid")
	require.NoError(t, err)
	assert.Empty(t, ur.UpdateOpenIDCalls, "OpenID already bound — no update")
}

func TestRefreshToken_Success(t *testing.T) {
	cfg := testCfg()
	refresh, err := middleware.SignRefresh(cfg, middleware.Claims{
		UserID: 42,
		Role:   "user",
	})
	require.NoError(t, err)

	svc := NewAuthService(&mockUserRepo{}, &mockAdminRepo{}, newMockSMSCodeCache(), &mockSMSProvider{}, cfg, nil)
	res, err := svc.RefreshToken(context.Background(), refresh)
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	assert.Equal(t, 7200, res.ExpiresIn)

	claims, err := middleware.ParseAccess(cfg, res.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, "user", claims.Role)
}

func TestRefreshToken_Invalid(t *testing.T) {
	svc := NewAuthService(&mockUserRepo{}, &mockAdminRepo{}, newMockSMSCodeCache(), &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.RefreshToken(context.Background(), "not-a-real-jwt")
	requireAppErr(t, err, apperrors.ErrRefreshInvalid)
}

func TestRefreshToken_Expired(t *testing.T) {
	cfg := testCfg()
	claims := middleware.Claims{
		UserID: 42,
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    cfg.JWT.Issuer,
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refresh, err := tok.SignedString([]byte(cfg.JWT.RefreshSecret))
	require.NoError(t, err)

	svc := NewAuthService(&mockUserRepo{}, &mockAdminRepo{}, newMockSMSCodeCache(), &mockSMSProvider{}, cfg, nil)
	_, err = svc.RefreshToken(context.Background(), refresh)
	requireAppErr(t, err, apperrors.ErrRefreshExpired)
}

func TestAdminLogin_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	require.NoError(t, err)

	ar := &mockAdminRepo{FindByUsernameFn: func(_ context.Context, _ repository.DBTX, _ string) (*models.Admin, error) {
		return &models.Admin{ID: 1, Username: "admin", PasswordHash: string(hash), Status: models.AdminStatusEnabled}, nil
	}}
	svc := NewAuthService(&mockUserRepo{}, ar, newMockSMSCodeCache(), &mockSMSProvider{}, testCfg(), nil)
	res, err := svc.AdminLogin(context.Background(), "admin", "secret123")
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	assert.Equal(t, "admin", res.Role)
	require.Len(t, ar.UpdateLastLoginCalls, 1)
	assert.Equal(t, int64(1), ar.UpdateLastLoginCalls[0])

	claims, err := middleware.ParseAccess(testCfg(), res.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, "admin", claims.Role)
}

func TestAdminLogin_WrongUsername(t *testing.T) {
	ar := &mockAdminRepo{} // returns sql.ErrNoRows by default
	svc := NewAuthService(&mockUserRepo{}, ar, newMockSMSCodeCache(), &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.AdminLogin(context.Background(), "ghost", "x")
	requireAppErr(t, err, apperrors.ErrAdminCredential)
}

func TestAdminLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	ar := &mockAdminRepo{FindByUsernameFn: func(_ context.Context, _ repository.DBTX, _ string) (*models.Admin, error) {
		return &models.Admin{ID: 1, Username: "admin", PasswordHash: string(hash), Status: models.AdminStatusEnabled}, nil
	}}
	svc := NewAuthService(&mockUserRepo{}, ar, newMockSMSCodeCache(), &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.AdminLogin(context.Background(), "admin", "wrong")
	requireAppErr(t, err, apperrors.ErrAdminCredential)
}

func TestAdminLogin_Disabled(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	ar := &mockAdminRepo{FindByUsernameFn: func(_ context.Context, _ repository.DBTX, _ string) (*models.Admin, error) {
		return &models.Admin{ID: 1, Username: "admin", PasswordHash: string(hash), Status: models.AdminStatusDisabled}, nil
	}}
	svc := NewAuthService(&mockUserRepo{}, ar, newMockSMSCodeCache(), &mockSMSProvider{}, testCfg(), nil)
	_, err := svc.AdminLogin(context.Background(), "admin", "secret123")
	requireAppErr(t, err, apperrors.ErrAdminDisabled)
}

// --- helpers ---

// requireAppErr asserts err is an *AppError with the given code.
func requireAppErr(t *testing.T, err error, code int) {
	t.Helper()
	require.Error(t, err)
	var ae *apperrors.AppError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, code, ae.Code, "expected code %d, got %d (msg=%q)", code, ae.Code, ae.Msg)
}
