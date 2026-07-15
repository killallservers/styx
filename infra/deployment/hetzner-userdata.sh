#!/bin/bash
# Hetzner cloud-init userData script for Styx registry server
# Copy/paste this into Hetzner console when creating server
# Or pass as: --user-data-from-file hetzner-userdata.sh

set -eu

# Configuration (edit these)
GITHUB_REPO="killallservers/styx"
SSH_PUBKEY="ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... your-public-key-here"

echo "========== Styx Registry - Hetzner cloud-init =========="
echo "Repository: $GITHUB_REPO"
echo ""

# ===== System-Level Setup (Root) =====

echo "========== System Update ==========="
apt-get update
apt-get upgrade -y
apt-get install -y curl wget git vim

echo "========== Create styx user ==========="
if ! id -u styx >/dev/null 2>&1; then
    useradd -m -s /bin/bash styx
    echo "✓ styx user created"
else
    echo "✓ styx user already exists"
fi

echo "========== Install Docker ==========="
if ! command -v docker >/dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    usermod -aG docker styx
    echo "✓ Docker installed"
else
    echo "✓ Docker already installed"
fi

echo "========== Install Docker Compose ==========="
if ! docker compose version >/dev/null 2>&1; then
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    echo "✓ Docker Compose installed"
else
    echo "✓ Docker Compose already installed"
fi

echo "========== Configure Firewall (UFW) ==========="
if ! command -v ufw >/dev/null; then
    apt-get install -y ufw
fi

ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
echo "✓ Firewall configured"

echo "========== Create Deployment Directory ==========="
mkdir -p /opt/styx
chown styx:styx /opt/styx
chmod 750 /opt/styx

echo "========== Add SSH Deploy Key ==========="
mkdir -p /home/styx/.ssh
chmod 700 /home/styx/.ssh
echo "$SSH_PUBKEY" >> /home/styx/.ssh/authorized_keys
chmod 600 /home/styx/.ssh/authorized_keys
chown -R styx:styx /home/styx/.ssh

echo "========== Setup Cron for Auto-Updates ==========="
cat > /etc/cron.daily/styx-security-updates <<'EOFCRON'
#!/bin/bash
apt-get update
apt-get install -y --only-upgrade $(apt list --upgradable 2>/dev/null | cut -d'/' -f1 | grep -E '^(linux-image|linux-headers|apt|openssl|openssh)' || true)
EOFCRON
chmod +x /etc/cron.daily/styx-security-updates

# ===== User-Level Setup (styx) =====

echo "========== Initialize styx user ==========="
sudo -u styx bash << EOFUSER
set -eu

cd /opt/styx

# Clone repository
if [ ! -d .git ]; then
    echo "Cloning repository..."
    git init
    git remote add origin "https://github.com/$GITHUB_REPO"
    git fetch origin main
    git checkout -b main origin/main
    echo "✓ Repository cloned"
else
    echo "✓ Repository already cloned"
fi

# Create deploy script
echo "Creating deploy script..."
cat > deploy.sh <<'EOFDEPLOY'
#!/bin/bash
set -eu

cd /opt/styx

# Pull latest code
git pull origin main

# Build and deploy (DOMAIN passed via GitHub Actions)
docker compose -f compose.prod.yml build --no-cache registry
docker compose -f compose.prod.yml up -d

# Verify
sleep 2
curl -f http://localhost:7506/health || exit 1
echo "✓ Registry deployed successfully"
EOFDEPLOY

chmod +x deploy.sh
echo "✓ Deploy script created"

# Add shell aliases
if ! grep -q "alias styx-deploy" ~/.bashrc 2>/dev/null; then
    cat >> ~/.bashrc <<'EOF'

# Styx registry aliases
alias styx-deploy="/opt/styx/deploy.sh"
alias styx-logs="docker compose -f /opt/styx/compose.prod.yml logs -f"
alias styx-status="docker compose -f /opt/styx/compose.prod.yml ps"
EOF
    echo "✓ Aliases added to ~/.bashrc"
fi

echo ""
echo "✅ Styx user initialization complete!"
EOFUSER

echo ""
echo "========================================="
echo "✅ Hetzner initialization complete!"
echo "========================================="
echo ""
echo "Next steps (from your local machine):"
echo "  1. Wait 2-3 minutes for boot to complete"
echo "  2. SSH into server: ssh styx@<server-ip>"
echo "  3. Deploy registry: cd /opt/styx && ./deploy.sh"
echo "  4. Verify: curl https://<your-domain>/health"
echo "  5. Configure GitHub Actions secrets:"
echo "     - STYX_DEPLOY_HOST: <server-ip>"
echo "     - STYX_DEPLOY_USER: styx"
echo "     - STYX_DEPLOY_KEY: <your-ssh-private-key>"
echo "     - STYX_REGISTRY_DOMAIN: <your-domain>"
echo ""
