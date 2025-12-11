# ADR 007: Mandatory API Authentication

**Date**: 2025-11-25  
**Status**: Accepted  
**Deciders**: Audun

## Context

The initial implementation included optional API authentication:
- Authentication could be toggled via environment variable
- Code had conditional branches: `if cfg.APISecret != "" { ... }`
- Some endpoints could be accessed without authentication
- Made testing easier but created security risk

This approach had several problems:
1. **Security risk**: Easy to accidentally deploy without authentication
2. **Code complexity**: Conditional logic throughout codebase
3. **Inconsistent behavior**: System behaved differently based on config
4. **Testing confusion**: Tests passed with/without auth, masking issues

## Decision

**Always require API authentication via `X-API-Secret` header.**

Remove all conditional authentication logic:
- `API_SECRET` environment variable is **required**
- All API requests must include `X-API-Secret` header
- No code branches for "auth disabled" mode
- Fail fast on startup if `API_SECRET` not set

## Options Considered

### Option 1: Optional Authentication (Status Quo)
**Implementation**: Allow running without authentication for development

**Pros**:
- Easier local development (no secret needed)
- Simpler testing (no headers to add)
- "Flexible" deployment options

**Cons**:
- Security risk if deployed to production without auth
- Code complexity (conditional logic everywhere)
- Tests might pass without auth, fail with auth
- Inconsistent system behavior
- False sense of security

### Option 2: Mandatory Authentication ✅ **CHOSEN**
**Implementation**: Always require valid API secret

**Pros**:
- Security by default
- Simpler codebase (no conditional logic)
- Consistent behavior across environments
- Forces proper security in all environments
- Tests always run with auth (more realistic)

**Cons**:
- Requires secret in all environments
- Slightly more setup for local development
- Need to manage secret in CI/CD

### Option 3: IP-Based Authentication
**Implementation**: Whitelist IP addresses instead of shared secret

**Rejected**: 
- More complex to manage
- Doesn't work well with dynamic IPs (Fly.io)
- Still need secret for external clients

## Implementation

### Code Changes

**Before** (conditional):
```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if cfg.APISecret != "" {  // Conditional!
            provided := r.Header.Get("X-API-Secret")
            if provided != cfg.APISecret {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}
```

**After** (mandatory):
```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        provided := r.Header.Get("X-API-Secret")
        if provided != cfg.APISecret {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### Configuration Changes

**Startup validation**:
```go
func Load() (*Config, error) {
    apiSecret := os.Getenv("API_SECRET")
    if apiSecret == "" {
        return nil, errors.New("API_SECRET environment variable is required")
    }
    // ... rest of config
}
```

### Test Updates
All tests updated to include `X-API-Secret` header:
```go
req.Header.Set("X-API-Secret", testAPISecret)
```

## Secret Generation

**Recommended**:
```bash
openssl rand -hex 32
```

Results in 64-character hexadecimal string (256 bits of entropy).

## Consequences

### Positive
- **Security by default**: Cannot deploy without authentication
- **Simpler code**: Removed ~50 lines of conditional logic
- **Better tests**: Tests always run in authenticated mode
- **Consistent behavior**: No environment-dependent behavior
- **Fail fast**: App won't start without secret

### Negative
- **Setup requirement**: Must generate and configure secret
  - *Mitigation*: Documented in README, simple command to generate
- **Test changes**: All tests needed auth headers added
  - *Impact*: One-time change, makes tests more realistic

### Neutral
- **Secret management**: Need to manage secret in environments
  - Use Fly.io secrets for production
  - Use `.env` or export for local development
  - Use GitHub Secrets for CI/CD

## Security Considerations

### Secret Storage
- **Production**: Fly.io secrets (`flyctl secrets set API_SECRET=...`)
- **Local**: `.env` file (gitignored) or shell export
- **CI/CD**: GitHub repository secrets

### Secret Rotation
1. Generate new secret
2. Update in all environments (Fly.io, local, CI/CD)
3. Restart services
4. Update any external clients

### Secret Strength
- Minimum 32 bytes (256 bits)
- Use cryptographically secure random generator
- Treat as sensitive credential (never commit to git)

## Related Commits

- `2c5f995` - Remove API_SECRET conditional branches - require authentication
- `47ab1cf` - Update TODO: Mark authentication as fully enabled
- `51c1598` - Implement shared secret API authentication
- `e82220a` - Add critical security TODO for API authentication

## Testing

All tests updated and passing:
- Unit tests include `X-API-Secret` header
- Integration tests include authentication
- E2E tests verify auth works correctly
- Auth middleware tests cover unauthorized cases

## Future Enhancements

If more sophisticated auth is needed:
1. **JWT tokens**: For user-specific authentication
2. **OAuth 2.0**: For third-party integrations
3. **API keys per client**: For tracking usage
4. **Rate limiting**: Per-key rate limits

**Current status**: Shared secret is sufficient for current needs

## Notes

This change embodies the "secure by default" principle. While it adds a small amount of setup friction, it eliminates an entire class of security vulnerabilities (accidental unauth deployment).

The simplification of removing conditional logic was an unexpected benefit - the code is now easier to understand and maintain.
