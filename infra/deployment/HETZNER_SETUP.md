# Quick Hetzner Setup with cloud-init

Automatically provision a cx23 server with one command.

---

## Quick Start (Fastest)

### 1. Edit the userdata script

```bash
# Edit your SSH public key into the script
nano infra/deployment/hetzner-userdata.sh

# Find this line and replace:
SSH_PUBKEY="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... your-public-key-here"

# With your actual key:
SSH_PUBKEY="$(cat ~/.ssh/id_rsa.pub)"
```

### 2. Create Hetzner server with userdata

**Via Hetzner console UI:**
1. Go to https://console.hetzner.cloud
2. Click "Add Server"
3. Choose **CX23**, **Debian 13**
4. Scroll to "User data" section
5. Copy/paste contents of `infra/deployment/hetzner-userdata.sh`
6. Create server
7. **Wait 2-3 minutes** for initialization to complete

**Via hcloud CLI:**
```bash
# Install: https://github.com/hetznercloud/cli

hcloud server create \
  --name styx-registry \
  --type cx23 \
  --image debian-13 \
  --ssh-key <your-ssh-key-name> \
  --user-data-from-file infra/deployment/hetzner-userdata.sh
```

### 3. Get server IP and configure DNS

```bash
# From Hetzner console, copy the IP address
IP="1.2.3.4"

# Point domain to server
# A record: registry.styx.sh → 1.2.3.4
```

### 4. SSH and deploy

```bash
ssh styx@1.2.3.4

# Inside server:
cd /opt/styx
./deploy.sh
```

### 5. Add GitHub Actions secrets

```
STYX_DEPLOY_HOST       = 1.2.3.4
STYX_DEPLOY_USER       = styx
STYX_DEPLOY_KEY        = (your SSH private key)
STYX_REGISTRY_DOMAIN   = registry.styx.sh
```

### 6. Test

```bash
curl https://registry.styx.sh/health
```

---

## What Happens Automatically

**On server boot (runs as root):**
- ✅ apt updates
- ✅ Docker + Docker Compose install
- ✅ UFW firewall (22, 80, 443)
- ✅ Create styx user
- ✅ Create /opt/styx directory
- ✅ Add SSH key
- ✅ Auto-security updates via cron

**As styx user:**
- ✅ Clone repository
- ✅ Create deploy.sh
- ✅ Add shell aliases (styx-deploy, styx-logs, styx-status)

**Total time:** 2-3 minutes

---

## Manual Alternative (No cloud-init)

If you prefer to not use cloud-init:

```bash
# 1. Create server without user data (just base Debian 13)

# 2. SSH in and run provisioning
ssh root@<server-ip>
bash -c "$(curl -fsSL https://raw.githubusercontent.com/killallservers/styx/main/infra/deployment/provision-cx23.sh)"

# 3. Then as styx user:
ssh styx@<server-ip>
bash -c "$(curl -fsSL https://raw.githubusercontent.com/killallservers/styx/main/infra/deployment/init-styx-user.sh)" killallservers/styx
```

---

## Troubleshooting

### Server still initializing

```bash
# Check cloud-init logs
ssh root@<server-ip> tail -f /var/log/cloud-init-output.log
```

### SSH key not working

```bash
# Verify key was added
ssh root@<server-ip> cat /home/styx/.ssh/authorized_keys
```

### Docker not found

```bash
ssh styx@<server-ip> docker ps
# If fails, wait another minute or run:
source ~/.bashrc
```

### Repository not cloned

```bash
ssh styx@<server-ip> ls -la /opt/styx
# Should see .git directory
```

---

## Security Notes

- ✅ SSH key-only auth (no passwords)
- ✅ UFW firewall enabled
- ✅ Auto-security updates daily
- ✅ styx user is non-root
- ✅ Docker daemon requires authentication
- ⚠️ SSH_PUBKEY in userdata is readable in Hetzner logs (use key-only auth)

---

## Next Steps

- [GitHub Actions Secrets Setup](../GITHUB_SECRETS_SETUP.md)
- [Deployment Guide](../DEPLOY.md)
- [Local Development](../LOCAL_DEVELOPMENT.md)
