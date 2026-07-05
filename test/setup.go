//go:build integration

// Package test provides shared integration-test infrastructure based on
// testcontainers (MySQL 8 + Redis 7). All tests in this package require
// Docker to be available on the host.
package test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"pickup-helper/internal/config"
	"pickup-helper/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// TestEnv bundles all resources needed by an integration test.
type TestEnv struct {
	DB             *sqlx.DB
	Rdb            *redis.Client
	Cfg            *config.Config
	MysqlContainer *tcmysql.MySQLContainer
	RedisContainer *tcredis.RedisContainer
	migrationsDir  string
}

// SetupTestEnv starts fresh MySQL 8 + Redis 7 containers, runs migrations,
// and returns a fully wired TestEnv. t.Cleanup terminates containers and
// closes connections.
func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()
	ctx := context.Background()

	// Disable Ryuk reaper container (would require Docker Hub access).
	// We terminate containers explicitly in t.Cleanup.
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	migDir := resolveMigrationsDir(t)

	mysqlC, dsn := startMySQL(t, ctx)
	db, err := sqlx.Connect("mysql", dsn)
	require.NoError(t, err, "connect testcontainers mysql")

	require.NoError(t, repository.RunMigrations(db.DB, migDir), "run migrations")

	redisC, redisAddr := startRedis(t, ctx)
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	require.NoError(t, rdb.Ping(ctx).Err(), "ping testcontainers redis")

	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8081, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second},
		MySQL: config.MySQLConfig{
			Host: "127.0.0.1", Port: 3306, Username: "root", Password: "test",
			Database: "pickup_helper_test", MaxOpenConns: 10, MaxIdleConns: 5, ConnMaxLifetime: 60 * time.Second,
		},
		Redis: config.RedisConfig{
			Host: "127.0.0.1", Port: 6379, DB: 0, PoolSize: 10,
		},
	}

	env := &TestEnv{
		DB:             db,
		Rdb:            rdb,
		Cfg:            cfg,
		MysqlContainer: mysqlC,
		RedisContainer: redisC,
		migrationsDir:  migDir,
	}

	t.Cleanup(func() {
		if env.DB != nil {
			_ = env.DB.Close()
		}
		if env.Rdb != nil {
			_ = env.Rdb.Close()
		}
		if env.MysqlContainer != nil {
			_ = env.MysqlContainer.Terminate(ctx)
		}
		if env.RedisContainer != nil {
			_ = env.RedisContainer.Terminate(ctx)
		}
	})

	return env
}

// MigrationsDir returns the absolute path to the migrations directory.
func (e *TestEnv) MigrationsDir() string { return e.migrationsDir }

func startMySQL(t *testing.T, ctx context.Context) (*tcmysql.MySQLContainer, string) {
	t.Helper()
	// Use tc.WithImage to override the module's default "mysql:8.0.36" tag
	// with "mysql:8.0" (cached locally) to avoid Docker Hub network pulls.
	mysqlC, err := tcmysql.RunContainer(ctx,
		tc.WithImage("mysql:8.0"),
		tcmysql.WithDatabase("pickup_helper_test"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("test"),
	)
	if err != nil {
		mysqlC, err = tcmysql.RunContainer(ctx,
			tc.WithImage("mysql:8.0"),
			tcmysql.WithDatabase("pickup_helper_test"),
			tcmysql.WithUsername("root"),
			tcmysql.WithPassword("test"),
		)
	}
	require.NoError(t, err, "start mysql container")

	dsn, err := mysqlC.ConnectionString(ctx, "parseTime=true", "loc=Asia%2FShanghai", "charset=utf8mb4")
	require.NoError(t, err, "build mysql dsn")
	return mysqlC, dsn
}

func startRedis(t *testing.T, ctx context.Context) (*tcredis.RedisContainer, string) {
	t.Helper()
	redisC, err := tcredis.RunContainer(ctx, tc.WithImage("redis:7-alpine"))
	if err != nil {
		redisC, err = tcredis.RunContainer(ctx, tc.WithImage("redis:7-alpine"))
	}
	require.NoError(t, err, "start redis container")
	addr, err := redisC.Endpoint(ctx, "")
	require.NoError(t, err, "redis endpoint")
	return redisC, addr
}

// resolveMigrationsDir returns the absolute path to the project's migrations
// directory regardless of the test working directory.
func resolveMigrationsDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	// thisFile = .../PickupHelper/test/setup.go → project root is two dirs up.
	projectRoot := filepath.Dir(filepath.Dir(thisFile))
	migDir := filepath.Join(projectRoot, "migrations")
	st, err := os.Stat(migDir)
	require.NoError(t, err, "stat migrations dir %s", migDir)
	require.True(t, st.IsDir(), "%s is not a directory", migDir)
	return migDir
}

// MySQLContainerStop stops the MySQL container to simulate DB going down.
// Used by readiness-failure integration tests.
func (e *TestEnv) MySQLContainerStop(ctx context.Context) error {
	if e.MysqlContainer == nil {
		return fmt.Errorf("no mysql container")
	}
	timeout := 10 * time.Second
	return e.MysqlContainer.Stop(ctx, &timeout)
}

// RedisContainerStop stops the Redis container to simulate Redis going down.
func (e *TestEnv) RedisContainerStop(ctx context.Context) error {
	if e.RedisContainer == nil {
		return fmt.Errorf("no redis container")
	}
	timeout := 10 * time.Second
	return e.RedisContainer.Stop(ctx, &timeout)
}
