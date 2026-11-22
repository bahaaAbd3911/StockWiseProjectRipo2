#!/usr/bin/env bash
# Restore a database backup made by deploy/backup.sh.
#
#   deploy/restore.sh --test  /var/backups/stockwise/stockwise-db-....dump
#       Restores into a separate database (ecommerce_restore_test), prints how
#       many rows each important table has, and drops it again. Your live data
#       is not touched. Do this once after setting up backups, and now and then.
#
#   deploy/restore.sh --live  /var/backups/stockwise/stockwise-db-....dump [uploads.tar.gz]
#       REPLACES the live database (and the uploaded images, if given) with
#       the backup. The API is stopped during the restore. Asks for confirmation.
#
# Uses the db service of docker-compose.yml (ENV_FILE, default .env.production).
set -euo pipefail

cd "$(dirname "$0")/.."
ENV_FILE=${ENV_FILE:-.env.production}
COMPOSE_PROJECT=${COMPOSE_PROJECT:-stockwise}

usage() { echo "usage: $0 --test|--live <db.dump> [uploads.tar.gz]"; exit 2; }
[ $# -ge 2 ] || usage
mode=$1 dump=$2 uploads=${3:-}
[ -f "$dump" ] || { echo "no such file: $dump"; exit 1; }
[ -f "$ENV_FILE" ] || { echo "$ENV_FILE not found"; exit 1; }

env_value() { grep -E "^$1=" "$ENV_FILE" | tail -n 1 | cut -d= -f2- || true; }
compose() { docker compose --env-file "$ENV_FILE" "$@"; }
psql_db() { compose exec -T db psql -v ON_ERROR_STOP=1 -U "$user" "$@"; }

user=$(env_value POSTGRES_USER); user=${user:-shop_api}
db=$(env_value POSTGRES_DB); db=${db:-ecommerce}

case "$mode" in
--test)
	test_db=${db}_restore_test
	echo "restoring $dump into $test_db ..."
	psql_db -d postgres -c "DROP DATABASE IF EXISTS $test_db" -c "CREATE DATABASE $test_db" > /dev/null
	compose exec -T db pg_restore --no-owner -U "$user" -d "$test_db" < "$dump"
	echo "rows per table in the backup:"
	psql_db -d "$test_db" -c "
		SELECT 'users' AS table_name, COUNT(*) FROM users UNION ALL
		SELECT 'products', COUNT(*) FROM products UNION ALL
		SELECT 'orders', COUNT(*) FROM orders UNION ALL
		SELECT 'order_items', COUNT(*) FROM order_items UNION ALL
		SELECT 'customer_debt_transactions', COUNT(*) FROM customer_debt_transactions UNION ALL
		SELECT 'schema_migrations', COUNT(*) FROM schema_migrations"
	psql_db -d postgres -c "DROP DATABASE $test_db" > /dev/null
	echo "restore test OK (test database removed)"
	;;
--live)
	echo "This REPLACES the live database '$db' with $dump."
	[ -n "$uploads" ] && echo "It also REPLACES all uploaded images with $uploads."
	read -r -p "Type 'restore' to continue: " answer
	[ "$answer" = "restore" ] || { echo "cancelled"; exit 1; }

	echo "stopping the API ..."
	compose stop api caddy
	echo "restoring the database ..."
	psql_db -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$db' AND pid <> pg_backend_pid()" > /dev/null
	psql_db -d postgres -c "DROP DATABASE $db" -c "CREATE DATABASE $db" > /dev/null
	compose exec -T db pg_restore --no-owner -U "$user" -d "$db" < "$dump"
	if [ -n "$uploads" ]; then
		echo "restoring uploaded images ..."
		docker run --rm -i -v "${COMPOSE_PROJECT}_uploads:/uploads" alpine \
			sh -c 'rm -rf /uploads/* && tar xzf - -C /uploads && chown -R 65532:65532 /uploads' < "$uploads"
	fi
	echo "starting the API (pending migrations run first) ..."
	compose up -d
	echo "restore done"
	;;
*) usage ;;
esac
