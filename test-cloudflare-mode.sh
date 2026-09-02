#!/bin/bash
# Test Cloudflare mode locally

echo "=== Testing Cloudflare Runtime Detection ==="

# Set Cloudflare environment
export CLOUDFLARE_APPLICATION_ID=test-app-123
export CF_ACCOUNT_ID=test-account
export CF_API_TOKEN=test-token
export CF_D1_DATABASE_ID=test-d1-db
export CF_R2_BUCKET_NAME=test-r2-bucket

echo "✓ Cloudflare environment variables set"
echo ""

# Build and run
echo "Building application..."
go build -o mycart ./cmd || exit 1

echo ""
echo "Testing runtime detection..."
./mycart serve --dev &
SERVER_PID=$!

sleep 3

# Test if server started
if curl -s http://localhost:8080/ping > /dev/null 2>&1; then
    echo "✓ Server started successfully"
    
    # Check maintenance status
    echo ""
    echo "Checking runtime detection via maintenance status..."
    curl -s http://localhost:8080/_/api/maintenance/status | jq '.'
    
    kill $SERVER_PID 2>/dev/null
    echo ""
    echo "✓ Test complete"
else
    echo "✗ Server failed to start"
    kill $SERVER_PID 2>/dev/null
    exit 1
fi
