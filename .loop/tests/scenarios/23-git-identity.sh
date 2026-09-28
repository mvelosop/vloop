#!/usr/bin/env bash
# The driver makes one commit per iteration, at the END of the iteration. A repo
# with no git identity configured therefore fails after both the work and the
# review session have been paid for, which is the most expensive possible place
# to discover a one-line setup problem. Preflight is free.
#
# This matters most in a repo the loop was just installed into, which is exactly
# where identity is least likely to be set.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default

note "── no identity configured: refuse before spending ──"
( cd "$FX/repo" && git config --unset user.email && git config --unset user.name )
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 1
assert_log "git user.name/user.email not set"
assert_log "preflight failed"
assert_no_state
ok "nothing was spent"

note "── identity restored: proceeds ──"
( cd "$FX/repo" && git config user.email fixture@test && git config user.name fixture )
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_log "git identity (fixture@test)"

note "── a pre-commit hook is surfaced, not blocked on ──"
mkdir -p "$FX/repo/.git/hooks"
printf '#!/bin/sh\nexit 0\n' > "$FX/repo/.git/hooks/pre-commit"
chmod +x "$FX/repo/.git/hooks/pre-commit"
fixture_run
assert_exit 0
assert_log "a pre-commit hook is active"
ok "a passing hook does not stop the run"
finish
