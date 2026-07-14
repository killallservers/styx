# Allora Security Hardening Guide

## Overview

This document details the security architecture for Allora deployed on Hetzner dedicated server, addressing common concerns including Docker + firewall compatibility.

## Key Security Decisions

### Docker + UFW Incompatibility: SOLVED ✅

**The Problem:**
Docker routes container traffic via iptables NAT rules, which bypass UFW's INPUT/OUTPUT chains. This means UFW firewall rules don't protect against access to Docker containers' exposed ports—they can still be reached even if UFW denies them.

**Our Solution:**
We use a **two-tier firewall approach**:

```
┌─────────────────────────────────────────┐
│      Hetzner Cloud Firewall             │
│  (External traffic filtering - Layer 1) │
│  ✓ SSH (22)                             │
│  ✓ HTTP (80)                            │
│  ✓ HTTPS (443)                          │
│  ✓ GPU model ports (7502, 7503) - DENY │
│  ✗ Internal ports - DENY                │
└────────────┬────────────────────────────┘
             │
        Ubuntu Host
        ┌────┴─────────────────────┐
        │   ufw (Local protection)  │
        │  (Internal network Layer) │
        │  ✓ SSH (22)               │
        │  ✓ HTTP (80)              │
        │  ✓ HTTPS (443)            │
        │  ✗ Everything else DENY   │
        └────┬──────────────────────┘
             │
    ┌────────┴──────────────────┐
    │    Docker Services        │
    │                           │
    │  - Ministral (7502)       │
    │    → 127.0.0.1:7502       │
    │    (Not exposed to network│
    │    unless via Caddy/443)  │
    │                           │
    │  - Qwen3 (7503)           │
    │    → 127.0.0.1:7503       │
    │    (Not exposed to network)
    │                           │
    │  - PostgreSQL (5432-5433) │
    │    → 127.0.0.1            │
    │    (Localhost only)       │
    │                           │
    │  - Caddy (80, 443)        │
    │    → 0.0.0.0:80/443       │
    │    (Public-facing only)   │
    └───────────────────────────┘
```

**Why this works:**

1. **Hetzner Cloud Firewall** (network layer):
   - Configured in OpenTofu `main.tf`
   - Blocks all traffic to internal ports from external networks
   - Only allows 22 (SSH), 80 (HTTP), 443 (HTTPS) from internet
   - Effectively prevents direct access to Docker service ports

2. **ufw** (host layer):
   - Prevents local processes from accidentally binding to unexpected ports
   - Protects against misconfigured services on the host
   - Doesn't conflict with Docker because we only expose necessary ports

3. **Docker binding** (container layer):
   - ALL internal services (Ministral, Qwen3, PostgreSQL, Prometheus, Grafana) bind to 127.0.0.1
   - Only Caddy binds to 0.0.0.0 (all interfaces)
   - Traffic to internal services is impossible from outside the host

**Result:** Defense in depth. Multiple layers ensure even if one fails, others protect.

---

## Security Architecture

### 1. Network Layer (Hetzner Cloud Firewall)

**Configured in:** `infra/opentofu/main.tf`

```hcl
# Allow SSH
- Direction: in, Protocol: tcp, Port: 22, Source: 0.0.0.0/0, ::/0

# Allow HTTP/HTTPS
- Direction: in, Protocol: tcp, Port: 80, Source: 0.0.0.0/0, ::/0
- Direction: in, Protocol: tcp, Port: 443, Source: 0.0.0.0/0, ::/0

# Internal ports (VPC only)
- Direction: in, Protocol: tcp, Port: 5432, Source: 10.0.0.0/8
- Direction: in, Protocol: tcp, Port: 7502-7503, Source: 10.0.0.0/8
- Direction: in, Protocol: tcp, Port: 9090, Source: 10.0.0.0/8
```

**Recommendations:**
- Restrict SSH (22) to your IP if known:
  ```hcl
  source_ips = ["YOUR.IP.ADDRESS/32"]
  ```
- Internal ports (5432, 7502, 7503, 9090) automatically restricted to VPC
- HTTP/S (80, 443) must remain open for public access

### 2. Host Layer (ufw Firewall)

**Configured in:** `bootstrap-secure.sh`

```bash
# Default policy: Deny incoming, allow outgoing
ufw default deny incoming
ufw default allow outgoing

# Whitelist specific services
ufw allow 22/tcp  # SSH
ufw allow 80/tcp  # HTTP
ufw allow 443/tcp # HTTPS
```

