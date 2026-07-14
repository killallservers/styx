# Styx Registry Infrastructure

Production infrastructure for deploying the Styx registry server to Hetzner.

**Architecture:**
- Styx registry server (Go binary, port 7506)
- Caddy reverse proxy (TLS via Let's Encrypt, ports 80/443)
- Automatic deployments via GitHub Actions on push to main

---

## Quick Start

1. **Provision a Hetzner cx23 server:**
   ```bash
   bash infra/deployment/provision-cx23.sh <public_ip> <domain>
   ```

2. **Configure GitHub Actions secrets:**
   See [GITHUB_SECRETS_SETUP.md](GITHUB_SECRETS_SETUP.md)

3. **Deploy:**
   ```bash
   git push origin main
   ```
   GitHub Actions automatically builds, pushes, and deploys.

---

## Files

```
infra/
├── compose/
│   ├── compose.prod.yml        # Production services (registry + caddy)
│   ├── compose.dev.yml         # Local development
│   ├── Dockerfile.registry     # Registry server image
│   ├── Caddyfile              # Reverse proxy config
│   └── .env.example           # Environment variables template
│
├── deployment/
│   ├── provision-cx23.sh      # Hetzner cx23 server setup
│   ├── bootstrap.sh           # Docker & Docker Compose install
│   └── deploy.sh              # SSH deployment from CI/CD
│
├── README.md                   # This file
├── DEPLOY.md                  # Detailed deployment guide
├── LOCAL_DEVELOPMENT.md        # Local testing
└── GITHUB_SECRETS_SETUP.md    # GitHub Actions configuration
```

---

## Deployment Options

### Option 1: Hetzner cx23 (Recommended)
- Minimal 2-core VM (~€3/month)
- Automatic provisioning script included
- Full TLS support via Let's Encrypt
- See [DEPLOY.md](DEPLOY.md)

### Option 2: Other Linux Server
- Same Docker Compose setup works anywhere
- Requires manual provisioning
- Follow [DEPLOY.md](DEPLOY.md) but adapt to your infrastructure

---

## Services

### Styx Registry (Port 7506, internal)
- HTTP API serving registry.json
- GET `/registry.json` — 26 tools, all versions
- GET `/health` — Liveness check
- POST `/webhook/github` — Rebuild on spec changes

### Caddy (Ports 80, 443)
- Automatic HTTPS (Let's Encrypt)
- Reverse proxy to registry (7506 → 443/registry.json)
- Security headers
- Certificate auto-renewal

---

## Monitoring & Troubleshooting

**Check status:**
```bash
curl https://registry.styx.sh/health
```

**View logs (on server):**
```bash
docker compose -f /opt/styx/compose.prod.yml logs -f registry
docker compose -f /opt/styx/compose.prod.yml logs -f caddy
```

**Restart services:**
```bash
docker compose -f /opt/styx/compose.prod.yml restart
```

**Manual rebuild (if webhook fails):**
```bash
cd /opt/styx
docker compose -f compose.prod.yml build --no-cache registry
docker compose -f compose.prod.yml up -d
```

---

## Security

- SSH key-only authentication (no password)
- UFW firewall (22, 80, 443 only)
- Automatic security updates
- TLS certificates auto-renewed
- Environment variables in `.env` (never committed)

---

## Cost

Hetzner cx23:
- **€3/month** base server
- **~€0.50/month** bandwidth (assuming light usage)
- **Total: ~€3.50/month**

---

## Next Steps

1. Read [DEPLOY.md](DEPLOY.md) for step-by-step instructions
2. Set up [GitHub Actions secrets](GITHUB_SECRETS_SETUP.md)
3. Test locally with [LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md)
4. Deploy to Hetzner
