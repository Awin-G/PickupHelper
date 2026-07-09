//go:build integration

package test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"pickup-helper/internal/models"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func userRepoSetup(t *testing.T) (context.Context, repository.DBTX, *repository.UserRepo, func()) {
	t.Helper()
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	r := repository.NewUserRepo()
	// DBTX is satisfied by *sqlx.DB.
	return context.Background(), env.DB, &r, func() {}
}

func TestUserRepo_Create_And_FindByPhone(t *testing.T) {
	ctx, db, r, _ := userRepoSetup(t)
	id, err := (*r).Create(ctx, db, "13800138000", "")
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	u, err := (*r).FindByPhone(ctx, db, "13800138000")
	require.NoError(t, err)
	assert.Equal(t, id, u.ID)
	assert.Equal(t, "13800138000", u.Phone)
	assert.Equal(t, int8(models.UserTypeNormal), u.UserType)
	assert.Equal(t, int8(models.RunnerStatusNone), u.RunnerStatus)
	assert.Equal(t, 100, u.CreditScore)
	assert.False(t, u.IsBlacklistedBool())
	assert.False(t, u.OpenID.Valid, "openid should be NULL for empty input")
}

func TestUserRepo_Create_DuplicatePhone(t *testing.T) {
	ctx, db, r, _ := userRepoSetup(t)
	id1, err := (*r).Create(ctx, db, "13900139000", "")
	require.NoError(t, err)
	// Second create with same phone should not error and return the same id.
	id2, err := (*r).Create(ctx, db, "13900139000", "")
	require.NoError(t, err)
	assert.Equal(t, id1, id2, "duplicate phone should return existing id")
}

