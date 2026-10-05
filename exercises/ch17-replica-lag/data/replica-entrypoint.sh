#!/bin/bash
# Entrypoint of the replica container. On first start the data directory is empty: clone the
# primary with pg_basebackup. -R writes standby.signal and primary_conninfo (plus
# primary_slot_name, from -S) into postgresql.auto.conf, which is what makes this a standby.
set -euo pipefail
if [ ! -s "$PGDATA/PG_VERSION" ]; then
  until pg_isready -h primary -p 5432 -U replicator; do sleep 1; done
  mkdir -p "$PGDATA"
  chown postgres:postgres "$PGDATA"
  chmod 0700 "$PGDATA"
  gosu postgres env PGPASSWORD="$REPLICATION_PASSWORD" \
    pg_basebackup -h primary -p 5432 -U replicator -D "$PGDATA" \
    --wal-method=stream --slot=replica1 --write-recovery-conf --checkpoint=fast --progress
fi
# The image's own entrypoint sees an initialised directory, skips initdb and starts postgres.
exec docker-entrypoint.sh postgres -c hot_standby=on
