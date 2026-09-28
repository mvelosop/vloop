#!/usr/bin/env bash
# A branch cut from main inherits whatever .loop/state/state.json the last squash left
# there. Resuming that as if it were this branch's work is silent nonsense.
#
# The discriminator is the BRIEF, not the branch. A branch name cannot separate
# "someone else's inherited plan" from "my plan, on a branch I renamed" — the
# two look identical and want opposite things, and guessing wrong in the
# destructive direction loses a run. The brief says which plan you are asking
# for, so it answers directly.
. "$(dirname "$0")/../lib.sh"

fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
[[ "$(fx_state '.branch')" == "$(git -C "$FX/repo" branch --show-current)" ]] \
  && ok "state stamped with its owning branch" || bad "state not stamped"

note "── a DIFFERENT brief: a different plan, so reset ──"
( cd "$FX/repo" && cp docs/briefs/0003-runstat-cli.md docs/briefs/0009-other.md )
# Not PLAN_TWO again: the first run already committed its artifacts, T1.out and
# T2.out, to HEAD -- so a second plan reusing PLAN_TWO's own verify commands
# would now name a HEAD file neither of ITS tasks owns, which is gate-shape
# rule 3, not what a reset-vs-resume test means to exercise. Swap in a plan
# whose verify never names a path.
fixture_plan '{"run_id":"fixture2","brief":"docs/briefs/0009-other.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"First","goal":"g","files":[],"depends_on":[],
  "acceptance":["ok"],"verify":"true","status":"pending","attempts":0,"notes":""},
 {"id":"T2","title":"Second","goal":"g","files":[],"depends_on":["T1"],
  "acceptance":["ok"],"verify":"true","status":"pending","attempts":0,"notes":""}]}'
fixture_stub_default
fixture_run docs/briefs/0009-other.md
assert_exit 0
assert_log "resetting and planning fresh"
last="$(ls -dt "$FX/repo"/.loop/state/runs/*/*/ | head -1)"
ls "$last"sessions/*-plan.json >/dev/null 2>&1 \
  && ok "a plan session ran — genuinely reset, not resumed" \
  || bad "no plan session: the old plan was resumed instead of reset"
[[ "$(fx_state '.brief')" == "docs/briefs/0009-other.md" ]] \
  && ok "state now holds the plan that was asked for" || bad "state brief did not change"

note "── same brief, branch RENAMED: must resume, never destroy ──"
( cd "$FX/repo" && jq '.branch = "the-old-branch-name"' .loop/state/state.json > s.tmp && mv s.tmp .loop/state/state.json )
before="$(fx_state '.run_id')"
fixture_run docs/briefs/0009-other.md
assert_exit 0
assert_no_log "resetting and planning fresh"
[[ "$(fx_state '.run_id')" == "$before" ]] \
  && ok "the renamed branch's plan survived ($before)" \
  || bad "renaming the branch destroyed the plan"
last="$(ls -dt "$FX/repo"/.loop/state/runs/*/*/ | head -1)"
ls "$last"sessions/*-plan.json >/dev/null 2>&1 \
  && bad "it re-planned instead of resuming" \
  || ok "no plan session — it resumed, as it should"

note "── no brief, and the stamp disagrees: refuse, do not guess ──"
( cd "$FX/repo" && jq '.branch = "the-old-branch-name"' .loop/state/state.json > s.tmp && mv s.tmp .loop/state/state.json )
fixture_run
assert_exit 1
assert_log "no way to tell an inherited plan"
[[ -f "$FX/repo/.loop/state/state.json" ]] && ok "refusing left the state untouched" \
  || bad "refusing destroyed the state"
finish
