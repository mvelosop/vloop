#!/usr/bin/env bash
# A plan with no verify command is a planning failure. Catch it before any
# iteration inherits it — every downstream session is measured against it.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"bad","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"No gate","goal":"g","files":[],"depends_on":[],
  "acceptance":["something"],"verify":"","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 1
assert_log "no verify command"
assert_iterations 0
finish
