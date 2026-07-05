//go:build integration

package test

import (
	"context"
	"testing"

	"pickup-helper/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations_TestContainers(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()

	var count int
	err := env.DB.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'pickup_helper_test'`)
	require.NoError(t, err)
	// 11 business tables + goose_db_version = 12
	assert.Equal(t, 12, count, "expected 11 business tables + goose_db_version")

	var tableNames []string
	err = env.DB.SelectContext(ctx, &tableNames,
		`SELECT table_name FROM information_schema.tables WHERE table_schema = 'pickup_helper_test' ORDER BY table_name`)
	require.NoError(t, err)
	assert.Contains(t, tableNames, "parcels")
	assert.Contains(t, tableNames, "users")
	assert.Contains(t, tableNames, "courier_companies")
	assert.Contains(t, tableNames, "goose_db_version")

	_ = repository.RunMigrations
}

func TestRollbackMigrations_TestContainers(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()

	require.NoError(t, repository.RollbackMigrations(env.DB.DB, env.MigrationsDir()))

	var count int
	err := env.DB.GetContext(ctx, &count, `SELECT COUNT(*) FROM courier_companies`)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "seed rows should be rolled back")
}

func TestResetMigrations_TestContainers(t *testing.T) {
	env := SetupTestEnv(t)
	ctx := context.Background()

	require.NoError(t, repository.ResetMigrations(env.DB.DB, env.MigrationsDir()))

	var tableNames []string
	err := env.DB.SelectContext(ctx, &tableNames,
		`SELECT table_name FROM information_schema.tables WHERE table_schema = 'pickup_helper_test'`)
	require.NoError(t, err)
	for _, name := range tableNames {
		if name == "goose_db_version" {
			continue
		}
		t.Errorf("unexpected table after reset: %s", name)
	}
}
