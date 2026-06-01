terraform {
  required_version = ">= 1.5"

  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "~> 2.34"
    }
  }
}

provider "digitalocean" {
  token = var.do_token
}

# --- SSH Key ---
data "digitalocean_ssh_key" "deploy" {
  name = var.ssh_key_name
}

# --- Staging Droplet ---
resource "digitalocean_droplet" "staging" {
  name     = "openpay-staging"
  image    = "ubuntu-24-04-x64"
  size     = var.droplet_size
  region   = var.region
  ssh_keys = [data.digitalocean_ssh_key.deploy.id]

  user_data = templatefile("${path.module}/cloud-init.yml", {
    postgres_password = var.postgres_password
    app_domain        = var.app_domain
  })

  tags = ["openpay", "staging"]
}

# --- Firewall ---
resource "digitalocean_firewall" "staging" {
  name        = "openpay-staging-fw"
  droplet_ids = [digitalocean_droplet.staging.id]

  # SSH
  inbound_rule {
    protocol         = "tcp"
    port_range       = "22"
    source_addresses = var.ssh_allow_cidrs
  }

  # HTTP / HTTPS (for API gateway behind reverse proxy)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "80"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  inbound_rule {
    protocol         = "tcp"
    port_range       = "443"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # Gateway API (direct access, useful before reverse proxy is set up)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "8080"
    source_addresses = ["0.0.0.0/0", "::/0"]
  }

  # Mailpit UI (staging only — email testing)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "8025"
    source_addresses = var.ssh_allow_cidrs
  }

  # MinIO console (staging only)
  inbound_rule {
    protocol         = "tcp"
    port_range       = "9001"
    source_addresses = var.ssh_allow_cidrs
  }

  # Allow all outbound
  outbound_rule {
    protocol              = "tcp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "udp"
    port_range            = "1-65535"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }

  outbound_rule {
    protocol              = "icmp"
    destination_addresses = ["0.0.0.0/0", "::/0"]
  }
}

# --- Optional: Floating IP for stable staging address ---
resource "digitalocean_reserved_ip" "staging" {
  count  = var.create_reserved_ip ? 1 : 0
  region = var.region
}

resource "digitalocean_reserved_ip_assignment" "staging" {
  count      = var.create_reserved_ip ? 1 : 0
  ip_address = digitalocean_reserved_ip.staging[0].ip_address
  droplet_id = digitalocean_droplet.staging.id
}
