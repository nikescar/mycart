# PostgreSQL Support Design

**Date:** 2026-08-07  
**Status:** Draft  
**Implementation Approach:** Bottom-Up (Database Layer First)

## Overview

Add PostgreSQL as an optional database backend for myCart alongside the existing SQLite support. Users will be able to choose their preferred database through environment variables, web installer, or CLI flags. The implementation uses sqlc for type-safe SQL code generation and maintains backward compatibility with existing deployments.

## Goals

1. **Optional PostgreSQL Support**: SQLite remains the default for simplicity
2. **Managed Database Compatibility**: Support Supabase, Heroku, Railway, Render, etc.
3. **Type Safety**: Use sqlc for compile-time SQL validation
4. **Bidirectional Migration**: Enable SQLite ↔ PostgreSQL data migration
5. **Zero Breaking Changes**: Existing SQLite users unaffected
6. **Docker & Kubernetes Ready**: Work seamlessly in containerized environments

## Non-Goals

- Replacing SQLite as the default database
- Supporting other databases (MySQL, MongoDB, etc.) in this phase
- Automatic database selection based on scale/load
- Real-time replication between SQLite and PostgreSQL

---

## 1. Architecture & Project Structure

### Current Structure
```
mycart/
├── migrations/               # SQLite migrations
│   └── *.sql
├── internal/
│   ├── queries/             # Direct SQL queries
│   │   ├── auth.go
│   │   ├── cart.go
│   │   ├── install.go
│   │   ├── pages.go
│   │   ├── products.go
│   │   ├── setting.go
│   │   └── queries.go       # Base struct
│   └── base/
│       └── base.go          # Database initialization
```

### New Structure
```
mycart/
├── db/
│   ├── migrations/
│   │   ├── sqlite/          # SQLite-specific migrations
│   │   │   └── *.sql
│   │   └── postgres/        # PostgreSQL-specific migrations
│   │       └── *.sql
│   └── queries/
│       ├── sqlite/          # SQLite sqlc queries
│       │   └── *.sql
│       └── postgres/        # PostgreSQL sqlc queries
│           └── *.sql
├── internal/
│   ├── db/
│   │   ├── interface.go     # Unified database interface
│   │   ├── types.go         # Common types
│   │   ├── sqlite/          # Generated sqlc code
│   │   │   └── *.go
│   │   └── postgres/        # Generated sqlc code
│   │       └── *.go
│   ├── database/
│   │   ├── database.go      # Connection factory
│   │   ├── config.go        # Configuration types
│   │   ├── connect.go       # Connection with retry
│   │   ├── migrate.go       # Migration logic
│   │   ├── detect.go        # Database detection
│   │   ├── health.go        # Health checking
│   │   ├── sqlite_adapter.go
│   │   ├── postgres_adapter.go
│   │   └── migrate_db.go    # Data migration tools
│   └── queries/             # Backward compatibility
│       └── queries.go       # Wraps internal/db interface
├── sqlc.yaml                # sqlc configuration
├── .env.example             # Environment variable template
└── docker/
    ├── .env.postgres.example
    ├── .env.supabase.example
    ├── docker-compose-postgres.yml
    └── docker-compose-supabase.yml
```

### Interface Design Pattern

The unified `Database` interface ensures both SQLite and PostgreSQL adapters implement the same methods:

```go
// internal/db/interface.go
type Database interface {
    // Transaction support
    BeginTx(ctx context.Context) (Database, error)
    Commit() error
    Rollback() error
    
    // Settings
    GetSettingByKey(ctx context.Context, key string) (map[string]Setting, error)
    UpdateSetting(ctx context.Context, key string, value interface{}) error
    ListSettings(ctx context.Context) ([]Setting, error)
    
    // Auth
    GetUserByEmail(ctx context.Context, email string) (User, error)
    CreateUser(ctx context.Context, email, password string) (User, error)
    
    // Install
    IsInstalled(ctx context.Context) (bool, error)
    
    // Pages, Products, Cart (all current methods)
    // ...
}
```

Existing `queries.DB()` returns this interface, so application code requires no changes.

---

## 2. Migration Structure & Strategy

### Migration File Organization

Each migration exists in **both** database directories with database-specific SQL syntax:

**SQLite** (`db/migrations/sqlite/20230714135923_init_db.sql`):
```sql
-- +goose Up
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE users;
```

**PostgreSQL** (`db/migrations/postgres/20230714135923_init_db.sql`):
```sql
-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- +goose Down
DROP TABLE users;
```

### Key SQL Differences

