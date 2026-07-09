// Package service implements the business logic for the User module
// (auth + user management) and, in later phases, parcel / pickup / etc.
package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"pickup-helper/internal/config"
	apperrors "pickup-helper/internal/errors"
	"pickup-helper/internal/middleware"
	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

// smsCodeTTL is how long a verification code remains valid.
const smsCodeTTL = 300 * time.Second

// maxIPRate is the per-IP SMS request ceiling within smsIPRateTTL.
const maxIPRate = 10

// LoginResult is the unified response for user / admin login.
type LoginResult struct {
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token,omitempty"`
	ExpiresIn    int                 `json:"expires_in"` // access-token TTL in seconds
	User         *models.UserInfoDTO `json:"user,omitempty"`
	Role         string              `json:"role"` // "user" or "admin"
}

// RefreshResult is the response for token refresh.
type RefreshResult struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// AuthService handles authentication: SMS send, login/registration,
// token refresh, and admin login.
type AuthService struct {
	userRepo  repository.UserRepo
	adminRepo repository.AdminRepo
	smsCache  repository.SMSCodeCache
	sms       SMSProvider
	cfg       *config.Config
	db        *sqlx.DB
}

// NewAuthService wires up an AuthService. db is used for the login
// registration flow (INSERT IGNORE + FindByPhone + optional UpdateOpenID).
func NewAuthService(
	ur repository.UserRepo,
	ar repository.AdminRepo,
	sc repository.SMSCodeCache,
	sms SMSProvider,
	cfg *config.Config,
	db *sqlx.DB,
) *AuthService {
	return &AuthService{
		userRepo:  ur,
		adminRepo: ar,
		smsCache:  sc,
		sms:       sms,
		cfg:       cfg,
		db:        db,
	}
}

// SendCode validates the phone, enforces per-phone (60s) and per-IP (10/600s)
// rate limits, then generates and stores a 6-digit code via the SMS cache
// and dispatches it through the SMS provider. Returns the expiry in seconds.
func (s *AuthService) SendCode(ctx context.Context, phone, ip string) (int, error) {
	if !models.IsValidPhone(phone) {
		return 0, apperrors.New(apperrors.ErrPhoneFormat, "")
	}

	// Reject already-blacklisted phones (registered users only).
	if u, err := s.userRepo.FindByPhone(ctx, s.db, phone); err == nil && u.IsBlacklistedBool() {
		return 0, apperrors.New(apperrors.ErrPhoneBlacklisted, "")
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	// Per-phone 60s rate limit: first call returns 1; >1 means too frequent.
	phoneCnt, err := s.smsCache.CheckAndIncrPhoneRate(ctx, phone)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if phoneCnt > 1 {
		return 0, apperrors.New(apperrors.ErrSMSTooFrequent, "")
	}

	// Per-IP 10/600s rate limit.
	ipCnt, err := s.smsCache.CheckAndIncrIPRate(ctx, ip)
	if err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if ipCnt > maxIPRate {
		return 0, apperrors.New(apperrors.ErrSMSTooFrequent, "")
	}

	code := s.sms.GenerateCode()
	if err := s.smsCache.SetCode(ctx, phone, code, smsCodeTTL); err != nil {
		return 0, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if err := s.sms.Send(ctx, phone, code); err != nil {
		return 0, apperrors.New(apperrors.ErrSMSChannelFail, "")
	}
	return int(smsCodeTTL.Seconds()), nil
}

// Login verifies the SMS code (one-shot consumption), registers the user
// if new (INSERT IGNORE), associates openid if supplied, and issues JWTs.
// Blacklisted users are rejected with ErrUserBlacklisted.
func (s *AuthService) Login(ctx context.Context, phone, code, openid string) (*LoginResult, error) {
	if !models.IsValidPhone(phone) {
		return nil, apperrors.New(apperrors.ErrPhoneFormat, "")
	}

	stored, err := s.smsCache.GetCode(ctx, phone)
	if errors.Is(err, redis.Nil) {
		return nil, apperrors.New(apperrors.ErrCodeInvalid, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if stored != code {
		return nil, apperrors.New(apperrors.ErrCodeInvalid, "")
	}
	// One-shot consumption — even if subsequent steps fail, the code is burned.
	if err := s.smsCache.DelCode(ctx, phone); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	if _, err := s.userRepo.Create(ctx, s.db, phone, openid); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	user, err := s.userRepo.FindByPhone(ctx, s.db, phone)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if user.IsBlacklistedBool() {
		return nil, apperrors.New(apperrors.ErrUserBlacklisted, "")
	}

	// Associate openid if provided and not yet bound.
	if openid != "" && !user.OpenID.Valid {
		if err := s.userRepo.UpdateOpenID(ctx, s.db, user.ID, openid); err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
		}
		user.OpenID = sql.NullString{String: openid, Valid: true}
	}

	claims := middleware.Claims{
		UserID:   user.ID,
		UserType: int(user.UserType),
		Role:     "user",
	}
	access, err := middleware.SignAccess(s.cfg, claims)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "sign access token")
	}
	refresh, err := middleware.SignRefresh(s.cfg, claims)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "sign refresh token")
	}
	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.cfg.JWT.AccessTTL.Seconds()),
		User:         user.ToDTO(),
		Role:         "user",
	}, nil
}

