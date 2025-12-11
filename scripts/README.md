# Grafana Dashboard Deployment Scripts

This directory contains scripts for programmatically deploying the Grafana dashboard to Fly.io or any Grafana instance.

## Available Scripts

### 1. Python Script (Recommended)

**File:** `deploy-grafana-dashboard.py`

**Advantages:**
- Cross-platform (Windows, macOS, Linux)
- Better error handling
- More informative output
- No external dependencies (uses stdlib only)

**Usage:**

```bash
# Using environment variables
export GRAFANA_URL="https://your-grafana-url"
export GRAFANA_API_KEY="your-api-key"
python3 scripts/deploy-grafana-dashboard.py

# Or with command line arguments
python3 scripts/deploy-grafana-dashboard.py \
  --url https://your-grafana-url \
  --key your-api-key

# Custom dashboard file
python3 scripts/deploy-grafana-dashboard.py \
  --dashboard path/to/dashboard.json
```

### 2. Bash Script (Unix/Linux)

**File:** `deploy-grafana-dashboard.sh`

**Requirements:** `jq`, `curl`

**Usage:**

```bash
# Make executable
chmod +x scripts/deploy-grafana-dashboard.sh

# Deploy
export GRAFANA_URL="https://your-grafana-url"
export GRAFANA_API_KEY="your-api-key"
./scripts/deploy-grafana-dashboard.sh
```

### 3. GitHub Actions (Automated)

**File:** `.github/workflows/deploy-grafana.yml`

**Triggers:**
- Manual workflow dispatch
- Automatic on push to `docs/grafana-dashboard.json`

**Setup:**

1. **Add GitHub Secrets:**
   - Go to: Settings → Secrets and variables → Actions
   - Add secret: `GRAFANA_API_KEY` with your API key

2. **Add GitHub Variables (optional):**
   - Go to: Settings → Secrets and variables → Actions → Variables
   - Add variable: `GRAFANA_URL` with your Grafana URL
   - This enables automatic deployment on dashboard changes

3. **Manual Trigger:**
   - Go to: Actions → Deploy Grafana Dashboard → Run workflow
   - Enter Grafana URL
   - Click "Run workflow"

---

## Setup Guide

### Step 1: Get Fly.io Grafana URL

**Option A: Via Dashboard**
1. Go to https://fly.io/dashboard
2. Navigate to your organization
3. Look for "Metrics" or "Grafana" link
4. Copy the URL (e.g., `https://fly-metrics-XXXXX.fly.dev`)

**Option B: Via CLI**
```bash
flyctl open --app status-app-backend
# Look for Grafana/Metrics link in the dashboard
```

**Note:** Fly.io Grafana URLs are typically:
- Organization-specific: `https://fly-metrics-{org-slug}.fly.dev`
- Or shared: `https://fly.io/apps/{app-name}/metrics`

### Step 2: Create Grafana API Key

1. **Access Grafana:**
   - Open your Grafana URL in browser
   - Log in with Fly.io credentials

