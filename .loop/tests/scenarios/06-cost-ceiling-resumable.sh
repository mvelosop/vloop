#!/usr/bin/env bash
# Same promise for the cost ceiling: stop, raise it, re-run, finish.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub <<STUB
STUB_COST=1.00
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$PLAN_TWO
PLANJSON
    ;;
  work) touch "\$TASK.out"
    jq -nc --arg t "\$TASK" '{task:\$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
LOOP_COST_CEILING=2 fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 6
assert_run_status cost_ceiling
assert_log "cost ceiling"
note "raising the ceiling and re-running"
LOOP_COST_CEILING=99 fixture_run
assert_exit 0
assert_run_status complete
finish
