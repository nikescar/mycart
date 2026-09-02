#!/bin/bash

echo "Testing Cloudflare Detection:"
echo ""

# Test 1: Local mode (no env vars)
echo "Test 1: Local Mode"
unset CLOUDFLARE_APPLICATION_ID RUNTIME
DB_TYPE=$(go run ./cmd serve --help 2>&1 | head -1)
echo "  (would use SQLite + Filesystem)"
echo ""

# Test 2: Cloudflare mode (with env var)
echo "Test 2: Cloudflare Mode"  
export CLOUDFLARE_APPLICATION_ID=test-123
echo "  CLOUDFLARE_APPLICATION_ID=$CLOUDFLARE_APPLICATION_ID"
echo "  (would use D1 + R2)"
echo ""

# Test 3: Explicit override
echo "Test 3: Explicit Override"
export RUNTIME=cloudflare
echo "  RUNTIME=$RUNTIME"
echo "  (forces Cloudflare mode even without other vars)"

echo ""
echo "To actually test:"
echo "1. export CLOUDFLARE_APPLICATION_ID=test-123"
echo "2. go run ./cmd serve --dev"
echo "3. curl http://localhost:8080/_/api/maintenance/status | jq '.'"
