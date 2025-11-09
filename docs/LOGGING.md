# Logging Guide

ActaLog uses a custom logging system that supports both stdout and optional file logging.

## Quick Start

### Development (Default)

By default, logs go to **stdout** (console):

```bash
make run
# Logs appear in terminal
```

### Enable File Logging

Set `LOG_FILE_ENABLED=true` in your `.env` file:

```env
LOG_LEVEL=info
LOG_FILE_ENABLED=true
# Optional: LOG_FILE_PATH=./logs/actalog.log
# Optional: LOG_MAX_SIZE_MB=100
# Optional: LOG_MAX_BACKUPS=3
```

Logs will be written to:
- **stdout** (console) - Always enabled
- **File** - `./logs/actalog.log` (if file logging enabled)

## Configuration Options

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `LOG_FILE_ENABLED` | `false` | Enable file logging |
| `LOG_FILE_PATH` | `./logs/actalog.log` | Path to log file (auto-detected if empty) |
| `LOG_MAX_SIZE_MB` | `100` | Max file size in MB before rotation |
| `LOG_MAX_BACKUPS` | `3` | Number of old log files to keep |

## Log Levels

The logger supports 4 levels:

1. **DEBUG** - Detailed diagnostic information
2. **INFO** - General informational messages (default)
3. **WARN** - Warning messages for potentially harmful situations
4. **ERROR** - Error messages for failures

Only messages at or above the configured level are logged.

## Log Format

All log entries include timestamp, level, and message:

```
2024-11-09 15:04:05 [INFO] Server listening on 0.0.0.0:8080
2024-11-09 15:04:12 [INFO] POST /api/auth/login 200 45ms
2024-11-09 15:04:15 [ERROR] Failed to connect to database: connection refused
```

### HTTP Request Logs

HTTP requests are logged with:
- Method (GET, POST, etc.)
- Path (/api/workouts)
- Status code (200, 404, 500)
- Duration (45ms, 1.2s)

Example:
```
2024-11-09 15:04:12 [INFO] POST /api/workouts 201 123ms
2024-11-09 15:04:15 [INFO] GET /api/movements 200 12ms
```

## File Logging Behavior

### Log File Location

If `LOG_FILE_PATH` is not specified, the log file is created in:

```
<executable-directory>/logs/actalog.log
```

For example:
- Binary at `/usr/local/bin/actalog` → Logs at `/usr/local/bin/logs/actalog.log`
- Binary at `./bin/actalog` → Logs at `./bin/logs/actalog.log`

### Automatic Rotation

When the log file reaches `LOG_MAX_SIZE_MB` (default: 100MB):

1. Current file is renamed with timestamp: `actalog.log.20241109-150405`
2. New empty `actalog.log` is created
3. Old backups beyond `LOG_MAX_BACKUPS` (default: 3) are automatically deleted

Example after rotation:
```
logs/
  actalog.log              # Current log file
  actalog.log.20241109-143022
  actalog.log.20241109-120145
  actalog.log.20241108-180022
```

## Usage Examples

### Development (stdout only)

```env
LOG_LEVEL=debug
LOG_FILE_ENABLED=false
```

```bash
make dev
# All logs appear in terminal for easy debugging
```

### Production (file logging)

```env
LOG_LEVEL=info
LOG_FILE_ENABLED=true
LOG_FILE_PATH=/var/log/actalog/app.log
LOG_MAX_SIZE_MB=100
LOG_MAX_BACKUPS=7
```

```bash
./bin/actalog
# Logs go to both stdout and /var/log/actalog/app.log
```

### Custom Log Location

```env
LOG_FILE_ENABLED=true
LOG_FILE_PATH=/custom/path/to/myapp.log
```

## Viewing Logs

### stdout Logs (Always Available)

```bash
# Follow logs in real-time
make run

# Run in background and redirect
./bin/actalog > app.log 2>&1 &
tail -f app.log
```

