#!/bin/sh
# Gate T3 — R3: the system temp directory masked as $TMPDIR. This tree's
# `vloop run` plans a fixture brief with a stub claude (stub-claude.sh) whose
# plan session was denied Bash commands naming files in the temp directory and
# wrote the directory to stderr. Three runs, each in a fresh fixture:
#   1. TMPDIR is a symbolic link to another directory: the denials, given in
#      both forms, read $TMPDIR/…, and neither form of the directory occurs
#      anywhere in the run folder (session record, stderr file, run log).
#   2. TMPDIR=/tmp: the denial keeps /tmp/x.1 as it is.
#   3. TMPDIR under HOME: the denial reads ~/tmp/x.1 (the home is masked first).
# Nothing touches this repository's tree or refs, the network or the real HOME.
set -u
H=.vloop/state/gates/T3
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
mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"

# venv <tmpdir> cmd...: cmd in the one environment every fixture invocation gets.
venv() {
  vt=$1; shift
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$vt" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
# fixture: $t/repo afresh, git on main, `vloop init` by this tree's binary, one
# check, the brief, committed; a fake home trusting it; the stub claude.
fixture() {
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/stub" || fail mkdir
  ln -s "$GITBIN" "$t/stub/git" && sed "s/@PLUGIN_VERSION@/$pv/" "$H/stub-claude.sh" >"$t/stub/claude" && chmod +x "$t/stub/claude" || fail stub
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  venv /tmp git -C "$t/repo" init -q -b main && venv /tmp git -C "$t/repo" config user.name Gate &&
    venv /tmp git -C "$t/repo" config user.email gate@example.invalid && venv /tmp git -C "$t/repo" config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  (cd "$t/repo" && venv /tmp "$t/bin/new" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  cp "$H/brief.md" "$t/repo/$BRIEF" || fail brief
  venv /tmp git -C "$t/repo" add -A && venv /tmp git -C "$t/repo" commit -q -m "gate fixture" || fail "fixture commit"
}
# plan <tmpdir> <denied path>...: a fresh fixture, then `vloop run` with TMPDIR
# set to <tmpdir>; the plan session is denied `rm -rf <path>` for each path
# and writes every path to stderr. $REC is the plan session's record.
plan() {
  pt=$1; shift
  fixture
  sep="" ; printf '[' >"$t/stub/denials.json"; : >"$t/stub/stderr.txt"
  for p in "$@"; do
    printf '%s{"tool_name":"Bash","tool_use_id":"x","tool_input":{"command":"rm -rf %s","description":"d"}}' "$sep" "$p" >>"$t/stub/denials.json"
    echo "stub: cannot remove $p" >>"$t/stub/stderr.txt"
    sep=,
  done
  printf ']' >>"$t/stub/denials.json"
  (cd "$t/repo" && venv "$pt" "$t/bin/new" run "$BRIEF" >"$t/run.out" 2>&1)
  REC=$(find "$t/repo/.vloop/state/runs" -path '*/sessions/001-plan.json' 2>/dev/null | head -1)
  [ -n "$REC" ] || fail "TMPDIR=$pt: no sessions/001-plan.json in the run folder: $(tail -5 "$t/run.out")"
  RUNF=$(dirname "$(dirname "$REC")")
}
# denied <n>: the n-th denial's command in $REC.
denied() { "$JQ" -r ".permission_denials[$1].command" "$REC"; }

# 1. A temp directory reached through a symbolic link.
mkdir -p "$t/real" && ln -s "$t/real" "$t/tl" || fail "temp dirs"
plan "$t/tl" "$t/tl/x.1" "$t/real/y.2"
d0=$(denied 0) d1=$(denied 1)
[ "$d0" = 'rm -rf $TMPDIR/x.1' ] || fail "a denial naming TMPDIR's path is recorded as '$d0', want 'rm -rf \$TMPDIR/x.1'"
[ "$d1" = 'rm -rf $TMPDIR/y.2' ] || fail "a denial naming TMPDIR's resolved path is recorded as '$d1', want 'rm -rf \$TMPDIR/y.2'"
for form in "$t/tl" "$t/real"; do
  hits=$(grep -rlF -- "$form" "$RUNF" | sed "s#^$t/repo/##")
  [ -z "$hits" ] || fail "the temp directory $form is kept in the run folder: $hits"
done
ls "$RUNF"/*.stderr >/dev/null 2>&1 || fail "the run folder has no stderr file for the plan session"
grep -qF -- '$TMPDIR/x.1' "$RUNF"/*.stderr || fail "the plan session's stderr file does not read \$TMPDIR/x.1"

# 2. /tmp is left as it is.
plan /tmp /tmp/x.1
d0=$(denied 0)
[ "$d0" = 'rm -rf /tmp/x.1' ] || fail "with TMPDIR=/tmp a denial of /tmp/x.1 is recorded as '$d0' — /tmp must stay as it is"

# 3. A temp directory under the home comes out as ~/….
plan "$t/home/tmp" "$t/home/tmp/x.1"
d0=$(denied 0)
[ "$d0" = 'rm -rf ~/tmp/x.1' ] || fail "with TMPDIR under HOME a denial is recorded as '$d0', want 'rm -rf ~/tmp/x.1'"
echo "gate T3: ok"
