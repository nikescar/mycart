package app

import (
	"github.com/shurco/mycart/internal/base"
	"github.com/shurco/mycart/internal/config"
	"github.com/shurco/mycart/migrations"
	"github.com/shurco/mycart/pkg/fsutil"
)

const (
	dbPath = "./lc_base/data.db"
)

var (
	requiredDirs = []string{"./lc_uploads", "./lc_digitals"}
)

// Init initializes the directory structure and database
// Only creates local directories and SQLite DB when not using D1/R2
func Init() error {
	// Load config to check if we're using cloud storage
	dbConfig := config.LoadDatabaseConfig()
	storageConfig := config.LoadStorageConfig()

	// Only create storage directories when using local filesystem (not R2)
	if storageConfig.Type != "r2" {
		for _, dir := range requiredDirs {
			if err := fsutil.MkDirs(0o775, dir); err != nil {
				if log != nil {
					log.Err(err).Send()
				}
				return err
			}
		}
	}

	// Only create SQLite database when using SQLite (not D1)
	if dbConfig.Type != "d1" {
		if _, err := base.New(dbPath, migrations.Embed()); err != nil {
			if log != nil {
				log.Err(err).Send()
			}
			return err
		}
	}

	return nil
}

// Migrate performs database migrations
func Migrate() error {
	return base.Migrate(dbPath, migrations.Embed())
}
