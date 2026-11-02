#!/bin/sh
# Gate T6 — close: every gate of this plan at once, in order. Together they
# hold the worked example line for line: the fence check tables, the
# apply_patch lines, the unchanged lines, the changed unknown phase and the
# hook's plan phase (T1); the stub run with its line, record, argv and hooks
# file, and the exit-3, no-plan and moved-HEAD planted failures (T2); the
# warning and the model.plan and K1 planted failures (T3); the docs (T4) and
# the kit (T5).
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for n in 1 2 3 4 5; do
  sh ".vloop/state/gates/T$n/gate.sh" || fail "gate T$n fails"
done
echo "gate T6: ok"
