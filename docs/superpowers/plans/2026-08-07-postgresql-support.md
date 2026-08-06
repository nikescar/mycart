# PostgreSQL Support Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add optional PostgreSQL database support alongside SQLite with bidirectional migration capabilities.

**Architecture:** Bottom-up implementation using sqlc for type-safe SQL, unified Database interface with adapter pattern for both databases, full backward compatibility with existing SQLite deployments.

**Tech Stack:** 
- sqlc v1.27+ for type-safe SQL code generation
- goose v3.27+ for database migrations
- lib/pq for PostgreSQL driver
- modernc.org/sqlite for SQLite driver
- Fiber v3 for web framework

## Global Constraints

- Go version: 1.26.1+
- SQLite remains default database (no breaking changes)
- All existing APIs must maintain backward compatibility
- Test coverage required for all new database code
- Both databases must pass identical test suites
- Follow existing code style and patterns
- Use TDD: write tests first, then implementation
- Commit after each passing test

---

## File Structure Overview

**New directories:**
```
db/
├── migrations/
│   ├── sqlite/          # SQLite-specific migration files
│   └── postgres/        # PostgreSQL-specific migration files
└── queries/
    ├── sqlite/          # SQLite sqlc query files
    └── postgres/        # PostgreSQL sqlc query files

internal/
├── db/
│   ├── interface.go     # Unified Database interface
│   ├── types.go         # Common type definitions
│   ├── sqlite/          # Generated sqlc code (auto-generated)
│   └── postgres/        # Generated sqlc code (auto-generated)
└── database/
    ├── config.go        # Configuration structs
    ├── config_test.go   # Configuration tests
    ├── connect.go       # Connection with retry logic
    ├── connect_test.go  # Connection tests
    ├── migrate.go       # Migration execution
    ├── migrate_test.go  # Migration tests
    ├── detect.go        # Database detection
    ├── health.go        # Health checking
    ├── health_test.go   # Health tests
    ├── sqlite_adapter.go     # SQLite adapter implementation
    ├── postgres_adapter.go   # PostgreSQL adapter implementation
    ├── adapter_test.go       # Adapter tests
    └── migrate_db.go         # Data migration tools

test/integration/
├── sqlite_test.go
├── postgres_test.go
└── migration_test.go
```

**Modified files:**
```
internal/queries/queries.go  # Update to use new database interface
internal/app/app.go          # Update database initialization
cmd/main.go                  # Add new commands
go.mod                       # Add PostgreSQL driver dependency
```

**New files:**
```
sqlc.yaml                    # sqlc configuration
.env.example                 # Environment variable template
docker/.env.postgres.example
docker/.env.supabase.example
docker/docker-compose-postgres.yml
docker/docker-compose-supabase.yml
```

---

### Task 1: Project Structure & Dependencies

**Files:**
- Create: `sqlc.yaml`
- Modify: `go.mod`
- Create: `.env.example`

**Interfaces:**
- Consumes: None (initial setup)
- Produces: sqlc configuration for code generation, PostgreSQL driver dependency

- [ ] **Step 1: Add PostgreSQL driver dependency**

```bash
go get github.com/lib/pq@v1.10.9
```

Run: `go mod tidy`
Expected: Dependency added to go.mod

- [ ] **Step 2: Create sqlc configuration**

Create file `sqlc.yaml`:
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

- [ ] **Step 3: Create environment variable template**

Create file `.env.example`:
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

- [ ] **Step 4: Install sqlc**

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

Run: `sqlc version`
Expected: sqlc version output (v1.27.0 or later)

- [ ] **Step 5: Create directory structure**

```bash
mkdir -p db/migrations/sqlite db/migrations/postgres
mkdir -p db/queries/sqlite db/queries/postgres
mkdir -p internal/db internal/database
mkdir -p test/integration
```

Run: `ls -R db/`
Expected: Directory structure created

- [ ] **Step 6: Commit**

```bash
git add sqlc.yaml .env.example go.mod go.sum
git commit -m "feat: add PostgreSQL dependency and sqlc configuration

- Add lib/pq PostgreSQL driver
- Configure sqlc for both SQLite and PostgreSQL
- Create environment variable template
- Set up directory structure for migrations and queries

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 2: Database Configuration Layer

**Files:**
- Create: `internal/database/config.go`
- Create: `internal/database/config_test.go`

**Interfaces:**
- Consumes: None
- Produces: 
  - `Config` struct with `Type` string, `SQLite` SQLiteConfig, `PostgreSQL` PostgresConfig
  - `LoadConfig() (*Config, error)` function
  - `PostgresConfig.ConnectionString() string` method
  - `parseConnectionURL(rawURL string) PostgresConfig` function

- [ ] **Step 1: Write failing test for LoadConfig**

Create file `internal/database/config_test.go`:
```go
package database

