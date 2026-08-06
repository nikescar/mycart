# SQLite to sqlc Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate from hand-written SQL queries to type-safe, generated sqlc code for SQLite database access.

**Architecture:** Replace manual *sql.DB query wrappers in `internal/queries/` with sqlc-generated code. Schema remains in existing goose migrations. Query definitions move to `.sql` files with sqlc annotations, generating type-safe Go code in `internal/database/sqlc/`.

**Tech Stack:** sqlc v1.27+, modernc.org/sqlite, database/sql, goose migrations

## Global Constraints

- Go 1.26.1+ required (already in go.mod)
- SQLite engine via modernc.org/sqlite (preserve existing)
- Goose migrations (preserve existing migration files and workflow)
- All existing tests must pass after migration
- No breaking changes to public API (handlers should work unchanged)
- Preserve existing database pragmas and connection settings
- Generate JSON tags for all sqlc structs (for API compatibility)

---

### Task 1: Setup Feature Branch & Install sqlc

**Files:**
- Create: feature branch `feature/migrate-sqlc`
- Verify: `sqlc` CLI installed

**Interfaces:**
- Consumes: main branch HEAD
- Produces: clean feature branch ready for work

- [ ] **Step 1: Create feature branch**

```bash
git checkout -b feature/migrate-sqlc
```

- [ ] **Step 2: Verify sqlc installation**

Run: `sqlc version`
Expected output: version number (e.g., `v1.27.0`)

If not installed, install via:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

- [ ] **Step 3: Verify sqlc is accessible**

Run: `which sqlc`
Expected: path to sqlc binary

- [ ] **Step 4: Commit branch creation**

```bash
git commit --allow-empty -m "chore: create feature branch for sqlc migration

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 2: Create sqlc Configuration

**Files:**
- Create: `sqlc.yaml`

**Interfaces:**
- Consumes: none
- Produces: sqlc.yaml config pointing to migrations (schema) and future queries directory

- [ ] **Step 1: Create sqlc.yaml at project root**

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "internal/database/queries"
    schema: "migrations"
    gen:
      go:
        package: "sqlc"
        out: "internal/database/sqlc"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_pointers_for_null_types: false
```

- [ ] **Step 2: Create queries directory structure**

Run: `mkdir -p internal/database/queries`
Expected: directory created

- [ ] **Step 3: Verify configuration syntax**

Run: `sqlc version`
Expected: no errors (validates sqlc is working)

- [ ] **Step 4: Commit configuration**

```bash
git add sqlc.yaml
git commit -m "feat(db): add sqlc configuration for SQLite

Sets up sqlc to generate type-safe queries from migrations schema.
Output package: internal/database/sqlc

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 3: Convert Setting Queries to sqlc

**Files:**
- Create: `internal/database/queries/setting.sql`
- Reference: `internal/queries/setting.go` for query patterns

**Interfaces:**
- Consumes: setting table schema from migrations/20230714135923_init_db.sql
- Produces: setting.sql with sqlc query definitions for GetSetting, UpdateSetting, ListSettings operations

- [ ] **Step 1: Create setting.sql with basic queries**

```sql
-- name: GetSettingByKey :one
SELECT id, key, value
FROM setting
WHERE key = ?
LIMIT 1;

-- name: GetSettingByID :one
SELECT id, key, value
FROM setting
WHERE id = ?
LIMIT 1;

-- name: UpdateSettingByKey :exec
UPDATE setting
SET value = ?
WHERE key = ?;

-- name: UpdateSettingByID :exec
UPDATE setting
SET value = ?
WHERE id = ?;

-- name: ListAllSettings :many
SELECT id, key, value
FROM setting
ORDER BY key;

-- name: CreateSetting :exec
INSERT INTO setting (id, key, value)
VALUES (?, ?, ?);

-- name: DeleteSettingByKey :exec
DELETE FROM setting
WHERE key = ?;

-- name: BulkUpdateSettings :exec
UPDATE setting
SET value = CASE key
  WHEN ? THEN ?
  WHEN ? THEN ?
  WHEN ? THEN ?
  ELSE value
END
WHERE key IN (?, ?, ?);
```

- [ ] **Step 2: Generate sqlc code**

Run: `sqlc generate`
Expected: Creates internal/database/sqlc/*.go files

- [ ] **Step 3: Verify generated files**

Run: `ls internal/database/sqlc/`
Expected output: db.go, models.go, querier.go, setting.sql.go

- [ ] **Step 4: Check generated code compiles**

Run: `go build ./internal/database/sqlc/`
Expected: no errors

- [ ] **Step 5: Commit setting queries**

```bash
git add internal/database/queries/setting.sql internal/database/sqlc/
git commit -m "feat(db): add sqlc queries for settings table

Generated type-safe query methods for setting CRUD operations.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 4: Convert Session Queries to sqlc

**Files:**
- Create: `internal/database/queries/session.sql`
- Reference: `internal/queries/session.go`

**Interfaces:**
- Consumes: session table schema from migrations
- Produces: session.sql with Get, Set, Delete, Cleanup queries

- [ ] **Step 1: Create session.sql**

```sql
-- name: GetSession :one
SELECT key, value, expires
FROM session
WHERE key = ?
LIMIT 1;

-- name: SetSession :exec
INSERT INTO session (key, value, expires)
VALUES (?, ?, ?)
ON CONFLICT(key) DO UPDATE SET
  value = excluded.value,
  expires = excluded.expires;

-- name: DeleteSession :exec
DELETE FROM session
WHERE key = ?;

-- name: CleanupExpiredSessions :exec
DELETE FROM session
WHERE expires < ?;

-- name: ListActiveSessions :many
SELECT key, value, expires
FROM session
WHERE expires >= ?
ORDER BY expires DESC;
```

- [ ] **Step 2: Generate sqlc code**

Run: `sqlc generate`
Expected: Updates internal/database/sqlc/ with session.sql.go

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/database/sqlc/`
Expected: no errors

- [ ] **Step 4: Commit session queries**

```bash
git add internal/database/queries/session.sql internal/database/sqlc/
git commit -m "feat(db): add sqlc queries for session table

Type-safe session storage with upsert and cleanup operations.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 5: Convert Auth Queries to sqlc

