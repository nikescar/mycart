# Building myCart

This guide explains how to build the myCart application from source.

## Overview

myCart consists of two main components:
1. **Frontend**: Two Vite-based web applications (admin panel + site)
2. **Backend**: Go application that embeds the built frontend assets

## Prerequisites

### Required
- **Go** 1.21 or higher
- **Node.js** 18 or higher
- **npm** or **bun** (bun is faster but npm works fine)

### Optional
- **Docker** (for containerized builds)
- **goreleaser** (for release builds)

## Quick Build (Development)

### 1. Build Frontend Assets

```bash
# Build admin panel
cd web/admin
npm ci
npm run build  # or: bun install && bun run build

# Build site
cd ../site
npm ci
npm run build  # or: bun install && bun run build

# Return to project root
cd ../..
```

This creates:
- `web/admin/build/` - Built admin panel
- `web/site/build/` - Built site

### 2. Build Go Binary

```bash
# Build for current platform
go build -o mycart ./cmd/main.go

# Or with version information
go build -ldflags="-X main.version=dev -X main.gitCommit=$(git rev-parse --short HEAD) -X main.buildDate=$(date +%Y-%m-%d)" -o mycart ./cmd/main.go
```

This creates the `mycart` executable in the current directory.

### 3. Run

```bash
./mycart serve
```

## Production Build

### Build Script

Create a build script for convenience:

```bash
#!/bin/bash
# build.sh

set -e

echo "Building myCart..."

# Clean previous builds
echo "Cleaning previous builds..."
rm -rf web/admin/build web/site/build mycart

# Build frontend
echo "Building admin panel..."
cd web/admin
npm ci
npm run build
cd ../..

echo "Building site..."
cd web/site
npm ci
npm run build
cd ../..

# Build backend with version info
VERSION=${VERSION:-dev}
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date +%Y-%m-%d)

echo "Building Go binary..."
CGO_ENABLED=0 go build \
  -ldflags="-w -s -X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}" \
  -o mycart \
  ./cmd/main.go

echo "Build complete! Binary: ./mycart"
echo "Version: ${VERSION}"
echo "Commit: ${GIT_COMMIT}"
echo "Date: ${BUILD_DATE}"
```

Make it executable and run:

```bash
chmod +x build.sh
./build.sh
```

## Cross-Platform Builds

### Build for Different Platforms

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o mycart-linux-amd64 ./cmd/main.go

# Linux (arm64)
GOOS=linux GOARCH=arm64 go build -o mycart-linux-arm64 ./cmd/main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o mycart-darwin-amd64 ./cmd/main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o mycart-darwin-arm64 ./cmd/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o mycart-windows-amd64.exe ./cmd/main.go
```

### Multi-Platform Build Script

```bash
#!/bin/bash
# build-all.sh

set -e

# Build frontend first (once)
cd web/admin && npm ci && npm run build && cd ../..
cd web/site && npm ci && npm run build && cd ../..

VERSION=${VERSION:-dev}
GIT_COMMIT=$(git rev-parse --short HEAD)
BUILD_DATE=$(date +%Y-%m-%d)
LDFLAGS="-w -s -X main.version=${VERSION} -X main.gitCommit=${GIT_COMMIT} -X main.buildDate=${BUILD_DATE}"

mkdir -p dist

# Build for each platform
for GOOS in linux darwin windows; do
  for GOARCH in amd64 arm64; do
    OUTPUT="dist/mycart-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
      OUTPUT="${OUTPUT}.exe"
    fi
    
    echo "Building for ${GOOS}/${GOARCH}..."
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
      -ldflags="$LDFLAGS" \
      -o "$OUTPUT" \
      ./cmd/main.go
  done
done

echo "All builds complete! Check ./dist/"
ls -lh dist/
```

## Docker Build

### Using Dockerfile.build (Recommended)

```bash
# Build Docker image
docker build -f Dockerfile.build -t mycart:latest .

# Run container
docker run -p 8080:8080 mycart:latest

# Or with environment variables
docker run -p 8080:8080 \
  -e DB_TYPE=postgresql \
  -e DATABASE_URL=postgresql://... \
  mycart:latest
```

### Using Docker Compose

```bash
# SQLite (default)
docker-compose build
docker-compose up

# PostgreSQL
docker-compose -f docker-compose.postgres.yml build
docker-compose -f docker-compose.postgres.yml up
```

## Release Build (goreleaser)

### Prerequisites

Install goreleaser:

```bash
# macOS
brew install goreleaser

# Linux
go install github.com/goreleaser/goreleaser@latest

# Or download from https://github.com/goreleaser/goreleaser/releases
```

### Local Release Build

```bash
# Test release build (doesn't publish)
goreleaser release --snapshot --clean

# This creates:
# - dist/mycart_*_linux-amd64.tar.gz
# - dist/mycart_*_darwin-amd64.tar.gz
# - dist/mycart_*_windows-amd64.zip
# - And more...
```

### Official Release (Maintainers Only)

```bash
# Tag a new version
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# Run goreleaser
GITHUB_TOKEN=xxx goreleaser release --clean
```

## Build Optimization

### Reducing Binary Size

```bash
# Strip symbols and debug info
go build -ldflags="-w -s" -o mycart ./cmd/main.go

