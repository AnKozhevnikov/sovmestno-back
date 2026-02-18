#!/bin/bash

# Sovmestno Backend - Server Setup Script
# This script prepares a fresh Ubuntu/Debian server for deployment

set -e

echo "Starting Sovmestno Backend server setup..."

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "This script should be run as root or with sudo"
  exit 1
fi

echo -e "${GREEN}[OK]${NC} Running as root"

# Update system
echo "Updating system packages..."
apt-get update
apt-get upgrade -y
echo -e "${GREEN}[OK]${NC} System updated"

# Install required packages
echo "Installing required packages..."
apt-get install -y \
  apt-transport-https \
  ca-certificates \
  curl \
  gnupg \
  lsb-release \
  git \
  ufw \
  fail2ban
echo -e "${GREEN}[OK]${NC} Required packages installed"

# Install Docker
echo "Installing Docker..."
if ! command -v docker &> /dev/null; then
  curl -fsSL https://get.docker.com | sh
  systemctl enable docker
  systemctl start docker
  echo -e "${GREEN}[OK]${NC} Docker installed"
else
  echo -e "${GREEN}[OK]${NC} Docker already installed"
fi

# Install Docker Compose (plugin)
echo "Installing Docker Compose..."
if ! docker compose version &> /dev/null; then
  apt-get install -y docker-compose-plugin
  echo -e "${GREEN}[OK]${NC} Docker Compose installed"
else
  echo -e "${GREEN}[OK]${NC} Docker Compose already installed"
fi

# Create deployment user
DEPLOY_USER="deploy"
echo "Setting up deployment user: ${DEPLOY_USER}"
if ! id "${DEPLOY_USER}" &>/dev/null; then
  useradd -m -s /bin/bash ${DEPLOY_USER}
  usermod -aG docker ${DEPLOY_USER}
  echo -e "${GREEN}[OK]${NC} User ${DEPLOY_USER} created"
else
  echo -e "${GREEN}[OK]${NC} User ${DEPLOY_USER} already exists"
fi

# Setup SSH for deploy user
echo "Setting up SSH for deploy user..."
DEPLOY_HOME="/home/${DEPLOY_USER}"
mkdir -p ${DEPLOY_HOME}/.ssh
chmod 700 ${DEPLOY_HOME}/.ssh
touch ${DEPLOY_HOME}/.ssh/authorized_keys
chmod 600 ${DEPLOY_HOME}/.ssh/authorized_keys
chown -R ${DEPLOY_USER}:${DEPLOY_USER} ${DEPLOY_HOME}/.ssh

echo -e "${YELLOW}[WARNING] Add your public SSH key to: ${DEPLOY_HOME}/.ssh/authorized_keys${NC}"
echo -e "${YELLOW}          Run: nano ${DEPLOY_HOME}/.ssh/authorized_keys${NC}"

# Setup firewall
echo "Configuring firewall..."
ufw --force enable
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
echo -e "${GREEN}[OK]${NC} Firewall configured"

# Configure fail2ban
echo "Configuring fail2ban..."
systemctl enable fail2ban
systemctl start fail2ban
echo -e "${GREEN}[OK]${NC} Fail2ban configured"

# Create project directory
echo "Creating project directory..."
PROJECT_DIR="${DEPLOY_HOME}/sovmestno-back"
if [ ! -d "$PROJECT_DIR" ]; then
  mkdir -p $PROJECT_DIR
  chown ${DEPLOY_USER}:${DEPLOY_USER} $PROJECT_DIR
  echo -e "${GREEN}[OK]${NC} Project directory created: $PROJECT_DIR"
else
  echo -e "${GREEN}[OK]${NC} Project directory already exists: $PROJECT_DIR"
fi

# Setup log rotation
echo "Setting up log rotation..."
cat > /etc/logrotate.d/sovmestno << 'EOF'
/home/deploy/sovmestno-back/logs/*.log {
  daily
  missingok
  rotate 14
  compress
  delaycompress
  notifempty
  create 0640 deploy deploy
  sharedscripts
}
EOF
echo -e "${GREEN}[OK]${NC} Log rotation configured"

# Docker optimization
echo "Optimizing Docker..."
cat > /etc/docker/daemon.json << 'EOF'
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "live-restore": true
}
EOF
systemctl restart docker
echo -e "${GREEN}[OK]${NC} Docker optimized"

# Print next steps
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Server setup complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo ""
echo "1. Add your SSH public key:"
echo "   sudo nano ${DEPLOY_HOME}/.ssh/authorized_keys"
echo ""
echo "2. Switch to deploy user:"
echo "   sudo su - ${DEPLOY_USER}"
echo ""
echo "3. Clone the repository:"
echo "   cd ~"
echo "   git clone https://github.com/YOUR_USERNAME/sovmestno-back.git"
echo "   cd sovmestno-back"
echo ""
echo "4. Create .env file from template:"
echo "   cp .env.staging.example .env"
echo "   nano .env"
echo ""
echo "5. Generate secrets:"
echo "   openssl rand -base64 32  # For JWT_SECRET"
echo "   openssl rand -base64 32  # For ADMIN_SECRET_KEY"
echo ""
echo "6. Setup SSL certificate (first time):"
echo "   ./setup-ssl.sh"
echo ""
echo "7. Deploy the application:"
echo "   docker compose -f docker-compose.staging.yml up -d"
echo ""
echo -e "${YELLOW}[WARNING] Don't forget to:${NC}"
echo "   - Update DNS records to point to this server"
echo "   - Configure GitHub Secrets for CI/CD"
echo "   - Setup monitoring and backups"
echo ""
