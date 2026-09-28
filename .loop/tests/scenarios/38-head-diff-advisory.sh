#!/usr/bin/env bash
# amend.sh check(), third advisory: a `git diff ... HEAD` scope guard that
# never reads LOOP_ACTIVE_TASK or LOOP_GATE_TASK (T3's two gate-only
# environment variables).
#
# Rule 4's gate-shape lint permits a HEAD diff because a work session cannot
# commit, so HEAD still separates that session's own edits from everything
# committed before it -- sound for the iteration that IS that session. But a
# finished task's gate re-runs for the life of the plan as a regression check
# of every LATER iteration too, and at that moment the only uncommitted work
# in the tree belongs to whichever task the driver is working now, not to the
# task whose gate is running. A guard that never consults the two variables
# reads that other task's in-progress edits as its own regression -- failure E
# in the brief, six false reversions from exactly this. Advisory, like the two
# smells already in check(): it does not fail the plan.
#
# amend.sh is not part of the fixture harness's throwaway repo (fixture_new
# copies run.sh, render-plan.sh and settings.json, not amend.sh), so this
# scenario builds its own tiny repo directly, the same shape as this task's
# own cheap direct gate in state.json.
. "$(dirname "$0")/../lib.sh"

D="$(mktemp -d "${TMPDIR:-/tmp}/loopfx.XXXXXX")"
trap 'rm -rf "$D"' EXIT
mkdir -p "$D/.loop/state"
cp "$SUITE_ROOT/../amend.sh" "$SUITE_ROOT/../render-plan.sh" "$D/.loop/"
git -C "$D" init -q
git -C "$D" -c user.email=fixture@test -c user.name=fixture commit -q --allow-empty -m init

plant() {   # <T2 verify>
  jq -nc --arg v "$1" '
    {run_id:"fx",brief:"docs/briefs/0003-runstat-cli.md",status:"running",iteration:0,
     created:"2026-08-15T00:00:00Z",updated:"2026-08-15T00:00:00Z",
     tasks:[
       {id:"T1",title:"t",goal:"g",files:[],depends_on:[],acceptance:["a"],
        verify:"test -f .never",status:"pending",attempts:0,notes:""},
       {id:"T2",title:"t",goal:"g",files:[],depends_on:[],acceptance:["a"],
        verify:$v,status:"pending",attempts:0,notes:""}
     ]}' >"$D/.loop/state/state.json"
}

chk() {   # <advise|quiet> <T2 verify>
  plant "$2"
  bash "$D/.loop/amend.sh" check >"$D/out" 2>&1
  rc=$?
  [[ $rc -eq 0 ]] && ok "check exits 0: $2" || bad "check exit $rc: $2"
  if [[ "$1" == advise ]]; then
    grep -q T2 "$D/out" && ok "advised about T2: $2" || bad "not advised: $2"
  else
    grep -q T2 "$D/out" && bad "advised wrongly: $2" || ok "no advisory: $2"
  fi
}

note "-- advised: diffs against HEAD, reads neither variable --"
chk advise 'test -f .never && test -z "$(git diff --name-only HEAD -- docs)"'
chk advise 'test -f .never && git diff --quiet HEAD -- docs'

note "-- quiet: diffs against HEAD, but reads one of the two variables --"
chk quiet 'test -f .never && { [ "$LOOP_GATE_TASK" != "$LOOP_ACTIVE_TASK" ] || test -z "$(git diff --name-only HEAD -- docs)"; }'
chk quiet 'test -f .never && test -n "${LOOP_ACTIVE_TASK:-}" && git diff --quiet HEAD -- docs'
chk quiet 'test -f .never && test -n "${LOOP_GATE_TASK:-}" && git diff --quiet HEAD -- docs'

note "-- quiet: no HEAD diff at all --"
chk quiet 'test -f .never'

finish
