# Deployment

## Fly.io

Already deployed and configured.

### Manual Deploy (if needed)

```bash
flyctl deploy --config fly.api.toml
```

Deploy all services:
```bash
for config in fly*.toml; do flyctl deploy --config $config; done
```

### Secrets

```bash
fly secrets set API_SECRET=<secret> -a status-app-commands
```

## GitHub Actions

**Automatic deployment on push to master.**

Workflow: `.github/workflows/ci.yml`
- Runs tests
- Deploys all services to Fly.io

No manual deployment needed.
