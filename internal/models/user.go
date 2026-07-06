// Package models defines domain structs that map 1:1 to database tables
// plus DTOs used by the service/handler layers. All business status
// constants live here to avoid import cycles between repository and service.
package models

import (
	"database/sql"
	"regexp"
	"time"
)

// User type and runner status constants (matching users.user_type / runner_status).
const (
	UserTypeNormal = 1 // 普通收件人
	UserTypeRunner = 2 // 跑腿员

	RunnerStatusNone     = 0 // 未申请
	RunnerStatusPending  = 1 // 审核中
	RunnerStatusApproved = 2 // 已通过
	RunnerStatusRejected = 3 // 已拒绝
)

// Admin status constants (admins.status).
const (
	AdminStatusDisabled = 0
	AdminStatusEnabled  = 1
)

// RunnerApplication status constants (runner_applications.status).
const (
	AppStatusPending  = 1
	AppStatusApproved = 2
	AppStatusRejected = 3
)

// User maps the `users` table.
type User struct {
	ID           int64          `db:"id" json:"id"`
	Phone        string         `db:"phone" json:"phone"`
	Nickname     string         `db:"nickname" json:"nickname"`
	Avatar       string         `db:"avatar" json:"avatar"`
	OpenID       sql.NullString `db:"openid" json:"openid,omitempty"`
	UserType     int8           `db:"user_type" json:"user_type"`
	RunnerStatus int8           `db:"runner_status" json:"runner_status"`
	CreditScore  int            `db:"credit_score" json:"credit_score"`
	IsBlacklisted int8          `db:"is_blacklisted" json:"is_blacklisted"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

// IsBlacklistedBool returns the blacklist flag as a bool.
func (u *User) IsBlacklistedBool() bool { return u.IsBlacklisted == 1 }

// Admin maps the `admins` table.
type Admin struct {
	ID           int64           `db:"id" json:"id"`
	Username     string          `db:"username" json:"username"`
	PasswordHash string          `db:"password_hash" json:"-"`
	RoleID       int64           `db:"role_id" json:"role_id"`
	StationID    sql.NullInt64   `db:"station_id" json:"station_id,omitempty"`
	RealName     sql.NullString  `db:"real_name" json:"real_name,omitempty"`
	Phone        sql.NullString  `db:"phone" json:"phone,omitempty"`
	Status       int8            `db:"status" json:"status"`
	LastLogin    sql.NullTime    `db:"last_login" json:"last_login,omitempty"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at" json:"updated_at"`
}

// RunnerApplication maps the `runner_applications` table.
type RunnerApplication struct {
	ID           int64          `db:"id" json:"id"`
	UserID       int64          `db:"user_id" json:"user_id"`
	RealName     string         `db:"real_name" json:"real_name"`
	StudentID    sql.NullString `db:"student_id" json:"student_id,omitempty"`
	IDCardImage  sql.NullString `db:"id_card_image" json:"id_card_image,omitempty"`
	Status       int8           `db:"status" json:"status"`
	AuditAdminID sql.NullInt64  `db:"audit_admin_id" json:"audit_admin_id,omitempty"`
	AuditRemark  sql.NullString `db:"audit_remark" json:"audit_remark,omitempty"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

// UserInfoDTO is the user-facing representation. Phone is masked and
// openid is never exposed.
type UserInfoDTO struct {
	ID            int64     `json:"id"`
	Phone         string    `json:"phone"`
	Nickname      string    `json:"nickname"`
	Avatar        string    `json:"avatar"`
	UserType      int8      `json:"user_type"`
	RunnerStatus  int8      `json:"runner_status"`
	CreditScore   int       `json:"credit_score"`
	IsBlacklisted bool      `json:"is_blacklisted"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToDTO converts a User to its masked DTO. Phone is always 11-digit masked;
// shorter strings are returned as-is.
func (u *User) ToDTO() *UserInfoDTO {
	return &UserInfoDTO{
		ID:            u.ID,
		Phone:         MaskPhone(u.Phone),
		Nickname:      u.Nickname,
		Avatar:        u.Avatar,
		UserType:      u.UserType,
		RunnerStatus:  u.RunnerStatus,
		CreditScore:   u.CreditScore,
		IsBlacklisted: u.IsBlacklistedBool(),
		CreatedAt:     u.CreatedAt,
	}
}

// phonePattern matches 11-digit Chinese mobile numbers (1[3-9]xxxxxxxxx).
var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

// MaskPhone masks the middle 4 digits of an 11-digit Chinese phone number,
// producing "138****8000". Strings of any other length are returned as-is.
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	if !phonePattern.MatchString(phone) {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}
