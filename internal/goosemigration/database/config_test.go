package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_DefaultSQLite(t *testing.T) {
	// Clear environment
	os.Clearenv()

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "sqlite", cfg.Type)
	assert.Equal(t, "./lc_base/data.db", cfg.SQLite.Path)
}

func TestLoadConfig_PostgresFromDatabaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("DATABASE_URL", "postgresql://testuser:testpass@localhost:5432/testdb?sslmode=require")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "postgres", cfg.Type)
	assert.Equal(t, "testuser", cfg.PostgreSQL.User)
	assert.Equal(t, "testpass", cfg.PostgreSQL.Password)
	assert.Equal(t, "localhost", cfg.PostgreSQL.Host)
	assert.Equal(t, 5432, cfg.PostgreSQL.Port)
	assert.Equal(t, "testdb", cfg.PostgreSQL.Database)
	assert.Equal(t, "require", cfg.PostgreSQL.SSLMode)
}

func TestLoadConfig_PostgresFromIndividualVars(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_TYPE", "postgres")
	os.Setenv("DB_HOST", "pghost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_NAME", "mydb")
	os.Setenv("DB_USER", "pguser")
	os.Setenv("DB_PASSWORD", "pgpass")
	os.Setenv("DB_SSLMODE", "disable")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "postgres", cfg.Type)
	assert.Equal(t, "pguser", cfg.PostgreSQL.User)
	assert.Equal(t, "pgpass", cfg.PostgreSQL.Password)
	assert.Equal(t, "pghost", cfg.PostgreSQL.Host)
	assert.Equal(t, 5433, cfg.PostgreSQL.Port)
	assert.Equal(t, "mydb", cfg.PostgreSQL.Database)
	assert.Equal(t, "disable", cfg.PostgreSQL.SSLMode)
}

func TestPostgresConfig_ConnectionString(t *testing.T) {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "pass",
		Database: "testdb",
		SSLMode:  "require",
	}

	expected := "host=localhost port=5432 user=test password=pass dbname=testdb sslmode=require"
	assert.Equal(t, expected, cfg.ConnectionString())
}
