#!/bin/bash
# Initialize styx user environment
# Run as styx (non-root)
# Usage: bash init-styx-user.sh <github_repo>
# Example: bash init-styx-user.sh killallservers/styx

set -eu

GITHUB_REPO="${1:-killallservers/styx}"
STYX_DIR="/opt/styx"

echo "====================================="
echo "Styx User Initialization"
echo "====================================="
echo "GitHub Repo: $GITHUB_REPO"
echo "Styx Dir:    $STYX_DIR"
echo ""

# Initialize git repository
if [ ! -d "$STYX_DIR/.git" ]; then
    echo "Initializing git repository..."
    cd "$STYX_DIR"
    git init
    git remote add origin "https://github.com/$GITHUB_REPO"
    git fetch origin main
    git checkout -b main origin/main
    echo "✓ Git repository initialized"
else
    echo "✓ Git repository already initialized"
fi

# Create deploy script
echo "Creating deploy script..."
cat > "$STYX_DIR/deploy.sh" <<'EOFDEPLOY'
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

chmod +x "$STYX_DIR/deploy.sh"
echo "✓ Deploy script created at $STYX_DIR/deploy.sh"

# Add shell aliases to .bashrc
echo "Setting up shell aliases..."
if ! grep -q "alias styx-deploy" ~/.bashrc 2>/dev/null; then
    cat >> ~/.bashrc <<'EOF'

# Styx registry aliases
alias styx-deploy="$HOME/opt/styx/deploy.sh"
alias styx-logs="docker compose -f $STYX_DIR/compose.prod.yml logs -f"
alias styx-status="docker compose -f $STYX_DIR/compose.prod.yml ps"
EOF
    echo "✓ Aliases added to ~/.bashrc"
else
    echo "✓ Aliases already in ~/.bashrc"
fi

echo ""
echo "====================================="
echo "✅ Styx user initialization complete!"
echo "====================================="
echo ""
echo "Next steps:"
echo "  1. Source your shell profile: source ~/.bashrc"
echo "  2. Deploy: $STYX_DIR/deploy.sh"
echo "  3. Check status: docker compose -f $STYX_DIR/compose.prod.yml ps"
echo ""
