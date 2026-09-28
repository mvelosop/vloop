#!/usr/bin/env bash
# A work session that keeps reporting blocked makes no recorded progress.
# Stop rather than spin.
#
# Each attempt touches a fresh marker file so the repeat-blocked halt (item 7,
# brief B20260924-1947 -- a SECOND blocked with nothing changed since the
# first) never fires here: something changes every time, so that rule keeps
# handing the retry back to the stall counter, which is what this scenario
# exists to exercise.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_ONE"
fixture_stub <<STUB
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$PLAN_ONE
PLANJSON
    ;;
  work)
    n=\$(( \$(cat n 2>/dev/null || echo 0) + 1 )); echo \$n > n
    touch "attempt-\$n.marker"
    jq -nc --arg t "\$TASK" '{task:\$t,outcome:"blocked",summary:"cannot proceed",files:[],verified:"",notes:"none"}' > .loop/tmp/proposal.json ;;
esac
STUB
LOOP_MAX_ATTEMPTS=99 LOOP_CONVERGENCE_MIN=99 fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 3
assert_run_status stalled
assert_log "no recorded progress"
finish
