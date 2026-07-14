# Production Verification Checklist

Complete verification guide for production deployment. Run through all checks before marking deployment as complete.

---

## Part 1: Infrastructure Readiness

### Domain Configuration

- [ ] DNS A record points to server IP: `www.allora.style` → `<SERVER_IP>`
- [ ] DNS A record points to server IP: `api.allora.style` → `<SERVER_IP>`
- [ ] DNS propagated (give ~5 minutes): `nslookup www.allora.style`
- [ ] Domain shown in GitHub secret: `PRODUCTION_DOMAIN=allora.style`

### Server Access

```bash
# Test SSH access
ssh -i ~/.ssh/hetzner/<ID>/deploy deploy@<SERVER_IP>

# Expected: SSH login successful
```

### Firewall Configuration

```bash
# Verify firewall rules (from Hetzner console)
# Should allow: 22 (SSH), 80 (HTTP), 443 (HTTPS)
# Should block: All other ports

# Test from local machine
curl http://www.allora.style  # Should respond (redirect to HTTPS)
curl https://www.allora.style # Should respond with web content
curl https://api.allora.style  # Should respond with 401 (missing API key)
```

---

## Part 2: TLS/HTTPS Configuration

### Certificate Provisioning

```bash
# Check TLS certificate
openssl s_client -connect www.allora.style:443 -showcerts 2>/dev/null | \
  grep -A 3 "Subject:"

# Expected output:
# Subject: CN = www.allora.style
# (OR Subject Alternative Names include www.allora.style, api.allora.style)
```

- [ ] Certificate issued by Let's Encrypt
- [ ] Certificate covers both domains (www.allora.style, api.allora.style)
- [ ] Certificate valid for 90 days
- [ ] Certificate not expired

### HTTPS Redirect

```bash
# Test HTTP → HTTPS redirect
curl -i http://www.allora.style 2>&1 | head -15

# Expected:
# HTTP/1.1 307 Temporary Redirect
# Location: https://www.allora.style
```

- [ ] HTTP requests redirect to HTTPS
- [ ] Status code is 301 or 307 (not 302)

### Security Headers

```bash
# Test HSTS header
curl -i https://www.allora.style 2>&1 | grep -i "strict-transport"

# Expected:
# Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

- [ ] HSTS header present with `max-age=31536000`
- [ ] HSTS includes `preload` directive
- [ ] X-Content-Type-Options: nosniff
- [ ] X-Frame-Options: SAMEORIGIN (web) or DENY (api)
- [ ] X-XSS-Protection header present

---

## Part 3: API Authentication

### Health Endpoint (No Auth Required)

```bash
# Test health endpoint
curl https://api.allora.style/health

# Expected: 200 OK
# Response: {"status":"ok"}
```

- [ ] Health endpoint responds without authentication
- [ ] Status code is 200

### API Key Authentication

```bash
# Test with valid API key (from seed output or database)
API_KEY="sk_prod_..."

curl -H "X-API-Key: $API_KEY" \
  https://api.allora.style/api/queries \
  -H "Content-Type: application/json" \
  -d '{"inputText":"test","limit":5}'

# Expected: 200 OK (or specific response from endpoint)
```

- [ ] Valid API key accepted (200 OK)
- [ ] Response includes query results

### Missing API Key (Should Fail)

```bash
# Test without API key
curl https://api.allora.style/api/queries \
  -H "Content-Type: application/json" \
  -d '{"inputText":"test","limit":5}'

# Expected: 401 Unauthorized
# Response: {"error":"Unauthorized","message":"X-API-Key header is required"}
```

- [ ] Missing API key returns 401
- [ ] Error message is clear but not leaking info

### Invalid API Key (Should Fail)

```bash
# Test with invalid key
curl -H "X-API-Key: invalid_key_12345" \
  https://api.allora.style/api/queries \
  -H "Content-Type: application/json" \
  -d '{"inputText":"test","limit":5}'

# Expected: 401 Unauthorized
# Response: {"error":"Unauthorized","message":"Invalid or revoked API key"}
```

- [ ] Invalid API key returns 401
- [ ] Does not reveal whether key exists or is revoked

---

## Part 4: Rate Limiting

### Caddy Global Rate Limiting (100 req/s)

```bash
# Rapid requests to trigger rate limit
for i in {1..200}; do
  curl -s https://www.allora.style/health > /dev/null &
done
wait

# Some requests should get 429 Too Many Requests
```

- [ ] Rate limit kicks in at ~100 req/s
- [ ] Status code is 429
- [ ] Request can retry after short delay

### API Key Rate Limiting (1000 req/min)

```bash
# Make 1001 requests with same API key rapidly
API_KEY="sk_prod_..."

for i in {1..1001}; do
  curl -s -H "X-API-Key: $API_KEY" \
    https://api.allora.style/api/queries \
    -H "Content-Type: application/json" \
    -d '{"inputText":"test","limit":1}' > /dev/null &
