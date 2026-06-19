# Monitoring and health checks

## Health endpoints

All Go services and frontend nginx expose:

| Endpoint | Purpose | Usage |
|----------|---------|-------|
| `/liveness` | Process alive | Kubernetes livenessProbe |
| `/readiness` | DB connected | Kubernetes readinessProbe |
| `/health` | Alias for readiness | Docker healthcheck, load balancers |
| `/metrics` | Prometheus metrics | Scraping (direct :8081/:8082/:8083 only) |

### Manual check

```bash
# Auth
curl http://localhost:8081/liveness
curl http://localhost:8081/readiness

# Server
curl http://localhost:8082/liveness
curl http://localhost:8082/readiness

# Sync
curl http://localhost:8083/liveness
curl http://localhost:8083/readiness

# Frontend (prod nginx)
curl http://localhost:8080/liveness
```

## Kubernetes probes

Configured in Helm values (enabled by default):

```yaml
readinessProbe:
  enabled: true
  path: /readiness
  initialDelaySeconds: 5
  periodSeconds: 10

livenessProbe:
  enabled: true
  path: /liveness
  initialDelaySeconds: 10
  periodSeconds: 30
```

## Prometheus metrics

Endpoint: `/metrics` on each Go service.

Example ServiceMonitor (bring your own):

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: treepage-backend
spec:
  selector:
    matchLabels:
      app.kubernetes.io/part-of: treepage
  endpoints:
    - port: http
      path: /metrics
      interval: 30s
```

## Logging

| Service | Level (default) | Configuration |
|---------|-----------------|---------------|
| auth | info | Helm: `auth.logging.level` |
| server | info | Helm: `server.logging.level` |
| sync | info | Helm: `sync.logging.level` |

UI override: **System settings** → **Platform** → **Log level**

### Audit log

When `enableAuditLog: true` (Helm: `server.security.enableAuditLog`):

- Admin actions are recorded in the `audit_log` table
- Fields: user_id, action, resource, timestamp, details

```sql
SELECT * FROM audit_log ORDER BY created_at DESC LIMIT 20;
```

## Sync monitoring

```sql
-- Latest sync jobs
SELECT r.name, j.status, j.started_at, j.finished_at, j.error
FROM sync_jobs j
JOIN repositories r ON r.id = j.repository_id
ORDER BY j.started_at DESC
LIMIT 10;

-- Repositories with errors
SELECT name, last_sync_status, last_sync_error, last_sync_at
FROM repositories
WHERE last_sync_status = 'failed';
```

## Resource usage

Recommended limits (from Helm defaults):

| Service | CPU limit | Memory limit |
|---------|-----------|--------------|
| auth | 200m | 256Mi |
| server | 500m | 512Mi |
| sync | 500m | 512Mi |
| frontend | 200m | 256Mi |

## Alerts (recommendations)

| Alert | Condition |
|-------|-----------|
| Pod not ready | readinessProbe failing > 5 min |
| Sync failures | last_sync_status = failed |
| High error rate | 5xx responses > threshold |
| DB connection | readiness failing on all backends |
| PVC usage | sync PVC > 80% |

## Related sections

- [Troubleshooting](troubleshooting.md)
- [Architecture](../reference/architecture.md)
