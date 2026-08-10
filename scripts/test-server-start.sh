#!/bin/sh
# Clean test environment before starting server
# This ensures the server opens a fresh database

echo "🧹 Cleaning test environment before server start..."
rm -rf lc_base lc_digitals lc_uploads
echo "✓ Environment cleaned"

echo "📁 Creating required directories..."
mkdir -p lc_base lc_digitals lc_uploads
echo "✓ Directories created"

echo "🗄️ Initializing SQLite database..."
touch lc_base/data.db
echo "✓ Database file created"

echo "🚀 Starting server..."
exec go run ./cmd serve
