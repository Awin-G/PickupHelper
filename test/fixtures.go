//go:build integration

package test

import (
	"testing"

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

// SeedAdmin inserts a test admin with the given username. The password is
// always "test-password-123" hashed with bcrypt cost 10.
func SeedAdmin(t *testing.T, db *sqlx.DB, username string) int64 {
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
