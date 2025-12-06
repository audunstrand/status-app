# Docker Images

Dockerfiles for all services.

## Files

- `Dockerfile.api` - Query API service
- `Dockerfile.commands` - Command handler service
- `Dockerfile.projections` - Projection builder service
- `Dockerfile.scheduler` - Scheduler service
- `Dockerfile.slackbot` - Slack bot service

## Build Locally

```bash
# From project root
docker build -f deployments/docker/Dockerfile.api -t status-app-api .
docker build -f deployments/docker/Dockerfile.commands -t status-app-commands .
docker build -f deployments/docker/Dockerfile.projections -t status-app-projections .
docker build -f deployments/docker/Dockerfile.scheduler -t status-app-scheduler .
docker build -f deployments/docker/Dockerfile.slackbot -t status-app-slackbot .
```

## Run Locally

```bash
# Run API service
docker run -p 8080:8080 \
  -e PROJECTION_DB_URL="postgres://..." \
  status-app-api
```

All images use:
- **Base**: `golang:1.23-alpine` for building
- **Runtime**: `alpine:latest` for minimal size
- **Size**: ~8-9MB per image
