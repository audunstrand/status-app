# Observability Guide

## Overview

The Status App is fully instrumented with Prometheus metrics for comprehensive observability of the event-sourced architecture.

## Metrics Endpoints

All services expose metrics on `/metrics`:

- **Backend** (port 8080): `http://localhost:8080/metrics`
- **Slackbot** (port 8081): `http://localhost:8081/metrics`
- **Scheduler** (port 8082): `http://localhost:8082/metrics`

## Available Metrics

### Event Store Metrics

**Event Storage:**
- `status_app_events_stored_total{event_type}` - Total events stored by type
- `status_app_events_stored_bytes_total` - Total bytes of event data
- `status_app_events_loaded_total{event_type}` - Total events loaded by type
- `status_app_events_errors_total{operation}` - Event store errors

**Key Queries:**
```promql
# Events per second by type
rate(status_app_events_stored_total[5m])

# Event type distribution
sum by (event_type) (status_app_events_stored_total)

# Event store growth rate
rate(status_app_events_stored_bytes_total[1h])
```

### Projection Metrics

**Projection Health:**
- `status_app_projections_updates_total{projection}` - Projection updates
- `status_app_projections_lag_seconds{projection}` - Lag between event and projection
- `status_app_projections_errors_total{projection}` - Projection errors
- `status_app_projections_processing_duration_seconds` - Processing time histogram

**Key Queries:**
```promql
# Projection lag (should be < 1 second)
status_app_projections_lag_seconds

# p95 projection processing time
histogram_quantile(0.95, rate(status_app_projections_processing_duration_seconds_bucket[5m]))

# Projection update rate
rate(status_app_projections_updates_total[1m])
```

### Slackbot Metrics

**Message Handling:**
- `status_app_slackbot_messages_received_total{type}` - Messages received (mention, direct_message, slash_command)
- `status_app_slackbot_messages_sent_total` - Messages sent
- `status_app_slackbot_commands_handled_total{command}` - Slash commands handled
- `status_app_slackbot_api_calls_total{endpoint,status}` - Slack API calls
- `status_app_slackbot_backend_api_calls_total{endpoint,status}` - Backend API calls
- `status_app_slackbot_errors_total{error_type}` - Slackbot errors

**Key Queries:**
```promql
# Message throughput
rate(status_app_slackbot_messages_received_total[5m])

# Command usage
sum by (command) (status_app_slackbot_commands_handled_total)

# API success rate
sum(rate(status_app_slackbot_backend_api_calls_total{status="success"}[5m])) 
/ 
sum(rate(status_app_slackbot_backend_api_calls_total[5m]))
```

### Scheduler Metrics

**Reminder Delivery:**
- `status_app_scheduler_reminders_scheduled_total` - Reminders scheduled
- `status_app_scheduler_reminders_sent_total{status}` - Reminders sent
- `status_app_scheduler_teams_reminder_count` - Teams that received reminders
- `status_app_scheduler_errors_total{error_type}` - Scheduler errors

**Key Queries:**
```promql
# Reminder success rate
rate(status_app_scheduler_reminders_sent_total{status="success"}[1h])
/ 
rate(status_app_scheduler_reminders_sent_total[1h])

# Teams receiving reminders
status_app_scheduler_teams_reminder_count
```

## Fly.io Integration

Fly.io automatically scrapes metrics from `/metrics` endpoints.

### Accessing Metrics

1. **Via Fly.io Dashboard:**
   ```bash
   flyctl open --app status-app-backend
   # Navigate to "Metrics" tab
   ```

2. **Via Grafana:**
   - Fly.io provides a Grafana instance
   - Access at: `https://fly.io/apps/status-app-backend/metrics`

### Deploying Dashboard

**Option 1: Automated Deployment (Recommended)**

Use the provided deployment scripts:

