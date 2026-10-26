#!/bin/sh
# Gate T6 — close: every gate of the run at once against the finished tree.
# T1 holds the worked example's schema lines and P-8's text, T2 its run lines
# (every planted failure among them), T3 to T5 the skills' citations of P-8,
# the finding kinds, the halt rows and the two eval cases' scaffolds. The evals
# themselves are the operator's, at close: no gate runs claude.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
for g in T1 T2 T3 T4 T5; do
  sh "$G/../$g/gate.sh" || { echo "GATE FAIL: gate $g fails on the finished tree" >&2; exit 1; }
done
echo "gate T6: ok"
