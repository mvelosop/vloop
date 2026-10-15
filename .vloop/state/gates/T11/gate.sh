#!/bin/sh
# Gate T11 — close: the worked example end to end with a fresh build, the
# real-data check on this repository, and the brief's planted failures.
#  - a run whose verify session finds 2: the run line, the close dry-run's
#    block, the metrics lines; a run with run.verify = false: none of them;
#    stale scratch gone at run start, .running held throughout;
#  - real data: vloop metrics of every consumed brief and vloop brief check of
#    every brief print what the plan's base prints, but for the retired
#    must-change list: and external format: lines;
#  - planted: a driver that ends blocked on a finding, a clearing that removes
#    .running, and a verify skill padded past 1,500 words are each caught.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
S=$t/stub
line() { sq "$1" | grep -E "^$2 " | head -n 1; }
# v1ok <bin> <name>: a run in a fresh repository whose verify session finds 2
# ends complete, exit 0, every task done, and says so; status 1 otherwise.
v1ok() {
  resetstub; runrepo "$2" T1; vreport "$S/verify-report.json" 2
  : >"$t/stub/argv.log"; inr "$1" run "$BRIEF" </dev/null >"$t/$2.run" 2>&1; got=$?
  [ "$got" = 0 ] && [ "$("$JQ" -r .status "$R/.vloop/state/state.json")" = complete ] &&
    "$JQ" -e '[.tasks[] | select(.status != "done")] | length == 0' "$R/.vloop/state/state.json" >/dev/null &&
    grep -Eq "^ *verify: 2 finding\\(s\\) — see $(lastdir | sed 's/\./\\./g')reports/verify\\.json\$" "$t/$2.run"
}

# --- the worked example
newrepo cfg
inr "$B" config list </dev/null >"$t/cl" 2>&1 || fail "config list: $(cat "$t/cl")"
grep -Eq '^run\.verify=true( |$)' "$t/cl" && grep -Eq '^model\.verify=sonnet( |$)' "$t/cl" && grep -Eq '^effort\.verify=' "$t/cl" ||
  fail "config list lacks run.verify = true, model.verify = sonnet or effort.verify: $(cat "$t/cl")"
v1ok "$B" w || fail "a run whose verify session finds 2: exit $got: $(tail -n 15 "$t/w.run")"
rd=$(lastdir)
git -C "$R" show --name-only --format= HEAD | grep -Eq "^$(echo "$rd" | sed 's/\./\\./g')sessions/[0-9]+-verify\.json\$" ||
  fail "the run commit holds no sessions/NNN-verify.json"
inr "$B" brief close "$BRIEF" --dry-run --no-findings </dev/null >"$t/dry" 2>"$t/dry.err" || fail "close --dry-run: $(cat "$t/dry.err")"
[ "$(head -n 1 "$t/dry")" = "verify of $RUN/$(basename "$rd"): 2 finding(s)" ] || fail "close --dry-run starts: $(head -n 3 "$t/dry")"
[ "$(sed -n 2p "$t/dry")" = "  1. example — Worked example, line 3; vloop greet --loud" ] || fail "close --dry-run's first finding: $(sed -n 2,4p "$t/dry")"
inr "$B" metrics "$BRIEF" </dev/null >"$t/m" 2>&1 || fail "metrics: $(cat "$t/m")"
line "$t/m" time | grep -Eq ' · gate review [0-9.]+ min · verify [0-9.]+ min · gates ' || fail "time line: $(line "$t/m" time)"
line "$t/m" cost | grep -Eq ' · gate review [0-9.]+ · verify [0-9.]+ · ' || fail "cost line: $(line "$t/m" cost)"

