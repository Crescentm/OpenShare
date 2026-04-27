#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test/docker-lib.sh
source "$SCRIPT_DIR/docker-lib.sh"

openshare_test_export_env
openshare_test_write_config
openshare_test_compose_cmd

"${DOCKER_COMPOSE[@]}" up -d --build

openshare_test_wait_http "Meilisearch" "http://127.0.0.1:${MEILI_HTTP_PORT}/health" 60
openshare_test_wait_http "OpenShare" "http://127.0.0.1:${OPENSHARE_HTTP_PORT}/healthz" 90

admin_line="$("${DOCKER_COMPOSE[@]}" logs --tail=120 openshare 2>/dev/null \
  | grep -E '\[bootstrap\] super admin initialized; username=.* password=.*' \
  | tail -n 1 || true)"

echo
echo "[test/up] docker test stack is up"
echo "  Public       : http://127.0.0.1:${OPENSHARE_HTTP_PORT}/"
echo "  Admin        : http://127.0.0.1:${OPENSHARE_HTTP_PORT}/admin"
echo "  API Health   : http://127.0.0.1:${OPENSHARE_HTTP_PORT}/healthz"
echo "  Meilisearch  : http://127.0.0.1:${MEILI_HTTP_PORT}/"
echo "  Runtime data : $OPENSHARE_TEST_RUNTIME_DIR"
echo

if [[ -n "$admin_line" ]]; then
  echo "[test/up] initial super admin:"
  echo "$admin_line"
  echo
fi

echo "[test/up] useful commands:"
echo "  ./test/check.sh"
echo "  ${DOCKER_COMPOSE[*]} logs -f openshare openshare-worker"
echo "  ./test/down.sh"
