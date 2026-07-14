# Allora Infrastructure

Complete infrastructure for deploying Allora to Hetzner dedicated server using Docker Compose, GitHub Actions, and Debian 13.

---

## Quick Start

**→ See [DEPLOY.md](DEPLOY.md)** for complete first-time and ongoing deployment instructions.

---

## Architecture

```
GitHub Repository (git push)
  ↓
GitHub Actions (build Docker images → push to GHCR)
  ↓
Production Server (Debian 13, Hetzner)
  ├── Docker Compose (9 services)
  │   ├── Ministral-3-3B LLM (GPU, port 7502)
  │   ├── Qwen3 Embeddings (GPU, port 7503)
  │   ├── PostgreSQL x2 (API + Web databases)
  │   ├── Bun API (port 7500)
  │   ├── Bun Web (port 7501)
  │   ├── Caddy (reverse proxy, TLS, ports 80/443)
  │   ├── Prometheus (metrics collection)
  │   └── Grafana (dashboards)
  │
  └── System Components
      ├── NVIDIA Driver + Container Toolkit (GPU support)
      ├── UFW firewall + Fail2ban (security)
      ├── Automatic security updates
      └── SSH key-only auth (no passwords)
```

---

## Services

### LLM Services (GPU)

**Ministral-3-3B** (Port 7502)
- Multimodal LLM (text + images)
- 2GB model (Q4 quantization)
- Full GPU offload
- JSON schema enforcement

**Qwen3-VL-Embedding** (Port 7503)
- Embedding model
- 1.8GB model
- 2048-dimensional embeddings
- Mean pooling

### Application Services

**API** (Port 7500)
- Bun HTTP server
- PostgreSQL backend (items, brands, queries)
- BetterAuth for API keys

**Web** (Port 7501)
- Bun HTTP server + React frontend
- PostgreSQL backend (users, sessions)
- BetterAuth for user sessions

### Infrastructure Services

**Caddy** (Ports 80, 443)
- Automatic HTTPS (Let's Encrypt)
- Reverse proxy for API/Web
- Security headers
- Rate limiting (optional)

**Prometheus** (Port 9090)
- Metrics collection
- 30-day retention
- Caddy metrics included

**Grafana** (Port 3000)
- Dashboard visualization
- Connected to Prometheus
- Admin access via SSH tunnel

---

## Files

```
infra/
├── deployment/
│   ├── bootstrap-secure.sh     # Security-hardened server setup
│   ├── bootstrap.sh            # Basic server setup
│   └── download-models.sh      # Download LLM models from Hugging Face
│
├── compose/
│   ├── compose.prod.yml        # Production orchestration
│   ├── compose.dev.yml         # Local development (mirrors prod)
│   ├── Dockerfile.api          # Multi-stage API build
│   ├── Dockerfile.web          # Multi-stage Web build
│   ├── .env.example            # Environment template
│   └── services/
│       ├── caddy/Caddyfile
│       ├── prometheus/prometheus.yml
│       └── postgres/backup.sh
│
├── opentofu/                   # Infrastructure provisioning (reference only)
│   ├── main.tf
│   ├── variables.tf
│   └── example.tfvars
│
├── DEPLOY.md                   # ⭐ Start here for deployment
├── LOCAL_DEVELOPMENT.md        # Local development setup
├── SECURITY.md                 # Security architecture & hardening
└── README.md                   # This file
```

---

## Deployment

### First Time (30-45 min)

1. Generate SSH keys (root + deploy)
2. Add root SSH key to server (one-time password)
3. Run bootstrap script (Docker, NVIDIA, security)
4. Copy Docker Compose files
5. Configure .env (domain, passwords)
6. Download LLM models (5-10 min)
7. Start services
8. Verify health (9 services healthy)
9. Optional: Set up GitHub Actions

**→ See [DEPLOY.md](DEPLOY.md) for step-by-step.**

### Ongoing

- View logs: `docker compose logs -f`
- Restart service: `docker compose restart <service>`
- Update: Push code → GitHub Actions auto-deploys
- Backup: Automated daily backups included

---

## Local Development

To test the full stack locally before deployment:

```bash
cd compose
docker compose -f compose.dev.yml --profile gpu up -d
```

See [LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md) for details.

---

## Security

All security measures are implemented:

✅ SSH key-only authentication (no passwords)
✅ UFW firewall + Fail2ban
✅ Automatic security updates
✅ Non-root service users
✅ Caddy automatic HTTPS with Let's Encrypt
✅ Security headers (HSTS, CSP, etc.)

See [SECURITY.md](SECURITY.md) for complete security documentation.

---

## Monitoring

**Grafana** (dashboards):
```bash
ssh -L 3000:localhost:3000 deploy@<SERVER_IP>
# Then: http://localhost:3000
```

**Prometheus** (metrics):
```bash
ssh -L 9090:localhost:9090 deploy@<SERVER_IP>
# Then: http://localhost:9090
```

**Caddy metrics**:
```bash
ssh deploy@<SERVER_IP> curl http://localhost:2019/metrics
```

---

## Troubleshooting

Services not starting?
```bash
docker compose -f compose.prod.yml logs
```

GPU not available?
```bash
docker run --rm --runtime=nvidia nvidia/cuda:12-runtime nvidia-smi
```

Can't SSH?
- Verify key exists: `ls -la ~/.ssh/allora-*`
- Check permissions: `chmod 600 ~/.ssh/allora-*`
- Verify server IP reachable: `ping <SERVER_IP>`

For complete troubleshooting, see [DEPLOY.md](DEPLOY.md).

---

## Cost

**Hetzner Dedicated Server (RTX 4000 Ada):** ~€2000/month
- Includes bandwidth, DDoS protection, dedicated resources
- 20GB GPU VRAM sufficient for both models + headroom

**Operating costs:**
- Backups: Automated locally (no external storage needed)
- Monitoring: Free (Prometheus + Grafana OSS)
- DNS: Free (Hetzner DNS)
- TLS: Free (Let's Encrypt via Caddy)

---

## Reference

- **Hetzner Cloud API:** https://docs.hetzner.cloud/
- **Docker Compose:** https://docs.docker.com/compose/
- **Caddy:** https://caddyserver.com/
- **Prometheus:** https://prometheus.io/
- **Grafana:** https://grafana.com/
- **Debian 13:** https://www.debian.org/

---

**Start with [DEPLOY.md](DEPLOY.md)** →
