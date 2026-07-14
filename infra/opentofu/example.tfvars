# Copy to terraform.tfvars and fill in actual values

# Hetzner Cloud API token (from dashboard)
hcloud_token = "your-hcloud-token-here"

# Environment name
environment = "prod"

# Server configuration
server_name     = "allora-prod"
server_image    = "ubuntu-26.04"
server_location = "nbg1"  # Nuremberg (eu-central)

# Domain configuration
domain           = "allora.example.com"
api_subdomain    = "api"
web_subdomain    = "www"

# SSH public key (cat ~/.ssh/id_rsa.pub)
ssh_public_key = "ssh-rsa AAAA..."

# Docker registry (ghcr.io/your-github-username)
docker_registry = "ghcr.io/erikwright"

# GitHub token for GHCR (personal access token with 'write:packages' scope)
github_token = "your-github-token-here"
