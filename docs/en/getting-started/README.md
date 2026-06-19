# Quick start

After [installing](../installation/README.md) TreePage, complete basic setup in 10–15 minutes.

## Steps

```
1. First login          →  admin@local / admin (local super_admin)
2. Welcome space        →  /spaces/welcome — docs from docs/
3. System settings      →  language, theme, OIDC
4. Your own spaces      →  if needed
5. Git + sync           →  for your documentation
6. Users and RBAC       →  groups, roles
```

## Section documents

| Step | Document |
|------|----------|
| 1 | [First login](first-login.md) |
| 2 | [Welcome space](welcome-space.md) |
| 3–6 | [Initial setup](initial-setup.md) |

## Roles during setup

| Task | Minimum role |
|------|--------------|
| First login (dev) | — |
| System settings, OIDC | `super_admin` |
| Creating spaces and repositories | `admin` |
| Git synchronization | `admin` or `editor` |
| Reading documentation | `viewer` or public access |

## Production readiness checklist

- [ ] Secrets changed (`JWT_SECRET`, `DB_PASSWORD`)
- [ ] OIDC provider configured
- [ ] Dev mode disabled (`DEV_MODE=false`, `ENV=prod`)
- [ ] Welcome space available (`/spaces/welcome`)
- [ ] First synchronization completed
- [ ] Roles assigned to users
- [ ] TLS configured on Ingress