| Feature | SQLite | PostgreSQL |
|---------|--------|------------|
| Auto-increment | `AUTOINCREMENT` | `SERIAL` or `BIGSERIAL` |
| Timestamps | `DATETIME`, `CURRENT_TIMESTAMP` | `TIMESTAMPTZ`, `NOW()` |
| Booleans | `INTEGER` (0/1) | `BOOLEAN` |
| JSON | `TEXT` | `JSONB` |
| Case sensitivity | Case-insensitive `LIKE` | Case-sensitive, use `ILIKE` |

### Migration Execution

```go
// internal/database/migrate.go
func RunMigrations(cfg *Config, migrationsFS embed.FS) error {
    var migrationPath string
    var driver string
    var dsn string
    
    switch cfg.Type {
    case "postgres":
        migrationPath = "migrations/postgres"
        driver = "postgres"
        dsn = cfg.PostgreSQL.ConnectionString()
    case "sqlite":
        migrationPath = "migrations/sqlite"
        driver = "sqlite3"
        dsn = buildSQLiteDSN(cfg.SQLite.Path)
    }
    
    goose.SetBaseFS(migrationsFS)
    db, err := goose.OpenDBWithDriver(driver, dsn)
    if err != nil {
        return err
    }
    defer db.Close()
    
    goose.SetTableName("migrate_db_version")
    return goose.Up(db, migrationPath)
}
```

### Migration Conversion Process

When converting existing SQLite migrations:
1. Copy migration file structure (keep same timestamp)
2. Convert SQL syntax for PostgreSQL compatibility
3. Test both up and down migrations
4. Verify schema matches between databases

---

## 3. sqlc Configuration & Code Generation

### sqlc.yaml

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "db/migrations/postgres"
    queries: "db/queries/postgres"
    gen:
      go:
        package: "postgres"
        out: "internal/db/postgres"
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: false
        emit_exact_table_names: false
        emit_empty_slices: true

  - engine: "sqlite"
    schema: "db/migrations/sqlite"
    queries: "db/queries/sqlite"
    gen:
      go:
        package: "sqlite"
        out: "internal/db/sqlite"
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: false
        emit_exact_table_names: false
        emit_empty_slices: true
```

### Query Files

Both databases share similar query logic with minor syntax differences:

**PostgreSQL** (`db/queries/postgres/settings.sql`):
```sql
-- name: GetSettingByKey :one
SELECT key, value, created_at, updated_at
FROM settings
WHERE key = $1 LIMIT 1;

-- name: UpdateSetting :exec
UPDATE settings
SET value = $2, updated_at = NOW()
WHERE key = $1;

-- name: ListSettings :many
SELECT key, value, created_at, updated_at
FROM settings
ORDER BY key;
```

**SQLite** (`db/queries/sqlite/settings.sql`):
```sql
-- name: GetSettingByKey :one
SELECT key, value, created_at, updated_at
FROM settings
WHERE key = ? LIMIT 1;

-- name: UpdateSetting :exec
UPDATE settings
SET value = ?, updated_at = CURRENT_TIMESTAMP
WHERE key = ?;

-- name: ListSettings :many
SELECT key, value, created_at, updated_at
FROM settings
ORDER BY key;
```

### Generated Code

sqlc generates type-safe Go interfaces and implementations:

```go
// internal/db/postgres/querier.go (generated)
type Querier interface {
    GetSettingByKey(ctx context.Context, key string) (Setting, error)
    UpdateSetting(ctx context.Context, arg UpdateSettingParams) error
    ListSettings(ctx context.Context) ([]Setting, error)
}
```

### Development Workflow

```bash
# After modifying migrations or queries
sqlc generate

# Generates code in:
# - internal/db/postgres/*.go
# - internal/db/sqlite/*.go
```

---

## 4. Database Interface & Adapter Pattern

### Common Types

```go
// internal/db/types.go
package db

import "time"

type Setting struct {
    Key       string      `json:"key"`
    Value     interface{} `json:"value"`
    CreatedAt time.Time   `json:"created_at"`
    UpdatedAt time.Time   `json:"updated_at"`
}

type User struct {
    ID        int64     `json:"id"`
    Email     string    `json:"email"`
    Password  string    `json:"-"`
    CreatedAt time.Time `json:"created_at"`
}

