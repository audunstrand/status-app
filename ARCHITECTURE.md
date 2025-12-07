# Status App Architecture - PostgreSQL Notifications

## The Problem You're Asking About

**You noticed:** The code has `NOTIFY` in postgres_store.go line 48, but **NOBODY IS LISTENING!** 🎯

You're absolutely right! Let me explain what's happening:

## Current Architecture (Simplified)

```
┌─────────────────┐
│  Commands API   │  Receives HTTP commands
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Event Store    │  Writes events to PostgreSQL
│  (postgres)     │  Sends NOTIFY (but nobody listens!)
└────────┬────────┘
         │
         │ Events table
         │
         ▼
┌─────────────────┐
│  Projections    │  Polls database periodically
│   Service       │  OR rebuilds on startup
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Projection DB   │  Read models for API
│  (postgres)     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   API Service   │  Serves HTTP queries
└─────────────────┘
```

## What's Actually Happening

### ❌ What's NOT Happening (Line 48 in postgres_store.go)

```go
// Line 48 - This NOTIFY has no listeners!
_, _ = s.db.ExecContext(ctx, "NOTIFY events, $1", event.ID)
```

This line sends a PostgreSQL NOTIFY, but **nobody is listening** because:

### ✅ What IS Happening (projector.go)

```go
// Line 34 - Subscribe returns an empty channel
eventsCh, err := p.eventStore.Subscribe(ctx, []string{})
```

Looking at `postgres_store.go` line 101-110:

```go
func (s *PostgresStore) Subscribe(ctx context.Context, eventTypes []string) (<-chan *Event, error) {
	// TODO: Implement LISTEN/NOTIFY for real-time event streaming
	// For now, return a simple implementation
	ch := make(chan *Event)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil  // ⚠️ Returns empty channel that never sends events!
}
```

## How It Works Instead

### The Projections Service Starts Up:

1. **Initial Load** (projector.go line 29):
   ```go
   if err := p.rebuildProjections(ctx); err != nil {
       return fmt.Errorf("failed to rebuild projections: %w", err)
   }
   ```
   - Reads ALL events from the database
   - Rebuilds the read models from scratch

2. **Subscribe** (line 34):
   ```go
   eventsCh, err := p.eventStore.Subscribe(ctx, []string{})
   ```
   - Gets an empty channel (stub implementation)
   - Nothing ever comes through this channel!

3. **Process Loop** (line 40-51):
   ```go
   go func() {
       for {
           select {
           case event := <-eventsCh:  // ⚠️ This never happens!
               if err := p.processEvent(ctx, event); err != nil {
                   log.Printf("failed to process event %s: %v", event.ID, err)
               }
           case <-ctx.Done():
               return
           }
       }
   }()
   ```
   - Waits for events that never arrive
   - Just sits there doing nothing after initial rebuild

## The Real Answer

**How do new events get projected?**

**They don't in real-time!** 😅

Currently, new events are only projected when:
1. The projections service **restarts** (rebuilds from all events)
2. Or you manually trigger a rebuild

## The TODO: Real Implementation

To make it work with PostgreSQL NOTIFY/LISTEN, you'd need to implement this:

```go
func (s *PostgresStore) Subscribe(ctx context.Context, eventTypes []string) (<-chan *Event, error) {
    // Import: "github.com/lib/pq"
    
    // Create listener
    listener := pq.NewListener(
        s.connString,
        10*time.Second,
        time.Minute,
        func(ev pq.ListenerEventType, err error) {
            if err != nil {
                log.Printf("Listener error: %v", err)
            }
        },
    )
    
    // Listen to the "events" channel
    if err := listener.Listen("events"); err != nil {
        return nil, err
    }
    
    ch := make(chan *Event)
    
    go func() {
        defer close(ch)
        defer listener.Close()
        
        for {
            select {
            case notification := <-listener.Notify:
                if notification == nil {
                    continue
                }
                
                // Get the event by ID from notification.Extra
                event, err := s.GetByID(ctx, notification.Extra)
                if err != nil {
                    log.Printf("Failed to fetch event: %v", err)
                    continue
                }
                
                ch <- event
                
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return ch, nil
}
```

## Summary

### Current Flow:
```
Command → Event Store → Database
                ↓
         (NOTIFY sent, nobody listening)
                ↓
         (Nothing happens)
         
         
Projections Service starts:
  → Reads ALL events from DB
  → Rebuilds read models
  → Waits for events that never come
```

### What It SHOULD Be:
```
Command → Event Store → Database
                ↓
         PostgreSQL NOTIFY "events"
                ↓
         Projections Service LISTENING
                ↓
         Receives notification
                ↓
         Fetches new event
                ↓
         Updates read model
                ↓
         API serves fresh data
```

## Why It Still Works

Even without real-time projections, the system works because:

1. **Commands Service** writes events to the event store ✅
2. **Projections Service** rebuilds on startup, so eventually gets all events ✅
3. **API Service** reads from projections database ✅

It's just not **real-time**. There's a delay between:
- Writing a status update
- It appearing in the API

To see new updates, you'd need to restart the projections service!

## Want to Fix It?

We can implement the real LISTEN/NOTIFY if you want real-time updates! Just say the word.
