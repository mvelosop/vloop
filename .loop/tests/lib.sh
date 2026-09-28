#!/usr/bin/env bash
#
# Fixture harness for the loop driver.
#
# Each scenario builds a throwaway repo with a scripted `claude` on PATH and a
# fake HOME, runs .loop/run.sh against it, and asserts on what came out. Nothing
# here talks to a real model, so the whole suite is free and deterministic.
#
# The point is that the driver's mechanics — task selection, the gate, attempts,
# stop conditions, masking — are checkable with a planted input and an expected
# exit code, rather than re-readable prose.
#
#   . "$(dirname "$0")/../lib.sh"
#   fixture_new
#   fixture_plan '<state.json>'
#   fixture_stub <<'EOF' ... EOF
#   fixture_run
#   assert_exit 0
#   finish

set -uo pipefail

SUITE_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SUITE_ROOT/../.." && pwd)"

FAILS=0
NAME="$(basename "${0%.sh}")"

ok()   { printf '    \033[32mok\033[0m   %s\n' "$*"; }
bad()  { printf '    \033[31mFAIL\033[0m %s\n' "$*"; FAILS=$((FAILS + 1)); }
note() { printf '    ---- %s\n' "$*"; }

# ------------------------------------------------------------- fixture ------

fixture_new() {
  FX="$(mktemp -d "${TMPDIR:-/tmp}/loopfx.XXXXXX")"
  FX="$(cd "$FX" && pwd -P)"
  mkdir -p "$FX/repo" "$FX/testuser" "$FX/bin"
  FX_REPO="$(cd "$FX/repo" && pwd -P)"

  # Preflight probes for tools it never runs: `uv`, for the demo target, and
  # `claude` in the scenarios that stop before any session. Whether the
  # developer happens to have those installed is not something a scenario
  # should be able to see -- 31-preflight-only asserts "a ready repo exits 0"
  # and failed on every machine without uv, which is the suite reporting on its
  # host rather than on the driver. Plant inert stubs instead; fixture_stub
  # overwrites the claude one in every scenario that actually runs a session.
  #
  # Safe to shadow because nothing here executes them -- fixture verify commands
  # are `test -f`, not builds. A scenario that ever needs the real tool should
  # delete the stub, not reorder PATH.
  for t in uv claude; do
    printf '#!/usr/bin/env bash\nexit 0\n' >"$FX/bin/$t"
    chmod +x "$FX/bin/$t"
  done

  cp -R "$REPO_ROOT/.claude" "$FX/repo/"
  # The skills are mechanism and must travel. The knowledge declaration is THIS
  # repo's content and must not: carried into a fixture it names roots the
  # fixture does not have, so every scenario ran against a declaration it never
  # asked for and preflight warned about roots nobody declared. Each scenario
  # writes its own, or none.
  rm -f "$FX/repo/.claude/loop-knowledge.md"
  cp "$REPO_ROOT/CLAUDE.md" "$REPO_ROOT/.gitignore" "$FX/repo/"
  mkdir -p "$FX/repo/.loop" "$FX/repo/docs/briefs"
  # settings.json travels with the driver: run.sh hands it to every session with
  # --settings and preflight fails without it, so a fixture missing it would be
  # testing a loop that never runs under its own fence.
  cp "$REPO_ROOT/.loop/run.sh" "$REPO_ROOT/.loop/render-plan.sh" \
     "$REPO_ROOT/.loop/settings.json" "$FX/repo/.loop/"
  echo "# fixture brief" >"$FX/repo/docs/briefs/0003-runstat-cli.md"

  # Preflight's trust check is real; give it a real file that says yes, rather
  # than adding a bypass to the driver that could be left switched on.
  jq -n --arg p "$FX_REPO" '{projects: {($p): {hasTrustDialogAccepted: true}}}' \
    >"$FX/testuser/.claude.json"

  git -C "$FX/repo" init -q
  git -C "$FX/repo" config user.email fixture@test
  git -C "$FX/repo" config user.name fixture
  git -C "$FX/repo" add -A
  git -C "$FX/repo" commit -qm "fixture init"
}

fixture_cleanup() { [[ -n "${FX:-}" && "$FX" == */loopfx.* ]] && rm -rf "$FX"; }
trap fixture_cleanup EXIT

# The plan the scripted planner will emit.
fixture_plan() { FX_PLAN="$1"; }

