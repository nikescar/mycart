# Developer Guide: Making Database Schema Changes

This guide explains the procedure for making changes to the database schema in myCart, which supports both SQLite and PostgreSQL.

## Overview

myCart uses a dual-database architecture with:
- **goose**: Database migration tool
- **sqlc**: Type-safe SQL query code generator
- **Separate migration paths**: One for SQLite, one for PostgreSQL

When you change the database schema, you must:
1. Create migrations for **both** databases
2. Update sqlc query files (if needed)
3. Regenerate type-safe code
4. Test on both databases
5. Commit all changes together

## Quick Reference

```bash
# 1. Create migration files
goose -dir db/migrations/sqlite create migration_name sql
goose -dir db/migrations/postgres create migration_name sql

# 2. Edit migration files (add SQL for both databases)

# 3. Update query files (if needed)
# Edit: db/queries/sqlite/*.sql
# Edit: db/queries/postgres/*.sql

# 4. Generate type-safe code
sqlc generate

# 5. Test migrations
go test ./internal/database -v

# 6. Test application
go test ./...

# 7. Commit everything
git add db/migrations/ db/queries/ internal/db/
git commit -m "feat: add column_name to table_name"
```

## Detailed Procedures

### 1. Adding a New Column

**Example: Add `phone` column to `setting` table**

#### Step 1: Create Migration Files

```bash
# Create timestamped migration files
goose -dir db/migrations/sqlite create add_phone_to_setting sql
goose -dir db/migrations/postgres create add_phone_to_setting sql
```

This creates:
- `db/migrations/sqlite/20260807120000_add_phone_to_setting.sql`
- `db/migrations/postgres/20260807120000_add_phone_to_setting.sql`

#### Step 2: Write SQLite Migration

Edit `db/migrations/sqlite/20260807120000_add_phone_to_setting.sql`:

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE setting ADD COLUMN phone TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE setting DROP COLUMN phone;
-- +goose StatementEnd
```

#### Step 3: Write PostgreSQL Migration

Edit `db/migrations/postgres/20260807120000_add_phone_to_setting.sql`:

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE setting ADD COLUMN phone TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE setting DROP COLUMN phone;
-- +goose StatementEnd
```

#### Step 4: Update sqlc Queries (if needed)

If you need to query the new column, update query files:

**SQLite:** `db/queries/sqlite/settings.sql`
```sql
-- name: GetSettingWithPhone :one
SELECT id, key, value, phone FROM setting WHERE key = ? LIMIT 1;

-- name: UpdateSettingPhone :exec
UPDATE setting SET phone = ? WHERE key = ?;
```

**PostgreSQL:** `db/queries/postgres/settings.sql`
```sql
-- name: GetSettingWithPhone :one
SELECT id, key, value, phone FROM setting WHERE key = $1 LIMIT 1;

-- name: UpdateSettingPhone :exec
UPDATE setting SET phone = $1 WHERE key = $2;
```

#### Step 5: Generate Type-Safe Code

```bash
sqlc generate
```

This updates:
- `internal/db/sqlite/models.go`
- `internal/db/postgres/models.go`
- `internal/db/sqlite/settings.sql.go`
- `internal/db/postgres/settings.sql.go`

#### Step 6: Test Migrations

```bash
# Test SQLite migration
go test ./internal/database -v -run TestRunMigrations_SQLite

# Test PostgreSQL migration (if DATABASE_URL is set)
DATABASE_URL="postgresql://..." go test ./internal/database -v -run TestRunMigrations_PostgreSQL
```

#### Step 7: Commit Changes

```bash
git add db/migrations/ db/queries/ internal/db/
git commit -m "feat: add phone column to setting table

- Add phone column to setting table (both databases)
- Update sqlc queries to include phone field
- Regenerate type-safe code
- Add Up and Down migrations

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

### 2. Modifying an Existing Column

**Example: Change `email` column from TEXT to VARCHAR(255)**

#### Create Migration Files

```bash
goose -dir db/migrations/sqlite create modify_email_column sql
goose -dir db/migrations/postgres create modify_email_column sql
```

#### SQLite Migration

**Note:** SQLite doesn't support ALTER COLUMN directly. You must:
1. Create new table with new schema
2. Copy data
3. Drop old table
4. Rename new table

```sql
-- +goose Up
-- +goose StatementBegin
-- Create new table with updated schema
CREATE TABLE setting_new (
    id     TEXT PRIMARY KEY NOT NULL,
    key    TEXT UNIQUE NOT NULL,
    value  TEXT DEFAULT NULL,
    email  VARCHAR(255) DEFAULT NULL
);

