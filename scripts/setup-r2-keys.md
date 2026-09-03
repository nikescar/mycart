# R2 Access Keys Setup

## Why Manual Setup?

Cloudflare's API can manage R2 buckets but doesn't currently support programmatic creation of R2 access keys (S3-compatible credentials). These must be created through the dashboard.

## Steps

1. **Go to Cloudflare Dashboard:**
   https://dash.cloudflare.com

2. **Navigate to R2 API Tokens:**
   - Click on "R2" in the sidebar
   - Click "Manage R2 API Tokens"

3. **Create New API Token:**
   - Click "Create API Token"
   - Name: `mycart-test-r2`
   - Permissions: Select "Admin Read & Write" for R2
   - Click "Create API Token"

4. **Copy Credentials:**
   ```
   Access Key ID: <copy this>
   Secret Access Key: <copy this>
   ```

5. **Update .env.test:**
   ```bash
   CLOUDFLARE_TEST_R2_ACCESS_KEY_ID=<paste Access Key ID>
   CLOUDFLARE_TEST_R2_SECRET_ACCESS_KEY=<paste Secret Access Key>
   ```

6. **Run Tests:**
   ```bash
   make test-cloudflare
   ```

## Current .env.test Status

Run this to generate the base .env.test file:
```bash
cat > .env.test << 'EOF'
# Cloudflare D1/R2 Integration Test Credentials

# Cloudflare Account ID (get from dashboard)
CLOUDFLARE_TEST_ACCOUNT_ID=your-account-id-here

# API Token with D1:Edit and R2:Edit permissions (create in dashboard)
CLOUDFLARE_TEST_API_TOKEN=your-api-token-here

# R2 Access Keys (create these manually via dashboard)
CLOUDFLARE_TEST_R2_ACCESS_KEY_ID=your-r2-access-key-id
CLOUDFLARE_TEST_R2_SECRET_ACCESS_KEY=your-r2-secret-access-key
EOF
```
