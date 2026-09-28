#!/usr/bin/env bash
# GATE REGRESSION <id> used to name only the task being reverted, never the
# task whose iteration caused the revert -- so when the cause is that task's
# own gate misreading the active task's uncommitted work as its own, the
# message names the only innocent party. This scenario has T2 clobber T1's
# output, which reverts T1 during T2's iteration, and asserts the warning
# names both T1 (reverted) and T2 (being worked).
. "$(dirname "$0")/../lib.sh"

unset LOOP_ACTIVE_TASK LOOP_GATE_TASK

fixture_new
export FXP="$FX/plan.json"
jq -nc '{run_id:"fx",brief:"docs/briefs/0003-runstat-cli.md",status:"running",
  iteration:0,created:"2026-08-15T00:00:00Z",updated:"2026-08-15T00:00:00Z",
  tasks:[
    {id:"T1",title:"a",goal:"g",files:[],depends_on:[],acceptance:["a"],
     verify:"test -f T1.out",status:"pending",attempts:0,notes:""},
    {id:"T2",title:"b",goal:"g",files:[],depends_on:["T1"],acceptance:["a"],
     verify:"test -f T2.out",status:"pending",attempts:0,notes:""}
  ]}' > "$FXP"
printf '%s\n' 'case "$PHASE" in
  plan) cp "$FXP" .loop/state/state.json ;;
  work)
    touch "$TASK.out"
    if [ "$TASK" = T2 ] && [ ! -f .broke ]; then rm -f T1.out; touch .broke; fi
    jq -nc --arg t "$TASK" "{task:\$t,outcome:\"done\",summary:\"s\",files:[],verified:\"ok\",notes:\"none\"}" > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" "{task:\$t,verdict:\"PASS\",criteria:[],findings:[],notes:\"none\"}" > .loop/tmp/verdict.json ;;
esac' | fixture_stub
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_attempts T1 1
assert_status T1 done
assert_status T2 done

L="$(grep 'GATE REGRESSION T1' "$FX/out.log")"
[ -n "$L" ] && ok "regression line: $L" || bad 'no GATE REGRESSION T1 line'
printf '%s' "$L" | sed 's/GATE REGRESSION T1//' | grep -q T2 \
  && ok 'it names T2, the task being worked' \
  || bad "the regression line does not name T2: $L"

finish
