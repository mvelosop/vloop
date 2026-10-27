#!/bin/sh
# Gate T5 — S1, the second review session. Fixture runs with the stub claude,
# whose primary review passes T1 and fails T2 once, and whose second review
# (told apart by model.second-review) fails T1 always and T2 once (see
# stub-claude.sh). Each mode of run.second-review, against the contract: when
# the second review runs, what reaches the task's notes, the run log's lines,
# the end-of-run dissent list, the kept second verdict, iteration/v3 and
# session/v4 records; that the second session runs /vloop:review under the
# review fence with its own model whatever the task says (P-6), never sees the
# primary's verdict, and that a changed file, no verdict or a moved ref are
# handled as for the primary. A copilot second reviewer runs under the hook's
# second-review phase. With the mode off nothing of it appears.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
cd "$top" || fail "repo root"
newbin
# seq <out>: the review sessions in order, "review T1|second T1|…".
seq_() { ev "$1" '^session (review|second|copilot-second) ' | awk '{ printf "%s%s %s", (n++ ? "|" : ""), $2, $3 }'; }
# notes2 <out>: the notes T2's second work session found.
notes2() { ev "$1" '^notes T2 ' | sed -n '2s/^notes T2 //p'; }
# task <field> <id>: the field of a task in the fixture's plan.
task() { "$JQ" -r --arg id "$2" ".tasks[] | select(.id == \$id) | .$1" "$t/repo/.vloop/state/state.json"; }
# itr <out> <n>: iteration n's record.
itr() { sed -n "${2}p" "$1.iters"; }
endline() { sed 's/^[[:space:]]*//' "$1" | grep -xF "$2" >/dev/null; }
valid() { (cd "$t/repo" && venv "$S" schema validate "$1" "$2" >"$t/val" 2>&1) || fail "$3 is not valid $1: $(cat "$t/val")"; }
DL1='second review: 1 dissent(s): T1 (primary PASS, second FAIL)'

# feedback, with a task-level review model on T1 (P-6).
CFG="run.second-review feedback
model.second-review $SECOND"; fixture plan-p6.json
runfix "$t/fb"
[ "$rc" = 0 ] || fail "feedback: vloop run exited $rc: $(tail -8 "$t/fb")"
[ "$(seq_ "$t/fb")" = 'review T1|second T1|review T2|second T2|review T2|second T2' ] ||
  fail "feedback: the reviews ran as '$(seq_ "$t/fb")', want a second review after every primary review"
[ "$(ev "$t/fb" '^session review T1 ' | sed 's/.*model=//')" = task-review-model ] || fail "feedback: the primary review of T1 did not use T1's own review model"
[ "$(ev "$t/fb" '^session second ' | grep -cv "model=$SECOND\$")" = 0 ] || fail "feedback: a second review did not run with model.second-review: $(ev "$t/fb" '^session second ')"
[ "$(ev "$t/fb" '^fence second ' | grep -cv ' ok$')" = 0 ] || fail "feedback: a second review did not run under the review fence: $(ev "$t/fb" '^fence second ')"
[ "$(ev "$t/fb" '^primary-visible ' | grep -cv ' no$')" = 0 ] || fail "feedback: a second review could see the primary's verdict under .vloop/tmp/: $(ev "$t/fb" '^primary-visible ')"
nt=$(notes2 "$t/fb")
case "$nt" in *"primary finding on T2"*"second review: second finding on T2"*) ;;
  *) fail "feedback: T2's retry found notes '$nt', want the primary's findings, then 'second review: second finding on T2'";; esac
[ "$(task status T1)" = done ] && [ "$(task status T2)" = done ] || fail "feedback: T1 is $(task status T1), T2 is $(task status T2), want both done"
[ "$(task attempts T1)" = 0 ] && [ "$(task attempts T2)" = 1 ] || fail "feedback: attempts T1=$(task attempts T1) T2=$(task attempts T2), want 0 and 1"
iterblock "$t/fb.runlog" 1 | awk '$0 == "   review: PASS" { r = NR } $0 == "   second review: FAIL (dissent)" && r { ok = 1 } END { exit !ok }' ||
  fail "feedback: iteration 1's run log has no '   second review: FAIL (dissent)' after '   review: PASS': $(iterblock "$t/fb.runlog" 1)"
