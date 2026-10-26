#!/bin/sh
# Gate T2 — J3, the check-gap halt, by the brief's worked example run lines. In
# a fixture run whose stub gate review writes fixed gate-verdict/v2 verdicts
# (this folder's *.json, one per round) and a two-task plan:
#   A  check-gap in round 1      -> the finding printed, the halt line, exit 2,
#                                   one plan session, no work session, nothing
#                                   committed, no plan left; a re-run plans afresh
#   B  uncovered, then PASS      -> sent back once, the plan committed, exit 0
#   C  check-gap and wrong-reason in round 1 -> both printed, the halt, exit 2,
#                                   no revision round
#   D  wrong-reason, then check-gap in round 2 -> the check-gap halt naming
#                                   gate-review-2.json, not "failed the plan twice"
#   E  two FAILs with no check-gap -> blocked as today: "failed the plan twice"
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
H=${H:-$G}
. "$H/lib.sh"
newbin
GAP="gate T2 (check-gap): pnpm smoke in a fresh copy: no check runs pnpm smoke"
WR="gate T1 (wrong-reason): the gate fails on a path typo, not for the missing T1.out"
UN="gate T2 (uncovered): T2.out holds what the operator meant: no gate, check or review can judge it"
TWICE="the gate review failed the plan twice"
has() { trimmed "$1" | grep -qxF -- "$2"; }
count() { sessions "$1" | grep -c "^$2" | tr -d ' '; }
# halted <case> <out> <round> <plan sessions>: the check-gap halt of that round.
halted() {
  [ "$rc" = 2 ] || fail "$1: vloop run exited $rc, want 2: $(tail -8 "$2")"
  l=$(runline "$2.err" "vloop: the gate review found")
  [ "$l" = "$(gapline "$3")" ] || fail "$1: the halt line is '$l', want '$(gapline "$3")' (stderr: $(tail -5 "$2.err"))"
  rp=$(runpath "$2.err" "vloop: the gate review found")reports/gate-review-$3.json
  [ -f "$t/repo/$rp" ] || fail "$1: the halt names $rp, which does not exist"
  (cd "$t/repo" && venv "$t/bin/new" schema validate gate-verdict/v2 "$rp" >/dev/null 2>&1) || fail "$1: $rp does not validate as gate-verdict/v2"
  "$JQ" -e '[.tasks[].findings[] | select(.kind == "check-gap")] | length == 1' "$t/repo/$rp" >/dev/null || fail "$1: $rp does not hold the check-gap finding"
  grep -qF "$TWICE" "$2" && fail "$1: the run says '$TWICE': $(grep -F "$TWICE" "$2")"
  [ "$(count "$2.log" /vloop:plan)" = "$4" ] || fail "$1: $(count "$2.log" /vloop:plan) plan session(s), want $4"
  [ "$(count "$2.log" /vloop:work)" = 0 ] || fail "$1: a work session ran: $(sessions "$2.log" | tr '\n' ';')"
  untouched "$1"
}

# A — a check gap in round 1 halts before any work and keeps nothing.
fixture; verdicts check-gap.json; runfix "$t/a"
has "$t/a" "$GAP" || fail "A: the finding line '$GAP' is not printed: $(tail -8 "$t/a")"
grep -qF "sending" "$t/a" && fail "A: the check gap was sent back to the plan session: $(grep -F sending "$t/a")"
halted A "$t/a" 1 1
[ "$(count "$t/a.log" /vloop:gate-review)" = 1 ] || fail "A: $(count "$t/a.log" /vloop:gate-review) gate-review sessions, want 1"
# ...and the operator's re-run, with the check added, plans afresh and runs.
rm -f "$t/stub/verdict."*.json "$t/stub/gate-review.count"
runfix "$t/a2"
[ "$rc" = 0 ] || fail "A: the re-run after the halt exited $rc, want 0: $(tail -8 "$t/a2")"
[ "$(count "$t/a2.log" /vloop:plan)" = 1 ] || fail "A: the re-run ran $(count "$t/a2.log" /vloop:plan) plan session(s), want 1 (it plans afresh)"
grep -qF "resuming acceptance" "$t/a2" && fail "A: the re-run resumed acceptance instead of planning afresh"

# B — an uncovered criterion goes back to the planner, as any finding does today.
fixture; verdicts uncovered.json; runfix "$t/b"
[ "$rc" = 0 ] || fail "B: vloop run exited $rc, want 0: $(tail -8 "$t/b")"
has "$t/b" "$UN" || fail "B: the finding line '$UN' is not printed"
grep -qF "sending 1 problem(s) back to the plan session" "$t/b" || fail "B: the uncovered finding was not sent back: $(tail -8 "$t/b")"
[ "$(count "$t/b.log" /vloop:plan)" = 2 ] || fail "B: $(count "$t/b.log" /vloop:plan) plan session(s), want 2"
grep -qF "check gap" "$t/b" && fail "B: the run speaks of a check gap: $(grep -F 'check gap' "$t/b")"
"$JQ" -e '.gate_review.rounds == 2 and .gate_review.verdict == "PASS"' "$t/repo/.vloop/state/state.json" >/dev/null ||
  fail "B: the plan's gate review is $("$JQ" -c .gate_review "$t/repo/.vloop/state/state.json" 2>&1), want round 2, PASS"
[ "$(venv git -C "$t/repo" rev-parse HEAD)" != "$FIX" ] || fail "B: the plan was not committed"

# C — a check gap beside another finding: both printed, the halt wins.
fixture; verdicts check-gap-and-wrong-reason.json; runfix "$t/c"
has "$t/c" "$GAP" || fail "C: the finding line '$GAP' is not printed: $(tail -8 "$t/c")"
has "$t/c" "$WR" || fail "C: the finding line '$WR' is not printed: $(tail -8 "$t/c")"
grep -qF "sending" "$t/c" && fail "C: the plan was sent back for a revision round: $(grep -F sending "$t/c")"
halted C "$t/c" 1 1

# D — a check gap in round 2, after round 1 was sent back for wrong-reason.
fixture; verdicts wrong-reason.json check-gap.json; runfix "$t/d"
has "$t/d" "$WR" || fail "D: the round-1 finding '$WR' is not printed"
has "$t/d" "$GAP" || fail "D: the round-2 finding '$GAP' is not printed: $(tail -8 "$t/d")"
[ "$(grep -cF "sending 1 problem(s) back to the plan session" "$t/d" | tr -d ' ')" = 1 ] || fail "D: round 1 was not sent back exactly once"
halted D "$t/d" 2 2

# E — two FAILs without a check gap: blocked, as today.
fixture; verdicts wrong-reason.json wrong-reason.json; runfix "$t/e"
[ "$rc" = 2 ] || fail "E: vloop run exited $rc, want 2: $(tail -8 "$t/e")"
grep -qF "$TWICE" "$t/e.err" || fail "E: no '$TWICE' line: $(tail -5 "$t/e.err")"
grep -qF "check gap" "$t/e" && fail "E: the run speaks of a check gap"
"$JQ" -e '.status == "blocked"' "$t/repo/.vloop/state/state.json" >/dev/null || fail "E: the plan is not kept blocked"
echo "gate T2: ok"
