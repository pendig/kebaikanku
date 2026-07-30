#!/usr/bin/env sh
set -eu

ENV_FILE=${ENV_FILE:-.env.production}
COMPOSE_FILE=${COMPOSE_FILE:-docker-compose.production.yml}
BACKUP_DIR=${BACKUP_DIR:-backups}

[ -f "$ENV_FILE" ] || { echo "Missing $ENV_FILE" >&2; exit 1; }
read_env() { sed -n "s/^$1=//p" "$ENV_FILE" | tail -n 1; }
POSTGRES_USER=$(read_env POSTGRES_USER)
POSTGRES_DB=$(read_env POSTGRES_DB)
[ -n "$POSTGRES_USER" ] && [ -n "$POSTGRES_DB" ] || { echo "POSTGRES_USER and POSTGRES_DB are required in $ENV_FILE" >&2; exit 1; }
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/kebaikanku-$(date -u +%Y%m%dT%H%M%SZ).dump"

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres \
	pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc >"$BACKUP_FILE"
echo "Backup written to $BACKUP_FILE"
