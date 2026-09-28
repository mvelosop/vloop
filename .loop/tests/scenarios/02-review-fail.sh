#!/usr/bin/env bash
# A review FAIL must revert the task, charge an attempt, and let a later
# iteration retry it. The gate cannot catch what this catches.
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
  review)
    # fail once, then pass
    if [ -f .reviewed ]; then v=PASS; f='[]'; else v=FAIL; f='["not really done"]'; touch .reviewed; fi
    jq -nc --arg t "\$TASK" --arg v "\$v" --argjson f "\$f" '{task:\$t,verdict:\$v,criteria:[],findings:\$f,notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_iterations 2
assert_iter_outcome 1 review_fail
assert_iter_outcome 2 done
assert_status T1 done
assert_attempts T1 1
finish
