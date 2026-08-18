package database

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestNewHealthChecker(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	checker := NewHealthChecker(adapter)

	if checker == nil {
		t.Fatal("NewHealthChecker returned nil")
	}

	if !checker.IsHealthy() {
		t.Error("Expected initial health to be true")
	}

	if checker.interval != 30*time.Second {
		t.Errorf("Expected interval to be 30s, got %v", checker.interval)
	}
}

func TestNewHealthCheckerWithInterval(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	interval := 5 * time.Second
	checker := NewHealthCheckerWithInterval(adapter, interval)

	if checker.interval != interval {
		t.Errorf("Expected interval to be %v, got %v", interval, checker.interval)
	}
}

func TestHealthChecker_IsHealthy(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	checker := NewHealthChecker(adapter)

	if !checker.IsHealthy() {
		t.Error("Expected database to be healthy")
	}
}

func TestHealthChecker_CheckNow(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	checker := NewHealthChecker(adapter)

	err = checker.CheckNow()
	if err != nil {
		t.Errorf("CheckNow failed: %v", err)
	}

	if !checker.IsHealthy() {
		t.Error("Expected database to be healthy after CheckNow")
	}
}

func TestHealthChecker_CheckNow_Unhealthy(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	adapter := NewSQLiteAdapter(db)
	checker := NewHealthChecker(adapter)

	// Close the database to make it unhealthy
	db.Close()

	err = checker.CheckNow()
	if err == nil {
		t.Error("Expected error when checking closed database")
	}

	if checker.IsHealthy() {
		t.Error("Expected database to be unhealthy after closing")
	}
}

func TestHealthChecker_StartStop(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	adapter := NewSQLiteAdapter(db)
	// Use short interval for testing
	checker := NewHealthCheckerWithInterval(adapter, 100*time.Millisecond)

	// Start monitoring
	checker.Start()

	// Wait for at least one check cycle
	time.Sleep(150 * time.Millisecond)

	if !checker.IsHealthy() {
		t.Error("Expected database to remain healthy during monitoring")
	}

	// Stop monitoring
	checker.Stop()

	// Wait a bit to ensure it stopped
	time.Sleep(50 * time.Millisecond)
}

func TestHealthChecker_BackgroundMonitoring(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	adapter := NewSQLiteAdapter(db)
	// Use very short interval for testing
	checker := NewHealthCheckerWithInterval(adapter, 50*time.Millisecond)

	checker.Start()
	defer checker.Stop()

	// Initially healthy
	if !checker.IsHealthy() {
		t.Error("Expected database to be initially healthy")
	}

	// Close database to make it unhealthy
	db.Close()

	// Wait for the health check to detect the issue
	time.Sleep(200 * time.Millisecond)

	if checker.IsHealthy() {
		t.Error("Expected health checker to detect closed database")
	}
}
