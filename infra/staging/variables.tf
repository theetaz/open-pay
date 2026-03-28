variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "ssh_key_name" {
  description = "Name of the SSH key registered in DigitalOcean"
  type        = string
}

variable "region" {
  description = "DigitalOcean region (e.g., sgp1, nyc1, lon1)"
  type        = string
  default     = "sgp1"
}

variable "droplet_size" {
  description = "Droplet size slug (s-2vcpu-4gb recommended minimum for 9 services)"
  type        = string
  default     = "s-2vcpu-4gb"
}

variable "postgres_password" {
  description = "PostgreSQL password for the olp user"
  type        = string
  sensitive   = true
}

variable "app_domain" {
  description = "Base domain for the staging environment (e.g., staging.yourdomain.com)"
  type        = string
  default     = ""
}

variable "ssh_allow_cidrs" {
  description = "CIDR blocks allowed to SSH into the droplet"
  type        = list(string)
  default     = ["0.0.0.0/0", "::/0"]
}

variable "create_reserved_ip" {
  description = "Whether to create a reserved (floating) IP for the staging droplet"
  type        = bool
  default     = false
}
