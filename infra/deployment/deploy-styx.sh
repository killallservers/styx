#!/bin/bash
# Styx Registry Deployment Script
# Called by GitHub Actions CI/CD to deploy to production
# Usage: bash deploy-styx.sh <ssh_host> <ssh_user> <registry_dir>

set -eu

SSH_HOST="${1}"
SSH_USER="${2}"
REGISTRY_DIR="${3:-/opt/styx}"

if [ -z "$SSH_HOST" ] || [ -z "$SSH_USER" ]; then
    echo "Usage: $0 <ssh_host> <ssh_user> [registry_dir]"
    echo ""
    echo "Example: $0 1.2.3.4 styx /opt/styx"
    exit 1
fi

echo "=== Styx Registry Deployment ==="
echo "Host:     $SSH_HOST"
echo "User:     $SSH_USER"
echo "Dir:      $REGISTRY_DIR"
echo ""

# Deploy via SSH
echo "Connecting to server and deploying..."
ssh -o StrictHostKeyChecking=no "$SSH_USER@$SSH_HOST" bash <<'EOFSCRIPT'
    set -eu

    cd /opt/styx

    echo "Pulling latest code..."
    git pull origin main

    echo "Building registry image..."
    docker compose -f compose.prod.yml build --no-cache registry

    echo "Starting services..."
    docker compose -f compose.prod.yml up -d

    echo "Waiting for services to be healthy..."
    sleep 2

    echo "Verifying health check..."
    if curl -f http://localhost:7506/health >/dev/null 2>&1; then
        echo "✓ Registry is healthy"
        exit 0
    else
        echo "✗ Registry health check failed"
        docker compose -f compose.prod.yml logs registry
        exit 1
    fi
EOFSCRIPT

echo ""
echo "✅ Deployment complete!"
echo ""
echo "Registry is now live at: https://$REGISTRY_DOMAIN/health"
