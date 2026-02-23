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

# Check if envsubst is installed
if ! command -v envsubst &> /dev/null; then
  echo -e "${RED}[ERROR] envsubst not found!${NC}"
  echo "Please install gettext-base package:"
  echo "  sudo apt-get update && sudo apt-get install -y gettext-base"
  exit 1
fi

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

if [ -z "$FRONTEND_DOMAIN" ]; then
  echo -e "${RED}[ERROR] FRONTEND_DOMAIN not set in .env file!${NC}"
  echo "Add FRONTEND_DOMAIN to your .env file"
  echo "Example: FRONTEND_DOMAIN=sovmestno-test.ru"
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
echo -e "${GREEN}API Domain: ${DOMAIN}${NC}"
echo -e "${GREEN}Frontend Domain: ${FRONTEND_DOMAIN}${NC}"
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
echo "[1/8] Creating temporary nginx configuration for ACME challenge..."

# Remove nginx config if it's a directory (Docker creates it if file doesn't exist)
if [ -d "$OUTPUT" ]; then
  echo "Removing directory created by Docker: ${OUTPUT}"
  rm -rf "$OUTPUT"
fi

TEMP_CONF=$(mktemp)
cat > $TEMP_CONF << 'EOF'
events {
    worker_connections 1024;
}

http {
    # API subdomain
    server {
        listen 80;
        server_name ${DOMAIN};

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            return 200 'API server is running - waiting for SSL setup';
            add_header Content-Type text/plain;
        }
    }

    # Frontend main domain
    server {
        listen 80;
        server_name ${FRONTEND_DOMAIN} www.${FRONTEND_DOMAIN};

        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }

        location / {
            return 200 'Frontend server is running - waiting for SSL setup';
            add_header Content-Type text/plain;
        }
    }
}
EOF

# Substitute DOMAIN and FRONTEND_DOMAIN in temp config
envsubst '${DOMAIN} ${FRONTEND_DOMAIN}' < $TEMP_CONF > ${TEMP_CONF}.processed
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
echo "[2/8] Starting nginx and certbot containers..."
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

# Step 3: Check if certificates already exist
echo ""
echo "[3/8] Checking for existing certificates..."

# Check API certificate
API_CERT_EXISTS=false
if docker compose -f ${COMPOSE_FILE} exec nginx test -d /etc/letsencrypt/live/${DOMAIN} 2>/dev/null; then
  echo -e "${YELLOW}[INFO] Certificate for API domain (${DOMAIN}) already exists${NC}"
  API_CERT_EXISTS=true
else
  echo "No certificate found for API domain (${DOMAIN})"
fi

# Check Frontend certificate
FRONTEND_CERT_EXISTS=false
if docker compose -f ${COMPOSE_FILE} exec nginx test -d /etc/letsencrypt/live/${FRONTEND_DOMAIN} 2>/dev/null; then
  echo -e "${YELLOW}[INFO] Certificate for frontend domain (${FRONTEND_DOMAIN}) already exists${NC}"
  FRONTEND_CERT_EXISTS=true
else
  echo "No certificate found for frontend domain (${FRONTEND_DOMAIN})"
fi

# Ask if user wants to proceed
if [ "$API_CERT_EXISTS" = true ] || [ "$FRONTEND_CERT_EXISTS" = true ]; then
  echo ""
  read -p "Request new certificates? This will renew existing ones. (y/N) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Using existing certificates"
    SKIP_CERT_REQUEST=true
  else
    SKIP_CERT_REQUEST=false
  fi
else
  SKIP_CERT_REQUEST=false
fi

# Step 4: Request certificates
if [ "$SKIP_CERT_REQUEST" = false ]; then
  echo ""
  echo "[4/8] Requesting SSL certificates from Let's Encrypt..."
  echo "This may take a couple of minutes..."
  echo ""

  # Request API certificate
  echo "Requesting certificate for API domain: ${DOMAIN}"
  if [ "$API_CERT_EXISTS" = true ]; then
    # Renew existing (only if expiring in <30 days)
    docker compose -f ${COMPOSE_FILE} run --rm certbot certonly \
      --webroot \
      --webroot-path=/var/www/certbot \
      --email ${SSL_EMAIL} \
      --agree-tos \
      --no-eff-email \
      --keep-until-expiring \
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

  echo ""
  echo "Requesting certificate for frontend domain: ${FRONTEND_DOMAIN}"
  if [ "$FRONTEND_CERT_EXISTS" = true ]; then
    # Renew existing (only if expiring in <30 days)
    docker compose -f ${COMPOSE_FILE} run --rm certbot certonly \
      --webroot \
      --webroot-path=/var/www/certbot \
      --email ${SSL_EMAIL} \
      --agree-tos \
      --no-eff-email \
      --keep-until-expiring \
      -d ${FRONTEND_DOMAIN} \
      -d www.${FRONTEND_DOMAIN}
  else
    # Request new
    docker compose -f ${COMPOSE_FILE} run --rm certbot certonly \
      --webroot \
      --webroot-path=/var/www/certbot \
      --email ${SSL_EMAIL} \
      --agree-tos \
      --no-eff-email \
      -d ${FRONTEND_DOMAIN} \
      -d www.${FRONTEND_DOMAIN}
  fi
