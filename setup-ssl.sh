#!/bin/bash

# SSL Certificate Setup Script using Let's Encrypt
# Run this script ONCE on the server after initial deployment
#
# This script:
# 1. Validates environment variables (DOMAIN, SSL_EMAIL)
# 2. Checks DNS configuration
# 3. Creates temporary nginx config for ACME challenge
# 4. Requests SSL certificate from Let's Encrypt
# 5. Generates final nginx config from template using envsubst
# 6. Tests and reloads nginx

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo "Starting SSL certificate setup..."

# Load environment variables
if [ ! -f .env ]; then
  echo -e "${RED}[ERROR] .env file not found!${NC}"
  echo "Please create .env file first (copy from .env.staging.example or .env.production.example)"
  exit 1
fi

# Source .env file
export $(grep -v '^#' .env | xargs)

# Validate required variables
if [ -z "$DOMAIN" ]; then
  echo -e "${RED}[ERROR] DOMAIN not set in .env file!${NC}"
  exit 1
fi

if [ -z "$SSL_EMAIL" ]; then
  echo -e "${RED}[ERROR] SSL_EMAIL not set in .env file!${NC}"
  exit 1
fi

if [ -z "$ENVIRONMENT" ]; then
  echo -e "${RED}[ERROR] ENVIRONMENT not set in .env file!${NC}"
  echo "ENVIRONMENT should be 'staging' or 'production'"
  exit 1
fi

echo -e "${GREEN}Environment: ${ENVIRONMENT}${NC}"
echo -e "${GREEN}Domain: ${DOMAIN}${NC}"
echo -e "${GREEN}Email: ${SSL_EMAIL}${NC}"

# Check if domain resolves to this server
echo "Checking DNS configuration..."
SERVER_IP=$(curl -s ifconfig.me)
DOMAIN_IP=$(dig +short ${DOMAIN} | tail -n1)

echo "Server IP: ${SERVER_IP}"
echo "Domain IP: ${DOMAIN_IP}"

if [ "$SERVER_IP" != "$DOMAIN_IP" ]; then
  echo -e "${YELLOW}[WARNING] Domain ${DOMAIN} does not resolve to this server IP${NC}"
  echo "Make sure DNS A record is configured:"
  echo "  Type: A"
  echo "  Name: api (for subdomain) or @ (for root domain)"
  echo "  Value: ${SERVER_IP}"
  echo ""
  read -p "Continue anyway? (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# Determine which environment we're in based on ENVIRONMENT variable
if [ "$ENVIRONMENT" = "staging" ]; then
  COMPOSE_FILE="docker-compose.staging.yml"
  TEMPLATE="nginx/nginx.staging.conf.template"
  OUTPUT="nginx/nginx.staging.conf"
  ENV_NAME="staging"
elif [ "$ENVIRONMENT" = "production" ]; then
  COMPOSE_FILE="docker-compose.production.yml"
  TEMPLATE="nginx/nginx.production.conf.template"
  OUTPUT="nginx/nginx.production.conf"
  ENV_NAME="production"
else
  echo -e "${RED}[ERROR] Invalid ENVIRONMENT value: ${ENVIRONMENT}${NC}"
  echo "ENVIRONMENT should be 'staging' or 'production'"
  exit 1
fi

# Check if compose file exists
if [ ! -f "$COMPOSE_FILE" ]; then
  echo -e "${RED}[ERROR] Compose file not found: ${COMPOSE_FILE}${NC}"
  exit 1
fi

echo ""
echo "Environment detected: ${ENV_NAME}"
echo "Compose file: ${COMPOSE_FILE}"
echo "Template: ${TEMPLATE}"
echo "Output: ${OUTPUT}"
echo ""

# Check if template exists
if [ ! -f "$TEMPLATE" ]; then
  echo -e "${RED}[ERROR] Template file not found: ${TEMPLATE}${NC}"
  exit 1
fi

# Step 1: Create temporary nginx config for ACME challenge
echo "[1/7] Creating temporary nginx configuration for ACME challenge..."
TEMP_CONF=$(mktemp)
cat > $TEMP_CONF << 'EOF'
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        server_name ${DOMAIN};

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            return 200 'Server is running - waiting for SSL setup';
            add_header Content-Type text/plain;
        }
    }
}
EOF

# Substitute DOMAIN in temp config
envsubst '${DOMAIN}' < $TEMP_CONF > ${TEMP_CONF}.processed
mv ${TEMP_CONF}.processed $TEMP_CONF

# Backup existing config if present
if [ -f "$OUTPUT" ]; then
  BACKUP_FILE="${OUTPUT}.backup.$(date +%Y%m%d_%H%M%S)"
  cp "$OUTPUT" "$BACKUP_FILE"
  echo "Backed up existing config to ${BACKUP_FILE}"
fi

# Copy temp config
cp $TEMP_CONF $OUTPUT
rm $TEMP_CONF
echo "Temporary config created"

# Step 2: Start nginx and certbot
echo ""
echo "[2/7] Starting nginx and certbot containers..."
docker compose -f ${COMPOSE_FILE} up -d nginx certbot

