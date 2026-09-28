#!/usr/bin/env bash
# A claude session that dies is not a task failure — it is an infrastructure
# failure, and charging a task an attempt for it would burn the retry budget on
# something the task never did.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_ONE"
fixture_stub <<STUB
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$PLAN_ONE
PLANJSON
    ;;
  work) exit 9 ;;   # dies without writing a proposal
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 7
assert_run_status session_error
assert_status T1 pending
assert_attempts T1 0
finish