### File Logs (When Enabled)

```bash
# View current log file
tail -f ./logs/actalog.log

# View last 100 lines
tail -100 ./logs/actalog.log

# Search for errors
grep ERROR ./logs/actalog.log

# View all log files including rotated ones
cat ./logs/actalog.log.* ./logs/actalog.log | grep "2024-11-09"
```

## Production Recommendations

### Option 1: Let Container/Systemd Handle Logs (Recommended)

**Best practice**: Keep `LOG_FILE_ENABLED=false` and let your deployment platform capture stdout:

#### Docker
```bash
# Logs automatically captured by Docker
docker logs actalog-backend -f

# Or with docker-compose
docker-compose logs -f backend
```

#### Systemd
```bash
# Logs automatically captured by journald
journalctl -u actalog -f
journalctl -u actalog --since "1 hour ago"
```

#### Kubernetes
```bash
kubectl logs deployment/actalog -f
```

**Advantages:**
- No file management needed
- Built-in log rotation
- Integration with log aggregation tools (Loki, ELK, Splunk)
- No disk space issues

### Option 2: Use File Logging with Log Aggregation

If you prefer file logging:

```env
LOG_FILE_ENABLED=true
LOG_FILE_PATH=/var/log/actalog/app.log
LOG_MAX_SIZE_MB=100
LOG_MAX_BACKUPS=7
```

Then use a log shipper to send files to your log aggregation system:
- **Filebeat** → Elasticsearch/Logstash
- **Promtail** → Loki/Grafana
- **Fluentd** → Various backends
- **CloudWatch Agent** → AWS CloudWatch

### Option 3: Hybrid (Both stdout and File)

```env
LOG_FILE_ENABLED=true
```

Logs go to both stdout and file:
- **stdout** → Captured by container/systemd
- **File** → Available for local debugging/backup

## Log Aggregation Integration

For production monitoring, integrate with log aggregation:

### Loki + Grafana

```yaml
# promtail-config.yaml
scrape_configs:
  - job_name: actalog
    static_configs:
      - targets:
          - localhost
        labels:
          job: actalog
          __path__: /var/log/actalog/*.log
```

### ELK Stack

```yaml
# filebeat.yml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/actalog/*.log
    fields:
      app: actalog
output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```

### CloudWatch

```bash
# CloudWatch Agent config
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/actalog/app.log",
            "log_group_name": "/actalog/application",
            "log_stream_name": "{instance_id}"
          }
        ]
      }
    }
  }
}
```

## Troubleshooting

### Logs not appearing in file

1. Check if file logging is enabled:
   ```bash
   grep LOG_FILE_ENABLED .env
   ```

2. Check log directory permissions:
   ```bash
   ls -la logs/
   # Should be writable by the user running actalog
   ```

3. Check for errors in startup logs

### Log file growing too large

Adjust rotation settings:

```env
LOG_MAX_SIZE_MB=50      # Smaller files
LOG_MAX_BACKUPS=5       # Keep more history
```

### Want JSON formatted logs

The current logger uses human-readable format. For JSON logs (better for parsing), consider:
- Using structured logging library (zerolog, zap)
- File an enhancement request

### Disk space issues

Monitor log directory size:

```bash
du -sh logs/
```

Solutions:
1. Reduce `LOG_MAX_SIZE_MB`
2. Reduce `LOG_MAX_BACKUPS`
3. Use log aggregation and disable file logging
4. Set up external log rotation with `logrotate`

## Future Enhancements

Planned improvements:

- [ ] Structured/JSON logging format
- [ ] Request ID tracking across services
- [ ] Configurable log format
- [ ] Performance metrics in logs
- [ ] Log sampling for high-traffic endpoints
- [ ] Integration with OpenTelemetry

## Related Documentation

- [Configuration Guide](../README.md#configuration)
- [Deployment Guide](DEPLOYMENT.md)
- [Architecture](ARCHITECTURE.md)