type Page struct {
    ID        int64     `json:"id"`
    Slug      string    `json:"slug"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Published bool      `json:"published"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// ... other domain types
```

### PostgreSQL Adapter

```go
// internal/database/postgres_adapter.go
package database

import (
    "context"
    "database/sql"
    
    "github.com/shurco/mycart/internal/db"
    pgdb "github.com/shurco/mycart/internal/db/postgres"
)

type PostgresAdapter struct {
    db      *sql.DB
    queries *pgdb.Queries
    tx      *sql.Tx
}

func NewPostgresAdapter(sqlDB *sql.DB) db.Database {
    return &PostgresAdapter{
        db:      sqlDB,
        queries: pgdb.New(sqlDB),
    }
}

func (a *PostgresAdapter) GetSettingByKey(ctx context.Context, key string) (map[string]db.Setting, error) {
    row, err := a.queries.GetSettingByKey(ctx, key)
    if err != nil {
        return nil, err
    }
    
    return map[string]db.Setting{
        key: {
            Key:       row.Key,
            Value:     row.Value,
            CreatedAt: row.CreatedAt,
            UpdatedAt: row.UpdatedAt,
        },
    }, nil
}

func (a *PostgresAdapter) BeginTx(ctx context.Context) (db.Database, error) {
    tx, err := a.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    
    return &PostgresAdapter{
        db:      a.db,
        queries: pgdb.New(tx),
        tx:      tx,
    }, nil
}

func (a *PostgresAdapter) Commit() error {
    if a.tx == nil {
        return nil
    }
    return a.tx.Commit()
}

func (a *PostgresAdapter) Rollback() error {
    if a.tx == nil {
        return nil
    }
    return a.tx.Rollback()
}

// ... implement all interface methods
```

### SQLite Adapter

```go
// internal/database/sqlite_adapter.go
package database

import (
    "context"
    "database/sql"
    
    "github.com/shurco/mycart/internal/db"
    sqlitedb "github.com/shurco/mycart/internal/db/sqlite"
)

type SQLiteAdapter struct {
    db      *sql.DB
    queries *sqlitedb.Queries
    tx      *sql.Tx
}

func NewSQLiteAdapter(sqlDB *sql.DB) db.Database {
    return &SQLiteAdapter{
        db:      sqlDB,
        queries: sqlitedb.New(sqlDB),
    }
}

// Same interface implementation as PostgresAdapter
```

### Backward Compatibility

```go
// internal/queries/queries.go (updated)
package queries

import (
    "embed"
    
    "github.com/shurco/mycart/internal/db"
    "github.com/shurco/mycart/internal/database"
)

var dbInstance db.Database

// New initializes database (signature unchanged)
func New(migrationsFS embed.FS) error {
    cfg, err := database.LoadConfig()
    if err != nil {
        return err
    }
    
    dbConn, err := database.ConnectWithRetry(cfg)
    if err != nil {
        return err
    }
    
    if err := database.RunMigrations(cfg, migrationsFS); err != nil {
        return err
    }
    
    switch cfg.Type {
    case "postgres":
        dbInstance = database.NewPostgresAdapter(dbConn)
    case "sqlite":
        dbInstance = database.NewSQLiteAdapter(dbConn)
    }
    
    return nil
}

// DB returns the database interface (existing API preserved)
func DB() db.Database {
    return dbInstance
}
```

---

## 5. Configuration System

### Environment Variables

`.env.example`:
```bash
# Database Type (sqlite or postgres)
DB_TYPE=sqlite

# SQLite Configuration
SQLITE_PATH=./lc_base/data.db

# PostgreSQL Configuration
# Option 1: Connection string (recommended)
DATABASE_URL=postgresql://user:password@localhost:5432/mycart?sslmode=require

# Option 2: Individual parameters
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mycart
DB_USER=postgres
DB_PASSWORD=your_password
DB_SSLMODE=require

# PostgreSQL Advanced (optional)
DB_SCHEMA=public
DB_TIMEZONE=UTC
DB_CONNECT_TIMEOUT=10

# Connection Pooling (optional, defaults provided)
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=300
```

### Configuration Types

```go
// internal/database/config.go
package database

import (
    "fmt"
    "net/url"
    "os"
    "strconv"
    "time"
)

type Config struct {
    Type       string // "sqlite" or "postgres"
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
    
    return PostgresConfig{
        Host:     u.Hostname(),
        Port:     port,
        Database: u.Path[1:], // Remove leading /
        User:     u.User.Username(),
        Password: password,
        SSLMode:  sslmode,
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
```

### Database Detection

```go
// internal/database/detect.go
package database

import "os"

type DetectedDatabase struct {
    Type        string
    Source      string
    Description string
    Available   bool
}

func DetectDatabases() []DetectedDatabase {
    var detected []DetectedDatabase
    
    // PostgreSQL from environment
    if os.Getenv("DATABASE_URL") != "" || os.Getenv("DB_HOST") != "" {
        desc := "PostgreSQL (configured via environment variables)"
        if url := os.Getenv("DATABASE_URL"); url != "" {
            cfg := parseConnectionURL(url)
            desc = fmt.Sprintf("PostgreSQL (configured: %s@%s)", cfg.User, cfg.Host)
        }
        
        detected = append(detected, DetectedDatabase{
            Type:        "postgres",
            Source:      "environment",
            Description: desc,
            Available:   true,
        })
    }
    
    // Existing SQLite
    if fileExists("./lc_base/data.db") {
        detected = append(detected, DetectedDatabase{
            Type:        "sqlite",
            Source:      "existing",
            Description: "SQLite (existing database found)",
            Available:   true,
        })
    }
    
    // New SQLite
    detected = append(detected, DetectedDatabase{
        Type:        "sqlite",
        Source:      "default",
        Description: "SQLite (create new database)",
        Available:   true,
    })
    
    return detected
}
```

---

## 6. Installer Updates

### Web Installer Flow

New database selection step before admin account creation:

1. Detect available databases via `/api/install/detect-databases`
2. Display options with radio buttons
3. If PostgreSQL selected, show connection form or use detected config
4. Test connection via `/api/install/test-connection`
5. Save configuration to `.env` file
6. Run migrations
7. Create admin account
8. Mark as installed

### Web Installer API

```go
// internal/handlers/private/install.go (updated)

type InstallRequest struct {
    DatabaseType   string `json:"database_type"`
    DatabaseURL    string `json:"database_url,omitempty"`
    PostgresConfig *PostgresConfigRequest `json:"postgres_config,omitempty"`
    Email          string `json:"email"`
    Password       string `json:"password"`
    Domain         string `json:"domain"`
}

type PostgresConfigRequest struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Database string `json:"database"`
    User     string `json:"user"`
    Password string `json:"password"`
    SSLMode  string `json:"ssl_mode"`
}

// GET /api/install/detect-databases
func DetectDatabases(c fiber.Ctx) error {
    databases := database.DetectDatabases()
    return c.JSON(databases)
}

// POST /api/install/test-connection
func TestConnection(c fiber.Ctx) error {
    var req struct {
        DatabaseType string `json:"database_type"`
        DatabaseURL  string `json:"database_url"`
    }
    
    if err := c.Bind().JSON(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }
    
    cfg := &database.Config{Type: req.DatabaseType}
    if req.DatabaseType == "postgres" {
        cfg.PostgreSQL = database.ParseConnectionURL(req.DatabaseURL)
    }
    
    db, err := database.Connect(cfg)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "success": false,
            "error":   err.Error(),
        })
    }
    defer db.Close()
    
    return c.JSON(fiber.Map{"success": true})
}