done
wait

# Request 1001 should get 429
```

- [ ] Rate limit per key applied after 1000 requests/min
- [ ] Response includes `Retry-After` header
- [ ] Other API keys not affected

---

## Part 5: CORS & Headers

### CORS Preflight (Web to API)

```bash
# Simulate browser preflight request
curl -i -X OPTIONS \
  https://api.allora.style/api/queries \
  -H "Origin: https://www.allora.style" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: X-API-Key"

# Expected: 200 OK (Caddy passes through to API)
```

- [ ] Preflight request succeeds
- [ ] CORS headers included in response

### CSP Header (API)

```bash
curl -i https://api.allora.style/health | grep -i "content-security"

# Expected:
# Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; upgrade-insecure-requests
```

- [ ] CSP header prevents inline scripts on API
- [ ] CSP restrictive (default-src 'none')

### CSP Header (Web)

```bash
curl -i https://www.allora.style/ | grep -i "content-security"

# Expected:
# Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; ...
```

- [ ] CSP allows self and necessary resources
- [ ] Blocks external scripts (except necessary APIs)

---

## Part 6: Logging & Monitoring

### Caddy Access Logs

```bash
# SSH to server
ssh -i ~/.ssh/hetzner/<ID>/deploy deploy@<SERVER_IP>

# View web access logs
docker logs allora-caddy | grep "www.allora.style" | tail -10 | jq .

# View API access logs
docker logs allora-caddy | grep "api.allora.style" | tail -10 | jq .
```

- [ ] Access logs in JSON format
- [ ] Logs include request timestamp, method, status
- [ ] Logs include client IP (X-Forwarded-For)
- [ ] Logs include request duration

### API Auth Logs

```bash
# View API authentication logs
docker logs allora-api | grep -i "authenticate" | tail -10

# Should show successful authentications and failures
```

- [ ] Auth events logged (successful & failed)
- [ ] Failed auth logs include reason
- [ ] No sensitive data in logs (API keys redacted)

### Prometheus Metrics

```bash
# Check Prometheus is running
curl http://localhost:7504/api/v1/targets

# Or from outside
ssh -i ~/.ssh/hetzner/<ID>/deploy deploy@<SERVER_IP>
curl http://localhost:7504/api/v1/targets

# Expected: 200 OK with target list
```

- [ ] Prometheus running and healthy
- [ ] Scraping targets (Caddy, API, etc.)
- [ ] Metrics available at `http://localhost:7504`

### Grafana Dashboards

```bash
# Verify Grafana is running
curl http://localhost:7505/api/health

# Or from outside (via Caddy, if configured)
curl https://api.allora.style/grafana/api/health
```

- [ ] Grafana running (port 7505 locally)
- [ ] Can login with admin/password

---

## Part 7: Database & Backend Services

### API Database

```bash
ssh -i ~/.ssh/hetzner/<ID>/deploy deploy@<SERVER_IP>

# Connect to API database
psql -h localhost -p 7432 -U allora -d allora -c "SELECT COUNT(*) FROM vendors;"

# Expected: Returns number of vendors
```

- [ ] API database accessible
- [ ] Vendors table has entries (at least 1)
- [ ] API keys table populated

### Web Database

```bash
# Connect to web database
psql -h localhost -p 7433 -U allora -d allora_web -c "SELECT COUNT(*) FROM users;"

# Expected: Returns user count (can be 0 for new deployment)
```

- [ ] Web database accessible
- [ ] Can query tables

### LLM Service (Ministral)

```bash
# Test LLM endpoint
curl http://localhost:7502/v1/models

# Expected: 200 OK with model list
# Response: {"object":"list","data":[{"id":"mistralai/Ministral-3-3B-Instruct-2512","object":"model"}]}
```

- [ ] LLM service running
- [ ] Model loaded (`mistralai/Ministral-3-3B-Instruct-2512`)
- [ ] Health endpoint responds

### Embeddings Service (Qwen3)

```bash
# Test embeddings endpoint
curl -X POST http://localhost:7503/v1/embeddings \
  -H "Content-Type: application/json" \
  -d '{"model":"Qwen3-VL-Embedding-2B","input":"test"}'

# Expected: 200 OK with embedding vector
```

- [ ] Embeddings service running
- [ ] Model loaded (`Qwen3-VL-Embedding-2B`)
- [ ] Can generate embeddings

---

## Part 8: End-to-End Integration Test

### Full Query Flow

```bash
# Set API key from database or seed output
API_KEY="sk_prod_..."

# Test query endpoint
curl -H "X-API-Key: $API_KEY" \
  https://api.allora.style/api/queries \
  -H "Content-Type: application/json" \
  -d '{
    "inputText": "quiet luxury summer evening",
    "limit": 5,
    "threshold": 0.5
  }' | jq .

# Expected: 200 OK with results
# Response: {
#   "id": "query-...",
#   "inputText": "quiet luxury summer evening",
#   "results": [
#     {
#       "itemId": "...",
#       "name": "...",
#       "category": "...",
#       "similarity": 0.92,
#       ...
#     }
#   ],
#   "totalResults": 5,
#   "executionTimeMs": 245
# }
```

