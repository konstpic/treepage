# TreePage operations

Documentation for DevOps and platform operators.

## Sections

| Topic | Document |
|-------|----------|
| Service configuration | [Configuration](configuration.md) |
| Secrets and env vars | [Secrets](secrets.md) |
| Database migrations | [Migrations](migrations.md) |
| Backup / restore | [Backup and restore](backup-restore.md) |
| Health checks and metrics | [Monitoring](monitoring.md) |
| Problem diagnosis | [Troubleshooting](troubleshooting.md) |
| Roadmap (phases 1–3) | [Roadmap](../reference/roadmap.md) |

## Components

| Service | Port | Required |
|---------|------|:--------:|
| frontend | 5173 (dev) / 80 (prod) | ✅ |
| backend-auth | 8081 | ✅ |
| backend-server | 8082 | ✅ |
| backend-sync | 8083 | ✅ |
| PostgreSQL | 5432 | ✅ |
| Redis | 6379 | ❌ |

## Configuration load order

```
1. YAML: /opt/app/conf/config.yml
2. Environment variables (override)
3. Validation → fail fast
```

## Production checklist

- [ ] Strong secrets (not defaults)
- [ ] `DEV_MODE=false`, `ENV=prod`
- [ ] OIDC configured
- [ ] TLS on Ingress
- [ ] Migrations applied
- [ ] PostgreSQL backup configured
- [ ] Health endpoints monitored
- [ ] PVC for sync (≥ 2Gi)
- [ ] Resource limits in Helm values

## Related sections

- [Kubernetes / Helm](../installation/kubernetes.md)
- [Architecture](../reference/architecture.md)