**Files:**
- Create: `internal/database/queries/auth.sql`
- Reference: `internal/queries/auth.go`

**Interfaces:**
- Consumes: setting table (auth stored as settings)
- Produces: auth.sql with authentication-specific setting queries

- [ ] **Step 1: Create auth.sql**

```sql
-- name: GetAuthCredentials :one
SELECT key, value
FROM setting
WHERE key IN ('email', 'password')
LIMIT 2;

-- name: GetEmailSetting :one
SELECT id, key, value
FROM setting
WHERE key = 'email'
LIMIT 1;

-- name: GetPasswordSetting :one
SELECT id, key, value
FROM setting
WHERE key = 'password'
LIMIT 1;

-- name: UpdatePassword :exec
UPDATE setting
SET value = ?
WHERE key = 'password';
```

- [ ] **Step 2: Generate sqlc code**

Run: `sqlc generate`
Expected: Updates internal/database/sqlc/ with auth.sql.go

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/database/sqlc/`
Expected: no errors

- [ ] **Step 4: Commit auth queries**

```bash
git add internal/database/queries/auth.sql internal/database/sqlc/
git commit -m "feat(db): add sqlc queries for authentication

Type-safe auth credential queries.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 6: Convert Pages Queries to sqlc

**Files:**
- Create: `internal/database/queries/pages.sql`
- Reference: `internal/queries/pages.go`

**Interfaces:**
- Consumes: page table schema from migrations
- Produces: pages.sql with ListPages, GetPage, CreatePage, UpdatePage, DeletePage

- [ ] **Step 1: Create pages.sql**

```sql
-- name: ListPages :many
SELECT id, name, slug, content, position, active, created, updated
FROM page
WHERE active = ?
ORDER BY name;

-- name: ListPagesByPosition :many
SELECT id, name, slug, content, position, active, created, updated
FROM page
WHERE position = ? AND active = ?
ORDER BY name;

-- name: GetPageBySlug :one
SELECT id, name, slug, content, position, active, created, updated
FROM page
WHERE slug = ?
LIMIT 1;

-- name: GetPageByID :one
SELECT id, name, slug, content, position, active, created, updated
FROM page
WHERE id = ?
LIMIT 1;

-- name: CreatePage :exec
INSERT INTO page (id, name, slug, content, position, active, created)
VALUES (?, ?, ?, ?, ?, ?, datetime('now'));

-- name: UpdatePage :exec
UPDATE page
SET name = ?,
    content = ?,
    position = ?,
    active = ?,
    updated = datetime('now')
WHERE id = ?;

-- name: DeletePage :exec
DELETE FROM page
WHERE id = ?;

-- name: CheckSlugExists :one
SELECT COUNT(*) as count
FROM page
WHERE slug = ? AND id != ?;
```

- [ ] **Step 2: Generate sqlc code**

Run: `sqlc generate`
Expected: Creates internal/database/sqlc/pages.sql.go

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/database/sqlc/`
Expected: no errors

- [ ] **Step 4: Commit pages queries**

```bash
git add internal/database/queries/pages.sql internal/database/sqlc/
git commit -m "feat(db): add sqlc queries for pages table

Type-safe page CRUD with position filtering and slug validation.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 7: Convert Product Queries to sqlc (Part 1: Core Product)

**Files:**
- Create: `internal/database/queries/products.sql`
- Reference: `internal/queries/products.go` (37KB - largest query file)

**Interfaces:**
- Consumes: product, product_image, digital_file, digital_data table schemas
- Produces: products.sql with core product CRUD operations

- [ ] **Step 1: Create products.sql with basic product queries**

```sql
-- name: GetProductByID :one
SELECT id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE id = ? AND deleted = 0
LIMIT 1;

-- name: GetProductBySlug :one
SELECT id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE slug = ? AND deleted = 0
LIMIT 1;

-- name: ListActiveProducts :many
SELECT id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE active = 1 AND deleted = 0
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: ListAllProducts :many
SELECT id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated
FROM product
WHERE deleted = 0
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: CountActiveProducts :one
SELECT COUNT(*) as count
FROM product
WHERE active = 1 AND deleted = 0;

-- name: CountAllProducts :one
SELECT COUNT(*) as count
FROM product
WHERE deleted = 0;

-- name: CreateProduct :exec
INSERT INTO product (id, name, desc, slug, amount, metadata, attribute, digital, active, created)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'));

-- name: UpdateProduct :exec
UPDATE product
SET name = ?,
    desc = ?,
    amount = ?,
    metadata = ?,
    attribute = ?,
    active = ?,
    updated = datetime('now')
WHERE id = ?;

-- name: SoftDeleteProduct :exec
UPDATE product
SET deleted = 1,
    updated = datetime('now')
WHERE id = ?;

-- name: CheckProductSlugExists :one
SELECT COUNT(*) as count
FROM product
WHERE slug = ? AND id != ? AND deleted = 0;
```

- [ ] **Step 2: Add product image queries to products.sql**

```sql
-- name: GetProductImages :many
SELECT id, product_id, name, ext, orig_name
FROM product_image
WHERE product_id = ?
ORDER BY id;

-- name: GetProductImageByID :one
SELECT id, product_id, name, ext, orig_name
FROM product_image
WHERE id = ?
LIMIT 1;

-- name: CreateProductImage :exec
INSERT INTO product_image (id, product_id, name, ext, orig_name)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteProductImage :exec
DELETE FROM product_image
WHERE id = ?;

-- name: DeleteProductImages :exec
DELETE FROM product_image
WHERE product_id = ?;
```

- [ ] **Step 3: Add digital product queries to products.sql**

```sql
-- name: GetDigitalFiles :many
SELECT id, product_id, name, ext, orig_name
FROM digital_file
WHERE product_id = ?
ORDER BY id;

-- name: CreateDigitalFile :exec
INSERT INTO digital_file (id, product_id, name, ext, orig_name)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteDigitalFile :exec
DELETE FROM digital_file
WHERE id = ?;

-- name: GetDigitalData :one
SELECT id, product_id, content, cart_id
FROM digital_data
WHERE product_id = ? AND (cart_id IS NULL OR cart_id = ?)
LIMIT 1;

-- name: CreateDigitalData :exec
INSERT INTO digital_data (id, product_id, content, cart_id)
VALUES (?, ?, ?, ?);

-- name: UpdateDigitalDataCart :exec
UPDATE digital_data
SET cart_id = ?
WHERE id = ?;
```