echo "Waiting for nginx to start..."
sleep 5

# Verify nginx is running
if ! docker compose -f ${COMPOSE_FILE} ps nginx | grep -q "Up"; then
  echo -e "${RED}[ERROR] Nginx failed to start!${NC}"
  docker compose -f ${COMPOSE_FILE} logs nginx
  exit 1
fi
echo "Nginx is running"

# Step 3: Check if certificate already exists
echo ""
echo "[3/7] Checking for existing certificates..."
if docker compose -f ${COMPOSE_FILE} exec nginx test -d /etc/letsencrypt/live/${DOMAIN} 2>/dev/null; then
  echo -e "${YELLOW}[WARNING] Certificate for ${DOMAIN} already exists${NC}"
  read -p "Request a new certificate? This will renew the existing one. (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Skipping certificate request, using existing certificate"
    SKIP_CERT_REQUEST=true
  else
    SKIP_CERT_REQUEST=false
  fi
else
  echo "No existing certificate found"
  SKIP_CERT_REQUEST=false
fi

# Step 4: Request certificate
if [ "$SKIP_CERT_REQUEST" = false ]; then
  echo ""
  echo "[4/7] Requesting SSL certificate from Let's Encrypt..."
  echo "This may take a minute..."

  if docker compose -f ${COMPOSE_FILE} exec nginx test -d /etc/letsencrypt/live/${DOMAIN} 2>/dev/null; then
    # Renew existing
    docker compose -f ${COMPOSE_FILE} run --rm certbot certonly \
      --webroot \
      --webroot-path=/var/www/certbot \
      --email ${SSL_EMAIL} \
      --agree-tos \
      --no-eff-email \
      --force-renewal \
      -d ${DOMAIN}
  else
    # Request new
    docker compose -f ${COMPOSE_FILE} run --rm certbot certonly \
      --webroot \
      --webroot-path=/var/www/certbot \
      --email ${SSL_EMAIL} \
      --agree-tos \
      --no-eff-email \
      -d ${DOMAIN}
  fi
else
  echo ""
  echo "[4/7] Skipped certificate request"
fi

# Step 5: Verify certificate exists
echo ""
echo "[5/7] Verifying SSL certificate..."
if docker compose -f ${COMPOSE_FILE} exec nginx test -f /etc/letsencrypt/live/${DOMAIN}/fullchain.pem 2>/dev/null; then
  echo -e "${GREEN}[OK] SSL certificate found${NC}"
else
  echo -e "${RED}[ERROR] SSL certificate not found!${NC}"
  echo "Certificate should be at: /etc/letsencrypt/live/${DOMAIN}/fullchain.pem"

  # Restore backup if exists
  if [ -f "${OUTPUT}.backup" ]; then
    mv "${OUTPUT}.backup" "$OUTPUT"
    echo "Restored previous configuration"
  fi

  exit 1
fi

# Step 6: Generate final nginx config from template
echo ""
echo "[6/7] Generating final nginx configuration from template..."
envsubst '${DOMAIN}' < ${TEMPLATE} > ${OUTPUT}
echo "Generated ${OUTPUT} from ${TEMPLATE}"

# Test nginx config
echo "Testing nginx configuration..."
if docker compose -f ${COMPOSE_FILE} exec nginx nginx -t 2>&1 | grep -q "successful"; then
  echo -e "${GREEN}[OK] Nginx configuration is valid${NC}"
else
  echo -e "${RED}[ERROR] Nginx configuration test failed!${NC}"
  docker compose -f ${COMPOSE_FILE} exec nginx nginx -t

  # Restore backup if exists
  LATEST_BACKUP=$(ls -t ${OUTPUT}.backup.* 2>/dev/null | head -1)
  if [ ! -z "$LATEST_BACKUP" ]; then
    mv "$LATEST_BACKUP" "$OUTPUT"
    echo "Restored previous configuration from ${LATEST_BACKUP}"
  fi

  exit 1
fi

# Step 7: Reload nginx
echo ""
echo "[7/7] Reloading nginx with new configuration..."
docker compose -f ${COMPOSE_FILE} exec nginx nginx -s reload
echo "Nginx reloaded successfully"

# Cleanup old backups (keep last 3)
echo ""
echo "Cleaning up old backups..."
ls -t ${OUTPUT}.backup.* 2>/dev/null | tail -n +4 | xargs rm -f 2>/dev/null || true

# Success message
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}SSL setup complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Your API is now available at:"
echo -e "${GREEN}https://${DOMAIN}${NC}"
echo ""
echo "Test your API:"
echo "  curl https://${DOMAIN}/health"
echo ""
echo "Configuration details:"
echo "  Environment: ${ENV_NAME}"
echo "  Template: ${TEMPLATE}"
echo "  Generated config: ${OUTPUT}"
echo "  Domain: ${DOMAIN}"
echo "  Certificate: /etc/letsencrypt/live/${DOMAIN}/fullchain.pem"
echo ""
echo "Certificate will auto-renew via certbot container"
echo "Certbot runs twice daily and renews certificates expiring in <30 days"
echo ""
