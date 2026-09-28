#!/usr/bin/env bash
# An untrusted workspace makes `claude -p` ignore .claude/settings.json without
# saying so, which cost a prior experiment an entire run. Preflight must refuse
# before spending anything — and this one fired for real on 2026-08-15.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default
fixture_untrust
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 1
assert_log "workspace NOT trusted"
assert_log "preflight failed"
assert_no_state
[[ ! -d "$FX/repo/.loop/state/runs" ]] || [[ -z "$(ls -A "$FX/repo"/.loop/state/runs/*/*/sessions 2>/dev/null)" ]] \
  && ok "no session was run" || bad "a session ran despite failed preflight"
finish
