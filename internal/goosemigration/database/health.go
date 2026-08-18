package database

import (
	"context"
	"sync"
	"time"
)

// HealthChecker monitors database connection health in the background.
type HealthChecker struct {
	db       Database
	healthy  bool
	mu       sync.RWMutex
	stopChan chan struct{}
	interval time.Duration
}

// NewHealthChecker creates a new health checker that monitors the database connection.
func NewHealthChecker(db Database) *HealthChecker {
	return &HealthChecker{
		db:       db,
		healthy:  true,
		stopChan: make(chan struct{}),
		interval: 30 * time.Second,
	}
}

// NewHealthCheckerWithInterval creates a health checker with a custom check interval.
func NewHealthCheckerWithInterval(db Database, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		db:       db,
		healthy:  true,
		stopChan: make(chan struct{}),
		interval: interval,
	}
}

// Start begins background health monitoring.
func (h *HealthChecker) Start() {
	go h.monitor()
}

// Stop stops the background health monitoring.
func (h *HealthChecker) Stop() {
	close(h.stopChan)
}

// IsHealthy returns true if the database connection is healthy.
func (h *HealthChecker) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.healthy
}

// monitor runs the health check loop in the background.
func (h *HealthChecker) monitor() {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	// Initial health check
	h.checkHealth()

	for {
		select {
		case <-ticker.C:
			h.checkHealth()
		case <-h.stopChan:
			return
		}
	}
}

// checkHealth performs a single health check.
func (h *HealthChecker) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.db.Ping(ctx)

	h.mu.Lock()
	wasHealthy := h.healthy
	h.healthy = (err == nil)
	h.mu.Unlock()

	// Log state changes (in a real implementation, use a proper logger)
	if wasHealthy && !h.healthy {
		// Connection lost
		_ = err // In production, log this error
	} else if !wasHealthy && h.healthy {
		// Connection restored
	}
}

// CheckNow performs an immediate health check and returns the result.
func (h *HealthChecker) CheckNow() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.db.Ping(ctx)

	h.mu.Lock()
	h.healthy = (err == nil)
	h.mu.Unlock()

	return err
}
