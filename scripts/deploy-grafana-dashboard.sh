#!/bin/bash
set -e

# Deploy Grafana Dashboard to Fly.io Grafana
# This script uses Grafana's HTTP API to programmatically deploy dashboards

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DASHBOARD_FILE="$PROJECT_ROOT/docs/grafana-dashboard.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Grafana Dashboard Deployment Script"
echo "========================================="
echo ""

# Check if dashboard file exists
if [ ! -f "$DASHBOARD_FILE" ]; then
    echo -e "${RED}Error: Dashboard file not found: $DASHBOARD_FILE${NC}"
    exit 1
fi

# Check for required environment variables
if [ -z "$GRAFANA_URL" ]; then
    echo -e "${YELLOW}GRAFANA_URL not set. Using Fly.io default...${NC}"
    # Fly.io Grafana URL format (you'll need to update this based on your org)
    # GRAFANA_URL="https://fly-metrics.net" # Example
    echo ""
    echo -e "${YELLOW}To deploy to Fly.io Grafana:${NC}"
    echo "1. Get your Grafana URL from: https://fly.io/dashboard"
    echo "2. Create a Grafana API key: Settings → API Keys → New API Key"
    echo "3. Set environment variables:"
    echo "   export GRAFANA_URL='https://your-grafana-url'"
    echo "   export GRAFANA_API_KEY='your-api-key'"
    echo "4. Run this script again"
    echo ""
    echo -e "${GREEN}Alternative: Manual import${NC}"
    echo "Upload $DASHBOARD_FILE manually via Grafana UI:"
    echo "  Dashboard → Import → Upload JSON file"
    echo ""
    exit 1
fi

if [ -z "$GRAFANA_API_KEY" ]; then
    echo -e "${RED}Error: GRAFANA_API_KEY environment variable is required${NC}"
    echo "Create an API key in Grafana: Settings → API Keys → New API Key"
    echo "Then: export GRAFANA_API_KEY='your-api-key'"
    exit 1
fi

echo "Grafana URL: $GRAFANA_URL"
echo "Dashboard file: $DASHBOARD_FILE"
echo ""

# Read dashboard JSON
DASHBOARD_JSON=$(cat "$DASHBOARD_FILE")

# Prepare the payload for Grafana API
# Grafana expects: {"dashboard": {...}, "overwrite": true}
PAYLOAD=$(jq -n \
  --argjson dashboard "$DASHBOARD_JSON" \
  '{
    dashboard: $dashboard.dashboard,
    overwrite: true,
    message: "Deployed via automation script"
  }')

echo "Deploying dashboard to Grafana..."

# Deploy to Grafana
RESPONSE=$(curl -s -w "\n%{http_code}" \
  -X POST \
  -H "Authorization: Bearer $GRAFANA_API_KEY" \
  -H "Content-Type: application/json" \
  -d "$PAYLOAD" \
  "$GRAFANA_URL/api/dashboards/db")

# Extract HTTP status code (last line)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Dashboard deployed successfully!${NC}"
    echo ""
    
    # Extract dashboard URL from response
    DASHBOARD_URL=$(echo "$RESPONSE_BODY" | jq -r '.url // empty')
    DASHBOARD_UID=$(echo "$RESPONSE_BODY" | jq -r '.uid // empty')
    
    if [ -n "$DASHBOARD_URL" ]; then
        echo "Dashboard URL: $GRAFANA_URL$DASHBOARD_URL"
    fi
    if [ -n "$DASHBOARD_UID" ]; then
        echo "Dashboard UID: $DASHBOARD_UID"
    fi
    echo ""
    echo -e "${GREEN}Dashboard is now live!${NC}"
    exit 0
else
    echo -e "${RED}✗ Failed to deploy dashboard${NC}"
    echo "HTTP Status: $HTTP_CODE"
    echo "Response:"
    echo "$RESPONSE_BODY" | jq '.' 2>/dev/null || echo "$RESPONSE_BODY"
    exit 1
fi
