#!/usr/bin/env bash
# The headline feature: a later task breaks an earlier one. Per-task gates
# cannot see this; re-running EVERY done task's verify command can.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub <<STUB
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$PLAN_TWO
PLANJSON
    ;;
  work)
    touch "\$TASK.out"
    # T2 clobbers T1's output, once.
    if [ "\$TASK" = T2 ] && [ ! -f .broke ]; then rm -f T1.out; touch .broke; fi
    jq -nc --arg t "\$TASK" '{task:\$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_log "GATE REGRESSION T1"
assert_attempts T1 1
assert_status T1 done
assert_status T2 done
assert_iter_task 3 T1
finish
