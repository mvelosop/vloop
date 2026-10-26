#!/bin/sh
# Gate T9 — close: every gate of this plan at once, in order. Together they
# hold the worked example line for line: the fence check tables (T1), the hook
# payloads (T2), the K1 refusal (T3), the second review's modes (T5).
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for n in 1 2 3 4 5 6 7 8; do
  sh ".vloop/state/gates/T$n/gate.sh" || fail "gate T$n fails"
done
echo "gate T9: ok"