# Build the scripted `claude`. Stdin is the body; it runs with $PHASE
# (plan|work|review), $TASK, and $FX_PLAN available, and may set $STUB_COST.
fixture_stub() {
  {
    cat <<'PROLOGUE'
#!/usr/bin/env bash
prompt=""
while [[ $# -gt 0 ]]; do case "$1" in -p) prompt="$2"; shift 2;; *) shift;; esac; done
PHASE="${prompt%% *}"; PHASE="${PHASE#/loop-}"
TASK="${prompt##* }"
STUB_COST=0.10
PROLOGUE
    cat
    cat <<'EPILOGUE'
jq -nc --argjson c "$STUB_COST" \
  '{type:"result",subtype:"success",is_error:false,duration_ms:1000,
    num_turns:3,total_cost_usd:$c,session_id:"stub",permission_denials:[]}'
EPILOGUE
  } >"$FX/bin/claude"
  chmod +x "$FX/bin/claude"
}

# The default stub: plans, does every task successfully, passes every review.
# Scenarios override one arm of it to plant the failure they are testing.
fixture_stub_default() {
  fixture_stub <<STUB
case "\$PHASE" in
  plan)   cat > .loop/state/state.json <<'PLANJSON'
$FX_PLAN
PLANJSON
    ;;
  work)   touch "\$TASK.out"
    jq -nc --arg t "\$TASK" '{task:\$t,outcome:"done",summary:("made "+\$t),files:[],verified:"ok",notes:"none"}' > .loop/tmp/proposal.json ;;
  review) jq -nc --arg t "\$TASK" '{task:\$t,verdict:"PASS",criteria:[],findings:[],notes:"none"}' > .loop/tmp/verdict.json ;;
esac
STUB
}

fixture_run() {
  ( cd "$FX/repo" && HOME="$FX/testuser" PATH="$FX/bin:$PATH" \
      bash .loop/run.sh "$@" >"$FX/out.log" 2>&1 )
  EXIT=$?
  return 0
}

fx_state() { jq -r "$@" "$FX/repo/.loop/state/state.json"; }
fx_iters() { cat "$FX/repo"/.loop/state/runs/*/*/iterations.jsonl 2>/dev/null; }
fx_log()   { cat "$FX/out.log"; }
# One journal per plan, so the harness resolves it from state rather than
# assuming a fixed path.
fx_journal() { echo "$FX/repo/.loop/state/journals/$(fx_state .run_id).md"; }
fx_dump()  { note "last 25 log lines:"; tail -25 "$FX/out.log" | sed 's/^/      /'; }

# ---------------------------------------------------------- assertions ------

assert_exit() {
  if [[ "$EXIT" == "$1" ]]; then ok "exit $1"
  else bad "exit: got $EXIT, want $1"; fx_dump; fi
}

assert_status() {
  local got; got="$(fx_state --arg id "$1" '.tasks[]|select(.id==$id)|.status')"
  [[ "$got" == "$2" ]] && ok "$1 status=$2" || bad "$1 status: got $got, want $2"
}

assert_attempts() {
  local got; got="$(fx_state --arg id "$1" '.tasks[]|select(.id==$id)|.attempts')"
  [[ "$got" == "$2" ]] && ok "$1 attempts=$2" || bad "$1 attempts: got $got, want $2"
}

assert_run_status() {
  local got; got="$(fx_state '.status')"
  [[ "$got" == "$1" ]] && ok "run status=$1" || bad "run status: got $got, want $1"
}

assert_iterations() {
  local got; got="$(fx_iters | wc -l | tr -d ' ')"
  [[ "$got" == "$1" ]] && ok "$1 iteration(s)" || bad "iterations: got $got, want $1"
}

# The task worked in iteration N (1-indexed).
assert_iter_task() {
  local got; got="$(fx_iters | sed -n "${1}p" | jq -r '.task')"
  [[ "$got" == "$2" ]] && ok "iteration $1 worked $2" || bad "iteration $1: got $got, want $2"
}

assert_iter_outcome() {
  local got; got="$(fx_iters | sed -n "${1}p" | jq -r '.outcome')"
  [[ "$got" == "$2" ]] && ok "iteration $1 outcome=$2" || bad "iteration $1 outcome: got $got, want $2"
}