// POST /api/install
func Install(c fiber.Ctx) error {
    var req InstallRequest
    if err := c.Bind().JSON(&req); err != nil {
        return err
    }
    
    // Build database config
    cfg := &database.Config{Type: req.DatabaseType}
    
    if req.DatabaseType == "postgres" {
        if req.DatabaseURL != "" {
            cfg.PostgreSQL = database.ParseConnectionURL(req.DatabaseURL)
        } else if req.PostgresConfig != nil {
            cfg.PostgreSQL = database.PostgresConfig{
                Host:     req.PostgresConfig.Host,
                Port:     req.PostgresConfig.Port,
                Database: req.PostgresConfig.Database,
                User:     req.PostgresConfig.User,
                Password: req.PostgresConfig.Password,
                SSLMode:  req.PostgresConfig.SSLMode,
            }
        }
    }
    
    // Write .env file
    if err := writeEnvFile(cfg); err != nil {
        return err
    }
    
    // Connect and run migrations
    // Create admin user
    // Mark as installed
    
    return c.JSON(fiber.Map{"success": true})
}
```

### CLI Installer

```go
// cmd/main.go (updated)
func cmdInstall() *cobra.Command {
    var (
        email       string
        password    string
        domain      string
        dbType      string
        databaseURL string
        dbHost      string
        dbPort      int
        dbName      string
        dbUser      string
        dbPassword  string
    )
    
    cmd := &cobra.Command{
        Use:   "install",
        Short: "Create the admin account (first-time setup)",
        Run: func(_ *cobra.Command, _ []string) {
            cfg := &database.Config{Type: dbType}
            
            if dbType == "postgres" {
                if databaseURL != "" {
                    cfg.PostgreSQL = database.ParseConnectionURL(databaseURL)
                } else {
                    cfg.PostgreSQL = database.PostgresConfig{
                        Host:     dbHost,
                        Port:     dbPort,
                        Database: dbName,
                        User:     dbUser,
                        Password: dbPassword,
                        SSLMode:  "require",
                    }
                }
            }
            
            handleCommandError(app.InstallWithDB(context.Background(), cfg, &models.Install{
                Email:    email,
                Password: password,
                Domain:   domain,
            }))
        },
    }
    
    cmd.Flags().StringVar(&email, "email", "", "admin email")
    cmd.Flags().StringVar(&password, "password", "", "admin password")
    cmd.Flags().StringVar(&domain, "domain", "localhost", "domain")
    cmd.Flags().StringVar(&dbType, "db-type", "sqlite", "database type (sqlite|postgres)")
    cmd.Flags().StringVar(&databaseURL, "database-url", "", "PostgreSQL connection string")
    cmd.Flags().StringVar(&dbHost, "db-host", "localhost", "PostgreSQL host")
    cmd.Flags().IntVar(&dbPort, "db-port", 5432, "PostgreSQL port")
    cmd.Flags().StringVar(&dbName, "db-name", "mycart", "PostgreSQL database")
    cmd.Flags().StringVar(&dbUser, "db-user", "postgres", "PostgreSQL user")
    cmd.Flags().StringVar(&dbPassword, "db-password", "", "PostgreSQL password")
    
    _ = cmd.MarkFlagRequired("email")
    _ = cmd.MarkFlagRequired("password")
    
    return cmd
}
```

---

## 7. Database Migration Commands

### Commands

```bash
# SQLite → PostgreSQL
./mycart migrate-to-postgres --database-url "postgresql://..."

