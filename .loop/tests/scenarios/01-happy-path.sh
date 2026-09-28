#!/usr/bin/env bash
# The baseline: plan, do every task, pass every review, stop clean.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_run_status complete
assert_status T1 done
assert_status T2 done
assert_iterations 2
assert_iter_task 1 T1
assert_iter_task 2 T2
assert_log "plan complete"
finish
