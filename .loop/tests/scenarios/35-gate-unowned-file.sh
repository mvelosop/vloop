#!/usr/bin/env bash
# Gate-shape lint, rule 3: a gate may not inspect the bytes of a HEAD-tracked
# file the task does not own. Not a heuristic -- the condition is
# gate_files_moved() (.loop/run.sh) evaluated one phase early. That function
# reverts any HEAD-existing file the current task's verify names and its files
# does not, and the revert runs before the gate does, so a task in that shape
# is unpassable by construction, for any implementation, forever
# (33-gate-rewrite.sh exercises the runtime half of the same guard).
#
# One rejected shape and two accepted escape hatches: own the file, or only
# hand it to a runner rather than reading it.
. "$(dirname "$0")/../lib.sh"

plan_one() {   # <files json> <verify>
  fixture_plan "{\"run_id\":\"fixture\",\"brief\":\"docs/briefs/0003-runstat-cli.md\",\"status\":\"running\",\"iteration\":0,
 \"created\":\"2026-08-15T00:00:00Z\",\"updated\":\"2026-08-15T00:00:00Z\",\"tasks\":[
 {\"id\":\"T1\",\"title\":\"t\",\"goal\":\"g\",\"files\":$1,\"depends_on\":[],
  \"acceptance\":[\"a\"],\"verify\":\"$2\",\"status\":\"pending\",\"attempts\":0,\"notes\":\"\"}]}"
}

new_fixture_with_todo() {
  fixture_new
  mkdir -p "$FX_REPO/docs" "$FX_REPO/tests"
  echo TIMEOUT=5 >"$FX_REPO/docs/todo.md"
  echo 'def test_x(): pass' >"$FX_REPO/tests/test_thing.py"
  git -C "$FX_REPO" add -A
  git -C "$FX_REPO" commit -qm "pre-existing files"
}

note "-- rejected: inspects docs/todo.md, which the task does not own --"
new_fixture_with_todo
plan_one '["lib/thing.py"]' 'grep -q TIMEOUT docs/todo.md && uv run pytest -q tests/test_thing.py'
fixture_stub_default
fixture_run --plan-only docs/briefs/0003-runstat-cli.md
assert_exit 1
assert_log "gate shape rejected"
assert_log 'T1 .*docs/todo.md'
assert_iterations 0

note "-- accepted: the same inspecting verify, with the path owned --"
new_fixture_with_todo
plan_one '["lib/thing.py","docs/todo.md"]' 'grep -q TIMEOUT docs/todo.md && uv run pytest -q tests/test_thing.py'
fixture_stub_default
fixture_run --plan-only docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_no_log "gate shape rejected"

note "-- accepted: the path is handed to a runner, not read --"
new_fixture_with_todo
plan_one '["lib/thing.py"]' 'uv run pytest -q tests/test_thing.py'
fixture_stub_default
fixture_run --plan-only docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_no_log "gate shape rejected"

assert_no_tool_errors
finish