**Rationale:**
- Any unexpected service binding to the host will be blocked
- Prevents accidental exposure of development tools, debug ports, etc.
- Doesn't conflict with Docker (Docker uses iptables directly, not ufw rules)

### 3. Container Layer (Docker Port Binding)

**Configured in:** `compose.prod.yml`

Every internal service binds to `127.0.0.1` (localhost only):

```yaml
api:
  ports:
    - "127.0.0.1:7500:7500"  # Local only

llm:
  ports:
    - "127.0.0.1:7502:8000"  # Local only

embeddings:
  ports:
    - "127.0.0.1:7503:8000"  # Local only

prometheus:
  ports:
    - "127.0.0.1:9090:9090"  # Local only

grafana:
  ports:
    - "127.0.0.1:3000:3000"  # Local only
```

Only Caddy is public-facing:

```yaml
caddy:
  ports:
    - "80:80"          # World-facing
    - "443:443"        # World-facing
```

**Traffic flow:**
- External request to `https://api.example.com` → Hetzner FW (allow) → ufw (allow 443) → Caddy (0.0.0.0:443) → reverse proxy to api (127.0.0.1:7500)
- Direct request to `http://server:7500` → Hetzner FW (block) → Never reaches server

---

## SSH Hardening

**Configured in:** `bootstrap-secure.sh`

```bash
# Key settings in /etc/ssh/sshd_config
PermitRootLogin prohibit-password        # Key-only auth
PasswordAuthentication no                 # Disable passwords
PubkeyAuthentication yes                  # Enable key auth
X11Forwarding no                          # Disable X11
MaxAuthTries 3                            # Fail2ban works better with strict SSH
MaxSessions 5                             # Limit concurrent sessions
ClientAliveInterval 300                   # Disconnect idle clients
StrictModes yes                           # Enforce strict file permissions
UseDNS no                                 # Speed up login (no reverse DNS lookup)
LogLevel VERBOSE                          # Detailed logging for Fail2ban
```

**Verification:**

The bootstrap script now verifies SSH hardening is actually applied:

```bash
# Verify password auth is disabled (should print exactly this)
grep "^PasswordAuthentication no" /etc/ssh/sshd_config
# Expected: PasswordAuthentication no

# Verify public key auth is enabled
grep "^PubkeyAuthentication yes" /etc/ssh/sshd_config
# Expected: PubkeyAuthentication yes

# Verify SSH config is valid
sshd -t
# Expected: exit 0 (no output = success)
```

**⚠️ Important:** If you can still login with password after bootstrap, SSH hardening failed. Run bootstrap again or manually fix:

```bash
# Remove all PasswordAuthentication lines
sed -i '/^#*PasswordAuthentication/d' /etc/ssh/sshd_config

# Add the correct setting
echo "PasswordAuthentication no" >> /etc/ssh/sshd_config

# Verify and restart
sshd -t && systemctl restart sshd
```

**Manual additional hardening:**

```bash
# Change SSH port (optional, requires firewall update)
# ⚠️ Only if you update Hetzner Firewall to allow new port
# vim /etc/ssh/sshd_config → Port 2222
# systemctl restart sshd
# ufw allow 2222/tcp

# Add authorized keys only (no passwords)
# Copy your public key to /root/.ssh/authorized_keys
# chmod 600 /root/.ssh/authorized_keys
```

---

## Fail2ban (Brute-Force Protection)

**Configured in:** `bootstrap-secure.sh`

```bash
# Configuration in /etc/fail2ban/jail.local
bantime = 3600              # Ban for 1 hour
findtime = 600              # Look back 10 minutes
maxretry = 3                # Ban after 3 failed attempts
```

**Monitor Fail2ban:**

```bash
# Check status
fail2ban-client status

# Check SSH jail specifically
fail2ban-client status sshd

# View banned IPs
fail2ban-client set sshd unbanip <ip>  # Unban if needed
```

**Logs:**

```bash
# View Fail2ban logs
tail -f /var/log/fail2ban.log

# See failed SSH attempts
grep "Failed password" /var/log/auth.log
```

---

## Automatic Security Updates

**Configured in:** `bootstrap-secure.sh`

Ubuntu's `unattended-upgrades` automatically applies security patches:

