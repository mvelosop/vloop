#!/bin/sh
# Gate T9 — E1: vloop defect add refuses, before writing anything, a brief
# (from --brief or --blame) that is not merged (M-5: no commit on the default
# branch where it is consumed), exit 1, with the message that says to record an
# intervention instead; with no default branch it refuses with M-5's lookup
# error, exit 1. A merged brief is recorded as before; --found-by stays
# required.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
U=B20260101-0700-unmerged.loop-brief
M=B20260101-0800-merged.loop-brief
# The refusal for brief $1: "<summary>" may be the placeholder or the summary.
refusal() {
  printf 'vloop: %s is not merged: record what you found before merge as an intervention — vloop intervention add "%s" --brief %s --kind verification-finding --origin <brief|plan|work|env> --phase verify --automatable partly --by operator\n' "$1" "<summary>" "$1"
  printf 'vloop: %s is not merged: record what you found before merge as an intervention — vloop intervention add "%s" --brief %s --kind verification-finding --origin <brief|plan|work|env> --phase verify --automatable partly --by operator\n' "$1" "$2" "$1"
}
# add <want exit> <args...>: vloop defect add in $R; stderr in $t/err.
add() {
  w=$1; shift
  inr "$B" defect add "$@" </dev/null >"$t/out" 2>"$t/err"; c=$?
  [ "$c" = "$w" ] || fail "vloop defect add $*: exit $c, want $w: $(cat "$t/err")"
}
nodefects() { [ -z "$(ls -A "$R/.vloop/defects" 2>/dev/null)" ] || fail "$1 wrote a defect file: $(ls "$R/.vloop/defects")"; }

newrepo e
putbrief "${U%.loop-brief}" ready
putbrief "${M%.loop-brief}" consumed
mkdir -p "$R/src" && printf 'from u\n' >"$R/src/u.txt" && printf 'from m\n' >"$R/src/m.txt" || fail fixture
git -C "$R" add docs .vloop && git -C "$R" commit -q -m briefs || fail "commit briefs"
git -C "$R" add src/u.txt && git -C "$R" commit -q -m "u's line" -m "Vloop-Brief: $U" || fail "commit u"
git -C "$R" add src/m.txt && git -C "$R" commit -q -m "m's line" -m "Vloop-Brief: $M" || fail "commit m"

add 1 "the summary" --brief "$U" --found-by user
refusal "$U" "the summary" | grep -Fxq -f - "$t/err" || fail "defect add on an unmerged brief printed
$(cat "$t/err")
want
$(refusal "$U" "the summary" | head -n 1)"
[ ! -s "$t/out" ] || fail "defect add on an unmerged brief printed on stdout: $(cat "$t/out")"
nodefects "defect add --brief <unmerged>"
add 1 "the summary" --blame src/u.txt:1 --found-by operator
refusal "$U" "the summary" | grep -Fxq -f - "$t/err" || fail "defect add --blame on a line of an unmerged brief printed
$(cat "$t/err")"
nodefects "defect add --blame <unmerged line>"
add 2 "the summary" --brief "$M"
grep -Fq -- '--found-by is required' "$t/err" || fail "defect add without --found-by: $(cat "$t/err")"

add 0 "merged one" --brief "$M" --found-by operator
p=$(tail -n 1 "$t/out"); [ -f "$R/$p" ] || fail "defect add on a merged brief printed '$p', not a defect file"
grep -Fxq "brief: $M" "$R/$p" || fail "$p is not on $M"
add 0 "merged two" --blame src/m.txt:1 --found-by user
p=$(tail -n 1 "$t/out"); [ -f "$R/$p" ] && grep -Fxq "brief: $M" "$R/$p" || fail "defect add --blame on a merged brief's line: '$p'"

# No default branch: only "trunk" exists.
newrepo nodef
git -C "$R" checkout -q -b trunk || fail "checkout trunk"
putbrief "${M%.loop-brief}" consumed
commit "on trunk"
git -C "$R" branch -q -D main 2>/dev/null
add 1 "x" --brief "$M" --found-by user
[ "$(cat "$t/err")" = "vloop: no default branch: origin/HEAD, main and master are all missing" ] ||
  fail "defect add with no default branch printed
$(cat "$t/err")"
nodefects "defect add with no default branch"
echo "gate T9: ok"