func TestUserRepo_Create_WithOpenID(t *testing.T) {
	ctx, db, r, _ := userRepoSetup(t)
	id, err := (*r).Create(ctx, db, "13700137000", "wx_openid_abc")
	require.NoError(t, err)

	u, err := (*r).FindByOpenID(ctx, db, "wx_openid_abc")
	require.NoError(t, err)
	assert.Equal(t, id, u.ID)
	require.True(t, u.OpenID.Valid)
	assert.Equal(t, "wx_openid_abc", u.OpenID.String)
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	ctx, db, r, _ := userRepoSetup(t)
	_, err := (*r).FindByID(ctx, db, 99999)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestUserRepo_UpdateProfile(t *testing.T) {
	ctx, db, r, _ := userRepoSetup(t)
	id, _ := (*r).Create(ctx, db, "13600136000", "")

	require.NoError(t, (*r).UpdateProfile(ctx, db, id, "alice", "https://cdn/a.png"))

	u, err := (*r).FindByID(ctx, db, id)
	require.NoError(t, err)
	assert.Equal(t, "alice", u.Nickname)
	assert.Equal(t, "https://cdn/a.png", u.Avatar)
}

func TestUserRepo_UpdateRunnerStatus(t *testing.T) {
	ctx, db, r, _ := userRepoSetup(t)
	id, _ := (*r).Create(ctx, db, "13500135000", "")

	require.NoError(t, (*r).UpdateRunnerStatus(ctx, db, id, models.UserTypeRunner, models.RunnerStatusApproved))

	u, _ := (*r).FindByID(ctx, db, id)
	assert.Equal(t, int8(models.UserTypeRunner), u.UserType)
	assert.Equal(t, int8(models.RunnerStatusApproved), u.RunnerStatus)
}

func TestUserRepo_SetBlacklist(t *testing.T) {
	ctx, db, r, _ := userRepoSetup(t)
	id, _ := (*r).Create(ctx, db, "13400134000", "")

	require.NoError(t, (*r).SetBlacklist(ctx, db, id, 1))
	u, _ := (*r).FindByID(ctx, db, id)
	assert.True(t, u.IsBlacklistedBool())

	require.NoError(t, (*r).SetBlacklist(ctx, db, id, 0))
	u, _ = (*r).FindByID(ctx, db, id)
	assert.False(t, u.IsBlacklistedBool())
}

func TestUserRepo_UpdateOpenID(t *testing.T) {
	ctx, db, r, _ := userRepoSetup(t)
	id, _ := (*r).Create(ctx, db, "13300133000", "")

	require.NoError(t, (*r).UpdateOpenID(ctx, db, id, "wx_oid_xyz"))
	u, _ := (*r).FindByID(ctx, db, id)
	require.True(t, u.OpenID.Valid)
	assert.Equal(t, "wx_oid_xyz", u.OpenID.String)
}

func TestAdminRepo_FindByUsername(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewAdminRepo()
	admin := SeedAdmin(t, env.DB, "admin01")

	got, err := r.FindByUsername(ctx, env.DB, "admin01")
	require.NoError(t, err)
	assert.Equal(t, admin.ID, got.ID)
	assert.Equal(t, "admin01", got.Username)
	assert.NotEmpty(t, got.PasswordHash)
	assert.Equal(t, int8(models.AdminStatusEnabled), got.Status)
}

func TestAdminRepo_FindByUsername_NotFound(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewAdminRepo()

	_, err := r.FindByUsername(ctx, env.DB, "ghost")
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestAdminRepo_UpdateLastLogin(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewAdminRepo()
	admin := SeedAdmin(t, env.DB, "admin02")

	now := time.Now()
	require.NoError(t, r.UpdateLastLogin(ctx, env.DB, admin.ID, now))

	got, _ := r.FindByUsername(ctx, env.DB, "admin02")
	require.True(t, got.LastLogin.Valid)
	// Compare within a second to tolerate MySQL datetime precision.
	assert.WithinDuration(t, now, got.LastLogin.Time, 2*time.Second)
}

func TestRunnerAppRepo_Create_And_FindByID(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewRunnerAppRepo()
	uid := SeedUser(t, env.DB, "13800000001")

	app := &models.RunnerApplication{
		UserID:      uid,
		RealName:    "张三",
		StudentID:   sql.NullString{String: "S001", Valid: true},
		IDCardImage: sql.NullString{String: "https://cdn/id.jpg", Valid: true},
		Status:      models.AppStatusPending,
	}
	id, err := r.Create(ctx, env.DB, app)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	got, err := r.FindByID(ctx, env.DB, id)
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
	assert.Equal(t, uid, got.UserID)
	assert.Equal(t, "张三", got.RealName)
	require.True(t, got.StudentID.Valid)
	assert.Equal(t, "S001", got.StudentID.String)
	require.True(t, got.IDCardImage.Valid)
	assert.Equal(t, "https://cdn/id.jpg", got.IDCardImage.String)
	assert.Equal(t, int8(models.AppStatusPending), got.Status)
}

func TestRunnerAppRepo_FindByID_NotFound(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewRunnerAppRepo()

	_, err := r.FindByID(ctx, env.DB, 99999)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestRunnerAppRepo_ListByFilter_StatusFilter(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewRunnerAppRepo()

	uid := SeedUser(t, env.DB, "13800000002")
	SeedRunnerApp(t, env.DB, uid, "张三", models.AppStatusPending)
	SeedRunnerApp(t, env.DB, uid, "李四", models.AppStatusApproved)
	SeedRunnerApp(t, env.DB, uid, "王五", models.AppStatusRejected)

	pending := int8(models.AppStatusPending)
	apps, total, err := r.ListByFilter(ctx, env.DB, repository.RunnerAppFilter{
		Status: &pending, Limit: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, apps, 1)
	assert.Equal(t, "张三", apps[0].RealName)
}

func TestRunnerAppRepo_ListByFilter_Keyword(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewRunnerAppRepo()

	uid := SeedUser(t, env.DB, "13800000003")
	SeedRunnerApp(t, env.DB, uid, "张三", models.AppStatusPending)
	SeedRunnerApp(t, env.DB, uid, "李四", models.AppStatusPending)

	// Keyword matches real_name "张"
	apps, total, err := r.ListByFilter(ctx, env.DB, repository.RunnerAppFilter{
		Keyword: "张", Limit: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, apps, 1)
	assert.Equal(t, "张三", apps[0].RealName)
}

func TestRunnerAppRepo_ListByFilter_Keyword_MatchesPhone(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewRunnerAppRepo()

	uid := SeedUser(t, env.DB, "13811110000")
	SeedRunnerApp(t, env.DB, uid, "王五", models.AppStatusPending)

	// Keyword matches the user phone digits.
	apps, total, err := r.ListByFilter(ctx, env.DB, repository.RunnerAppFilter{
		Keyword: "1381111", Limit: 20,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, apps, 1)
	assert.Equal(t, "王五", apps[0].RealName)
}

func TestRunnerAppRepo_ListByFilter_Pagination(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewRunnerAppRepo()

	uid := SeedUser(t, env.DB, "13800000004")
	for i := 0; i < 5; i++ {
		SeedRunnerApp(t, env.DB, uid, "申请人", models.AppStatusPending)
	}

	apps, total, err := r.ListByFilter(ctx, env.DB, repository.RunnerAppFilter{
		Limit: 2, Offset: 0,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	require.Len(t, apps, 2, "page size 2")
}

func TestRunnerAppRepo_UpdateStatus(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewRunnerAppRepo()
	uid := SeedUser(t, env.DB, "13800000005")
	app := SeedRunnerApp(t, env.DB, uid, "赵六", models.AppStatusPending)

	require.NoError(t, r.UpdateStatus(ctx, env.DB, app.ID, 1001, models.AppStatusApproved, "ok"))

	got, _ := r.FindByID(ctx, env.DB, app.ID)
	assert.Equal(t, int8(models.AppStatusApproved), got.Status)
	require.True(t, got.AuditAdminID.Valid)
	assert.Equal(t, int64(1001), got.AuditAdminID.Int64)
	require.True(t, got.AuditRemark.Valid)
	assert.Equal(t, "ok", got.AuditRemark.String)
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewUserRepo()

	// Run a UserRepo.Create inside a transaction. *sqlx.Tx satisfies DBTX.
	err := repository.WithTx(ctx, env.DB, func(tx *sqlx.Tx) error {
		_, err := r.Create(ctx, tx, "13200132000", "")
		return err
	})
	require.NoError(t, err)

	u, err := r.FindByPhone(ctx, env.DB, "13200132000")
	require.NoError(t, err)
	assert.Equal(t, "13200132000", u.Phone)
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewUserRepo()

	// Force a failure mid-transaction; the user should NOT be persisted.
	boom := errors.New("boom")
	err := repository.WithTx(ctx, env.DB, func(tx *sqlx.Tx) error {
		if _, e := r.Create(ctx, tx, "13100131000", ""); e != nil {
			return e
		}
		return boom
	})
	assert.ErrorIs(t, err, boom)

	_, err = r.FindByPhone(ctx, env.DB, "13100131000")
	assert.True(t, errors.Is(err, sql.ErrNoRows), "rollback should prevent insert")
}

func TestWithTx_RollsBackOnPanic(t *testing.T) {
	env := SetupTestEnv(t)
	TruncateAll(t, env.DB)
	ctx := context.Background()
	r := repository.NewUserRepo()

	assert.Panics(t, func() {
		_ = repository.WithTx(ctx, env.DB, func(tx *sqlx.Tx) error {
			_, _ = r.Create(ctx, tx, "13000130000", "")
			panic("kaboom")
		})
	})

	_, err := r.FindByPhone(ctx, env.DB, "13000130000")
	assert.True(t, errors.Is(err, sql.ErrNoRows), "panic should trigger rollback")
}
