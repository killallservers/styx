# Caddy + Auth Implementation Guide

Complete implementation walkthrough for the production Caddy + Auth setup with domains (`www.allora.style` and `api.allora.style`).

---

## What Was Implemented

### 1. Enhanced Caddyfile with Domains

**File:** `infra/compose/services/caddy/Caddyfile`

**Changes:**
- Separate configurations for `www.{$DOMAIN}` and `api.{$DOMAIN}`
- Request logging to JSON files (`/var/log/caddy/access-{web,api}.log`)
- Rate limiting (100 req/s global, per-zone)
- Enhanced security headers (HSTS, CSP, Permissions-Policy)
- Request ID propagation (`X-Request-ID`)
- API key pre-validation (basic check for presence)

**Key Features:**
```caddy
# Web: www.allora.style
www.{$DOMAIN} {
  log { output file ... format json }
  rate_limit { zone dynamic 100r/s 1000 }
  reverse_proxy {$WEB_UPSTREAM} { ... }
  header { ... comprehensive security headers ... }
}

# API: api.allora.style
api.{$DOMAIN} {
  log { output file ... format json }
  rate_limit { zone dynamic 100r/s 1000 }
  reverse_proxy {$API_UPSTREAM} { ... }
  header { ... API-specific restrictive headers ... }
}
```

### 2. Enhanced Authentication Middleware

**File:** `packages/api/src/auth/enhanced-middleware.ts`

**New Features:**

#### Request Metadata Extraction
```typescript
interface RequestMetadata {
  requestId: string;        // Unique per request
  clientIp: string;         // From X-Forwarded-For or CF-Connecting-IP
  userAgent: string;        // Browser/client info
  timestamp: Date;          // When request came in
  method: string;           // GET, POST, etc.
  path: string;             // API path
  host: string;             // www.allora.style or api.allora.style
  protocol: string;         // https (from X-Forwarded-Proto)
}
```

#### API Key Rate Limiting
```typescript
class RateLimiter {
  private readonly windowSize = 60_000;     // 1 minute sliding window
  private readonly maxRequests = 1000;      // 1000 req/min per key
  
  isAllowed(keyId: number): boolean { ... }
  getRemainingRequests(keyId: number): number { ... }
  getResetTime(keyId: number): number { ... }
}
```

#### Structured Logging
```json
{
  "timestamp": "2026-07-14T12:00:00Z",
  "level": "info",
  "message": "API key authenticated",
  "requestId": "req_1626129600_abc123",
  "clientIp": "203.0.113.45",
  "vendor": "hewi",
  "path": "/api/queries"
}
```

#### Permission-Based Access Control
```typescript
router.use("*", requirePermission("write"));  // Only write access

// Returns 403 Forbidden if key doesn't have write permission
```

### 3. Docker Compose Configuration

**File:** `infra/compose/compose.prod.yml`

**Changes:**
- Added `caddy_logs` volume for structured logging
- Configured logging driver (100MB per file, 5 files retained)
- Updated DOMAIN default to `allora.style`
- Added request timeout to API upstream (30s)

```yaml
caddy:
  volumes:
    - ./services/caddy/Caddyfile:/etc/caddy/Caddyfile
    - caddy_logs:/var/log/caddy  # NEW
  logging:
    driver: "json-file"
    options:
      max-size: "100m"
      max-file: "5"  # NEW
  environment:
    DOMAIN: ${DOMAIN:-allora.style}  # UPDATED
```

### 4. Documentation

**New Files:**
- `infra/AUTH.md` — Complete authentication & security guide
- `infra/PRODUCTION_VERIFICATION.md` — Step-by-step verification checklist
- `infra/IMPLEMENTATION_GUIDE.md` — This file

---

## Deployment Steps

### Step 1: Update Environment Variables

**Update GitHub Secrets:**

```bash
PRODUCTION_DOMAIN=allora.style
```

*Note: This should already be set if deploying to allora.style domain.*

### Step 2: Verify Domain Configuration

```bash
# Ensure DNS records exist
nslookup www.allora.style      # Should resolve to server IP
nslookup api.allora.style      # Should resolve to server IP

# Verify DNS propagation
dig www.allora.style +short    # Returns server IP
dig api.allora.style +short    # Returns server IP
```