assert_log()    { grep -q -- "$1" "$FX/out.log" && ok "log: $1" || { bad "log missing: $1"; fx_dump; }; }
assert_no_log() { grep -q -- "$1" "$FX/out.log" && { bad "log should not contain: $1"; fx_dump; } || ok "log clean of: $1"; }

assert_no_state() {
  [[ -f "$FX/repo/.loop/state/state.json" ]] && bad "state.json should not exist" || ok "no state.json written"
}

# Hard rule 1 + 2: nothing outside the repo, and no tracked file names the
# machine. Scoped to *tracked* files deliberately — .loop/tmp/proposal.json and
# .loop/tmp/verdict.json are raw session output, masked only on the way into
# .loop/state/runs/, so the invariant that matters is that they are never committed.
assert_contained() {
  local hits
  hits="$(cd "$FX/repo" && git ls-files -z | xargs -0 grep -l "testuser" 2>/dev/null || true)"
  [[ -z "$hits" ]] && ok "no username in any tracked file" \
    || bad "username leaked into tracked: $hits"

  local leaked=()
  for f in .loop/tmp/proposal.json .loop/tmp/verdict.json; do
    ( cd "$FX/repo" && git check-ignore -q "$f" ) || leaked+=("$f")
  done
  [[ ${#leaked[@]} -eq 0 ]] && ok "transient handoffs are gitignored" \
    || bad "unmasked session output is committable: ${leaked[*]}"

  # Not just the username: ANY absolute path in a tracked file names the
  # machine. The mask keys on $HOME, so a path that does not sit under it —
  # the repo's own location, say — slips straight through. That is exactly how
  # an absolute journal path reached a committed loop.log.
  hits="$(cd "$FX/repo" && git ls-files -z | xargs -0 grep -l -- "$FX" 2>/dev/null || true)"
  [[ -z "$hits" ]] && ok "no absolute sandbox path in any tracked file" \
    || bad "absolute path leaked into tracked: $hits"

  hits="$(find "$FX/testuser" -mindepth 1 -not -name '.claude.json' 2>/dev/null || true)"
  [[ -z "$hits" ]] && ok "nothing written to fake HOME" \
    || bad "wrote outside the repo: $hits"
}

finish() {
  if [[ $FAILS -eq 0 ]]; then printf '  \033[32mPASS\033[0m %s\n' "$NAME"; exit 0
  else printf '  \033[31mFAIL\033[0m %s (%d assertion(s))\n' "$NAME" "$FAILS"; exit 1; fi
}

# ----------------------------------------------------------- plan bodies ----

# Two independent tasks, each satisfied by the default stub's `touch $TASK.out`.
PLAN_TWO='{"run_id":"fixture","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"First","goal":"g","files":[],"depends_on":[],
  "acceptance":["T1.out exists"],"verify":"test -f T1.out","status":"pending","attempts":0,"notes":""},
 {"id":"T2","title":"Second","goal":"g","files":[],"depends_on":["T1"],
  "acceptance":["T2.out exists"],"verify":"test -f T2.out","status":"pending","attempts":0,"notes":""}]}'

PLAN_ONE='{"run_id":"fixture","brief":"docs/briefs/0003-runstat-cli.md","status":"running","iteration":0,
 "created":"2026-08-15T00:00:00Z","updated":"2026-08-15T00:00:00Z","tasks":[
 {"id":"T1","title":"Only","goal":"g","files":[],"depends_on":[],
  "acceptance":["T1.out exists"],"verify":"test -f T1.out","status":"pending","attempts":0,"notes":""}]}'

# Preflight's trust check, from the other side: a workspace the operator has
# not accepted. This is the failure that silently voided all permission rules
# in a prior experiment.
fixture_untrust() {
  jq -n --arg p "$FX_REPO" '{projects: {($p): {hasTrustDialogAccepted: false}}}' \
    >"$FX/testuser/.claude.json"
}

# A tool erroring mid-run is never harmless: awk exiting 2 on a malformed value
# makes an `if awk ...` budget check silently false, so the budget stops being
# enforced while everything still looks fine.
assert_no_tool_errors() {
  local hits
  hits="$(grep -nE '^(awk|jq|sed|grep|bash): ' "$FX/out.log" | head -3 || true)"
  [[ -z "$hits" ]] && ok "no tool errors leaked into the run" \
    || bad "tool error in run output: $hits"
}
