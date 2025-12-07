# Security

## Authentication

**Required**: All services require `API_SECRET` environment variable.

### Setup

```bash
# Generate secret
openssl rand -hex 32

# Local
export API_SECRET=<secret>

# Production (already configured)
fly secrets set API_SECRET=<secret> -a status-app-commands
```

### Protected Endpoints

- `POST /commands/*` - Requires `Authorization: Bearer <secret>`
- `GET /api/*` - Requires `Authorization: Bearer <secret>`
- `GET /health` - Public

### Troubleshooting

**Service won't start**: Set `API_SECRET` environment variable

**401 errors**: Verify all services have same `API_SECRET`
