#!/bin/bash
# Allora Server Bootstrap Script
# Run as root on fresh Debian 13 Hetzner server
# Usage: curl https://raw.githubusercontent.com/user/allora/main/infra/deployment/bootstrap.sh | bash

set -e

echo "=== Allora Server Bootstrap Starting ==="
echo "Environment: ${environment}"

# Update system
echo "Updating system packages..."
apt-get update
apt-get upgrade -y

# Install dependencies
echo "Installing dependencies..."
apt-get install -y \
  curl \
  wget \
  git \
  ca-certificates \
  gnupg \
  lsb-release \
  apt-transport-https

# Install Docker
echo "Installing Docker..."
curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Start Docker daemon
systemctl enable docker
systemctl start docker

# Install NVIDIA container toolkit (for GPU support)
echo "Installing NVIDIA container toolkit..."
curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg
curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | \
    sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | \
    tee /etc/apt/sources.list.d/nvidia-container-toolkit.list > /dev/null
apt-get update
apt-get install -y nvidia-container-toolkit

# Configure Docker to use NVIDIA runtime
mkdir -p /etc/docker
cat > /etc/docker/daemon.json <<EOF
{
  "runtimes": {
    "nvidia": {
      "path": "nvidia-container-runtime",
      "runtimeArgs": []
    }
  },
  "default-runtime": "runc",
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

systemctl restart docker

# Create application directories
echo "Setting up application directories..."
mkdir -p /opt/allora/{compose,deployments,backups,data}
cd /opt/allora

# Configure GHCR authentication
echo "Configuring Docker registry authentication..."
mkdir -p /root/.docker
cat > /root/.docker/config.json <<EOF
{
  "auths": {
    "ghcr.io": {
      "auth": "$(echo -n "USERNAME:${github_token}" | base64)"
    }
  }
}
EOF

# Create environment file
echo "Creating environment configuration..."
cat > /opt/allora/compose/.env <<EOF
# Allora Environment Configuration
ENVIRONMENT=${environment}
DOMAIN=\${DOMAIN:-allora.example.com}
DOCKER_REGISTRY=${docker_registry}
IMAGE_TAG=latest

# Database
POSTGRES_PASSWORD=$(openssl rand -base64 32)

# Grafana
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=$(openssl rand -base64 16)
EOF

chmod 600 /opt/allora/compose/.env

# Log file
echo "Setup complete at $(date)" > /var/log/allora-bootstrap.log
echo "Environment file created at /opt/allora/compose/.env"

echo ""
echo "=== Bootstrap Complete ==="
echo "Next steps:"
echo "1. Edit /opt/allora/compose/.env with actual configuration"
echo "2. Copy Docker Compose files to /opt/allora/compose/"
echo "3. Run: cd /opt/allora/compose && docker compose -f compose.prod.yml up -d"
echo ""
echo "SSH access: ssh root@<server-ip>"
