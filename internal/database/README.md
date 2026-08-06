# Database Package

This package provides type-safe database access using [sqlc](https://sqlc.dev/).

## Architecture

- **sqlc/**: Generated code from SQL queries (DO NOT EDIT)
- **queries/**: SQL query definitions with sqlc annotations
- **schema/**: SQLite schema representing current database state
- **database.go**: Database wrapper with transaction support

## Generating Code

After modifying queries in `queries/*.sql` or schema in `schema/schema.sql`:

```bash
sqlc generate
```

Note: Use `~/go/bin/sqlc generate` if sqlc is not in your PATH.

## Adding New Queries

1. Add query to appropriate file in `queries/` (e.g., `queries/products.sql`)
2. Use sqlc annotations:
   - `-- name: QueryName :one` for single row
   - `-- name: QueryName :many` for multiple rows
   - `-- name: QueryName :exec` for INSERT/UPDATE/DELETE
3. Run `sqlc generate`
4. Use generated methods via `sqlc.Queries` or `Database` wrapper

## Transaction Example

```go
db := database.New(sqlDB)

err := db.WithTx(ctx, func(q *sqlc.Queries) error {
    if err := q.CreateProduct(ctx, params); err != nil {
        return err
    }
    return q.CreateProductImage(ctx, imageParams)
})
```

## Migration Notes

- Migrated from hand-written `internal/queries` package to sqlc in 2026-08
- Schema in `schema/schema.sql` represents final state after all goose migrations
- Goose migrations in `/migrations` remain unchanged and are still used at runtime
- sqlc uses `schema/` for type generation only

## Status

**Completed:**
- ✅ sqlc configuration and setup
- ✅ Schema extraction from migrations
- ✅ All query definitions (setting, session, auth, pages, products, cart)
- ✅ Type-safe Go code generation
- ✅ Database wrapper with transaction support

**TODO:**
- ⏳ Adapter layer for backward compatibility with existing code
- ⏳ Integration with handlers and routes
- ⏳ Test coverage for generated queries
- ⏳ Migration of existing query usages

## Next Steps

1. Create adapter layer that matches existing `internal/models` types
2. Update handlers to use sqlc queries
3. Run full test suite and fix any integration issues
4. Remove old `internal/queries` code once migration is complete
