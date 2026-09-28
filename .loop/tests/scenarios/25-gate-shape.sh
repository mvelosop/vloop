#!/usr/bin/env bash
# A gate that parses a structure and then substring-matches its re-serialised
# text has thrown away the parse. It rejects correct implementations (a $ref
# does not contain the property name) and accepts corrupted ones (a hand-written
# duplicate does) — exactly backwards, and both halves were observed in one real
# run of the loop against a NestJS target.
#
# The rule cannot be stated unambiguously in prose, so it is stated as a check.
# What makes that possible is that the compliant form is essentially unique:
# navigate to the value and assert on it.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_stub <<'STUB'
case "$PHASE" in
  plan) cat > .loop/state/state.json <<'PLANJSON'
{"run_id":"badgate","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"Serve the document","goal":"g","files":[],"depends_on":[],
  "acceptance":["the document describes url"],
  "verify":"node -e \"const d=require('./doc.json');const op=d.paths['/links'].post;if(JSON.stringify(op.requestBody).indexOf('url')<0)process.exit(1)\"",
  "status":"pending","attempts":0,"notes":""}]}
PLANJSON
    ;;
esac
STUB
fixture_run docs/briefs/0003-runstat-cli.md

assert_exit 1
assert_log "gate shape rejected"
assert_log "T1"
# Refused before a single iteration is spent, like every other plan-validation
# failure: the plan phase is cheap to redo, a run built on a bad gate is not.
assert_iterations 0
assert_no_tool_errors
finish
