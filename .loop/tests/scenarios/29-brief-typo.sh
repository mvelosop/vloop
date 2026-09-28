#!/usr/bin/env bash
# A typo must not be able to destroy a committed plan.
#
# Naming a brief other than the one the current plan holds resets that plan and
# starts fresh — which is right, and is how a branch refuses to resume another
# branch's inherited state. But the reset used to happen without first checking
# that the brief you named exists, so a mistyped path, or `run.sh --help`, read
# as "a different brief" and deleted a finished plan on the way to failing.
#
# Found by running `.loop/run.sh --help` against this repo and watching a
# completed plan disappear. A resumable run is the loop's central promise and a
# slip of the keyboard must not be able to spend it.
. "$(dirname "$0")/../lib.sh"
fixture_new

plant_state() {
  mkdir -p "$FX/repo/.loop/state"
  cat >"$FX/repo/.loop/state/state.json" <<'PLANJSON'
{"run_id":"0003-runstat-cli","brief":"docs/briefs/0003-runstat-cli.md","branch":"main",
 "status":"complete","iteration":4,"created":"2026-08-15T00:00:00Z","updated":"2026-08-15T01:00:00Z",
 "tasks":[{"id":"T1","title":"t","goal":"g","files":[],"depends_on":[],"references":[],
 "acceptance":["a"],"verify":"true","status":"done","attempts":1,"notes":""}]}
PLANJSON
  printf 'rendered\n' >"$FX/repo/.loop/state/plan.md"
}

note "── a mistyped brief path leaves the plan alone ──"
plant_state
before="$(cat "$FX/repo/.loop/state/state.json")"
( cd "$FX/repo" && HOME="$FX/testuser" bash .loop/run.sh docs/briefs/0003-runstat-cli-typo.md >"$FX/out.log" 2>&1 )
rc=$?
[[ $rc -eq 2 ]] && ok "exit 2" || bad "exit: got $rc, want 2"
assert_log "brief not found"
assert_log "the current plan is untouched"
[[ -f "$FX/repo/.loop/state/state.json" ]] && ok "state.json still exists" || bad "state.json was deleted by a typo"
[[ "$(cat "$FX/repo/.loop/state/state.json" 2>/dev/null)" == "$before" ]] \
  && ok "the plan is byte-identical" || bad "the plan was modified"
[[ -f "$FX/repo/.loop/state/plan.md" ]] && ok "plan.md survived" || bad "plan.md was deleted by a typo"
# Nothing opened means nothing to clean up: no lock, no run directory.
[[ ! -e "$FX/repo/.loop/tmp/.running" ]] && ok "no lock was taken" || bad "a lock was left behind"
[[ -z "$(ls -A "$FX/repo/.loop/state/runs" 2>/dev/null)" ]] \
  && ok "no run directory was created" || bad "a run directory was created for a run that never started"

note "── --help prints usage and changes nothing ──"
( cd "$FX/repo" && HOME="$FX/testuser" bash .loop/run.sh --help >"$FX/out.log" 2>&1 )
rc=$?
[[ $rc -eq 0 ]] && ok "exit 0" || bad "exit: got $rc, want 0"
assert_log "usage"
[[ "$(cat "$FX/repo/.loop/state/state.json" 2>/dev/null)" == "$before" ]] \
  && ok "--help left the plan alone" || bad "--help modified the plan"

note "── an unknown option is refused, not read as a brief ──"
( cd "$FX/repo" && HOME="$FX/testuser" bash .loop/run.sh --resume >"$FX/out.log" 2>&1 )
rc=$?
[[ $rc -eq 2 ]] && ok "exit 2" || bad "exit: got $rc, want 2"
assert_log "unknown option"
[[ "$(cat "$FX/repo/.loop/state/state.json" 2>/dev/null)" == "$before" ]] \
  && ok "an unknown option left the plan alone" || bad "an unknown option modified the plan"

finish
