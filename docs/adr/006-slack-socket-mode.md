# ADR 006: Slack Socket Mode vs HTTP Webhooks

**Date**: 2025-11-20 (Initial Implementation)  
**Status**: Accepted  
**Deciders**: Audun

## Context

Slack apps can receive events and interactions through two primary mechanisms:

1. **HTTP Webhooks**: Slack sends HTTP POST requests to a public URL
2. **Socket Mode**: App opens WebSocket connection to Slack

We needed to choose an approach for receiving:
- Message events (mentions, direct messages)
- Slash command interactions
- Interactive component events (modals, buttons)

## Decision

**Use Slack Socket Mode for all Slack event handling.**

## Options Considered

### Option 1: HTTP Webhooks
**Implementation**: Expose public HTTPS endpoint, handle incoming POST requests

**Pros**:
- Standard webhook pattern
- Stateless (no persistent connection)
- Can scale horizontally easily
- Works with serverless/edge functions
- Native to Slack's original architecture

**Cons**:
- Requires public HTTPS endpoint
- Need to handle webhook verification
- Must implement request signature validation
- Requires URL configuration in Slack App settings
- More complex local development (ngrok/tunneling)
- Need to manage webhook retry logic
- Requires deploying webhook endpoint first

**Implementation Complexity**:
```go
// Need to:
// 1. Verify request signatures
// 2. Handle challenge requests
// 3. Return 200 OK immediately
// 4. Process events asynchronously
// 5. Handle retries and deduplication
```

### Option 2: Socket Mode ✅ **CHOSEN**
**Implementation**: Open WebSocket connection to Slack, receive events over socket

**Pros**:
- No public endpoint required
- Built-in authentication (App-Level Token)
- Simpler local development (no tunneling needed)
- Official Go library support (`slack-go/slack/socketmode`)
- Automatic reconnection handling
- No signature verification needed
- Works behind firewalls/NAT
- Easy to run locally for testing

**Cons**:
- Requires persistent connection
- Single instance limitation (can't load balance events)
- More stateful (connection management)
- WebSocket connection overhead
- Not suitable for serverless

**Implementation Simplicity**:
```go
// Simple setup:
client := socketmode.New(api)
for evt := range client.Events {
    // Handle events
}
```

## Rationale

**Socket Mode was chosen** because:

1. **Development Experience**: Dramatically simpler local development
   - No ngrok or tunneling needed
   - Just run the app locally
   - Slack events work immediately

2. **Deployment Simplicity**: No webhook URL management
   - Don't need to update Slack App config per environment
   - No coordination between deployment and configuration

3. **Security**: Less attack surface
   - No public endpoint to secure
   - No signature verification code to maintain
   - Built-in authentication via App-Level Token

4. **Scale Appropriate**: Our scale doesn't need multi-instance load balancing
   - Single Slackbot instance handles current load
   - Can revisit if scaling needs change

5. **Official Support**: First-class support in `slack-go` library
   - Well-maintained
   - Good documentation
   - Active community

## Implementation

### Configuration
```go
api := slack.New(
    cfg.SlackBotToken,           // Bot User OAuth Token
    slack.OptionAppLevelToken(cfg.SlackSigningKey),  // App-Level Token
)

client := socketmode.New(api)
```

### Event Handling
```go
for evt := range client.Events {
    switch evt.Type {
    case socketmode.EventTypeEventsAPI:
        // Handle message events
    case socketmode.EventTypeSlashCommand:
        // Handle slash commands
    case socketmode.EventTypeInteractive:
        // Handle interactive components
    }
    client.Ack(*evt.Request)
}
```

### Connection Management
- Automatic reconnection on disconnect
- Built-in heartbeat/ping-pong
- Error handling via callbacks

## Consequences

### Positive
- **Simple development**: No tunneling for local testing
- **Less infrastructure**: No public endpoint needed
- **Built-in reconnection**: Handles network issues automatically
- **Cleaner code**: No signature verification complexity

### Negative
- **Single instance**: Cannot load balance across multiple instances
  - *Impact*: Acceptable for current scale
  - *Mitigation*: Can switch to webhooks if scaling needs change
- **Persistent connection**: Requires long-running process
  - *Impact*: Already running long-lived service (not serverless)

### Neutral
- **Different debugging**: Connection-based debugging vs HTTP logs
- **Monitoring**: Monitor WebSocket health vs HTTP endpoint health

## When to Reconsider

Consider switching to HTTP webhooks if:
1. Need to scale to multiple Slackbot instances
2. Moving to serverless/edge architecture
3. Want fully stateless deployment
4. Need to share event processing across multiple systems

**Current Status**: Socket Mode meets all requirements, no need to change

## Related Configuration

**Slack App Settings**:
- Socket Mode: Enabled
- App-Level Token: Required (starts with `xapp-`)
- OAuth Scopes: `app_mentions:read`, `chat:write`, `commands`
- Event Subscriptions: Delivered via Socket Mode (not HTTP)

## Testing

Socket Mode works well with testing:
- Can run bot locally against production Slack workspace
- No special test infrastructure needed
- Events received in real-time for manual testing

## Related Commits

- Initial implementation in early project commits
- `6e9e2ba` - Implement Slack event handlers for status updates
- `0a192fb` - Add app_mention event handler for @bot mentions

## Notes

Socket Mode is particularly well-suited for small to medium scale Slack apps. It eliminates much of the complexity around webhook handling while providing a reliable, real-time event delivery mechanism.

For this application, Socket Mode is the right choice and should not be changed unless scaling requirements fundamentally change.
