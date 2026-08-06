package database

import (
	"os"
	"testing"
)

func TestDetectDatabases_NewSQLite(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DB_HOST")

	databases := DetectDatabases()

	if len(databases) == 0 {
		t.Fatal("Expected at least one database to be detected")
	}

	// Should always have new SQLite option
	foundNewSQLite := false
	for _, db := range databases {
		if db.Type == "sqlite" && db.Source == "new installation" {
			foundNewSQLite = true
			if !db.Available {
				t.Error("New SQLite should always be available")
			}
			if db.Config == nil {
				t.Error("New SQLite should have a config")
			}
		}
	}

	if !foundNewSQLite {
		t.Error("Expected to find new SQLite option")
	}
}

func TestDetectDatabases_WithDatabaseURL(t *testing.T) {
	// Set DATABASE_URL
	testURL := "postgresql://user:pass@localhost:5432/testdb"
	os.Setenv("DATABASE_URL", testURL)
	defer os.Unsetenv("DATABASE_URL")

	databases := DetectDatabases()

	foundPostgreSQL := false
	for _, db := range databases {
		if db.Type == "postgresql" && db.Source == "DATABASE_URL" {
			foundPostgreSQL = true
			if db.Config == nil {
				t.Error("PostgreSQL should have a config")
			}
			if db.Config.PostgreSQL.Host != "localhost" {
				t.Errorf("Expected host 'localhost', got '%s'", db.Config.PostgreSQL.Host)
			}
			// Don't check Available since we may not have a real database
		}
	}

	if !foundPostgreSQL {
		t.Error("Expected to find PostgreSQL from DATABASE_URL")
	}
}

func TestDetectDatabases_ExistingSQLite(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Save current directory and change to temp
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	// Create lc_base directory and database file
	os.MkdirAll("lc_base", 0755)
	file, err := os.Create("lc_base/data.db")
	if err != nil {
		t.Fatalf("Failed to create test database file: %v", err)
	}
	file.Close()

	// Clear environment
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DB_HOST")

	databases := DetectDatabases()

	foundExistingSQLite := false
	for _, db := range databases {
		if db.Type == "sqlite" && db.Source == "existing file" {
			foundExistingSQLite = true
			if db.Config == nil {
				t.Error("Existing SQLite should have a config")
			}
		}
	}

	if !foundExistingSQLite {
		t.Error("Expected to find existing SQLite database")
	}
}

func TestDetectPreferredDatabase_NewSQLite(t *testing.T) {
	// Clear environment
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DB_HOST")

	preferred := DetectPreferredDatabase()

	if preferred == nil {
		t.Fatal("Expected to find a preferred database")
	}

	if preferred.Type != "sqlite" {
		t.Errorf("Expected preferred database to be sqlite, got %s", preferred.Type)
	}
}

func TestDetectPreferredDatabase_ExistingSQLite(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)

	os.Chdir(tmpDir)

	// Create existing SQLite database
	os.MkdirAll("lc_base", 0755)
	file, err := os.Create("lc_base/data.db")
	if err != nil {
		t.Fatalf("Failed to create test database file: %v", err)
	}
	file.Close()

	// Clear environment
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("DB_HOST")

	preferred := DetectPreferredDatabase()

	if preferred == nil {
		t.Fatal("Expected to find a preferred database")
	}

	if preferred.Source != "existing file" {
		t.Errorf("Expected preferred database to be existing file, got %s", preferred.Source)
	}
}

func TestDetectedDatabase_Fields(t *testing.T) {
	db := DetectedDatabase{
		Type:        "sqlite",
		Source:      "test",
		Description: "Test database",
		Available:   true,
		Config: &Config{
			Type: "sqlite",
		},
	}

	if db.Type != "sqlite" {
		t.Errorf("Expected Type 'sqlite', got '%s'", db.Type)
	}
	if db.Source != "test" {
		t.Errorf("Expected Source 'test', got '%s'", db.Source)
	}
	if !db.Available {
		t.Error("Expected Available to be true")
	}
	if db.Config == nil {
		t.Error("Expected Config to be set")
	}
}
