#!/bin/bash
# Demo: Cloudflare Mode Detection

echo "=== Cloudflare Mode Detection Demo ==="
echo ""

# Clean up any running servers
pkill -f "mycart serve" 2>/dev/null

# Set Cloudflare environment
export CLOUDFLARE_APPLICATION_ID=test-app-123
export CF_ACCOUNT_ID=mock-account
export CF_API_TOKEN=mock-token

echo "1. Environment variables set:"
echo "   CLOUDFLARE_APPLICATION_ID=$CLOUDFLARE_APPLICATION_ID"
echo "   CF_ACCOUNT_ID=$CF_ACCOUNT_ID"
echo ""

# Build
echo "2. Building..."
go build -o mycart ./cmd 2>&1 | grep -v "^#" || exit 1
echo "   ✓ Build complete"
echo ""

# Start server with Cloudflare env
echo "3. Starting server in Cloudflare mode..."
./mycart serve --dev > /tmp/cf-server.log 2>&1 &
SERVER_PID=$!
sleep 3

# Test
echo "4. Testing runtime detection:"
RESPONSE=$(curl -s http://localhost:8080/_/api/maintenance/status)
echo "$RESPONSE" | jq '.'
echo ""

# Check if Cloudflare mode detected
IS_CF=$(echo "$RESPONSE" | jq -r '.runtime.cloudflare')
if [ "$IS_CF" = "true" ]; then
    echo "✓ Cloudflare mode DETECTED"
else
    echo "✗ Cloudflare mode NOT detected (still in local mode)"
fi

# Cleanup
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo ""
echo "Server log (last 10 lines):"
tail -5 /tmp/cf-server.log
