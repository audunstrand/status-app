# Verifying Slack Bot Integration

## Pre-requisites
- Slack bot is installed in your workspace
- Fly.io secrets are set (SLACK_BOT_TOKEN, SLACK_SIGNING_SECRET, API_SHARED_SECRET)
- Bot is deployed and running

## Verification Steps

### 1. Check Bot Deployment Status
```bash
fly status -a status-bot
fly logs -a status-bot
```

### 2. Verify Slack App Configuration
In your [Slack App Settings](https://api.slack.com/apps):

1. **Event Subscriptions**
   - Enabled: ✅
   - Request URL: `https://status-bot.fly.dev/slack/events`
   - Should show "Verified ✓"
   - Subscribe to bot events:
     - `app_mention` - When someone @mentions the bot
     - `message.im` - Direct messages to the bot

2. **OAuth & Permissions**
   - Bot Token Scopes needed:
     - `chat:write` - Send messages
     - `app_mentions:read` - Read mentions
     - `im:history` - Read DM history
     - `im:write` - Send DMs

3. **Interactivity & Shortcuts**
   - Enabled: ✅
   - Request URL: `https://status-bot.fly.dev/slack/events`

### 3. Test the Bot

#### Test 1: Direct Message
1. In Slack, find your bot in the Apps section
2. Send a direct message: `Hello`
3. Expected: Bot should respond (check logs if no response)

#### Test 2: Channel Mention
1. Invite the bot to a channel: `/invite @YourBotName`
2. Mention the bot: `@YourBotName hello`
3. Expected: Bot should respond

#### Test 3: Submit Status Update
1. Send to bot: `update`
2. Expected: Bot should prompt for status
3. Reply with a status message
4. Expected: Bot should confirm and send to API

### 4. Check Logs for Debugging

```bash
# Real-time logs
fly logs -a status-bot

# Check for errors
fly logs -a status-bot | grep -i error

# Check API connectivity
fly logs -a status-bot | grep -i "api"
```

### 5. Common Issues

**Bot doesn't respond:**
- Check Event Subscriptions URL is verified
- Check fly logs for errors
- Verify SLACK_SIGNING_SECRET matches Slack app

**"Failed to send event" errors:**
- Check API_SHARED_SECRET matches between bot and API
- Verify API is running: `fly status -a status-api`
- Check API URL in bot code (should be internal: `http://status-api.internal:8080`)

**Authorization errors:**
- Reinstall bot to workspace to refresh token
- Verify SLACK_BOT_TOKEN in Fly secrets

### 6. Manual Testing with curl

Test the event endpoint (requires valid Slack signature):
```bash
fly ssh console -a status-bot
# Inside the container
curl -v http://localhost:8080/health
```

## Next Steps After Verification

Once bot responds:
1. Test status submission flow end-to-end
2. Verify events are stored in API database
3. Test weekly scheduling (TODO item)
4. Set up team registration