2. **Create API Key:**
   - Click on Configuration (⚙️) → API Keys
   - Click "New API Key"
   - Settings:
     - **Key name:** `Dashboard Deployment`
     - **Role:** `Editor` (required for creating/updating dashboards)
     - **Time to live:** Set as needed (or leave empty for no expiration)
   - Click "Add"
   - **Copy the API key immediately** (you won't see it again!)

3. **Save API Key Securely:**
   ```bash
   # For local use
   echo "export GRAFANA_API_KEY='glsa_...'" >> ~/.bashrc
   # Or use a password manager
   ```

### Step 3: Deploy Dashboard

**Local deployment:**
```bash
export GRAFANA_URL="https://your-grafana-url"
export GRAFANA_API_KEY="glsa_your_key_here"
python3 scripts/deploy-grafana-dashboard.py
```

**GitHub Actions deployment:**
- Add secrets and variables as described above
- Push changes or trigger manually

---

## Troubleshooting

### Error: "Dashboard file not found"

**Solution:** Run script from project root:
```bash
cd /path/to/status-app
python3 scripts/deploy-grafana-dashboard.py
```

### Error: "401 Unauthorized"

**Causes:**
1. Invalid or expired API key
2. API key doesn't have `Editor` role

**Solution:**
- Create a new API key with `Editor` role
- Verify API key is correct (starts with `glsa_`)

### Error: "404 Not Found"

**Causes:**
1. Wrong Grafana URL
2. Grafana instance not accessible

**Solution:**
- Verify Grafana URL is correct
- Test access in browser first
- Check if URL needs authentication

### Error: "Connection refused" or "SSL error"

**Causes:**
1. Grafana instance is down
2. Network connectivity issues
3. SSL certificate problems

**Solution:**
- Check Grafana is accessible via browser
- Verify network connection
- Try with explicit port: `https://grafana:3000`

### Dashboard deploys but doesn't show data

**Causes:**
1. Prometheus datasource not configured
2. Wrong datasource name in dashboard
3. Metrics not being scraped

**Solution:**
1. Verify Prometheus datasource exists in Grafana
2. Check datasource name matches in dashboard panels
3. Test metrics endpoint: `curl http://localhost:8080/metrics`
4. Verify Fly.io is scraping metrics

---

## Advanced Usage

### Update Existing Dashboard

The scripts automatically overwrite existing dashboards with the same title.

```bash
# Edit dashboard
vim docs/grafana-dashboard.json

# Deploy update
python3 scripts/deploy-grafana-dashboard.py
```

### Deploy to Multiple Grafana Instances

```bash
# Production
GRAFANA_URL="https://prod-grafana" \
GRAFANA_API_KEY="$PROD_KEY" \
python3 scripts/deploy-grafana-dashboard.py

# Staging
GRAFANA_URL="https://staging-grafana" \
GRAFANA_API_KEY="$STAGING_KEY" \
python3 scripts/deploy-grafana-dashboard.py
```

### CI/CD Integration

**Example with other CI systems:**

```yaml
# GitLab CI
deploy-dashboard:
  script:
    - python3 scripts/deploy-grafana-dashboard.py
  only:
    - master
  variables:
    GRAFANA_URL: $GRAFANA_URL
    GRAFANA_API_KEY: $GRAFANA_API_KEY
```

```groovy
// Jenkins
stage('Deploy Dashboard') {
    environment {
        GRAFANA_URL = credentials('grafana-url')
        GRAFANA_API_KEY = credentials('grafana-api-key')
    }
    steps {
        sh 'python3 scripts/deploy-grafana-dashboard.py'
    }
}
```

---

## Security Best Practices

1. **Never commit API keys to git**
   - Use environment variables
   - Use CI/CD secrets
   - Use password managers

2. **Use limited-scope API keys**
   - Only `Editor` role (not `Admin`)
   - Set expiration dates
   - Rotate regularly

3. **Restrict API key access**
   - Store in secrets management system
   - Use different keys for different environments
   - Revoke unused keys

4. **Audit deployments**
   - Check Grafana audit logs
   - Monitor API key usage
   - Track dashboard changes

---

## Alternatives

### Manual Import (No Script Needed)

1. Open Grafana
2. Click "+" → "Import"
3. Click "Upload JSON file"
4. Select `docs/grafana-dashboard.json`
5. Select Prometheus datasource
6. Click "Import"

**Pros:** No setup required, works immediately  
**Cons:** Manual process, not automated

### Grafana Provisioning (For Self-Hosted)

If you're running your own Grafana (not Fly.io):

```yaml
# /etc/grafana/provisioning/dashboards/status-app.yaml
apiVersion: 1

providers:
  - name: 'Status App'
    folder: 'Applications'
    type: file
    options:
      path: /var/lib/grafana/dashboards
```

Copy `grafana-dashboard.json` to `/var/lib/grafana/dashboards/`

---

## References

- [Grafana HTTP API Documentation](https://grafana.com/docs/grafana/latest/developers/http_api/)
- [Grafana Dashboard Provisioning](https://grafana.com/docs/grafana/latest/administration/provisioning/)
- [Fly.io Metrics Documentation](https://fly.io/docs/reference/metrics/)
