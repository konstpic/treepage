#!/usr/bin/env bash
# Deploy TreePage production compose stack (static frontend, required secrets in .env.prod).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.prod.yml}"
ENV_FILE="${ENV_FILE:-.env.prod}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Missing $ENV_FILE — copy from .env.prod.example and fill secrets."
  exit 1
fi

export COMPOSE_BAKE="${COMPOSE_BAKE:-false}"

echo "=== TreePage PROD deploy ($COMPOSE_FILE, env: $ENV_FILE) ==="
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" build "$@"
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" ps
echo "Done."
