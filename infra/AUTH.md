# Allora Production Authentication & Security Guide

Complete guide for authentication, rate limiting, logging, and security monitoring on production.

## Overview

**Architecture:**
- **API Gateway:** Caddy reverse proxy (www.allora.style, api.allora.style)
- **API Authentication:** X-API-Key header → database lookup
- **Rate Limiting:** Per-API-key at middleware level (1000 req/min), global at Caddy level (100 req/s)
- **Logging:** Caddy access logs (JSON) + API auth logs (structured)
- **TLS:** Let's Encrypt automatic certificate management

**Domains:**
- `www.allora.style` — Web frontend (public)
- `api.allora.style` — API backend (B2B + B2C)

---

## Part 1: API Key Authentication

### How It Works

1. **Client sends request with API key:**
   ```bash
   curl -H "X-API-Key: sk_prod_abc123def456" \
     https://api.allora.style/api/queries \
     -H "Content-Type: application/json" \
     -d '{"inputText":"quiet luxury summer evening","limit":5}'
   ```

2. **Caddy receives request:**
   - Adds `X-Forwarded-Proto: https`
   - Adds `X-Forwarded-For: <client-ip>`
   - Adds `X-Request-ID: req_<timestamp>_<random>`
   - Routes to API backend

3. **API middleware verifies key:**
   - Extracts `X-API-Key` header
   - Looks up key in `api_keys` table
   - Checks if key is revoked
   - Validates vendor permissions
   - Applies rate limiting

4. **Response includes rate limit info:**
   ```json
   {
     "status": "ok",
     "results": [...],
     "X-RateLimit-Remaining": 999
   }
   ```

### API Key Permissions

Three permission levels:

| Permission | Read | Write | Use Case |
|-----------|------|-------|----------|
| `read` | ✅ | ❌ | Query products, search |
| `write` | ❌ | ✅ | Create/update products (reserved) |
| `read_write` | ✅ | ✅ | Full access (ingestion, queries) |

**Permission enforcement:**
```typescript
// Check permission in route handler
const apiKey = c.get("apiKey");
if (!hasPermission(apiKey.permissions, "write")) {
  return c.json({ error: "Forbidden" }, 403);
}
```

### Creating API Keys

**From database (manual):**
```sql
-- Create vendor
INSERT INTO vendors (slug, name, type)
VALUES ('hewi', 'Hardly Ever Worn It', 'read_write');

-- Create API key
INSERT INTO api_keys (vendor_id, name, key)
VALUES (1, 'HEWI Ingestion', 'sk_prod_abc123def456');
```

**From application (recommended):**
```bash
# Run seed script
bun run db:seed

# Output will show the generated API key
```

---

## Part 2: Rate Limiting

### Two-Layer Rate Limiting

**Layer 1: Caddy (100 req/s global)**
- Protects API from traffic spikes
- Applied to all requests on `api.allora.style`
- Returns `429 Too Many Requests` when exceeded

**Layer 2: API Middleware (1000 req/min per key)**
- Per-API-key rate limiting
- Protects against individual key abuse
- Tracked in memory (per application instance)
- Returns `429 Too Many Requests` with `Retry-After` header

### Rate Limit Response

```json
{
  "error": "Too many requests",
  "message": "Rate limit exceeded for this API key",
  "retryAfter": 30
}
```

**Headers:**
- `X-RateLimit-Remaining: 999` — Requests remaining in current window
- `X-RateLimit-Reset: 1626129600` — Unix timestamp when limit resets
- `Retry-After: 30` — Seconds to wait before retrying

### Disabling Rate Limiting (Development Only)

**In enhanced-middleware.ts:**
```typescript
// Set to very high number for testing
private readonly maxRequests = 100_000; // 100k req/min
```

---

## Part 3: Logging & Monitoring

### Caddy Access Logs

**Location:** `/var/log/caddy/access-{web,api}.log`

**Format:** JSON (structured, machine-readable)

**Example log entry:**
```json
{
  "ts": 1626129600.123,
  "request": {
    "method": "POST",
    "uri": "/api/queries",
    "proto": "HTTP/2.0",
    "remote_ip": "203.0.113.45",
    "remote_port": "54321",
    "host": "api.allora.style",
    "headers": {
      "User-Agent": "curl/7.68.0",
      "X-Request-ID": "req_1626129600123_abc123",
      "X-API-Key": "sk_prod_***" // Redacted
    }
  },
  "response": {
    "status": 200,
    "size": 1024
  },
  "duration": 0.123
}
```

### API Auth Logs

**Format:** JSON (structured)

**Example log entries:**
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

