#!/bin/sh
# Gate T5 — A1, the acceptance kit, tools/copilot-acceptance/. Nothing here
# runs the kit, copilot, claude, gh or the network: the kit is checked for
# syntax, for a script that runs the fixture brief with agent.verify = copilot
# and model.verify set while work and review stay on Claude, and for a README
# section, under a heading naming verify, that says how to run it, what its
# summary reports and where the operator records the outcome.
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
# A script that sets agent.verify to copilot and model.verify, on the fixture brief.
found=
for f in "$K"/*.sh; do
  [ -f "$f" ] || continue
  grep -qE 'agent\.verify[[:space:]"'\''=]+copilot|VLOOP_AGENT_VERIFY=["'\'']?copilot' "$f" || continue
  grep -qE 'model\.verify|VLOOP_MODEL_VERIFY' "$f" || continue
  grep -qF 'greet.loop-brief.md' "$f" || continue
  found=$f
done
[ -n "$found" ] || fail "no script in $K sets agent.verify = copilot and model.verify and runs the fixture brief (greet.loop-brief.md)"
grep -qE 'agent\.(work|review|second-review)[[:space:]"'\''=]+copilot|VLOOP_AGENT_(WORK|REVIEW|SECOND_REVIEW)=["'\'']?copilot' "$found" &&
  fail "$found puts work or review sessions on copilot: plan, work and review stay on Claude"
grep -qE 'run\.verify[[:space:]"'\''=]+false|VLOOP_RUN_VERIFY=["'\'']?false' "$found" && fail "$found turns run.verify off"
# Its README section: from the first heading naming verify to the next ## heading.
R=$K/README.md
awk 'p && /^## / { exit } p || (/^#/ && tolower($0) ~ /verify/) { p = 1; print }' "$R" >"$t/sec"
[ -s "$t/sec" ] || fail "$R has no section under a heading naming verify"
grep -qF 'agent.verify' "$t/sec" || fail "$R's verify section does not describe the run with agent.verify = copilot"
grep -qF "$(basename "$found")" "$t/sec" || fail "$R's verify section does not say how to run $(basename "$found")"
grep -qF 'permission_denials' "$t/sec" || fail "$R's verify section does not say the summary reports the verify session's permission_denials"
grep -qF 'session/v4' "$t/sec" || fail "$R's verify section does not say the summary reports the record's session/v4 validity"
grep -qF 'verify/v1' "$t/sec" || fail "$R's verify section does not say the summary reports the kept report's verify/v1 validity"
grep -qF 'copilot-cli-experiments.md' "$t/sec" || fail "$R's verify section does not say where the operator records the outcome"
echo "gate T5: ok"
