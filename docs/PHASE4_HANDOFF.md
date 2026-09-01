# Phase 4 Handoff: Cloudflare Container Workers Setup

## Completed Work (Phase 1-3)

### Phase 1: Infrastructure ✅
- Database abstraction layer (`pkg/database`)
- Storage abstraction layer (`pkg/storage`)
- Environment detection (`internal/config/detect.go`)
- Factory patterns for auto-routing

**Commits:**
- `b31ceea` - Storage interface and filesystem adapter
- `26967dc` - R2 adapter and storage factory
- `50bc343` - Environment detection and configuration

### Phase 2: Database Integration ✅
- All 7 query modules use `database.Database` interface
- Updated 200+ method calls
- Full compilation success

**Commits:**
- `172e2d4` - Integrate database abstraction layer into query modules

### Phase 3: Storage Integration ✅
- Created `internal/store` package with singleton accessor
- Initialized storage in `app.go`
- Storage factory integrated

**Commits:**
- `8dc574b` - Integrate storage abstraction into application

**Current State:**
- ✅ Database abstraction: Complete and working
- ✅ Storage abstraction: Complete and working
- ✅ Environment detection: Complete
- ✅ Application compiles and runs locally
- ⚠️  Some test failures (pre-existing from Phase 2, not blocking)

---

## Phase 4 Plan: Cloudflare Infrastructure

### Goal
Create the necessary infrastructure files to deploy myCart as a Cloudflare Container Worker.

### Architecture Overview

**Deployment Modes:**
1. **Local CLI** (existing): SQLite + Filesystem storage
2. **Cloudflare Container Workers** (new): D1 + R2 storage

**Container Worker Flow:**
```
Request → Container Worker (Go binary) → D1 Database
                                      → R2 Storage
                                      → Response
```

### Files to Create

#### 1. Dockerfile
**Path:** `Dockerfile`
**Purpose:** Build distroless container image with Go binary

**Key Requirements:**
- Multi-stage build (builder + runtime)
- Use distroless base (`gcr.io/distroless/static-debian12`)
- Copy compiled binary + embedded assets
- Expose port 8080 (Cloudflare Container Workers requirement)
- No CGO (pure Go build)

**Template:**
```dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -ldflags="-s -w" -o mycart ./cmd

FROM gcr.io/distroless/static-debian12
COPY --from=builder /build/mycart /mycart
EXPOSE 8080
ENTRYPOINT ["/mycart", "serve", "--http", ":8080"]
```

#### 2. wrangler.jsonc
**Path:** `wrangler.jsonc`
**Purpose:** Cloudflare deployment configuration

**Template:**
```jsonc
{
  "name": "mycart",
  "type": "container",
  "compatibility_date": "2024-01-01",
  
  "container": {
    "dockerfile": "./Dockerfile",
    "port": 8080
  },
  
  "bindings": [
    {
      "type": "d1",
      "name": "DB",
      "database_id": "<YOUR_D1_DATABASE_ID>"
    },
    {
      "type": "r2_bucket",
      "name": "UPLOADS",
      "bucket_name": "<YOUR_R2_BUCKET_NAME>"
    }
  ],
  
  "env": {
    "CF_WORKER": "true",
    "D1_DATABASE_ID": "<YOUR_D1_DATABASE_ID>",
    "D1_API_TOKEN": "<YOUR_API_TOKEN>",
    "R2_BUCKET_NAME": "<YOUR_R2_BUCKET_NAME>",
    "R2_ACCOUNT_ID": "<YOUR_ACCOUNT_ID>",
    "R2_ACCESS_KEY_ID": "<YOUR_ACCESS_KEY_ID>",
    "R2_SECRET_ACCESS_KEY": "<YOUR_SECRET_KEY>"
  }
}
```

#### 3. .containerignore
**Path:** `.containerignore`
**Purpose:** Exclude unnecessary files from Docker context

**Template:**
```
.git
.github
*.md
docs/
cmd/lc_base/
cmd/lc_uploads/
cmd/lc_digitals/
*.db
*.db-wal
*.db-shm
node_modules/
web/admin/node_modules/
web/site/node_modules/
```

### Implementation Steps

#### Step 1: Create Dockerfile
- Multi-stage build with Alpine builder
- Distroless runtime image
- Copy binary with embedded assets
- Expose port 8080
- Set entrypoint to `mycart serve`

#### Step 2: Create wrangler.jsonc
- Define container configuration
- Set up D1 database binding
- Set up R2 bucket binding
- Configure environment variables
- Set compatibility date

#### Step 3: Create .containerignore
- Exclude development files
- Exclude local data directories
- Exclude git metadata
- Exclude node_modules

#### Step 4: Test Local Docker Build
```bash
# Build Docker image
docker build -t mycart:local .

# Run locally (simulates Cloudflare environment)
docker run -p 8080:8080 mycart:local

# Test endpoints
curl http://localhost:8080/ping
curl http://localhost:8080/api/settings
```

### Testing Strategy

**Local Docker Test:**
1. Build image: `docker build -t mycart:local .`
2. Run container: `docker run -p 8080:8080 mycart:local`
3. Test endpoints: `/ping`, `/api/settings`, `/_/`
4. Verify compilation and startup

**Cloudflare Test (optional):**
1. Create D1 database: `wrangler d1 create mycart-db`
2. Create R2 bucket: `wrangler r2 bucket create mycart-uploads`
3. Update `wrangler.jsonc` with actual IDs
4. Deploy: `wrangler deploy`

### Success Criteria

- ✅ Dockerfile builds successfully
- ✅ Container runs locally on port 8080
- ✅ Application serves requests in container
- ✅ wrangler.jsonc configuration complete
- ✅ .containerignore excludes unnecessary files

### Files Reference

**To Create:**
- `Dockerfile` - Container build configuration
- `wrangler.jsonc` - Cloudflare deployment config
- `.containerignore` - Docker context exclusions

**Current Branch:**
```
feature/cloudflare-container-workers
```

**Test Commands:**
```bash
# Build Docker image
docker build -t mycart:local .

# Run locally
docker run -p 8080:8080 mycart:local

# Test endpoints
curl http://localhost:8080/ping
```

---

## Next Steps

1. Create Dockerfile with multi-stage build
2. Create wrangler.jsonc with bindings
3. Create .containerignore
4. Test local Docker build
5. (Optional) Test Cloudflare deployment

**Ready for Phase 4 implementation.**
