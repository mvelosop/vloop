#!/usr/bin/env bash
# --check answers "is this repo ready?" without spending anything.
#
# The checks it runs already existed — workspace trust, the fence, the
# knowledge roots, index coverage — but only inside a run, which meant the only
# way to ask "have I set this up right?" was to start one. That is the wrong
# price for a question you ask again every time you edit a declaration or add a
# document.
#
# It runs the same preflight the real thing runs, not a reimplementation: a
# sanity check that can disagree with the thing it checks is worse than none.
. "$(dirname "$0")/../lib.sh"

# $FX/bin first, like fixture_run: the tools preflight only probes for are
# stubbed there, so "is this repo ready?" answers about the fixture and not
# about what the developer happens to have installed.
check_run() { ( cd "$FX/repo" && HOME="$FX/testuser" PATH="$FX/bin:$PATH" \
                  bash .loop/run.sh "$@" >"$FX/out.log" 2>&1 ); }

note "── a ready repo: exit 0, and nothing is created ──"
fixture_new
check_run --check
rc=$?
[[ $rc -eq 0 ]] && ok "exit 0" || bad "exit: got $rc, want 0"
assert_log "preflight ok"
assert_log "nothing was run and nothing was changed"
# No brief was named and none is needed — the question is about the repo.
[[ ! -f "$FX/repo/.loop/state/state.json" ]] && ok "no state was written" || bad "a check wrote state"
[[ -z "$(ls -A "$FX/repo/.loop/state/runs" 2>/dev/null)" ]] \
  && ok "no run directory was left behind" \
  || bad "a run directory was created — a check that litters is a check people stop running"
[[ ! -e "$FX/repo/.loop/tmp/.running" ]] && ok "no lock was taken" || bad "a lock was left behind"

note "── --preflight is the same thing ──"
check_run --preflight
[[ $? -eq 0 ]] && ok "exit 0" || bad "--preflight did not behave like --check"
assert_log "preflight ok"

note "── a check must not block, or be blocked by, a running loop ──"
mkdir -p "$FX/repo/.loop/tmp"
jq -nc --arg p "$$" '{pid:$p, branch:"other", started:"2026-08-18T00:00:00Z", run:"x"}' \
  >"$FX/repo/.loop/tmp/.running"
check_run --check
[[ $? -eq 0 ]] && ok "runs while a loop holds the lock" \
  || bad "a check refused because a loop was running"
[[ -f "$FX/repo/.loop/tmp/.running" ]] && ok "the live lock was left alone" || bad "a check cleared someone's lock"
rm -f "$FX/repo/.loop/tmp/.running"

note "── advisories exit non-zero, so it can gate a hook ──"
mkdir -p "$FX/repo/docs/domain"
printf '# transaction\n' >"$FX/repo/docs/domain/transaction.md"
cat >"$FX/repo/.claude/loop-knowledge.md" <<'KNOW'
# Knowledge roots
| Surface | Path | How to find what is in it |
|---|---|---|
| Domain model | `docs/domain/` | nothing yet |
KNOW
check_run --check
rc=$?
# A run would still start on these — they are advisories, not blockers. But the
# two commands are asked different questions, and answering "is anything wrong?"
# with a silent 0 makes the flag useless in the hook it exists to be put in.
[[ $rc -eq 1 ]] && ok "exit 1 on advisories" || bad "exit: got $rc, want 1"
assert_log "no index and no descriptions"
assert_log "advisory finding(s)"

note "── and it reports an index that has stopped naming its own folder ──"
fixture_cleanup; fixture_new
mkdir -p "$FX/repo/docs/decisions"
printf '# Decisions\n\n- `0001-storage.md` where uploads live\n' >"$FX/repo/docs/decisions/README-decisions.md"
printf '# storage\n'  >"$FX/repo/docs/decisions/0001-storage.md"
printf '# currency\n' >"$FX/repo/docs/decisions/0002-currency.md"
cat >"$FX/repo/.claude/loop-knowledge.md" <<'KNOW'
# Knowledge roots
| Surface | Path | How to find what is in it |
|---|---|---|
| Decisions | `docs/decisions/` | index |
KNOW
check_run --check
[[ $? -eq 1 ]] && ok "exit 1" || bad "an unlisted document did not fail the check"
assert_log "index does not name"
assert_log "0002-currency.md"

finish
