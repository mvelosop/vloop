#!/bin/sh
# Gate T5 — A1, the acceptance kit, tools/copilot-acceptance/. Nothing here
# runs the kit, copilot, claude, gh or the network: the kit is checked for
# syntax, for a script that runs the fixture brief with agent.work = copilot
# and model.work set, and for a README that says how to run it, what it
# reports and where the operator records the outcome.
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
# A script that sets agent.work to copilot and model.work, on the fixture brief.
found=
for f in "$K"/*.sh; do
  [ -f "$f" ] || continue
  grep -qE 'agent\.work[[:space:]"'\'']+copilot|VLOOP_AGENT_WORK=["'\'']?copilot' "$f" || continue
  grep -qE 'model\.work|VLOOP_MODEL_WORK' "$f" || continue
  grep -qF 'greet.loop-brief.md' "$f" || grep -qF 'run.sh' "$f" || continue
  found=$f
done
[ -n "$found" ] || fail "no script in $K sets agent.work = copilot and model.work and runs the fixture brief (greet.loop-brief.md)"
grep -qE 'agent\.review[[:space:]"'\'']+copilot|VLOOP_AGENT_REVIEW=["'\'']?copilot' "$found" &&
  fail "$found puts the reviews on copilot: plan and review stay on Claude"
R=$K/README.md
grep -qF 'agent.work' "$R" || fail "$R does not describe the work run with agent.work = copilot"
grep -qF "$(basename "$found")" "$R" || fail "$R does not say how to run $(basename "$found")"
grep -qF 'permission_denials' "$R" || fail "$R does not say the summary reports each work session's permission_denials"
grep -qF 'session/v4' "$R" || fail "$R does not say the summary reports each work session record's session/v4 validity"
grep -qF 'copilot-cli-experiments.md' "$R" || fail "$R does not say where the operator records the outcome"
echo "gate T5: ok"