else
  echo ""
  echo "[4/8] Skipped certificate request"
fi

# Step 5: Verify certificates exist
echo ""
echo "[5/8] Verifying SSL certificates..."

# Verify API certificate
if docker compose -f ${COMPOSE_FILE} exec nginx test -f /etc/letsencrypt/live/${DOMAIN}/fullchain.pem 2>/dev/null; then
  echo -e "${GREEN}[OK] API certificate found: ${DOMAIN}${NC}"
else
  echo -e "${RED}[ERROR] API certificate not found!${NC}"
  echo "Certificate should be at: /etc/letsencrypt/live/${DOMAIN}/fullchain.pem"

  # Restore backup if exists
  LATEST_BACKUP=$(ls -t ${OUTPUT}.backup.* 2>/dev/null | head -1)
  if [ ! -z "$LATEST_BACKUP" ]; then
    mv "$LATEST_BACKUP" "$OUTPUT"
    echo "Restored previous configuration from ${LATEST_BACKUP}"
  fi

  exit 1
fi

# Verify Frontend certificate
if docker compose -f ${COMPOSE_FILE} exec nginx test -f /etc/letsencrypt/live/${FRONTEND_DOMAIN}/fullchain.pem 2>/dev/null; then
  echo -e "${GREEN}[OK] Frontend certificate found: ${FRONTEND_DOMAIN}${NC}"
else
  echo -e "${RED}[ERROR] Frontend certificate not found!${NC}"
  echo "Certificate should be at: /etc/letsencrypt/live/${FRONTEND_DOMAIN}/fullchain.pem"

  # Restore backup if exists
  LATEST_BACKUP=$(ls -t ${OUTPUT}.backup.* 2>/dev/null | head -1)
  if [ ! -z "$LATEST_BACKUP" ]; then
    mv "$LATEST_BACKUP" "$OUTPUT"
    echo "Restored previous configuration from ${LATEST_BACKUP}"
  fi

  exit 1
fi

# Step 6: Generate final nginx config from template
echo ""
echo "[6/8] Generating final nginx configuration from template..."
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

# Step 7: Verify frontend directory exists
echo ""
echo "[7/8] Verifying frontend deployment directory..."
if [ -d /var/www/sovmestno-frontend ]; then
  echo -e "${GREEN}[OK] Frontend directory exists: /var/www/sovmestno-frontend${NC}"
else
  echo -e "${YELLOW}[WARNING] Frontend directory does not exist!${NC}"
  echo "Creating /var/www/sovmestno-frontend..."
  mkdir -p /var/www/sovmestno-frontend
  chown $(whoami):$(whoami) /var/www/sovmestno-frontend
  echo -e "${GREEN}[OK] Frontend directory created${NC}"
fi

# Step 8: Reload nginx
echo ""
echo "[8/8] Reloading nginx with new configuration..."
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
echo "Your services are now available at:"
echo -e "${GREEN}API: https://${DOMAIN}${NC}"
echo -e "${GREEN}Frontend: https://${FRONTEND_DOMAIN}${NC}"
echo -e "${GREEN}Frontend (www): https://www.${FRONTEND_DOMAIN}${NC}"
echo ""
echo "Test your services:"
echo "  curl https://${DOMAIN}/health"
echo "  curl https://${FRONTEND_DOMAIN}"
echo ""
echo "Configuration details:"
echo "  Environment: ${ENV_NAME}"
echo "  Template: ${TEMPLATE}"
echo "  Generated config: ${OUTPUT}"
echo "  API Domain: ${DOMAIN}"
echo "  Frontend Domain: ${FRONTEND_DOMAIN}"
echo "  API Certificate: /etc/letsencrypt/live/${DOMAIN}/fullchain.pem"
echo "  Frontend Certificate: /etc/letsencrypt/live/${FRONTEND_DOMAIN}/fullchain.pem"
echo ""
echo "Next steps:"
echo "1. Set up GitHub Secrets in sovmestno-front repository"
echo "2. Push frontend code to main branch"
echo ""
echo "Certificates will auto-renew via certbot container"
echo "Certbot runs twice daily and renews certificates expiring in <30 days"
echo ""