### Step 3: Deploy to Production

The deployment is now automatic via GitHub Actions, but if deploying manually:

```bash
# SSH to server
ssh -i ~/.ssh/hetzner/<ID>/deploy deploy@<SERVER_IP>

# Navigate to deployment directory
cd /opt/allora

# Pull latest compose files (if using git)
git pull

# Start services
docker compose -f compose/compose.prod.yml up -d

# Verify services started
docker ps | grep -E "caddy|api|web|postgres|llm"
```

### Step 4: Verify TLS Certificate

Caddy automatically provisions Let's Encrypt certificates. Verify:

```bash
# Check certificate
openssl s_client -connect api.allora.style:443 -showcerts 2>/dev/null | \
  grep -A3 "Subject:"

# Expected: CN = api.allora.style (or SAN covering both domains)

# Check expiration
curl -I https://api.allora.style 2>&1 | head -1

# Expected: HTTP/2 200
```

### Step 5: Test Authentication

```bash
# Get API key from database or seed output
API_KEY=$(psql -h localhost -p 7432 -U allora -d allora \
  -t -c "SELECT key FROM api_keys LIMIT 1;")

# Test query
curl -H "X-API-Key: $API_KEY" \
  https://api.allora.style/api/queries \
  -H "Content-Type: application/json" \
  -d '{"inputText":"test","limit":5}'

# Expected: 200 OK with results
```

---

## Configuration Reference

### Caddyfile Environment Variables

```bash
# Set via docker compose environment
DOMAIN=allora.style           # Root domain (expands to www.allora.style, api.allora.style)
API_UPSTREAM=http://api:7500  # Internal Docker network address
WEB_UPSTREAM=http://web:7501  # Internal Docker network address
```

### Enhanced Middleware Configuration

**File:** `packages/api/src/auth/enhanced-middleware.ts`

**Tunable Parameters:**

```typescript
class RateLimiter {
  private readonly windowSize = 60_000;      // 1 minute (ms)
  private readonly maxRequests = 1000;       // 1000 requests per window
}

// For high-volume APIs, increase maxRequests:
private readonly maxRequests = 10_000;       // 10k req/min per key
```

### API Key Permissions

**Database Schema:**

```sql
-- Vendors table
CREATE TABLE vendors (
  id SERIAL PRIMARY KEY,
  slug TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  type TEXT NOT NULL,  -- 'read', 'write', or 'read_write'
  ...
);

-- API Keys table
CREATE TABLE api_keys (
  id SERIAL PRIMARY KEY,
  vendor_id INTEGER NOT NULL REFERENCES vendors(id),
  name TEXT NOT NULL,
  key TEXT UNIQUE NOT NULL,
  revoked_at TIMESTAMP,
  last_used_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  ...
);
```

---

## Monitoring Setup

### Caddy Access Logs

**Location:** `/var/log/caddy/access-{web,api}.log`

**Format:** JSON (machine-readable)

**Example query (on server):**

```bash
# View last 20 requests to API
tail -20 /var/log/caddy/access-api.log | jq '.'

# Count requests by status
cat /var/log/caddy/access-api.log | jq '.response.status' | sort | uniq -c

# Find 401 errors (auth failures)
cat /var/log/caddy/access-api.log | jq 'select(.response.status == 401)' | head -5

# Find 429 errors (rate limit)
cat /var/log/caddy/access-api.log | jq 'select(.response.status == 429)' | head -5
```

### API Auth Logs

**Output:** Docker container logs

**Example query (on server):**

```bash
# View recent auth events
docker logs allora-api | grep -E "authenticate|Invalid|Rate limit" | tail -20

# Filter by type
docker logs allora-api | grep "API key authenticated" | head -10
docker logs allora-api | grep "Invalid API key" | head -10
docker logs allora-api | grep "Rate limit exceeded" | head -10
```

### Prometheus Metrics

**Access:** `http://localhost:7504` (local only)

