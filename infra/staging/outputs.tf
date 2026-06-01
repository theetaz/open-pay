output "droplet_ip" {
  description = "Public IPv4 address of the staging droplet"
  value       = digitalocean_droplet.staging.ipv4_address
}

output "droplet_id" {
  description = "DigitalOcean droplet ID"
  value       = digitalocean_droplet.staging.id
}

output "reserved_ip" {
  description = "Reserved IP address (if created)"
  value       = var.create_reserved_ip ? digitalocean_reserved_ip.staging[0].ip_address : null
}

output "ssh_command" {
  description = "SSH command to connect to the staging server"
  value       = "ssh root@${digitalocean_droplet.staging.ipv4_address}"
}

output "next_steps" {
  description = "Post-provisioning steps"
  value       = <<-EOT
    1. SSH into the droplet: ssh root@${digitalocean_droplet.staging.ipv4_address}
    2. Wait for cloud-init to finish: cloud-init status --wait
    3. Clone the repo: git clone <repo-url> /opt/openpay
    4. Copy env template: cp /opt/openpay/.env.staging.example /opt/openpay/.env.staging
    5. Edit .env.staging with real credentials
    6. Run first deploy: cd /opt/openpay && ./deploy.sh staging
    7. Add DROPLET_IP and DROPLET_SSH_KEY to GitHub Environment "staging" secrets
  EOT
}
