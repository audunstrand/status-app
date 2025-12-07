# API Security Setup Guide

This guide explains the API authentication setup for the Status App services.

## Overview

**⚠️ REQUIRED**: The app requires shared secret authentication for all services. Services will not start without API_SECRET configured.

All service-to-service communication requires a valid API key via the `Authorization: Bearer <secret>` header.

## Generate API Secret

First, generate a secure random secret:

```bash
openssl rand -hex 32
```

Save this secret - you'll need it for all services.

Example output: `c69517c23d7ce83edd2e097a265e24b04861b3110282bae1e711deffd9fe70b4`

## Local Development

Set the API_SECRET environment variable (required):

```bash
export API_SECRET=$(openssl rand -hex 32)
export COMMANDS_URL=http://localhost:8081
```

Or create a `.env` file:

```bash
API_SECRET=c69517c23d7ce83edd2e097a265e24b04861b3110282bae1e711deffd9fe70b4
COMMANDS_URL=http://localhost:8081
```

**Note**: Services will fail to start if API_SECRET is not set.

## Production Deployment (Fly.io)

✅ **Already configured** for this project! 

If you need to update the secret:

```bash
# Generate new secret
SECRET=$(openssl rand -hex 32)

# Set for all apps
fly secrets set API_SECRET=$SECRET -a status-app-commands
fly secrets set API_SECRET=$SECRET -a status-app-api
fly secrets set API_SECRET=$SECRET -a status-app-slackbot
fly secrets set API_SECRET=$SECRET -a status-app-scheduler
fly secrets set API_SECRET=$SECRET -a status-app-projections
```

## Verify Setup

After setting secrets, verify authentication is working:

```bash
# This should fail with 401 Unauthorized
curl https://status-app-commands.fly.dev/commands/submit-update \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"team_id":"test","content":"test","author":"test"}'

# This should succeed
curl https://status-app-commands.fly.dev/commands/submit-update \
  -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_SECRET_HERE" \
  -d '{"team_id":"test","content":"test","author":"test"}'

# Health check should work without auth
curl https://status-app-commands.fly.dev/health
```

## Protected Endpoints

### Commands Service
- `POST /commands/submit-update` - Submit status update (requires auth)
- `POST /commands/register-team` - Register new team (requires auth)
- `GET /health` - Health check (public)

### API Service
- `GET /api/teams` - List all teams (requires auth)
- `GET /api/teams/{id}` - Get team details (requires auth)
- `GET /api/updates` - Get recent updates (requires auth)
- `GET /api/teams/{id}/updates` - Get team updates (requires auth)
- `GET /health` - Health check (public)

## Authentication Flow

```
┌─────────────┐
│  Slackbot   │
└──────┬──────┘
       │ POST /commands/submit-update
       │ Authorization: Bearer <secret>
       ▼
┌─────────────────┐
│ Auth Middleware │──── Validate token
└────────┬────────┘
         │ ✅ Token valid
         ▼
┌─────────────────┐
│ Command Handler │
└─────────────────┘
```

## Rotating Secrets

To rotate the API secret:

```bash
# Generate new secret
NEW_SECRET=$(openssl rand -hex 32)

# Update all services (they will restart automatically)
fly secrets set API_SECRET=$NEW_SECRET -a status-app-commands
fly secrets set API_SECRET=$NEW_SECRET -a status-app-api
fly secrets set API_SECRET=$NEW_SECRET -a status-app-slackbot
fly secrets set API_SECRET=$NEW_SECRET -a status-app-scheduler
```

**Note**: Services will restart when secrets are updated. Consider doing this during low-traffic periods.

## Troubleshooting

### Service won't start

**Problem**: Service fails to start with error

**Solution**: 
- Check that API_SECRET environment variable is set
- Services now REQUIRE API_SECRET and will fail fast if missing
- Set it using `export API_SECRET=<secret>` or Fly.io secrets

### 401 Unauthorized errors

**Problem**: Getting 401 errors when services try to communicate

**Solution**: 
1. Verify all services have the same API_SECRET set
2. Check logs: `fly logs -a status-app-commands`
3. Ensure Authorization header is being sent

### Services can't communicate

**Problem**: Services can't reach each other

**Solution**:
1. Verify COMMANDS_URL is set correctly
2. Check network connectivity between Fly.io apps
3. Verify services are running: `fly status -a status-app-commands`

## Security Best Practices

1. ✅ **Never commit secrets to git**
   - Secrets are in .gitignore
   - Use Fly.io secrets management

2. ✅ **Use strong secrets**
   - Always use `openssl rand -hex 32`
   - Never use simple passwords

3. ✅ **Rotate secrets periodically**
   - Recommend rotating every 90 days
   - Rotate immediately if compromised

4. ✅ **Monitor access logs**
   - Watch for unauthorized access attempts
   - Set up alerts for 401 errors

5. ✅ **Keep secrets secret**
   - Don't share in chat/email
   - Use secure secret management tools
