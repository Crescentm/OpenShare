#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/.env"

if ! docker compose version >/dev/null 2>&1; then
  echo "[up] docker compose is required" >&2
  exit 1
fi

"$SCRIPT_DIR/init-env.sh"

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

(
  cd "$SCRIPT_DIR"
  docker compose up -d
)

echo
echo "[up] OpenShare is starting"
echo "  Public      : http://127.0.0.1:${OPENSHARE_HTTP_PORT:-8080}/"
echo "  Admin       : http://127.0.0.1:${OPENSHARE_HTTP_PORT:-8080}/admin"
echo "  Meilisearch : http://127.0.0.1:${MEILI_HTTP_PORT:-7700}/"
echo
echo "[up] Useful commands:"
echo "  cd \"$SCRIPT_DIR\""
echo "  docker compose logs -f openshare openshare-worker"
echo "  docker compose logs openshare | grep 'super admin initialized'"