# PostgreSQL → SQLite
./mycart migrate-to-sqlite --path "./lc_base/data.db"
```

### Implementation

```go
// cmd/main.go
func cmdMigrateToPostgres() *cobra.Command {
    var databaseURL string
    var dryRun bool
    
    cmd := &cobra.Command{
        Use:   "migrate-to-postgres",
        Short: "Migrate data from SQLite to PostgreSQL",
        Long: `Exports data from SQLite and imports into PostgreSQL.

Steps:
1. Verify PostgreSQL connection
2. Create PostgreSQL schema (run migrations)
3. Export data from SQLite
4. Import data into PostgreSQL
5. Verify data integrity
6. Update .env configuration`,
        Run: func(_ *cobra.Command, _ []string) {
            if dryRun {
                fmt.Println("DRY RUN MODE - No changes will be made")
            }
            handleCommandError(app.MigrateToPostgres(databaseURL, dryRun))
        },
    }
    
    cmd.Flags().StringVar(&databaseURL, "database-url", "", "PostgreSQL connection string (required)")
    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate migration without making changes")
    _ = cmd.MarkFlagRequired("database-url")
    
    return cmd
}

func cmdMigrateToSQLite() *cobra.Command {
    var sqlitePath string
    var dryRun bool
    
    cmd := &cobra.Command{
        Use:   "migrate-to-sqlite",
        Short: "Migrate data from PostgreSQL to SQLite",
        Run: func(_ *cobra.Command, _ []string) {
            if dryRun {
                fmt.Println("DRY RUN MODE - No changes will be made")
            }
            handleCommandError(app.MigrateToSQLite(sqlitePath, dryRun))
        },
    }
    
    cmd.Flags().StringVar(&sqlitePath, "path", "./lc_base/data.db", "SQLite database path")
    cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate migration without making changes")
    
    return cmd
}
```

### Migration Logic

```go
// internal/app/migrate_db.go
package app

import (
    "context"
    "fmt"
    "time"
    
    "github.com/shurco/mycart/internal/database"
)

