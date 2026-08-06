package database

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Type       string
	SQLite     SQLiteConfig
	PostgreSQL PostgresConfig
}

type SQLiteConfig struct {
	Path string
}

type PostgresConfig struct {
	Host            string
	Port            int
	Database        string
	User            string
	Password        string
	SSLMode         string
	Schema          string
	Timezone        string
	ConnectTimeout  int
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func (c *PostgresConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Type: getEnv("DB_TYPE", "sqlite"),
	}

	cfg.SQLite = SQLiteConfig{
		Path: getEnv("SQLITE_PATH", "./lc_base/data.db"),
	}

	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		cfg.PostgreSQL = parseConnectionURL(databaseURL)
	} else {
		cfg.PostgreSQL = PostgresConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5432),
			Database:        getEnv("DB_NAME", "mycart"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        os.Getenv("DB_PASSWORD"),
			SSLMode:         getEnv("DB_SSLMODE", "require"),
			Schema:          getEnv("DB_SCHEMA", "public"),
			Timezone:        getEnv("DB_TIMEZONE", "UTC"),
			ConnectTimeout:  getEnvInt("DB_CONNECT_TIMEOUT", 10),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME", 300)) * time.Second,
		}
	}

	return cfg, nil
}

func parseConnectionURL(rawURL string) PostgresConfig {
	u, _ := url.Parse(rawURL)

	port := 5432
	if u.Port() != "" {
		port, _ = strconv.Atoi(u.Port())
	}

	password, _ := u.User.Password()

	sslmode := "require"
	if q := u.Query().Get("sslmode"); q != "" {
		sslmode = q
	}

	dbName := "postgres"
	if len(u.Path) > 1 {
		dbName = u.Path[1:]
	}

	return PostgresConfig{
		Host:            u.Hostname(),
		Port:            port,
		Database:        dbName,
		User:            u.User.Username(),
		Password:        password,
		SSLMode:         sslmode,
		Schema:          "public",
		Timezone:        "UTC",
		ConnectTimeout:  10,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 300 * time.Second,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