resetstub; runrepo off T1
inr "$B" config set run.verify false </dev/null >"$t/o" 2>&1 || fail "config set run.verify false: $(cat "$t/o")"
commit off
mkdir -p "$R/.vloop/tmp" && : >"$R/.vloop/tmp/old.txt"
printf '[ -e .vloop/tmp/old.txt ] && echo "the plan session saw the stale old.txt" >>"$dir/norunning"\n' >"$S/plan.sh"
vrun off.run
[ "$got" = 0 ] || fail "run.verify = false: exit $got: $(tail -n 10 "$t/off.run")"
grep -q -- '-p /vloop:verify' "$S/argv.log" && fail "a verify session ran with run.verify = false"
grep -Eq '^ *verify:' "$t/off.run" && fail "a verify line with run.verify = false"
[ ! -s "$S/norunning" ] || fail "$(cat "$S/norunning")"
inr "$B" brief close "$BRIEF" --dry-run --no-findings </dev/null >"$t/dry2" 2>&1 || fail "close --dry-run: $(cat "$t/dry2")"
[ "$(head -n 1 "$t/dry2")" = "no verify report — verify by hand" ] || fail "close --dry-run with no report starts: $(head -n 2 "$t/dry2")"

# --- real data, on this repository, against the plan's base
basebin
for b in $(grep -l '^status: consumed' docs/briefs/*.md); do
  "$t/vloop-base" metrics "$b" </dev/null >"$t/rb" 2>&1; eb=$?
  "$B" metrics "$b" </dev/null >"$t/rn" 2>&1; en=$?
  [ "$eb" = "$en" ] && cmp -s "$t/rb" "$t/rn" || fail "vloop metrics $b differs from the base's:
$(diff "$t/rb" "$t/rn")"
done
for b in docs/briefs/*.md; do
  "$t/vloop-base" brief check "$b" </dev/null >"$t/cb0" 2>&1; retire "$t/cb0" >"$t/cb"
  "$B" brief check "$b" </dev/null >"$t/cn" 2>&1
  cmp -s "$t/cb" "$t/cn" || fail "vloop brief check $b differs from the base's beyond the retired lines:
$(diff "$t/cb" "$t/cn")"
done

# --- planted: a driver that ends blocked on a verify finding is caught
cat >"$t/plant-vloop" <<P
#!/bin/sh
"$B" "\$@"; c=\$?
if [ "\$1" = run ] && [ "\$c" = 0 ] && ls .vloop/state/runs/*/*/reports/verify.json >/dev/null 2>&1; then exit 2; fi
exit \$c
P
chmod +x "$t/plant-vloop"
v1ok "$t/plant-vloop" plant1 && fail "a driver that ends the run blocked on a verify finding is not caught"
# --- planted: a clearing that removes .running is caught
resetstub; runrepo plant2 T1
printf 'rm -f .vloop/tmp/.running\n' >"$S/plan.sh"
vrun plant2.run
[ -s "$S/norunning" ] || fail ".running removed during the run is not caught by the sessions' check"
# --- planted: a verify skill padded past 1,500 words fails the budget test
mkdir -p "$t/pad/internal" "$t/pad/plugin" && cp "$top/go.mod" "$top/go.sum" "$t/pad/" && cp -R "$top/internal/budget" "$t/pad/internal/" &&
  cp -R "$top/plugin/skills" "$t/pad/plugin/" || fail "copy for the padded skill"
(cd "$t/pad" && go test -count=1 ./internal/budget) >"$t/pad0" 2>&1 || fail "the budget test fails on the unpadded copy: $(tail -n 10 "$t/pad0")"
i=0; while [ $i -lt 1501 ]; do printf 'pad ' ; i=$((i + 1)); done >>"$t/pad/plugin/skills/verify/SKILL.md"
(cd "$t/pad" && go test -count=1 ./internal/budget) >"$t/pad1" 2>&1 && fail "plugin/skills/verify/SKILL.md padded past 1,500 words does not fail the budget test"
grep -q 'verify' "$t/pad1" || fail "the budget test's failure does not name the verify skill: $(tail -n 10 "$t/pad1")"
echo "gate T11: ok"