func MigrateToPostgres(databaseURL string, dryRun bool) error {
    ctx := context.Background()
    
    fmt.Println("🔄 Starting migration from SQLite to PostgreSQL...")
    
    // Create backup
    if !dryRun {
        backupPath := fmt.Sprintf("./lc_base/data.db.backup.%d", time.Now().Unix())
        fmt.Printf("📦 Creating backup: %s\n", backupPath)
        if err := copyFile("./lc_base/data.db", backupPath); err != nil {
            return fmt.Errorf("backup failed: %w", err)
        }
    }
    
    // Load source (SQLite)
    fmt.Println("📖 Reading SQLite database...")
    sqliteCfg := &database.Config{Type: "sqlite"}
    sourceDB, err := database.New(sqliteCfg)
    if err != nil {
        return fmt.Errorf("failed to open SQLite: %w", err)
    }
    
    // Connect to target (PostgreSQL)
    fmt.Println("🔌 Connecting to PostgreSQL...")
    pgCfg := &database.Config{
        Type: "postgres",
        PostgreSQL: database.ParseConnectionURL(databaseURL),
    }
    
    if err := database.TestConnection(pgCfg); err != nil {
        return fmt.Errorf("PostgreSQL connection failed: %w", err)
    }
    
    // Run migrations
    fmt.Println("📦 Creating PostgreSQL schema...")
    if !dryRun {
        if err := database.RunMigrations(pgCfg, migrations.Embed()); err != nil {
            return fmt.Errorf("migration failed: %w", err)
        }
    }
    
    targetDB, err := database.New(pgCfg)
    if err != nil {
        return fmt.Errorf("failed to open PostgreSQL: %w", err)
    }
    
    // Migrate data
    fmt.Println("📊 Migrating data...")
    if err := migrateData(ctx, sourceDB, targetDB, dryRun); err != nil {
        return fmt.Errorf("data migration failed: %w", err)
    }
    
    // Verify
    fmt.Println("✓ Verifying data integrity...")
    if err := verifyMigration(ctx, sourceDB, targetDB); err != nil {
        return fmt.Errorf("verification failed: %w", err)
    }
    
    // Update config
    if !dryRun {
        fmt.Println("📝 Updating configuration...")
        if err := updateEnvFile("postgres", databaseURL); err != nil {
            return fmt.Errorf("failed to update .env: %w", err)
        }
    }
    
    fmt.Println("✅ Migration completed successfully!")
    return nil
}

func migrateData(ctx context.Context, source, target database.Database, dryRun bool) error {
    tables := []string{
        "settings",
        "users",
        "pages",
        "products",
        "product_variants",
        "carts",
        "cart_items",
    }
    
    for _, table := range tables {
        fmt.Printf("  • Migrating %s...\n", table)
        
        if dryRun {
            count, _ := source.CountTable(ctx, table)
            fmt.Printf("    Would migrate %d rows\n", count)
            continue
        }
        
        // Begin transaction
        tx, err := target.BeginTx(ctx)
        if err != nil {
            return err
        }
        
        // Migrate table data
        if err := migrateTable(ctx, source, tx, table); err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to migrate %s: %w", table, err)
        }
        
        if err := tx.Commit(); err != nil {
            return err
        }
    }
    
    return nil
}

func verifyMigration(ctx context.Context, source, target database.Database) error {
    tables := []string{"settings", "users", "pages", "products", "carts"}
    
    for _, table := range tables {
        sourceCount, err := source.CountTable(ctx, table)
        if err != nil {
            return err
        }
        
        targetCount, err := target.CountTable(ctx, table)
        if err != nil {
            return err
        }
        
        if sourceCount != targetCount {
            return fmt.Errorf("row count mismatch in %s: source=%d, target=%d", 
                table, sourceCount, targetCount)
        }
        
        fmt.Printf("  ✓ %s: %d rows\n", table, sourceCount)
    }
    
    return nil
}
```

---

## 8. Docker & Docker Compose Setup

### PostgreSQL Compose File

`docker/docker-compose-postgres.yml`:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: mycart-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${DB_NAME:-mycart}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD:-changeme}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "${DB_PORT:-5432}:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-postgres}"]
      interval: 10s
      timeout: 5s
      retries: 5

  mycart:
    image: shurco/mycart:latest
    container_name: mycart
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DB_TYPE: postgres
      DB_HOST: postgres
      DB_PORT: 5432
      DB_NAME: ${DB_NAME:-mycart}
      DB_USER: ${DB_USER:-postgres}
      DB_PASSWORD: ${DB_PASSWORD:-changeme}
      DB_SSLMODE: disable
    ports:
      - "${HTTP_PORT:-8080}:8080"
    volumes:
      - ./lc_digitals:/lc_digitals
      - ./lc_uploads:/lc_uploads
      - ./site:/site

volumes:
  postgres_data:
```

### Supabase Compose File

`docker/docker-compose-supabase.yml`:
```yaml
version: '3.8'

services:
  mycart:
    image: shurco/mycart:latest
    container_name: mycart
    restart: unless-stopped
    environment:
      DATABASE_URL: ${DATABASE_URL}
      DB_TYPE: postgres
    ports:
      - "${HTTP_PORT:-8080}:8080"
    volumes:
      - ./lc_digitals:/lc_digitals
      - ./lc_uploads:/lc_uploads
      - ./site:/site
```

