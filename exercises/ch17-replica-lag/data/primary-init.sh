#!/bin/bash
# Runs once, when the primary's data directory is first created.
# 1. A role that may only stream WAL. 2. A physical slot so the primary keeps WAL until the
# replica has it. 3. A pg_hba.conf line: the "replication" pseudo-database is not matched by "all".
set -euo pipefail
psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -v pw="$REPLICATION_PASSWORD" <<'SQL'
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD :'pw';
SELECT pg_create_physical_replication_slot('replica1');
SQL
echo "host replication replicator all scram-sha-256" >> "$PGDATA/pg_hba.conf"
