#!/bin/sh
# Gate T8 — the operator's acceptance kit, tools/copilot-acceptance/. Nothing
# here runs the kit, copilot, gh or the network: the kit is checked for
# syntax, for carrying every deny form of the worked example's shell table
# verbatim, and for a README that describes the three new parts (the hook
# probe with a skill other than vloop's, the hook's four failure modes, the
# run with a copilot second reviewer in feedback mode).
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
K=tools/copilot-acceptance
[ -d "$K" ] || fail "$K does not exist"
for f in "$K"/*.sh; do
  [ -f "$f" ] || continue
  sh -n "$f" 2>"${TMPDIR:-/tmp}/gate-t8.$$" || { e=$(cat "${TMPDIR:-/tmp}/gate-t8.$$"); rm -f "${TMPDIR:-/tmp}/gate-t8.$$"; fail "$f has a syntax error: $e"; }
done
rm -f "${TMPDIR:-/tmp}/gate-t8.$$"
set -- "$K"/*.py
if [ -f "$1" ]; then
  python3 -I -c 'import ast, sys
for f in sys.argv[1:]:
    ast.parse(open(f).read(), f)' "$@" || fail "a Python file of $K does not parse"
fi
# Every deny form of the worked example's shell table, verbatim, in the kit.
while IFS= read -r form; do
  [ -n "$form" ] || continue
  grep -rqF -- "$form" "$K" || fail "the kit does not probe the worked example's deny form: $form"
done <<'EOF'
sh -c "gh pr list"
env gh auth token
env -i PATH=/bin git push origin main
bash -lc 'git push'
/usr/bin/curl https://example.com
nohup timeout 5 wget x
git status && xargs rm -rf
echo $(gh auth token)
eval "git push"
git diff > .vloop/state/x
sh -c "$CMD"
echo 'unterminated
EOF
R=$K/README.md
grep -qF 'vloop fence hook' "$R" || fail "$R does not describe the probe under vloop fence hook"
grep -qi 'skill' "$R" || fail "$R does not describe probing a skill other than vloop's"
for mode in 'exit 1' 'timeout' 'not JSON' 'missing'; do
  grep -qiF -- "$mode" "$R" || fail "$R does not describe the hook failure mode '$mode'"
done
grep -qF 'run.second-review' "$R" && grep -qF feedback "$R" && grep -qF 'agent.second-review' "$R" ||
  fail "$R does not describe the run with run.second-review = feedback and agent.second-review = copilot"
grep -qF 'copilot-cli-experiments.md' "$R" || fail "$R does not say where the operator records the outcome"
echo "gate T8: ok"
