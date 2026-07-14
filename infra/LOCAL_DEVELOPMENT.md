# Local Development Setup

Run the complete Allora stack locally with `compose.dev.yml`, which mirrors production as closely as possible.

**This is for local testing BEFORE deploying to the production server.**

For production deployment, see: [`DEPLOY.md`](DEPLOY.md).

## Quick Start

### Option A: With GPU (Full Stack)

If you have NVIDIA GPU:

```bash
cd infra/compose
docker compose -f compose.dev.yml --profile gpu up -d
```

This runs ALL services:
- ✅ Ministral-3-3B (LLM, GPU-accelerated)
- ✅ Qwen3-VL-Embedding (Embeddings, GPU-accelerated)
- ✅ PostgreSQL (API database)
- ✅ PostgreSQL (Web database)
- ✅ API service (Bun)
- ✅ Web service (Bun)
- ✅ Prometheus (Metrics)
- ✅ Grafana (Dashboards)

### Option B: Without GPU (Database + API/Web Only)

If you don't have NVIDIA GPU or want to skip LLM/embeddings:

```bash
cd infra/compose
docker compose -f compose.dev.yml up -d
```

This runs:
- ✅ PostgreSQL (API database)
- ✅ PostgreSQL (Web database)
- ✅ API service (Bun)
- ✅ Web service (Bun)
- ✅ Prometheus (Metrics)
- ✅ Grafana (Dashboards)

API will fail health checks (LLM/embeddings unavailable), but database operations work.

### Option C: With Caddy (TLS Testing)

If you want to test HTTPS locally with self-signed certificates:

```bash
cd infra/compose
docker compose -f compose.dev.yml --profile gpu --profile caddy up -d
```