{
  "timestamp": "2026-07-14T12:00:01Z",
  "level": "warn",
  "message": "Rate limit exceeded",
  "requestId": "req_1626129601_def456",
  "clientIp": "203.0.113.46",
  "vendor": "hewi",
  "path": "/api/queries",
  "statusCode": 429,
  "reason": "Rate limit exceeded"
}

{
  "timestamp": "2026-07-14T12:00:02Z",
  "level": "warn",
  "message": "Invalid API key",
  "requestId": "req_1626129602_ghi789",
  "clientIp": "203.0.113.47",
  "path": "/api/queries",
  "statusCode": 401,
  "reason": "API key not found or revoked"
}
```

### Monitoring Dashboard

**Prometheus metrics available at:**
```
http://localhost:7504/metrics
```

**Key metrics to track:**
- `caddy_http_requests_total` — Total requests by endpoint
- `caddy_http_request_duration_seconds` — Response time distribution
- `http_requests_total{status="401"}` — Authentication failures
- `http_requests_total{status="429"}` — Rate limit hits
- `http_requests_total{status="403"}` — Permission denials

---

## Part 4: Security Hardening Checklist

### ✅ Network Layer

- [x] Hetzner Cloud Firewall restricts external access to 22 (SSH), 80 (HTTP), 443 (HTTPS)
- [x] Internal ports (5432, 7502, 7503, 9090) not exposed externally
- [x] ufw on host blocks all incoming except SSH/HTTP/HTTPS
- [x] Docker services bind to 127.0.0.1 except Caddy (0.0.0.0)

### ✅ TLS/HTTPS

- [ ] Let's Encrypt certificate provisioned (automatic via Caddy)
- [ ] Certificate auto-renewal configured (Caddy handles this)
- [ ] HSTS header enforces HTTPS (max-age=31536000)
- [ ] TLS version ≥ 1.2 only

### ✅ API Key Security

- [x] API keys stored as-is in database (bcrypt not needed, they're cryptographically random)
- [x] API keys never logged (redacted in Caddy logs)
- [x] API key revocation supported via `revokedAt` column
- [x] Revoked keys rejected immediately
- [ ] API key rotation policy enforced (recommend: quarterly)
- [ ] Rate limiting prevents brute force attacks

### ✅ Headers

- [x] `Strict-Transport-Security` (HSTS) — Forces HTTPS
- [x] `X-Content-Type-Options: nosniff` — Prevents MIME sniffing
- [x] `X-Frame-Options: DENY` (API) / `SAMEORIGIN` (Web) — Prevents clickjacking
- [x] `X-XSS-Protection` — Additional XSS protection
- [x] `Referrer-Policy: strict-origin-when-cross-origin` — Limits referrer info
- [x] `Permissions-Policy` — Disables unnecessary browser features
- [x] `Content-Security-Policy` — Restricts resource loading

### ✅ CORS

- [x] Web frontend can reach API (CSP allows `connect-src 'self' https://api.allora.style`)
- [x] External origins cannot access API (no wildcard CORS)
- [ ] Verify CORS preflight working correctly

### ✅ Authentication

- [x] API key required for all `/api/*` endpoints
- [x] Invalid keys rejected with 401
- [x] Missing keys rejected with 401
- [x] Permission denied returns 403 (not 401)
- [ ] Error messages don't leak information (generic "Invalid or revoked API key")

### ✅ Logging & Monitoring

- [x] Caddy logs all requests (JSON format)
- [x] API logs authentication events
- [x] Rate limit hits logged as warnings
- [x] Failed authentications logged
- [ ] Logs rotated (100MB per file, 5 files retained)
- [ ] Log aggregation configured (optional: centralized logging)

### ✅ Infrastructure

- [x] SSH hardening: key-only auth, no passwords
- [x] Fail2ban: 3 failed attempts = 1 hour ban
- [x] Automatic security updates via unattended-upgrades
- [x] Docker daemon only accessible to root
- [ ] Backup strategy configured
- [ ] Disaster recovery plan documented

---

## Part 5: Common Tasks

### Viewing Logs

**Web frontend logs:**
```bash
# SSH to server
ssh -i ~/.ssh/hetzner/<ID>/deploy deploy@<IP>

# View recent requests
docker logs -f allora-caddy | grep "api.allora.style"

# View specific file
tail -100 /var/log/caddy/access-api.log | jq .
```

**API auth events:**
```bash
docker logs allora-api | grep "authenticate\|Invalid\|Rate limit"
```

### Monitoring Metrics

**Check API health:**
```bash
curl https://api.allora.style/health
```

**View rate limit status:**
```bash
# Make request and check headers
curl -H "X-API-Key: sk_prod_abc123" \
  https://api.allora.style/api/queries \
  -v 2>&1 | grep -i "rate-limit"
```

### Revoking API Keys

