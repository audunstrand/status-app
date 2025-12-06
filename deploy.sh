#!/bin/bash
set -e

echo "🚀 Deploying Status App to Fly.io"
echo ""

# Check if flyctl is installed
if ! command -v flyctl &> /dev/null; then
    echo "❌ flyctl not found. Install it: https://fly.io/docs/hands-on/install-flyctl/"
    exit 1
fi

# Check if logged in
if ! flyctl auth whoami &> /dev/null; then
    echo "❌ Not logged in to Fly.io. Run: flyctl auth login"
    exit 1
fi

echo "📦 Step 1: Create PostgreSQL databases (if not exists)"
echo ""

# Create event store database
echo "Creating event store database..."
flyctl postgres create --name status-app-eventstore --region arn --initial-cluster-size 1 --vm-size shared-cpu-1x --volume-size 1 || echo "Event store DB may already exist"

# Create projections database
echo "Creating projections database..."
flyctl postgres create --name status-app-projections-db --region arn --initial-cluster-size 1 --vm-size shared-cpu-1x --volume-size 1 || echo "Projections DB may already exist"

echo ""
echo "📝 Step 2: Note your database connection strings:"
echo "Run these commands to get connection strings:"
echo "  flyctl postgres connect -a status-app-eventstore"
echo "  flyctl postgres connect -a status-app-projections-db"
echo ""
echo "Press Enter to continue when ready..."
read

echo ""
echo "🏗️  Step 3: Deploy services"
echo ""

# Deploy API (public-facing)
echo "Deploying API service..."
flyctl deploy --config fly.api.toml --ha=false

# Deploy Commands (public-facing)
echo "Deploying Commands service..."
flyctl deploy --config fly.toml --ha=false

# Deploy Projections (background worker)
echo "Deploying Projections service..."
flyctl deploy --config fly.projections.toml --ha=false

# Deploy Scheduler (background worker)
echo "Deploying Scheduler service..."
flyctl deploy --config fly.scheduler.toml --ha=false

# Deploy Slackbot (background worker)
echo "Deploying Slackbot service..."
flyctl deploy --config fly.slackbot.toml --ha=false

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📋 Next steps:"
echo "1. Set secrets for each app:"
echo "   flyctl secrets set -a status-app-commands EVENT_STORE_URL='postgres://...'"
echo "   flyctl secrets set -a status-app-api PROJECTION_DB_URL='postgres://...'"
echo "   flyctl secrets set -a status-app-projections EVENT_STORE_URL='...' PROJECTION_DB_URL='...'"
echo "   flyctl secrets set -a status-app-slackbot SLACK_BOT_TOKEN='...' SLACK_SIGNING_KEY='...'"
echo "   flyctl secrets set -a status-app-scheduler PROJECTION_DB_URL='postgres://...'"
echo ""
echo "2. Run migrations:"
echo "   flyctl ssh console -a status-app-commands"
echo "   # Then run migrations manually or use a migration job"
echo ""
echo "3. Check status:"
echo "   flyctl status -a status-app-api"
echo "   flyctl logs -a status-app-commands"
