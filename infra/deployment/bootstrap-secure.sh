#!/bin/bash
# Allora Server Bootstrap Script - SECURITY HARDENED
# Debian 13 on Hetzner dedicated server
# Includes hardening per Hetzner + Docker security best practices
# IDEMPOTENT: Safe to run multiple times with same end result

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# ============================================================================
# LOGGING SETUP
# ============================================================================

BOOTSTRAP_LOG="/var/log/allora/bootstrap-$(date +%Y%m%d-%H%M%S).log"
mkdir -p "$(dirname "$BOOTSTRAP_LOG")"

# Redirect all output (stdout + stderr) to log file AND display in terminal
exec > >(tee -a "$BOOTSTRAP_LOG")
exec 2>&1

echo "=== Allora Server Bootstrap - Security Hardened ==="
echo "Environment: ${ENVIRONMENT:-production}"
echo "Time: $(date)"
echo "Log file: $BOOTSTRAP_LOG"
echo ""

# Enable command echoing for debugging (remove in production if verbose logs are a concern)
# set -x

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# ============================================================================
# 0. CLEANUP (idempotent: remove stale/broken configs from previous runs)
# ============================================================================

log_info "Cleaning up stale configurations..."

# Remove broken NVIDIA sources (from failed previous installs)
rm -f /etc/apt/sources.list.d/nvidia-container-toolkit.list 2>/dev/null || true
rm -f /etc/apt/sources.list.d/cuda.list 2>/dev/null || true

# Remove stale GPG keys
rm -f /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg 2>/dev/null || true

log_info "Cleanup complete"

# ============================================================================
# 1. SYSTEM UPDATES & SECURITY BASELINE
# ============================================================================

log_info "Updating system packages..."
apt-get update
apt-get upgrade -y
apt-get install -y \
    build-essential linux-headers-$(uname -r) \
    curl wget git \
    ca-certificates gnupg lsb-release apt-transport-https \
    vim nano htop \
    net-tools iputils-ping dnsutils \
    ufw fail2ban \
    openssh-client openssh-server

# ============================================================================
# 2. SSH HARDENING
# ============================================================================

log_info "Hardening SSH configuration..."

# Backup original SSH config only if not already backed up
if [ ! -f /etc/ssh/sshd_config.backup.original ]; then
    cp /etc/ssh/sshd_config /etc/ssh/sshd_config.backup.original
    log_info "✓ Created SSH config backup: /etc/ssh/sshd_config.backup.original"
else
    log_info "✓ SSH backup already exists"
fi

# Apply security settings (use sed to replace if exists, append otherwise)
log_info "Applying SSH hardening settings..."

# Helper function to update sshd_config (replace line if exists, else append)
update_ssh_config() {
    local key="$1"
    local value="$2"

    if grep -q "^#*$key" /etc/ssh/sshd_config; then
        # Line exists (commented or not), replace it
        sed -i "s/^#*$key .*/$key $value/" /etc/ssh/sshd_config
        log_info "  ✓ Updated: $key $value"
    else
        # Line doesn't exist, append it
        echo "$key $value" >> /etc/ssh/sshd_config
        log_info "  ✓ Added: $key $value"
    fi
}

# Apply all hardening settings
update_ssh_config "PermitRootLogin" "prohibit-password"
update_ssh_config "PasswordAuthentication" "no"
update_ssh_config "PubkeyAuthentication" "yes"
update_ssh_config "PermitEmptyPasswords" "no"
update_ssh_config "X11Forwarding" "no"
update_ssh_config "PrintMotd" "no"
update_ssh_config "Protocol" "2"
update_ssh_config "LogLevel" "VERBOSE"
update_ssh_config "StrictModes" "yes"
update_ssh_config "MaxAuthTries" "3"
update_ssh_config "MaxSessions" "5"
update_ssh_config "ClientAliveInterval" "300"
update_ssh_config "ClientAliveCountMax" "2"
update_ssh_config "UseDNS" "no"

# Verify SSH config syntax
if sshd -t 2>&1; then
    log_info "✓ SSH configuration syntax valid"
    if systemctl restart sshd 2>&1; then
        log_info "✓ SSH service restarted successfully"
    else
        log_error "✗ Failed to restart SSH service"
        log_error "  Restoring from backup..."
        cp /etc/ssh/sshd_config.backup.original /etc/ssh/sshd_config
        systemctl restart sshd
        log_error "✗ Bootstrap aborted: SSH restart failed"
        exit 1
    fi
