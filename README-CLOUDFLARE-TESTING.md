# Testing Cloudflare Mode Locally

## Quick Test

```bash
# Test Cloudflare runtime detection
./test-cloudflare-mode.sh
```

## Manual Testing

### 1. Simulate Cloudflare Environment

```bash
# Option A: Use environment override
export RUNTIME=cloudflare
export CLOUDFLARE_ACCOUNT_ID=your-account-id
export CLOUDFLARE_API_TOKEN=your-api-token

# Option B: Use Cloudflare environment variable
export CLOUDFLARE_APPLICATION_ID=test-app-id
export CLOUDFLARE_ACCOUNT_ID=your-account-id
export CLOUDFLARE_API_TOKEN=your-api-token
```

### 2. Run the Server

```bash
go run ./cmd serve --dev
```

### 3. Check Runtime Detection

```bash
# Check if Cloudflare mode is detected
curl http://localhost:8080/_/api/maintenance/status | jq '.runtime'

# Expected output:
# {
#   "cloudflare": true
# }
```

### 4. Test Maintenance Flag Path

```bash
# In Cloudflare mode (writable /app):
# Flag: /app/maintenance.flag

# In local mode:
# Flag: ./maintenance.flag

# Enable maintenance
go run ./cmd maintenance enable

# Check status
go run ./cmd maintenance status
```

### 5. Test D1/R2 Operations

```bash
# Backup (will use D1 REST API if in Cloudflare mode)
curl -X POST http://localhost:8080/_/api/maintenance/backup

# Restore
curl -X POST http://localhost:8080/_/api/maintenance/restore \
  -H "Content-Type: application/json" \
  -d '{"backup_path": "./lc_base/backups/backup-20260903-123456.db"}'
```

## Environment Files

### Local Development
```bash
source .env.local
npm run devrun
```

### Cloudflare Simulation
```bash
source .env.cloudflare
npm run devrun
```

## Runtime Detection Logic

The app detects Cloudflare environment by checking:

1. `RUNTIME=cloudflare` (explicit override)
2. `CLOUDFLARE_APPLICATION_ID` (auto-detect)

When in Cloudflare mode:
- Database: Auto-selects D1 (unless DB_TYPE is set)
- Storage: Auto-selects R2 (unless STORAGE_TYPE is set)
- Maintenance flag: /app/maintenance.flag (or ./maintenance.flag if /app not writable)

## Testing Without Real Cloudflare Account

You can test the runtime detection and code paths without a real Cloudflare account by:

1. Setting mock environment variables
2. Server will detect "Cloudflare mode"
3. D1/R2 API calls will fail (no valid credentials), but you can see the detection works

```bash
export CLOUDFLARE_APPLICATION_ID=mock-test
export CLOUDFLARE_ACCOUNT_ID=mock-account
export CLOUDFLARE_API_TOKEN=mock-token

go run ./cmd serve --dev
```

## Actual Cloudflare Deployment

For real Cloudflare deployment:

1. Install Wrangler manually:
   ```bash
   npm install -D wrangler@latest
   ```

2. Configure wrangler.toml with your D1/R2 IDs

3. Set environment variables in Cloudflare dashboard

4. Deploy:
   ```bash
   npm run deploy:cf
   ```

The app will auto-detect it's running in Cloudflare and use D1/R2 automatically.