- [ ] **Step 4: Generate sqlc code**

Run: `sqlc generate`
Expected: Creates internal/database/sqlc/products.sql.go

- [ ] **Step 5: Verify compilation**

Run: `go build ./internal/database/sqlc/`
Expected: no errors

- [ ] **Step 6: Commit product queries**

```bash
git add internal/database/queries/products.sql internal/database/sqlc/
git commit -m "feat(db): add sqlc queries for products and digital assets

Type-safe product CRUD with images, digital files, and digital data.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 8: Convert Cart Queries to sqlc

**Files:**
- Create: `internal/database/queries/cart.sql`
- Reference: `internal/queries/cart.go` (21KB - complex cart logic)

**Interfaces:**
- Consumes: cart, cart_product table schemas
- Produces: cart.sql with cart and cart_product queries

- [ ] **Step 1: Create cart.sql with cart queries**

```sql
-- name: GetCartByID :one
SELECT id, email, amount_total, currency, payment_id, payment_status, payment_system, created, updated
FROM cart
WHERE id = ?
LIMIT 1;

-- name: ListCarts :many
SELECT id, email, amount_total, currency, payment_id, payment_status, payment_system, created, updated
FROM cart
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: CountCarts :one
SELECT COUNT(*) as count
FROM cart;

-- name: CreateCart :exec
INSERT INTO cart (id, email, amount_total, currency, payment_system, created)
VALUES (?, ?, ?, ?, ?, datetime('now'));

-- name: UpdateCartPayment :exec
UPDATE cart
SET payment_id = ?,
    payment_status = ?,
    updated = datetime('now')
WHERE id = ?;

-- name: UpdateCartTotal :exec
UPDATE cart
SET amount_total = ?,
    updated = datetime('now')
WHERE id = ?;

-- name: DeleteCart :exec
DELETE FROM cart
WHERE id = ?;
```

- [ ] **Step 2: Add cart_product queries to cart.sql**

```sql
-- name: GetCartProducts :many
SELECT id, cart_id, product_id, quantity, amount
FROM cart_product
WHERE cart_id = ?
ORDER BY id;

-- name: GetCartProductByID :one
SELECT id, cart_id, product_id, quantity, amount
FROM cart_product
WHERE id = ?
LIMIT 1;

-- name: CreateCartProduct :exec
INSERT INTO cart_product (id, cart_id, product_id, quantity, amount)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateCartProductQuantity :exec
UPDATE cart_product
SET quantity = ?,
    amount = ?
WHERE id = ?;

-- name: DeleteCartProduct :exec
DELETE FROM cart_product
WHERE id = ?;

-- name: DeleteCartProducts :exec
DELETE FROM cart_product
WHERE cart_id = ?;

-- name: GetCartProductByCartAndProduct :one
SELECT id, cart_id, product_id, quantity, amount
FROM cart_product
WHERE cart_id = ? AND product_id = ?
LIMIT 1;
```

- [ ] **Step 3: Add payment-specific queries to cart.sql**

```sql
-- name: ListCartsWithPaymentStatus :many
SELECT id, email, amount_total, currency, payment_id, payment_status, payment_system, created, updated
FROM cart
WHERE payment_status = ?
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: ListCartsByPaymentSystem :many
SELECT id, email, amount_total, currency, payment_id, payment_status, payment_system, created, updated
FROM cart
WHERE payment_system = ?
ORDER BY created DESC
LIMIT ? OFFSET ?;
```

- [ ] **Step 4: Generate sqlc code**

Run: `sqlc generate`
Expected: Creates internal/database/sqlc/cart.sql.go

- [ ] **Step 5: Verify compilation**

Run: `go build ./internal/database/sqlc/`
Expected: no errors

- [ ] **Step 6: Commit cart queries**

```bash
git add internal/database/queries/cart.sql internal/database/sqlc/
git commit -m "feat(db): add sqlc queries for cart and cart_product tables

Type-safe cart operations with payment tracking and product line items.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 9: Create Database Adapter Layer

**Files:**
- Create: `internal/database/database.go`
- Modify: none yet

**Interfaces:**
- Consumes: internal/database/sqlc package, internal/base/base.go connection logic
- Produces: Database struct wrapping sqlc.Queries with *sql.DB for transactions

- [ ] **Step 1: Create database.go with adapter struct**

```go
package database

import (
	"context"
	"database/sql"

	"github.com/shurco/mycart/internal/database/sqlc"
)

// Database wraps sqlc.Queries and provides transaction support
type Database struct {
	db      *sql.DB
	queries *sqlc.Queries
}

// New creates a new Database instance from an existing *sql.DB connection
func New(db *sql.DB) *Database {
	return &Database{
		db:      db,
		queries: sqlc.New(db),
	}
}

// Queries returns the underlying sqlc.Queries instance
func (d *Database) Queries() *sqlc.Queries {
	return d.queries
}

// DB returns the underlying *sql.DB connection
func (d *Database) DB() *sql.DB {
	return d.db
}

// WithTx executes a function within a database transaction
func (d *Database) WithTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := sqlc.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/database/`
Expected: no errors

- [ ] **Step 3: Commit database adapter**

```bash
git add internal/database/database.go
git commit -m "feat(db): add database adapter layer for sqlc

Wraps sqlc.Queries with transaction support and DB access.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 10: Update Application Initialization

**Files:**
- Modify: `internal/app.go:1-50`
- Reference: current db initialization pattern

**Interfaces:**
- Consumes: internal/base.New(), internal/database.New()
- Produces: Updated app initialization using Database wrapper

- [ ] **Step 1: Add database import to app.go**

Find the imports section and add:
```go
import (
	// ... existing imports
	"github.com/shurco/mycart/internal/database"
)
```

- [ ] **Step 2: Update db field in app struct (if exists)**

Find the app struct definition and verify it has a db field of type `*sql.DB`. The database.Database wrapper will be added to handlers, not the app struct directly.

- [ ] **Step 3: Test compilation**

Run: `go build ./internal/`
Expected: no errors

- [ ] **Step 4: Commit app changes**

```bash
git add internal/app.go
git commit -m "feat(db): add database package import to app

