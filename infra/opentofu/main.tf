# ============================================================================
# REFERENCE: Infrastructure as Code for Hetzner Dedicated Server
# ============================================================================
#
# ⚠️  IMPORTANT: This configuration is for REFERENCE and drift detection only.
#
# Your server is already provisioned. Do NOT run:
#   - tofu apply
#   - tofu plan
#   - Any OpenTofu commands that would create/modify resources
#
# This configuration documents what the infra looks like.
# Use it to:
#   1. Understand the server setup
#   2. Detect drift (if infra changes outside of IaC)
#   3. Rebuild from scratch (if needed in future)
#
# For now: use bootstrap-secure.sh to configure the existing server.
# See: infra/DEPLOY_TO_EXISTING_SERVER.md
# ============================================================================

# ============================================================================
# REFERENCE ONLY: Resource Definitions (commented out)
# ============================================================================

# SSH Key for server access
# resource "hcloud_ssh_key" "allora" {
#   name       = "allora-${var.environment}"
#   public_key = var.ssh_public_key
# }

# Network for internal communication (VPC)
# resource "hcloud_network" "allora" {
#   name = "allora-${var.environment}"
# }

# resource "hcloud_network_subnet" "allora" {
#   network_id   = hcloud_network.allora.id
#   type         = "cloud"
#   network_zone = "eu-central"
#   ip_range     = "10.0.0.0/8"
# }

# Firewall rules (SSH, HTTP, HTTPS, internal services)
# resource "hcloud_firewall" "allora" {
#   name = "allora-${var.environment}"
#
#   rule {
#     direction = "in"
#     protocol  = "tcp"
#     port      = "22"
#     source_ips = ["0.0.0.0/0", "::/0"]
#   }
#
#   rule {
#     direction = "in"
#     protocol  = "tcp"
#     port      = "80"
#     source_ips = ["0.0.0.0/0", "::/0"]
#   }
#
#   rule {
#     direction = "in"
#     protocol  = "tcp"
#     port      = "443"
#     source_ips = ["0.0.0.0/0", "::/0"]
#   }
# }

# Dedicated Server
# resource "hcloud_server" "allora" {
#   name       = var.server_name
#   image      = var.server_image
#   server_type = "cx102"
#   location   = var.server_location
#   ssh_keys   = [hcloud_ssh_key.allora.id]
#   public_net {
#     ipv4_enabled = true
#     ipv6_enabled = true
#   }
#
#   labels = {
#     environment = var.environment
#     application = "allora"
#   }
# }

# Network attachment
# resource "hcloud_server_network" "allora" {
#   server_id  = hcloud_server.allora.id
#   network_id = hcloud_network.allora.id
#   ip         = "10.0.1.10"
# }

# Firewall attachment
# resource "hcloud_firewall_attachment" "allora" {
#   firewall_id = hcloud_firewall.allora.id
#   server_ids  = [hcloud_server.allora.id]
# }

# Static IP assignment
# resource "hcloud_primary_ip" "allora" {
#   name              = "allora-${var.environment}-ip"
#   type              = "ipv4"
#   datacenter        = var.server_location
#   auto_delete       = false
#   delete_protection = true
#   labels = {
#     environment = var.environment
#   }
# }

# resource "hcloud_server_primary_ip_assignment" "allora" {
#   primary_ip_id = hcloud_primary_ip.allora.id
#   server_id     = hcloud_server.allora.id
# }