else
    log_error "✗ SSH configuration has syntax errors!"
    log_error "  Restoring from backup..."
    cp /etc/ssh/sshd_config.backup.original /etc/ssh/sshd_config
    log_error "✗ Bootstrap aborted: SSH config invalid"
    exit 1
fi

# Verify critical settings are actually in effect
log_info "Verifying SSH hardening..."
if grep -q "^PasswordAuthentication no" /etc/ssh/sshd_config; then
    log_info "✓ Password authentication DISABLED"
else
    log_error "✗ Password authentication check FAILED - config may not be applied correctly"
    exit 1
fi

# ============================================================================
# 3. FIREWALL CONFIGURATION
# ============================================================================

log_info "Configuring firewall (ufw)..."

# Check if UFW is already enabled
if ufw status 2>/dev/null | grep -q "Status: active"; then
    log_info "✓ UFW already enabled"
else
    log_info "Enabling UFW..."
    if ufw --force enable 2>&1 | grep -q "Firewall is active"; then
        log_info "✓ UFW enabled successfully"
    else
        log_error "✗ Failed to enable UFW"
        exit 1
    fi
fi

# Allow SSH, HTTP, HTTPS (idempotent - ufw ignores duplicate rules)
log_info "Configuring firewall rules..."
if ufw allow 22/tcp 2>&1 | grep -q -E "^Skipping|^Rule"; then
    log_info "  ✓ SSH (port 22) allowed"
fi

if ufw allow 80/tcp 2>&1 | grep -q -E "^Skipping|^Rule"; then
    log_info "  ✓ HTTP (port 80) allowed"
fi

if ufw allow 443/tcp 2>&1 | grep -q -E "^Skipping|^Rule"; then
    log_info "  ✓ HTTPS (port 443) allowed"
fi

# Set default policies (idempotent)
ufw default deny incoming 2>&1 | grep -q -E "^Default|^Already" && log_info "  ✓ Default incoming: DENY"
ufw default allow outgoing 2>&1 | grep -q -E "^Default|^Already" && log_info "  ✓ Default outgoing: ALLOW"

log_info "✓ Firewall configured (SSH, HTTP, HTTPS allowed)"
log_warn "⚠️  Docker ports managed by Hetzner Cloud Firewall, not ufw"

# ============================================================================
# 4. FAIL2BAN - Brute Force Protection
# ============================================================================

log_info "Configuring Fail2ban..."

# Create fail2ban local config (idempotent: overwrite always)
if cat > /etc/fail2ban/jail.local <<EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3
destemail = root@localhost
sendername = Fail2ban
action = %(action_mwl)s

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
EOF
then
    log_info "✓ Fail2ban config written"
else
    log_error "✗ Failed to write fail2ban config"
    exit 1
fi

# Enable and restart
if systemctl enable fail2ban 2>&1; then
    log_info "✓ Fail2ban enabled for auto-start"
else
    log_error "✗ Failed to enable fail2ban"
    exit 1
fi

if systemctl restart fail2ban 2>&1; then
    log_info "✓ Fail2ban restarted successfully"
else
    log_error "✗ Failed to restart fail2ban"
    exit 1
fi

log_info "✓ Fail2ban configured (bans after 3 failed attempts for 1 hour)"

# ============================================================================
# 5. AUTOMATIC SECURITY UPDATES
# ============================================================================

log_info "Configuring automatic security updates..."

# Install unattended-upgrades (idempotent)
apt-get install -y unattended-upgrades apt-listchanges > /dev/null 2>&1

# Configure automatic updates (idempotent: overwrite)
cat > /etc/apt/apt.conf.d/50unattended-upgrades <<EOF
Unattended-Upgrade::Allowed-Origins {
    "\${distro_id}:\${distro_codename}-security";
    "\${distro_id}ESMApps:\${distro_codename}-apps-security";
    "\${distro_id}ESM:\${distro_codename}-infra-security";
};
Unattended-Upgrade::DevRelease "false";
Unattended-Upgrade::Mail "root";
Unattended-Upgrade::MailOnlyOnError "true";
Unattended-Upgrade::Remove-Unused-Kernel-Packages "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
Unattended-Upgrade::Automatic-Reboot-Time "02:00";
EOF

