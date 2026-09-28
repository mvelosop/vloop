#!/usr/bin/env bash
# The driver picks the next READY task — first pending whose dependencies are
# all done — not simply the first pending one. Plan lists T1 before T2 while
# T1 depends on T2, so file order and ready order disagree.
. "$(dirname "$0")/../lib.sh"
fixture_new
PLAN_REV='{"run_id":"fixture","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"Depends on T2","goal":"g","files":[],"depends_on":["T2"],
  "acceptance":["T1.out"],"verify":"test -f T1.out","status":"pending","attempts":0,"notes":""},
 {"id":"T2","title":"No deps","goal":"g","files":[],"depends_on":[],
  "acceptance":["T2.out"],"verify":"test -f T2.out","status":"pending","attempts":0,"notes":""}]}'
fixture_plan "$PLAN_REV"
fixture_stub_default
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_iter_task 1 T2
assert_iter_task 2 T1
finish
