#!/bin/sh
# Gate T4 — close: every gate of the run at once, against the finished tree.
# T1 holds the worked example's config lines, T2 the stub comparison (argv,
# working directory and environment as the base's, the planted flag caught)
# and the session/v3 records, T3 metrics/v3 and the base-written run directory.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
for g in T1 T2 T3; do
  sh "$G/../$g/gate.sh" || { echo "GATE FAIL: gate $g fails on the finished tree" >&2; exit 1; }
done
echo "gate T4: ok"
