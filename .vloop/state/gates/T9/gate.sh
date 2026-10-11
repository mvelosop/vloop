#!/bin/sh
# Gate T9 — metrics say when nothing was measured (brief R7): a brief that
# delivered only shell and markdown lines under metrics.stacks = ["go"] shows
# n/a for code and test in the table and a line naming the unclassified lines
# under the summary's delivered line; the JSON is unchanged; a brief that
# delivered Go code keeps its numbers.
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

# col <table file> <header name> <brief>: the brief's value in the named column of vloop metrics' table.
col() { awk -v h="$2" -v b="$3" 'NR == 1 { for (i = 1; i <= NF; i++) if ($i == h) c = i; next } $1 == b { print $c }' "$1"; }

# --- a brief that delivered only shell and markdown lines, with metrics.stacks = ["go"] ---
CONFIG='[metrics]
stacks = ["go"]
' runrepo o
cat >"$S/script.sh" <<'EOF'
stub_after() {
  [ "$PHASE" = work ] || return 0
  printf '#!/bin/sh\nset -u\necho one\necho two\n\necho three\n' >tool.sh
  printf '# Notes\n\nSome notes.\n' >notes.md
}
EOF
vrun o.out; got=$?
[ "$got" = 0 ] || fail "the stubbed run exited $got, want 0 — $(tail -5 "$t/o.out.err")"
(cd "$R" && "$B" metrics "$N" --json) >"$t/m.json" 2>"$t/err" || fail "metrics --json: $(cat "$t/err")"
jq -e '.size.delivered.code == 0 and .size.delivered.test == 0 and .size.delivered.other > 0' "$t/m.json" >/dev/null \
  || fail "the fixture did not deliver only unclassified lines: $(jq -c .size.delivered "$t/m.json")"
n=$(jq .size.delivered.other "$t/m.json")
(cd "$R" && "$B" metrics) >"$t/table" 2>"$t/err" || fail "metrics: $(cat "$t/err")"
[ "$(col "$t/table" code "$N")" = n/a ] && [ "$(col "$t/table" test "$N")" = n/a ] \
  || fail "the table's code and test for a brief with no code or test lines are '$(col "$t/table" code "$N")' and '$(col "$t/table" test "$N")', want n/a: $(cat "$t/table")"
(cd "$R" && "$B" metrics "$N") >"$t/sum" 2>"$t/err" || fail "metrics $N: $(cat "$t/err")"
d=$(grep -n '^ size      delivered ' "$t/sum" | cut -d: -f1)
[ -n "$d" ] || fail "the summary has no delivered line: $(cat "$t/sum")"
sed -n "${d}p" "$t/sum" | grep -q '^ size      delivered  code 0 · test 0 · docs [0-9,]* · test:code n/a$' \
  || fail "the delivered line changed: $(sed -n "${d}p" "$t/sum")"
want="           $n delivered lines match no stack in metrics.stacks"
[ "$(sed -n "$((d + 1))p" "$t/sum")" = "$want" ] || fail "the line after delivered is '$(sed -n "$((d + 1))p" "$t/sum")', want '$want'"
# The JSON is unchanged: no new keys in size.
[ "$(jq -c '.size.delivered | keys' "$t/m.json")" = '["code","deleted","docs","other","test"]' ] || fail "size.delivered's keys changed: $(jq -c '.size.delivered | keys' "$t/m.json")"

# --- a brief that delivered Go code keeps its numbers ---
CONFIG='[metrics]
stacks = ["go"]
' runrepo g
cat >"$S/script.sh" <<'EOF'
stub_after() {
  [ "$PHASE" = work ] || return 0
  printf 'package main\n\nfunc main() {\n}\n' >main.go
  printf '#!/bin/sh\necho one\n' >tool.sh
}
EOF
vrun g.out; got=$?
[ "$got" = 0 ] || fail "the stubbed Go run exited $got, want 0 — $(tail -5 "$t/g.out.err")"
(cd "$R" && "$B" metrics "$N" --json) >"$t/m.json" 2>"$t/err" || fail "metrics --json: $(cat "$t/err")"
code=$(jq .size.delivered.code "$t/m.json")
[ "$code" -gt 0 ] || fail "the Go fixture delivered no code"
(cd "$R" && "$B" metrics) >"$t/table" 2>"$t/err" || fail "metrics: $(cat "$t/err")"
[ "$(col "$t/table" code "$N")" = "$code" ] && [ "$(col "$t/table" test "$N")" = 0 ] \
  || fail "a brief with $code code lines shows code '$(col "$t/table" code "$N")' and test '$(col "$t/table" test "$N")' in the table"
echo "gate T9: ok"
