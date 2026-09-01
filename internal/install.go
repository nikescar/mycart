package app

import (
	"context"
	"fmt"

	"github.com/shurco/mycart/internal/config"
	"github.com/shurco/mycart/internal/models"
	"github.com/shurco/mycart/internal/queries"
	"github.com/shurco/mycart/migrations"
	"github.com/shurco/mycart/pkg/database"
)

// InstallAdmin performs first-time setup with the given admin credentials.
// It initializes storage directories, runs migrations when needed, and
// creates the admin account. Safe to run from a one-shot container job.
func InstallAdmin(ctx context.Context, install *models.Install) error {
	if err := install.Validate(); err != nil {
		return fmt.Errorf("validate install: %w", err)
	}

	if err := Init(); err != nil {
		return fmt.Errorf("init: %w", err)
	}

	dbConfig := config.LoadDatabaseConfig()
	db, err := database.New(dbConfig)
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	if err := queries.New(db, migrations.Embed()); err != nil {
		return fmt.Errorf("connect database: %w", err)
	}

	if err := queries.DB().Install(ctx, install); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	return nil
}