```bash
# Configuration in /etc/apt/apt.conf.d/50unattended-upgrades
# Only security updates, not major version upgrades
# Automatic reboot disabled (you control reboot timing)
# Runs daily
```

**Monitor updates:**

```bash
# Check update log
tail -f /var/log/unattended-upgrades/unattended-upgrades.log

# Check what was updated
apt-get install needrestart  # Shows if reboot needed
needrestart -r l             # List what needs restart
```

---

## Docker Security

### Image Security

```bash
# Use only official, trusted images
- ghcr.io/ggml-org/llama.cpp:server-cuda13  ✓ Official GGML project
- postgres:18-alpine                        ✓ Official
- caddy:2-alpine                            ✓ Official
- prom/prometheus:latest                    ✓ Official
- grafana/grafana:latest                    ✓ Official

# Your images from GitHub
- ghcr.io/your-username/allora-api         ✓ Built from your Dockerfile
- ghcr.io/your-username/allora-web         ✓ Built from your Dockerfile
```

### Runtime Security

```yaml
# compose.prod.yml security settings
deploy:
  resources:
    limits:
      cpus: '2'
      memory: 4G
    reservations:
      cpus: '1'
      memory: 2G

# Non-root user (in Dockerfile)
USER appuser:1000  # ✓ Containers don't run as root

# Resource isolation
- CPU limits prevent runaway processes
- Memory limits prevent OOM crashes
- Logging driver limits disk usage
```

### Docker Daemon Security

```json
{
  "log-driver": "json-file",
  "log-opts": {"max-size": "10m", "max-file": "3"},
  "storage-driver": "overlay2",
  "icc": false,           // ✓ Disable inter-container communication
  "userland-proxy": false // ✓ Use kernel proxying (faster, more secure)
}
```

---

## Data Protection

### Database Passwords

```bash
# Generated securely
POSTGRES_PASSWORD=$(openssl rand -base64 32)

# Stored safely in /opt/allora/compose/.env
chmod 600 /opt/allora/compose/.env

# Never committed to git (.env in .gitignore)
```

### Backups

```bash
# Daily automated backup at 2 AM
0 2 * * * docker exec allora-api-db /usr/local/bin/backup.sh

# Retention: 7 days (older backups auto-deleted)
# Location: /var/lib/postgresql/backups/ (inside container)

# Restore (if needed)
docker exec allora-api-db psql -U allora < backup.sql
```

### GHCR Registry Authentication

```bash
# Stored in Docker config
/root/.docker/config.json (chmod 600)

# Uses GitHub token (scoped to write:packages)
# Not committed to git

# Token can be revoked anytime from GitHub settings
```

---

## TLS/HTTPS Security

### Automatic HTTPS via Let's Encrypt

**Caddy handles this:**

```
1. Server starts
2. Caddy tries to renew existing certificate
3. If none exists, requests new cert from Let's Encrypt
4. Certificate auto-renews 30 days before expiration
5. HTTPS enabled by default
```

**Verify:**

```bash
# Check certificate
docker compose -f compose.prod.yml exec caddy caddy list-certs

# Manual renew (if needed)
docker compose -f compose.prod.yml exec caddy caddy reload
```

### Security Headers

**Configured in:** `Caddyfile`

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

These prevent:
- SSL stripping attacks
- MIME type confusion
- Clickjacking
- XSS attacks

---

## Monitoring & Logging

### Service Logs

```bash
# All services
docker compose -f compose.prod.yml logs -f

# Specific service
docker compose -f compose.prod.yml logs -f api

# Last 100 lines, follow
docker compose -f compose.prod.yml logs --tail 100 -f nginx
```

### System Logs

```bash
# SSH attempts
tail -f /var/log/auth.log

# Fail2ban
tail -f /var/log/fail2ban.log

# Docker daemon
journalctl -u docker -f

# System messages
tail -f /var/log/syslog
```

### Metrics (Prometheus)

```bash
# Access via SSH tunnel
ssh -L 9090:localhost:9090 root@<server-ip>

# Then: http://localhost:9090 in browser
```

---

## Security Checklist

### After Bootstrap

- [ ] SSH into server successfully
- [ ] Verify Fail2ban is running: `fail2ban-client status`
- [ ] Verify ufw is enabled: `ufw status`
- [ ] Verify Docker is running: `docker ps`
- [ ] Verify NVIDIA GPU: `docker run --rm --runtime=nvidia nvidia/cuda:12-runtime nvidia-smi`
- [ ] Check bootstrap log: `cat /var/log/allora/bootstrap.log`

