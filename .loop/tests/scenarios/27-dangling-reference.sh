#!/usr/bin/env bash
# A task's references are what the work session is told to read and what the
# review session holds the diff against. A path that does not resolve is
# therefore worse than no reference at all: the work session cannot recover
# from it, and the way it finds out is by burning an attempt.
#
# exploring-claude earned this rule the expensive way — a cited-but-missing
# guideline blocked its implementer — and it is mechanically checkable, so it
# is checked at plan time rather than written down as advice.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"badref","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"Load a run","goal":"g","files":[],"depends_on":[],
  "references":[{"path":"docs/runstat.md","why":"the signal formulas"},
                {"path":"docs/never-written.md","why":"a document nobody wrote"}],
  "acceptance":["it loads"],"verify":"test -f loaded","status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md

assert_exit 1
assert_log "do not resolve"
assert_log "docs/never-written.md"
# Refused before anything is spent, like every other plan-validation failure.
assert_iterations 0
assert_no_tool_errors
finish
