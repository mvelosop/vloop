#!/usr/bin/env bash
# --plan-only buys the cheap half of a run.
#
# The plan is the highest-leverage artefact the loop produces and the cheapest
# thing to get wrong: every gate the rest of the run is measured against was
# authored in that one session, and a weak one silently lowers the bar for
# everything after it. Reading the plan before paying for iterations is the
# whole reason this exists.
#
# It was already possible with LOOP_MAX_ITERATIONS=0. That works, but it reads
# as a workaround and it puts a number in the one place a typo is expensive —
# LOOP_MAX_ITERATIONS=10 is a plausible slip and an unrecoverable one.
#
# The assertion that matters is the last: a plan-only run must leave a plan the
# next invocation can actually execute. A stopping flag that produced something
# unresumable would be worse than the workaround.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"planonly","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"first","goal":"the first thing","files":[],"depends_on":[],"references":[],
  "acceptance":["a"],"verify":"true","status":"pending","attempts":0,"notes":""},
 {"id":"T2","title":"second","goal":"the second thing","files":[],"depends_on":["T1"],"references":[],
  "acceptance":["b"],"verify":"true","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
  work)   jq -nc --arg t "$TASK" '{task:$t,outcome:"done",summary:"s",files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "$TASK" '{task:$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB

note "── --plan-only plans, commits, and stops ──"
fixture_run --plan-only docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_log "plan only"
assert_log "2 task(s)"
# The first ready task is named, because the next thing the operator does is
# read that task's verify command.
assert_log "T1"
assert_iterations 0

n_sessions="$(ls "$FX/repo"/.loop/state/runs/*/*/sessions/*.json 2>/dev/null | wc -l | tr -d ' ')"
[[ "$n_sessions" -eq 1 ]] && ok "exactly one session was paid for" \
  || bad "sessions: got $n_sessions, want 1 (the plan)"
[[ "$(fx_state '[.tasks[]|select(.status!="pending")]|length')" -eq 0 ]] \
  && ok "every task is still pending" || bad "a task changed status without an iteration"
[[ -f "$FX/repo/.loop/state/plan.md" ]] && ok "plan.md was rendered" || bad "no plan.md to read"
git -C "$FX/repo" log --oneline -1 | grep -q '\[loop\] plan' \
  && ok "the plan was committed" || bad "the plan was not committed"

note "── the plan it left is resumable, which is the point ──"
fixture_run
assert_exit 0
[[ "$(fx_state .status)" == "complete" ]] && ok "the resumed run completed" \
  || bad "status: got $(fx_state .status), want complete"
[[ "$(fx_state '[.tasks[]|select(.status=="done")]|length')" -eq 2 ]] \
  && ok "both tasks done" || bad "not every task finished"

note "── --plan-only on an already-planned run does nothing, cheaply ──"
before="$(cat "$FX/repo/.loop/state/state.json")"
fixture_run --plan-only docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_log "already planned"
[[ "$(cat "$FX/repo/.loop/state/state.json")" == "$before" ]] \
  && ok "the finished plan was left alone" || bad "--plan-only modified an existing plan"

assert_no_tool_errors
finish
