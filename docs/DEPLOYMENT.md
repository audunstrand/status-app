# Deployment

## Fly.io

**Current Architecture**: 3 services + 1 database

Services:
- `status-app-backend` - Commands + API + Projections (consolidated)
- `status-app-slackbot` - Slack integration
- `status-app-scheduler` - Reminder scheduling
- `status-app-db` - PostgreSQL database

### Manual Deploy (if needed)

Deploy individual service:
```bash
flyctl deploy --config fly.backend.toml --remote-only
flyctl deploy --config fly.slackbot.toml --remote-only
flyctl deploy --config fly.scheduler.toml --remote-only
```

Deploy all services:
```bash
for config in fly.backend.toml fly.slackbot.toml fly.scheduler.toml; do 
  flyctl deploy --config $config --remote-only
done
```

### Secrets

Backend service:
```bash
flyctl secrets set API_SECRET=<secret> -a status-app-backend
flyctl secrets set EVENT_STORE_URL=<url> -a status-app-backend
flyctl secrets set PROJECTION_DB_URL=<url> -a status-app-backend
```

Slackbot service:
```bash
flyctl secrets set API_SECRET=<secret> -a status-app-slackbot
flyctl secrets set SLACK_BOT_TOKEN=<token> -a status-app-slackbot
flyctl secrets set SLACK_SIGNING_KEY=<key> -a status-app-slackbot
flyctl secrets set COMMANDS_URL=https://status-app-backend.fly.dev -a status-app-slackbot
```

Scheduler service:
```bash
flyctl secrets set API_SECRET=<secret> -a status-app-scheduler
flyctl secrets set PROJECTION_DB_URL=<url> -a status-app-scheduler
flyctl secrets set COMMANDS_URL=https://status-app-backend.fly.dev -a status-app-scheduler
```

### Database

Attach database to services:
```bash
flyctl postgres attach status-app-db -a status-app-backend
flyctl postgres attach status-app-db -a status-app-scheduler
```

### Monitoring

Check service status:
```bash
flyctl status -a status-app-backend
flyctl status -a status-app-slackbot
flyctl status -a status-app-scheduler
```

View logs:
```bash
flyctl logs -a status-app-backend
flyctl logs -a status-app-slackbot
flyctl logs -a status-app-scheduler
```

Check health:
```bash
curl https://status-app-backend.fly.dev/health
```

## GitHub Actions

**Automatic deployment on push to master.**

Workflow: `.github/workflows/deploy.yml`
- Runs all tests (unit + E2E)
- Deploys 3 services to Fly.io (backend, slackbot, scheduler)

No manual deployment needed for normal development.