// RefreshToken validates a refresh token and issues a new access token
// with the same subject claims. Expired refresh tokens yield ErrRefreshExpired;
// any other parse failure yields ErrRefreshInvalid.
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	claims, err := middleware.ParseRefresh(s.cfg, refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.New(apperrors.ErrRefreshExpired, "")
		}
		return nil, apperrors.New(apperrors.ErrRefreshInvalid, "")
	}
	if claims.UserID <= 0 {
		return nil, apperrors.New(apperrors.ErrRefreshInvalid, "")
	}

	// Re-sign access token with refreshed expiry.
	access, err := middleware.SignAccess(s.cfg, middleware.Claims{
		UserID:   claims.UserID,
		UserType: claims.UserType,
		Role:     firstNonEmpty(claims.Role, "user"),
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "sign access token")
	}
	return &RefreshResult{
		AccessToken: access,
		ExpiresIn:   int(s.cfg.JWT.AccessTTL.Seconds()),
	}, nil
}

// AdminLogin authenticates an admin by username + password (bcrypt) and
// issues a JWT with role="admin" and the admin's station_id (if any).
// Disabled accounts (status=0) are rejected with ErrAdminDisabled.
func (s *AuthService) AdminLogin(ctx context.Context, username, password string) (*LoginResult, error) {
	admin, err := s.adminRepo.FindByUsername(ctx, s.db, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrAdminCredential, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, apperrors.New(apperrors.ErrAdminCredential, "")
	}
	if admin.Status != models.AdminStatusEnabled {
		return nil, apperrors.New(apperrors.ErrAdminDisabled, "")
	}

	if err := s.adminRepo.UpdateLastLogin(ctx, s.db, admin.ID, time.Now()); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}

	claims := middleware.Claims{
		UserID: admin.ID,
		Role:   "admin",
	}
	if admin.StationID.Valid {
		claims.StationID = admin.StationID.Int64
	}
	access, err := middleware.SignAccess(s.cfg, claims)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "sign access token")
	}
	refresh, err := middleware.SignRefresh(s.cfg, claims)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "sign refresh token")
	}
	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.cfg.JWT.AccessTTL.Seconds()),
		Role:         "admin",
	}, nil
}

// firstNonEmpty returns s if non-empty, else fallback.
func firstNonEmpty(s, fallback string) string {
	if s != "" {
		return s
	}
	return fallback
}

// LoginByOpenID looks up a user by openid and returns JWT tokens.
func (s *AuthService) LoginByOpenID(ctx context.Context, openid string) (*LoginResult, error) {
	u, err := s.userRepo.FindByOpenID(ctx, s.db, openid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.New(apperrors.ErrUserNotFound, "")
	}
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if u.IsBlacklistedBool() {
		return nil, apperrors.New(apperrors.ErrUserBlacklisted, "")
	}
	claims := middleware.Claims{
		UserID:   u.ID,
		UserType: int(u.UserType),
		Role:     "user",
	}
	access, err := middleware.SignAccess(s.cfg, claims)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	refresh, err := middleware.SignRefresh(s.cfg, claims)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.cfg.JWT.AccessTTL.Seconds()),
		Role:         "user",
		User:         u.ToDTO(),
	}, nil
}

// RegisterByWechat creates a new user from WeChat login data.
func (s *AuthService) RegisterByWechat(ctx context.Context, openid, phone, nickname, avatarURL string) (*LoginResult, error) {
	if openid == "" || phone == "" {
		return nil, apperrors.New(apperrors.ErrInvalidParam, "openid 和 phone 必填")
	}
	if !models.IsValidPhone(phone) {
		return nil, apperrors.New(apperrors.ErrPhoneFormat, "")
	}

	// Check if phone already exists → bind openid to existing user.
	u, err := s.userRepo.FindByPhone(ctx, s.db, phone)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if u != nil {
		if u.IsBlacklistedBool() {
			return nil, apperrors.New(apperrors.ErrUserBlacklisted, "")
		}
		if !u.OpenID.Valid {
			if e := s.userRepo.UpdateOpenID(ctx, s.db, u.ID, openid); e != nil {
				return nil, apperrors.Wrap(e, apperrors.ErrInternal, "")
			}
			u.OpenID = sql.NullString{String: openid, Valid: true}
		}
		if nickname != "" && u.Nickname == "" {
			_ = s.userRepo.UpdateProfile(ctx, s.db, u.ID, nickname, avatarURL)
			u.Nickname = nickname
		}
		return s.signLoginUser(u), nil
	}

	// Create new user.
	id, err := s.userRepo.Create(ctx, s.db, phone, openid)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	if nickname != "" {
		_ = s.userRepo.UpdateProfile(ctx, s.db, id, nickname, avatarURL)
	}

	u2, err := s.userRepo.FindByID(ctx, s.db, id)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrInternal, "")
	}
	return s.signLoginUser(u2), nil
}

func (s *AuthService) signLoginUser(u *models.User) *LoginResult {
	claims := middleware.Claims{
		UserID:   u.ID,
		UserType: int(u.UserType),
		Role:     "user",
	}
	access, _ := middleware.SignAccess(s.cfg, claims)
	refresh, _ := middleware.SignRefresh(s.cfg, claims)
	return &LoginResult{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.cfg.JWT.AccessTTL.Seconds()),
		Role:         "user",
		User:         u.ToDTO(),
	}
}