iterblock "$t/fb.runlog" 2 | grep -qxF '   second review: FAIL' || fail "feedback: iteration 2's run log has no '   second review: FAIL' (no dissent): $(iterblock "$t/fb.runlog" 2)"
iterblock "$t/fb.runlog" 3 | grep -qxF '   second review: PASS' || fail "feedback: iteration 3's run log has no '   second review: PASS': $(iterblock "$t/fb.runlog" 3)"
endline "$t/fb" "$DL1" || fail "feedback: the run's output does not end with '$DL1': $(tail -12 "$t/fb")"
endline "$t/fb.runlog" "$DL1" || fail "feedback: the run log does not hold '$DL1'"
i=0
for w in '{"verdict":"FAIL","dissent":true}' '{"verdict":"FAIL","dissent":false}' '{"verdict":"PASS","dissent":false}'; do
  i=$((i + 1)); itr "$t/fb" $i >"$t/it.json"
  [ "$("$JQ" -r .schema "$t/it.json")" = iteration/v3 ] || fail "feedback: iteration $i's record is $("$JQ" -r .schema "$t/it.json"), want iteration/v3"
  "$JQ" -e --argjson w "$w" '.second_review == $w' "$t/it.json" >/dev/null || fail "feedback: iteration $i's second_review is $("$JQ" -c .second_review "$t/it.json"), want $w"
  valid iteration/v3 "$t/it.json" "feedback: iteration $i's record"
  v=$(echo "$w" | "$JQ" -r .verdict)
  f=$(printf '%s/%03d-second-verdict.json' "$t/fb.reports" $i)
  [ -f "$f" ] || fail "feedback: no reports/$(basename "$f")"
  [ "$("$JQ" -r .verdict "$f")" = "$v" ] || fail "feedback: reports/$(basename "$f") holds $("$JQ" -r .verdict "$f"), want $v"
done
[ "$("$JQ" -r .verdict "$t/fb.reports/001-verdict.json" 2>/dev/null)" = PASS ] || fail "feedback: reports/001-verdict.json is not the primary's PASS"
recs=$(phaserecs "$t/fb" second-review)
[ "$(echo "$recs" | grep -c .)" = 3 ] || fail "feedback: $(echo "$recs" | grep -c .) session records of phase second-review, want 3"
for f in $recs; do
  [ "$("$JQ" -r .schema "$f")" = session/v4 ] || fail "feedback: $(basename "$f") is $("$JQ" -r .schema "$f"), want session/v4"
  [ "$("$JQ" -r .model "$f")" = "$SECOND" ] && [ "$("$JQ" -r .agent "$f")" = claude ] || fail "feedback: $(basename "$f") is not claude $SECOND"
