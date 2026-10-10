#!/bin/sh
# Gate T10 — the domain and the guides follow the brief ("The domain, updated"),
# and docs/guide/commands.md is regenerated.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
# Go keeps its caches under HOME unless told; keep the real ones across the change.
GOMODCACHE=$(go env GOMODCACHE) GOCACHE=$(go env GOCACHE) GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
mkdir -p "$t/Users/gateuser" || fail mkdir
export HOME="$t/Users/gateuser"
N=B20260101-0900-demo
BRIEF=docs/briefs/$N.loop-brief.md
trust() { printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$R" >"$HOME/.claude.json"; }
# runrepo <name> [<plan>]: a vloop repository on main with a ready brief and a
# check (after $CONFIG, when set), committed; its own stub claude playing
# <plan> (plan.json by default) from $t/<name>.stub.
runrepo() {
  R="$t/$1" S="$t/$1.stub"
  mkdir -p "$R" "$S" && git -C "$R" init -q -b main || fail "git init $1"
  cp "$G/claude" "$S/claude" && chmod +x "$S/claude" && cp "$G/${2:-plan.json}" "$S/plan.json" || fail "stub for $1"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail "identity"
  trust
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
  printf '%s[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' "${CONFIG:-}" >"$R/.vloop/config.toml" || fail config
  mkdir -p "$R/docs/briefs" && echo x >"$R/docs/x.md" && cp "$G/brief.md" "$R/$BRIEF" || fail brief
  git -C "$R" add -A && git -C "$R" commit -q -m harness || fail "commit $1"
}
# vrun <out> [VAR=value...]: vloop run <brief> in $R with its stub on PATH;
# stdout in $t/<out>, stderr in $t/<out>.err; returns vloop's exit code.
vrun() {
  o=$1; shift
  trust
  (cd "$R" && env PATH="$S:$PATH" "$@" "$B" run "$BRIEF") >"$t/$o" 2>"$t/$o.err"
}
# records <glob>: the session records of the brief's run folders matching <glob>, one path per line.
records() { ls "$R"/.vloop/state/runs/"$N"/*/sessions/$1 2>/dev/null; }

# The docs say what this brief changed. Each check names the document and what
# it must now state; how it is worded is the review's to judge.
# has <file> <label> <pattern...>: the file matches every extended pattern, case-insensitively.
has() {
  f=$1 label=$2; shift 2
  [ -f "$f" ] || fail "$f is missing"
  for p in "$@"; do grep -Eiq -- "$p" "$f" || fail "$f does not state $label (no match for '$p')"; done
}
has docs/domain/domain-model.md "S-2's ignored .git/config keys" 'run\.git-config-ignore'
has docs/domain/execution/plan.md "that acceptance resumes after a halt, and the stamp on load" 'resum[a-z]* (the )?acceptance' 'stamp'
has docs/domain/execution/run.md "the resume of acceptance after exit 9" 'resum[a-z]* (the )?acceptance'
awk '/^\| *9 *\|/' docs/domain/execution/run.md | grep -iq 'accept' || fail "docs/domain/execution/run.md's exit table row for 9 does not say what happens during acceptance"
has docs/domain/execution/session.md "the record's ended, wall_ms and a denial's command" '`ended`' '`wall_ms`' '`command`'
has docs/domain/measurement/metrics.md "session time from wall_ms, the wall from the first plan session and n/a" 'wall_ms' 'plan session' 'n/a'
has docs/guide/configuration.md "the new key and its environment name" 'run\.git-config-ignore' 'VLOOP_RUN_GIT_CONFIG_IGNORE'
awk '/\*\*wall\*\*/ { on = 1 } on && /^$/ { exit } on' docs/guide/metrics.md | grep -iq 'plan session' \
  || fail "docs/guide/metrics.md's wall entry does not start the wall at the plan session"
has docs/guide/metrics.md "the n/a for unmeasured size" 'match no stack'
has docs/guide/concepts.md "resuming acceptance, the per-session line and status's running line" 'resum[a-z]* (the )?acceptance' 'running:' 'gate review ·|work T[0-9]+ ·'
# The command reference is regenerated.
go build -o "$t/gendocs" ./cmd/gendocs || fail "go build ./cmd/gendocs"
(cd "$t" && "$t/gendocs" "$t/commands.md") >"$t/gd.out" 2>&1 || fail "gendocs: $(cat "$t/gd.out")"
cmp -s "$t/commands.md" docs/guide/commands.md || fail "docs/guide/commands.md is stale: $(diff "$t/commands.md" docs/guide/commands.md | head -6)"
echo "gate T10: ok"