Adds Caddy reverse proxy (but uses `localhost` for domain, so TLS won't verify in browser).

---

## Configuration

### Environment Variables

Default values are in `.env.dev`:

```bash
# Default is fine for local dev
POSTGRES_PASSWORD=allora
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin
```

To override:

```bash
# Option 1: Edit .env.dev
nano .env.dev

# Option 2: Pass via environment
export POSTGRES_PASSWORD=my-secure-password
docker compose -f compose.dev.yml up -d

# Option 3: Docker command
docker compose -f compose.dev.yml -e POSTGRES_PASSWORD=custom up -d
```

---

## Service Ports

All services are bound to `127.0.0.1` (localhost-only), matching production security:

| Service | Port | Local URL | Notes |
|---------|------|-----------|-------|
| **API** | 7500 | http://localhost:7500 | Bun API service |
| **Web** | 7501 | http://localhost:7501 | Bun web app |
| **PostgreSQL (API)** | 5432 | localhost:5432 | Use psql or tools |
| **PostgreSQL (Web)** | 5433 | localhost:5433 | Use psql or tools |
| **Ministral LLM** | 7502 | http://localhost:7502/v1/chat/completions | OpenAI-compatible |
| **Qwen3 Embeddings** | 7503 | http://localhost:7503/v1/embedding | OpenAI-compatible |
| **Prometheus** | 9090 | http://localhost:9090 | Metrics dashboard |
| **Grafana** | 3000 | http://localhost:3000 | Visualization (admin/admin) |

---

## Accessing Services

### API Service

```bash
# Health check
curl http://localhost:7500/health

# Test endpoint
curl -X POST http://localhost:7500/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"test","description":"test product"}'
```

### Web Service

```bash
# Home page
open http://localhost:7501

# Or in browser
firefox http://localhost:7501
```

### Prometheus

```bash
# Metrics dashboard
open http://localhost:9090
```

### Grafana

```bash
# Dashboards (admin/admin)
open http://localhost:3000
```

### PostgreSQL

Connect with any PostgreSQL client:

```bash
# API database
psql -h localhost -U allora -d allora

# Web database
psql -h localhost -p 5433 -U allora -d allora_web

# Inside Docker
docker compose -f compose.dev.yml exec api-db psql -U allora -d allora
docker compose -f compose.dev.yml exec web-db psql -U allora -d allora_web
```

### LLM/Embeddings

```bash
# Test Ministral LLM
curl http://localhost:7502/v1/models

curl -X POST http://localhost:7502/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello"}],
    "temperature": 0.1
  }'

# Test Qwen3 Embeddings
curl http://localhost:7503/v1/models

curl -X POST http://localhost:7503/v1/embedding \
  -H "Content-Type: application/json" \
  -d '{"input":"test text"}'
```

---

## Common Commands

### View Logs

```bash
# All services
docker compose -f compose.dev.yml logs -f

# Specific service
docker compose -f compose.dev.yml logs -f api
docker compose -f compose.dev.yml logs -f web
docker compose -f compose.dev.yml logs -f llm

# Follow + tail (last 50 lines)
docker compose -f compose.dev.yml logs --tail 50 -f
```

### Check Service Status

```bash
# List all services
docker compose -f compose.dev.yml ps

# Expected output (all Up, or Exited if not using --profile gpu)
# NAME                    STATUS
# allora-api-db-dev       Up
# allora-web-db-dev       Up
# allora-api-dev          Up
# allora-web-dev          Up
# allora-llm-dev          Up (if --profile gpu)
# allora-embeddings-dev   Up (if --profile gpu)
# allora-prometheus-dev   Up
# allora-grafana-dev      Up
```

### Stop/Start Services

```bash
# Stop all
docker compose -f compose.dev.yml down

# Stop and remove volumes (WARNING: deletes data!)
docker compose -f compose.dev.yml down -v

# Restart single service
docker compose -f compose.dev.yml restart api

# Start specific services only
docker compose -f compose.dev.yml up -d api-db api web-db prometheus
```

### View Resource Usage

```bash
docker stats

# Or in compose
docker compose -f compose.dev.yml stats
```

### Clean Up

```bash
# Remove stopped containers
docker compose -f compose.dev.yml rm

# Prune unused images/volumes
docker image prune
docker volume prune
docker system prune
```

---

## Development Workflow

### 1. Start Stack

```bash
cd infra/compose
docker compose -f compose.dev.yml --profile gpu up -d
```

### 2. Watch Logs

```bash
docker compose -f compose.dev.yml logs -f api web
```

### 3. Make Code Changes

Edit code in `packages/api/` or `packages/web/`

### 4. Rebuild Services

```bash
# Rebuild without stopping
docker compose -f compose.dev.yml build api web

# Or restart services (if using auto-reload)
docker compose -f compose.dev.yml restart api web
```

### 5. Test Changes

```bash
# API
curl http://localhost:7500/api/...

# Web
open http://localhost:7501
```

### 6. View Metrics

```bash
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
# Check container stats: docker stats
```

---

## Troubleshooting

### GPU Not Available

```bash
# Check NVIDIA Docker support
docker run --rm --runtime=nvidia nvidia/cuda:12-runtime nvidia-smi

# If error, install NVIDIA Container Toolkit:
# https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html

# Or run without GPU:
docker compose -f compose.dev.yml up -d  # Omit --profile gpu
```

### Database Connection Refused

```bash
# Check if databases are running
docker compose -f compose.dev.yml ps api-db web-db

# Check database logs
docker compose -f compose.dev.yml logs api-db

# Try connecting
docker compose -f compose.dev.yml exec api-db psql -U allora -d allora
```

### API Service Won't Start

```bash
# Check logs
docker compose -f compose.dev.yml logs api

# Common issues:
# - LLM/embeddings not running (use --profile gpu or mock)
# - Database not ready (wait 10s and restart: docker compose restart api)
# - Port already in use: lsof -i :7500
```

### Out of Disk Space

```bash
# Check disk usage
docker system df

# Prune unused images/volumes
docker system prune -a

# Remove specific volume
docker volume rm <volume-name>
```

### Port Already in Use

```bash
# Find process using port
lsof -i :7500  # API
lsof -i :3000  # Grafana
lsof -i :5432  # PostgreSQL

# Kill process (or change compose port binding)
kill <PID>

# Or edit compose.dev.yml to use different port:
# ports: - "127.0.0.1:7500:7500" → "127.0.0.1:8500:7500"
```

---

## Differences: Local vs Production

| Aspect | Local (compose.dev.yml) | Production (compose.prod.yml) |
|--------|-------------------------|-------------------------------|
| **Environment** | development | production |
| **Ports** | All on 127.0.0.1 (localhost) | Same (all local except Caddy) |
| **Resource Limits** | None (dev machine may vary) | Strict CPU/memory limits |
| **Logging** | json-file, unlimited | json-file, 10MB max, 3 files |
| **TLS/HTTPS** | Not configured (use localhost) | Automatic via Let's Encrypt |
| **Reverse Proxy** | Optional Caddy (commented) | Required Caddy (always on) |
| **Caddy Security Headers** | Not applied (localhost) | Full security headers (HSTS, CSP) |
| **Health Checks** | Same as prod | Same as dev |
| **Firewall** | Host firewall (if enabled) | Hetzner FW + ufw + Docker binding |
| **Restart Policy** | unless-stopped | unless-stopped |
| **Volumes** | Local (deleted on `down -v`) | Persistent (not deleted) |
| **LLM/Embeddings** | Optional (--profile gpu) | Required |
| **API/Web** | Built from local code | Pre-built from Docker images |

**Key Point:** Local setup is as close to production as possible, so bugs found locally will likely work in production.

---

## Performance Tuning

### Docker Desktop Settings (Mac/Windows)

If running on Docker Desktop:

```
Preferences → Resources:
- CPUs: 4-6 (adjust to your machine)
- Memory: 8-16 GB (LLM needs 6-8GB)
- Disk Image Size: 100+ GB
- Swap: 2GB
```

### Memory Issues

If services crash with OOM (Out Of Memory):

```yaml
# Edit compose.dev.yml, add to each service:
deploy:
  resources:
    limits:
      memory: 2G  # Adjust per service
    reservations:
      memory: 1G
```

### CPU Issues

If LLM is slow:

```yaml
# Edit compose.dev.yml llm service:
deploy:
  resources:
    limits:
      cpus: '4'
    reservations:
      cpus: '2'
```

### Disk Space

LLM model files are ~5GB:

```bash
du -sh ~/.cache/huggingface/hub/
```

---

## Database Persistence

### Keep Data Across Restarts

By default, volumes persist data:

```bash
docker compose -f compose.dev.yml down  # Data still there
docker compose -f compose.dev.yml up -d   # Data restored
```

### Reset Data

```bash
# Delete all volumes (WARNING: Deletes all data!)
docker compose -f compose.dev.yml down -v

# Start fresh
docker compose -f compose.dev.yml up -d
```

### Backup Database

```bash
# Backup API database
docker compose -f compose.dev.yml exec api-db pg_dump -U allora allora > backup-api.sql

# Backup Web database
docker compose -f compose.dev.yml exec web-db pg_dump -U allora allora_web > backup-web.sql

# Restore
docker compose -f compose.dev.yml exec api-db psql -U allora allora < backup-api.sql
```

---

## CI/CD Testing

### Run Tests Locally

```bash
# Type check
bun run typecheck

# Lint
bun run lint

# Format
bun run format

# Check (all in one)
bun run check

# Tests (if you have them)
bun run test
```

### Simulate GitHub Actions Build

```bash
# Build API image (as GitHub Actions does)
docker build -f infra/compose/Dockerfile.api -t test-api:latest .

# Build Web image
docker build -f infra/compose/Dockerfile.web -t test-web:latest .

# Tag for GHCR (if testing push)
docker tag test-api:latest ghcr.io/your-username/allora-api:latest
docker tag test-web:latest ghcr.io/your-username/allora-web:latest
```

---

## Next Steps After Setup

1. ✅ Start stack: `docker compose -f compose.dev.yml --profile gpu up -d`
2. ✅ Verify all services: `docker compose -f compose.dev.yml ps`
3. ✅ Test API: `curl http://localhost:7500/health`
4. ✅ Access Web: `open http://localhost:7501`
5. ✅ View metrics: `open http://localhost:9090`
6. ✅ Make code changes in `packages/api/` or `packages/web/`
7. ✅ Rebuild: `docker compose -f compose.dev.yml build`
8. ✅ Test: `curl` or browser

---

## Reference

- **Compose file:** `infra/compose/compose.dev.yml`
- **Environment:** `infra/compose/.env.dev`
- **Production file:** `infra/compose/compose.prod.yml`
- **Docker Compose docs:** https://docs.docker.com/compose/
- **Troubleshooting:** See section above
