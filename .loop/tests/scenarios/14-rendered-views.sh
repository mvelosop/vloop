#!/usr/bin/env bash
# JSON is what the driver computes over; markdown is what a human reads.
# .loop/state/plan.md must track state after every change, and the plan's journal must
# carry the narrative a reader needs — the planner's report at the front and
# the run's outcome at the end — without anyone opening a JSON file.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0

P="$FX/repo/.loop/state/plan.md"
J="$(fx_journal)"

[[ -f "$P" ]] && ok "plan.md rendered" || bad "plan.md not rendered"
grep -q 'Do NOT edit' "$P" && ok "plan.md warns it is generated" || bad "plan.md has no do-not-edit banner"
grep -q '2/2 done' "$P" && ok "plan.md tracks final state" || { bad "plan.md is stale"; grep -m1 'Status' "$P"; }
grep -q -- '- \[x\] \*\*T1\*\*' "$P" && ok "T1 rendered as checked" || bad "T1 not checked in plan.md"
grep -q -- '- \[x\] \*\*T2\*\*' "$P" && ok "T2 rendered as checked" || bad "T2 not checked in plan.md"
grep -q '<details><summary>verify command' "$P" && ok "verify commands are collapsible" || bad "verify not in <details>"

grep -q '^## Plan — fixture' "$J" && ok "journal opens with the plan report" || bad "journal has no plan section"
grep -q '^## Run ended — complete' "$J" && ok "journal closes with the outcome" || bad "journal has no run-end section"
grep -q 'Signals:' "$J" && ok "journal carries the signals" || bad "journal has no signals line"
grep -q '^## T1 —' "$J" && ok "journal has per-iteration entries" || bad "journal missing iteration entries"

# The rendered view is never a source of truth: regenerating from unchanged
# state must reproduce it byte for byte.
cp "$P" "$FX/plan.before"
# ...and it must actually RUN. Swallowing the exit code here is how a broken
# no-arg default survived: render-plan failed, wrote nothing, and the diff
# below passed precisely BECAUSE nothing had happened.
( cd "$FX/repo" && ./.loop/render-plan.sh >/dev/null ) \
  && ok "render-plan runs with no arguments" || bad "render-plan.sh failed with no arguments"
diff -q "$FX/plan.before" "$P" >/dev/null && ok "render is deterministic" || bad "re-render changed the file"
finish
