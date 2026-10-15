#!/bin/sh
# Gate T8 — V4: vloop brief close and close --dry-run print first, on stdout,
# the report of the newest of the brief's run folders that holds
# reports/verify.json — "verify of <run id>/<folder>: <n> finding(s)" and one
# block per finding in the report's order, "…: no findings" for an empty list,
# "no verify report — verify by hand" when no folder holds one — then close's
# output as before. Close still needs --finding or --no-findings, and a verify
# finding is recorded only when passed as a --finding.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
S=$t/stub
# dry <out>: close --dry-run --no-findings in $R, stdout in $t/<out>.
dry() {
  inr "$B" brief close "$BRIEF" --dry-run --no-findings </dev/null >"$t/$1" 2>"$t/$1.err" || fail "close --dry-run --no-findings: $(cat "$t/$1.err")"
  [ -z "$(git -C "$R" status --porcelain)" ] || fail "close --dry-run wrote: $(git -C "$R" status --porcelain)"
}
# head_is <out> <expected file>: the output starts with exactly those lines.
head_is() {
  n=$(wc -l <"$2" | tr -d ' ')
  head -n "$n" "$t/$1" >"$t/$1.head"
  cmp -s "$2" "$t/$1.head" || fail "$1 does not start with
$(cat "$2")
it starts with
$(head -n $((n + 2)) "$t/$1")"
}
# rest_ok <out> <lines of the head>: what follows is close's dry-run output.
rest_ok() {
  tail -n +$(($2 + 1)) "$t/$1" | grep -v '^[[:space:]]*$' >"$t/$1.rest"
  head -n 1 "$t/$1.rest" | grep -q "^$RUN " || fail "$1: after the verify block, close's output does not follow: $(head -n 3 "$t/$1.rest")"
  [ "$(tail -n 1 "$t/$1.rest")" = "would commit [vloop] close $RUN" ] || fail "$1: the last line is '$(tail -n 1 "$t/$1.rest")'"
}

runrepo v T1
cat >>"$R/.vloop/config.toml" <<CFG

[[check]]
name = "final"
paths = ["nomatch/**"]
run = "test ! -e '$S/failfinal'"
CFG
commit final
printf ': >"$dir/failfinal"\n' >"$S/plan.sh"
vreport "$S/verify-report.json" 2
vrun v1.run
[ "$got" = 2 ] || fail "harness: the failing final pass did not end the first run blocked (exit $got): $(tail -n 10 "$t/v1.run")"
f1=$(basename "$(lastdir)")
rm -f "$S/failfinal" "$S/plan.sh"
sleep 1
vrun v2.run
[ "$got" = 0 ] || fail "the resumed run: exit $got: $(tail -n 10 "$t/v2.run")"
f2=$(basename "$(lastdir)")
[ "$f1" != "$f2" ] || fail "harness: both runs share the folder $f1"
[ -f "$R/.vloop/state/runs/$RUN/$f2/reports/verify.json" ] || fail "the complete run left no reports/verify.json (T4, the driver's verify session, comes first)"
after=$(git -C "$R" rev-parse HEAD)

cat >"$t/two" <<EOF
verify of $RUN/$f2: 2 finding(s)
  1. example — Worked example, line 3; vloop greet --loud
     brief:    exit 2
     observed: exit 1
  2. surface — README.md:12
     brief:    greet prints in upper case with --loud
     observed: README.md says --loud is not supported
EOF
dry d1
head_is d1 "$t/two"
rest_ok d1 7

# An older folder's report is not the newest.
mkdir -p "$R/.vloop/state/runs/$RUN/$f1/reports" || fail mkdir
vreport "$R/.vloop/state/runs/$RUN/$f1/reports/verify.json" 0
commit "an older report"
dry d2
head_is d2 "$t/two"
git -C "$R" rm -q ".vloop/state/runs/$RUN/$f2/reports/verify.json" && commit "drop the newest report"
dry d3
printf 'verify of %s/%s: no findings\n' "$RUN" "$f1" >"$t/none"
head_is d3 "$t/none"
rest_ok d3 1
git -C "$R" rm -q ".vloop/state/runs/$RUN/$f1/reports/verify.json" && commit "drop every report"
dry d4
printf 'no verify report — verify by hand\n' >"$t/noreport"
head_is d4 "$t/noreport"
rest_ok d4 1

# The real close, back at the run's own state.
git -C "$R" reset -q --hard "$after" || fail "reset to the run"
inr "$B" brief close "$BRIEF" </dev/null >"$t/c0" 2>&1; c=$?
[ "$c" = 2 ] || fail "close with neither --finding nor --no-findings: exit $c, want 2"
[ -z "$(git -C "$R" status --porcelain)" ] || fail "a refused close wrote: $(git -C "$R" status --porcelain)"
inr "$B" brief close "$BRIEF" --no-findings </dev/null >"$t/c1" 2>"$t/c1.err" || fail "close --no-findings: $(cat "$t/c1.err")"
head_is c1 "$t/two"
[ "$(git -C "$R" log -1 --format=%s)" = "[vloop] close $RUN" ] || fail "close made no close commit"
[ -z "$(ls -A "$R/.vloop/interventions" 2>/dev/null)" ] || fail "close recorded a verify finding the operator did not pass: $(ls "$R/.vloop/interventions")"
echo "gate T8: ok"