- [ ] Query endpoint returns 200 OK
- [ ] Response includes results array
- [ ] Similarity scores between 0 and 1
- [ ] Execution time reasonable (<1s for small datasets)

### Verify Request Headers

```bash
# Check that request headers are preserved
curl -v -H "X-API-Key: $API_KEY" \
  https://api.allora.style/api/queries \
  -H "Content-Type: application/json" \
  -d '{"inputText":"test","limit":1}' 2>&1 | grep -E "X-Request-ID|X-Forwarded|X-RateLimit"

# Expected:
# X-Request-ID: req_... (added by Caddy)
# X-Forwarded-Proto: https (added by Caddy)
# X-Forwarded-For: 203.0.113.45 (added by Caddy)
```

- [ ] X-Request-ID added to every request
- [ ] X-Forwarded-Proto set to https
- [ ] X-Forwarded-For contains client IP

---

## Part 9: Error Scenarios

### Database Unavailable

```bash
# Stop API database (simulated failure)
docker stop allora-api-db

# Try API query (will fail)
curl https://api.allora.style/health

# Restart database
docker start allora-api-db
```

- [ ] API returns 500 when database unavailable
- [ ] Error message is generic (not exposing implementation)
- [ ] Service recovers when database restored

### LLM Model Unavailable

```bash
# Stop LLM service (simulated)
docker stop allora-llm

# Try query (will timeout or fail)
curl https://api.allora.style/api/queries \
  -H "X-API-Key: $API_KEY" \
  -d '{"inputText":"test"}'

# Restart LLM
docker start allora-llm
```

- [ ] API returns error when LLM unavailable
- [ ] Error message is clear (but doesn't expose internals)
- [ ] Service recovers when LLM restarted

---

## Part 10: Performance Baseline

### Response Time (P50, P95, P99)

```bash
# Make 100 requests and measure response time
API_KEY="sk_prod_..."

for i in {1..100}; do
  curl -H "X-API-Key: $API_KEY" \
    https://api.allora.style/api/queries \
    -H "Content-Type: application/json" \
    -d '{"inputText":"test","limit":5}' \
    -w "%{time_total}\n" \
    -o /dev/null -s
done | sort -n

# Expected:
# P50: ~100-200ms
# P95: ~300-500ms
# P99: ~800-1200ms
```

- [ ] P50 latency < 200ms
- [ ] P95 latency < 500ms
- [ ] P99 latency < 1200ms

### Throughput (req/sec)

```bash
# Measure sustained throughput for 10 seconds
API_KEY="sk_prod_..."

ab -n 100 -c 10 \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -p query.json \
  https://api.allora.style/api/queries

# Where query.json contains: {"inputText":"test","limit":5}

# Expected:
# Requests per second: ~10-50 (depends on dataset size & LLM performance)
# Failed requests: 0
```

- [ ] Throughput > 1 req/sec
- [ ] No failed requests under sustained load
- [ ] Response time stable (no spikes)

---

## Part 11: Deployment Readiness Sign-Off

### Checklist Summary

- [ ] All infrastructure checks passed (Part 1)
- [ ] TLS/HTTPS configured and working (Part 2)
- [ ] API authentication working (Part 3)
- [ ] Rate limiting functional (Part 4)
- [ ] CORS and headers correct (Part 5)
- [ ] Logging and monitoring configured (Part 6)
- [ ] All backend services healthy (Part 7)
- [ ] End-to-end integration test passed (Part 8)
- [ ] Error scenarios handled gracefully (Part 9)
- [ ] Performance baseline acceptable (Part 10)

### Issues Found (if any)

Document any issues found during verification:

| Issue | Severity | Action | Status |
|-------|----------|--------|--------|
| (describe) | Critical/High/Medium/Low | (fix/document/monitor) | (done/pending) |

### Sign-Off

- **Verified by:** (Name)
- **Date:** (YYYY-MM-DD)
- **Notes:** (Any additional observations)

**Production deployment is ready when all checks are ✓ and no critical/high issues remain.**

---

## Ongoing Monitoring

After deployment, monitor these metrics:

- **Error rate:** `docker logs allora-api | grep "error" | wc -l`
- **Rate limit hits:** `docker logs allora-caddy | grep "429" | wc -l`
- **Auth failures:** `docker logs allora-api | grep "Invalid API key" | wc -l`
- **Response time:** Check Prometheus dashboard
- **Disk usage:** `df -h /opt/` (model cache size)
- **Memory usage:** `docker stats --no-stream`

---

## Related Documentation

- `infra/AUTH.md` — Authentication guide
- `infra/SECURITY.md` — Security hardening details
- `infra/DEPLOY.md` — Deployment procedures
