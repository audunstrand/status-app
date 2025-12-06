# Fly.io Deployment

Configuration files for deploying to Fly.io.

## Files

- `fly.toml` - Commands service configuration
- `fly.api.toml` - API service configuration
- `fly.projections.toml` - Projections service configuration
- `fly.scheduler.toml` - Scheduler service configuration
- `fly.slackbot.toml` - Slackbot service configuration
- `deploy.sh` - Automated deployment script

## Quick Deploy

```bash
cd deployments/fly
./deploy.sh
```

## Manual Deploy

```bash
# Deploy individual services
flyctl deploy --config fly.api.toml --ha=false
flyctl deploy --config fly.toml --ha=false
flyctl deploy --config fly.projections.toml --ha=false
flyctl deploy --config fly.scheduler.toml --ha=false
flyctl deploy --config fly.slackbot.toml --ha=false
```

See `../../docs/FLY_DEPLOYMENT.md` for complete setup instructions.