Preparation for sqlc query integration.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 11: Migrate Setting Queries Usage (Test First)

**Files:**
- Modify: `internal/queries/setting_test.go`
- Create: `internal/database/setting_adapter.go` (temporary compatibility layer)

**Interfaces:**
- Consumes: sqlc.Queries methods (GetSettingByKey, UpdateSettingByKey, etc.)
- Produces: Adapter matching old SettingQueries interface for backward compatibility

- [ ] **Step 1: Create setting adapter with old interface**

```go
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/shurco/mycart/internal/database/sqlc"
	"github.com/shurco/mycart/internal/models"
)

// SettingAdapter provides backward-compatible interface for setting queries
type SettingAdapter struct {
	db *Database
}

// NewSettingAdapter creates a new setting adapter
func NewSettingAdapter(db *Database) *SettingAdapter {
	return &SettingAdapter{db: db}
}

// GetSettingByKey retrieves a setting by key and returns it as a map
func (a *SettingAdapter) GetSettingByKey(ctx context.Context, key string) (map[string]struct{ Value any }, error) {
	setting, err := a.db.queries.GetSettingByKey(ctx, key)
	if err != nil {
		return nil, err
	}

	result := map[string]struct{ Value any }{
		key: {Value: setting.Value.String},
	}
	return result, nil
}

// UpdateSettingByKey updates a setting value by key
func (a *SettingAdapter) UpdateSettingByKey(ctx context.Context, key string, value any) error {
	valueStr, err := convertToString(value)
	if err != nil {
		return err
	}

	return a.db.queries.UpdateSettingByKey(ctx, sqlc.UpdateSettingByKeyParams{
		Value: sql.NullString{String: valueStr, Valid: true},
		Key:   key,
	})
}

// convertToString converts various types to string for database storage
func convertToString(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int64:
		return strconv.FormatInt(v.(int64), 10), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}
```

- [ ] **Step 2: Run existing setting tests with adapter**

Run: `go test ./internal/queries/ -run TestSetting -v`
Expected: Tests should reveal what needs to be adapted

- [ ] **Step 3: Update setting_test.go to use new adapter**

Note the test failures and update the test to use SettingAdapter. Document failures for next step.

- [ ] **Step 4: Commit adapter layer**