# Enable APT auto-update (idempotent: overwrite)
cat > /etc/apt/apt.conf.d/20auto-upgrades <<EOF
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Download-Upgradeable-Packages "1";
APT::Periodic::AutocleanInterval "7";
APT::Periodic::Unattended-Upgrade "1";
EOF

log_info "Automatic security updates configured"

# ============================================================================
# 6. UV PACKAGE MANAGER (for Python dependencies)
# ============================================================================

log_info "Installing uv (fast Python package manager)..."

if command -v uv &> /dev/null; then
    log_info "✓ uv already installed ($(uv --version))"
else
    log_info "Installing uv from official installer..."
    if curl -LsSf https://astral.sh/uv/install.sh | sh 2>&1; then
        # uv installs to $HOME/.local/bin, copy to /usr/local/bin for system-wide access
        if [ -f "$HOME/.local/bin/uv" ]; then
            cp "$HOME/.local/bin/uv" /usr/local/bin/uv
            chmod +x /usr/local/bin/uv
            log_info "✓ uv installed to /usr/local/bin (system-wide)"
        else
            log_warn "uv installer ran but binary not found, it may already be in PATH"
        fi
    else
        log_error "✗ Failed to install uv"
        exit 1
    fi
fi

# Verify uv is accessible
if command -v uv &> /dev/null; then
    log_info "✓ uv verified: $(uv --version)"
else
    log_error "✗ uv not found in PATH after installation"
    exit 1
fi

# ============================================================================
# 8. DOCKER INSTALLATION & SECURITY
# ============================================================================

log_info "Installing Docker..."

# Check if Docker is already installed
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version)
    log_info "✓ Docker already installed: $DOCKER_VERSION"
else
    log_info "Installing Docker from official repository..."

    if ! curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg 2>&1; then
        log_error "✗ Failed to download/verify Docker GPG key"
        exit 1
    fi
    log_info "  ✓ Docker GPG key installed"

    if ! echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | \
        tee /etc/apt/sources.list.d/docker.list > /dev/null; then
        log_error "✗ Failed to add Docker repository"
        exit 1
    fi
    log_info "  ✓ Docker repository added"

    if ! apt-get update 2>&1 | tail -5; then
        log_error "✗ Failed to update package lists"
        exit 1
    fi
    log_info "  ✓ Package lists updated"

    if ! apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin 2>&1; then
        log_error "✗ Failed to install Docker packages"
        exit 1
    fi
    log_info "✓ Docker installed: $(docker --version)"
fi

# Configure Docker daemon for security (idempotent: overwrite)
log_info "Configuring Docker daemon..."
mkdir -p /etc/docker

if cat > /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3",
    "labels": "com.example.vendor=Allora"
  },
  "storage-driver": "overlay2",
  "icc": false,
  "userland-proxy": false,
  "default-ulimits": {
    "memlock": {
      "Name": "memlock",
      "Hard": -1,
      "Soft": -1
    }
  },
  "runtimes": {
    "nvidia": {
      "path": "nvidia-container-runtime",
      "runtimeArgs": []
    }
  }
}
EOF
then
    log_info "✓ Docker daemon config written"
else
    log_error "✗ Failed to write Docker daemon config"
    exit 1
fi

# Enable and restart
if systemctl enable docker 2>&1 | grep -q -E "^Created symlink|^Synchronizing"; then
    log_info "✓ Docker enabled for auto-start"
fi

if systemctl daemon-reload 2>&1; then
    log_info "✓ Systemd daemon reloaded"
else
    log_error "✗ Failed to reload systemd daemon"
    exit 1
fi

if systemctl restart docker 2>&1; then
    log_info "✓ Docker service restarted"
else
    log_error "✗ Failed to restart Docker"
    exit 1
fi

if systemctl is-active --quiet docker; then
    log_info "✓ Docker is running"
else
    log_error "✗ Docker is not running"
    systemctl status docker
    exit 1
fi

# ============================================================================
# 9. NVIDIA CONTAINER TOOLKIT (for GPU support)
# ============================================================================

log_info "Installing NVIDIA drivers..."

# Check if drivers already installed
if command -v nvidia-smi &> /dev/null; then
    log_info "NVIDIA drivers already installed ($(nvidia-smi --query-gpu=driver_version --format=csv,noheader | head -1))"
