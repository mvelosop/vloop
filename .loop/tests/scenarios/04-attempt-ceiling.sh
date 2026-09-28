#!/usr/bin/env bash
# A task that keeps failing review is blocked, not retried forever.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_ONE"
fixture_stub <<STUB
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$PLAN_ONE
PLANJSON
    ;;
  work) touch "\$TASK.out"
    jq -nc --arg t "\$TASK" '{task:\$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"FAIL",criteria:[],findings:["nope"],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
LOOP_MAX_ATTEMPTS=2 LOOP_CONVERGENCE_MIN=99 fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 2
assert_run_status blocked
assert_status T1 blocked
assert_attempts T1 2
assert_log "attempt ceiling"
finish
