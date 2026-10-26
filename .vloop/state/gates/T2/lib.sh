# Shared harness for B20's run gates (T2's, and T6's worked example). The
# sourcing gate sets H, this folder. Every repository, binary, stub and HOME a
# gate makes lives under $t, a temporary directory removed on exit. Nothing
# touches this repository's tree or refs, the network or the real HOME: runs
# happen in a fixture repository with the stub claude (stub-claude.sh) first on
# PATH. No gate runs the real claude.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT GOFLAGS GOBIN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
BRIEF=docs/briefs/B20260101-0900-demo.loop-brief.md
RUNID=B20260101-0900-demo
# The line the brief pins for a check-gap halt, with the run folder written
# <run folder>; gapline <round> is the line for that round's report.
gapline() {
  echo "vloop: the gate review found 1 check gap(s): the plan needs a check .vloop/config.toml does not have — add it, commit it, then re-run (see <run folder>reports/gate-review-$1.json)"
}

# newbin: $t/bin/new, vloop built from this tree.
newbin() {
  mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
}
# venv cmd...: cmd in the one environment every fixture invocation gets.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
# fixture: $t/repo afresh — git on main, `vloop init` by this tree's binary,
# one check, the harness brief, all committed — with a fresh stub claude (the
# plan it hands out is plan.json) and a fake home trusting the repository.
# $FIX is the commit the fixture ends at.
fixture() {
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
  pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
  cp "$H/plan.json" "$t/stub/plan.json" && ln -s "$GITBIN" "$t/stub/git" &&
    sed "s/@PLUGIN_VERSION@/$pv/" "$H/stub-claude.sh" >"$t/stub/claude" && chmod +x "$t/stub/claude" || fail "stub"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  venv git -C "$t/repo" init -q -b main && venv git -C "$t/repo" config user.name Gate &&
    venv git -C "$t/repo" config user.email gate@example.invalid && venv git -C "$t/repo" config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  (cd "$t/repo" && venv "$t/bin/new" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  cp "$H/brief.md" "$t/repo/$BRIEF" || fail brief
  venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "gate fixture" || fail "fixture commit"
  FIX=$(venv git -C "$t/repo" rev-parse HEAD) || fail "rev-parse"
}
# verdicts <file>...: the gate-review session of round n writes the n-th file
# of the gate folder; a round past the list writes a gate-verdict/v2 PASS.
verdicts() {
  n=0
  for f in "$@"; do n=$((n + 1)); cp "$H/$f" "$t/stub/verdict.$n.json" || fail "verdict $f"; done
}
# runfix <out>: this tree's `vloop run` of the brief in the fixture; $rc is
# its exit code, <out> its stdout and stderr, <out>.err its stderr alone,
# <out>.log the stub's log.
runfix() {
  o=$1
  (cd "$t/repo" && venv "$t/bin/new" run "$BRIEF" >"$o.out" 2>"$o.err"); rc=$?
  cat "$o.out" "$o.err" >"$o"
  if [ -f "$t/stub/log" ]; then mv "$t/stub/log" "$o.log"; else : >"$o.log"; fi
  return 0
}
# sessions <log>: the -p prompts of the stub's sessions, in order.
sessions() {
  awk '/^=== / { prev = ""; next } prev == "arg -p" && /^arg / { print substr($0, 5) } { prev = $0 }' "$1"
}
# trimmed <file>: the file's lines without leading blanks.
trimmed() { sed 's/^[[:space:]]*//' "$1"; }
# runline <file> <prefix>: the first line of the file starting with the prefix,
# with the run folder (.vloop/state/runs/<run id>/<stamp>/) written <run folder>.
runline() {
  trimmed "$1" | grep -F -- "$2" | head -1 |
    sed -E "s#\\.vloop/state/runs/$RUNID/[0-9]{8}-[0-9]{6}(-[0-9]+)?/#<run folder>#"
}
# runpath <file> <prefix>: the run-folder path the first such line names.
runpath() {
  trimmed "$1" | grep -F -- "$2" | head -1 | sed -E -n "s#.*(\\.vloop/state/runs/$RUNID/[0-9]{8}-[0-9]{6}(-[0-9]+)?/).*#\\1#p"
}
# untouched <what>: the fixture ends as it began — no commit on any branch, no
# plan, no gate folders, no plan.md.
untouched() {
  for b in $(venv git -C "$t/repo" for-each-ref --format='%(refname)' refs/heads); do
    [ "$(venv git -C "$t/repo" rev-parse "$b")" = "$FIX" ] || fail "$1: $b moved — a commit was made: $(venv git -C "$t/repo" log --oneline -3 "$b")"
  done
  [ ! -e "$t/repo/.vloop/state/state.json" ] || fail "$1: .vloop/state/state.json was left behind ($("$JQ" -c '{status, gate_review}' "$t/repo/.vloop/state/state.json"))"
  [ ! -e "$t/repo/.vloop/state/gates" ] || fail "$1: .vloop/state/gates/ was left behind"
  [ ! -e "$t/repo/.vloop/state/plan.md" ] || fail "$1: .vloop/state/plan.md was left behind"
}