```bash
# Set credentials
export GRAFANA_URL="https://your-grafana-url"
export GRAFANA_API_KEY="your-api-key"

# Deploy dashboard
python3 scripts/deploy-grafana-dashboard.py
```

**Option 2: GitHub Actions (CI/CD)**

1. Add secrets to GitHub: `GRAFANA_API_KEY`
2. Add variable to GitHub: `GRAFANA_URL`
3. Push changes or trigger manually
4. Dashboard deploys automatically

**Option 3: Manual Import**

1. Go to Fly.io Grafana instance
2. Click "+" → "Import"
3. Upload `docs/grafana-dashboard.json`
4. Select Fly.io Prometheus datasource
5. Click "Import"

**See:** `scripts/README.md` for detailed deployment instructions

## Dashboard Panels

The included Grafana dashboard provides:

1. **Events Stored per Second** - Real-time event ingestion rate
2. **Projection Lag** - Gauge showing projection freshness
3. **Event Type Distribution** - Pie chart of event types
4. **Projection Updates per Minute** - Projection activity
5. **Projection Processing Duration (p95)** - Performance monitoring
6. **Slack Messages** - Message throughput
7. **Slack Commands Handled** - Command usage
8. **Backend API Calls** - API health
9. **Scheduler - Reminders Sent** - Reminder delivery count
10. **Error Rates** - All service errors
11. **Event Store Size** - Storage growth

## Alerting

### Recommended Alerts

**Critical:**
- Projection lag > 5 seconds for 5 minutes
- Event store errors > 5/minute
- Backend API error rate > 10%

**Warning:**
- Projection lag > 1 second for 2 minutes
- Slack API errors > 2/minute
- Scheduler reminder failure > 10%

### Setting Up Alerts in Fly.io

Fly.io supports Prometheus alerting rules:

```yaml
# Example alert rules
groups:
  - name: status_app
    rules:
      - alert: HighProjectionLag
        expr: status_app_projections_lag_seconds > 5
        for: 5m
        annotations:
          summary: "Projection lag is high"
          
      - alert: HighErrorRate
        expr: rate(status_app_events_errors_total[5m]) > 0.1
        for: 5m
        annotations:
          summary: "Event store error rate is high"
```

## Local Development

For local testing with Prometheus and Grafana:

```bash
# Start Prometheus
docker run -d -p 9090:9090 \
  -v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# Start Grafana
docker run -d -p 3000:3000 grafana/grafana

# prometheus.yml example:
scrape_configs:
  - job_name: 'backend'
    static_configs:
      - targets: ['host.docker.internal:8080']
  - job_name: 'slackbot'
    static_configs:
      - targets: ['host.docker.internal:8081']
  - job_name: 'scheduler'
    static_configs:
      - targets: ['host.docker.internal:8082']
```

Then access:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

## Troubleshooting

### Metrics Not Showing

1. **Check endpoint is accessible:**
   ```bash
   curl http://localhost:8080/metrics
   ```

2. **Verify Prometheus scraping:**
   - Go to Prometheus: http://localhost:9090/targets
   - Check target status

3. **Check Fly.io logs:**
   ```bash
   flyctl logs --app status-app-backend
   ```

### High Projection Lag

- Check event store performance
- Verify database connection pool settings
- Look for projection errors in metrics
- Check `status_app_projections_errors_total`

### Missing Metrics

- Ensure all services are deployed
- Check service health: `/health` endpoints
- Verify services are running: `flyctl status --app <app-name>`

## Best Practices

1. **Monitor projection lag continuously** - This is the most critical metric for event sourcing
2. **Set up alerting** - Don't wait for users to report issues
3. **Track error rates** - Catch issues before they impact users
4. **Review dashboards regularly** - Understand normal patterns
5. **Use PromQL for investigation** - Powerful queries for debugging

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Grafana Dashboards](https://grafana.com/docs/grafana/latest/dashboards/)
- [Fly.io Metrics](https://fly.io/docs/reference/metrics/)
