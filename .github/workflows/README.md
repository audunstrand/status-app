# GitHub Actions Workflows

This directory contains GitHub Actions workflows for CI/CD.

## Workflows

### `deploy.yml` - CI/CD Pipeline

**Triggers:**
- Push to `master` branch
- Pull requests to `master` branch

**Jobs:**

1. **Test** (runs on every push/PR)
   - Unit tests
   - E2E tests with real PostgreSQL

2. **Deploy** (runs only on push to master, after tests pass)
   - Deploys all 5 services to Fly.io in parallel
   - Commands service
   - API service
   - Projections service
   - Slackbot service
   - Scheduler service

## Setup Required

See `docs/GITHUB_ACTIONS.md` for complete setup instructions.

**Quick setup:**
```bash
# 1. Get Fly.io token
fly tokens create deploy -x 999999h

# 2. Add to GitHub Secrets:
#    Name: FLY_API_TOKEN
#    Value: <your-token>

# 3. Push to master
git push origin master
```

## Status Badge

Add this to your README.md:
```markdown
![CI/CD](https://github.com/audunstrand/status-app/workflows/CI%2FCD%20Pipeline/badge.svg)
```
