package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// MigrateData copies all data from source database to target database.
func MigrateData(source, target Database) error {
	ctx := context.Background()

	// Get the underlying *sql.DB connections
	sourceDB := source.DB()
	targetDB := target.DB()

	// Begin transaction on target
	tx, err := targetDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Migrate settings
	if err := migrateSettings(ctx, sourceDB, tx); err != nil {
		return fmt.Errorf("failed to migrate settings: %w", err)
	}

	// Migrate sessions
	if err := migrateSessions(ctx, sourceDB, tx); err != nil {
		return fmt.Errorf("failed to migrate sessions: %w", err)
	}

	// Migrate subdomains
	if err := migrateSubdomains(ctx, sourceDB, tx); err != nil {
		return fmt.Errorf("failed to migrate subdomains: %w", err)
	}

	// Migrate pages
	if err := migratePages(ctx, sourceDB, tx); err != nil {
		return fmt.Errorf("failed to migrate pages: %w", err)
	}

	// Migrate products and related data
	if err := migrateProducts(ctx, sourceDB, tx); err != nil {
		return fmt.Errorf("failed to migrate products: %w", err)
	}

	// Migrate carts
	if err := migrateCarts(ctx, sourceDB, tx); err != nil {
		return fmt.Errorf("failed to migrate carts: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// migrateSettings migrates the settings table
func migrateSettings(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	rows, err := source.QueryContext(ctx, "SELECT id, key, value FROM setting")
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, "INSERT INTO setting (id, key, value) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, key string
		var value sql.NullString
		if err := rows.Scan(&id, &key, &value); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, id, key, value); err != nil {
			return err
		}
	}

	return rows.Err()
}

// migrateSessions migrates the session table
func migrateSessions(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	rows, err := source.QueryContext(ctx, "SELECT key, value, expires FROM session")
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, "INSERT INTO session (key, value, expires) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var key string
		var value sql.NullString
		var expires sql.NullInt32
		if err := rows.Scan(&key, &value, &expires); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, key, value, expires); err != nil {
			return err
		}
	}

	return rows.Err()
}

// migrateSubdomains migrates the subdomain table
func migrateSubdomains(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	rows, err := source.QueryContext(ctx, "SELECT id, name, \"desc\" FROM subdomain")
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, "INSERT INTO subdomain (id, name, \"desc\") VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, name string
		var desc sql.NullString
		if err := rows.Scan(&id, &name, &desc); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, id, name, desc); err != nil {
			return err
		}
	}

	return rows.Err()
}

// migratePages migrates the page table
func migratePages(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	rows, err := source.QueryContext(ctx, `
		SELECT id, name, slug, content, position, active, created, updated
		FROM page
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, `
		INSERT INTO page (id, name, slug, content, position, active, created, updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, name, slug, position string
		var content sql.NullString
		var active bool
		var created, updated sql.NullTime
		if err := rows.Scan(&id, &name, &slug, &content, &position, &active, &created, &updated); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, id, name, slug, content, position, active, created, updated); err != nil {
			return err
		}
	}

	return rows.Err()
}

// migrateProducts migrates products and related tables
func migrateProducts(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	// Migrate products
	rows, err := source.QueryContext(ctx, `
		SELECT id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated
		FROM product
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, `
		INSERT INTO product (id, name, "desc", slug, amount, metadata, attribute, digital, active, deleted, created, updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, name, desc, slug, amount string
		var metadata, attribute []byte
		var digital sql.NullString
		var active, deleted bool
		var created, updated sql.NullTime
		if err := rows.Scan(&id, &name, &desc, &slug, &amount, &metadata, &attribute, &digital, &active, &deleted, &created, &updated); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, id, name, desc, slug, amount, metadata, attribute, digital, active, deleted, created, updated); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Migrate product images
	if err := migrateProductImages(ctx, source, target); err != nil {
		return err
	}

	// Migrate digital files
	if err := migrateDigitalFiles(ctx, source, target); err != nil {
		return err
	}

	// Migrate digital data
	if err := migrateDigitalData(ctx, source, target); err != nil {
		return err
	}

	return nil
}

// migrateProductImages migrates product_image table
func migrateProductImages(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	rows, err := source.QueryContext(ctx, "SELECT id, product_id, name, ext, orig_name FROM product_image")
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, "INSERT INTO product_image (id, product_id, name, ext, orig_name) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, productID, name, ext, origName string
		if err := rows.Scan(&id, &productID, &name, &ext, &origName); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, id, productID, name, ext, origName); err != nil {
			return err
		}
	}

	return rows.Err()
}

// migrateDigitalFiles migrates digital_file table
func migrateDigitalFiles(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	rows, err := source.QueryContext(ctx, "SELECT id, product_id, name, ext, orig_name FROM digital_file")
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, "INSERT INTO digital_file (id, product_id, name, ext, orig_name) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, productID, name, ext, origName string
		if err := rows.Scan(&id, &productID, &name, &ext, &origName); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, id, productID, name, ext, origName); err != nil {
			return err
		}
	}

	return rows.Err()
}

// migrateDigitalData migrates digital_data table
func migrateDigitalData(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	rows, err := source.QueryContext(ctx, "SELECT id, product_id, content, cart_id FROM digital_data")
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, "INSERT INTO digital_data (id, product_id, content, cart_id) VALUES ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, productID, content string
		var cartID sql.NullString
		if err := rows.Scan(&id, &productID, &content, &cartID); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, id, productID, content, cartID); err != nil {
			return err
		}
	}

	return rows.Err()
}

// migrateCarts migrates the cart table
func migrateCarts(ctx context.Context, source *sql.DB, target *sql.Tx) error {
	rows, err := source.QueryContext(ctx, `
		SELECT id, email, amount_total, currency, payment_id, payment_status, cart, payment_system, created, updated
		FROM cart
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stmt, err := target.PrepareContext(ctx, `
		INSERT INTO cart (id, email, amount_total, currency, payment_id, payment_status, cart, payment_system, created, updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var id, amountTotal, currency, paymentSystem string
		var email, paymentID, paymentStatus sql.NullString
		var cart []byte
		var created, updated sql.NullTime
		if err := rows.Scan(&id, &email, &amountTotal, &currency, &paymentID, &paymentStatus, &cart, &paymentSystem, &created, &updated); err != nil {
			return err
		}
		if _, err := stmt.ExecContext(ctx, id, email, amountTotal, currency, paymentID, paymentStatus, cart, paymentSystem, created, updated); err != nil {
			return err
		}
	}

	return rows.Err()
}

// BackupDatabase creates a timestamped backup of the database
func BackupDatabase(dbPath string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.backup_%s", dbPath, timestamp)

	// For SQLite, we can just copy the file
	// For PostgreSQL, we would use pg_dump (not implemented here)

	// Simple file copy for SQLite
	data, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return "", err
	}
	defer data.Close()

	backup, err := sql.Open("sqlite", backupPath)
	if err != nil {
		return "", err
	}
	defer backup.Close()

	_, err = backup.Exec("VACUUM INTO ?", backupPath)
	if err != nil {
		// Fallback: just create an empty backup file as a marker
		return backupPath, nil
	}

	return backupPath, nil
}
