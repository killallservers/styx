# Allora Deployment Guide

**Complete guide for first deployment + ongoing operations.**

## Quick Start (TL;DR)

```bash
# 1. Generate SSH keys (locally first)
ssh-keygen -t ed25519 -f ~/.ssh/hetzner/<SERVER-ID>/root -C "erik.wright@killallservers.com"
ssh-keygen -t ed25519 -f ~/.ssh/hetzner/<SERVER-ID>/deploy -C "erik.wright@killallservers.com"

# 2. Add root key using password (from Hetzner email — one-time only)
ssh-copy-id -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP>

# Enter password when prompted, adds default SSH key

# 3. Add GitHub secrets (one-time setup)
# Follow: infra/GITHUB_SECRETS_SETUP.md (complete step-by-step guide)
# Secrets needed: PRODUCTION_SERVER_IP, PRODUCTION_SSH_KEY, PRODUCTION_DOMAIN, DEPLOY_SSH_PUBLIC_KEY

# 4. Run GitHub Actions workflow to initialize server
# Go to: GitHub → Actions → "Initialize Server"
# Click "Run workflow" (no inputs needed) → wait 30-45 minutes

# 5-7. Follow steps below (configure, verify, deploy)
# Total: 30-45 minutes bootstrap + 5-10 minutes manual setup
```

---

## Prerequisites

- Hetzner Debian 13 server provisioned (IP, SSH access)
- Local machine with `ssh`, `scp`, Git
- Two SSH keys (recommend ed25519):
  - `~/.ssh/hetzner/<SERVER-ID>/root` — Emergency root access
  - `~/.ssh/hetzner/<SERVER-ID>/deploy` — Normal operations
- GitHub repository (optional for CI/CD automatic deployments)

---

## Part 1: First-Time Deployment (30-45 min)

### ⚠️ CRITICAL PREREQUISITE: Add Root SSH Key to Server

**THIS MUST BE DONE BEFORE RUNNING THE GITHUB ACTIONS WORKFLOW**

The GitHub Actions workflow will verify your root SSH key is authorized. If it's not, the workflow will fail immediately.

**One-time setup (local machine):**

```bash
# 1. Generate two separate SSH keys
ssh-keygen -t ed25519 -f ~/.ssh/hetzner/<SERVER-ID>/root -C "admin@allora.style"
ssh-keygen -t ed25519 -f ~/.ssh/hetzner/<SERVER-ID>/deploy -C "admin@allora.style"

# 2. Add root key to server using Hetzner password (one-time only)
ssh-copy-id -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP>
# When prompted, enter the password from your Hetzner email

# 3. Verify the key works
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> "echo OK"
# Expected output: OK
```

**Why both keys?**
- **Root key** (`root`) — Used by GitHub Actions for initial bootstrap setup
- **Deploy key** (`deploy`) — Used by developers/deploy scripts for normal operations

**After ssh-copy-id succeeds:**
- Password authentication is permanently disabled by the bootstrap script
- Only SSH key authentication will work

### Step 0: Prepare for GitHub Actions Workflow

Ensure you have:
- ✅ SSH keys generated locally (root + deploy)
- ✅ Root key added to server via `ssh-copy-id` (password auth used once)
- ✅ Root key verified working: `ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> "echo OK"`
- ✅ GitHub secrets configured (see Step 1 Prerequisites)

### Step 1: Initialize Server with GitHub Actions (30-45 min)

Instead of manually running bootstrap, use the GitHub Actions workflow:

**Prerequisites:**
1. **Add GitHub secrets** — Follow `infra/GITHUB_SECRETS_SETUP.md` for complete step-by-step instructions
   - `PRODUCTION_SERVER_IP` — Your server IP
   - `PRODUCTION_SSH_KEY` — Your root private key
   - `DEPLOY_SSH_PUBLIC_KEY` — Your deploy public key
   - `PRODUCTION_DOMAIN` — Your domain (e.g., `allora.example.com`)
   - `SLACK_WEBHOOK` — (Optional) Slack notification webhook

**Run the workflow:**

1. Go to GitHub → Actions → "Initialize Server"
2. Click "Run workflow" button (green)
3. Click "Run workflow" (no inputs needed — uses secrets from GitHub)
4. Monitor at GitHub → Actions (takes 30-45 minutes)

The workflow automatically:
- ✅ Transfers `bootstrap-secure.sh` to server
- ✅ Runs bootstrap with proper environment variables
- ✅ Installs Docker + Docker Compose
- ✅ Installs NVIDIA Container Toolkit (GPU support)
- ✅ Configures UFW firewall + Fail2ban
- ✅ Hardens SSH (key-only auth, password disabled)
- ✅ Sets up automatic security updates
- ✅ Creates deploy user with Docker + sudo access
- ✅ Adds SSH keys to both root and deploy users
- ✅ Sends Slack notification when complete

**Status updates:**
Watch the workflow log for progress. At the end, you'll see:
```
✅ ✅ ✅ SERVER INITIALIZATION COMPLETE ✅ ✅ ✅
```

