#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root/deployments/authentik"
if [ ! -f .env ]; then
    printf '%s\n' 'Run make setup first.' >&2
    exit 1
fi
if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    exec docker compose --project-name tenantledger-authentik --env-file .env -f compose.yml "$@"
elif command -v podman-compose >/dev/null 2>&1; then
    exec podman-compose --project-name tenantledger-authentik --env-file .env -f compose.yml "$@"
else
    printf '%s\n' 'Install Docker with Compose v2 or Podman with podman-compose.' >&2
    exit 1
fi
