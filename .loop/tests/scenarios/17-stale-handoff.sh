#!/usr/bin/env bash
# .loop/tmp/proposal.json is transient. If the driver did not clear it before each
# iteration, a work session that wrote nothing would silently inherit the
# previous iteration's report — a task marked done on the strength of another
# task's evidence.
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
    # Only T1 reports. T2's session writes nothing, leaving T1's proposal on
    # disk as the most recent one.
    if [ "\$TASK" = T1 ]; then
      jq -nc --arg t "\$TASK" '{task:\$t,outcome:"done",summary:"T1 report",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json
    fi ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
LOOP_MAX_ATTEMPTS=2 LOOP_CONVERGENCE_MIN=99 fixture_run docs/briefs/0003-runstat-cli.md
assert_status T1 done
assert_iter_outcome 2 blocked
assert_log "no valid proposal"
grep -q 'T1 report' "$(fx_journal)" && ok "T1's report is in its own entry" || bad "T1 report missing"
[[ "$(grep -c 'T1 report' "$(fx_journal)")" == "1" ]] \
  && ok "T1's report was not reused for T2" || bad "stale proposal leaked into a second entry"
finish
