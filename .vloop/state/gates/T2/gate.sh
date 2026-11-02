#!/bin/sh
# Gate T2 — V1, a finished plan does not block the next brief. This tree's
# binary, built into a temporary directory, in fixture repositories under $t
# with the stub claude (stub-claude.sh) first on PATH, a fake HOME trusting
# the fixture, and nothing of this repository's tree, refs or HOME touched.
#
# The fixture: on work branch B20260101-0900-a, a committed state/v2 plan of
# docs/briefs/B20260101-0900-a.loop-brief.md (run id B20260101-0900-a, branch
# the current one) whose T1 references docs/old.md, which a later commit
# removed; and a ready brief docs/briefs/B20260101-0910-b.loop-brief.md.
# <n> is what `vloop task validate` counts on it.
#   complete plan: `vloop doctor` prints the pass line (and --json carries it,
#     result pass); `vloop run <b>` resets it and plans fresh, exit 0;
#     `vloop run <a>` and a bare `vloop run` refuse in preflight, exit 1, with
#     the plan line and the preflight line, nothing run.
#   planted, status running: `vloop run <b>` refuses the same way, exit 1.
#   `vloop task validate` still counts the complete plan's problems.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p; s/^\(COPILOT_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT GOFLAGS GOBIN GH_TOKEN GITHUB_TOKEN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/vloop" ./cmd/vloop || fail "go build of this tree"
S=$t/bin/vloop
A=docs/briefs/B20260101-0900-a.loop-brief.md
B=docs/briefs/B20260101-0910-b.loop-brief.md
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
g() { venv git -C "$t/repo" "$@"; }
# fixture <plan status> <task status>: $t/repo afresh, as above.
fixture() {
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
  pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
  cp "$G/plan-b.json" "$t/stub/plan.json" && ln -s "$GITBIN" "$t/stub/git" && ln -s "$JQ" "$t/stub/jq" &&
    sed "s/@PLUGIN_VERSION@/$pv/" "$G/stub-claude.sh" >"$t/stub/claude" && chmod +x "$t/stub/claude" || fail "stub"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  g init -q -b main && g config user.name Gate && g config user.email gate@example.invalid && g config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  echo "the old contract" >"$t/repo/docs/old.md"
  echo "the fixture's readme" >"$t/repo/README.md"
  (cd "$t/repo" && venv "$S" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  sed 's/@NAME@/B20260101-0900-a/' "$G/brief.md" >"$t/repo/$A" && sed 's/@NAME@/B20260101-0910-b/' "$G/brief.md" >"$t/repo/$B" || fail brief
  g add -A && g commit -q -m "gate fixture" || fail "fixture commit"
  g switch -q -c B20260101-0900-a || fail "work branch"
  base=$(g rev-parse HEAD) || fail "rev-parse"
  mkdir -p "$t/repo/.vloop/state" && sed -e "s/@BASE@/$base/" -e "s/@STATUS@/$1/" -e "s/@TSTATUS@/$2/" "$G/plan-a.json" >"$t/repo/.vloop/state/state.json" || fail plan
  g add -A && g commit -q -m "[vloop] B20260101-0900-a: the plan" || fail "plan commit"
  g rm -q docs/old.md && g commit -q -m "move docs/old.md away" || fail "remove docs/old.md"
}
# inrepo <name> cmd...: run cmd in the fixture; $rc, $t/<name>.out and .err.
inrepo() { n=$1; shift; (cd "$t/repo" && venv "$S" "$@" >"$t/$n.out" 2>"$t/$n.err"); rc=$?; }
# refused <name>: exit 1 from the preflight, with the plan line and the
# preflight line on stderr, no session run and the plan as it was.
refused() {
  [ "$rc" = 1 ] || fail "$1: exit $rc, want 1: $(tail -5 "$t/$1.err")"
  grep -qxF "  ✗ plan $np problem(s) — run vloop task validate" "$t/$1.err" || fail "$1: no '  ✗ plan $np problem(s) — run vloop task validate' on stderr: $(cat "$t/$1.err")"
  grep -qxF "vloop: preflight failed — 1 problem(s), nothing has run" "$t/$1.err" || fail "$1: no 'vloop: preflight failed — 1 problem(s), nothing has run' on stderr: $(cat "$t/$1.err")"
  [ -s "$t/stub/events" ] && fail "$1: a session ran: $(cat "$t/stub/events")"
  [ "$("$JQ" -r .run_id "$t/repo/.vloop/state/state.json")" = B20260101-0900-a ] || fail "$1: the plan was replaced"
  return 0
}

# The complete plan: what the doctor and task validate say.
fixture complete done
inrepo tv task validate
np=$(sed -n 's/^\([0-9][0-9]*\) problem(s).*/\1/p' "$t/tv.out" "$t/tv.err" | head -1)
[ -n "$np" ] && [ "$np" -ge 1 ] || fail "vloop task validate does not count the complete plan's problems (exit $rc): $(cat "$t/tv.out" "$t/tv.err")"
[ "$rc" != 0 ] || fail "vloop task validate passes the complete plan whose reference is gone"
grep -qF docs/old.md "$t/tv.out" "$t/tv.err" || fail "vloop task validate does not name docs/old.md: $(cat "$t/tv.out" "$t/tv.err")"
WANT="✓ plan a complete plan of $A — the next brief's plan replaces it"
inrepo doc doctor
grep -qxF "$WANT" "$t/doc.out" || fail "vloop doctor: want the line '$WANT', got: $(grep -F plan "$t/doc.out")"
inrepo docj --json doctor
m=$("$JQ" -r '.checks[] | select(.name == "plan") | .result + "|" + .message' "$t/docj.out" 2>/dev/null)
[ "$m" = "pass|a complete plan of $A — the next brief's plan replaces it" ] || fail "vloop doctor --json's plan check: want pass with the message, got '$m'"

# A run naming the plan's own brief, and a bare run, are resumes: validated.
inrepo resume run "$A"; refused resume
inrepo bare run; refused bare

# A run naming another brief: the plan is reset and the run goes on.
inrepo fresh run "$B"
[ "$rc" = 0 ] || fail "vloop run $B on a complete plan of another brief exited $rc, want 0: $(tail -8 "$t/fresh.err")"
grep -qxF "the plan B20260101-0900-a was for $A — you asked for $B: resetting and planning fresh" "$t/fresh.out" ||
  fail "vloop run $B: no reset line on stdout: $(head -8 "$t/fresh.out")"
grep -qF '✗ plan' "$t/fresh.err" && fail "vloop run $B: the preflight printed the plan check: $(grep -F '✗ plan' "$t/fresh.err")"
st=$("$JQ" -r '.brief + " " + .status' "$t/repo/.vloop/state/state.json")
[ "$st" = "$B complete" ] || fail "vloop run $B: the plan reads '$st', want '$B complete'"
grep -q '^session plan ' "$t/stub/events" || fail "vloop run $B: no plan session ran"

# Planted: the same plan, running: validated, refused.
fixture running pending
inrepo planted run "$B"; refused planted
inrepo pdoc doctor
grep -qF "a complete plan of" "$t/pdoc.out" && fail "vloop doctor passes a running plan as complete: $(grep -F plan "$t/pdoc.out")"
grep -qxF "✗ plan $np problem(s) — run vloop task validate" "$t/pdoc.out" || fail "vloop doctor on a running plan: want '✗ plan $np problem(s) — run vloop task validate', got: $(grep -F plan "$t/pdoc.out")"
echo "gate T2: ok"
