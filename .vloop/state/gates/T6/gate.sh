#!/bin/sh
# Gate T6 — close: every gate of this plan at once, in order. Together they
# hold the worked example line for line: the fence check tables, the
# apply_patch lines, the unchanged lines and the changed unknown phase (T1);
# the hook's verify phase, the stub run with its record, argv, hooks file and
# kept report, and the README.md and moved-HEAD planted failures (T2); the
# warning and the other planted failures (T3); the docs (T4) and the kit (T5).
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for n in 1 2 3 4 5; do
  sh ".vloop/state/gates/T$n/gate.sh" || fail "gate T$n fails"
done
echo "gate T6: ok"