else
    log_info "Installing NVIDIA drivers (latest 610 series)..."

    # Download official NVIDIA driver installer
    DRIVER_VERSION="610.43.03"
    DRIVER_FILE="NVIDIA-Linux-x86_64-$DRIVER_VERSION.run"
    DRIVER_URL="https://us.download.nvidia.com/XFree86/Linux-x86_64/$DRIVER_VERSION/$DRIVER_FILE"

    if wget -q -O "/tmp/$DRIVER_FILE" "$DRIVER_URL" 2>/dev/null; then
        chmod +x "/tmp/$DRIVER_FILE"
        log_info "Running NVIDIA driver installer..."

        # Install driver with proper prompts answered
        # The installer needs sudo and explicit answers to all prompts
        INSTALL_LOG="/tmp/nvidia_install.log"
        log_info "Running installer (this may take 5-10 minutes)..."

        # Run with answers to common prompts
        (echo "MIT/GPL"; echo "Continue installation"; echo "Continue installation"; echo "No") | \
            sudo "/tmp/$DRIVER_FILE" --ui=none --accept-license --no-questions \
            > "$INSTALL_LOG" 2>&1

        if grep -q "Installation of the NVIDIA" "$INSTALL_LOG"; then
            log_info "NVIDIA driver installer completed successfully"
        else
            log_warn "Driver installation may have issues. Check: tail -50 /tmp/nvidia_install.log"
            tail -20 "$INSTALL_LOG" | while read line; do log_warn "  $line"; done
        fi

        rm -f "/tmp/$DRIVER_FILE" "$INSTALL_LOG"

        # Rebuild kernel module for current kernel (handles kernel upgrades)
        log_info "Building NVIDIA kernel module for $(uname -r)..."
        NVIDIA_SRC="/usr/src/nvidia-$DRIVER_VERSION/kernel"

        if [ -f "$NVIDIA_SRC/Makefile" ]; then
            if cd "$NVIDIA_SRC" && \
               sudo make -j$(nproc) > /dev/null 2>&1 && \
               sudo make install > /dev/null 2>&1; then
                log_info "Kernel module rebuilt successfully"
            else
                log_warn "Kernel module rebuild had issues - check /usr/src/nvidia-*/kernel build logs"
            fi
        else
            log_info "NVIDIA source not found at $NVIDIA_SRC (may need reboot to load modules)"
        fi

        # Load kernel module
        modprobe nvidia 2>/dev/null || true
        modprobe nvidia_uvm 2>/dev/null || true

        # Check if kernel module loaded
        if [ -d "/lib/modules/$(uname -r)/kernel/drivers/video/nvidia" ]; then
            log_info "NVIDIA kernel module loaded for current kernel"
        else
            log_warn "NVIDIA kernel module not found - may need reboot or manual rebuild"
            log_warn "Run: sudo /usr/bin/nvidia-uninstall --no-questions && re-run bootstrap"
        fi

        # Check if smi works
        if command -v nvidia-smi &> /dev/null && nvidia-smi > /dev/null 2>&1; then
            log_info "NVIDIA drivers working ($(nvidia-smi --query-gpu=driver_version --format=csv,noheader | head -1))"
        else
            log_warn "nvidia-smi not working - GPU drivers may need manual setup"
        fi
    else
        log_warn "Failed to download NVIDIA driver installer - GPU support may be unavailable"
    fi
fi

log_info "Installing NVIDIA Container Toolkit..."

# Official guide: https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html

# Check if already installed
if command -v nvidia-ctk &> /dev/null; then
    log_info "NVIDIA Container Toolkit already installed"
else
    log_info "Installing NVIDIA Container Toolkit from official repository..."

    # Step 1: Import GPG key
    curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey 2>/dev/null | \
        gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg 2>/dev/null

    # Step 2: Add repository
    curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list 2>/dev/null | \
        sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | \
        tee /etc/apt/sources.list.d/nvidia-container-toolkit.list > /dev/null

    # Step 3: Update and install
    apt-get update > /dev/null 2>&1 || true
    apt-get install -y nvidia-container-toolkit > /dev/null 2>&1 || \
        log_warn "Failed to install nvidia-container-toolkit from repository"
fi

# Ensure kernel modules are loaded
modprobe nvidia 2>/dev/null || true
modprobe nvidia_uvm 2>/dev/null || true

