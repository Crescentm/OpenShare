#!/usr/bin/env bash
set -euo pipefail

OPENSHARE_TEST_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OPENSHARE_TEST_RUNTIME_DIR="${OPENSHARE_TEST_RUNTIME_DIR:-$OPENSHARE_TEST_ROOT/test/.runtime/docker}"
OPENSHARE_TEST_COMPOSE_PROJECT="${OPENSHARE_TEST_COMPOSE_PROJECT:-openshare-test}"

abs_path() {
  local path="$1"
  if [[ "$path" = /* ]]; then
    printf '%s\n' "$path"
  else
    printf '%s\n' "$OPENSHARE_TEST_ROOT/$path"
  fi
}

openshare_test_load_dotenv() {
  local dotenv="$OPENSHARE_TEST_ROOT/.env"
  local line key value

  [[ -f "$dotenv" ]] || return

  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"

    [[ -z "$line" || "${line:0:1}" == "#" ]] && continue

    if [[ "$line" == export\ * ]]; then
      line="${line#export }"
    fi

    key="${line%%=*}"
    value="${line#*=}"
    key="${key%"${key##*[![:space:]]}"}"
    value="${value#"${value%%[![:space:]]*}"}"

    [[ "$key" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || continue
    [[ -n "${!key+x}" ]] && continue

    if [[ "${#value}" -ge 2 ]]; then
      if [[ "${value:0:1}" == '"' && "${value: -1}" == '"' ]]; then
        value="${value:1:${#value}-2}"
      elif [[ "${value:0:1}" == "'" && "${value: -1}" == "'" ]]; then
        value="${value:1:${#value}-2}"
      fi
    fi

    export "$key=$value"
  done <"$dotenv"
}

openshare_test_export_env() {
  openshare_test_load_dotenv

  export OPENSHARE_CONTAINER_NAME="${OPENSHARE_CONTAINER_NAME:-openshare-test}"
  export OPENSHARE_WORKER_CONTAINER_NAME="${OPENSHARE_WORKER_CONTAINER_NAME:-openshare-test-worker}"
  export MEILI_CONTAINER_NAME="${MEILI_CONTAINER_NAME:-openshare-test-meilisearch}"

  export OPENSHARE_HTTP_PORT="${OPENSHARE_HTTP_PORT:-18080}"
  export MEILI_HTTP_PORT="${MEILI_HTTP_PORT:-17700}"

  export OPENSHARE_SESSION_SECRET="${OPENSHARE_SESSION_SECRET:-openshare-local-docker-test-secret-change-me}"
  export OPENSHARE_SEARCH_ENGINE_ENABLED="${OPENSHARE_SEARCH_ENGINE_ENABLED:-true}"
  export OPENSHARE_SEARCH_ENGINE_INDEX_NAME="${OPENSHARE_SEARCH_ENGINE_INDEX_NAME:-openshare_resources}"
  export OPENSHARE_SEARCH_ENGINE_SEMANTIC_PROFILE_PATH="${OPENSHARE_SEARCH_ENGINE_SEMANTIC_PROFILE_PATH:-config/search_semantics.openwhu.json}"
  export OPENSHARE_IMPORTS_CONTAINER_PATH="${OPENSHARE_IMPORTS_CONTAINER_PATH:-/imports}"
  export MEILI_MASTER_KEY="${MEILI_MASTER_KEY:-openshare-development-master-key-change-me}"

  export OPENSHARE_DATA_PATH
  OPENSHARE_DATA_PATH="$(abs_path "${OPENSHARE_DATA_PATH:-$OPENSHARE_TEST_RUNTIME_DIR/data}")"

  export OPENSHARE_IMPORTS_PATH
  OPENSHARE_IMPORTS_PATH="$(abs_path "${OPENSHARE_IMPORTS_PATH:-$OPENSHARE_TEST_RUNTIME_DIR/imports}")"

  export MEILI_DATA_PATH
  MEILI_DATA_PATH="$(abs_path "${MEILI_DATA_PATH:-$OPENSHARE_TEST_RUNTIME_DIR/meili_data}")"

  export OPENSHARE_CONFIG
  OPENSHARE_CONFIG="$(abs_path "${OPENSHARE_CONFIG:-$OPENSHARE_TEST_RUNTIME_DIR/config.local.json}")"
}

openshare_test_write_config() {
  mkdir -p "$OPENSHARE_DATA_PATH" "$OPENSHARE_IMPORTS_PATH" "$MEILI_DATA_PATH" "$(dirname "$OPENSHARE_CONFIG")"

  if [[ -f "$OPENSHARE_CONFIG" ]]; then
    return
  fi

  cat >"$OPENSHARE_CONFIG" <<JSON
{
  "session": {
    "secret": "$OPENSHARE_SESSION_SECRET"
  },
  "search_engine": {
    "enabled": true,
    "host": "http://meilisearch:7700",
    "api_key": "$MEILI_MASTER_KEY",
    "index_name": "$OPENSHARE_SEARCH_ENGINE_INDEX_NAME",
    "semantic_profile_path": "$OPENSHARE_SEARCH_ENGINE_SEMANTIC_PROFILE_PATH"
  }
}
JSON
}

openshare_test_compose_cmd() {
  local compose_file="$OPENSHARE_TEST_ROOT/docker/docker-compose.yml"

  if docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE=(docker compose --project-name "$OPENSHARE_TEST_COMPOSE_PROJECT" -f "$compose_file")
    return
  fi

  if command -v docker-compose >/dev/null 2>&1; then
    DOCKER_COMPOSE=(docker-compose -p "$OPENSHARE_TEST_COMPOSE_PROJECT" -f "$compose_file")
    return
  fi

  echo "[test/docker] docker compose is required" >&2
  return 1
}

openshare_test_wait_http() {
  local name="$1"
  local url="$2"
  local retries="${3:-60}"
  local i

  for ((i = 1; i <= retries; i++)); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "[test/docker] $name is ready: $url"
      return 0
    fi
    sleep 1
  done

  echo "[test/docker] $name not ready after ${retries}s: $url" >&2
  return 1
}
