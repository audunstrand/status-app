# ADR 002: Architecture Consolidation

**Date**: 2025-12-07  
**Status**: Accepted  
**Deciders**: Audun

## Context

The initial architecture consisted of 5 separate services and 2 databases:
- **Services**: Commands, API, Projections, Slackbot, Scheduler
- **Databases**: Events database, Projections database

This architecture had several operational challenges:
- Deployment complexity (5 separate Fly.io apps)
- Inter-service communication overhead
- Increased monitoring and logging complexity
- Higher operational costs
- More complex testing setup

The services were logically separated but tightly coupled in practice, with Commands, API, and Projections all working on the same event stream and projection data.

## Decision

**Consolidate to 3 services and 1 database:**
- **Backend service**: Combines Commands + API + Projections
- **Slackbot service**: Slack integration (unchanged)
- **Scheduler service**: Reminder scheduling (unchanged)
- **Single PostgreSQL database** with two schemas: `events` and `projections`

## Options Considered

### Option 1: Keep Separate (Status Quo)
**Pros**:
- Clear service boundaries
- Independent scaling
- True microservices architecture

**Cons**:
- High operational complexity
- Deployment overhead
- Network latency between services
- More moving parts to monitor

### Option 2: Consolidate Databases Only
**Pros**:
- Reduced database management
- Simplified connection management

**Cons**:
- Still 5 services to deploy and monitor
- Doesn't address service complexity

### Option 3: Full Consolidation ✅ **CHOSEN**
**Pros**:
- Significantly reduced operational complexity
- Faster inter-component communication (in-process)
- Simpler deployment pipeline
- Lower infrastructure costs
- Easier local development
- Maintains logical separation via schemas

**Cons**:
- Cannot scale Commands/API/Projections independently
- Larger service footprint
- Slightly couples three concerns

## Implementation

### Phase 1: Database Consolidation
- Migrated from 2 databases → 1 database
- Used PostgreSQL schemas (`events`, `projections`) for logical separation
- Created migrations: 003 (events schema), 004 (projections schema)
- Updated all services to use single database connection

### Phase 2: Service Consolidation
- Created new `cmd/backend` combining three services
- Shared single HTTP server on port 8080
- Commands write events, API reads projections, Projections listen to events
- Removed obsolete service directories

### Phase 3: Deployment
- Updated Fly.io configuration (`fly.backend.toml`)
- Updated CI/CD pipeline
- Verified all tests passing
- Successfully deployed to production

## Consequences

### Positive
- **Operational simplicity**: 3 services instead of 5
- **Faster communication**: In-process instead of HTTP
- **Cost reduction**: Fewer Fly.io apps running
- **Simpler testing**: Single service to test
- **Better developer experience**: Less context switching

### Negative
- **Scaling granularity**: Cannot scale Commands/API/Projections independently
  - *Mitigation*: Current traffic doesn't require independent scaling
- **Service coupling**: Three concerns in one deployment unit
  - *Mitigation*: Clear internal boundaries maintained

### Neutral
- **Database schemas**: Logical separation maintained
- **Slackbot and Scheduler**: Remain independent (correct isolation)

## Related Commits

- `27fee10` - Complete Phase 1: Database consolidation
- `3d86ccf` - Add backend service combining commands, api, and projections
- `b3650f7` - Complete Phase 2: Service consolidation
- `840b04f` - CONSOLIDATION COMPLETE! Final verification and summary

## Notes

The consolidation was done incrementally with verification at each step. All tests were updated and passing throughout the process. The architecture now follows a more pragmatic "modular monolith" pattern while retaining the benefits of event sourcing and CQRS.