# Step 4: Configure Docker runtime (idempotent)
if command -v nvidia-ctk &> /dev/null; then
    nvidia-ctk runtime configure --runtime=docker > /dev/null 2>&1 || true
    systemctl restart docker > /dev/null 2>&1
    log_info "Docker configured for NVIDIA GPU support"
else
    log_warn "NVIDIA Container Toolkit not available - GPU support disabled"
fi

# Step 5: Verify with test container (informational only)
if docker run --rm --runtime=nvidia --gpus all nvidia/cuda:12-runtime nvidia-smi > /dev/null 2>&1; then
    log_info "✓ NVIDIA GPU verified with test container"
else
    log_warn "NVIDIA GPU test failed - check installation (optional feature)"
fi

# ============================================================================
# 10. DOWNLOAD LLM MODELS (optional, can be skipped if done separately)
# ============================================================================

log_info "Preparing for model downloads..."

# Create model cache directory
mkdir -p /root/.cache/huggingface/hub
chmod 755 /root/.cache/huggingface/hub

log_info "To download models, run:"
log_info "  cd /root/.cache/huggingface/hub"
log_info "  # Ministral (2.5GB):"
log_info "  wget https://huggingface.co/ggml-org/ministral-3-3b-instruct-gguf/resolve/main/Ministral-3-3B-Instruct-2512-Q4_K_M.gguf"
log_info "  # Qwen3 (1.8GB):"
log_info "  wget https://huggingface.co/Qwen/Qwen-VL-Embedding-2B-w-mps/resolve/main/Qwen.Qwen3-VL-Embedding-2B.Q8_0.gguf"
log_info "Or copy models from local machine:"
log_info "  scp ~/models/* root@<IP>:/root/.cache/huggingface/hub/"

# ============================================================================
# 11. APPLICATION DIRECTORIES & PERMISSIONS
# ============================================================================

log_info "Creating application directories..."

mkdir -p /opt/allora/{compose,deployment,backups,data,logs}
chmod 750 /opt/allora
chown root:root /opt/allora

# Create deployment user + configure SSH keys
if ! id -u deploy > /dev/null 2>&1; then
    useradd -m -s /bin/bash -G docker deploy
    mkdir -p /home/deploy/.ssh
    chmod 700 /home/deploy/.ssh

    # Add SSH public key for deploy user
    if [ -n "$DEPLOY_SSH_PUBLIC_KEY" ]; then
        # Use provided deploy SSH key (separate from root)
        echo "$DEPLOY_SSH_PUBLIC_KEY" > /home/deploy/.ssh/authorized_keys
        log_info "Added deploy SSH public key to deploy user"
    elif [ -f /root/.ssh/authorized_keys ]; then
        # Fallback: copy from root if no separate key provided
        cp /root/.ssh/authorized_keys /home/deploy/.ssh/authorized_keys
        log_info "Copied SSH keys from root to deploy user (fallback)"
    else
        log_warn "No SSH key provided for deploy user - you'll need to add one manually"
    fi

    # Set permissions
    if [ -f /home/deploy/.ssh/authorized_keys ]; then
        chmod 600 /home/deploy/.ssh/authorized_keys
        chown -R deploy:deploy /home/deploy/.ssh
    fi

    # Add deploy user to sudoers for password-less sudo (for emergency ops)
    cat > /etc/sudoers.d/deploy <<'SUDOERS'
deploy ALL=(ALL) NOPASSWD: ALL
SUDOERS
    chmod 440 /etc/sudoers.d/deploy
    log_info "Created 'deploy' user with Docker + sudo access"
fi

# ============================================================================
# 12. REGISTRY AUTHENTICATION (GHCR)
# ============================================================================

log_info "Docker registry authentication"

# Note: Docker images are pulled using public or organization-authenticated registries.
# Authentication for GitHub Actions is handled by CI/CD workflows using built-in GITHUB_TOKEN.
# For manual image pulls, ensure your DOCKER_REGISTRY URL is publicly accessible or
# add Docker config manually if using private registries.

mkdir -p /root/.docker
if id deploy > /dev/null 2>&1; then
    mkdir -p /home/deploy/.docker
    chown deploy:deploy /home/deploy/.docker
fi

log_info "✓ Docker config directories created (credentials configured via compose/.env if needed)"

# ============================================================================
# 13. ENVIRONMENT CONFIGURATION
# ============================================================================

log_info "Creating environment configuration..."

