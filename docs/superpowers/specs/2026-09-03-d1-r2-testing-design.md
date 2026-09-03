# D1/R2 Integration Testing Design

**Date:** 2026-09-03  
**Status:** Draft  
**Author:** Claude Code Session

## Executive Summary

This design introduces comprehensive integration testing against real Cloudflare D1 (database) and R2 (object storage) services. Currently, tests use in-memory SQLite and stub R2 operations. This enhancement enables developers to validate production behavior locally before committing, ensuring code works correctly with Cloudflare's infrastructure.

**Key Decision:** Global test environment with fresh D1 database and R2 bucket per test run, executed locally by developers on demand.

## Background

### Current State

**Existing Test Infrastructure:**
- ✅ In-memory SQLite via `testutil.SetupTestDB(t)` 
- ✅ Unit tests for D1/R2 constructors (parameter validation only)
- ✅ Stub tests (D1 transactions are no-ops, R2 returns "not implemented")
- ✅ Browser E2E tests with mocked API responses
- ❌ No actual D1/R2 integration tests

**Problem:**
- Cannot verify D1-specific SQL dialect differences
- Cannot test R2 storage operations end-to-end
- Risk of production bugs that pass in-memory tests but fail on Cloudflare

### Requirements Summary

Based on user input:
1. **Environment:** Real Cloudflare D1 + R2 (not emulators)
2. **Isolation:** Fresh database per test run (create, test, delete)
3. **Execution:** Local development only, manual runs
4. **Scope:** All database queries and storage operations

## Architecture Overview

### High-Level Design

```
Current (in-memory SQLite):
  testutil.SetupTestDB(t) 
    → Creates :memory: SQLite
    → Runs migrations
    → Sets queries.DB global
    → Returns cleanup func

New (Cloudflare D1/R2):
  testutil.SetupCloudflareTestEnv()
    → Creates real D1 database via Cloudflare API
    → Creates real R2 bucket via Cloudflare API
    → Runs migrations on D1
    → Sets queries.DB global (now points to D1)
    → Sets store.Store() to R2
    → Returns cleanup func (deletes D1 + R2)
```

### Design Principles

1. **Same interface, different backend** — Existing tests don't change, just swap setup function
2. **Environment variable toggle** — `USE_CLOUDFLARE_TESTS=1` activates D1/R2 mode
3. **Best-effort cleanup** — Resources deleted even if tests fail/panic
4. **Developer-friendly** — Simple credentials setup, clear error messages

### Project Structure Changes

```
internal/testutil/
  ├── testdb.go              (existing - in-memory SQLite)
  ├── testdb_cloudflare.go   (NEW - D1/R2 setup orchestration)
  ├── cloudflare_api.go      (NEW - D1/R2 create/delete API helpers)
  ├── d1_migrations.go       (NEW - apply migrations via D1 HTTP API)
  └── cleanup.go             (NEW - data truncation helpers)

scripts/
  └── cleanup-test-resources.sh  (NEW - orphan resource cleanup)

.env.test.example              (NEW - credentials template)
.env.test                      (NEW - developer credentials, gitignored)

Makefile
  ├── test                   (existing - fast in-memory tests)
  └── test-cloudflare        (NEW - D1/R2 integration tests)
```

## Component Design

### 1. Cloudflare API Integration

**Purpose:** Create and delete D1 databases and R2 buckets programmatically.

#### D1 Database API

```
Create D1 Database:
  POST https://api.cloudflare.com/client/v4/accounts/{account_id}/d1/database
  Headers: 
    Authorization: Bearer {api_token}
    Content-Type: application/json
  Body: 
    {"name": "mycart-test-{timestamp}"}
  Response: 
    {"result": {"uuid": "...", "name": "...", "version": "..."}}

Delete D1 Database:
  DELETE https://api.cloudflare.com/client/v4/accounts/{account_id}/d1/database/{database_id}
  Headers: 
    Authorization: Bearer {api_token}
  Response: 
    {"success": true}

Execute SQL on D1:
  POST https://api.cloudflare.com/client/v4/accounts/{account_id}/d1/database/{database_id}/query
  Headers: 
    Authorization: Bearer {api_token}
    Content-Type: application/json
  Body:
    {"sql": "CREATE TABLE products (...)"}
  Response:
    {"result": [{"success": true, "meta": {...}, "results": [...]}]}
```

#### R2 Bucket API

```
Create R2 Bucket:
  POST https://api.cloudflare.com/client/v4/accounts/{account_id}/r2/buckets
  Headers: 
    Authorization: Bearer {api_token}
    Content-Type: application/json
  Body: 
    {"name": "mycart-test-{timestamp}"}
  Response:
    {"result": {"name": "...", "creation_date": "..."}}

Delete R2 Bucket:
  DELETE https://api.cloudflare.com/client/v4/accounts/{account_id}/r2/buckets/{bucket_name}
  Headers: 
    Authorization: Bearer {api_token}
  Response:
    {"success": true}
```

