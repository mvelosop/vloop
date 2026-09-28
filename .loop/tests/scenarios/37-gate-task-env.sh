#!/usr/bin/env bash
# The driver hands every gate two ids: LOOP_ACTIVE_TASK, the task this
# iteration is working, and LOOP_GATE_TASK, the task whose verify is running
# right now. They differ exactly when a gate re-runs as a regression check of
# a task other than the one being worked -- the case that used to have no
# session whose scope it could speak to, and cost the arc six false
# reversions (failure E in the brief this answers).
#
# Outside the gate sweep neither variable is set, even when the driver's own
# environment carries a stale value inherited from whatever invoked it --
# checked below by exporting both onto fixture_run itself, the way a gate
# re-running run-all.sh would inherit them from an outer driver.
. "$(dirname "$0")/../lib.sh"

fixture_new
export FXE="$FX/env.log"

V1='test -f T1.out && printf "%s %s\n" "${LOOP_ACTIVE_TASK-unset}" "${LOOP_GATE_TASK-unset}" >> "$FXE"'
V2='test -f T2.out && printf "%s %s\n" "${LOOP_ACTIVE_TASK-unset}" "${LOOP_GATE_TASK-unset}" >> "$FXE"'

FX_PLAN="$(jq -nc --arg v1 "$V1" --arg v2 "$V2" '
  {run_id:"fixture",brief:"docs/briefs/0003-runstat-cli.md",status:"running",
   iteration:0,created:"2026-08-15T00:00:00Z",updated:"2026-08-15T00:00:00Z",
   tasks:[
     {id:"T1",title:"First",goal:"g",files:[],depends_on:[],acceptance:["a"],
      verify:$v1,status:"pending",attempts:0,notes:""},
     {id:"T2",title:"Second",goal:"g",files:[],depends_on:["T1"],acceptance:["a"],
      verify:$v2,status:"pending",attempts:0,notes:""}
   ]}')"
fixture_plan "$FX_PLAN"

fixture_stub <<STUB
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$FX_PLAN
PLANJSON
    ;;
  work)
    touch "\$TASK.out"
    printf "work %s %s\n" "\${LOOP_ACTIVE_TASK-unset}" "\${LOOP_GATE_TASK-unset}" >> "$FXE"
    jq -nc --arg t "\$TASK" '{task:\$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review)
    printf "review %s %s\n" "\${LOOP_ACTIVE_TASK-unset}" "\${LOOP_GATE_TASK-unset}" >> "$FXE"
    jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB

# A stale value already in the environment when the driver starts -- as
# happens when a gate re-runs run-all.sh inside itself -- must not leak into
# any session either.
LOOP_ACTIVE_TASK=stale LOOP_GATE_TASK=stale fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_status T1 done
assert_status T2 done

note "recorded: $(tr '\n' '|' <"$FXE")"
for l in 'T1 T1' 'T2 T1' 'T2 T2'; do
  grep -qx "$l" "$FXE" && ok "a gate saw active/gate = $l" || bad "no gate saw active/gate = $l"
done
[[ "$(grep -cx 'work unset unset' "$FXE")" == 2 ]] \
  && ok 'work sessions see neither variable' || bad 'a work session saw a gate variable'
[[ "$(grep -cx 'review unset unset' "$FXE")" == 2 ]] \
  && ok 'review sessions see neither variable' || bad 'a review session saw a gate variable'

assert_no_tool_errors
finish