### Environment Templates

`docker/.env.postgres.example`:
```bash
DB_TYPE=postgres
DB_HOST=postgres
DB_PORT=5432
DB_NAME=mycart
DB_USER=postgres
DB_PASSWORD=your_secure_password_here
DB_SSLMODE=disable

HTTP_PORT=8080
```

`docker/.env.supabase.example`:
```bash
DB_TYPE=postgres
DATABASE_URL=postgresql://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres
DB_SSLMODE=require

HTTP_PORT=8080
```

---

## 9. Error Handling & Connection Retry

### Retry Strategy

```go
// internal/database/connect.go
type RetryConfig struct {
    MaxAttempts  int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

var DefaultRetryConfig = RetryConfig{
    MaxAttempts:  5,
    InitialDelay: 1 * time.Second,
    MaxDelay:     30 * time.Second,
    Multiplier:   2.0,
}

func ConnectWithRetry(cfg *Config) (*sql.DB, error) {
    var db *sql.DB
    var err error
    
    delay := DefaultRetryConfig.InitialDelay
    
    for attempt := 1; attempt <= DefaultRetryConfig.MaxAttempts; attempt++ {
        db, err = connect(cfg)
        if err == nil {
            return db, nil
        }
        
        if attempt < DefaultRetryConfig.MaxAttempts {
            fmt.Printf("⚠️  Database connection failed (attempt %d/%d): %v\n", 
                attempt, DefaultRetryConfig.MaxAttempts, err)
            fmt.Printf("   Retrying in %v...\n", delay)
            
            time.Sleep(delay)
            delay = time.Duration(float64(delay) * DefaultRetryConfig.Multiplier)
            if delay > DefaultRetryConfig.MaxDelay {
                delay = DefaultRetryConfig.MaxDelay
            }
        }
    }
    
    return nil, fmt.Errorf("failed to connect after %d attempts: %w", 
        DefaultRetryConfig.MaxAttempts, err)
}

func connect(cfg *Config) (*sql.DB, error) {
    var db *sql.DB
    var err error
    
    switch cfg.Type {
    case "postgres":
        db, err = sql.Open("postgres", cfg.PostgreSQL.ConnectionString())
        if err != nil {
            return nil, err
        }
        
        db.SetMaxOpenConns(cfg.PostgreSQL.MaxOpenConns)
        db.SetMaxIdleConns(cfg.PostgreSQL.MaxIdleConns)
        db.SetConnMaxLifetime(cfg.PostgreSQL.ConnMaxLifetime)
        
    case "sqlite":
        dsn := buildSQLiteDSN(cfg.SQLite.Path)
        db, err = sql.Open("sqlite", dsn)
        if err != nil {
            return nil, err
        }
        
        db.SetMaxOpenConns(1)
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        db.Close()
        return nil, fmt.Errorf("ping failed: %w", err)
    }
    
    return db, nil
}
```

### Health Monitoring

```go
// internal/database/health.go
type HealthChecker struct {
    db            *sql.DB
    isHealthy     bool
    lastCheck     time.Time
    mu            sync.RWMutex
    checkInterval time.Duration
}

func NewHealthChecker(db *sql.DB) *HealthChecker {
    hc := &HealthChecker{
        db:            db,
        isHealthy:     true,
        checkInterval: 30 * time.Second,
    }
    go hc.monitorHealth()
    return hc
}

func (hc *HealthChecker) monitorHealth() {
    ticker := time.NewTicker(hc.checkInterval)
    defer ticker.Stop()
    
    for range ticker.C {
        healthy := hc.check()
        
        hc.mu.Lock()
        wasHealthy := hc.isHealthy
        hc.isHealthy = healthy
        hc.lastCheck = time.Now()
        hc.mu.Unlock()
        
        if !healthy && wasHealthy {
            fmt.Println("⚠️  Database connection lost")
        } else if healthy && !wasHealthy {
            fmt.Println("✅ Database connection restored")
        }
    }
}
```

### Health Check Endpoint

```go
// GET /api/health
func HealthCheck(c fiber.Ctx) error {
    ctx := c.Context()
    db := queries.DB()
    
    _, err := db.GetSettingByKey(ctx, "installed")
    if err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status":   "unhealthy",
            "database": "disconnected",
        })
    }
    
    return c.JSON(fiber.Map{
        "status":    "healthy",
        "database":  "connected",
        "timestamp": time.Now().UTC(),
    })
}
```

---

## 10. Testing Strategy

### Test Structure