### Before Running Services

- [ ] Edit `.env` with actual DOMAIN
- [ ] Set strong POSTGRES_PASSWORD (32+ chars)
- [ ] Set strong GRAFANA_ADMIN_PASSWORD (16+ chars)
- [ ] Review Hetzner Firewall rules
- [ ] Review SSH config: `cat /etc/ssh/sshd_config | grep -v "^#"`

### After Starting Services

- [ ] All services healthy: `docker compose -f compose.prod.yml ps`
- [ ] External access works: `curl https://your-domain/`
- [ ] HTTPS certificate valid: Check browser lock icon
- [ ] Internal services not accessible:
  ```bash
  # From your machine, should fail:
  curl http://server-ip:7502  # ✗ Should timeout/refuse
  curl http://server-ip:7503  # ✗ Should timeout/refuse
  curl http://server-ip:9090  # ✗ Should timeout/refuse
  ```
- [ ] Docker logs clean: No error messages in `docker compose logs`

### Monthly

- [ ] Review Fail2ban bans: `fail2ban-client set sshd unbanip <suspicious-ips>`
- [ ] Check for unpatched packages: `apt list --upgradable`
- [ ] Verify backups exist: `docker exec allora-api-db ls /var/lib/postgresql/backups/`
- [ ] Test backup restore (optional): `docker exec allora-api-db psql -U allora < test-restore.sql`
- [ ] Review logs for anomalies: `grep "ERROR\|CRITICAL" /var/log/syslog`

### Quarterly

- [ ] Rotate SSH key
- [ ] Rotate GitHub token (if used for GHCR)
- [ ] Review Hetzner Cloud API token access
- [ ] Full security audit (penetration test if critical)

---

## Incident Response

### SSH Brute-Force Attack

**Detection:**
```bash
fail2ban-client status sshd | grep "Currently banned"
```

**Response:**
```bash
# Check what was banned
grep "Ban" /var/log/fail2ban.log | tail -20

# If legitimate IP was banned:
fail2ban-client set sshd unbanip <your-ip>

# Tighten SSH rules:
nano /etc/fail2ban/jail.local  # Lower maxretry to 2
systemctl restart fail2ban
```

### Docker Container Compromised

**If you suspect a container is compromised:**

```bash
# Stop all services
docker compose -f compose.prod.yml down

# Inspect container logs
docker logs <container-id>

# Remove compromised image (if necessary)
docker rmi <image-id>

# Update to patched image
docker compose -f compose.prod.yml pull
docker compose -f compose.prod.yml up -d
```

### Database Breach

**If database is compromised:**

```bash
# Stop services
docker compose -f compose.prod.yml down

# Restore from clean backup
docker exec allora-api-db psql -U allora < /var/lib/postgresql/backups/postgres_CLEAN_DATE.sql

# Change all passwords
# Update .env file

# Restart services
docker compose -f compose.prod.yml up -d

# Review database access logs
docker exec allora-api-db psql -U allora -c "SELECT * FROM pg_stat_statements"
```

---

## Future Security Improvements

- [ ] Enable PostgreSQL audit logging
- [ ] Add centralized logging (ELK stack or Grafana Loki)
- [ ] Implement API rate limiting (Caddy templates provided)
- [ ] Add image scanning to CI/CD (Trivy, Snyk)
- [ ] Implement secrets rotation (HashiCorp Vault, AWS Secrets Manager)
- [ ] Add DDoS protection (Cloudflare, AWS Shield)
- [ ] Regular penetration testing
- [ ] Security incident playbook

---

## References

- [Hetzner Security Best Practices](https://docs.hetzner.cloud/)
- [Docker Security Guide](https://docs.docker.com/engine/security/)
- [UFW Documentation](https://manpages.ubuntu.com/manpages/focal/man8/ufw.8.html)
- [Fail2ban Documentation](https://www.fail2ban.org/)
- [Let's Encrypt Security](https://letsencrypt.org/docs/)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)

---

## Questions?

- Security concerns: Review this document + infra/README.md
- Bootstrap issues: Check `/var/log/allora/bootstrap.log`
- Docker issues: Run `docker system df` and `docker system prune`
- Firewall issues: Review Hetzner Firewall in OpenTofu main.tf
