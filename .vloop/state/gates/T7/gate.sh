#!/bin/sh
# Gate T7 — the three eval repairs (brief L7): plan-checks-its-gates' fixture
# grader reads files under .vloop/state/gates/, not the trace;
# operate-proposes-options' scaffold puts the plan and the iterations on a work
# branch named for the run id, main untouched, with at least three tasks, the
# first done and the second blocked on a gate dispute.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
PC=plugin/evals/plan-checks-its-gates/graders
for f in "$PC"/*.md; do
  fm=$(awk 'NR == 1 && $0 == "---" { p = 1; next } p && $0 == "---" { exit } p { print }' "$f")
  if printf '%s\n' "$fm" | grep -Eq '^(focus|target): trace$' && grep -q 'greet\.toml' "$f"; then
    fail "$f judges the gate-folder fixtures from the trace"
  fi
done
grep -h '^ *path: ' "$PC"/*.md | sed 's/^ *path: *//; s/^["'\'']//; s/["'\'']$//' | grep -q '^\.vloop/state/gates/.*[^/]$' ||
  fail "no grader of plan-checks-its-gates reads a file under .vloop/state/gates/"

S="$t/operate"
mkdir -p "$S" || fail mkdir
(cd "$S" && bash "$top/plugin/evals/operate-proposes-options/scaffold.sh") >"$t/scaffold.log" 2>&1 ||
  fail "operate-proposes-options/scaffold.sh exits non-zero: $(tail -n 20 "$t/scaffold.log")"
P="$S/.vloop/state/state.json"
[ -f "$P" ] || fail "the operate scaffold left no .vloop/state/state.json in the working tree"
"$B" schema validate state/v2 "$P" >"$t/validate" 2>&1 || fail "the operate plan is not state/v2: $(cat "$t/validate")"
id=$("$JQ" -r .run_id "$P")
br=$(git -C "$S" symbolic-ref --quiet --short HEAD) || fail "the operate scaffold leaves HEAD detached"
[ -n "$id" ] && [ "$br" = "$id" ] || fail "the work branch is '$br', not the run id '$id'"
[ "$("$JQ" -r .branch "$P")" = "$id" ] || fail "the plan's branch is not the run id"
git -C "$S" rev-parse --verify -q main >/dev/null || fail "the operate scaffold has no main"
git -C "$S" cat-file -e main:.vloop/state/state.json 2>/dev/null && fail "main holds a plan: the plan belongs on the work branch, main untouched"
[ "$(git -C "$S" rev-list --count main..HEAD)" -ge 2 ] || fail "the work branch does not carry the plan and an iteration as commits of its own"
[ "$("$JQ" '.tasks | length' "$P")" -ge 3 ] || fail "the operate plan has fewer than three tasks"
[ "$("$JQ" -r '.tasks[0].status' "$P")" = done ] || fail "the first task is not done"
[ "$("$JQ" -r '.tasks[1].status' "$P")" = blocked ] || fail "the second task is not blocked"
"$JQ" -r '.tasks[1].notes' "$P" | grep -iq 'dispute' || fail "the second task's notes do not record a gate dispute"
echo "gate T7: ok"