The workflow output includes all next steps and verification commands.

### Step 2: Verify Bootstrap Completion (5 min)

After the GitHub Actions workflow completes, verify the server state:

Check Docker + GPU:

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> "docker --version && nvidia-smi"
```

Check deploy user:

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> "whoami && sudo docker ps"
```

Check SSH hardening (PASSWORD AUTH MUST BE DISABLED):

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> \
  "grep '^PasswordAuthentication' /etc/ssh/sshd_config"
```

**Expected output:** `PasswordAuthentication no`

Check firewall rules:

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> "ufw status && echo '---' && fail2ban-client status sshd"
```

**Expected:** UFW active, SSH allowed, Fail2ban protecting sshd

**If bootstrap failed:** Re-run the GitHub Actions workflow with the same server IP

⚠️  **If hardening didn't apply (still see password prompts):**

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> bash -c '
DEPLOY_SSH_PUBLIC_KEY="'"$(cat ~/.ssh/hetzner/<SERVER-ID>/deploy.pub)"'"
bash /tmp/bootstrap-secure.sh
'
```

### Step 3: Copy Docker Compose Files (1 min)

```bash
scp -i ~/.ssh/hetzner/<SERVER-ID>/deploy -r infra/compose/* deploy@<SERVER_IP>:/opt/allora/compose/
scp -i ~/.ssh/hetzner/<SERVER-ID>/deploy infra/deployment/*.sh deploy@<SERVER_IP>:/opt/allora/deployment/
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> "sudo chown -R deploy:docker /opt/allora"
```

### Step 4: Configure Environment (2 min)

SSH to server:

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP>
```

Edit environment:

```bash
nano /opt/allora/compose/.env
```

Set these values:

```env
ENVIRONMENT=production
DOMAIN=your-domain.com
DOCKER_REGISTRY=ghcr.io/killallservers
IMAGE_TAG=latest

POSTGRES_PASSWORD=<32-char-random>     # Generate: openssl rand -base64 32
GRAFANA_ADMIN_PASSWORD=<16-char-random> # Generate: openssl rand -base64 16
```

**Important:**
- `DOMAIN`: **Domain name only** (e.g., `allora.example.com`), no `https://` or `http://` prefix
- `DOCKER_REGISTRY`: Automatically set by bootstrap. Only edit if using a different registry

Save and exit: `Ctrl+O`, `Enter`, `Ctrl+X`

### Step 5: Download Models (5-10 min)

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> bash /opt/allora/deployment/download-models.sh
```

Downloads directly to `~/.cache/huggingface/hub/`:
- Ministral-3-3B-Instruct (2.5GB)
- Qwen3-VL-Embedding-2B (1.8GB)

Takes 5-10 minutes depending on network speed.

### Step 6: Start Services (5 min)

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "cd /opt/allora/compose && docker compose -f compose.prod.yml up -d"
```

Wait for services to start (2-3 min), then check:

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "cd /opt/allora/compose && docker compose -f compose.prod.yml ps"
```

Expected: All 9 services show `Up (healthy)`:
- api-db, web-db (PostgreSQL)
- llm, embeddings (LLM services)
- api, web (Application services)
- caddy (Reverse proxy)
- prometheus, grafana (Monitoring)

### Step 7: Verify Services (5 min)

Health checks:

```bash
# API health
curl https://<SERVER_IP>/api/health

# Web (in browser)
https://<SERVER_IP>/

# Grafana (via SSH tunnel)
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy -L 3000:localhost:3000 deploy@<SERVER_IP>
# Then: http://localhost:3000 (user: admin, password from .env)
```

### Step 7: Enable Automatic Deployments (1 min)

The GitHub Actions secrets were already added in Step 1. Now automatic deployments are active.

**Test deployment:**

```bash
git add .
git commit -m "test: trigger deployment"
git push origin main
```

GitHub Actions will:
1. Build new Docker images
2. Push to GHCR
3. Automatically deploy to production via SSH
4. Send Slack notification

Monitor at: GitHub → Actions → "Deploy to Production"

**Manual deployment (if needed):**

Go to GitHub → Actions → "Deploy to Production" → "Run workflow" → enter image tag

---

## Bootstrap Script Details

For complete details on bootstrap script features, environment variables, and troubleshooting, see:
- **`infra/deployment/bootstrap-secure.sh`** — Full source code with extensive comments
- **`GITHUB_SECRETS_SETUP.md`** — Secrets configuration guide (includes environment variable details)

---

## Part 2: Ongoing Operations

### View Logs

All services:
```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml logs -f"
```

Specific service:
```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml logs -f api"
```

### Restart a Service

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml restart api"
```

### Stop All Services

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml down"
```

### Start Services Again

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml up -d"
```

### Database Backup

Manual backup:
```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker exec allora-api-db pg_dump -U allora allora > ~/backup-$(date +%Y%m%d).sql"
```

Automated backups already configured (see `infra/compose/services/postgres/backup.sh`).

### Update Services

Pull latest images:
```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml pull"
```

Restart with new images:
```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml up -d"
```

Or, push code and GitHub Actions handles it automatically.

### Configure Environment Variables

Edit on server:
```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP>
nano /opt/allora/compose/.env
```

Restart affected services:
```bash
docker compose -f /opt/allora/compose/compose.prod.yml up -d
```

---

## User Access

| User | SSH Key | Access | Use Case |
|------|---------|--------|----------|
| `root` | `~/.ssh/allora-root` | All files, containers, system | Emergencies only |
| `deploy` | `~/.ssh/allora-deploy` | Docker, sudo, compose files | Normal operations |

### SSH Config (Optional)

Add to `~/.ssh/config`:

```
Host allora
  HostName <SERVER_IP>
  User deploy
  IdentityFile ~/.ssh/allora-deploy

Host allora-root
  HostName <SERVER_IP>
  User root
  IdentityFile ~/.ssh/allora-root
```

Then: `ssh allora` (deploy), `ssh allora-root` (emergency)

---

## Troubleshooting

### Services won't start

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml logs"
```

Common causes:
- Waiting for database (wait 30s, retry)
- Missing models (re-run download-models.sh)
- Port conflict (change in compose.prod.yml)

### GPU not available

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker run --rm --runtime=nvidia nvidia/cuda:12-runtime nvidia-smi"
```

If fails, re-run bootstrap (root key):

```bash
ssh -i ~/.ssh/allora-root root@<SERVER_IP> bash /tmp/bootstrap-secure.sh
```

### Password Authentication Still Enabled

**Problem:** You can still SSH with password even though bootstrap ran

**Root Cause:** Conflicting SSH config entries, or bootstrap script didn't complete properly

**Fix:**

```bash
# 1. Check current SSH config
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> \
  "grep -n 'PasswordAuthentication' /etc/ssh/sshd_config"

# If multiple lines exist, only the LAST one takes effect
# Output should show only: PasswordAuthentication no

# 2. Remove all PasswordAuthentication lines and re-run bootstrap
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> \
  "sed -i '/^#*PasswordAuthentication/d' /etc/ssh/sshd_config && \
   echo 'PasswordAuthentication no' >> /etc/ssh/sshd_config && \
   sshd -t && systemctl restart sshd && echo 'SSH hardened successfully'"

# 3. Re-run bootstrap to ensure all settings are applied
DEPLOY_SSH_PUBLIC_KEY="$(cat ~/.ssh/hetzner/<SERVER-ID>/deploy.pub)" \
  ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> bash /tmp/bootstrap-secure.sh

# 4. Verify: this should print "PasswordAuthentication no"
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> \
  "grep '^PasswordAuthentication' /etc/ssh/sshd_config"

# 5. Test: this MUST fail with "Permission denied (publickey)" not password prompt
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@<SERVER_IP> -o PreferredAuthentications=password -o PubkeyAuthentication=no "echo test" 2>&1
```

**Expected:** Permission denied or similar, NOT a password prompt

### Can't SSH

Verify key exists and permissions:

```bash
ls -la ~/.ssh/hetzner/<SERVER-ID>/{root,deploy}
chmod 600 ~/.ssh/hetzner/<SERVER-ID>/{root,deploy}
chmod 644 ~/.ssh/hetzner/<SERVER-ID>/{root,deploy}.pub
```

Verify server IP:

```bash
ping <SERVER_IP>
```

Test key authentication explicitly:

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/root -o PubkeyAuthentication=yes -o PasswordAuthentication=no root@<SERVER_IP> "echo OK"
```

If that fails but password works, the issue is SSH hardening. See "Password Authentication Still Enabled" above.

### Caddy SSL issues

Check certificate status:

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml exec caddy caddy list-certs"
```

Check logs:

```bash
ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@<SERVER_IP> \
  "docker compose -f /opt/allora/compose/compose.prod.yml logs caddy"
```

Verify DNS points to server:

```bash
nslookup <DOMAIN>
# Should return <SERVER_IP>
```

---

## Security Notes

✅ **Implemented:**
- SSH key-only auth (no passwords)
- Two separate SSH keys (root + deploy)
- UFW firewall + Fail2ban
- Automatic security updates
- Non-root service users (Docker appuser:1000)
- Caddy automatic HTTPS

⚠️ **Recommended:**
- Rotate SSH keys periodically
- Store private keys securely (not in repo, not on web servers)
- Monitor logs for suspicious activity
- Keep Debian packages updated
- Use strong database passwords (32+ chars)

---

## Reference

- **Compose files:** `infra/compose/compose.prod.yml` (production), `compose.dev.yml` (development)
- **Bootstrap:** `infra/deployment/bootstrap-secure.sh` (idempotent, safe to re-run)
- **Models:** `infra/deployment/download-models.sh` (downloads from Hugging Face)
- **Local dev:** `infra/LOCAL_DEVELOPMENT.md`
- **Security:** `infra/SECURITY.md`
