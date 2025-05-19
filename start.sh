#!/bin/bash
set -e

# Initialize environment variables if not set
export DOMAIN=${DOMAIN:-localhost}
export PUID=${PUID:-$(id -u)}
export PGID=${PGID:-$(id -g)}
export TZ=${TZ:-UTC}

# Create required directories if they don't exist
mkdir -p static/svg
mkdir -p svg-cache

# Check if the basic-auth.svg exists, if not copy from original location
if [ ! -f static/svg/basic-auth.svg ]; then
  echo "Copying default SVG template..."
  cp svg/basic-auth.svg static/svg/ || echo "Warning: Could not copy default SVG template"
fi

# Build and start the containers
echo "Starting SVG Web Elements service..."
docker compose up -d

# Show status
echo "Service started! Container status:"
docker compose ps

echo ""
echo "SVG Web Elements service is now available at: https://svg.${DOMAIN}"
echo "Or locally at: http://localhost:8082"
echo ""
echo "To monitor logs: docker compose logs -f"
echo "To stop service: docker compose down"