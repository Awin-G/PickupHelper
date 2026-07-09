//go:build integration

package test

import (
	"database/sql"
	"testing"
	"time"

	"pickup-helper/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// SeedStation inserts a test station and returns its id.
func SeedStation(t *testing.T, db *sqlx.DB) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO stations (name, address, status) VALUES (?, ?, ?)`,
		"测试驿站", "北京市海淀区中关村大街1号", 1,
	)
	require.NoError(t, err, "seed station")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return id
}

// SeedUser inserts a test user with the given phone and returns its id.
func SeedUser(t *testing.T, db *sqlx.DB, phone string) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO users (phone, nickname, user_type, runner_status, credit_score, is_blacklisted) VALUES (?, ?, ?, ?, ?, ?)`,
		phone, "测试用户", 1, 0, 100, 0,
	)
	require.NoError(t, err, "seed user")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return id
}

// SeedUserWithStatus inserts a user with explicit user_type / runner_status
// / is_blacklisted. Useful for testing the runner application flow and
// blacklist login rejection.
func SeedUserWithStatus(t *testing.T, db *sqlx.DB, phone string, userType, runnerStatus int8) *models.User {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO users (phone, nickname, user_type, runner_status, credit_score, is_blacklisted)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		phone, "测试用户", userType, runnerStatus, 100, 0,
	)
	require.NoError(t, err, "seed user with status")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return &models.User{
		ID: id, Phone: phone, Nickname: "测试用户", UserType: userType,
		RunnerStatus: runnerStatus, CreditScore: 100, IsBlacklisted: 0,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

// SeedBlackUser inserts a blacklisted user with the given phone.
func SeedBlackUser(t *testing.T, db *sqlx.DB, phone string) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO users (phone, nickname, user_type, runner_status, credit_score, is_blacklisted)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		phone, "黑名单用户", 1, 0, 100, 1,
	)
	require.NoError(t, err, "seed black user")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return id
}

// SeedAdmin inserts a test admin with the given username. The password is
// always "test-password-123" hashed with bcrypt cost 10.
func SeedAdmin(t *testing.T, db *sqlx.DB, username string) *models.Admin {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("test-password-123"), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt hash")
	res, err := db.Exec(
		`INSERT INTO admins (username, password_hash, role_id, status) VALUES (?, ?, ?, ?)`,
		username, string(hash), 1, 1,
	)
	require.NoError(t, err, "seed admin")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return &models.Admin{
		ID: id, Username: username, PasswordHash: string(hash),
		RoleID: 1, Status: 1,
	}
}

// SeedAdminWithPassword inserts a test admin with a custom plaintext password.
func SeedAdminWithPassword(t *testing.T, db *sqlx.DB, username, password string) *models.Admin {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt hash")
	res, err := db.Exec(
		`INSERT INTO admins (username, password_hash, role_id, status) VALUES (?, ?, ?, ?)`,
		username, string(hash), 1, 1,
	)
	require.NoError(t, err, "seed admin with password")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return &models.Admin{
		ID: id, Username: username, PasswordHash: string(hash),
		RoleID: 1, Status: 1,
	}
}

// SeedDisabledAdmin inserts an admin whose status=0 (disabled).
func SeedDisabledAdmin(t *testing.T, db *sqlx.DB, username string) *models.Admin {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("test-password-123"), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt hash")
	res, err := db.Exec(
		`INSERT INTO admins (username, password_hash, role_id, status) VALUES (?, ?, ?, ?)`,
		username, string(hash), 1, 0,
	)
	require.NoError(t, err, "seed disabled admin")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return &models.Admin{
		ID: id, Username: username, PasswordHash: string(hash),
		RoleID: 1, Status: 0,
	}
}

// SeedRunnerApp inserts a runner application and returns the persisted row.
func SeedRunnerApp(t *testing.T, db *sqlx.DB, userID int64, realName string, status int8) *models.RunnerApplication {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO runner_applications (user_id, real_name, student_id, id_card_image, status)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, realName,
		sql.NullString{String: "S001", Valid: true},
		sql.NullString{String: "https://cdn/id.jpg", Valid: true},
		status,
	)
	require.NoError(t, err, "seed runner app")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return &models.RunnerApplication{
		ID: id, UserID: userID, RealName: realName,
		StudentID:   sql.NullString{String: "S001", Valid: true},
		IDCardImage: sql.NullString{String: "https://cdn/id.jpg", Valid: true},
		Status:      status,
	}
}

// SeedShelf inserts a shelf_layout row for a station.
func SeedShelf(t *testing.T, db *sqlx.DB, stationID int64, shelfCode string, maxCapacity int) int64 {
	t.Helper()
	res, err := db.Exec(
		`INSERT INTO shelf_layout (station_id, shelf_code, row_num, col_num, current_capacity, max_capacity)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		stationID, shelfCode, 1, maxCapacity, 0, maxCapacity,
	)
	require.NoError(t, err, "seed shelf")
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return id
}

// TruncateAll clears all business tables (preserving goose_db_version)
// for test isolation.
func TruncateAll(t *testing.T, db *sqlx.DB) {
	t.Helper()
	tables := []string{
		"operation_logs", "notifications", "shelf_layout", "proxy_orders",
		"pickup_logs", "parcels", "runner_applications", "admins", "users",
		"stations", "courier_companies",
	}
	for _, tbl := range tables {
		_, err := db.Exec("DELETE FROM `" + tbl + "`")
		require.NoError(t, err, "truncate %s", tbl)
	}
}

// MaskPhoneEqual asserts that models.MaskPhone(phone) equals expected.
func MaskPhoneEqual(t *testing.T, phone, expected string) {
	t.Helper()
	require.Equal(t, expected, models.MaskPhone(phone),
		"MaskPhone(%q) should equal %q", phone, expected)
}
