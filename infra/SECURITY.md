# Styx Registry Server Security

Security architecture for the Styx registry server on Hetzner cx23.

---

## Threat Model

### Intended Threats

1. **Network-level attacks** — Unauthorized access to server
2. **Compromised SSH keys** — Malicious deployments
3. **DDoS attacks** — Service availability
4. **Spec tampering** — Altered tool specifications

### Out of Scope

- Compromised hosting provider
- Nation-state attacks
- Endpoint compromise (if attacker has root)

---

## Security Layers

### Layer 1: Firewall (UFW)

On the Hetzner cx23, UFW restricts incoming traffic:

```bash
ufw allow 22/tcp    # SSH (password-disabled, key-only)
ufw allow 80/tcp    # HTTP (redirects to HTTPS)
ufw allow 443/tcp   # HTTPS (TLS via Let's Encrypt)
ufw deny incoming   # Everything else blocked
```

**Why:** Prevents unauthorized access attempts to internal services (Docker internal ports, etc.)

### Layer 2: TLS/HTTPS (Caddy + Let's Encrypt)

- **Automatic certificate renewal** — Caddy renews every 30 days
- **HSTS header** — Forces HTTPS, prevents downgrade attacks
- **Security headers** — X-Content-Type-Options, X-Frame-Options, etc.

### Layer 3: SSH Hardening

From `infra/deployment/provision-cx23.sh`:

```bash
# Key-only authentication
PasswordAuthentication no
PubkeyAuthentication yes

# Restrict cipher suites
KexAlgorithms diffie-hellman-group-exchange-sha256

# Prevent brute-force
MaxAuthTries 3
MaxSessions 10
```

### Layer 4: Non-Root Service User

Registry server runs as `styx:styx` (UID 1000) inside Docker, not root.

```dockerfile
RUN addgroup -g 1000 styx && adduser -D -u 1000 -G styx styx
USER styx
```

**Benefit:** If registry container is compromised, attacker has limited privileges.

---

## Attack Scenarios & Mitigations

### Scenario 1: Spec File Tampering

**Attack:** Attacker modifies `pkg/registry/specs/golang.toml` to serve malicious binary

**Mitigation:**
1. Git history is immutable (hosted on GitHub)
2. Specs are version-controlled (each change is tracked)
3. GitHub Actions require code review before merge
4. Each build is auditable (commit hash visible in logs)

**To verify authenticity:**
```bash
# Check which commit deployed the registry
curl https://registry.styx.sh/health

# See that commit's specs in GitHub
git log --oneline -- pkg/registry/specs/
```

### Scenario 2: SSH Key Compromise

**Attack:** Attacker obtains GitHub Actions SSH key

**Mitigation:**
1. Separate SSH key for deployments (not root)
2. Deploy user (`styx`) has limited permissions (only docker, no system access)
3. SSH key stored as GitHub secret (encrypted, not in repo)
4. Deployments are audited in GitHub Actions logs

**Recovery:**
```bash
# 1. Revoke compromised key immediately in GitHub
Settings → Secrets and variables → Actions → Delete STYX_DEPLOY_KEY

# 2. Generate new SSH key locally
ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519_styx -C "styx"

# 3. Add new key to server
ssh-copy-id -i ~/.ssh/id_ed25519_styx.pub styx@<server_ip>

# 4. Add new key to GitHub secrets
Settings → Secrets → New secret → STYX_DEPLOY_KEY
```

### Scenario 3: DDoS Attack

**Attack:** Attacker floods registry.json with requests

**Mitigation:**
1. Registry.json is small (< 2MB) and cached by Caddy
2. Caddy reverse proxy handles rate limiting (can be enabled)
3. Hetzner DDoS protection (optional, paid feature)

**To enable rate limiting in Caddy:**
Edit `infra/compose/Caddyfile`:
```
rate_limit /registry.json 100 100 by ip
```
This limits to 100 requests per IP per second.

### Scenario 4: Container Escape

**Attack:** Attacker exploits Docker/container vulnerability

