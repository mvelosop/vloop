#!/usr/bin/env bash
# A session must not move git refs. A planning session checking its own gate
# fixtures once ran their git commands in the real repository: it renamed the
# branch, created and checked out another, and planted `origin/trunk` with
# `origin/HEAD` pointing at it. The driver then committed the whole run onto the
# new branch without noticing.
#
# The driver snapshots every ref, and where HEAD points, before each session and
# compares after. Any difference halts with exit 9 and commits nothing -- not
# the iteration, not the run's closing commit -- because HEAD may now be on
# another branch. The stub moves refs directly, as `git -C` or a test script
# would, so this exercises the driver's check, not the fence.
. "$(dirname "$0")/../lib.sh"

loop_commits() { git -C "$FX/repo" log --all --format=%s | grep -c "$1" || true; }

refs_case() {   # <phase that moves refs> <shell snippet that moves them>
  fixture_cleanup; fixture_new
  export MOVE_PHASE="$1" MOVE_CMD="$2"
  fixture_stub <<STUB
case "\$PHASE" in
  plan)   cat > .loop/state/state.json <<'PLANJSON'
$PLAN_ONE
PLANJSON
    ;;
  work)   touch "\$TASK.out"
    jq -nc --arg t "\$TASK" '{task:\$t,outcome:"done",summary:("made "+\$t),files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
[ "\$PHASE" = "\$MOVE_PHASE" ] && eval "\$MOVE_CMD"
STUB
  BRANCH0="$(git -C "$FX/repo" symbolic-ref --short HEAD)"
  fixture_run docs/briefs/0003-runstat-cli.md
}

note "── the planning session renames the branch ──"
refs_case plan 'git branch -m "$(git symbolic-ref --short HEAD)" renamed-by-session'
assert_exit 9
assert_log "REFS MOVED"
assert_log "refs/heads/renamed-by-session"
[ "$(loop_commits '^\[loop\] plan')" = 0 ] && ok "the plan was not committed" \
  || bad "the plan was committed onto moved refs"

note "── the work session plants a remote ref and repoints origin/HEAD ──"
refs_case work 'git update-ref refs/remotes/origin/trunk HEAD && git symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/trunk'
assert_exit 9
assert_log "REFS MOVED T1"
assert_log "refs/remotes/origin/trunk"
[ "$(loop_commits '^\[loop\] T1')" = 0 ] && ok "the iteration was not committed" \
  || bad "the iteration was committed after refs moved"
[ "$(loop_commits '^\[loop\] run ')" = 0 ] && ok "the run's closing commit was skipped" \
  || bad "the run's closing commit was made after refs moved"

note "── the review session checks out a new branch ──"
refs_case review 'git checkout -q -b work'
assert_exit 9
assert_log "REFS MOVED T1"
assert_log "refs/heads/work"
[ "$(loop_commits '^\[loop\] T1')" = 0 ] && ok "the iteration was not committed" \
  || bad "the iteration was committed onto the new branch"
[ "$(git -C "$FX/repo" rev-list --count work 2>/dev/null)" = "$(git -C "$FX/repo" rev-list --count "$BRANCH0")" ] \
  && ok "nothing landed on the new branch" || bad "commits landed on the branch the session created"

note "── control: a session that leaves refs alone runs to completion ──"
refs_case none 'true'
assert_exit 0
assert_no_log "REFS MOVED"

finish
