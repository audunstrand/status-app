# Deployment

## Fly.io

Already deployed and configured.

### Update Deployment

```bash
./deploy.sh
```

Or manually:
```bash
flyctl deploy --config fly.api.toml
```

### Secrets

```bash
fly secrets set API_SECRET=<secret> -a status-app-commands
```

## GitHub Actions

Auto-deploys on push to master.

Workflow: `.github/workflows/ci.yml`
- Runs tests
- Deploys all services to Fly.io
