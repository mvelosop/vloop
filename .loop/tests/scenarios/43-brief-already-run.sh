#!/usr/bin/env bash
# Nothing retires a brief, so a spent one reads plannable forever. A run's
# journal is named for the brief's own stem -- the number-and-slug a planner
# writes as run_id -- so its existence is the check: no stamp, no new field.
#
# check-brief.sh must report an already-run brief as a problem, and run.sh
# must refuse to plan from one (before touching state.json or plan.md, and
# before spending a session) unless the operator passes the explicit
# --replan override.
set -uo pipefail
. "$(dirname "${BASH_SOURCE[0]}")/../lib.sh"

note "── check-brief.sh: an already-run brief is reported ──"
out="$( (cd "$REPO_ROOT" && .loop/check-brief.sh docs/briefs/0003-runstat-cli.md) 2>&1 )"
rc=$?
[[ $rc -eq 1 ]] && ok "check-brief: an already-run brief exits 1" \
  || bad "check-brief on an already-run brief exited $rc"
case "$out" in
  *.loop/state/journals/0003-runstat-cli.md*) ok "it names the journal" ;;
  *) bad "check-brief does not name the journal: $out" ;;
esac

note "── check-brief.sh: a never-run plannable brief is unaffected ──"
( cd "$REPO_ROOT" && .loop/check-brief.sh docs/briefs/0005-production-and-insights.md ) >/dev/null 2>&1 \
  && ok "a never-run brief still passes" \
  || bad "a never-run plannable brief fails check-brief"

# A throwaway repo: a plannable brief with its journal already committed, and
# a second, never-run brief with the same body -- so the fixture can prove
# the refusal is keyed on the journal, not on anything else about the brief.
fx() {
  fixture_cleanup; fixture_new
  printf '# brief\n\n- **Status:** ready to plan\n' >"$FX_REPO/docs/briefs/0003-runstat-cli.md"
  cp "$FX_REPO/docs/briefs/0003-runstat-cli.md" "$FX_REPO/docs/briefs/0009-fresh.md"
  mkdir -p "$FX_REPO/.loop/state/journals"
  echo '# Journal' >"$FX_REPO/.loop/state/journals/0003-runstat-cli.md"
  git -C "$FX_REPO" add -A
  git -C "$FX_REPO" commit -qm pre
  fixture_plan "$PLAN_ONE"
  fixture_stub_default
}

note "── run.sh: refuses to plan a brief whose journal already exists ──"
fx
fixture_run --plan-only docs/briefs/0003-runstat-cli.md
assert_exit 1
assert_log '.loop/state/journals/0003-runstat-cli.md'
assert_no_state
assert_no_log 'planning from'

note "── run.sh: refuses before an existing plan for another brief is reset ──"
fx
mkdir -p "$FX_REPO/.loop/state"
printf '%s' "$PLAN_ONE" | jq '.brief="docs/briefs/0009-fresh.md"' >"$FX_REPO/.loop/state/state.json"
git -C "$FX_REPO" add -A
git -C "$FX_REPO" commit -qm plan
S0="$(cksum <"$FX_REPO/.loop/state/state.json")"
fixture_run --plan-only docs/briefs/0003-runstat-cli.md
assert_exit 1
[[ "$(cksum <"$FX_REPO/.loop/state/state.json" 2>/dev/null)" == "$S0" ]] \
  && ok "the existing plan is untouched" \
  || bad "the refusal destroyed or changed the existing plan"

note "── run.sh --replan: the explicit override re-plans anyway ──"
fx
fixture_run --plan-only --replan docs/briefs/0003-runstat-cli.md
assert_exit 0
[[ -f "$FX_REPO/.loop/state/state.json" ]] && ok "plan written" || bad "no plan written under --replan"

note "── run.sh: a plannable brief that was never run plans normally ──"
fx
fixture_run --plan-only docs/briefs/0009-fresh.md
assert_exit 0
[[ -f "$FX_REPO/.loop/state/state.json" ]] && ok "plan written" || bad "no plan written for a fresh brief"

finish
