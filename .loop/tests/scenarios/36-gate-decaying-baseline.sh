#!/usr/bin/env bash
# Gate-shape lint, rule 4: a gate may not diff, log or rev-list against any ref
# other than HEAD. A gate re-runs for the life of the plan, so a baseline fixed
# at plan time -- a commit, a tag, a branch, a relative ref, a merge-base --
# decays the moment any other task commits (failure E in the brief this rule
# answers: a scope guard baselined on the plan commit went unpassable the
# moment either stack closed its first task, and reverted six finished tasks).
# A range that merely ENDS in HEAD is still rejected -- the left side is the
# fixed part and is what decays.
#
# git diff/log/rev-list against HEAD itself stays allowed: a work session
# cannot commit, so HEAD still discriminates a session's edits from committed
# history, which is what the loop's own gate-rewrite guard depends on.
. "$(dirname "$0")/../lib.sh"

plan_one() {   # <verify>
  fixture_plan "$(jq -nc --arg v "$1" '
    {run_id:"fixture",brief:"docs/briefs/0003-runstat-cli.md",status:"running",
     iteration:0,created:"2026-08-15T00:00:00Z",updated:"2026-08-15T00:00:00Z",
     tasks:[{id:"T1",title:"t",goal:"g",files:[],depends_on:[],acceptance:["a"],
             verify:$v,status:"pending",attempts:0,notes:""}]}')"
}

rej() {   # <verify> <ref that should be named>
  note "-- rejected: diffs against $2 --"
  fixture_new
  plan_one "$1"
  fixture_stub_default
  fixture_run --plan-only docs/briefs/0003-runstat-cli.md
  assert_exit 1
  assert_log "gate shape rejected"
  assert_log "T1 .*$2"
  assert_iterations 0
}

acc() {   # <verify>
  note "-- accepted: diffs against HEAD --"
  fixture_new
  plan_one "$1"
  fixture_stub_default
  fixture_run --plan-only docs/briefs/0003-runstat-cli.md
  assert_exit 0
  assert_no_log "gate shape rejected"
}

rej 'test -z "$(git diff --name-only 4c911d7 -- docs)"' '4c911d7'
rej 'test -z "$(git diff --name-only HEAD~1 -- docs)"' 'HEAD~1'
rej 'test -z "$(git log --format=%H v1.0..HEAD -- docs)"' 'v1.0'
rej 'test "$(git rev-list --count origin/main..HEAD)" -gt 0' 'origin/main'
rej 'git diff --quiet "$(git merge-base HEAD main)" -- docs' ''

acc 'test -z "$(git diff --name-only HEAD -- docs)"'
acc 'git diff --quiet HEAD -- docs'

assert_no_tool_errors
finish
