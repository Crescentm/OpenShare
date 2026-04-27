#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test/docker-lib.sh
source "$SCRIPT_DIR/docker-lib.sh"

openshare_test_export_env
openshare_test_compose_cmd

"${DOCKER_COMPOSE[@]}" down "$@"

echo "[test/down] docker test stack stopped"
echo "[test/down] runtime data is kept at $OPENSHARE_TEST_RUNTIME_DIR"
echo "[test/down] pass --volumes if you also want docker compose to remove named volumes"
