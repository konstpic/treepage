#!/usr/bin/env bash
# Dev stack + Authentik IdP for OIDC testing.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export COMPOSE_FILE="docker-compose.dev.yml:docker-compose.authentik.yml"
export COMPOSE_BAKE="${COMPOSE_BAKE:-false}"

if [[ -f .env.authentik ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env.authentik
  set +a
fi

exec "$ROOT/scripts/deploy-dev.sh" "$@"

echo ""
echo ">>> Authentik bootstrap admin password sync"
chmod +x "$ROOT/scripts/authentik-reset-bootstrap.sh"
"$ROOT/scripts/authentik-reset-bootstrap.sh"