```
test/
├── integration/
│   ├── sqlite_test.go
│   ├── postgres_test.go
│   └── migration_test.go
└── fixtures/
    └── test_data.sql

internal/
├── database/
│   ├── adapter_test.go
│   ├── config_test.go
│   └── migrate_test.go
└── testutil/
    └── testdb.go
```

### Unit Tests

Both adapters must pass identical test suites:

```go
// internal/database/adapter_test.go
func TestSQLiteAdapter(t *testing.T) {
    db := testutil.NewTestDB(t, "sqlite")
    defer db.Close()
    
    ctx := context.Background()
    
    err := db.CreateSetting(ctx, "test_key", "test_value")
    assert.NoError(t, err)
    
    result, err := db.GetSettingByKey(ctx, "test_key")
    assert.NoError(t, err)
    assert.Equal(t, "test_value", result["test_key"].Value)
}

func TestPostgresAdapter(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping PostgreSQL test")
    }
    
    db := testutil.NewTestDB(t, "postgres")
    defer db.Close()
    
    // Same test as SQLite
}
```

### Integration Tests

```go
// test/integration/migration_test.go
func TestMigrateSQLiteToPostgres(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    ctx := context.Background()
    
    sqliteDB := testutil.NewTestDB(t, "sqlite")
    testutil.SeedTestData(t, sqliteDB)
    
    pgDB := testutil.NewTestDB(t, "postgres")
    
    err := database.MigrateData(ctx, sqliteDB, pgDB)
    assert.NoError(t, err)
    
    // Verify data matches
    sqliteSettings, _ := sqliteDB.ListSettings(ctx)
    pgSettings, _ := pgDB.ListSettings(ctx)
    assert.Equal(t, len(sqliteSettings), len(pgSettings))
}
```

### CI/CD

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: test
          POSTGRES_DB: mycart_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
        ports:
          - 5432:5432
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.26'
      
      - name: Install sqlc
        run: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
      
      - name: Generate sqlc code
        run: sqlc generate
      
      - name: Run tests
        env:
          TEST_DATABASE_URL: postgresql://postgres:test@localhost:5432/mycart_test?sslmode=disable
        run: go test -v ./...
```

### Test Checklist

- ✅ Unit tests pass for both SQLite and PostgreSQL adapters
- ✅ Integration tests verify bidirectional migration
- ✅ sqlc generates valid code without errors
- ✅ Connection retry works correctly
- ✅ Health check endpoint responds accurately
- ✅ Web installer handles both database types
- ✅ CLI installer accepts database flags
- ✅ Docker Compose configurations work
- ✅ Migration commands complete successfully

---

## Implementation Timeline

### Phase 1: Database Layer (2-3 days)
1. Set up directory structure
2. Create PostgreSQL migrations from SQLite migrations
3. Write sqlc query files for both databases
4. Configure sqlc and generate code
5. Create unified Database interface
6. Implement SQLite and PostgreSQL adapters
7. Write unit tests

### Phase 2: Configuration & Connection (1 day)
1. Implement configuration loading
2. Add database detection logic
3. Implement connection with retry
4. Add health checking
5. Create backward compatibility layer
6. Test configuration loading

### Phase 3: Installer & CLI (1 day)
1. Update web installer UI
2. Add database selection API endpoints
3. Update CLI install command with database flags
4. Test both installation methods

### Phase 4: Migration Tools (1-2 days)
1. Implement data migration logic
2. Add migrate-to-postgres command
3. Add migrate-to-sqlite command
4. Add verification and rollback
5. Test migrations with real data

### Phase 5: Docker & Documentation (0.5 days)
1. Create PostgreSQL Docker Compose file
2. Create Supabase Docker Compose file
3. Add environment file examples
4. Update README with PostgreSQL instructions

**Total Estimated Time: 5.5-7.5 days**

---

## Success Criteria

1. ✅ Existing SQLite users experience no changes
2. ✅ New users can choose PostgreSQL during installation
3. ✅ Supabase connection works out-of-the-box
4. ✅ Both databases pass identical test suites
5. ✅ Migration commands successfully convert data bidirectionally
6. ✅ Docker deployments work for both databases
7. ✅ Health checks and retry logic work reliably
8. ✅ All CI/CD tests pass

---

## Future Enhancements

- [ ] Database backup/restore commands
- [ ] Connection pool monitoring dashboard
- [ ] Support for read replicas
- [ ] Multi-database support (write to both simultaneously)
- [ ] Performance benchmarking tools
- [ ] Database migration progress UI
- [ ] Support for other PostgreSQL-compatible databases (CockroachDB, YugabyteDB)
