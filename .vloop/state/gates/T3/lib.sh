# Shared harness: a fresh build, a temporary HOME, a stub claude answering
# --version and `plugin list --json` from $t/plugins.json, and a PATH that holds
# only the system's tools (and, when asked, the stub).
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) GOCACHE=$(go env GOCACHE) GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
V=$("$B" version --json | "$JQ" -r .version) && [ -n "$V" ] && [ "$V" != null ] || fail "vloop version --json gave no version"
SYS=/usr/bin:/bin
PATH="$SYS" command -v claude >/dev/null 2>&1 && fail "fixture: claude is on $SYS, so 'no claude' cannot be staged"
mkdir -p "$t/stub" "$t/Users/gateuser" || fail mkdir
cat >"$t/stub/claude" <<'STUB'
#!/bin/sh
case "$1" in
--version) echo "2.0.0 (Claude Code)" ;;
plugin) [ "$2" = list ] && cat "$STUB_PLUGINS" && exit 0; exit 1 ;;
*) exit 1 ;;
esac
STUB
chmod +x "$t/stub/claude" || fail chmod
export HOME="$t/Users/gateuser"
# plugins <json>: what the stub's `claude plugin list --json` prints.
plugins() { printf '%s\n' "$1" >"$t/plugins.json"; }
plugins '[]'
# newrepo <name>: a git repository where vloop init ran, with a [[check]] and
# trusted in $HOME/.claude.json, so doctor's only warnings are the ones staged.
newrepo() {
  R="$t/$1"
  mkdir -p "$R" && git -C "$R" init -q -b main || fail "git init $1"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail identity
  (cd "$R" && env PATH="$t/stub:$SYS" "$B" init </dev/null) >"$t/$1.init" 2>&1 || fail "vloop init in $1: $(cat "$t/$1.init")"
  printf '\n[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >>"$R/.vloop/config.toml" || fail config
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$R" >"$HOME/.claude.json" || fail trust
}
# doc <out> [args...]: vloop doctor in $R with the stub on PATH; stdout in
# $t/<out>; sets got to its exit code. docnc: the same with no claude on PATH.
doc() { o=$1; shift; (cd "$R" && env PATH="$t/stub:$SYS" STUB_PLUGINS="$t/plugins.json" "$B" doctor "$@") >"$t/$o" 2>"$t/$o.err"; got=$?; }
docnc() { o=$1; shift; (cd "$R" && env PATH="$SYS" "$B" doctor "$@") >"$t/$o" 2>"$t/$o.err"; got=$?; }
# has <out> <line>: <out> has exactly that line.
has() { grep -Fxq -- "$2" "$t/$1" || fail "$1: no line '$2' in:
$(cat "$t/$1")"; }
PLUGIN_W='! plugin /vloop:operate is not available in your interactive sessions — claude plugin install vloop@vloop (after claude plugin marketplace add mvelosop/vloop), or claude --plugin-dir "$(vloop plugin path)" for one session'
GATE_W='! gate edits no allow rule lets an interactive session edit .vloop/state/gates/ — add "Edit(./.vloop/state/gates/**)" to permissions.allow in .claude/settings.json'
