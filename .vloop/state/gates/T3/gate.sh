#!/bin/sh
# Gate T3 — V1's surfaces: run.verify (true|false, default true), model.verify
# (default sonnet) and effort.verify, shown by vloop config list and checked by
# vloop config set; verify/v1, the report's schema, listed and enforced by
# vloop schema; session/v2's phase gains verify, which, like plan and
# gate-review, has no task.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
# x <want exit> <out> <args...>: vloop in $R; stdout in $t/<out>, stderr in $t/<out>.err.
x() {
  w=$1 o=$2; shift 2
  inr "$B" "$@" </dev/null >"$t/$o" 2>"$t/$o.err"; c=$?
  [ "$c" = "$w" ] || fail "vloop $*: exit $c, want $w: $(cat "$t/$o.err") $(cat "$t/$o")"
}

newrepo c3
x 0 l config list
grep -Eq '^run\.verify=true( |$)' "$t/l" || fail "vloop config list has no run.verify=true:
$(cat "$t/l")"
grep -Eq '^model\.verify=sonnet( |$)' "$t/l" || fail "vloop config list has no model.verify=sonnet:
$(cat "$t/l")"
grep -Eq '^effort\.verify=( |$)' "$t/l" || fail "vloop config list has no unset effort.verify:
$(cat "$t/l")"
x 0 g1 config get run.verify
[ "$(cat "$t/g1")" = true ] || fail "config get run.verify printed '$(cat "$t/g1")', want true"
x 0 g2 config get model.verify
[ "$(cat "$t/g2")" = sonnet ] || fail "config get model.verify printed '$(cat "$t/g2")', want sonnet"
x 0 g3 config get effort.verify
[ -z "$(cat "$t/g3")" ] || fail "config get effort.verify printed '$(cat "$t/g3")', want nothing"
x 2 s1 config set run.verify maybe
x 2 s2 config set effort.verify enormous
x 0 s3 config set run.verify false
x 0 g4 config get run.verify
[ "$(cat "$t/g4")" = false ] || fail "after config set run.verify false, get printed '$(cat "$t/g4")'"
x 0 s4 config set run.verify true
x 0 s5 config set model.verify opus
x 0 s6 config set effort.verify high
x 0 g5 config get effort.verify
[ "$(cat "$t/g5")" = high ] || fail "after config set effort.verify high, get printed '$(cat "$t/g5")'"

x 0 sl schema list
grep -Eq '(^|[^a-z])verify/v1([^0-9]|$)' "$t/sl" || fail "vloop schema list does not list verify/v1:
$(cat "$t/sl")"
for f in verify-ok verify-empty; do
  x 0 v schema validate verify/v1 "$G/$f.json"
done
for f in verify-bad-kind verify-no-checked verify-bad-checked verify-no-where verify-bad-schema; do
  inr "$B" schema validate verify/v1 "$G/$f.json" </dev/null >"$t/v" 2>&1 && fail "vloop schema validate verify/v1 accepts $f.json: $(cat "$t/v")"
done
x 0 s schema validate session/v2 "$G/session-verify.json"
x 0 s schema validate session/v2 "$G/session-gate-review.json"
inr "$B" schema validate session/v2 "$G/session-verify-task.json" </dev/null >"$t/v" 2>&1 &&
  fail "session/v2 accepts a verify record with a task: $(cat "$t/v")"
echo "gate T3: ok"