```bash
git add internal/database/setting_adapter.go
git commit -m "feat(db): add setting adapter for backward compatibility

Temporary adapter layer to migrate setting queries incrementally.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 12: Migrate Setting Queries - Implementation

**Files:**
- Modify: `internal/models/setting.go` (if exists)
- Modify: `internal/handlers/*` (setting-related handlers)

**Interfaces:**
- Consumes: SettingAdapter from Task 11
- Produces: Updated handlers using sqlc-based setting queries

- [ ] **Step 1: Find all setting query usages**

Run: `git grep -n "SettingQueries" internal/`
Expected: List of files using SettingQueries

Document the list for migration.

- [ ] **Step 2: Update first handler using settings**

Pick the simplest handler from the grep results. Replace:
```go
// Old
sq := &queries.SettingQueries{DB: db}
setting, err := sq.GetSettingByKey(ctx, "key")

// New
dbAdapter := database.New(db)
adapter := database.NewSettingAdapter(dbAdapter)
setting, err := adapter.GetSettingByKey(ctx, "key")
```

- [ ] **Step 3: Run handler tests**

Run: `go test ./internal/handlers/ -v`
Expected: Tests for updated handler should pass

- [ ] **Step 4: Update remaining handlers iteratively**

Repeat step 2-3 for each handler using SettingQueries.

- [ ] **Step 5: Commit handler migrations**

```bash
git add internal/handlers/
git commit -m "feat(db): migrate setting queries to sqlc in handlers

Replace manual SettingQueries with sqlc-based adapter.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 13: Migrate Session Queries

**Files:**
- Create: `internal/database/session_adapter.go`
- Modify: handler files using SessionQueries

**Interfaces:**
- Consumes: sqlc session queries (GetSession, SetSession, DeleteSession, CleanupExpiredSessions)
- Produces: SessionAdapter matching old SessionQueries interface

- [ ] **Step 1: Create session adapter**

```go
package database

import (
	"context"
	"database/sql"

	"github.com/shurco/mycart/internal/database/sqlc"
)

// SessionAdapter provides backward-compatible interface for session queries
type SessionAdapter struct {
	db *Database
}

// NewSessionAdapter creates a new session adapter
func NewSessionAdapter(db *Database) *SessionAdapter {
	return &SessionAdapter{db: db}
}

// Get retrieves a session by key
func (a *SessionAdapter) Get(ctx context.Context, key string) (string, error) {
	session, err := a.db.queries.GetSession(ctx, key)
	if err != nil {
		return "", err
	}
	return session.Value.String, nil
}

// Set stores or updates a session
func (a *SessionAdapter) Set(ctx context.Context, key, value string, expires int64) error {
	return a.db.queries.SetSession(ctx, sqlc.SetSessionParams{
		Key:     key,
		Value:   sql.NullString{String: value, Valid: true},
		Expires: sql.NullInt64{Int64: expires, Valid: true},
	})
}

// Delete removes a session by key
func (a *SessionAdapter) Delete(ctx context.Context, key string) error {
	return a.db.queries.DeleteSession(ctx, key)
}

// Cleanup removes expired sessions
func (a *SessionAdapter) Cleanup(ctx context.Context, now int64) error {
	return a.db.queries.CleanupExpiredSessions(ctx, sql.NullInt64{Int64: now, Valid: true})
}
```

- [ ] **Step 2: Find session query usages**

Run: `git grep -n "SessionQueries" internal/`
Expected: List of files to update

- [ ] **Step 3: Update code to use SessionAdapter**

Replace SessionQueries instantiation with SessionAdapter.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/... -run Session -v`
Expected: All session tests pass

- [ ] **Step 5: Commit session migration**

```bash
git add internal/database/session_adapter.go internal/
git commit -m "feat(db): migrate session queries to sqlc

Replace SessionQueries with sqlc-based SessionAdapter.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 14: Migrate Auth Queries

**Files:**
- Create: `internal/database/auth_adapter.go`
- Modify: authentication-related handlers

**Interfaces:**
- Consumes: sqlc auth queries and setting queries
- Produces: AuthAdapter for authentication operations

- [ ] **Step 1: Create auth adapter**

```go
package database

import (
	"context"
	"errors"

	"github.com/shurco/mycart/internal/database/sqlc"
	"golang.org/x/crypto/bcrypt"
)

// AuthAdapter provides authentication query operations
type AuthAdapter struct {
	db *Database
}

// NewAuthAdapter creates a new auth adapter
func NewAuthAdapter(db *Database) *AuthAdapter {
	return &AuthAdapter{db: db}
}

// VerifyCredentials checks if email and password match
func (a *AuthAdapter) VerifyCredentials(ctx context.Context, email, password string) (bool, error) {
	emailSetting, err := a.db.queries.GetEmailSetting(ctx)
	if err != nil {
		return false, err
	}

	if emailSetting.Value.String != email {
		return false, nil
	}

	passwordSetting, err := a.db.queries.GetPasswordSetting(ctx)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordSetting.Value.String), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// UpdatePassword updates the hashed password
func (a *AuthAdapter) UpdatePassword(ctx context.Context, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return a.db.queries.UpdatePassword(ctx, sqlc.NullString{
		String: string(hashedPassword),
		Valid:  true,
	})
}
```

- [ ] **Step 2: Find auth query usages**

Run: `git grep -n "AuthQueries" internal/`
Expected: List of files using auth queries

- [ ] **Step 3: Update handlers to use AuthAdapter**

Replace AuthQueries with AuthAdapter in authentication handlers.

- [ ] **Step 4: Run auth tests**

Run: `go test ./internal/... -run Auth -v`
Expected: All auth tests pass

- [ ] **Step 5: Commit auth migration**

```bash
git add internal/database/auth_adapter.go internal/
git commit -m "feat(db): migrate auth queries to sqlc

Replace AuthQueries with sqlc-based AuthAdapter with bcrypt support.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 15: Migrate Pages Queries

**Files:**
- Create: `internal/database/pages_adapter.go`
- Modify: page-related handlers

**Interfaces:**
- Consumes: sqlc pages queries
- Produces: PagesAdapter with CRUD operations for pages

- [ ] **Step 1: Create pages adapter**

```go
package database

import (
	"context"
	"database/sql"

	"github.com/shurco/mycart/internal/database/sqlc"
	"github.com/shurco/mycart/internal/models"
)

// PagesAdapter provides page query operations
type PagesAdapter struct {
	db *Database
}

// NewPagesAdapter creates a new pages adapter
func NewPagesAdapter(db *Database) *PagesAdapter {
	return &PagesAdapter{db: db}
}

// ListActive retrieves all active pages
func (a *PagesAdapter) ListActive(ctx context.Context) ([]*models.Page, error) {
	pages, err := a.db.queries.ListPages(ctx, sql.NullBool{Bool: true, Valid: true})
	if err != nil {
		return nil, err
	}

	result := make([]*models.Page, len(pages))
	for i, p := range pages {
		result[i] = &models.Page{
			ID:       p.ID,
			Name:     p.Name,
			Slug:     p.Slug,
			Content:  p.Content.String,
			Position: p.Position,
			Active:   p.Active.Bool,
			Created:  p.Created.Time,
			Updated:  p.Updated.Time,
		}
	}
	return result, nil
}

// GetBySlug retrieves a page by slug
func (a *PagesAdapter) GetBySlug(ctx context.Context, slug string) (*models.Page, error) {
	p, err := a.db.queries.GetPageBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return &models.Page{
		ID:       p.ID,
		Name:     p.Name,
		Slug:     p.Slug,
		Content:  p.Content.String,
		Position: p.Position,
		Active:   p.Active.Bool,
		Created:  p.Created.Time,
		Updated:  p.Updated.Time,
	}, nil
}

// Create creates a new page
func (a *PagesAdapter) Create(ctx context.Context, page *models.Page) error {
	return a.db.queries.CreatePage(ctx, sqlc.CreatePageParams{
		ID:       page.ID,
		Name:     page.Name,
		Slug:     page.Slug,
		Content:  sql.NullString{String: page.Content, Valid: true},
		Position: page.Position,
		Active:   sql.NullBool{Bool: page.Active, Valid: true},
	})
}

// Update updates an existing page
func (a *PagesAdapter) Update(ctx context.Context, page *models.Page) error {
	return a.db.queries.UpdatePage(ctx, sqlc.UpdatePageParams{
		Name:     page.Name,
		Content:  sql.NullString{String: page.Content, Valid: true},
		Position: page.Position,
		Active:   sql.NullBool{Bool: page.Active, Valid: true},
		ID:       page.ID,
	})
}

// Delete deletes a page
func (a *PagesAdapter) Delete(ctx context.Context, id string) error {
	return a.db.queries.DeletePage(ctx, id)
}
```

- [ ] **Step 2: Find pages query usages**

Run: `git grep -n "PagesQueries\|PageQueries" internal/`
Expected: List of handler files using page queries

- [ ] **Step 3: Update handlers to use PagesAdapter**

Replace old queries with PagesAdapter methods.

- [ ] **Step 4: Run pages tests**

Run: `go test ./internal/... -run Page -v`
Expected: All page tests pass

- [ ] **Step 5: Commit pages migration**

```bash
git add internal/database/pages_adapter.go internal/
git commit -m "feat(db): migrate pages queries to sqlc

Replace manual page queries with sqlc-based PagesAdapter.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 16: Migrate Product Queries (Part 1: Basic CRUD)

**Files:**
- Create: `internal/database/products_adapter.go`
- Reference: `internal/queries/products.go` for complex logic

**Interfaces:**
- Consumes: sqlc product queries
- Produces: ProductsAdapter with basic CRUD (full complex queries in Part 2)

- [ ] **Step 1: Create products adapter skeleton**

```go
package database

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/shurco/mycart/internal/database/sqlc"
	"github.com/shurco/mycart/internal/models"
)

// ProductsAdapter provides product query operations
type ProductsAdapter struct {
	db *Database
}

// NewProductsAdapter creates a new products adapter
func NewProductsAdapter(db *Database) *ProductsAdapter {
	return &ProductsAdapter{db: db}
}

// GetByID retrieves a product by ID
func (a *ProductsAdapter) GetByID(ctx context.Context, id string) (*models.Product, error) {
	p, err := a.db.queries.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return a.convertToModel(p)
}

// GetBySlug retrieves a product by slug
func (a *ProductsAdapter) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	p, err := a.db.queries.GetProductBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	return a.convertToModel(p)
}

// ListActive retrieves active products with pagination
func (a *ProductsAdapter) ListActive(ctx context.Context, limit, offset int64) ([]*models.Product, error) {
	products, err := a.db.queries.ListActiveProducts(ctx, sqlc.ListActiveProductsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return a.convertToModelSlice(products)
}

// Create creates a new product
func (a *ProductsAdapter) Create(ctx context.Context, product *models.Product) error {
	metadataJSON, err := json.Marshal(product.Metadata)
	if err != nil {
		return err
	}

	attributeJSON, err := json.Marshal(product.Attribute)
	if err != nil {
		return err
	}

	return a.db.queries.CreateProduct(ctx, sqlc.CreateProductParams{
		ID:        product.ID,
		Name:      product.Name,
		Desc:      product.Desc,
		Slug:      product.Slug,
		Amount:    product.Amount,
		Metadata:  string(metadataJSON),
		Attribute: string(attributeJSON),
		Digital:   sql.NullString{String: product.Digital, Valid: product.Digital != ""},
		Active:    sql.NullBool{Bool: product.Active, Valid: true},
	})
}

// convertToModel converts sqlc.Product to models.Product
func (a *ProductsAdapter) convertToModel(p sqlc.Product) (*models.Product, error) {
	var metadata []map[string]string
	if err := json.Unmarshal([]byte(p.Metadata), &metadata); err != nil {
		return nil, err
	}

	var attribute []map[string]any
	if err := json.Unmarshal([]byte(p.Attribute), &attribute); err != nil {
		return nil, err
	}

	return &models.Product{
		ID:        p.ID,
		Name:      p.Name,
		Desc:      p.Desc,
		Slug:      p.Slug,
		Amount:    p.Amount,
		Metadata:  metadata,
		Attribute: attribute,
		Digital:   p.Digital.String,
		Active:    p.Active.Bool,
		Deleted:   p.Deleted.Bool,
		Created:   p.Created.Time,
		Updated:   p.Updated.Time,
	}, nil
}

// convertToModelSlice converts []sqlc.Product to []*models.Product
func (a *ProductsAdapter) convertToModelSlice(products []sqlc.Product) ([]*models.Product, error) {
	result := make([]*models.Product, len(products))
	for i, p := range products {
		converted, err := a.convertToModel(p)
		if err != nil {
			return nil, err
		}
		result[i] = converted
	}
	return result, nil
}
```

- [ ] **Step 2: Add product image operations**

```go
// GetImages retrieves all images for a product
func (a *ProductsAdapter) GetImages(ctx context.Context, productID string) ([]*models.ProductImage, error) {
	images, err := a.db.queries.GetProductImages(ctx, productID)
	if err != nil {
		return nil, err
	}

	result := make([]*models.ProductImage, len(images))
	for i, img := range images {
		result[i] = &models.ProductImage{
			ID:        img.ID,
			ProductID: img.ProductID,
			Name:      img.Name,
			Ext:       img.Ext,
			OrigName:  img.OrigName,
		}
	}
	return result, nil
}

// CreateImage creates a new product image
func (a *ProductsAdapter) CreateImage(ctx context.Context, image *models.ProductImage) error {
	return a.db.queries.CreateProductImage(ctx, sqlc.CreateProductImageParams{
		ID:        image.ID,
		ProductID: image.ProductID,
		Name:      image.Name,
		Ext:       image.Ext,
		OrigName:  image.OrigName,
	})
}
```

- [ ] **Step 3: Test basic product operations**

Run: `go test ./internal/database/ -v`
Expected: Adapter compiles and basic tests pass

- [ ] **Step 4: Commit basic product adapter**

```bash
git add internal/database/products_adapter.go
git commit -m "feat(db): add basic product adapter with CRUD operations

Type-safe product queries with images support. Complex list queries in next task.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 17: Migrate Cart Queries

**Files:**
- Create: `internal/database/cart_adapter.go`
- Reference: `internal/queries/cart.go` for payment logic

**Interfaces:**
- Consumes: sqlc cart and cart_product queries
- Produces: CartAdapter with cart and line item operations

- [ ] **Step 1: Create cart adapter**

```go
package database

import (
	"context"
	"database/sql"

	"github.com/shurco/mycart/internal/database/sqlc"
	"github.com/shurco/mycart/internal/models"
)

// CartAdapter provides cart query operations
type CartAdapter struct {
	db *Database
}

// NewCartAdapter creates a new cart adapter
func NewCartAdapter(db *Database) *CartAdapter {
	return &CartAdapter{db: db}
}

// GetByID retrieves a cart by ID
func (a *CartAdapter) GetByID(ctx context.Context, id string) (*models.Cart, error) {
	c, err := a.db.queries.GetCartByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.Cart{
		ID:            c.ID,
		Email:         c.Email,
		AmountTotal:   c.AmountTotal.Int64,
		Currency:      c.Currency.String,
		PaymentID:     c.PaymentID.String,
		PaymentStatus: c.PaymentStatus.String,
		PaymentSystem: c.PaymentSystem.String,
		Created:       c.Created.Time,
		Updated:       c.Updated.Time,
	}, nil
}

// List retrieves carts with pagination
func (a *CartAdapter) List(ctx context.Context, limit, offset int64) ([]*models.Cart, int, error) {
	carts, err := a.db.queries.ListCarts(ctx, sqlc.ListCartsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	countRow, err := a.db.queries.CountCarts(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*models.Cart, len(carts))
	for i, c := range carts {
		result[i] = &models.Cart{
			ID:            c.ID,
			Email:         c.Email,
			AmountTotal:   c.AmountTotal.Int64,
			Currency:      c.Currency.String,
			PaymentID:     c.PaymentID.String,
			PaymentStatus: c.PaymentStatus.String,
			PaymentSystem: c.PaymentSystem.String,
			Created:       c.Created.Time,
			Updated:       c.Updated.Time,
		}
	}

	return result, int(countRow), nil
}

// Create creates a new cart
func (a *CartAdapter) Create(ctx context.Context, cart *models.Cart) error {
	return a.db.queries.CreateCart(ctx, sqlc.CreateCartParams{
		ID:            cart.ID,
		Email:         cart.Email,
		AmountTotal:   sql.NullInt64{Int64: cart.AmountTotal, Valid: true},
		Currency:      sql.NullString{String: cart.Currency, Valid: true},
		PaymentSystem: sql.NullString{String: cart.PaymentSystem, Valid: true},
	})
}

// UpdatePayment updates cart payment information
func (a *CartAdapter) UpdatePayment(ctx context.Context, id, paymentID, status string) error {
	return a.db.queries.UpdateCartPayment(ctx, sqlc.UpdateCartPaymentParams{
		PaymentID:     sql.NullString{String: paymentID, Valid: true},
		PaymentStatus: sql.NullString{String: status, Valid: true},
		ID:            id,
	})
}

// GetProducts retrieves all products in a cart
func (a *CartAdapter) GetProducts(ctx context.Context, cartID string) ([]*models.CartProduct, error) {
	products, err := a.db.queries.GetCartProducts(ctx, cartID)
	if err != nil {
		return nil, err
	}

	result := make([]*models.CartProduct, len(products))
	for i, p := range products {
		result[i] = &models.CartProduct{
			ID:        p.ID,
			CartID:    p.CartID,
			ProductID: p.ProductID,
			Quantity:  int(p.Quantity.Int64),
			Amount:    p.Amount.Int64,
		}
	}
	return result, nil
}

// AddProduct adds a product to a cart
func (a *CartAdapter) AddProduct(ctx context.Context, cp *models.CartProduct) error {
	return a.db.queries.CreateCartProduct(ctx, sqlc.CreateCartProductParams{
		ID:        cp.ID,
		CartID:    cp.CartID,
		ProductID: cp.ProductID,
		Quantity:  sql.NullInt64{Int64: int64(cp.Quantity), Valid: true},
		Amount:    sql.NullInt64{Int64: cp.Amount, Valid: true},
	})
}
```

- [ ] **Step 2: Find cart query usages**

Run: `git grep -n "CartQueries" internal/`
Expected: List of handlers using cart queries

- [ ] **Step 3: Update handlers to use CartAdapter**

Replace CartQueries with CartAdapter.

- [ ] **Step 4: Run cart tests**

Run: `go test ./internal/... -run Cart -v`
Expected: Cart tests pass

- [ ] **Step 5: Commit cart migration**

```bash
git add internal/database/cart_adapter.go internal/
git commit -m "feat(db): migrate cart queries to sqlc

Replace CartQueries with sqlc-based CartAdapter for cart and line items.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 18: Run Full Test Suite

**Files:**
- None modified
- Verify: all tests pass with new sqlc queries

**Interfaces:**
- Consumes: all migrated adapters
- Produces: passing test suite confirmation

- [ ] **Step 1: Run all internal tests**

Run: `go test ./internal/... -v`
Expected: All tests pass (may need adapter fixes)

- [ ] **Step 2: Run handler tests**

Run: `go test ./internal/handlers/... -v`
Expected: All handler tests pass

- [ ] **Step 3: Run integration tests if present**

Run: `go test ./internal/integration/... -v`
Expected: Integration tests pass

- [ ] **Step 4: Check for test coverage gaps**

Run: `go test ./internal/database/... -cover`
Expected: Coverage report for new adapters

- [ ] **Step 5: Document test results**

Create a simple summary of test results in commit message.

- [ ] **Step 6: Fix any failing tests**

If tests fail, identify the adapter or query issue and fix it. Re-run tests until all pass.

- [ ] **Step 7: Commit test fixes if needed**

```bash
git add internal/
git commit -m "fix(db): resolve failing tests after sqlc migration

All internal and handler tests now passing with sqlc adapters.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 19: Remove Old Query Code

**Files:**
- Delete: `internal/queries/*.go` (except tests to be migrated)
- Verify: no imports of old queries package

**Interfaces:**
- Consumes: verified working sqlc adapters
- Produces: clean codebase without old query code

- [ ] **Step 1: Verify no usages of old queries**

Run: `git grep -n "github.com/shurco/mycart/internal/queries" --and --not -e "_test.go"`
Expected: No matches (only test files if any)

- [ ] **Step 2: Remove old query implementation files**

```bash
rm internal/queries/auth.go \
   internal/queries/cart.go \
   internal/queries/install.go \
   internal/queries/pages.go \
   internal/queries/products.go \
   internal/queries/queries.go \
   internal/queries/session.go \
   internal/queries/setting.go
```

Run: `git status`
Expected: Deleted files listed

- [ ] **Step 3: Verify project still compiles**

Run: `go build ./...`
Expected: No errors

- [ ] **Step 4: Run tests again**

Run: `go test ./... -v`
Expected: All tests still pass

- [ ] **Step 5: Commit removal**

```bash
git add -u
git commit -m "chore(db): remove old manual query implementations

All queries migrated to sqlc. Old query files no longer needed.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 20: Migrate or Remove Old Query Tests

**Files:**
- Evaluate: `internal/queries/*_test.go`
- Decision: Migrate valuable tests to adapter tests or remove if redundant

**Interfaces:**
- Consumes: old query tests
- Produces: adapter tests or removed redundant tests

- [ ] **Step 1: Evaluate test coverage**

Run: `ls internal/queries/*_test.go`
Expected: List of test files to evaluate

- [ ] **Step 2: Review each test file**

For each test file, determine if:
- Tests are redundant (already covered by handler tests)
- Tests should be migrated to internal/database/*_adapter_test.go
- Tests are integration tests that should remain

- [ ] **Step 3: Migrate valuable tests to adapter tests**

Example: Create `internal/database/setting_adapter_test.go` if setting_test.go has unique test cases.

- [ ] **Step 4: Remove redundant test files**

```bash
rm internal/queries/*_test.go  # After migration or verification
```

- [ ] **Step 5: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass, no decrease in coverage

- [ ] **Step 6: Commit test migration/removal**

```bash
git add internal/
git commit -m "test(db): migrate query tests to adapter tests

Migrated valuable test cases to adapter tests. Removed redundant tests.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 21: Update Documentation

**Files:**
- Modify: `README.md` (if database section exists)
- Create: `internal/database/README.md` (architecture doc)

**Interfaces:**
- Consumes: completed migration
- Produces: documentation for sqlc usage

- [ ] **Step 1: Create database package README**

```markdown
# Database Package

This package provides type-safe database access using [sqlc](https://sqlc.dev/).

## Architecture

- **sqlc/**: Generated code from SQL queries (DO NOT EDIT)
- **queries/**: SQL query definitions with sqlc annotations
- **database.go**: Database wrapper with transaction support
- **\*_adapter.go**: Backward-compatible adapters for gradual migration

## Generating Code

After modifying queries in `queries/*.sql` or schema in `migrations/*.sql`:

\`\`\`bash
sqlc generate
\`\`\`

## Adding New Queries

1. Add query to appropriate file in `queries/` (e.g., `queries/products.sql`)
2. Use sqlc annotations:
   - `-- name: QueryName :one` for single row
   - `-- name: QueryName :many` for multiple rows
   - `-- name: QueryName :exec` for INSERT/UPDATE/DELETE
3. Run `sqlc generate`
4. Use generated methods via adapter or directly

## Transaction Example

\`\`\`go
db := database.New(sqlDB)

err := db.WithTx(ctx, func(q *sqlc.Queries) error {
    if err := q.CreateProduct(ctx, params); err != nil {
        return err
    }
    return q.CreateProductImage(ctx, imageParams)
})
\`\`\`

## Migration Notes

Migrated from hand-written `internal/queries` package to sqlc in 2026-08.
Adapters provide backward compatibility during transition period.
```

- [ ] **Step 2: Write README file**

Write the above content to `internal/database/README.md`.

- [ ] **Step 3: Update main README if database section exists**

Check if README.md mentions database access. If yes, add note about sqlc.

- [ ] **Step 4: Commit documentation**

```bash
git add internal/database/README.md README.md
git commit -m "docs(db): add database package documentation

Document sqlc architecture, code generation, and usage patterns.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 22: Final Integration Test & Cleanup

**Files:**
- Verify: entire application works end-to-end
- Clean: remove any temporary adapter code if no longer needed

**Interfaces:**
- Consumes: completed migration
- Produces: production-ready sqlc integration

- [ ] **Step 1: Build the application**

Run: `go build ./cmd/...`
Expected: Binary builds successfully

- [ ] **Step 2: Run application locally**

Run the application with a test database and verify:
- Database connection works
- Migrations run successfully
- Basic CRUD operations work (create product, cart, etc.)

- [ ] **Step 3: Run full test suite**

Run: `go test ./... -race -v`
Expected: All tests pass with race detector

- [ ] **Step 4: Check for unused adapters**

Review adapter files. If handlers now use sqlc directly (no backward compat needed), adapters can be removed or noted for future removal.

- [ ] **Step 5: Verify no old imports remain**

Run: `git grep "internal/queries"` 
Expected: No imports of old queries package

- [ ] **Step 6: Final cleanup commit if needed**

```bash
git add .
git commit -m "chore(db): final sqlc migration cleanup

Application fully migrated to sqlc. All tests passing.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

### Task 23: Create Pull Request

**Files:**
- None modified
- Action: Create PR for review

**Interfaces:**
- Consumes: completed feature branch
- Produces: Pull request ready for review

- [ ] **Step 1: Push feature branch**

Run: `git push -u origin feature/migrate-sqlc`
Expected: Branch pushed to remote

- [ ] **Step 2: Create pull request**

Run: 
```bash
gh pr create --title "Migrate SQLite queries to sqlc" --body "$(cat <<'EOF'
## Summary

- Migrated from hand-written SQL queries in `internal/queries/` to type-safe sqlc-generated code
- Created `internal/database/` package with sqlc-generated queries and adapters
- All existing tests passing
- No breaking changes to handlers or public API
- Database schema unchanged (reused existing goose migrations)

## Changes

- Added `sqlc.yaml` configuration for SQLite
- Created SQL query definitions in `internal/database/queries/`
- Generated type-safe Go code in `internal/database/sqlc/`
- Added adapter layer for backward compatibility
- Removed old `internal/queries/*.go` files (except framework)
- Updated all handlers to use sqlc queries
- Added database package documentation

## Test Plan

- [x] All existing unit tests pass
- [x] All handler tests pass
- [x] All integration tests pass (if present)
- [x] Manual testing: CRUD operations work
- [x] Race detector: no issues found
- [x] Build succeeds: binary compiles

## Migration Benefits

- **Type Safety**: Compile-time checking of SQL queries
- **Maintainability**: SQL and Go code separated, easier to review
- **Performance**: Same performance, no ORM overhead
- **Reliability**: Fewer runtime SQL errors, caught at compile time

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Expected: PR created and URL returned

- [ ] **Step 3: Add labels and reviewers**

If using GitHub, add relevant labels (e.g., `database`, `refactor`, `enhancement`).

- [ ] **Step 4: Note PR URL**

Document the PR URL for tracking.

---

## Plan Complete

This plan migrates the mycart SQLite database access from hand-written queries to sqlc-generated type-safe code. Each task is atomic and testable. The migration preserves existing behavior while improving code quality and maintainability.

**Key Decisions:**
- Keep existing goose migrations (no schema changes)
- Create adapters for backward compatibility during migration
- Migrate incrementally (setting → session → auth → pages → products → cart)
- Remove old code only after all tests pass
- Use sqlc v2 config format with JSON tags for API compatibility

**Risk Mitigation:**
- Work in feature branch (easy rollback)
- Test after each component migration
- Preserve existing test suite
- Adapters allow incremental migration
- No database schema changes required
