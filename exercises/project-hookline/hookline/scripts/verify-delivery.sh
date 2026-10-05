#!/usr/bin/env bash
# The proof of spec N4 (milestone M10, chapter 39): every accepted event reached every endpoint that was
# subscribed to it when it was accepted, at least once. The expected set comes from PostgreSQL (events
# joined with subscriptions and endpoints, not from the deliveries the dispatcher wrote, so a lost
# fan-out shows up as missing); the received set comes from each sink's /receipts, which dedupes by
# (hook, event id). Missing pairs fail the check; duplicates are counted, not failed: they are the price
# of "at least once".
#
#   scripts/verify-delivery.sh                      # every event in the database
#   SINCE='2026-10-05 12:00:00+00' scripts/verify-delivery.sh
#   WAIT=300 SINCE=... scripts/verify-delivery.sh   # re-check every 5 s until nothing is missing
#
# Only endpoints whose URL starts with one of SINKS are checked (default: the two compose sinks), since
# those are the receivers whose receipts we can read. Prints per endpoint: expected, received, missing,
# duplicates, and the delivery lag (ingest to the sink's first receipt) p50, p99 and max. Exit 1 on any
# missing pair. Needs: docker (the compose stack), curl, python3.
set -euo pipefail
cd "$(dirname "$0")/.."

SINCE=${SINCE:-1970-01-01 00:00:00+00}
WAIT=${WAIT:-0}
SINKS=${SINKS:-http://localhost:8091 http://localhost:8092}
OUT=${OUT:-/tmp/hookline-verify}
COMPOSE=${COMPOSE:-docker compose -p hookline -f deploy/compose.yaml}
mkdir -p "$OUT"

expected() {
  # One row per (event, endpoint) that must be delivered: id, endpoint id, url, created_at (epoch), status.
  $COMPOSE exec -T postgres psql -U hookline -d hookline -At -F ',' -c "
    SELECT DISTINCT e.id, ep.id, ep.url, extract(epoch FROM e.created_at), coalesce(d.status, 'none')
    FROM events e
    JOIN subscriptions s ON s.tenant_id = e.tenant_id AND s.event_type IN (e.type, '*')
                        AND s.created_at <= e.created_at
    JOIN endpoints ep ON ep.id = s.endpoint_id AND ep.enabled
    LEFT JOIN deliveries d ON d.event_id = e.id AND d.endpoint_id = ep.id
    WHERE e.created_at >= '$SINCE'" > "$OUT/expected.csv"
}

receipts() {
  : > "$OUT/receipts.jsonl"
  for sink in $SINKS; do
    curl -sf "$sink/receipts" | SINK="$sink" python3 -c '
import json, os, sys
for r in json.load(sys.stdin):
    r["sink"] = os.environ["SINK"]
    print(json.dumps(r))' >> "$OUT/receipts.jsonl"
  done
}

compare() {
  SINKS="$SINKS" python3 - "$OUT/expected.csv" "$OUT/receipts.jsonl" << 'EOF'
import csv, json, os, sys
from collections import defaultdict

sinks = os.environ["SINKS"].split()
got = {}
for line in open(sys.argv[2]):
    r = json.loads(line)
    got[(r["sink"], r["hook"], r["event_id"])] = r

def target(url):
    for sink in sinks:
        prefix = sink.rstrip("/") + "/hooks/"
        if url.startswith(prefix):
            return sink, url[len(prefix):].split("/")[0].split("?")[0]
    return None

per = defaultdict(lambda: {"expected": 0, "received": 0, "missing": [], "duplicates": 0, "lags": []})
skipped = 0
for event_id, _endpoint_id, url, created, status in csv.reader(open(sys.argv[1])):
    t = target(url)
    if t is None:
        skipped += 1
        continue
    row = per[url]
    row["expected"] += 1
    r = got.get((t[0], t[1], event_id))
    if r is None:
        row["missing"].append((event_id, status))
        continue
    row["received"] += 1
    row["duplicates"] += r["count"] - 1
    row["lags"].append(r["first_at"] - float(created))

def pct(xs, p):
    xs = sorted(xs)
    return xs[min(len(xs) - 1, int(p * len(xs)))] if xs else float("nan")

missing = sum(len(r["missing"]) for r in per.values())
print(f"{'endpoint':<44} {'expected':>8} {'received':>8} {'missing':>7} {'dups':>5} {'lag p50':>8} {'p99':>7} {'max':>7}")
for url, r in sorted(per.items()):
    lags = r["lags"]
    print(f"{url:<44} {r['expected']:>8} {r['received']:>8} {len(r['missing']):>7} {r['duplicates']:>5}"
          f" {pct(lags, .5):>7.2f}s {pct(lags, .99):>6.2f}s {max(lags, default=float('nan')):>6.2f}s")
for url, r in sorted(per.items()):
    for event_id, status in r["missing"][:10]:
        print(f"MISSING  {url} event {event_id} (delivery status: {status})")
if skipped:
    print(f"({skipped} pairs for endpoints outside SINKS were not checked)")
total = sum(r["expected"] for r in per.values())
dups = sum(r["duplicates"] for r in per.values())
print(f"TOTAL expected={total} missing={missing} duplicates={dups}")
sys.exit(1 if missing or not total else 0)
EOF
}

deadline=$((SECONDS + WAIT))
while true; do
  expected
  receipts
  if compare > "$OUT/report.txt"; then
    cat "$OUT/report.txt"
    echo "PASS: 0 missing; every expected (event, endpoint) pair reached its sink at least once"
    exit 0
  fi
  if [ "$SECONDS" -ge "$deadline" ]; then
    cat "$OUT/report.txt"
    echo "FAIL: pairs are missing (or nothing was expected); see $OUT/" >&2
    exit 1
  fi
  echo "waiting: $(tail -1 "$OUT/report.txt")"
  sleep 5
done
