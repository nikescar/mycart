# Cloudflare Container Workers Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Cloudflare Container Workers deployment support to mycart while maintaining full CLI compatibility through unified database and storage abstractions.

**Architecture:** Single Go binary with abstraction interfaces that auto-detect environment (local CLI vs Cloudflare) and initialize appropriate backends (SQLite/filesystem vs D1/R2). Cloudflare deployment uses Container Workers with TypeScript proxy routing API requests to container and static assets to Workers Assets.

**Tech Stack:** Go 1.26, Fiber v3, SQLite (modernc.org/sqlite), Cloudflare D1, Cloudflare R2, AWS S3 SDK (for R2), Docker (distroless), Wrangler CLI, TypeScript

**Spec:** `docs/superpowers/specs/2026-09-01-cloudflare-container-workers-design.md`

## Global Constraints

- Go 1.26+ required
- Test coverage >= 80% for all new code
- Use `t.Parallel()` for all tests
- Use `t.TempDir()` and `t.Setenv()` for test isolation
- Table-driven tests for multiple scenarios
- No breaking changes to existing CLI functionality
- Follow existing mycart patterns (dependency injection, error wrapping with context)
- Use `pkg/httpclient.New()` for HTTP clients (never `http.DefaultClient`)
- All new migrations must use `CURRENT_TIMESTAMP` not `datetime('now')`
- Commit after each task completion

---

## Phase 1: Foundation (Tasks 1-7)

Implement core abstractions with TDD approach.

---

### Task 1: Database Interface & Tests

Create the database abstraction interface and comprehensive SQLite tests following TDD.

**Implementation:** Write failing tests first (RED), then minimal implementation (GREEN).

**Validation:** `cd pkg/database && go test -v -race -cover`

**Expected coverage:** >= 80%

---

### Task 2: SQLite Adapter

Implement SQLite database adapter to make tests pass.

**Implementation:** Wrap `modernc.org/sqlite` with Database interface.

**Validation:** `cd pkg/database && go test -v -race`

**Expected:** All tests PASS

---

### Task 3: D1 Adapter Stub

Create D1 adapter structure (HTTP API implementation deferred).

**Implementation:** Constructor validation and no-op transaction support.

**Validation:** `cd pkg/database && go test -v -run TestD1`

**Expected:** Constructor tests PASS

---

### Task 4: Database Factory

Implement factory pattern for database creation.

**Implementation:** Factory function with table-driven tests.

**Validation:** `cd pkg/database && go test -v`

**Expected:** All database tests PASS, coverage >= 80%

---

### Task 5: Storage Interface & Filesystem

Create storage abstraction and implement filesystem storage.

**Implementation:** Interface definition + filesystem implementation with tests.

**Validation:** `cd pkg/storage && go test -v -race`

**Expected:** All tests PASS, coverage >= 80%

---

### Task 6: R2 Adapter & Factory

Implement R2 storage using AWS S3 SDK and create factory.

**Implementation:** Add AWS SDK dependency, implement R2, create factory.

**Validation:** `cd pkg/storage && go test -v -race`

**Expected:** All tests PASS, coverage >= 80%

---

### Task 7: Environment Detection & Config

Implement environment detection and configuration loading with auto-detection.

**Implementation:** Detect Cloudflare vs local, load appropriate configs.

**Validation:** `cd internal/config && go test -v -race`

**Expected:** All tests PASS, coverage >= 80%

---

## Next Steps

After completing Phase 1 (Tasks 1-7), additional phases needed:

**Phase 2: Application Integration**
- Update `internal/app.go` to use abstractions
- Update `internal/queries/*.go` to use `database.Database`
- Update handlers to use `storage.Storage`

**Phase 3: Cloudflare Infrastructure**
- Create Dockerfile (distroless)
- Create wrangler configuration
- Create TypeScript worker
- Create build/deploy scripts

**Phase 4: Testing & Validation**
- Integration tests
- End-to-end testing
- Performance testing

**Phase 5: Documentation**
- README updates
- Environment variable documentation
- Troubleshooting guide

---

**Note:** This plan contains the foundational tasks (1-7). Execute these first, then we'll create detailed plans for the remaining phases.
