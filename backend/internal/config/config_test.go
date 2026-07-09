package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chdirProjectRoot walks up from the test working directory until it finds
// `configs/config.dev.yaml`, then chdirs there. This makes Load() work from
// any subdirectory of the project.
func chdirProjectRoot(t *testing.T) {
	t.Helper()
	cwd, err := os.Getwd()
	require.NoError(t, err)
	dir := cwd
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "configs", "config.dev.yaml")); err == nil {
			require.NoError(t, os.Chdir(dir))
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("project root not found from %s", cwd)
}

func TestLoad_DefaultEnv(t *testing.T) {
	chdirProjectRoot(t)
	require.NoError(t, os.Unsetenv("APP_ENV"))
	// Clear env overrides that might leak from other tests.
	require.NoError(t, os.Unsetenv("MYSQL__HOST"))

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "pickup_helper", cfg.MySQL.Database)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestLoad_TestEnv(t *testing.T) {
	chdirProjectRoot(t)
	t.Setenv("APP_ENV", "test")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "pickup_helper_test", cfg.MySQL.Database)
	assert.Equal(t, 8081, cfg.Server.Port)
}

func TestLoad_EnvOverride(t *testing.T) {
	chdirProjectRoot(t)
	t.Setenv("APP_ENV", "dev")
	t.Setenv("MYSQL__HOST", "10.0.0.1")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", cfg.MySQL.Host)
}

func TestMySQLDSN(t *testing.T) {
	cfg := &Config{
		MySQL: MySQLConfig{
			Username: "root", Password: "1973", Host: "127.0.0.1", Port: 3306, Database: "x",
		},
	}
	dsn := cfg.MySQLDSN()
	assert.Contains(t, dsn, "parseTime=true")
	assert.Contains(t, dsn, "loc=Asia%2FShanghai")
	assert.True(t, strings.HasPrefix(dsn, "root:1973@tcp(127.0.0.1:3306)/x"))
}
