#!/usr/bin/env bash
# A plan-only run (LOOP_MAX_ITERATIONS=0) closes nothing and records nothing.
# Every signal must read zero. jq's `//` binds to the whole path expression, so
# `[.[].tasks_done // 0]` on an empty stream yields ONE phantom 0 rather than
# none — which made the driver report a no-progress streak of 1 on a run that
# had no iterations at all, disagreeing with what runstat's Python computes for
# the same input. Brief 0002 acceptance item 6 does not tolerate that.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default
LOOP_MAX_ITERATIONS=0 fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 4
assert_iterations 0
assert_log "iterations:            0"
assert_log "tasks closed:          0/0"
assert_log "iterations per closed: n/a"
assert_log "no-progress streak:    0"
assert_log "gate failures:         0"
assert_log "attempts burned:       0"
finish