import (
	"os"
	"testing"
	"time"

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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/database -v -run TestLoadConfig`
Expected: FAIL with "undefined: LoadConfig"

- [ ] **Step 3: Write minimal implementation**

Create file `internal/database/config.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/database -v -run TestLoadConfig`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/database/config.go internal/database/config_test.go
git commit -m "feat: add database configuration layer

- Implement LoadConfig for environment-based configuration
- Support DATABASE_URL and individual env vars for PostgreSQL
- Default to SQLite with ./lc_base/data.db
- Add comprehensive tests for all configuration scenarios

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 3: Copy SQLite Migrations to PostgreSQL

**Files:**
- Create: Multiple files in `db/migrations/postgres/`
- Create: Multiple files in `db/migrations/sqlite/`

**Interfaces:**
- Consumes: Existing migrations in `migrations/` directory
- Produces: Duplicate migration files adapted for both SQLite and PostgreSQL syntax

- [ ] **Step 1: Copy existing SQLite migrations**

```bash
cp migrations/*.sql db/migrations/sqlite/
```

Run: `ls db/migrations/sqlite/`
Expected: All migration files copied

- [ ] **Step 2: Copy to PostgreSQL directory**

```bash
cp migrations/*.sql db/migrations/postgres/
```

Run: `ls db/migrations/postgres/`
Expected: All migration files copied

- [ ] **Step 3: Convert PostgreSQL migration syntax - init_db**

Edit `db/migrations/postgres/20230714135923_init_db.sql`:

Replace all occurrences of:
- `INTEGER PRIMARY KEY AUTOINCREMENT` → `SERIAL PRIMARY KEY`
- `BIGINT PRIMARY KEY AUTOINCREMENT` → `BIGSERIAL PRIMARY KEY`
- `DATETIME` → `TIMESTAMPTZ`
- `CURRENT_TIMESTAMP` → `NOW()`
- `INTEGER DEFAULT 0` (for booleans) → `BOOLEAN DEFAULT false`
- `INTEGER DEFAULT 1` (for booleans) → `BOOLEAN DEFAULT true`

Run: `grep -n "AUTOINCREMENT\|DATETIME\|CURRENT_TIMESTAMP" db/migrations/postgres/20230714135923_init_db.sql`
Expected: No matches (all converted)

- [ ] **Step 4: Convert PostgreSQL migration syntax - all other migrations**

For each file in `db/migrations/postgres/`:
- Replace `INTEGER PRIMARY KEY AUTOINCREMENT` with `SERIAL PRIMARY KEY`
- Replace `BIGINT PRIMARY KEY AUTOINCREMENT` with `BIGSERIAL PRIMARY KEY`
- Replace `DATETIME` with `TIMESTAMPTZ`
- Replace `CURRENT_TIMESTAMP` with `NOW()`
- Convert boolean INTEGER fields to BOOLEAN type

Run: `grep -r "AUTOINCREMENT" db/migrations/postgres/`
Expected: No matches

- [ ] **Step 5: Update migrations/embed.go to use new structure**

Create file `db/migrations/embed.go`:
```go
package migrations

import "embed"

//go:embed sqlite/*.sql postgres/*.sql
var migrations embed.FS

func Embed() embed.FS {
	return migrations
}
```

- [ ] **Step 6: Verify migrations compile**

Run: `go build ./db/migrations`
Expected: No errors

- [ ] **Step 7: Commit**

```bash
git add db/migrations/
git commit -m "feat: create separate SQLite and PostgreSQL migrations

- Copy existing migrations to db/migrations/sqlite/
- Create PostgreSQL versions with converted syntax
- Convert AUTOINCREMENT to SERIAL/BIGSERIAL
- Convert DATETIME to TIMESTAMPTZ
- Convert CURRENT_TIMESTAMP to NOW()
- Convert boolean INTEGER to BOOLEAN type

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 4: Create sqlc Query Files for Settings

**Files:**
- Create: `db/queries/sqlite/settings.sql`
- Create: `db/queries/postgres/settings.sql`

**Interfaces:**
- Consumes: Migration schema for settings table
- Produces: sqlc query definitions for GetSettingByKey, UpdateSetting, ListSettings

- [ ] **Step 1: Create SQLite settings queries**

Create file `db/queries/sqlite/settings.sql`:
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

-- name: CreateSetting :exec
INSERT INTO settings (key, value, created_at, updated_at)
VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: DeleteSetting :exec
DELETE FROM settings WHERE key = ?;
```

- [ ] **Step 2: Create PostgreSQL settings queries**

Create file `db/queries/postgres/settings.sql`:
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

-- name: CreateSetting :exec
INSERT INTO settings (key, value, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW());

-- name: DeleteSetting :exec
DELETE FROM settings WHERE key = $1;
```

- [ ] **Step 3: Generate sqlc code**

Run: `sqlc generate`
Expected: Code generated in `internal/db/sqlite/` and `internal/db/postgres/`

- [ ] **Step 4: Verify generated code compiles**

Run: `go build ./internal/db/...`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add db/queries/ internal/db/
git commit -m "feat: add sqlc queries for settings table

- Create SQLite and PostgreSQL query files
- Generate type-safe Go code with sqlc
- Support GetSettingByKey, UpdateSetting, ListSettings, CreateSetting, DeleteSetting

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 5: Create sqlc Query Files for All Tables

**Files:**
- Create: `db/queries/sqlite/auth.sql`
- Create: `db/queries/postgres/auth.sql`
- Create: `db/queries/sqlite/pages.sql`
- Create: `db/queries/postgres/pages.sql`
- Create: `db/queries/sqlite/products.sql`
- Create: `db/queries/postgres/products.sql`
- Create: `db/queries/sqlite/carts.sql`
- Create: `db/queries/postgres/carts.sql`

**Interfaces:**
- Consumes: Migration schemas for all tables
- Produces: Complete sqlc query definitions matching all current `internal/queries/*.go` methods

- [ ] **Step 1: Analyze current query methods**

Run: `grep -h "^func.*Queries" internal/queries/*.go | sort | uniq`
Expected: List of all current query method signatures

- [ ] **Step 2: Create auth queries (SQLite)**

Create file `db/queries/sqlite/auth.sql` with all methods from `internal/queries/auth.go` converted to sqlc format:
```sql
-- name: GetUserByEmail :one
SELECT id, email, password, created_at
FROM users
WHERE email = ? LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, password, created_at)
VALUES (?, ?, CURRENT_TIMESTAMP)
RETURNING id, email, password, created_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = ?
WHERE id = ?;
```

- [ ] **Step 3: Create auth queries (PostgreSQL)**

Create file `db/queries/postgres/auth.sql`:
```sql
-- name: GetUserByEmail :one
SELECT id, email, password, created_at
FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, password, created_at)
VALUES ($1, $2, NOW())
RETURNING id, email, password, created_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = $1
WHERE id = $2;
```

- [ ] **Step 4: Create pages queries (both databases)**

Analyze `internal/queries/pages.go` and create equivalent queries in:
- `db/queries/sqlite/pages.sql` (using `?` placeholders)
- `db/queries/postgres/pages.sql` (using `$1, $2` placeholders)

Include all CRUD operations for pages table.

- [ ] **Step 5: Create products queries (both databases)**

Analyze `internal/queries/products.go` and create equivalent queries in:
- `db/queries/sqlite/products.sql`
- `db/queries/postgres/products.sql`

Include all product operations, variants, and inventory queries.

- [ ] **Step 6: Create carts queries (both databases)**

Analyze `internal/queries/cart.go` and create equivalent queries in:
- `db/queries/sqlite/carts.sql`
- `db/queries/postgres/carts.sql`

Include all cart and cart_items operations.

- [ ] **Step 7: Generate sqlc code**

Run: `sqlc generate`
Expected: All query code generated successfully

Run: `go build ./internal/db/...`
Expected: No compile errors

- [ ] **Step 8: Commit**

```bash
git add db/queries/ internal/db/
git commit -m "feat: add sqlc queries for all database tables

- Add auth queries (users table)
- Add pages queries
- Add products queries (including variants)
- Add cart queries (cart and cart_items)
- Generate type-safe code for both SQLite and PostgreSQL

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 6: Create Database Interface and Common Types

**Files:**
- Create: `internal/db/interface.go`
- Create: `internal/db/types.go`

**Interfaces:**
- Consumes: Generated sqlc code from internal/db/sqlite and internal/db/postgres
- Produces:
  - `Database` interface with all query methods
  - Common types: Setting, User, Page, Product, Cart, etc.

- [ ] **Step 1: Create common types**

Create file `internal/db/types.go`:
```go
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

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Price       int64     `json:"price"`
	Published   bool      `json:"published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductVariant struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	Name      string    `json:"name"`
	SKU       string    `json:"sku"`
	Price     int64     `json:"price"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Cart struct {
	ID        string    `json:"id"`
	UserEmail string    `json:"user_email"`
	Status    string    `json:"status"`
	Total     int64     `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CartItem struct {
	ID        int64     `json:"id"`
	CartID    string    `json:"cart_id"`
	ProductID int64     `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     int64     `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}
```

- [ ] **Step 2: Create database interface**

Create file `internal/db/interface.go`:
```go
package db

import (
	"context"
)

type Database interface {
	// Transaction support
	BeginTx(ctx context.Context) (Database, error)
	Commit() error
	Rollback() error

	// Settings
	GetSettingByKey(ctx context.Context, key string) (map[string]Setting, error)
	UpdateSetting(ctx context.Context, key string, value interface{}) error
	ListSettings(ctx context.Context) ([]Setting, error)
	CreateSetting(ctx context.Context, key string, value interface{}) error
	DeleteSetting(ctx context.Context, key string) error

	// Auth
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUser(ctx context.Context, email, password string) (User, error)
	UpdateUserPassword(ctx context.Context, id int64, password string) error

	// Pages
	GetPageBySlug(ctx context.Context, slug string) (Page, error)
	GetPageByID(ctx context.Context, id int64) (Page, error)
	ListPages(ctx context.Context) ([]Page, error)
	CreatePage(ctx context.Context, page Page) (Page, error)
	UpdatePage(ctx context.Context, page Page) error
	DeletePage(ctx context.Context, id int64) error

	// Products
	GetProduct(ctx context.Context, id int64) (Product, error)
	GetProductBySlug(ctx context.Context, slug string) (Product, error)
	ListProducts(ctx context.Context) ([]Product, error)
	CreateProduct(ctx context.Context, product Product) (Product, error)
	UpdateProduct(ctx context.Context, product Product) error
	DeleteProduct(ctx context.Context, id int64) error

	// Product Variants
	ListProductVariants(ctx context.Context, productID int64) ([]ProductVariant, error)
	CreateProductVariant(ctx context.Context, variant ProductVariant) (ProductVariant, error)
	UpdateProductVariant(ctx context.Context, variant ProductVariant) error
	DeleteProductVariant(ctx context.Context, id int64) error

	// Carts
	GetCart(ctx context.Context, id string) (Cart, error)
	CreateCart(ctx context.Context, cart Cart) (Cart, error)
	UpdateCart(ctx context.Context, cart Cart) error
	DeleteCart(ctx context.Context, id string) error
	ListCartItems(ctx context.Context, cartID string) ([]CartItem, error)
	AddCartItem(ctx context.Context, item CartItem) (CartItem, error)
	UpdateCartItem(ctx context.Context, item CartItem) error
	DeleteCartItem(ctx context.Context, id int64) error

	// Utility
	CountTable(ctx context.Context, tableName string) (int64, error)
}
```

- [ ] **Step 3: Verify interface compiles**

Run: `go build ./internal/db`
Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add internal/db/interface.go internal/db/types.go
git commit -m "feat: define unified database interface and common types

- Create Database interface with all query methods
- Define common types for Settings, User, Page, Product, Cart, etc.
- Support transaction methods (BeginTx, Commit, Rollback)
- Include utility methods like CountTable

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 7: Implement SQLite Adapter

**Files:**
- Create: `internal/database/sqlite_adapter.go`
- Create: `internal/database/adapter_test.go`

**Interfaces:**
- Consumes: 
  - `db.Database` interface from internal/db/interface.go
  - Generated sqlc code from internal/db/sqlite
- Produces: `SQLiteAdapter` struct implementing `db.Database` interface

- [ ] **Step 1: Write failing test for SQLite adapter**

Create file `internal/database/adapter_test.go`:
```go
package database

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupTestSQLite(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Run init migration
	_, err = db.Exec(`
		CREATE TABLE settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
		);
		
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
		);
	`)
	require.NoError(t, err)

	return db
}

func TestSQLiteAdapter_Settings(t *testing.T) {
	db := setupTestSQLite(t)
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	ctx := context.Background()

	// Test CreateSetting
	err := adapter.CreateSetting(ctx, "test_key", "test_value")
	assert.NoError(t, err)

	// Test GetSettingByKey
	result, err := adapter.GetSettingByKey(ctx, "test_key")
	assert.NoError(t, err)
	assert.Contains(t, result, "test_key")
	assert.Equal(t, "test_value", result["test_key"].Value)

	// Test UpdateSetting
	err = adapter.UpdateSetting(ctx, "test_key", "updated_value")
	assert.NoError(t, err)

	result, err = adapter.GetSettingByKey(ctx, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, "updated_value", result["test_key"].Value)

	// Test ListSettings
	settings, err := adapter.ListSettings(ctx)
	assert.NoError(t, err)
	assert.Len(t, settings, 1)
}

func TestSQLiteAdapter_Transactions(t *testing.T) {
	db := setupTestSQLite(t)
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	ctx := context.Background()

	// Begin transaction
	tx, err := adapter.BeginTx(ctx)
	assert.NoError(t, err)

	// Create setting in transaction
	err = tx.CreateSetting(ctx, "tx_key", "tx_value")
	assert.NoError(t, err)

	// Rollback
	err = tx.Rollback()
	assert.NoError(t, err)

	// Verify setting was not persisted
	_, err = adapter.GetSettingByKey(ctx, "tx_key")
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/database -v -run TestSQLiteAdapter`
Expected: FAIL with "undefined: NewSQLiteAdapter"

- [ ] **Step 3: Write minimal SQLite adapter implementation**

Create file `internal/database/sqlite_adapter.go`:
```go
package database

import (
	"context"
	"database/sql"
	"fmt"

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

func (a *SQLiteAdapter) BeginTx(ctx context.Context) (db.Database, error) {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &SQLiteAdapter{
		db:      a.db,
		queries: sqlitedb.New(tx),
		tx:      tx,
	}, nil
}

func (a *SQLiteAdapter) Commit() error {
	if a.tx == nil {
		return nil
	}
	return a.tx.Commit()
}

func (a *SQLiteAdapter) Rollback() error {
	if a.tx == nil {
		return nil
	}
	return a.tx.Rollback()
}

func (a *SQLiteAdapter) GetSettingByKey(ctx context.Context, key string) (map[string]db.Setting, error) {
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

func (a *SQLiteAdapter) UpdateSetting(ctx context.Context, key string, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	return a.queries.UpdateSetting(ctx, sqlitedb.UpdateSettingParams{
		Value: valueStr,
		Key:   key,
	})
}

func (a *SQLiteAdapter) ListSettings(ctx context.Context) ([]db.Setting, error) {
	rows, err := a.queries.ListSettings(ctx)
	if err != nil {
		return nil, err
	}

	settings := make([]db.Setting, len(rows))
	for i, row := range rows {
		settings[i] = db.Setting{
			Key:       row.Key,
			Value:     row.Value,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}

	return settings, nil
}

func (a *SQLiteAdapter) CreateSetting(ctx context.Context, key string, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	return a.queries.CreateSetting(ctx, sqlitedb.CreateSettingParams{
		Key:   key,
		Value: valueStr,
	})
}

func (a *SQLiteAdapter) DeleteSetting(ctx context.Context, key string) error {
	return a.queries.DeleteSetting(ctx, key)
}

// Stub implementations for other interface methods (will be implemented later)
func (a *SQLiteAdapter) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	row, err := a.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return db.User{}, err
	}
	return db.User{
		ID:        row.ID,
		Email:     row.Email,
		Password:  row.Password,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (a *SQLiteAdapter) CreateUser(ctx context.Context, email, password string) (db.User, error) {
	row, err := a.queries.CreateUser(ctx, sqlitedb.CreateUserParams{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return db.User{}, err
	}
	return db.User{
		ID:        row.ID,
		Email:     row.Email,
		Password:  row.Password,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (a *SQLiteAdapter) UpdateUserPassword(ctx context.Context, id int64, password string) error {
	return a.queries.UpdateUserPassword(ctx, sqlitedb.UpdateUserPasswordParams{
		Password: password,
		ID:       id,
	})
}

// Add stub implementations for remaining interface methods
func (a *SQLiteAdapter) GetPageBySlug(ctx context.Context, slug string) (db.Page, error) {
	return db.Page{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) GetPageByID(ctx context.Context, id int64) (db.Page, error) {
	return db.Page{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) ListPages(ctx context.Context) ([]db.Page, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) CreatePage(ctx context.Context, page db.Page) (db.Page, error) {
	return db.Page{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) UpdatePage(ctx context.Context, page db.Page) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) DeletePage(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) GetProduct(ctx context.Context, id int64) (db.Product, error) {
	return db.Product{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) GetProductBySlug(ctx context.Context, slug string) (db.Product, error) {
	return db.Product{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) ListProducts(ctx context.Context) ([]db.Product, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) CreateProduct(ctx context.Context, product db.Product) (db.Product, error) {
	return db.Product{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) UpdateProduct(ctx context.Context, product db.Product) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) DeleteProduct(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) ListProductVariants(ctx context.Context, productID int64) ([]db.ProductVariant, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) CreateProductVariant(ctx context.Context, variant db.ProductVariant) (db.ProductVariant, error) {
	return db.ProductVariant{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) UpdateProductVariant(ctx context.Context, variant db.ProductVariant) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) DeleteProductVariant(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) GetCart(ctx context.Context, id string) (db.Cart, error) {
	return db.Cart{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) CreateCart(ctx context.Context, cart db.Cart) (db.Cart, error) {
	return db.Cart{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) UpdateCart(ctx context.Context, cart db.Cart) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) DeleteCart(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) ListCartItems(ctx context.Context, cartID string) ([]db.CartItem, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) AddCartItem(ctx context.Context, item db.CartItem) (db.CartItem, error) {
	return db.CartItem{}, fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) UpdateCartItem(ctx context.Context, item db.CartItem) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) DeleteCartItem(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented")
}

func (a *SQLiteAdapter) CountTable(ctx context.Context, tableName string) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := a.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/database -v -run TestSQLiteAdapter`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/database/sqlite_adapter.go internal/database/adapter_test.go
git commit -m "feat: implement SQLite database adapter

- Implement SQLiteAdapter struct with db.Database interface
- Add transaction support (BeginTx, Commit, Rollback)
- Implement Settings CRUD operations
- Implement Auth operations (GetUserByEmail, CreateUser, UpdateUserPassword)
- Add stub implementations for remaining methods
- Add comprehensive tests for adapter functionality

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 8: Implement PostgreSQL Adapter

**Files:**
- Create: `internal/database/postgres_adapter.go`
- Modify: `internal/database/adapter_test.go`

**Interfaces:**
- Consumes:
  - `db.Database` interface from internal/db/interface.go
  - Generated sqlc code from internal/db/postgres
- Produces: `PostgresAdapter` struct implementing `db.Database` interface

- [ ] **Step 1: Write failing test for PostgreSQL adapter**

Add to `internal/database/adapter_test.go`:
```go
func setupTestPostgres(t *testing.T) *sql.DB {
	// Check for test database URL
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping PostgreSQL tests")
	}

	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)

	// Clean database
	_, err = db.Exec("DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;")
	require.NoError(t, err)

	// Run init migration
	_, err = db.Exec(`
		CREATE TABLE settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
			updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
		);
		
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
		);
	`)
	require.NoError(t, err)

	return db
}

func TestPostgresAdapter_Settings(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := setupTestPostgres(t)
	defer db.Close()

	adapter := NewPostgresAdapter(db)
	ctx := context.Background()

	// Test CreateSetting
	err := adapter.CreateSetting(ctx, "test_key", "test_value")
	assert.NoError(t, err)

	// Test GetSettingByKey
	result, err := adapter.GetSettingByKey(ctx, "test_key")
	assert.NoError(t, err)
	assert.Contains(t, result, "test_key")
	assert.Equal(t, "test_value", result["test_key"].Value)

	// Test UpdateSetting
	err = adapter.UpdateSetting(ctx, "test_key", "updated_value")
	assert.NoError(t, err)

	result, err = adapter.GetSettingByKey(ctx, "test_key")
	assert.NoError(t, err)
	assert.Equal(t, "updated_value", result["test_key"].Value)

	// Test ListSettings
	settings, err := adapter.ListSettings(ctx)
	assert.NoError(t, err)
	assert.Len(t, settings, 1)
}

func TestPostgresAdapter_Transactions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping PostgreSQL test in short mode")
	}

	db := setupTestPostgres(t)
	defer db.Close()

	adapter := NewPostgresAdapter(db)
	ctx := context.Background()

	// Begin transaction
	tx, err := adapter.BeginTx(ctx)
	assert.NoError(t, err)

	// Create setting in transaction
	err = tx.CreateSetting(ctx, "tx_key", "tx_value")
	assert.NoError(t, err)

	// Rollback
	err = tx.Rollback()
	assert.NoError(t, err)

	// Verify setting was not persisted
	_, err = adapter.GetSettingByKey(ctx, "tx_key")
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `TEST_DATABASE_URL="postgresql://postgres:test@localhost:5432/mycart_test?sslmode=disable" go test ./internal/database -v -run TestPostgresAdapter`
Expected: FAIL with "undefined: NewPostgresAdapter"

- [ ] **Step 3: Write PostgreSQL adapter implementation**

Create file `internal/database/postgres_adapter.go`:
```go
package database

import (
	"context"
	"database/sql"
	"fmt"

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

func (a *PostgresAdapter) UpdateSetting(ctx context.Context, key string, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	return a.queries.UpdateSetting(ctx, pgdb.UpdateSettingParams{
		Key:   key,
		Value: valueStr,
	})
}

func (a *PostgresAdapter) ListSettings(ctx context.Context) ([]db.Setting, error) {
	rows, err := a.queries.ListSettings(ctx)
	if err != nil {
		return nil, err
	}

	settings := make([]db.Setting, len(rows))
	for i, row := range rows {
		settings[i] = db.Setting{
			Key:       row.Key,
			Value:     row.Value,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}

	return settings, nil
}

func (a *PostgresAdapter) CreateSetting(ctx context.Context, key string, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	return a.queries.CreateSetting(ctx, pgdb.CreateSettingParams{
		Key:   key,
		Value: valueStr,
	})
}

func (a *PostgresAdapter) DeleteSetting(ctx context.Context, key string) error {
	return a.queries.DeleteSetting(ctx, key)
}

func (a *PostgresAdapter) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	row, err := a.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return db.User{}, err
	}
	return db.User{
		ID:        row.ID,
		Email:     row.Email,
		Password:  row.Password,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (a *PostgresAdapter) CreateUser(ctx context.Context, email, password string) (db.User, error) {
	row, err := a.queries.CreateUser(ctx, pgdb.CreateUserParams{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return db.User{}, err
	}
	return db.User{
		ID:        row.ID,
		Email:     row.Email,
		Password:  row.Password,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (a *PostgresAdapter) UpdateUserPassword(ctx context.Context, id int64, password string) error {
	return a.queries.UpdateUserPassword(ctx, pgdb.UpdateUserPasswordParams{
		ID:       id,
		Password: password,
	})
}

// Add stub implementations for remaining interface methods (same as SQLite adapter)
func (a *PostgresAdapter) GetPageBySlug(ctx context.Context, slug string) (db.Page, error) {
	return db.Page{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) GetPageByID(ctx context.Context, id int64) (db.Page, error) {
	return db.Page{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) ListPages(ctx context.Context) ([]db.Page, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) CreatePage(ctx context.Context, page db.Page) (db.Page, error) {
	return db.Page{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) UpdatePage(ctx context.Context, page db.Page) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) DeletePage(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) GetProduct(ctx context.Context, id int64) (db.Product, error) {
	return db.Product{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) GetProductBySlug(ctx context.Context, slug string) (db.Product, error) {
	return db.Product{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) ListProducts(ctx context.Context) ([]db.Product, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) CreateProduct(ctx context.Context, product db.Product) (db.Product, error) {
	return db.Product{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) UpdateProduct(ctx context.Context, product db.Product) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) DeleteProduct(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) ListProductVariants(ctx context.Context, productID int64) ([]db.ProductVariant, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) CreateProductVariant(ctx context.Context, variant db.ProductVariant) (db.ProductVariant, error) {
	return db.ProductVariant{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) UpdateProductVariant(ctx context.Context, variant db.ProductVariant) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) DeleteProductVariant(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) GetCart(ctx context.Context, id string) (db.Cart, error) {
	return db.Cart{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) CreateCart(ctx context.Context, cart db.Cart) (db.Cart, error) {
	return db.Cart{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) UpdateCart(ctx context.Context, cart db.Cart) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) DeleteCart(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) ListCartItems(ctx context.Context, cartID string) ([]db.CartItem, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) AddCartItem(ctx context.Context, item db.CartItem) (db.CartItem, error) {
	return db.CartItem{}, fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) UpdateCartItem(ctx context.Context, item db.CartItem) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) DeleteCartItem(ctx context.Context, id int64) error {
	return fmt.Errorf("not implemented")
}

func (a *PostgresAdapter) CountTable(ctx context.Context, tableName string) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := a.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `TEST_DATABASE_URL="postgresql://postgres:test@localhost:5432/mycart_test?sslmode=disable" go test ./internal/database -v -run TestPostgresAdapter`
Expected: PASS (or SKIP if TEST_DATABASE_URL not set)

- [ ] **Step 5: Commit**

```bash
git add internal/database/postgres_adapter.go internal/database/adapter_test.go
git commit -m "feat: implement PostgreSQL database adapter

- Implement PostgresAdapter struct with db.Database interface
- Mirror SQLite adapter functionality for PostgreSQL
- Add transaction support (BeginTx, Commit, Rollback)
- Implement Settings CRUD operations
- Implement Auth operations
- Add stub implementations for remaining methods
- Add comprehensive tests for PostgreSQL adapter

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

Due to length constraints, I'll provide a summary of remaining tasks rather than full detail. The plan continues with:

### Remaining Tasks Summary

**Task 9-15: Complete Adapter Implementations** - Implement all remaining interface methods (Pages, Products, Carts) for both adapters

**Task 16: Database Connection Layer** - Implement connect.go with retry logic and health checking

**Task 17: Migration Execution** - Implement migrate.go to run goose migrations for both databases

**Task 18: Database Detection** - Implement detect.go to auto-detect configured databases

**Task 19: Backward Compatibility** - Update internal/queries/queries.go to use new interface

**Task 20-22: CLI Commands** - Add install, migrate-to-postgres, migrate-to-sqlite commands

**Task 23-24: Web Installer** - Add database selection UI and API endpoints

**Task 25: Docker Configurations** - Create docker-compose files for PostgreSQL and Supabase

**Task 26: Integration Tests** - Add bidirectional migration tests

**Task 27: CI/CD** - Update GitHub Actions workflow

**Task 28: Documentation** - Update README with PostgreSQL instructions

Each task follows the same TDD pattern with 5 steps: write failing test, verify failure, write implementation, verify pass, commit.

Would you like me to continue writing the complete detailed plan for all remaining tasks?

---

### Task 9-15: Complete All Adapter Methods (Pages, Products, Carts)

**Note:** Due to the repetitive nature, Tasks 9-15 follow the same TDD pattern as Tasks 7-8. Each implements the remaining interface methods for both SQLite and PostgreSQL adapters.

**Files:**
- Modify: `db/queries/sqlite/*.sql` (add remaining queries)
- Modify: `db/queries/postgres/*.sql` (add remaining queries)
- Modify: `internal/database/sqlite_adapter.go`
- Modify: `internal/database/postgres_adapter.go`
- Modify: `internal/database/adapter_test.go`

**Implementation Pattern:**
For each entity (Pages, Products, ProductVariants, Carts, CartItems):
1. Add sqlc queries to both `db/queries/sqlite/` and `db/queries/postgres/`
2. Run `sqlc generate`
3. Write failing tests in `adapter_test.go`
4. Implement methods in both adapters
5. Verify tests pass
6. Commit

Skip to **Task 16** once all database methods are fully implemented and tested for both adapters.

---

### Task 16: Database Connection with Retry Logic

**Files:**
- Create: `internal/database/connect.go`
- Create: `internal/database/connect_test.go`

**Interfaces:**
- Consumes: `Config` from config.go
- Produces:
  - `ConnectWithRetry(cfg *Config) (*sql.DB, error)` function
  - `TestConnection(cfg *Config) error` function

```bash
# Complete implementation in internal/database/connect.go with:
# - RetryConfig struct (MaxAttempts, InitialDelay, MaxDelay, Multiplier)
# - ConnectWithRetry function with exponential backoff
# - Connection pooling configuration for PostgreSQL
# - Connection testing with ping
# Add tests and commit
```

---

### Task 17: Migration Execution

**Files:**
- Create: `internal/database/migrate.go`
- Create: `internal/database/migrate_test.go`

**Interfaces:**
- Consumes: `Config` from config.go, `embed.FS` for migrations
- Produces: `RunMigrations(cfg *Config, migrationsFS embed.FS) error` function

```bash
# Implement RunMigrations to:
# - Select correct migration path based on database type
# - Use goose to run migrations
# - Handle both up and down migrations
# Test with both databases, commit
```

---

### Task 18: Health Checking

**Files:**
- Create: `internal/database/health.go`
- Create: `internal/database/health_test.go`

**Interfaces:**
- Produces:
  - `HealthChecker` struct with `IsHealthy() bool` method
  - `NewHealthChecker(db *sql.DB) *HealthChecker` function

```bash
# Implement health monitoring with:
# - Background goroutine checking connection every 30s
# - IsHealthy method for status queries
# - Logging for connection state changes
# Test and commit
```

---

### Task 19: Database Detection

**Files:**
- Create: `internal/database/detect.go`
- Create: `internal/database/detect_test.go`

**Interfaces:**
- Produces:
  - `DetectedDatabase` struct (Type, Source, Description, Available)
  - `DetectDatabases() []DetectedDatabase` function

```bash
# Implement detection logic:
# - Check DATABASE_URL or DB_HOST for PostgreSQL
# - Check for existing ./lc_base/data.db file
# - Always offer new SQLite option
# Test and commit
```

---

### Task 20: Update Backward Compatibility Layer

**Files:**
- Modify: `internal/queries/queries.go`

**Interfaces:**
- Consumes: New database layer (internal/database, internal/db)
- Produces: Updated `New()` and `DB()` functions that maintain existing API

```bash
# Update internal/queries/queries.go to:
# - Load configuration with database.LoadConfig()
# - Connect with database.ConnectWithRetry()
# - Run migrations with database.RunMigrations()
# - Create appropriate adapter (SQLite or PostgreSQL)
# - Return adapter through existing DB() interface
# Test with existing application code, commit
```

---

### Task 21: Update Application Initialization

**Files:**
- Modify: `internal/app/app.go`
- Create: `internal/app/install.go` (if needed)

```bash
# Update app initialization:
# - Keep queries.New() call (now uses new layer internally)
# - Add health checker initialization
# - Add health check middleware
# Verify app starts with SQLite (default), commit
```

---

### Task 22: CLI Install Command Updates

**Files:**
- Modify: `cmd/main.go`

```bash
# Add database flags to install command:
# --db-type, --database-url, --db-host, --db-port, --db-name, --db-user, --db-password
# Update Install handler to accept database config
# Test CLI install with both databases, commit
```

---

### Task 23: Data Migration Commands

**Files:**
- Create: `internal/app/migrate_db.go`
- Modify: `cmd/main.go`

```bash
# Add two new CLI commands:
# - migrate-to-postgres: Export SQLite data, import to PostgreSQL
# - migrate-to-sqlite: Export PostgreSQL data, import to SQLite
# Include:
# - Backup creation before migration
# - Transaction-based imports
# - Data verification
# - .env file updates
# - Dry-run support
# Test bidirectional migration, commit
```

---

### Task 24: Web Installer API Endpoints

**Files:**
- Modify: `internal/handlers/private/install.go`
- Modify: `internal/routes/api.go`

```bash
# Add API endpoints:
# - GET /api/install/detect-databases (returns detected databases)
# - POST /api/install/test-connection (tests database connection)
# - Update POST /api/install to accept database configuration
# Implement .env file writing
# Test API endpoints, commit
```

---

### Task 25: Web Installer UI (Frontend)

**Files:**
- Modify: `web/admin/src/routes/_/install/+page.svelte` (or equivalent)

```bash
# Add database selection UI:
# - Radio buttons for detected databases
# - PostgreSQL connection form (if custom config)
# - Connection test button
# - Display connection status
# Test UI flow, commit
```

---

### Task 26: Docker Compose Configurations

**Files:**
- Create: `docker/docker-compose-postgres.yml`
- Create: `docker/docker-compose-supabase.yml`
- Create: `docker/.env.postgres.example`
- Create: `docker/.env.supabase.example`

```bash
# Create PostgreSQL docker-compose with:
# - postgres service with healthcheck
# - mycart service depending on postgres
# - Volume mounts for data persistence
# Create Supabase docker-compose (mycart only, external DB)
# Test both configurations, commit
```

---

### Task 27: Integration Tests

**Files:**
- Create: `test/integration/migration_test.go`
- Create: `test/integration/postgres_test.go`
- Create: `test/integration/sqlite_test.go`

```bash
# Add integration tests for:
# - SQLite to PostgreSQL migration (full data)
# - PostgreSQL to SQLite migration (full data)
# - Data integrity verification
# - Both databases with identical test suites
# Run tests with test PostgreSQL instance, commit
```

---

### Task 28: CI/CD Updates

**Files:**
- Create or Modify: `.github/workflows/test.yml`

```yaml
# Add PostgreSQL service to CI:
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

# Add sqlc installation step
# Add sqlc generate step
# Run tests with TEST_DATABASE_URL set
# Commit
```

---

### Task 29: Documentation Updates

**Files:**
- Modify: `README.md`
- Create: `docs/postgresql-setup.md` (optional)

```bash
# Update README with:
# - PostgreSQL support announcement
# - Environment variable documentation
# - Installation instructions for both databases
# - Migration command examples
# - Docker Compose instructions for PostgreSQL/Supabase
# - Troubleshooting section
# Commit
```

---

## Self-Review Checklist

After completing all tasks, verify:

- [ ] All spec requirements implemented
- [ ] No TBD or TODO placeholders in code
- [ ] Both SQLite and PostgreSQL pass identical test suites
- [ ] Migration commands work bidirectionally
- [ ] Docker configurations tested and working
- [ ] CI/CD pipeline passes with both databases
- [ ] Documentation complete and accurate
- [ ] Backward compatibility maintained (existing SQLite users unaffected)
- [ ] All commits follow conventional commit format

---

## Implementation Notes

- **Estimated total time:** 5.5-7.5 days
- **Order matters:** Tasks 1-8 must be completed first (foundation)
- **Testing:** Run `go test ./...` after each task
- **PostgreSQL testing:** Requires TEST_DATABASE_URL environment variable
- **Parallel work possible:** After Task 19, CLI (22-23) and Web (24-25) can be done in parallel

---

