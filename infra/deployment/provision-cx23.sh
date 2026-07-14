#!/bin/bash
# Provision a Hetzner cx23 for Styx registry server
# Usage: bash provision-cx23.sh <server_ip> <domain> <github_repo> <ssh_public_key>
#
# Example:
#   bash provision-cx23.sh 1.2.3.4 registry.styx.sh killallservers/styx ~/.ssh/id_rsa.pub

set -eu

if [ $# -lt 3 ]; then
    echo "Usage: $0 <server_ip> <domain> <github_repo> [ssh_public_key]"
    echo ""
    echo "Arguments:"
    echo "  server_ip       - IP address of Hetzner cx23 instance"
    echo "  domain          - Domain for registry (e.g., registry.styx.sh)"
    echo "  github_repo     - GitHub repo path (e.g., killallservers/styx)"
    echo "  ssh_public_key  - Path to SSH public key (default: ~/.ssh/id_rsa.pub)"
    echo ""
    echo "Example:"
    echo "  bash provision-cx23.sh 1.2.3.4 registry.styx.sh killallservers/styx ~/.ssh/id_rsa.pub"
    exit 1
fi

SERVER_IP="$1"
DOMAIN="$2"
GITHUB_REPO="$3"
SSH_KEY="${4:-$HOME/.ssh/id_rsa.pub}"

if [ ! -f "$SSH_KEY" ]; then
    echo "Error: SSH public key not found: $SSH_KEY"
    exit 1
fi

SSH_PUBKEY=$(cat "$SSH_KEY")
DEPLOY_KEY_PATH="/home/styx/.ssh/deploy_key"

echo "====================================="
echo "Styx Registry - Hetzner cx23 Setup"
echo "====================================="
echo "Server IP:    $SERVER_IP"
echo "Domain:       $DOMAIN"
echo "GitHub Repo:  $GITHUB_REPO"
echo ""
echo "This script will:"
echo "  1. Create 'styx' user (non-root deployment)"
echo "  2. Install Docker & Docker Compose"
echo "  3. Set up firewall (SSH, HTTP, HTTPS)"
echo "  4. Create deployment directory (/opt/styx)"
echo "  5. Add SSH key for automated CI/CD deployments"
echo "  6. Configure auto-security updates"
echo ""
read -p "Continue? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
fi

echo ""
echo "Connecting to server and running setup..."
echo ""

# Create provisioning script to run on remote
REMOTE_SCRIPT=$(cat <<'EOFSCRIPT'
#!/bin/bash
set -eu

DOMAIN="$1"
GITHUB_REPO="$2"
SSH_PUBKEY="$3"

echo "========== System Update ==========="
apt-get update
apt-get upgrade -y
apt-get install -y curl wget git vim

echo "========== Create styx user ==========="
if ! id -u styx >/dev/null 2>&1; then
    useradd -m -s /bin/bash styx
    echo "styx user created"
else
    echo "styx user already exists"
fi

echo "========== Install Docker ==========="
if ! command -v docker >/dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    usermod -aG docker styx
    echo "Docker installed"
else
    echo "Docker already installed"
fi

echo "========== Install Docker Compose ==========="
docker --version
if ! docker compose version >/dev/null 2>&1; then
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
    echo "Docker Compose installed"
else
    echo "Docker Compose already installed"
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
ufw status

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

echo "========== Create Deploy Hook ==========="
cat > /home/styx/deploy.sh <<'EOFDEPLOY'
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

chmod +x /home/styx/deploy.sh
chown styx:styx /home/styx/deploy.sh

echo "========== Setup Cron for Auto-Updates ==========="
cat > /etc/cron.daily/styx-security-updates <<'EOFCRON'
#!/bin/bash
apt-get update
apt-get install -y --only-upgrade $(apt list --upgradable 2>/dev/null | cut -d'/' -f1 | grep -E '^(linux-image|linux-headers|apt|openssl|openssh)' || true)
EOFCRON

chmod +x /etc/cron.daily/styx-security-updates

echo ""
echo "========================================="
echo "✓ Hetzner cx23 provisioning complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "  1. Clone repository into /opt/styx:"
echo "     cd /opt/styx && git init && git remote add origin https://github.com/$GITHUB_REPO"
echo "  2. Pull main branch"
echo "  3. Run: /home/styx/deploy.sh"
echo "  4. Verify: curl https://$DOMAIN/health"
echo ""
EOFSCRIPT
)

# Run remote provisioning
ssh -o StrictHostKeyChecking=no root@"$SERVER_IP" bash -s "$DOMAIN" "$GITHUB_REPO" "$SSH_PUBKEY" <<< "$REMOTE_SCRIPT"

echo ""
echo "========================================="
echo "Server provisioned! Now:"
echo "========================================="
echo ""
echo "1. SSH into the server and initialize the repository:"
echo "   ssh styx@$SERVER_IP"
echo "   cd /opt/styx"
echo "   git init"
echo "   git remote add origin https://github.com/$GITHUB_REPO"
echo "   git pull origin main"
echo ""
echo "2. Deploy the registry:"
echo "   ./deploy.sh"
echo ""
echo "3. Verify it's working:"
echo "   curl https://$DOMAIN/health"
echo ""
echo "4. Update GitHub Actions secrets (see GITHUB_SECRETS_SETUP.md):"
echo "   - STYX_DEPLOY_HOST: $SERVER_IP"
echo "   - STYX_DEPLOY_USER: styx"
echo "   - STYX_DEPLOY_KEY: (private key corresponding to $SSH_KEY)"
echo "   - STYX_REGISTRY_DOMAIN: $DOMAIN"
echo ""
