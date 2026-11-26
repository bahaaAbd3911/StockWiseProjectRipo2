#!/usr/bin/env bash
# Daily backup of the StockWise database and uploaded images.
#
#   deploy/backup.sh                 # run by hand
#   crontab -e  →  0 3 * * * /opt/stockwise/backend/deploy/backup.sh >> /var/log/stockwise-backup.log 2>&1
#
# What it does:
#   1. dumps the database (pg_dump custom format: compressed, restorable table by table)
#   2. checks the dump can be read back (pg_restore --list) before keeping it
#   3. archives the uploaded images volume
#   4. deletes backups older than KEEP_DAYS
#   5. optionally copies everything off the server with rclone (BACKUP_REMOTE)
#
# Settings (environment variables, all optional):
#   ENV_FILE       .env.production     the same file docker compose uses
#   BACKUP_DIR     /var/backups/stockwise
#   KEEP_DAYS      30
#   BACKUP_REMOTE  e.g. r2:stockwise-backups (an rclone remote; set up with "rclone config")
#   COMPOSE_PROJECT stockwise          prefix of the Docker volumes
#
# Works with the db service of docker-compose.yml, or with a managed database
# when DATABASE_URL is set in the env file (pg_dump then runs in a temporary
# postgres container, so nothing needs to be installed on the server).
set -euo pipefail
# Git Bash on Windows rewrites container paths like /uploads into Windows
# paths; this turns that off (no effect on Linux).
export MSYS_NO_PATHCONV=1

cd "$(dirname "$0")/.."
ENV_FILE=${ENV_FILE:-.env.production}
BACKUP_DIR=${BACKUP_DIR:-/var/backups/stockwise}
KEEP_DAYS=${KEEP_DAYS:-30}
BACKUP_REMOTE=${BACKUP_REMOTE:-}
COMPOSE_PROJECT=${COMPOSE_PROJECT:-stockwise}
PG_IMAGE=postgres:18-alpine

log() { echo "$(date -u +%FT%TZ) backup: $*"; }
fail() { log "FAILED: $*"; exit 1; }

[ -f "$ENV_FILE" ] || fail "$ENV_FILE not found (run from the backend folder or set ENV_FILE)"

# Read one value from the env file without executing it (values may contain < > spaces).
env_value() { grep -E "^$1=" "$ENV_FILE" | tail -n 1 | cut -d= -f2- || true; }

compose() { docker compose --env-file "$ENV_FILE" "$@"; }

umask 077 # backups contain customer data: readable by this user only
mkdir -p "$BACKUP_DIR"
stamp=$(date -u +%Y-%m-%dT%H%MZ)
db_file="$BACKUP_DIR/stockwise-db-$stamp.dump"
uploads_file="$BACKUP_DIR/stockwise-uploads-$stamp.tar.gz"
tmp="$db_file.partial"
uploads_tmp="$uploads_file.partial"
# A failed step never leaves a half-written file that looks like a backup.
trap 'rm -f "$tmp" "$uploads_tmp"' EXIT

# ---- 1 + 2: database ----
database_url=$(env_value DATABASE_URL)
if [ -n "$database_url" ]; then
	log "dumping the managed database"
	docker run --rm "$PG_IMAGE" pg_dump --format=custom --no-owner "$database_url" > "$tmp"
	docker run --rm -i "$PG_IMAGE" pg_restore --list > /dev/null < "$tmp" || fail "the dump cannot be read back"
else
	user=$(env_value POSTGRES_USER); user=${user:-shop_api}
	db=$(env_value POSTGRES_DB); db=${db:-ecommerce}
	log "dumping database $db from the db container"
	compose exec -T db pg_dump --format=custom --no-owner -U "$user" -d "$db" > "$tmp"
	compose exec -T db pg_restore --list > /dev/null < "$tmp" || fail "the dump cannot be read back"
fi
[ -s "$tmp" ] || fail "the dump is empty"
mv "$tmp" "$db_file"
log "database: $db_file ($(du -h "$db_file" | cut -f1))"

# ---- 3: uploaded images (only with STORAGE_DRIVER=local; S3 buckets have their own versioning) ----
volume="${COMPOSE_PROJECT}_uploads"
if docker volume inspect "$volume" > /dev/null 2>&1; then
	docker run --rm -v "$volume:/uploads:ro" alpine tar czf - -C /uploads . > "$uploads_tmp"
	mv "$uploads_tmp" "$uploads_file"
	log "uploads: $uploads_file ($(du -h "$uploads_file" | cut -f1))"
else
	log "no $volume volume (S3 storage?) - skipping uploads"
fi

# ---- 4: rotation ----
deleted=$(find "$BACKUP_DIR" -maxdepth 1 -name 'stockwise-*' -type f -mtime +"$KEEP_DAYS" -print -delete | wc -l)
log "removed $deleted backup file(s) older than $KEEP_DAYS days"

# ---- 5: off the server ----
if [ -n "$BACKUP_REMOTE" ]; then
	command -v rclone > /dev/null || fail "BACKUP_REMOTE is set but rclone is not installed"
	rclone copy "$db_file" "$BACKUP_REMOTE/"
	[ -f "$uploads_file" ] && rclone copy "$uploads_file" "$BACKUP_REMOTE/"
	log "copied to $BACKUP_REMOTE"
else
	log "WARNING: BACKUP_REMOTE is not set - backups exist only on this server"
fi
log "done"
