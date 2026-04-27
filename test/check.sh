#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test/docker-lib.sh
source "$SCRIPT_DIR/docker-lib.sh"

openshare_test_export_env
openshare_test_compose_cmd

check_url() {
  local name="$1"
  local url="$2"

  if curl -fsS "$url" >/dev/null; then
    echo "[test/check] PASS: $name ($url)"
    return
  fi

  echo "[test/check] FAIL: $name ($url)" >&2
  return 1
}

"${DOCKER_COMPOSE[@]}" ps

check_url "OpenShare health" "http://127.0.0.1:${OPENSHARE_HTTP_PORT}/healthz"
check_url "worker health" "http://127.0.0.1:${OPENSHARE_HTTP_PORT}/healthz/worker"
check_url "Meilisearch health" "http://127.0.0.1:${MEILI_HTTP_PORT}/health"
check_url "public homepage" "http://127.0.0.1:${OPENSHARE_HTTP_PORT}/"

echo "[test/check] all good"
