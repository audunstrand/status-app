# Deploying to Fly.io

## Prerequisites

1. **Install Fly.io CLI**
   ```bash
   # macOS
   brew install flyctl
   
   # Linux
   curl -L https://fly.io/install.sh | sh
   
   # Windows
   # Download from https://fly.io/docs/hands-on/install-flyctl/
   ```

2. **Sign up and login**
   ```bash
   flyctl auth signup  # Or flyctl auth login if you have an account
   ```

3. **Add a credit card** (required even for free tier)
   - Visit https://fly.io/dashboard/personal/billing

## Step-by-Step Deployment

### 1. Create PostgreSQL Databases

```bash
# Event Store database (Stockholm region - closest to Norway)
flyctl postgres create \
  --name status-app-eventstore \
  --region arn \
  --initial-cluster-size 1 \
  --vm-size shared-cpu-1x \
  --volume-size 1

# Projections database  
flyctl postgres create \
  --name status-app-projections-db \
  --region arn \
  --initial-cluster-size 1 \
  --vm-size shared-cpu-1x \
  --volume-size 1
```

**Save the connection strings** shown after creation!

### 2. Create Apps (One-time setup)

```bash
# API service
flyctl apps create status-app-api

# Commands service
flyctl apps create status-app-commands

# Projections service
flyctl apps create status-app-projections

# Scheduler service
flyctl apps create status-app-scheduler

# Slackbot service
flyctl apps create status-app-slackbot
```

### 3. Attach Databases to Apps

```bash
# Attach event store to commands and projections
flyctl postgres attach status-app-eventstore -a status-app-commands
flyctl postgres attach status-app-eventstore -a status-app-projections

# Attach projections DB to API, projections, and scheduler
flyctl postgres attach status-app-projections-db -a status-app-api
flyctl postgres attach status-app-projections-db -a status-app-projections
flyctl postgres attach status-app-projections-db -a status-app-scheduler
```

### 4. Set Secrets

```bash
# Commands service
flyctl secrets set -a status-app-commands \
  EVENT_STORE_URL="postgres://user:pass@host/db"

# API service
flyctl secrets set -a status-app-api \
  PROJECTION_DB_URL="postgres://user:pass@host/db"

# Projections service
flyctl secrets set -a status-app-projections \
  EVENT_STORE_URL="postgres://..." \
  PROJECTION_DB_URL="postgres://..."

# Scheduler service
flyctl secrets set -a status-app-scheduler \
  PROJECTION_DB_URL="postgres://..."

# Slackbot service (get tokens from Slack App dashboard)
flyctl secrets set -a status-app-slackbot \
  SLACK_BOT_TOKEN="xoxb-your-token" \
  SLACK_SIGNING_KEY="xapp-your-key" \
  EVENT_STORE_URL="postgres://..." # For sending commands
```

### 5. Run Migrations

Option A: Manual via SSH
```bash
# SSH into commands app
flyctl ssh console -a status-app-commands

# Inside the container, run migrations
# (You'll need to install golang-migrate first or run SQL directly)
```

Option B: Run migrations locally pointing to Fly databases
```bash
# Get connection string
flyctl postgres connect -a status-app-eventstore

# Run locally
export EVENT_STORE_URL="postgres://..."
export PROJECTION_DB_URL="postgres://..."
make migrate-up
```

### 6. Deploy All Services

Use the automated script:
```bash
./deploy.sh
```

Or deploy manually:
```bash
flyctl deploy --config fly.toml --ha=false              # Commands
flyctl deploy --config fly.api.toml --ha=false          # API
flyctl deploy --config fly.projections.toml --ha=false  # Projections
flyctl deploy --config fly.scheduler.toml --ha=false    # Scheduler
flyctl deploy --config fly.slackbot.toml --ha=false     # Slackbot
```

## Verify Deployment

```bash
# Check app status
flyctl status -a status-app-api
flyctl status -a status-app-commands

# View logs
flyctl logs -a status-app-commands
flyctl logs -a status-app-projections
flyctl logs -a status-app-api

# Get API URL
flyctl info -a status-app-api
```

## Access Your API

```bash
# Get API endpoint
flyctl info -a status-app-api

# Test it
curl https://status-app-api.fly.dev/api/teams
```

## Useful Commands

```bash
# Scale a service
flyctl scale count 2 -a status-app-api

# SSH into a service
flyctl ssh console -a status-app-commands

# Restart a service
flyctl apps restart status-app-api

# View metrics
flyctl dashboard -a status-app-api

# Destroy an app (careful!)
flyctl apps destroy status-app-api
```

## Costs Estimate (Free Tier)

Fly.io free tier includes:
- 3 shared-cpu-1x VMs (256MB RAM each)
- 3GB persistent storage
- 160GB outbound data transfer

**Our setup uses:**
- 5 VMs (commands, api, projections, scheduler, slackbot) - **2 over free tier**
- 2 Postgres instances (2GB storage) - within limits

**Expected monthly cost:** ~$10-15/month for the 2 extra VMs

**To stay free:**
- Combine services (e.g., merge scheduler into projections)
- Use auto-stop for low-traffic services

## Troubleshooting

**Build fails:**
```bash
# Check logs
flyctl logs -a status-app-commands

# Try local build
docker build -f Dockerfile.commands -t test .
```

**Database connection issues:**
```bash
# Test connection
flyctl postgres connect -a status-app-eventstore

# Check if attached
flyctl postgres list
```

**Service won't start:**
```bash
# Check secrets
flyctl secrets list -a status-app-commands

# Increase memory
flyctl scale memory 512 -a status-app-commands
```
