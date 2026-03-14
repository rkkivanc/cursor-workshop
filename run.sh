#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

require_cmd docker

if ! docker info >/dev/null 2>&1; then
  echo "Docker is not running or not accessible (docker info failed)." >&2
  echo "Start Docker Desktop (or your Docker daemon) and try again." >&2
  exit 1
fi

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
  else
    echo "Docker Compose not found. Install Docker Compose v2 (recommended)." >&2
    exit 1
  fi
}

wait_for_health() {
  local service="$1"
  local timeout_s="${2:-90}"
  local started
  started="$(date +%s)"

  local cid
  cid="$(compose ps -q "$service" 2>/dev/null || true)"
  if [[ -z "${cid}" ]]; then
    echo "Service '$service' is not running (no container id)." >&2
    return 1
  fi

  while true; do
    local status
    status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}no-healthcheck{{end}}' "$cid" 2>/dev/null || true)"
    case "$status" in
      healthy) return 0 ;;
      no-healthcheck)
        echo "Service '$service' has no healthcheck; skipping wait."
        return 0
        ;;
      unhealthy)
        echo "Service '$service' is unhealthy. Recent logs:" >&2
        compose logs --no-color --tail=200 "$service" >&2 || true
        return 1
        ;;
    esac

    local now elapsed
    now="$(date +%s)"
    elapsed=$((now - started))
    if (( elapsed > timeout_s )); then
      echo "Timed out waiting for '$service' to become healthy (${timeout_s}s)." >&2
      compose logs --no-color --tail=200 "$service" >&2 || true
      return 1
    fi

    sleep 2
  done
}

cmd="${1:-up}"
shift || true

case "$cmd" in
  up)
    compose up -d --build "$@"
    wait_for_health postgres 90
    echo "All set."
    echo "Backend:  http://localhost:8080"
    echo "Frontend: http://localhost:3000"
    ;;
  down)
    compose down "$@"
    ;;
  restart)
    compose restart "$@"
    ;;
  logs)
    compose logs -f "$@"
    ;;
  ps)
    compose ps "$@"
    ;;
  *)
    echo "Usage: ./run.sh {up|down|restart|logs|ps} [compose args...]" >&2
    exit 2
    ;;
esac

