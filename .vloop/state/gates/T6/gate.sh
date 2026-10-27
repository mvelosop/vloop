#!/bin/sh
# Gate T6 — close: every gate of this plan at once, in order. Together they
# hold the worked example line for line: the fence check tables, the review
# phase's rm line and the unknown phase (T1); the hook's work phase, the stub
# run with its record, argv and hooks file, agent.plan refused and the moved
# HEAD (T2); the warning and the planted failures (T3); the docs (T4) and the
# kit (T5).
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for n in 1 2 3 4 5; do
  sh ".vloop/state/gates/T$n/gate.sh" || fail "gate T$n fails"
done
echo "gate T6: ok"
