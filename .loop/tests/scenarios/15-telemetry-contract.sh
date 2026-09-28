#!/usr/bin/env bash
# The seam with runstat. Brief 0003's "Input format" section is a contract the
# driver has to emit and runstat has to read; scenario 12 only checks the
# formulas over a hand-built fixture, so without this nothing would notice the
# driver drifting away from the shape runstat parses.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0

RUN="$(ls -d "$FX/repo"/.loop/state/runs/*/*/ | head -1)"

n="$(ls "$RUN"sessions/*.json 2>/dev/null | wc -l | tr -d ' ')"
[[ "$n" == "5" ]] && ok "5 session files (1 plan + 2x work/review)" || bad "session files: got $n, want 5"

names="$(cd "$RUN"sessions && ls *.json | tr '\n' ' ')"
[[ "$names" == "001-plan.json 002-work.json 003-review.json 004-work.json 005-review.json " ]] \
  && ok "session files are name-sortable in run order" || bad "session names: $names"

bad_phase="$(jq -s '[.[] | select((.phase | IN("plan","work","review")) | not)] | length' "$RUN"sessions/*.json)"
[[ "$bad_phase" == "0" ]] && ok "every session stamped with a valid phase" || bad "$bad_phase session(s) have no valid phase"

bad_iter="$(jq -s '[.[] | select((.iteration | type) != "number")] | length' "$RUN"sessions/*.json)"
[[ "$bad_iter" == "0" ]] && ok "every session stamped with a numeric iteration" || bad "$bad_iter session(s) lack a numeric iteration"

pi="$(jq -r 'select(.phase=="plan") | .iteration' "$RUN"sessions/001-plan.json)"
[[ "$pi" == "0" ]] && ok "plan session is iteration 0" || bad "plan iteration: got $pi, want 0"

pairs="$(jq -s -r '[.[] | select(.phase!="plan") | "\(.phase):\(.iteration)"] | join(" ")' "$RUN"sessions/*.json)"
[[ "$pairs" == "work:1 review:1 work:2 review:2" ]] && ok "work/review sessions pair per iteration" || bad "pairs: $pairs"

# The fields runstat reads must survive the driver's rewrite of the result JSON.
missing="$(jq -s '[.[] | select(has("total_cost_usd") and has("num_turns") and has("duration_ms") and has("is_error") and has("permission_denials") | not)] | length' "$RUN"sessions/*.json)"
[[ "$missing" == "0" ]] && ok "every session keeps the fields runstat reads" || bad "$missing session(s) lost a field runstat needs"

keys="$(jq -s -r '[.[] | keys] | unique | .[0] | join(",")' "$RUN"iterations.jsonl)"
[[ "$keys" == "attempts,iteration,task,tasks_done,tasks_total,outcome" || "$keys" == "attempts,iteration,outcome,task,tasks_done,tasks_total" ]] \
  && ok "iterations.jsonl carries exactly the six contract keys" || bad "iteration keys: $keys"

bad_out="$(jq -s '[.[] | select((.outcome | IN("done","gate_fail","review_fail","blocked")) | not)] | length' "$RUN"iterations.jsonl)"
[[ "$bad_out" == "0" ]] && ok "every outcome is one of the four contract values" || bad "$bad_out record(s) have an unknown outcome"
finish
