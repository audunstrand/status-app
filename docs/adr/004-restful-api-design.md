# ADR 004: RESTful API Design Over CQRS-Exposing URLs

**Date**: 2025-12-09  
**Status**: Accepted  
**Deciders**: Audun

## Context

The initial API design exposed internal CQRS (Command Query Responsibility Segregation) implementation details through URL structure:
- `POST /commands/submit-update`
- `POST /commands/register-team`
- `GET /api/teams`
- `GET /api/updates`

This approach had several issues:
1. **Leaky abstraction**: URLs exposed implementation details (CQRS pattern)
2. **Non-standard**: Not following REST conventions
3. **Not intuitive**: API consumers had to understand CQRS
4. **Inconsistent**: Mixed `/commands/` and `/api/` prefixes
5. **Team ID location**: Passed in request body instead of URL path

## Decision

**Refactor to resource-based RESTful URLs that hide implementation details.**

Map commands and queries to standard REST operations on resources:
- Commands → `POST`, `PUT`, `DELETE` on resources
- Queries → `GET` on resources
- Team ID as path parameter instead of body field

## URL Mapping

### Teams Resource
| Old URL | New URL | Method | Description |
|---------|---------|--------|-------------|
| `POST /commands/register-team` | `POST /teams` | POST | Register team |
| `GET /api/teams` | `GET /teams` | GET | List all teams |
| `GET /api/teams/:id` | `GET /teams/{id}` | GET | Get team details |
| `PUT /commands/update-team-name` | `PUT /teams/{id}/name` | PUT | Update team name |

### Updates Resource
| Old URL | New URL | Method | Description |
|---------|---------|--------|-------------|
| `POST /commands/submit-update` | `POST /teams/{id}/updates` | POST | Submit update |
| `GET /api/updates` | `GET /updates` | GET | Recent updates |
| N/A | `GET /teams/{id}/updates` | GET | Team updates |

### Request Body Changes
**Old** (team_id in body):
```json
{
  "team_id": "C123456",
  "content": "Status update"
}
```

**New** (team_id in URL):
```
POST /teams/C123456/updates
{
  "content": "Status update"
}
```

## Options Considered

### Option 1: Keep CQRS-Exposing URLs (Status Quo)
**Pros**:
- Explicit about architecture
- Clear separation of commands and queries

**Cons**:
- Leaks implementation details
- Not RESTful
- Harder to understand for API consumers
- Non-standard URL structure

### Option 2: RESTful Resource URLs ✅ **CHOSEN**
**Pros**:
- Standard REST conventions
- Hides implementation details
- Intuitive for API consumers
- Better resource modeling
- Team ID in URL path (REST best practice)

**Cons**:
- Less explicit about CQRS architecture
- Requires updating all clients
- Need to update tests

### Option 3: Hybrid Approach
Keep CQRS URLs but add REST aliases

**Rejected**: Added complexity without benefits

## Implementation

### Backend Changes
1. Updated route definitions in `cmd/backend/main.go`
2. Modified handlers to extract team_id from URL path
3. Updated request validation

### Test Updates
1. Updated all E2E tests
2. Updated integration tests
3. Updated authentication middleware tests
4. Updated Docker E2E tests

### Client Updates
1. Updated Slackbot to use new endpoints
2. Updated Scheduler (no changes needed)

## Consequences

### Positive
- **Better API design**: Follows REST conventions
- **Intuitive**: Clear resource hierarchy (`/teams/{id}/updates`)
- **Implementation hiding**: CQRS is internal detail
- **Standard patterns**: Familiar to API consumers
- **Better URL semantics**: Team ID in path shows ownership

### Negative
- **Migration effort**: All clients needed updates
  - *Impact*: Small codebase, completed quickly
- **Test updates**: All API tests needed updating
  - *Impact*: Tests now more readable with REST URLs

### Neutral
- **Internal CQRS**: Still using CQRS internally (unchanged)
- **Handler logic**: Minimal changes, mostly routing

## Design Principles Applied

1. **Hide implementation details**: API consumers don't need to know about CQRS
2. **Resource-oriented**: URLs represent resources, not operations
3. **HTTP verbs**: Use POST/GET/PUT/DELETE for operations
4. **Path parameters**: IDs in URL path, not request body
5. **Consistent structure**: All endpoints follow same pattern

## Related Commits

- `9e20f74` - Refactor URL structure to RESTful endpoints

## Testing

All tests updated and passing:
- Unit tests: ✅
- Integration tests: ✅
- E2E tests: ✅
- Docker E2E tests: ✅
- Auth middleware tests: ✅

## Notes

This refactoring improved API usability without changing internal architecture. CQRS is still used internally (commands write events, queries read projections), but this is now an implementation detail hidden from API consumers.

The RESTful design also makes it easier to add new endpoints following established patterns.
