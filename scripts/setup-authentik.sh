#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
target="$root/deployments/authentik/.env"
if [ -e "$target" ] || [ -L "$target" ]; then
    printf '%s\n' 'Existing Authentik .env preserved.'
    exit 0
fi
command -v openssl >/dev/null 2>&1 || { printf '%s\n' 'openssl is required.' >&2; exit 1; }
umask 077
# noclobber also prevents overwriting a file created by another setup process.
set -C
{
    printf 'PG_DB=authentik\nPG_USER=authentik\n'
    printf 'PG_PASS=%s\n' "$(openssl rand -hex 32)"
    printf 'AUTHENTIK_SECRET_KEY=%s\n' "$(openssl rand -hex 60)"
    printf 'AUTHENTIK_IMAGE=ghcr.io/goauthentik/server\nAUTHENTIK_TAG=2026.8.5\n'
    printf 'COMPOSE_PORT_HTTP=9000\nCOMPOSE_PORT_HTTPS=9443\n'
} > "$target"
printf '%s\n' 'Created deployments/authentik/.env (mode 600). Keep this file private.'
