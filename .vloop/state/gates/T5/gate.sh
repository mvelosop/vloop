#!/bin/sh
# Gate T5 — A1, the acceptance kit, tools/copilot-acceptance/. Nothing here
# runs the kit, copilot, claude, gh or the network: the kit is checked for
# syntax, for a script that runs the fixture brief with agent.plan = copilot
# and model.plan set while gate review, work, review and verify stay on
# Claude, and for a README section, under a heading naming plan sessions,
# that says how to run it, what its summary reports and where the operator
# records the outcome.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
K=tools/copilot-acceptance
[ -d "$K" ] || fail "$K does not exist"
[ -f "$K/greet.loop-brief.md" ] || fail "$K/greet.loop-brief.md, the fixture brief, is gone"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
for f in "$K"/*.sh; do
  [ -f "$f" ] || continue
  sh -n "$f" 2>"$t/err" || fail "$f has a syntax error: $(cat "$t/err")"
done
set -- "$K"/*.py
if [ -f "$1" ]; then
  python3 -I -c 'import ast, sys
for f in sys.argv[1:]:
    ast.parse(open(f).read(), f)' "$@" || fail "a Python file of $K does not parse"
fi
# A script that sets agent.plan to copilot and model.plan, on the fixture brief.
found=
for f in "$K"/*.sh; do
  [ -f "$f" ] || continue
  grep -qE 'agent\.plan[[:space:]"'\''=]+copilot|VLOOP_AGENT_PLAN=["'\'']?copilot' "$f" || continue
  grep -qE 'model\.plan|VLOOP_MODEL_PLAN' "$f" || continue
  grep -qF 'greet.loop-brief.md' "$f" || continue
  found=$f
done
[ -n "$found" ] || fail "no script in $K sets agent.plan = copilot and model.plan and runs the fixture brief (greet.loop-brief.md)"
grep -qE 'agent\.(gate-review|work|review|second-review|verify)[[:space:]"'\''=]+copilot|VLOOP_AGENT_(GATE_REVIEW|WORK|REVIEW|SECOND_REVIEW|VERIFY)=["'\'']?copilot' "$found" &&
  fail "$found puts another phase on copilot: gate review, work, review and verify stay on Claude"
# Its README section: from the first heading naming plan to the next ## heading.
R=$K/README.md
awk 'p && /^## / { exit } p || (/^#/ && tolower($0) ~ /(^|[^a-z-])plan([^a-z-]|$)/) { p = 1; print }' "$R" >"$t/sec"
[ -s "$t/sec" ] || fail "$R has no section under a heading naming plan sessions"
grep -qF 'agent.plan' "$t/sec" || fail "$R's plan section does not describe the run with agent.plan = copilot"
grep -qF "$(basename "$found")" "$t/sec" || fail "$R's plan section does not say how to run $(basename "$found")"
grep -qF 'permission_denials' "$t/sec" || fail "$R's plan section does not say the summary reports each session's permission_denials"
grep -qF 'session/v4' "$t/sec" || fail "$R's plan section does not say the summary reports the records' session/v4 validity"
grep -qF 'state/v2' "$t/sec" || fail "$R's plan section does not say the summary reports the plan's state/v2 validity"
grep -qiw 'round' "$t/sec" || fail "$R's plan section does not say the summary gives the gate review's verdict per round"
grep -qF 'copilot-cli-experiments.md' "$t/sec" || fail "$R's plan section does not say where the operator records the outcome"
echo "gate T5: ok"