**Mark key as revoked:**
```sql
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = <key_id>;
```

**Verify revocation:**
```bash
# This should now return 401
curl -H "X-API-Key: sk_prod_abc123" \
  https://api.allora.style/api/queries
```

### Rotating API Keys

**Current approach (manual):**
1. Create new key in database
2. Provide new key to client
3. Wait for client to switch (grace period)
4. Revoke old key

**Recommended: Automated key rotation**
- Generate new key quarterly
- Notify vendor 2 weeks in advance
- Set grace period of 1 week
- Auto-revoke old key after grace period

---

## Part 6: Troubleshooting

### 401 Unauthorized

**Check:**
1. Is X-API-Key header present?
2. Is key value correct (copy-paste)?
3. Is key in database?
4. Is key revoked (`revokedAt IS NOT NULL`)?

**Debug:**
```bash
# Check if key exists
curl -H "X-API-Key: sk_prod_abc123" \
  https://api.allora.style/api/queries \
  -v

# Check API logs
docker logs allora-api | grep "Invalid API key"
```

### 429 Too Many Requests

**Possible causes:**
1. Rate limit exceeded at Caddy level (100 req/s global)
2. Rate limit exceeded at API level (1000 req/min per key)
3. Legitimate traffic spike

**Check:**
```bash
# Check Caddy logs for rate limit hits
docker logs allora-caddy | grep "429"

# Check remaining requests in header
curl -H "X-API-Key: sk_prod_abc123" \
  https://api.allora.style/api/queries \
  -i | grep -i "rate-limit"
```

**Fix:**
1. Wait for rate limit window to reset
2. If legitimate traffic, increase rate limit (edit enhanced-middleware.ts)
3. If abuse, revoke API key

### TLS Certificate Issues

**Check certificate:**
```bash
# View current cert
curl --insecure -v https://api.allora.style 2>&1 | grep -A5 "certificate"

# Check expiration
openssl s_client -connect api.allora.style:443 -showcerts 2>/dev/null | \
  openssl x509 -noout -dates
```

**Caddy auto-renews 30 days before expiration.**

If renewal fails:
```bash
# Check Caddy logs
docker logs allora-caddy | grep -i "certificate\|tls\|acme"

# Force renewal
docker exec allora-caddy caddy reload
```

### API Slow/Timing Out

**Check:**
1. API container running? `docker ps | grep allora-api`
2. Database connection healthy? `docker logs allora-api | grep "database"`
3. LLM/Embeddings containers running? `docker ps | grep llm\|embeddings`

**Metrics:**
```bash
# Check response times in Caddy logs
tail -100 /var/log/caddy/access-api.log | \
  jq '.response.duration' | sort -n | tail -20
```

---

## Part 7: Security Incident Response

### Suspected Compromised API Key

1. **Immediate:** Revoke the key
   ```sql
   UPDATE api_keys SET revoked_at = NOW() WHERE id = <key_id>;
   ```

2. **Investigation:** Check audit logs
   ```bash
   # View all requests with that key (last 1000 lines)
   tail -1000 /var/log/caddy/access-api.log | \
     jq 'select(.request.headers["X-API-Key"] == "sk_prod_...")'
   ```

3. **Containment:** Review what data was accessed
   - Check queries made (from logs)
   - Check items accessed (from DB query logs)

4. **Communication:** Notify affected vendor
   - What was exposed
   - When compromise was detected
   - What actions were taken
   - What they should do (new key, etc.)

5. **Prevention:** Post-incident review
   - How was key compromised? (exposed in code, git history, email, etc.)
   - How to prevent in future? (education, key rotation, monitoring, etc.)
   - Were other keys exposed? (same compromise vector)

### DDoS or Traffic Spike

1. **Detection:** Caddy logs show high 429 rate limit hits
2. **Immediate:** No action needed (rate limiting prevents damage)
3. **Investigation:** Where is traffic coming from?
   ```bash
   tail -1000 /var/log/caddy/access-api.log | \
     jq '.request.remote_ip' | sort | uniq -c | sort -rn | head -10
   ```

4. **If malicious:**
   - Identify attacking IP ranges
   - Add Hetzner Firewall rules to block
   - Revoke API keys used in attack

5. **If legitimate:** Increase rate limits temporarily
   ```typescript
   // In enhanced-middleware.ts
   private readonly maxRequests = 5000; // Increased from 1000
   ```

---

## Related Documentation

- `infra/SECURITY.md` — Network security, firewall setup, SSH hardening
- `infra/DEPLOY.md` — Deployment procedures
- `packages/api/src/auth/enhanced-middleware.ts` — Implementation details
- `infra/compose/services/caddy/Caddyfile` — Caddy configuration
