#!/usr/bin/env bash
# A review session that returns nothing usable must FAIL, never pass. Failing
# open here would mean a crashed or confused reviewer silently rubber-stamps
# every task, which is indistinguishable from having no reviewer at all.
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
  review) : ;;   # writes no verdict at all
esac
STUB
LOOP_MAX_ATTEMPTS=2 LOOP_CONVERGENCE_MIN=99 fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 2
assert_status T1 blocked
assert_iter_outcome 1 review_fail
assert_log "no valid verdict"
assert_no_log "T1 done"
finish