**Mitigation:**
1. Minimal Alpine Linux image (3.19) reduces attack surface
2. Non-root user inside container (styx, not root)
3. Read-only filesystems where possible (specs are baked in)
4. No shell in container (can't exec into it easily)

**To verify:**
```bash
# What's in the image?
docker run --rm styx-registry ls -la /

# No bash/sh? (output will be minimal)
docker run --rm styx-registry which bash  # should fail
```

---

## Data & Access

### Data at Rest

- **Specs:** Baked into Docker image at build time (immutable)
- **Certificates:** In Docker volume `caddy_config` (owned by caddy service)
- **Logs:** Auto-rotated via compose config (10MB max, 3 files)

### Data in Transit

- HTTP → HTTPS redirect (all traffic encrypted)
- TLS 1.3 required (enforced by Caddy)
- Mutual TLS not implemented (not needed for public registry)

### Access Control

| User | Access | Can |
|------|--------|-----|
| styx (deploy) | Docker, sudo, git | Deploy new versions, view logs, restart services |
| root | Everything | System administration (emergency only) |
| Public | HTTP/HTTPS | Read registry.json, health checks |

---

## Compliance & Auditing

### What's Auditable

1. **GitHub Commits** — Every spec change is tracked
2. **GitHub Actions** — Every deployment is logged
3. **Docker Image** — Built from tagged commit (repeatable)
4. **Server Logs** — Caddy access logs show all requests
5. **Git Log** — `git log` shows who changed what, when

### How to Audit

```bash
# See deployment history
gh run list --workflow deploy-registry.yml --limit 10

# See spec changes
git log --oneline -- pkg/registry/specs/

# See server logs (on server)
docker logs styx-registry | head -20
docker logs styx-caddy | head -20
```

---

## Incident Response

### If Registry is Down

1. Check server health:
   ```bash
   ssh styx@<server_ip> "docker compose -f /opt/styx/compose.prod.yml ps"
   ```

2. View logs:
   ```bash
   ssh styx@<server_ip> "docker compose -f /opt/styx/compose.prod.yml logs registry"
   ```

3. Restart if needed:
   ```bash
   ssh styx@<server_ip> "docker compose -f /opt/styx/compose.prod.yml restart registry"
   ```

### If Specs are Corrupted

1. Revert to known-good commit:
   ```bash
   git revert <bad_commit>
   git push origin main
   ```
   GitHub Actions will auto-deploy.

2. Or manually restore:
   ```bash
   ssh styx@<server_ip> "cd /opt/styx && git checkout <good_commit>"
   /home/styx/deploy.sh
   ```

### If SSH Keys are Compromised

1. **Immediately:**
   - Delete compromised SSH key from GitHub secrets
   - Generate new SSH keypair

2. **Add new key to server:**
   ```bash
   ssh-copy-id -i ~/.ssh/id_ed25519_new.pub styx@<server_ip>
   ```

3. **Update GitHub secrets with new private key**

4. **Test deployment:**
   ```bash
   git commit --allow-empty -m "test: verify new deploy key works"
   git push origin main
   ```

---

## Security Checklist

Before deploying to production:

- [ ] SSH key-only auth enabled (no passwords)
- [ ] UFW firewall active (22, 80, 443 only)
- [ ] GitHub secrets configured (SSH key, domain, IP)
- [ ] TLS certificate auto-renewal enabled
- [ ] Non-root service user (styx)
- [ ] Logs are auto-rotated (10MB max)
- [ ] Specs are version-controlled
- [ ] Deployments are auditable in GitHub Actions
- [ ] Health checks pass: `curl https://registry.styx.sh/health`

---

## Regular Maintenance

**Daily:**
- Monitor health: `curl https://registry.styx.sh/health`

**Weekly:**
- Review GitHub Actions logs for failures

**Monthly:**
- Check disk usage: `ssh styx@<server_ip> "df -h"`
- Review access logs for suspicious activity

**Quarterly:**
- Rotate SSH keys (generate new pair, update GitHub)
- Review and update security practices

---

## References

- Hetzner Security Recommendations: https://docs.hetzner.cloud/
- Caddy Security Guide: https://caddyserver.com/docs/caddyfile/directives/header
- OWASP Top 10: https://owasp.org/www-project-top-ten/
- CIS Docker Benchmark: https://www.cisecurity.org/benchmark/docker
