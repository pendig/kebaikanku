#!/usr/bin/env sh
set -eu

ENV_FILE=${ENV_FILE:-.env.production}
COMPOSE_FILE=${COMPOSE_FILE:-docker-compose.production.yml}
BACKUP_FILE=${1:-}

[ -n "$BACKUP_FILE" ] && [ -f "$BACKUP_FILE" ] || { echo "Usage: RESTORE_CONFIRM=restore $0 path/to/backup.dump" >&2; exit 1; }
[ "${RESTORE_CONFIRM:-}" = "restore" ] || { echo "Set RESTORE_CONFIRM=restore after stopping the API." >&2; exit 1; }
[ -f "$ENV_FILE" ] || { echo "Missing $ENV_FILE" >&2; exit 1; }
API_CONTAINER=$(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps -q api)
[ -z "$API_CONTAINER" ] || [ "$(docker inspect -f '{{.State.Running}}' "$API_CONTAINER")" != "true" ] || { echo "Stop the api service before restoring." >&2; exit 1; }
read_env() { sed -n "s/^$1=//p" "$ENV_FILE" | tail -n 1; }
POSTGRES_USER=$(read_env POSTGRES_USER)
POSTGRES_DB=$(read_env POSTGRES_DB)
[ -n "$POSTGRES_USER" ] && [ -n "$POSTGRES_DB" ] || { echo "POSTGRES_USER and POSTGRES_DB are required in $ENV_FILE" >&2; exit 1; }

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres \
	pg_restore --clean --if-exists --no-owner -U "$POSTGRES_USER" -d "$POSTGRES_DB" <"$BACKUP_FILE"
echo "Restore complete. Run migrations, start the API, then verify /readyz."
