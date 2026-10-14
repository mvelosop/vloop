#!/bin/sh
# Gate T6 — the eval case gate-review-real-misses (brief L6): its scaffold, run
# in an empty directory on this host, exits 0 and leaves a state/v2 plan of
# seven tasks in shell `sh` whose seven gates all fail on the base; its graders
# hold the verdict to FAIL and name `unpassable`; the evals README and the
# evals guide list it, the README saying it is for macOS hosts.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
C=plugin/evals/gate-review-real-misses
[ -f "$C/scaffold.sh" ] && [ -f "$C/case.yaml" ] && [ -f "$C/prompt.md" ] && [ -d "$C/graders" ] ||
  fail "$C does not have scaffold.sh, case.yaml, prompt.md and graders/"
grep -Fxq 'name: gate-review-real-misses' "$C/case.yaml" || fail "$C/case.yaml does not name the case"
grep -Fxq '  scaffold_script: scaffold.sh' "$C/case.yaml" || fail "$C/case.yaml has no context.scaffold_script: scaffold.sh"
[ "$(awk 'n == 2 && NF { print; exit } /^---$/ { n++ }' "$C/prompt.md")" = '/vloop:gate-review' ] ||
  fail "$C/prompt.md's body is not /vloop:gate-review"

S="$t/scaffold"
mkdir -p "$S" "$t/bin" && cp "$B" "$t/bin/vloop" || fail "mkdir, and vloop on PATH"
(cd "$S" && env PATH="$t/bin:$PATH" bash "$top/$C/scaffold.sh") >"$t/scaffold.log" 2>&1 ||
  fail "scaffold.sh in an empty directory exits non-zero: $(tail -n 20 "$t/scaffold.log")"
P="$S/.vloop/state/state.json"
[ -f "$P" ] || fail "the scaffold wrote no .vloop/state/state.json"
"$B" schema validate state/v2 "$P" >"$t/validate" 2>&1 || fail "the plan is not state/v2: $(cat "$t/validate")"
[ "$("$JQ" -r .shell "$P")" = sh ] || fail "the plan's shell is not \"sh\""
[ "$("$JQ" '.tasks | length' "$P")" = 7 ] || fail "the plan has $("$JQ" '.tasks | length' "$P") tasks, not seven"
i=0
while [ $i -lt 7 ]; do
  id=$("$JQ" -r ".tasks[$i].id" "$P")
  "$JQ" -r ".tasks[$i].verify" "$P" >"$t/verify.$i"
  (cd "$S" && env PATH="$t/bin:$PATH" sh -c "$(cat "$t/verify.$i")") >"$t/gate.$i" 2>&1 &&
    fail "$id's gate passes on the scaffold's base: $(cat "$t/verify.$i")"
  i=$((i + 1))
done

# Graders: the verdict FAIL, and the defective gates as `unpassable`.
grep -l '^type: regex' "$C"/graders/*.md 2>/dev/null | xargs grep -l 'path: .vloop/tmp/gate-verdict.json' 2>/dev/null |
  xargs grep -l 'FAIL' >/dev/null 2>&1 || fail "no regex grader holds .vloop/tmp/gate-verdict.json's verdict to FAIL"
u=$(grep -l 'unpassable' "$C"/graders/*.md 2>/dev/null | wc -l | tr -d ' ')
[ "$u" -ge 6 ] || fail "$u grader(s) name unpassable; the six defective gates need one each"

R=plugin/evals/README.md
grep -q '^| `gate-review-real-misses` |' "$R" || fail "$R has no table row for gate-review-real-misses"
grep 'macOS' "$R" | grep -Eq 'gate-review-real-misses|BSD' || fail "$R does not say, beside the case, that it is for macOS hosts"
grep '^| `/vloop:gate-review` |' docs/guide/evals.md | grep -Fq 'gate-review-real-misses' ||
  fail "docs/guide/evals.md's /vloop:gate-review row does not list gate-review-real-misses"
echo "gate T6: ok"
