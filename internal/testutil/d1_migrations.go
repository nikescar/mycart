package testutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/shurco/mycart/migrations"
)

type migrationFile struct {
	Name string
	SQL  string
}

// readMigrationFiles reads migration SQL from embedded filesystem
func readMigrationFiles() ([]migrationFile, error) {
	var migs []migrationFile

	embedFS := migrations.Embed()

	err := fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := fs.ReadFile(embedFS, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		migs = append(migs, migrationFile{
			Name: filepath.Base(path),
			SQL:  string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by name for consistent ordering
	sort.Slice(migs, func(i, j int) bool {
		return migs[i].Name < migs[j].Name
	})

	return migs, nil
}

// readFixtureFiles reads fixture SQL from filesystem
func readFixtureFiles() ([]migrationFile, error) {
	var fixtures []migrationFile

	// Get project root
	_, src, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(src), "..", "..")
	fixturesDir := filepath.Join(projectRoot, "fixtures", "migration")

	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		return nil, fmt.Errorf("read fixtures dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(fixturesDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		fixtures = append(fixtures, migrationFile{
			Name: entry.Name(),
			SQL:  string(content),
		})
	}

	// Sort by name
	sort.Slice(fixtures, func(i, j int) bool {
		return fixtures[i].Name < fixtures[j].Name
	})

	return fixtures, nil
}

// runMigrationsOnD1 applies all schema migrations to a D1 database
func runMigrationsOnD1(client interface {
	ExecuteD1SQL(databaseID, sql string) error
}, databaseID string) error {
	migs, err := readMigrationFiles()
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	for _, mig := range migs {
		if err := client.ExecuteD1SQL(databaseID, mig.SQL); err != nil {
			return fmt.Errorf("migration %s failed: %w", mig.Name, err)
		}
	}

	return nil
}

// runFixturesOnD1 applies fixture data to a D1 database
func runFixturesOnD1(client interface {
	ExecuteD1SQL(databaseID, sql string) error
}, databaseID string) error {
	fixtures, err := readFixtureFiles()
	if err != nil {
		return fmt.Errorf("read fixtures: %w", err)
	}

	for _, fixture := range fixtures {
		if err := client.ExecuteD1SQL(databaseID, fixture.SQL); err != nil {
			return fmt.Errorf("fixture %s failed: %w", fixture.Name, err)
		}
	}

	return nil
}