-- Copy existing data
INSERT INTO setting_new (id, key, value)
SELECT id, key, value FROM setting;

-- Drop old table
DROP TABLE setting;

-- Rename new table
ALTER TABLE setting_new RENAME TO setting;

-- Recreate indexes
CREATE INDEX idx_setting_key ON setting (key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Reverse the change (convert back to TEXT)
CREATE TABLE setting_old (
    id     TEXT PRIMARY KEY NOT NULL,
    key    TEXT UNIQUE NOT NULL,
    value  TEXT DEFAULT NULL,
    email  TEXT DEFAULT NULL
);

INSERT INTO setting_old (id, key, value, email)
SELECT id, key, value, email FROM setting;

DROP TABLE setting;
ALTER TABLE setting_old RENAME TO setting;
CREATE INDEX idx_setting_key ON setting (key);
-- +goose StatementEnd
```

#### PostgreSQL Migration

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE setting ALTER COLUMN email TYPE VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE setting ALTER COLUMN email TYPE TEXT;
-- +goose StatementEnd
```

### 3. Removing a Column

**Example: Remove deprecated `old_field` column**

#### SQLite Migration

```sql
-- +goose Up
-- +goose StatementBegin
-- SQLite requires table recreation
CREATE TABLE setting_new (
    id     TEXT PRIMARY KEY NOT NULL,
    key    TEXT UNIQUE NOT NULL,
    value  TEXT DEFAULT NULL
    -- old_field removed
);

INSERT INTO setting_new (id, key, value)
SELECT id, key, value FROM setting;

DROP TABLE setting;
ALTER TABLE setting_new RENAME TO setting;
CREATE INDEX idx_setting_key ON setting (key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE setting_old (
    id     TEXT PRIMARY KEY NOT NULL,
    key    TEXT UNIQUE NOT NULL,
    value  TEXT DEFAULT NULL,
    old_field TEXT DEFAULT NULL
);

INSERT INTO setting_old (id, key, value)
SELECT id, key, value FROM setting;

DROP TABLE setting;
ALTER TABLE setting_old RENAME TO setting;
CREATE INDEX idx_setting_key ON setting (key);
-- +goose StatementEnd
```

#### PostgreSQL Migration

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE setting DROP COLUMN old_field;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE setting ADD COLUMN old_field TEXT DEFAULT NULL;
-- +goose StatementEnd
```

### 4. Adding a New Table

**Example: Add `notifications` table**

#### SQLite Migration

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE notification (
    id         TEXT PRIMARY KEY NOT NULL,
    user_id    TEXT NOT NULL,
    message    TEXT NOT NULL,
    read       BOOLEAN DEFAULT FALSE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(id) ON DELETE CASCADE
);
CREATE INDEX idx_notification_user ON notification (user_id);
CREATE INDEX idx_notification_read ON notification (read);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE notification;
-- +goose StatementEnd
```

#### PostgreSQL Migration

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE notification (
    id         TEXT PRIMARY KEY NOT NULL,
    user_id    TEXT NOT NULL,
    message    TEXT NOT NULL,
    read       BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);
CREATE INDEX idx_notification_user ON notification (user_id);
CREATE INDEX idx_notification_read ON notification (read);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE notification;
-- +goose StatementEnd
```

#### Create Query Files

**SQLite:** `db/queries/sqlite/notifications.sql`
```sql
-- name: GetNotification :one
SELECT id, user_id, message, read, created_at
FROM notification WHERE id = ? LIMIT 1;

-- name: ListUserNotifications :many
SELECT id, user_id, message, read, created_at
FROM notification WHERE user_id = ?
ORDER BY created_at DESC;

-- name: CreateNotification :one
INSERT INTO notification (id, user_id, message)
VALUES (?, ?, ?)
RETURNING id, user_id, message, read, created_at;

-- name: MarkAsRead :exec
UPDATE notification SET read = TRUE WHERE id = ?;
```

**PostgreSQL:** `db/queries/postgres/notifications.sql`
```sql
-- name: GetNotification :one
SELECT id, user_id, message, read, created_at
FROM notification WHERE id = $1 LIMIT 1;

-- name: ListUserNotifications :many
SELECT id, user_id, message, read, created_at
FROM notification WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateNotification :one
INSERT INTO notification (id, user_id, message)
VALUES ($1, $2, $3)
RETURNING id, user_id, message, read, created_at;

-- name: MarkAsRead :exec
UPDATE notification SET read = TRUE WHERE id = $1;
```

## Database-Specific Considerations

### SQLite Syntax

```sql
-- Auto-increment
id INTEGER PRIMARY KEY AUTOINCREMENT

-- Timestamps
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
updated_at DATETIME

-- Placeholders
WHERE id = ?

-- JSON
metadata JSON

-- Text types (all are TEXT)
name TEXT
description TEXT
```

### PostgreSQL Syntax

```sql
-- Auto-increment
id SERIAL PRIMARY KEY
id BIGSERIAL PRIMARY KEY

-- Timestamps
created_at TIMESTAMPTZ DEFAULT NOW()
updated_at TIMESTAMPTZ

-- Placeholders (numbered)
WHERE id = $1 AND name = $2

-- JSON (use JSONB for better performance)
metadata JSONB

-- Text types
name TEXT
description TEXT
email VARCHAR(255)

-- Reserved keywords (must quote)
"user"
"desc"
"order"
```

### Common Differences

| Feature | SQLite | PostgreSQL |
|---------|--------|------------|
| Boolean | `BOOLEAN`, `0/1` | `BOOLEAN`, `true/false` |
| Timestamp | `DATETIME` | `TIMESTAMPTZ` |
| Auto-increment | `AUTOINCREMENT` | `SERIAL`, `BIGSERIAL` |
| JSON | `JSON` | `JSONB` (preferred) |
| Placeholder | `?` | `$1`, `$2`, `$3` |
| Reserved words | No quotes needed | Must quote: `"user"`, `"desc"` |
| Column modification | Requires table recreation | `ALTER TABLE ... ALTER COLUMN` |
| Current time | `CURRENT_TIMESTAMP` | `NOW()` |

## Testing Migrations

### Test SQLite Migration

```bash
# Create test database
sqlite3 test.db < db/migrations/sqlite/20260807120000_your_migration.sql

# Or use Go tests
go test ./internal/database -v -run TestRunMigrations_SQLite
```

### Test PostgreSQL Migration

```bash
# Create test database
createdb mycart_test

# Run migration
psql -d mycart_test -f db/migrations/postgres/20260807120000_your_migration.sql

# Or use Go tests
DATABASE_URL="postgresql://localhost/mycart_test" \
  go test ./internal/database -v -run TestRunMigrations_PostgreSQL

# Clean up
dropdb mycart_test
```

### Test Migration Rollback

```bash
# Test the Down migration
goose -dir db/migrations/sqlite sqlite3 test.db down
goose -dir db/migrations/postgres postgres "postgresql://localhost/mycart_test" down
```

## Common Pitfalls

### ❌ Don't: Modify existing migrations

```sql
-- WRONG: Editing an old migration that's already deployed
-- db/migrations/postgres/20230714135923_init_db.sql
ALTER TABLE setting ADD COLUMN new_field TEXT; -- DON'T DO THIS!
```

✅ **Do:** Create a new migration

```bash
goose -dir db/migrations/postgres create add_new_field sql
```

### ❌ Don't: Forget to update both databases

```bash
# WRONG: Only creating PostgreSQL migration
goose -dir db/migrations/postgres create add_field sql
```

✅ **Do:** Create migrations for both

```bash
goose -dir db/migrations/sqlite create add_field sql
goose -dir db/migrations/postgres create add_field sql
```

### ❌ Don't: Use different column names

```sql
-- SQLite
ALTER TABLE setting ADD COLUMN phone_number TEXT;

-- PostgreSQL (WRONG: different name)
ALTER TABLE setting ADD COLUMN telephone TEXT;
```

✅ **Do:** Use identical column names

```sql
-- Both databases
ALTER TABLE setting ADD COLUMN phone TEXT;
```

### ❌ Don't: Forget to regenerate sqlc

```bash
# Edit migration files
vim db/migrations/*/add_column.sql

# Commit without regenerating
git commit -m "Add column"  # WRONG!
```

✅ **Do:** Regenerate before committing

```bash
sqlc generate
git add db/migrations/ internal/db/
git commit -m "Add column"
```

## Best Practices

### 1. Use Timestamps in Migration Names

```bash
# Good: Includes date/time for ordering
20260807120000_add_phone_column.sql

# Bad: No timestamp
add_phone_column.sql
```

### 2. Write Reversible Migrations

Always implement both `Up` and `Down`:

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE setting ADD COLUMN phone TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE setting DROP COLUMN phone;
-- +goose StatementEnd
```

### 3. Test Before Committing

```bash
# Test both databases
go test ./internal/database -v

# Test full application
go test ./...

# Manual verification
./mycart serve
# Test the feature that uses the new column
```

### 4. Keep Migrations Small

```bash
# Good: One logical change per migration
20260807120000_add_phone_column.sql
20260807120100_add_address_column.sql

# Bad: Multiple unrelated changes
20260807120000_add_lots_of_stuff.sql
```

### 5. Document Complex Migrations

```sql
-- +goose Up
-- +goose StatementBegin
-- Migration to support multi-currency pricing
-- Adds currency column and converts existing prices to USD
-- See: https://github.com/shurco/mycart/issues/123

ALTER TABLE product ADD COLUMN currency TEXT DEFAULT 'USD';
UPDATE product SET currency = 'USD' WHERE currency IS NULL;
-- +goose StatementEnd
```

### 6. Handle Data Migration Carefully

```sql
-- +goose Up
-- +goose StatementBegin
-- Add new column
ALTER TABLE product ADD COLUMN price_cents INTEGER;

-- Migrate existing data (price dollars -> cents)
UPDATE product SET price_cents = CAST(price * 100 AS INTEGER);

-- Make it required after data migration
-- (Do this in a separate migration to ensure safety)
-- +goose StatementEnd
```

## Workflow Checklist

- [ ] Create migration files for **both** SQLite and PostgreSQL
- [ ] Write Up migration SQL
- [ ] Write Down migration SQL (must be reversible)
- [ ] Update sqlc query files (if accessing new columns)
- [ ] Run `sqlc generate`
- [ ] Test SQLite migration locally
- [ ] Test PostgreSQL migration (if possible)
- [ ] Run full test suite
- [ ] Test manually in running application
- [ ] Review generated code in `internal/db/`
- [ ] Commit all changes together (migrations + queries + generated code)
- [ ] Document any special migration considerations

## Troubleshooting

### Migration Fails on SQLite

```bash
# Error: "table setting has no column named phone"

# Check migration was applied
sqlite3 ./lc_base/data.db "SELECT * FROM goose_db_version;"

# Manually run migration
sqlite3 ./lc_base/data.db < db/migrations/sqlite/20260807120000_add_phone.sql
```

### Migration Fails on PostgreSQL

```bash
# Error: column "phone" does not exist

# Check migration status
psql -d mycart -c "SELECT * FROM goose_db_version;"

# Manually run migration
psql -d mycart -f db/migrations/postgres/20260807120000_add_phone.sql
```

### sqlc Generate Fails

```bash
# Error: "syntax error at or near..."

# Check SQL syntax in query files
cat db/queries/postgres/settings.sql

# Common issues:
# - Wrong placeholder (? vs $1)
# - Reserved keyword not quoted
# - Typo in table/column name
```

### Generated Code Doesn't Match

```bash
# If internal/db/ code doesn't reflect your changes:

# 1. Check sqlc.yaml configuration
cat sqlc.yaml

# 2. Delete generated files
rm -rf internal/db/sqlite/*.go internal/db/postgres/*.go

# 3. Regenerate
sqlc generate

# 4. Verify models.go updated
git diff internal/db/*/models.go
```

## Examples from myCart

See these files for real examples:

- `db/migrations/postgres/20230714135923_init_db.sql` - Initial schema
- `db/migrations/postgres/20260418120000_fix_socials.sql` - Adding settings
- `db/migrations/postgres/20231114104637_litepay.sql` - Column changes
- `db/queries/postgres/settings.sql` - Basic queries
- `db/queries/postgres/products.sql` - Complex queries

## Getting Help

If you're unsure about a migration:

1. **Check existing migrations** for similar patterns
2. **Test in a dev database** before committing
3. **Ask for review** in pull request
4. **Reference documentation**:
   - SQLite: https://www.sqlite.org/lang_altertable.html
   - PostgreSQL: https://www.postgresql.org/docs/current/sql-altertable.html
   - sqlc: https://docs.sqlc.dev/
   - goose: https://github.com/pressly/goose

## Summary

**Key principle:** Keep SQLite and PostgreSQL schemas **identical** at all times.

Every database change requires:
1. ✅ SQLite migration
2. ✅ PostgreSQL migration  
3. ✅ Updated sqlc queries (if needed)
4. ✅ Regenerated code (`sqlc generate`)
5. ✅ Tests passing
6. ✅ All committed together

Follow this guide to maintain schema consistency across both databases! 🚀
