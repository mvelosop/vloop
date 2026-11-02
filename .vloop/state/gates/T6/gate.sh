#!/bin/sh
# Gate T6 — close: every gate of this plan at once, in order. Together they
# hold the worked example line for line: the schemas (T1); the fence-check
# table, the work line, the changed unknown-phase line and the unchanged
# gate-review line (T2); the stub Claude FIX and PASS runs with their output,
# record, argv and fence, and every planted failure with the stub Claude
# (T3); the stub Copilot run with its warning, record, deny list and hooks
# file, and the model.review-brief refusal (T4); the skills and docs (T5).
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for n in 1 2 3 4 5; do
  sh ".vloop/state/gates/T$n/gate.sh" || fail "gate T$n fails"
done
echo "gate T6: ok"
