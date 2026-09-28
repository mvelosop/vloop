#!/usr/bin/env bash
# Rule 7: nothing else in the loop notices a run going nowhere. Every gate is
# green here and every review is thorough — and nothing ever closes.
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
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"FAIL",criteria:[],findings:["never good enough"],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
LOOP_CONVERGENCE_MIN=2 LOOP_MAX_ATTEMPTS=99 fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 5
assert_run_status not_converging
assert_log "not converging"
finish