# With UPX compression (optional)
upx --best --lzma mycart
```

### Build Cache

```bash
# Clear Go build cache
go clean -cache

# Clear module cache
go clean -modcache

# Rebuild everything
go build -a -o mycart ./cmd/main.go
```

## Troubleshooting

### Frontend Build Fails

```bash
# Clear node_modules and reinstall
rm -rf web/admin/node_modules web/site/node_modules
cd web/admin && npm ci && npm run build
cd ../site && npm ci && npm run build
```

### "Embeds" Error

```bash
# Error: pattern admin/build: cannot embed directory admin/build: contains no embeddable files

# Solution: Ensure frontend is built before Go build
cd web/admin && npm run build
cd ../site && npm run build
cd ../.. && go build ./cmd/main.go
```

### Wrong Node Version

```bash
# Use nvm to switch versions
nvm install 18
nvm use 18

# Or use specific version
docker run -v $(pwd):/app node:18-alpine sh -c "cd /app/web/admin && npm ci && npm run build"
```

### Permission Denied

```bash
# Make build script executable
chmod +x build.sh

# Or run with bash
bash build.sh
```

## Build Verification

### Check Binary Info

```bash
# Show version
./mycart --version

# Check file size
ls -lh mycart

# Show build details
file mycart

# Check dependencies (Linux)
ldd mycart

# Check symbols
nm mycart | grep main.version
```

### Test Built Binary

```bash
# Initialize
./mycart init

# Install (creates admin account)
./mycart install --email=admin@example.com --password=secure123 --domain=localhost

# Serve
./mycart serve

# Test health endpoint
curl http://localhost:8080/api/health
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      
      - name: Build Frontend
        run: |
          cd web/admin && npm ci && npm run build
          cd ../site && npm ci && npm run build
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: Build Binary
        run: |
          go build -v -o mycart ./cmd/main.go
      
      - name: Test Binary
        run: |
          ./mycart --version
```

## Build Configurations

### Development Build

Fast build for development:

```bash
# No optimization, includes debug symbols
go build -o mycart ./cmd/main.go
```

### Production Build

Optimized for size and performance:

```bash
go build -ldflags="-w -s" -trimpath -o mycart ./cmd/main.go
```

### Debug Build

With race detector and debug symbols:

```bash
go build -race -gcflags="all=-N -l" -o mycart-debug ./cmd/main.go
```

## Environment Variables During Build

### Frontend Build Variables

```bash
# Set in web/admin/.env or web/site/.env
VITE_API_URL=https://api.example.com
VITE_PUBLIC_PATH=/admin/

# Then build
cd web/admin && npm run build
```

### Backend Build Variables

```bash
# Set version info
VERSION=1.2.3 \
GIT_COMMIT=$(git rev-parse --short HEAD) \
BUILD_DATE=$(date +%Y-%m-%d) \
./build.sh
```

## Static Assets

The Go binary embeds these directories:

- `web/admin/build/` → Embedded in binary
- `web/site/build/` → Embedded in binary
- `migrations/` → Embedded for database migrations

Ensure all are present before building:

```bash
# Check what will be embedded
ls -la web/admin/build web/site/build migrations/
```

## Build Output

After a successful build, you'll have:

```
mycart                    # Single executable
├── web/admin/build/      # (embedded)
├── web/site/build/       # (embedded)
└── migrations/           # (embedded)
```

The executable is **completely standalone** and requires no external files!

## Platform-Specific Notes

### macOS

```bash
# Notarization for distribution (maintainers only)
codesign -s "Developer ID" mycart
xcrun notarytool submit mycart.zip --wait

# Remove quarantine attribute
xattr -d com.apple.quarantine mycart
```

### Windows

```bash
# Build Windows executable on Linux/Mac
GOOS=windows GOARCH=amd64 go build -o mycart.exe ./cmd/main.go

# With Windows icon (requires rsrc)
go install github.com/akavel/rsrc@latest
rsrc -ico icon.ico -o rsrc.syso
go build -o mycart.exe ./cmd/main.go
```

### Linux

```bash
# Static binary (no glibc dependency)
CGO_ENABLED=0 go build -o mycart ./cmd/main.go

# Verify static linking
ldd mycart  # Should say "not a dynamic executable"
```

## Summary

**Simple build:**
```bash
cd web/admin && npm ci && npm run build && cd ../..
cd web/site && npm ci && npm run build && cd ../..
go build -o mycart ./cmd/main.go
```

**Production build:**
```bash
./build.sh  # Use the build script above
```

**Docker build:**
```bash
docker build -f Dockerfile.build -t mycart:latest .
```

**Release build:**
```bash
goreleaser release --snapshot --clean
```

That's it! You now have a fully built myCart executable. 🚀

## Further Reading

- [Go Build Documentation](https://pkg.go.dev/cmd/go#hdr-Compile_packages_and_dependencies)
- [Vite Build Guide](https://vitejs.dev/guide/build.html)
- [GoReleaser Documentation](https://goreleaser.com/)
- [Docker Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
