#!/bin/sh
# Gate T6 — the close: gates T1 to T5 together on the final tree. Between them
# they hold every line of the brief's worked example (P1 and its real-data line
# in T1; the goldens in T2; P2 in T3; P3 in T4; P4 in T5), each through the
# binaries built from this tree and the starting point's Python as reference.
for g in T1 T2 T3 T4 T5; do
  sh ".vloop/state/gates/$g/gate.sh" || { echo "GATE FAIL: gate $g" >&2; exit 1; }
done
echo "gate T6: ok"