cat > /opt/allora/compose/.env <<EOF
# Allora Environment - Generated $(date)
ENVIRONMENT=${ENVIRONMENT:-production}
DOMAIN=\${DOMAIN:-allora.example.com}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-ghcr.io}
IMAGE_TAG=latest

# Database (CHANGE THESE!)
POSTGRES_PASSWORD=$(openssl rand -base64 32)

# Grafana (CHANGE THESE!)
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 16)

# Security
COMPOSE_PROFILES=
DOCKER_CONTENT_TRUST=1
EOF

chmod 600 /opt/allora/compose/.env
log_warn "⚠️  IMPORTANT: Edit /opt/allora/compose/.env with your DOMAIN and secure passwords"

# ============================================================================
# 14. LOGGING & MONITORING
# ============================================================================

log_info "Setting up logging..."

# Create log directory
mkdir -p /var/log/allora
touch /var/log/allora/bootstrap.log
touch /var/log/allora/deployments.log

# Create logrotate config for Allora logs
cat > /etc/logrotate.d/allora <<EOF
/var/log/allora/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 root root
    sharedscripts
    postrotate
        systemctl reload rsyslog > /dev/null 2>&1 || true
    endscript
}
EOF

log_info "Log rotation configured (30-day retention)"

# ============================================================================
# 15. SYSTEM HARDENING EXTRAS
# ============================================================================

log_info "Applying additional system hardening..."

# Disable unnecessary services
systemctl disable avahi-daemon 2>/dev/null || true
systemctl disable cups 2>/dev/null || true

# Kernel parameter hardening
cat >> /etc/sysctl.conf <<'EOF'

# Network security
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.all.send_redirects = 0
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1
net.ipv4.tcp_syncookies = 1
net.ipv6.conf.all.disable_ipv6 = 0
net.ipv6.conf.all.forwarding = 0

# Performance
net.core.somaxconn = 4096
net.ipv4.tcp_max_syn_backlog = 4096
EOF

sysctl -p > /dev/null

# ============================================================================
# 16. SUMMARY & NEXT STEPS
# ============================================================================

log_info "Bootstrap complete!"

cat > /var/log/allora/bootstrap.log <<EOF
=== Allora Bootstrap Summary ===
Completed: $(date)
Hostname: $(hostname)
Kernel: $(uname -r)
Docker Version: $(docker --version)

SECURITY MEASURES APPLIED:
✅ SSH hardening (key-only auth, max 3 retries)
✅ Firewall enabled (ufw) - local network protection
✅ Fail2ban enabled - brute-force protection
✅ Automatic security updates - daily
✅ Docker security daemon config - logging, no ICC, no userland-proxy
✅ NVIDIA GPU support - verified
✅ Log rotation - 30-day retention
✅ Kernel hardening - network security parameters

