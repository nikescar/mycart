package app

import (
	"fmt"

	"github.com/shurco/mycart/db/migrations"
	"github.com/shurco/mycart/internal/goosemigration/database"
)

// MigrateToPostgres migrates data from SQLite to PostgreSQL
func MigrateToPostgres() error {
	log.Info().Msg("Starting migration from SQLite to PostgreSQL...")

	// Connect to source SQLite database
	sqliteCfg := &database.Config{
		Type: "sqlite",
		SQLite: database.SQLiteConfig{
			Path: "./lc_base/data.db",
		},
	}

	log.Info().Msg("Connecting to SQLite database...")
	sourceDB, err := database.ConnectWithRetry(sqliteCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}
	defer sourceDB.Close()

	// Load PostgreSQL configuration from environment
	log.Info().Msg("Loading PostgreSQL configuration...")
	targetCfg, err := database.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load PostgreSQL config: %w", err)
	}

	// Override to ensure we're connecting to PostgreSQL
	if targetCfg.Type != "postgresql" && targetCfg.Type != "postgres" {
		targetCfg.Type = "postgresql"
	}

	// Connect to target PostgreSQL database
	log.Info().Msgf("Connecting to PostgreSQL at %s:%d...", targetCfg.PostgreSQL.Host, targetCfg.PostgreSQL.Port)
	targetDB, err := database.ConnectWithRetry(targetCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer targetDB.Close()

	// Run migrations on target database
	log.Info().Msg("Running migrations on PostgreSQL...")
	if err := database.RunMigrations(targetDB.DB(), "postgresql", migrations.Embed()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create backup of SQLite database
	log.Info().Msg("Creating backup of SQLite database...")
	backupPath, err := database.BackupDatabase(sqliteCfg.SQLite.Path)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create backup, continuing anyway...")
	} else {
		log.Info().Msgf("Backup created at: %s", backupPath)
	}

	// Migrate data
	log.Info().Msg("Migrating data from SQLite to PostgreSQL...")
	if err := database.MigrateData(sourceDB, targetDB); err != nil {
		return fmt.Errorf("failed to migrate data: %w", err)
	}

	log.Info().Msg("✓ Migration completed successfully!")
	log.Info().Msg("Next steps:")
	log.Info().Msg("  1. Verify your data in PostgreSQL")
	log.Info().Msg("  2. Update your environment variables to use PostgreSQL")
	log.Info().Msgf("     Set: DB_TYPE=postgresql or DATABASE_URL=...")
	log.Info().Msgf("  3. SQLite backup available at: %s", backupPath)

	return nil
}

// MigrateToSQLite migrates data from PostgreSQL to SQLite
func MigrateToSQLite() error {
	log.Info().Msg("Starting migration from PostgreSQL to SQLite...")

	// Load PostgreSQL configuration from environment
	log.Info().Msg("Loading PostgreSQL configuration...")
	sourceCfg, err := database.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load PostgreSQL config: %w", err)
	}

	// Connect to source PostgreSQL database
	log.Info().Msgf("Connecting to PostgreSQL at %s:%d...", sourceCfg.PostgreSQL.Host, sourceCfg.PostgreSQL.Port)
	sourceDB, err := database.ConnectWithRetry(sourceCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer sourceDB.Close()

	// Set up target SQLite configuration
	sqlitePath := "./lc_base/data_new.db"
	if path := sourceCfg.SQLite.Path; path != "" {
		sqlitePath = path
	}

	targetCfg := &database.Config{
		Type: "sqlite",
		SQLite: database.SQLiteConfig{
			Path: sqlitePath,
		},
	}

	// Connect to target SQLite database
	log.Info().Msgf("Creating SQLite database at %s...", sqlitePath)
	targetDB, err := database.ConnectWithRetry(targetCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}
	defer targetDB.Close()

	// Run migrations on target database
	log.Info().Msg("Running migrations on SQLite...")
	if err := database.RunMigrations(targetDB.DB(), "sqlite", migrations.Embed()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Migrate data
	log.Info().Msg("Migrating data from PostgreSQL to SQLite...")
	if err := database.MigrateData(sourceDB, targetDB); err != nil {
		return fmt.Errorf("failed to migrate data: %w", err)
	}

	log.Info().Msg("✓ Migration completed successfully!")
	log.Info().Msg("Next steps:")
	log.Info().Msg("  1. Verify your data in SQLite")
	log.Info().Msg("  2. Update your environment variables to use SQLite")
	log.Info().Msg("     Set: DB_TYPE=sqlite")
	log.Info().Msgf("  3. SQLite database created at: %s", sqlitePath)

	return nil
}