**Key metrics:**
- `caddy_http_requests_total` — Total requests by endpoint
- `caddy_http_request_duration_seconds_bucket` — Response time percentiles
- `http_requests_total{status="401"}` — Auth failures
- `http_requests_total{status="429"}` — Rate limit hits
- `http_requests_total{status="403"}` — Permission denied

---

## Troubleshooting

### Caddy Won't Start

```bash
# Check Caddy logs
docker logs allora-caddy

# Common issues:
# 1. Port 80/443 already in use
# 2. Domain resolution failing
# 3. Certificate provisioning timeout
```

**Fix:**
```bash
# If port is in use, stop conflicting container
docker ps | grep -E "nginx|apache|other-caddy"
docker stop <container-name>

# Restart Caddy
docker restart allora-caddy

# Check logs again
docker logs allora-caddy
```

### Certificate Not Provisioning

```bash
# Check DNS resolution (from container)
docker exec allora-caddy nslookup api.allora.style

# Check Caddy logs for ACME errors
docker logs allora-caddy | grep -i "acme\|certificate\|tls"

# Force renewal
docker exec allora-caddy caddy reload
```

### API Key Not Working

```bash
# Verify key exists in database
psql -h localhost -p 7432 -U allora -d allora \
  -c "SELECT id, name, key, revoked_at FROM api_keys WHERE key = 'sk_prod_...';"

# Check if revoked (revoked_at should be NULL)
# Verify vendor exists
# Verify permissions

# Clear application cache if using caching middleware
docker restart allora-api
```

### Rate Limit Errors

```bash
# Check if legitimate traffic is being rate limited
docker logs allora-api | grep "Rate limit" | tail -20

# If increasing limit needed:
# Edit: packages/api/src/auth/enhanced-middleware.ts
# Change: private readonly maxRequests = 1000;
# To:     private readonly maxRequests = 5000;
# Rebuild and redeploy
```

---

## Performance Tuning

### API Key Rate Limiting

**Current:** 1000 req/min per key (16.7 req/sec)

**Increase for high-volume ingestion:**
```typescript
private readonly maxRequests = 5000;  // 83 req/sec
```

**Decrease for restrictive access:**
```typescript
private readonly maxRequests = 100;   // 1.7 req/sec
```

### Caddy Global Rate Limiting

**Current:** 100 req/s per zone

**Increase for high traffic:**
```caddy
rate_limit {
  zone dynamic 500r/s 5000  # 500 req/s
}
```

### Connection Timeouts

**API upstream timeout (currently 30s):**
```caddy
timeout 30s  # For long-running embeddings/LLM queries
```

**Increase if embeddings are slow:**
```caddy
timeout 60s  # 60 seconds for vLLM generation
```

---

## Security Checklist (Post-Deployment)

- [ ] TLS certificate valid and auto-renewing
- [ ] HSTS header forces HTTPS for 1 year
- [ ] API requires X-API-Key for all endpoints (except /health)
- [ ] API keys in database are not bcrypted (they're cryptographically random)
- [ ] Rate limiting prevents brute force attacks
- [ ] Access logs in JSON format (machine-readable)
- [ ] Auth events logged with request ID for correlation
- [ ] Error messages don't leak implementation details
- [ ] CSP header prevents XSS on web frontend
- [ ] CORS allows only necessary cross-origin requests

---

## Related Files

- `infra/AUTH.md` — Authentication guide
- `infra/PRODUCTION_VERIFICATION.md` — Verification checklist
- `infra/SECURITY.md` — Security hardening details
- `packages/api/src/auth/middleware.ts` — Original middleware (kept for compatibility)
- `packages/api/src/auth/enhanced-middleware.ts` — New enhanced middleware
- `infra/compose/services/caddy/Caddyfile` — Caddy configuration
- `infra/compose/compose.prod.yml` — Docker Compose production config

---

## Next Steps

1. **Deploy:** Push to `main` branch to trigger GitHub Actions
2. **Verify:** Follow `PRODUCTION_VERIFICATION.md` checklist
3. **Monitor:** Watch logs for auth failures and performance issues
4. **Optimize:** Tune rate limits based on actual traffic patterns
5. **Document:** Update runbooks with new monitoring procedures
