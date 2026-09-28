#!/usr/bin/env bash
# A blocked task's own gate now runs in the sweep, and when it passes the
# driver says so: the deliverable satisfies its gate, so whatever the
# session could not do is about something else. Failures B and C in the
# brief both read blocked with a complete, green deliverable, and an
# operator had to run the gate by hand to learn that. This does not change
# the status transition -- a blocked outcome still charges an attempt and
# returns the task to pending, because a session's claim about its own work
# is still unverified. Signal, not verdict.
. "$(dirname "$0")/../lib.sh"

blk() {   # <1|0 -- does the work session leave a deliverable that satisfies the gate?>
  fixture_cleanup; fixture_new
  fixture_plan "$PLAN_ONE"
  export MAKE="$1"
  fixture_stub <<STUB
case "\$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
$PLAN_ONE
PLANJSON
    ;;
  work)
    [ "\$MAKE" = 1 ] && touch T1.out
    jq -nc --arg t "\$TASK" '{task:\$t,outcome:"blocked",summary:"stub summary",files:[],verified:"none",notes:"stub summary"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
  LOOP_MAX_ITERATIONS=1 fixture_run docs/briefs/0003-runstat-cli.md
}

# "gate" and "pass" together, naming T1, is the contract -- not exact wording.
says() { grep -iE 'gate.*pass|pass.*gate' | grep -q T1; }

note "-- blocked, and the deliverable satisfies the gate --"
blk 1
assert_iter_outcome 1 blocked
assert_status T1 pending
assert_attempts T1 1
assert_no_log "review: PASS"

notes="$(fx_state '.tasks[0].notes')"
printf 'T1 %s' "$notes" | says \
  && ok "notes say the gate passes: $notes" \
  || bad "notes do not say the gate passes: $notes"

iter_log="$(sed -n '/── iteration/,/═══/p' "$FX/out.log")"
printf '%s\n' "$iter_log" | says \
  && ok "the iteration warning says it" \
  || bad "no iteration warning names T1 and its passing gate"

end_log="$(sed -n '/═══/,$p' "$FX/out.log")"
printf '%s\n' "$end_log" | says \
  && ok "the run-end report says it" \
  || bad "the run-end report does not name T1 and its passing gate"

note "-- blocked, and the gate fails: nothing to report --"
blk 0
assert_iter_outcome 1 blocked
assert_status T1 pending
assert_attempts T1 1
assert_no_log "GATE REGRESSION"
assert_no_log "GATE FAIL"

notes="$(fx_state '.tasks[0].notes')"
printf 'T1 %s' "$notes" | says \
  && bad "notes claim a passing gate: $notes" \
  || ok "notes make no gate claim"

sed -n '/── iteration/,$p' "$FX/out.log" | says \
  && bad "the log claims a passing gate" \
  || ok "the log makes no gate claim"

finish