done
for f in "$t/fb.sessions"/*.json; do valid "$("$JQ" -r .schema "$f")" "$f" "feedback: $(basename "$f")"; done

# record: the same runs and outcomes; nothing reaches the notes.
CFG="run.second-review record
model.second-review $SECOND"; fixture
runfix "$t/rc"
[ "$rc" = 0 ] || fail "record: vloop run exited $rc: $(tail -8 "$t/rc")"
[ "$(seq_ "$t/rc")" = 'review T1|second T1|review T2|second T2|review T2|second T2' ] || fail "record: the reviews ran as '$(seq_ "$t/rc")'"
nt=$(notes2 "$t/rc")
case "$nt" in *"primary finding on T2"*) ;; *) fail "record: T2's retry found notes '$nt', want the primary's findings";; esac
case "$nt" in *"second review"*) fail "record: T2's notes hold the second review's findings: '$nt'";; esac
[ "$(task status T1)" = done ] && [ "$(task status T2)" = done ] && [ "$(task attempts T1)" = 0 ] && [ "$(task attempts T2)" = 1 ] ||
  fail "record: the outcomes changed: T1 $(task status T1)/$(task attempts T1), T2 $(task status T2)/$(task attempts T2)"
endline "$t/rc" "$DL1" || fail "record: no '$DL1': $(tail -12 "$t/rc")"

# on-pass: only after a primary PASS.
CFG="run.second-review on-pass
model.second-review $SECOND"; fixture
runfix "$t/op"
[ "$rc" = 0 ] || fail "on-pass: vloop run exited $rc: $(tail -8 "$t/op")"
[ "$(seq_ "$t/op")" = 'review T1|second T1|review T2|review T2|second T2' ] || fail "on-pass: the reviews ran as '$(seq_ "$t/op")', want a second review after each primary PASS only"
case "$(notes2 "$t/op")" in *"second review"*) fail "on-pass: T2's notes hold second-review findings";; esac
[ "$(task status T1)" = done ] && [ "$(task status T2)" = done ] && [ "$(task attempts T2)" = 1 ] || fail "on-pass: the outcomes changed"
iterblock "$t/op.runlog" 2 | grep -q 'second review' && fail "on-pass: iteration 2 (a primary FAIL) has a second review line"
itr "$t/op" 2 | "$JQ" -e 'has("second_review")' >/dev/null && fail "on-pass: iteration 2's record has second_review though none ran"
endline "$t/op" 'second review: 2 dissent(s): T1 (primary PASS, second FAIL), T2 (primary PASS, second FAIL)' ||
  fail "on-pass: no 'second review: 2 dissent(s): T1 (primary PASS, second FAIL), T2 (primary PASS, second FAIL)': $(tail -12 "$t/op")"

# both-pass: a second FAIL rejects as a primary FAIL does.
CFG="run.second-review both-pass
model.second-review $SECOND
run.max-attempts 2"; fixture plan1.json
runfix "$t/bp"
[ "$(seq_ "$t/bp")" = 'review T1|second T1|review T1|second T1' ] || fail "both-pass: the reviews ran as '$(seq_ "$t/bp")', want T1 reviewed by both twice"
[ "$(task status T1)" = blocked ] && [ "$(task attempts T1)" = 2 ] || fail "both-pass: T1 is $(task status T1) after $(task attempts T1) attempts, want blocked after 2"
case "$(ev "$t/bp" '^notes T1 ' | sed -n '2s/^notes T1 //p')" in *"second review: second finding on T1"*) ;;
  *) fail "both-pass: T1's retry found notes '$(ev "$t/bp" '^notes T1 ' | sed -n 2p)', want 'second review: second finding on T1'";; esac
[ "$(itr "$t/bp" 1 | "$JQ" -r .outcome)" = rejected ] || fail "both-pass: iteration 1's outcome is $(itr "$t/bp" 1 | "$JQ" -r .outcome), want rejected"
CFG="run.second-review both-pass
model.second-review $SECOND"; fixture
: >"$t/stub/second.pass"
runfix "$t/bq"
[ "$rc" = 0 ] || fail "both-pass, second passing: vloop run exited $rc: $(tail -8 "$t/bq")"
[ "$(seq_ "$t/bq")" = 'review T1|second T1|review T2|review T2|second T2' ] || fail "both-pass, second passing: the reviews ran as '$(seq_ "$t/bq")'"
[ "$(task status T1)" = done ] && [ "$(task status T2)" = done ] || fail "both-pass, second passing: T1 $(task status T1), T2 $(task status T2)"
grep -q 'dissent(s)' "$t/bq" && fail "both-pass, second passing: a dissent list with no dissent"

# off: nothing of it.
CFG=; fixture
runfix "$t/off"
[ "$rc" = 0 ] || fail "off: vloop run exited $rc"
[ -z "$(ev "$t/off" '^session second ')" ] || fail "off: a second review ran"
grep -qi 'second review' "$t/off" "$t/off.runlog" && fail "off: the run speaks of a second review"
"$JQ" -e 'has("second_review")' "$t/off.iters" | grep -q true && fail "off: an iteration record has second_review"
ls "$t/off.reports" | grep -q second && fail "off: a second verdict was kept"
[ -z "$(phaserecs "$t/off" second-review)" ] || fail "off: a second-review session record"

# Faults of the second session, in record mode: none changes T1's outcome.
for fault in touch noverdict; do
  CFG="run.second-review record
model.second-review $SECOND"; fixture plan1.json
  : >"$t/stub/second.$fault"
  runfix "$t/f-$fault"
  [ "$rc" = 0 ] || fail "record, second.$fault: vloop run exited $rc: $(tail -8 "$t/f-$fault")"
  [ "$(task status T1)" = done ] && [ "$(task attempts T1)" = 0 ] || fail "record, second.$fault: T1 is $(task status T1)/$(task attempts T1), want done at 0 attempts"
  [ "$(itr "$t/f-$fault" 1 | "$JQ" -r .second_review.verdict)" = FAIL ] || fail "record, second.$fault: the second verdict is not FAIL"
  iterblock "$t/f-$fault.runlog" 1 | grep -qxF '   second review: FAIL (dissent)' || fail "record, second.$fault: no '   second review: FAIL (dissent)'"
done
CFG="run.second-review record
model.second-review $SECOND"; fixture plan1.json
: >"$t/stub/second.touch"
runfix "$t/f2"
[ "$(cat "$t/repo/docs/x.md")" = x ] && [ "$(venv git -C "$t/repo" show HEAD:docs/x.md)" = x ] || fail "record, second.touch: the second review's change to docs/x.md was kept"
CFG="run.second-review record
model.second-review $SECOND"; fixture plan1.json
: >"$t/stub/second.moveref"
runfix "$t/f3"
[ "$rc" = 9 ] || fail "record, second.moveref: a second review that moved a ref: vloop run exited $rc, want 9: $(tail -5 "$t/f3")"

# A copilot second reviewer: the hook's second-review phase.
CFG="run.second-review record
agent.second-review copilot
model.second-review $SECOND"; fixture plan1.json
runfix "$t/cp"
[ "$rc" = 0 ] || fail "copilot second review: vloop run exited $rc: $(tail -8 "$t/cp")"
[ "$(ev "$t/cp" '^session copilot-second ' | wc -l | tr -d ' ')" = 1 ] || fail "copilot second review: $(ev "$t/cp" '^session copilot-second ' | wc -l | tr -d ' ') copilot second sessions, want 1: $(cat "$t/cp.ev")"
ev "$t/cp" '^hooks ' | grep -qxE 'hooks [1-9][0-9]* bash=yes powershell=yes' || fail "copilot second review: no hooks file with both forms"
ev "$t/cp" '^hook-phase ' | grep -qx 'hook-phase second-review' || fail "copilot second review: the hook is not run with --phase second-review: $(ev "$t/cp" '^hook-phase ')"
ev "$t/cp" '^hook shgh ' | grep -qx 'hook shgh deny 0' || fail "copilot second review: the hook did not deny sh -c gh"
f=$(phaserecs "$t/cp" second-review)
[ -n "$f" ] && [ "$("$JQ" -r .agent "$f")" = copilot ] && [ "$("$JQ" -r .schema "$f")" = session/v4 ] || fail "copilot second review: no session/v4 record of agent copilot"
"$JQ" -e '.permission_denials | map(.tool_use_id) == ["c1","c4"]' "$f" >/dev/null || fail "copilot second review: permission_denials is $("$JQ" -c .permission_denials "$f")"
valid session/v4 "$f" "copilot second review: $(basename "$f")"
echo "gate T5: ok"
