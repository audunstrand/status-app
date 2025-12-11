# Grafana Dashboard Deployment Guide

## Quick Start

### Option 1: GitHub Actions (Recommended)

**Setup once:**
1. Go to: Settings → Secrets and variables → Actions
2. Add **Secret**: `GRAFANA_API_KEY` = `glsa_your_api_key_here`
3. Add **Variable**: `GRAFANA_URL` = `https://your-grafana-url` (optional)

**Then:**
- **Automatic**: Push changes to `docs/grafana-dashboard.json` → deploys automatically
- **Manual**: Actions → Deploy Grafana Dashboard → Run workflow

### Option 2: Makefile (Local)

```bash
export GRAFANA_URL="https://your-grafana-url"
export GRAFANA_API_KEY="glsa_your_api_key_here"
make deploy-dashboard
```

### Option 3: Python Script

```bash
python3 scripts/deploy-grafana-dashboard.py --url $GRAFANA_URL --key $GRAFANA_API_KEY
```

### Option 4: Manual Upload

1. Open Grafana
2. Dashboard → Import → Upload JSON
3. Select `docs/grafana-dashboard.json`
4. Choose Prometheus datasource
5. Import

---

## Getting Credentials

### Step 1: Find Your Grafana URL

**For Fly.io:**
```bash
flyctl open --app status-app-backend
# Look for "Metrics" or "Grafana" link
```

Typical format: `https://fly-metrics-{org}.fly.dev`

### Step 2: Create API Key

1. Open Grafana in browser
2. Go to: Configuration ⚙️ → API Keys
3. Click "New API Key"
   - **Name**: `GitHub Actions Dashboard Deploy`
   - **Role**: `Editor` (required)
   - **Time to live**: Leave empty or set expiration
4. Click "Add"
5. **Copy the key immediately** (starts with `glsa_`)

---

## GitHub Actions Workflow

The workflow (`.github/workflows/deploy-grafana.yml`) automatically:

✅ Triggers on push to `docs/grafana-dashboard.json`  
✅ Can be manually triggered with custom URL  
✅ Uses native `curl` + `jq` (no custom actions)  
✅ Validates deployment and shows dashboard URL  
✅ Adds summary to Actions output  

**How it works:**
```yaml
- Uses standard curl to call Grafana HTTP API
- Reads dashboard JSON from docs/
- Posts to /api/dashboards/db endpoint
- Overwrites existing dashboard
- Reports success/failure
```

**No external dependencies** - uses only GitHub Actions built-in tools.

---

## Testing Deployment

### Dry Run (Check Configuration)

```bash
# Test that dashboard JSON is valid
jq '.' docs/grafana-dashboard.json > /dev/null && echo "✓ Dashboard JSON is valid"

# Test credentials (without deploying)
curl -s -H "Authorization: Bearer $GRAFANA_API_KEY" \
  "$GRAFANA_URL/api/dashboards/db" \
  | jq '.message' # Should not error
```

### Actual Deployment

```bash
# Deploy
make deploy-dashboard

# Or directly
python3 scripts/deploy-grafana-dashboard.py
```

### Verify in Grafana

1. Open Grafana
2. Search for "Status App - Event Sourcing Observability"
3. Dashboard should appear with 12 panels
4. Check that panels show data from Prometheus

---

## Troubleshooting

### "401 Unauthorized"

❌ **Problem**: Invalid or expired API key

✅ **Solution**:
1. Create new API key with `Editor` role
2. Update `GRAFANA_API_KEY` secret/variable
3. Verify key starts with `glsa_`

### "404 Not Found"

❌ **Problem**: Wrong Grafana URL

✅ **Solution**:
1. Verify URL in browser first
2. Check URL doesn't have trailing slash
3. Confirm Grafana is accessible

### Dashboard deploys but shows no data

❌ **Problem**: Datasource not configured or metrics not flowing

✅ **Solution**:
1. Verify Prometheus datasource exists in Grafana
2. Test metrics endpoint: `curl https://your-app.fly.dev/metrics`
3. Check Fly.io is scraping metrics
4. Verify datasource name in dashboard matches

### Workflow doesn't trigger

❌ **Problem**: Missing secrets/variables

✅ **Solution**:
1. Verify `GRAFANA_API_KEY` secret exists
2. For auto-deploy: Set `GRAFANA_URL` variable
3. Check workflow file is on master branch

---

## Advanced Usage

### Deploy to Multiple Environments

```bash
# Production
GRAFANA_URL=$PROD_GRAFANA_URL \
GRAFANA_API_KEY=$PROD_API_KEY \
make deploy-dashboard

# Staging
GRAFANA_URL=$STAGING_GRAFANA_URL \
GRAFANA_API_KEY=$STAGING_API_KEY \
make deploy-dashboard
```

### Update Dashboard

1. Edit `docs/grafana-dashboard.json`
2. Commit and push
3. GitHub Actions automatically deploys
4. Dashboard updates in Grafana (overwrite mode)

### Rollback Dashboard

```bash
# Checkout previous version
git checkout HEAD~1 docs/grafana-dashboard.json

# Deploy old version
make deploy-dashboard

# Or commit and push to trigger workflow
```

---

## Security Best Practices

✅ **DO:**
- Store API keys in GitHub Secrets (encrypted)
- Use `Editor` role (not Admin)
- Set API key expiration dates
- Rotate keys regularly
- Use different keys per environment

❌ **DON'T:**
- Commit API keys to git
- Share API keys in chat/email
- Use same key for dev and prod
- Give keys `Admin` role unnecessarily

---

## What Gets Deployed

**Dashboard Title**: Status App - Event Sourcing Observability

**Panels (12 total)**:
1. Events Stored per Second (graph)
2. Projection Lag (gauge)
3. Event Type Distribution (pie chart)
4. Projection Updates per Minute (graph)
5. Projection Processing Duration p95 (graph)
6. Slack Messages (stat)
7. Slack Commands Handled (graph)
8. Backend API Calls (graph)
9. Scheduler - Reminders Sent (stat)
10. Error Rates (graph)
11. Event Store Size (graph)

**Metrics Source**: Prometheus (scraped from `/metrics` endpoints)

**Auto-refresh**: Every 30 seconds

**Time range**: Last 1 hour (adjustable)

---

## Files Reference

```
.github/workflows/deploy-grafana.yml  # GitHub Actions workflow
docs/grafana-dashboard.json          # Dashboard definition
scripts/deploy-grafana-dashboard.py  # Python deployment script
scripts/deploy-grafana-dashboard.sh  # Bash deployment script
scripts/README.md                    # Detailed script documentation
Makefile                            # make deploy-dashboard target
```

---

## Support

**Issues with deployment?**
1. Check troubleshooting section above
2. Verify credentials are correct
3. Test Grafana access in browser
4. Check GitHub Actions logs for errors

**Issues with dashboard data?**
1. Verify `/metrics` endpoints are accessible
2. Check Prometheus is scraping metrics
3. Confirm services are running and healthy

**Dashboard customization?**
1. Export from Grafana after making changes
2. Save to `docs/grafana-dashboard.json`
3. Commit and push to deploy changes

---

## Summary

✅ **Simple**: One command or automatic on push  
✅ **Secure**: Uses GitHub Secrets, never exposes keys  
✅ **Automated**: No manual steps after setup  
✅ **Reliable**: Uses native GitHub Actions tools  
✅ **Flexible**: Multiple deployment options  

The dashboard can be deployed programmatically using GitHub Actions, eliminating manual import steps! 🚀
