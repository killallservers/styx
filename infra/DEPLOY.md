# Deploying Styx Registry to Hetzner cx23

Step-by-step guide for provisioning and deploying the Styx registry server.

---

## Prerequisites

- Hetzner account (sign up at https://hetzner.cloud)
- SSH keypair for deployment
- GitHub account with push access to styx repo
- Domain name (or use subdomain like registry.yourdomain.com)

---

## Step 1: Provision Hetzner cx23

### Create the Server in Hetzner Console

1. Go to https://console.hetzner.cloud
2. Create new project (or use existing)
3. Click **Add Server**
4. **Location:** Choose your region (Falkenstein recommended for EU)
5. **OS Image:** Debian 13 (latest)
6. **Server Type:** CX23 (2 vCPU, 4 GB RAM, 40 GB SSD) — ~€3/month
7. **SSH Keys:** Add your public key from `~/.ssh/id_rsa.pub`
8. **Name:** styx-registry
9. Click **Create Server**
10. **Copy the IP address** (you'll need it next)

### Run Provisioning Script

On your local machine:

```bash
cd styx
bash infra/deployment/provision-cx23.sh \
    <server_ip> \
    registry.styx.sh \
    killallservers/styx \
    ~/.ssh/id_rsa.pub
```

**What this does:**
- Creates `styx` user (non-root deployment)
- Installs Docker and Docker Compose
- Sets up UFW firewall (SSH, HTTP, HTTPS)
- Creates `/opt/styx` deployment directory
- Adds SSH key for automated CI/CD deployments
- Configures auto-security updates
- No configuration files needed (all via GitHub Actions secrets)

---

## Step 2: Configure DNS

Point your domain to the Hetzner server:

1. Go to your DNS provider (Cloudflare, Route53, etc.)
2. Create an **A record:**
   - Name: `registry` (or `registry.yourdomain.com`)
   - Type: A
   - Value: `<your-server-ip>`
   - TTL: 3600

Example for Cloudflare:
```
Type: A
Name: registry
IPv4: 1.2.3.4
TTL: Auto
```

Wait for DNS propagation (usually < 5 minutes, max 48 hours).

---

## Step 3: Initialize Repository on Server

SSH into the server:

```bash
ssh styx@<server_ip>
cd /opt/styx
```

Initialize git repository:

```bash
git init
git remote add origin https://github.com/killallservers/styx
git fetch origin main
git checkout -b main origin/main
```

---

## Step 4: Deploy Registry

Run the deploy script:

```bash
/home/styx/deploy.sh
```

This will:
1. Pull latest code
2. Build registry Docker image
3. Start registry + Caddy containers
4. Verify health check

Watch the logs:

```bash
docker compose -f /opt/styx/compose.prod.yml logs -f
```

---

## Step 5: Verify Deployment

From your local machine:

```bash
# Check health
curl https://registry.styx.sh/health

# Expected response:
# {"status":"ok","tools":26,"lastBuild":"2026-07-14T17:14:07Z"}

# Check registry.json
curl https://registry.styx.sh/registry.json | jq '.tools | length'

# Should output: 26
```

---

## Step 6: Configure GitHub Actions Deployments

See [GITHUB_SECRETS_SETUP.md](GITHUB_SECRETS_SETUP.md) for setting up automated deployments.

After secrets are configured, push to main to trigger auto-deployment:

```bash
git push origin main
```

---

## Step 7: Test Automated Deployment

Make a small change to trigger CI/CD:

```bash
echo "# Test deployment" >> README.md
git add README.md
git commit -m "test: trigger deployment"
git push origin main
```

Monitor GitHub Actions:
1. Go to https://github.com/killallservers/styx/actions
2. Watch the `deploy-registry` workflow
3. Should see: build → push → SSH deploy → restart

Verify the change is live:

```bash
curl https://registry.styx.sh/registry.json | jq '.metadata'
```

---

## Troubleshooting

### Domain not resolving

```bash
# Check DNS propagation
nslookup registry.styx.sh
dig registry.styx.sh

# Wait a few minutes if it's new
```

### Registry container failing

```bash
ssh styx@<server_ip>
docker compose -f /opt/styx/compose.prod.yml logs registry
```

Common issues:
- **Port 7506 in use:** Check `docker ps` and kill conflicting containers
- **TOML spec parsing error:** Check `/opt/styx/pkg/registry/specs/*.toml` are valid
- **Out of memory:** cx23 has 4GB; should be plenty for registry

### Caddy TLS not working

```bash
docker compose -f /opt/styx/compose.prod.yml logs caddy

# Caddy gets TLS from Let's Encrypt. If stuck:
# 1. Check domain is actually pointing to server
# 2. Check ports 80/443 are open in firewall
# 3. Restart Caddy: docker compose restart caddy
```

### Health check failing

```bash
# From server:
curl -v http://localhost:7506/health

# From local:
curl -v https://registry.styx.sh/health
```

Expected response:
```json
{
  "status": "ok",
  "tools": 26,
  "lastBuild": "2026-07-14T17:14:07Z"
}
```

---

## Maintenance

### Manual Updates

If GitHub Actions fails, manually deploy:

```bash
ssh styx@<server_ip>
cd /opt/styx
git pull origin main
/home/styx/deploy.sh
```

### View Logs

```bash
docker compose -f /opt/styx/compose.prod.yml logs -f registry
docker compose -f /opt/styx/compose.prod.yml logs -f caddy
```

### Stop Services

```bash
docker compose -f /opt/styx/compose.prod.yml down
```

### Restart Services

```bash
docker compose -f /opt/styx/compose.prod.yml up -d
```

### Rebuild Registry Image

```bash
docker compose -f /opt/styx/compose.prod.yml build --no-cache registry
docker compose -f /opt/styx/compose.prod.yml up -d registry
```

---

## Monitoring

### Health Checks

Add to your monitoring system (Uptime Kuma, Pingdom, etc.):
```
https://registry.styx.sh/health
```

### Logs

Caddy auto-rotates logs. View recent entries:

```bash
docker compose -f /opt/styx/compose.prod.yml logs --tail 100 caddy
docker compose -f /opt/styx/compose.prod.yml logs --tail 100 registry
```

### Disk Usage

Monitor `/opt/styx` disk usage:

```bash
du -sh /opt/styx
df -h /opt/styx
```

Registry should use < 100MB. If growing, check for:
- Build artifact accumulation
- Caddy certificate logs
- Container logs (auto-rotated via compose.prod.yml config)

---

## Cost & Performance

**Hetzner cx23 specs:**
- 2 vCPU (AMD EPYC)
- 4 GB RAM
- 40 GB NVMe SSD
- 1 Gbps network
- **€3/month base**
- **~€0.50/month bandwidth** (light usage)

**Expected performance:**
- Registry.json response: < 10ms (cached)
- 26 tools, ~2MB JSON
- ~100 requests/minute (healthy)
- Easily handles 1000+ requests/minute

**Upgrading:**
If you need more resources:
- CPX21 (2 vCPU, 8GB RAM): ~€6/month
- CPX31 (4 vCPU, 16GB RAM): ~€20/month

Upgrade in Hetzner console without downtime.

---

## Disaster Recovery

### Backup Registry Specs

The source of truth is GitHub. To restore:

```bash
# On server
rm -rf /opt/styx
git clone https://github.com/killallservers/styx /opt/styx
cd /opt/styx
/home/styx/deploy.sh
```

### Restore from Disaster

If the server dies, provision a new one with the same IP/domain:

```bash
bash infra/deployment/provision-cx23.sh <new_ip> registry.styx.sh killallservers/styx ~/.ssh/id_rsa.pub
```

DNS still points to old IP. Update Hetzner console or DNS provider to new IP.

---

## Next Steps

- [Set up GitHub Actions secrets](GITHUB_SECRETS_SETUP.md)
- [Local development testing](LOCAL_DEVELOPMENT.md)
- Monitor at: https://registry.styx.sh/health
