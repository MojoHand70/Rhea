#!/usr/bin/env sh
# Start a shell: ./shell.sh 8070 (the default log, database "rhea") or
# ./shell.sh 8071 (the Poland demo log, database "rhea_pl"). Builds first,
# reads the API key from .env, runs in the foreground; Ctrl-C stops it.
set -e
cd "$(dirname "$0")"
port="${1:-8070}"
case "$port" in
  8071) export RHEA_PG_DSN="postgres://rhea:rhea@127.0.0.1:5432/rhea_pl" RHEA_DUCKDB="data/rhea_pl.duckdb" ;;
esac
go build -o rhea ./cmd/rhea
set -a; [ -f .env ] && . ./.env; set +a
exec ./rhea serve -addr ":$port"
