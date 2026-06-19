#!/usr/bin/env bash
# Deploy TreePage dev stack with visible build progress.
# Usage: ./scripts/deploy-dev.sh [docker compose build/up args...]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"
export COMPOSE_BAKE="${COMPOSE_BAKE:-false}"

echo "=== TreePage deploy (compose file: $COMPOSE_FILE) ==="
echo ""
echo "Build notes:"
echo "  - Step 'go build' takes 1-3 minutes per backend (CPU-bound, not frozen)"
echo "  - With -v you will see Go packages scroll during compile"
echo "  - Three backends may build in parallel — slower on small VMs"
echo ""

if [[ "${1:-}" == "build-only" ]]; then
  shift
  echo ">>> docker compose build $*"
  docker compose -f "$COMPOSE_FILE" build "$@"
  exit 0
fi

echo ">>> docker compose build $*"
docker compose -f "$COMPOSE_FILE" build "$@"

echo ""
echo ">>> docker compose up -d"
docker compose -f "$COMPOSE_FILE" up -d

echo ""
echo ">>> status"
docker compose -f "$COMPOSE_FILE" ps

echo ""
echo "Done. UI: http://$(hostname -I 2>/dev/null | awk '{print $1}' || echo localhost):5173"
