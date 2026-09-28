#!/usr/bin/env bash
# Two properties that let loops run in parallel without fighting at merge time:
#
#   1. Telemetry lives under .loop/state/runs/<branch>/<timestamp>/, so two loops that
#      start in the same second still write to different paths. Second-
#      resolution timestamps alone are not enough.
#   2. The journal is per PLAN, not per repo. A single .loop/journal.md is the
#      only append-shared file in the system — it conflicts between branches,
#      and because it is created only when missing it silently accumulates
#      unrelated plans one after another.
#
# Per-plan must not be confused with per-run: a resumed run has to keep
# appending to the same narrative, or the work session's "read the last two
# entries" loses its handoff. That is the third assertion here.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default

LOOP_MAX_ITERATIONS=1 fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 4

branch="$(git -C "$FX/repo" branch --show-current)"
runs="$(cd "$FX/repo" && ls -d .loop/state/runs/*/ 2>/dev/null | head -1)"
[[ "$runs" == ".loop/state/runs/$branch/" ]] \
  && ok "telemetry grouped by branch: $runs" \
  || bad "run dir is not branch-scoped: got '$runs', want '.loop/state/runs/$branch/'"

[[ -f "$FX/repo/.loop/state/journals/fixture.md" ]] \
  && ok "journal is per plan: .loop/state/journals/fixture.md" \
  || bad "no per-plan journal at .loop/state/journals/fixture.md"

[[ ! -f "$FX/repo/.loop/journal.md" ]] \
  && ok "no repo-wide .loop/journal.md" \
  || bad "the old shared journal is still being written"

before="$(grep -c '^## ' "$(fx_journal)")"
note "resuming — the narrative must continue, not restart"
fixture_run
assert_exit 0

n="$(cd "$FX/repo" && ls .loop/state/journals/*.md | wc -l | tr -d ' ')"
[[ "$n" == "1" ]] && ok "still one journal for the plan across both runs" \
  || bad "a resumed run started a new journal ($n files)"

after="$(grep -c '^## ' "$(fx_journal)")"
[[ "$after" -gt "$before" ]] \
  && ok "resumed run appended to it ($before → $after sections)" \
  || bad "resumed run did not append (before $before, after $after)"

grep -q '^## T1 —' "$(fx_journal)" && grep -q '^## T2 —' "$(fx_journal)" \
  && ok "both runs' iterations are in one narrative" \
  || bad "the narrative is split across runs"

# Two runs back to back land in the same second. Second-resolution timestamps
# alone would have them share a directory, and the later one truncates the
# earlier one's iterations.jsonl — which made scenario 21 flaky before the
# driver started disambiguating.
runs_n="$(cd "$FX/repo" && ls -d .loop/state/runs/"$branch"/*/ | wc -l | tr -d ' ')"
[[ "$runs_n" == "2" ]] && ok "each run got its own telemetry dir, same second or not" \
  || bad "expected 2 run dirs under the branch, got $runs_n"
finish
