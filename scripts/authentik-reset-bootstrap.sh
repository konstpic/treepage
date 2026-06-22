#!/usr/bin/env bash
# Reset Authentik bootstrap admin password to match .env (bootstrap runs only on fresh DB).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml:docker-compose.authentik.yml}"
if [[ -f .env.authentik ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env.authentik
  set +a
fi
if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

EMAIL="${AUTHENTIK_BOOTSTRAP_EMAIL:-admin@authentik.local}"
PASS="${AUTHENTIK_BOOTSTRAP_PASSWORD:-authentik}"

compose_args=()
IFS=':' read -r -a compose_files <<< "$COMPOSE_FILE"
for f in "${compose_files[@]}"; do
  compose_args+=(-f "$f")
done

if ! docker compose "${compose_args[@]}" ps authentik-server --status running &>/dev/null; then
  echo "authentik-server is not running — skip bootstrap password reset"
  exit 0
fi

echo ">>> Reset Authentik bootstrap admin ($EMAIL)"

docker compose "${compose_args[@]}" exec -T authentik-server ak shell <<EOF
from authentik.core.models import User

email = "${EMAIL}"
password = "${PASS}"
user = User.objects.filter(email=email).first() or User.objects.filter(username="akadmin").first()
if user is None:
    raise SystemExit("Bootstrap admin not found")
user.set_password(password)
user.is_active = True
user.save()
print(f"Password reset for {user.username} ({user.email})")
EOF