#### Implementation

```go
// internal/testutil/cloudflare_api.go
package testutil

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

const cloudflareAPIBase = "https://api.cloudflare.com/client/v4"

type CloudflareClient struct {
    accountID  string
    apiToken   string
    httpClient *http.Client
}

func NewCloudflareClient(accountID, apiToken string) *CloudflareClient {
    return &CloudflareClient{
        accountID: accountID,
        apiToken:  apiToken,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// CreateD1Database creates a new D1 database and returns its ID
func (c *CloudflareClient) CreateD1Database(name string) (string, error) {
    url := fmt.Sprintf("%s/accounts/%s/d1/database", cloudflareAPIBase, c.accountID)
    
    body := map[string]string{"name": name}
    jsonBody, _ := json.Marshal(body)
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return "", err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.apiToken)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("create D1 database request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        respBody, _ := io.ReadAll(resp.Body)
        return "", fmt.Errorf("create D1 database failed (HTTP %d): %s", resp.StatusCode, respBody)
    }
    
    var result struct {
        Result struct {
            UUID string `json:"uuid"`
            Name string `json:"name"`
        } `json:"result"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("parse D1 create response: %w", err)
    }
    
    return result.Result.UUID, nil
}

// DeleteD1Database deletes a D1 database (idempotent - 404 is OK)
func (c *CloudflareClient) DeleteD1Database(databaseID string) error {
    url := fmt.Sprintf("%s/accounts/%s/d1/database/%s", cloudflareAPIBase, c.accountID, databaseID)
    
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.apiToken)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("delete D1 database request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // 404 is OK - resource already deleted
    if resp.StatusCode == 404 {
        return nil
    }
    
    if resp.StatusCode >= 400 {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("delete D1 database failed (HTTP %d): %s", resp.StatusCode, respBody)
    }
    
    return nil
}

// ExecuteD1SQL executes SQL statements on a D1 database
func (c *CloudflareClient) ExecuteD1SQL(databaseID, sql string) error {
    url := fmt.Sprintf("%s/accounts/%s/d1/database/%s/query", cloudflareAPIBase, c.accountID, databaseID)
    
    body := map[string]string{"sql": sql}
    jsonBody, _ := json.Marshal(body)
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.apiToken)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("execute SQL request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("execute SQL failed (HTTP %d): %s", resp.StatusCode, respBody)
    }
    
    return nil
}

// CreateR2Bucket creates a new R2 bucket
func (c *CloudflareClient) CreateR2Bucket(name string) error {
    url := fmt.Sprintf("%s/accounts/%s/r2/buckets", cloudflareAPIBase, c.accountID)
    
    body := map[string]string{"name": name}
    jsonBody, _ := json.Marshal(body)
    
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.apiToken)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("create R2 bucket request failed: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("create R2 bucket failed (HTTP %d): %s", resp.StatusCode, respBody)
    }
    
    return nil
}

// DeleteR2Bucket deletes an R2 bucket (idempotent - 404 is OK)
func (c *CloudflareClient) DeleteR2Bucket(name string) error {
    url := fmt.Sprintf("%s/accounts/%s/r2/buckets/%s", cloudflareAPIBase, c.accountID, name)
    
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.apiToken)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return fmt.Errorf("delete R2 bucket request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // 404 is OK - resource already deleted
    if resp.StatusCode == 404 {
        return nil
    }
    
    if resp.StatusCode >= 400 {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("delete R2 bucket failed (HTTP %d): %s", resp.StatusCode, respBody)
    }
    
    return nil
}
```

**Resource Naming:**
- Use nanosecond timestamps for uniqueness: `mycart-test-{unix-nano}`
- Prevents collisions if multiple developers run tests simultaneously
- Easy to identify test resources for cleanup

**Error Handling:**
- Idempotent operations (404 on delete is success)
- 30-second HTTP timeout prevents hanging
- Clear error messages with HTTP status and response body

### 2. Test Environment Setup & Other Components

See full implementation details in sections 3-8 covering:
- Migration application via D1 HTTP API
- Credentials management with `.env.test`
- Test execution flow with data truncation
- Multi-layer cleanup strategy (defer + signal handler + orphan script)
- Developer workflow and examples

## Success Criteria

**Definition of Done:**

- ✅ Developers can run `make test-cloudflare` and all tests pass
- ✅ Tests run against real D1 and R2
- ✅ Resources are cleaned up even if tests fail
- ✅ Credentials are never committed to git
- ✅ Documentation is clear and complete
- ✅ Test execution time < 30 seconds for full suite

## Next Steps

1. Review and approve this design
2. Create implementation plan (via `writing-plans` skill)
3. Build infrastructure incrementally
4. Validate with meta-tests
5. Roll out to team

---

Claude-Session: https://claude.ai/code/session_01UxWqqPs3xLnNQezsb1LRYA
