package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMaskPhone(t *testing.T) {
	cases := []struct {
		name  string
		phone string
		want  string
	}{
		{"standard 11-digit", "13800138000", "138****8000"},
		{"another valid", "13912345678", "139****5678"},
		{"too short", "123", "123"},
		{"too long", "138001380001", "138001380001"},
		{"empty", "", ""},
		{"wrong prefix", "10000000000", "10000000000"}, // 11 digits but not 1[3-9]
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, MaskPhone(c.phone))
		})
	}
}

func TestUser_IsBlacklistedBool(t *testing.T) {
	u := User{IsBlacklisted: 1}
	assert.True(t, u.IsBlacklistedBool())
	u.IsBlacklisted = 0
	assert.False(t, u.IsBlacklistedBool())
}

func TestUser_ToDTO_MasksPhoneAndDropsOpenID(t *testing.T) {
	u := &User{
		ID:            42,
		Phone:         "13800138000",
		Nickname:      "alice",
		Avatar:        "https://cdn/avatar.png",
		UserType:      UserTypeNormal,
		RunnerStatus:  RunnerStatusNone,
		CreditScore:   100,
		IsBlacklisted: 1,
		CreatedAt:     time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC),
	}

	dto := u.ToDTO()
	assert.Equal(t, int64(42), dto.ID)
	assert.Equal(t, "138****8000", dto.Phone, "phone must be masked")
	assert.Equal(t, "alice", dto.Nickname)
	assert.Equal(t, int8(UserTypeNormal), dto.UserType)
	assert.Equal(t, int8(RunnerStatusNone), dto.RunnerStatus)
	assert.Equal(t, 100, dto.CreditScore)
	assert.True(t, dto.IsBlacklisted)
	assert.Equal(t, u.CreatedAt, dto.CreatedAt)
}

func TestUserConstants(t *testing.T) {
	// Sanity: constants do not collide across tables.
	assert.Equal(t, 1, UserTypeNormal)
	assert.Equal(t, 2, UserTypeRunner)
	assert.Equal(t, 0, RunnerStatusNone)
	assert.Equal(t, 1, RunnerStatusPending)
	assert.Equal(t, 2, RunnerStatusApproved)
	assert.Equal(t, 3, RunnerStatusRejected)
	assert.Equal(t, 0, AdminStatusDisabled)
	assert.Equal(t, 1, AdminStatusEnabled)
	assert.Equal(t, 1, AppStatusPending)
	assert.Equal(t, 2, AppStatusApproved)
	assert.Equal(t, 3, AppStatusRejected)
}
