#!/usr/bin/env bash
# Parallel loops are supported one per git WORKTREE, and git already refuses to
# check one branch out twice — so different worktrees are necessarily different
# branches. What is left to prevent is two loops in the SAME working tree,
# where they share .loop/state/state.json and .loop/tmp/proposal.json. That is not a merge
# annoyance: one loop's review session reading the other's proposal is the
# stale-handoff failure scenario 17 exists to prevent, happening across runs
# where nothing clears it.
#
# A lock that can brick the loop is worse than no lock, so a dead pid must be
# cleared rather than obeyed.
. "$(dirname "$0")/../lib.sh"
fixture_new
fixture_plan "$PLAN_TWO"
fixture_stub_default

note "── a live lock must stop a second loop ──"
mkdir -p "$FX/repo/.loop/tmp"
jq -nc --arg p "$$" '{pid:$p, branch:"other-branch", started:"2026-08-18T00:00:00Z", run:"x"}' \
  >"$FX/repo/.loop/tmp/.running"        # $$ is this test process — genuinely alive
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 1
assert_log "a loop is already running in this working tree"
assert_log "git worktree add"
assert_no_state
[[ -f "$FX/repo/.loop/tmp/.running" ]] && ok "the live lock was left alone" \
  || bad "a live lock was deleted by the run it blocked"

note "── a dead pid must be cleared, not obeyed ──"
dead=999999; kill -0 "$dead" 2>/dev/null && dead=999998
jq -nc --arg p "$dead" '{pid:$p, branch:"crashed-run", started:"2026-08-18T00:00:00Z", run:"x"}' \
  >"$FX/repo/.loop/tmp/.running"
fixture_run docs/briefs/0003-runstat-cli.md
assert_exit 0
assert_log "clearing a stale lock"
ok "a crashed run does not brick the loop"

note "── the lock is released on exit ──"
[[ ! -f "$FX/repo/.loop/tmp/.running" ]] && ok "lock cleared after a normal run" \
  || bad "lock survived a completed run — the next run would see it"

note "── and it is never committed ──"
( cd "$FX/repo" && git check-ignore -q .loop/tmp/.running ) \
  && ok ".loop/tmp/.running is gitignored" \
  || bad "a machine-specific pid file is committable"
finish
