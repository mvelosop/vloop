#!/bin/sh
# Gate T1 — S1: the driver empties .vloop/tmp/ of everything but .running, the
# run's lock: at run start before any session, and before every work session,
# the final pass (and, from T4, the verify session). It does not clear between
# plan rounds or between a work session and its review. A symbolic link is
# removed, never followed; an entry that cannot be removed is a warning and the
# run goes on. vloop plugin path clears nothing.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"

runrepo s1 T1 T2
cat >>"$R/.vloop/config.toml" <<'CFG'

[[check]]
name = "scratch"
paths = ["**"]
run = "test ! -e .vloop/tmp/scratch/x.txt"

[[check]]
name = "late"
paths = ["nomatch/**"]
run = "test ! -e .vloop/tmp/late.txt"
CFG
commit checks
# Stale scratch from before the run: files, a folder, links out of the tree, and
# an entry the driver cannot remove.
mkdir -p "$t/outside" "$R/.vloop/tmp/olddir" "$R/.vloop/tmp/locked" || fail mkdir
: >"$t/outside/keep.txt"; : >"$t/outside/keep2.txt"
: >"$R/.vloop/tmp/old.txt"; : >"$R/.vloop/tmp/olddir/x"; : >"$R/.vloop/tmp/locked/f"
ln -s "$t/outside" "$R/.vloop/tmp/linkdir" && ln -s "$t/outside/keep2.txt" "$R/.vloop/tmp/linkfile" || fail "ln -s"
chmod 555 "$R/.vloop/tmp/locked" || fail chmod
[ -z "$(git -C "$R" status --porcelain)" ] || fail "the stale scratch is not ignored by git: $(git -C "$R" status --porcelain)"

S=$t/stub
cat >"$S/plan.sh" <<'P'
n=$(cat "$dir/plan-n" 2>/dev/null || echo 0); n=$((n + 1)); echo "$n" >"$dir/plan-n"
for e in old.txt olddir linkdir linkfile; do
  if [ -e ".vloop/tmp/$e" ] || [ -L ".vloop/tmp/$e" ]; then echo "plan round $n saw the stale .vloop/tmp/$e" >>"$dir/s1"; fi
done
if [ "$n" = 1 ]; then
  : >.vloop/tmp/plan-scratch.txt; mkdir -p .vloop/tmp/scratch; : >.vloop/tmp/scratch/x.txt
else
  [ -f .vloop/tmp/plan-problems.md ] || echo "plan round $n: no .vloop/tmp/plan-problems.md" >>"$dir/s1"
  [ -f .vloop/tmp/plan-scratch.txt ] || echo "plan round $n: round 1's .vloop/tmp/plan-scratch.txt was cleared" >>"$dir/s1"
  mkdir -p .vloop/tmp/scratch; : >.vloop/tmp/scratch/x.txt
fi
P
cat >"$S/work-T1.sh" <<'W'
for e in scratch/x.txt plan-scratch.txt gate-verdict.json; do
  [ -e ".vloop/tmp/$e" ] && echo "T1's work session saw .vloop/tmp/$e" >>"$dir/s1"
done
: >.vloop/tmp/w1.txt
W
cat >"$S/review.sh" <<'V'
if [ "$TASK" = T1 ]; then
  for e in proposal.json checks.json w1.txt; do
    [ -e ".vloop/tmp/$e" ] || echo "T1's review session found no .vloop/tmp/$e" >>"$dir/s1"
  done
fi
V
cat >"$S/work-T2.sh" <<'W'
[ -e .vloop/tmp/w1.txt ] && echo "T2's work session saw T1's .vloop/tmp/w1.txt" >>"$dir/s1"
: >.vloop/tmp/late.txt
W
: >"$S/gr-fail"; rm -f "$S/plan-n" "$S/s1" "$S/norunning"

vrun s1.run
[ ! -s "$S/s1" ] || fail "the scratch was not cleared as S1 requires:
$(cat "$S/s1")"
[ "$(cat "$S/plan-n" 2>/dev/null)" = 2 ] || fail "the gate review's FAIL did not lead to a second plan round (plan rounds: $(cat "$S/plan-n" 2>/dev/null)): $(tail -n 15 "$t/s1.run")"
[ ! -s "$S/norunning" ] || fail ".vloop/tmp/.running was missing in a session: $(cat "$S/norunning")"
[ "$got" = 0 ] || fail "vloop run: exit $got, want 0: $(tail -n 15 "$t/s1.run")"
grep -Eq '^warning: cannot clear \.vloop/tmp/locked[:/]' "$t/s1.run" || fail "no 'warning: cannot clear .vloop/tmp/locked: <error>' for an entry that cannot be removed:
$(cat "$t/s1.run")"
[ -f "$t/outside/keep.txt" ] && [ -f "$t/outside/keep2.txt" ] || fail "clearing followed a symbolic link out of .vloop/tmp/ and removed what it points at"
[ -f "$R/.vloop/tmp/locked/f" ] || fail "the unremovable entry vanished"
for e in old.txt olddir linkdir linkfile; do
  { [ -e "$R/.vloop/tmp/$e" ] || [ -L "$R/.vloop/tmp/$e" ]; } && fail "the stale .vloop/tmp/$e outlived the run"
done
[ "$("$JQ" -r .status "$R/.vloop/state/state.json")" = complete ] || fail "the plan did not end complete"

# vloop plugin path is not a clearing point.
chmod 755 "$R/.vloop/tmp/locked"
: >"$R/.vloop/tmp/keepme"
inr "$B" plugin path </dev/null >"$t/pp" 2>&1 || fail "vloop plugin path: $(cat "$t/pp")"
[ -f "$R/.vloop/tmp/keepme" ] || fail "vloop plugin path cleared .vloop/tmp/"
p=$(tail -n 1 "$t/pp")
case "$p" in /*) f=$p ;; *) f=$R/$p ;; esac
[ -f "$f/skills/plan/SKILL.md" ] || fail "vloop plugin path printed '$p', not an extracted plugin"
echo "gate T1: ok"
