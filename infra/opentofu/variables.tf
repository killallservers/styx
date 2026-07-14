variable "hcloud_token" {
  description = "Hetzner Cloud API token"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

variable "server_name" {
  description = "Name of the dedicated server"
  type        = string
  default     = "allora-prod"
}

variable "server_image" {
  description = "OS image (e.g., 'ubuntu-26.04')"
  type        = string
  default     = "ubuntu-26.04"
}

variable "server_location" {
  description = "Hetzner location (e.g., 'nbg1' for Nuremberg)"
  type        = string
  default     = "nbg1"
}

variable "domain" {
  description = "Root domain for the application"
  type        = string
  default     = "allora.local"
}

variable "api_subdomain" {
  description = "API subdomain"
  type        = string
  default     = "api"
}

variable "web_subdomain" {
  description = "Web subdomain"
  type        = string
  default     = "www"
}

variable "ssh_public_key" {
  description = "SSH public key for root access"
  type        = string
}

variable "docker_registry" {
  description = "Docker registry (e.g., 'ghcr.io/username')"
  type        = string
}

variable "github_token" {
  description = "GitHub token for GHCR authentication"
  type        = string
  sensitive   = true
}
