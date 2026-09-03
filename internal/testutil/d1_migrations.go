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

// isOnlyComments checks if a statement contains only SQL comments
func isOnlyComments(stmt string) bool {
	lines := strings.Split(stmt, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			return false
		}
	}
	return true
}

// parseGooseSQL strips goose directives and splits SQL into individual statements
// Only processes the "Up" section, stops at "-- +goose Down"
func parseGooseSQL(sql string) []string {
	var statements []string
	lines := strings.Split(sql, "\n")
	var currentStmt strings.Builder
	inUpSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start processing when we see "-- +goose Up"
		if strings.HasPrefix(trimmed, "-- +goose Up") {
			inUpSection = true
			continue
		}

		// Stop processing when we see "-- +goose Down"
		if strings.HasPrefix(trimmed, "-- +goose Down") {
			break
		}

		// Skip other goose directives
		if strings.HasPrefix(trimmed, "-- +goose") {
			continue
		}

		// Only process lines in Up section
		if !inUpSection {
			continue
		}

		// Add line to current statement
		currentStmt.WriteString(line)
		currentStmt.WriteString("\n")

		// If line ends with semicolon, it's a complete statement
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(currentStmt.String())
			if stmt != "" && stmt != ";" && !isOnlyComments(stmt) {
				statements = append(statements, stmt)
			}
			currentStmt.Reset()
		}
	}

	// Add any remaining statement
	if currentStmt.Len() > 0 {
		stmt := strings.TrimSpace(currentStmt.String())
		if stmt != "" && stmt != ";" && !isOnlyComments(stmt) {
			statements = append(statements, stmt)
		}
	}

	return statements
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
		// Parse goose migration into individual statements
		statements := parseGooseSQL(mig.SQL)

		// Execute each statement separately (D1 doesn't support batch execution)
		for i, stmt := range statements {
			if err := client.ExecuteD1SQL(databaseID, stmt); err != nil {
				return fmt.Errorf("migration %s (statement %d) failed: %w", mig.Name, i+1, err)
			}
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
		// Parse fixture SQL into individual statements
		statements := parseGooseSQL(fixture.SQL)

		// Execute each statement separately (D1 doesn't support batch execution)
		for i, stmt := range statements {
			if err := client.ExecuteD1SQL(databaseID, stmt); err != nil {
				return fmt.Errorf("fixture %s (statement %d) failed: %w", fixture.Name, i+1, err)
			}
		}
	}

	return nil
}
