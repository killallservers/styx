# ============================================================================
# REFERENCE ONLY: Outputs (commented out)
# ============================================================================
# These would be available if resources were active.
# To re-enable, uncomment and run: tofu apply
# ============================================================================

# output "server_id" {
#   value       = hcloud_server.allora.id
#   description = "Hetzner Cloud Server ID"
# }

# output "server_ipv4" {
#   value       = hcloud_server.allora.public_net[0].ipv4.ip
#   description = "Server public IPv4 address"
# }

# output "server_ipv6" {
#   value       = hcloud_server.allora.public_net[0].ipv6.ip
#   description = "Server public IPv6 address"
# }

# output "server_name" {
#   value       = hcloud_server.allora.name
#   description = "Server name"
# }

# output "internal_ip" {
#   value       = hcloud_server_network.allora.ip
#   description = "Server internal IP (in VPC)"
# }

# output "ssh_command" {
#   value       = "ssh root@${hcloud_server.allora.public_net[0].ipv4.ip}"
#   description = "SSH command to connect to server"
# }

# output "firewall_id" {
#   value       = hcloud_firewall.allora.id
#   description = "Firewall ID"
# }
