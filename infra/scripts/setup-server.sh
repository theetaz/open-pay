#!/bin/bash
set -e

# =============================================================================
# Open Pay — Ubuntu VPS Setup Script
# Run this on a fresh Ubuntu 22.04/24.04 VPS to prepare it for deployments.
# Usage: curl -sSL <raw-url> | bash -s -- [staging|production]
#   or:  ./setup-server.sh [staging|production]
# =============================================================================

ENVIRONMENT="${1:-staging}"
REPO_URL="${2:-}"

echo "============================================"
echo "  Open Pay — Server Setup ($ENVIRONMENT)"
echo "============================================"

# --- System update ---
echo "==> Updating system packages..."
apt-get update -qq
apt-get upgrade -y -qq

# --- Essential packages ---
echo "==> Installing dependencies..."
apt-get install -y -qq \
  apt-transport-https \
  ca-certificates \
  curl \
  gnupg \
  lsb-release \
  ufw \
  fail2ban \
  unzip \
  wget \
  git \
  jq

# --- Docker ---
if ! command -v docker &>/dev/null; then
  echo "==> Installing Docker..."
  curl -fsSL https://get.docker.com | sh
  systemctl enable docker
  systemctl start docker
  echo "Docker installed: $(docker --version)"
else
  echo "==> Docker already installed: $(docker --version)"
fi

# --- Docker Compose (v2 plugin) ---
echo "==> Docker Compose: $(docker compose version)"

# --- golang-migrate ---
if ! command -v migrate &>/dev/null; then
  echo "==> Installing golang-migrate..."
  curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.deb -o /tmp/migrate.deb
  dpkg -i /tmp/migrate.deb
  rm /tmp/migrate.deb
  echo "migrate installed: $(migrate --version 2>&1 || true)"
else
  echo "==> migrate already installed: $(migrate --version 2>&1 || true)"
fi

# --- Firewall ---
echo "==> Configuring firewall..."
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw allow 8080/tcp  # Gateway API
if [ "$ENVIRONMENT" = "staging" ]; then
  ufw allow 8025/tcp  # Mailpit UI
  ufw allow 9001/tcp  # MinIO Console
fi
ufw --force enable
echo "Firewall configured."

# --- Fail2ban ---
echo "==> Configuring fail2ban..."
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = 22
filter = sshd
logpath = /var/log/auth.log
EOF
systemctl enable fail2ban
systemctl restart fail2ban

# --- Docker log rotation ---
echo "==> Configuring Docker log rotation..."
mkdir -p /etc/docker
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF
systemctl restart docker

# --- Swap (for small VPS during Docker builds) ---
if [ ! -f /swapfile ]; then
  echo "==> Creating 2GB swap..."
  fallocate -l 2G /swapfile
  chmod 600 /swapfile
  mkswap /swapfile
  swapon /swapfile
  echo '/swapfile none swap sw 0 0' >> /etc/fstab
  echo 'vm.swappiness=10' >> /etc/sysctl.conf
  sysctl -p
fi

# --- App directory ---
mkdir -p /opt/openpay

# --- Clone repo if URL provided ---
if [ -n "$REPO_URL" ]; then
  if [ ! -d /opt/openpay/.git ]; then
    echo "==> Cloning repository..."
    git clone "$REPO_URL" /opt/openpay
  else
    echo "==> Repository already cloned."
  fi
fi

# --- Done ---
echo ""
echo "============================================"
echo "  Server setup complete!"
echo "============================================"
echo ""
echo "Next steps:"
echo "  1. Clone the repo (if not done): git clone <repo-url> /opt/openpay"
echo "  2. Copy env template:  cp /opt/openpay/.env.${ENVIRONMENT}.example /opt/openpay/.env.${ENVIRONMENT}"
if [ "$ENVIRONMENT" = "staging" ]; then
  echo "  3. Edit /opt/openpay/.env.staging with real credentials"
  echo "  4. Deploy: cd /opt/openpay && ./deploy.sh staging"
  echo "  5. Add DROPLET_IP and DROPLET_SSH_KEY to GitHub Environment 'staging' secrets"
else
  echo "  3. Edit /opt/openpay/.env.prod with real credentials"
  echo "  4. Deploy: cd /opt/openpay && ./deploy.sh production"
  echo "  5. Add DROPLET_IP and DROPLET_SSH_KEY to GitHub Environment 'production' secrets"
fi
echo ""