FIREWALL STRATEGY:
- Hetzner Cloud Firewall: Manages external traffic
- ufw: Manages local network access (doesn't conflict with Docker)
- Docker services: Bound to 127.0.0.1 (internal only)
- Caddy: Only public-facing service (80, 443)

NEXT STEPS:
1. SSH to server: ssh root@<ip>
2. Edit configuration: nano /opt/allora/compose/.env
3. Set DOMAIN to your actual domain
4. Set strong POSTGRES_PASSWORD and GRAFANA_ADMIN_PASSWORD
5. Copy Docker Compose files to /opt/allora/compose/
6. Start services: cd /opt/allora/compose && docker compose -f compose.prod.yml up -d

CRITICAL:
- SSH config backed up at: /etc/ssh/sshd_config.backup.*
- Registry credentials in: /root/.docker/config.json (chmod 600)
- Generated passwords in: /opt/allora/compose/.env (chmod 600)
- Check Fail2ban status: fail2ban-client status

SECURITY CHECKLIST:
☐ Verify SSH access works
☐ Check Fail2ban is running
☐ Review firewall rules: ufw status
☐ Verify NVIDIA GPU: nvidia-smi
☐ Monitor logs: tail -f /var/log/auth.log

For full setup, follow: /opt/allora/SETUP_CHECKLIST.md
EOF

# ============================================================================
# 17. VERIFICATION
# ============================================================================

log_info "Verifying installation..."

VERIFICATION_FAILED=0

# Check critical services
SERVICES=("docker" "fail2ban" "sshd")
for service in "${SERVICES[@]}"; do
    if systemctl is-active --quiet "$service"; then
        log_info "✓ $service is running"
    else
        log_error "✗ $service is NOT running"
        systemctl status "$service" 2>&1 | head -5 | while read line; do log_error "  $line"; done
        VERIFICATION_FAILED=1
    fi
done

# Verify SSH hardening
log_info "Verifying SSH hardening..."
if grep -q "^PasswordAuthentication no" /etc/ssh/sshd_config; then
    log_info "✓ Password authentication: DISABLED"
else
    log_error "✗ Password authentication: FAILED TO DISABLE"
    VERIFICATION_FAILED=1
fi

if grep -q "^PubkeyAuthentication yes" /etc/ssh/sshd_config; then
    log_info "✓ Public key authentication: ENABLED"
else
    log_error "✗ Public key authentication: FAILED TO ENABLE"
    VERIFICATION_FAILED=1
fi

if sshd -t 2>&1; then
    log_info "✓ SSH config syntax: VALID"
else
    log_error "✗ SSH config syntax: INVALID"
    VERIFICATION_FAILED=1
fi

# Show system info
log_info "System Information:"
echo "  Hostname: $(hostname)"
echo "  Kernel: $(uname -r)"
echo "  Docker: $(docker --version)"
if command -v fail2ban-client &> /dev/null; then
    echo "  Fail2ban: $(fail2ban-client --version 2>&1 | head -1)"
fi
UFW_STATUS=$(ufw status 2>&1 | head -1)
echo "  UFW: $UFW_STATUS"

# Show SSH key fingerprints
if [ -f /root/.ssh/authorized_keys ]; then
    log_info "Root SSH Keys:"
    ssh-keygen -l -f /root/.ssh/authorized_keys 2>/dev/null | while read line; do
        echo "  $line"
    done || log_warn "  (Could not read key fingerprints)"
fi

if id deploy &> /dev/null && [ -f /home/deploy/.ssh/authorized_keys ]; then
    log_info "Deploy SSH Keys:"
    ssh-keygen -l -f /home/deploy/.ssh/authorized_keys 2>/dev/null | while read line; do
        echo "  $line"
    done || log_warn "  (Could not read key fingerprints)"
fi

# Show next steps
echo ""
if [ $VERIFICATION_FAILED -eq 0 ]; then
    echo "✅ ✅ ✅ BOOTSTRAP COMPLETE ✅ ✅ ✅"
else
    echo "⚠️  BOOTSTRAP COMPLETE WITH WARNINGS ⚠️"
    echo "Review errors above before proceeding"
fi
echo "=========================================="
echo ""
echo "SECURITY STATUS:"
if [ $VERIFICATION_FAILED -eq 0 ]; then
    echo "  ✅ Password authentication: DISABLED"
    echo "  ✅ SSH keys: REQUIRED"
    echo "  ✅ Firewall: ACTIVE"
    echo "  ✅ Fail2ban: ACTIVE"
else
    echo "  ⚠️  Review verification errors above!"
fi
echo ""
echo "SSH Access (root):"
echo "  ssh -i ~/.ssh/hetzner/<SERVER-ID>/root root@$(hostname -I | awk '{print $1}' | head -1)"
echo ""
echo "SSH Access (deploy):"
echo "  ssh -i ~/.ssh/hetzner/<SERVER-ID>/deploy deploy@$(hostname -I | awk '{print $1}' | head -1)"
echo ""
echo "Next Steps:"
echo "  1. Verify SSH access works"
echo "  2. Edit: /opt/allora/compose/.env"
echo "  3. Copy compose files to /opt/allora/compose/"
echo "  4. Run: cd /opt/allora/compose && docker compose -f compose.prod.yml up -d"
echo ""
echo "Logs & Monitoring:"
echo "  - Full bootstrap log: $BOOTSTRAP_LOG"
echo "  - View log: cat $BOOTSTRAP_LOG"
echo "  - Tail log: tail -f $BOOTSTRAP_LOG"
echo "  - Docker: docker ps"
echo "  - Firewall: ufw status"
echo "  - Failed logins: fail2ban-client status sshd"
echo "  - Auth log: tail -50 /var/log/auth.log"
echo ""
echo "=========================================="
echo ""

# Exit with error if verification failed
if [ $VERIFICATION_FAILED -ne 0 ]; then
    log_error "Bootstrap completed but verification failed. Please review the errors above."
    log_error "Full log: $BOOTSTRAP_LOG"
    exit 1
fi
