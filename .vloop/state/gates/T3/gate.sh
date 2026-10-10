#!/bin/sh
# Gate T3 — the driver times every session (brief R3, the record): a stub plan
# session that runs 3 s and reports duration_ms 20 leaves a record with
# duration_ms 20, wall_ms >= 3000 and ended; every record has both; the
# session schemas declare them, optional, with no version change.
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

# A plan session that runs 3 s and reports duration_ms 20.
runrepo w
cat >"$S/script.sh" <<'EOF'
[ "$PHASE" = plan ] && STUB_SLEEP=3 && STUB_DUR=20
EOF
vrun w.out; got=$?
[ "$got" = 0 ] || fail "the stubbed run exited $got, want 0 — $(tail -5 "$t/w.out.err")"
P=$(records '*-plan.json')
[ "$(echo "$P" | grep -c .)" = 1 ] || fail "want one plan session record, found: $P"
jq -e '.schema == "session/v1" and .duration_ms == 20' "$P" >/dev/null || fail "the plan record does not keep duration_ms 20 as claude reported it, in session/v1: $(jq -c '{schema, duration_ms}' "$P")"
jq -e '(.wall_ms | type) == "number" and .wall_ms >= 3000 and .wall_ms < 60000' "$P" >/dev/null \
  || fail "the plan record's wall_ms is not the driver's measurement of a 3 s session: $(jq -c '{wall_ms, duration_ms}' "$P")"
for f in $(records '*.json'); do
  jq -e '(.ended | type) == "string" and (.ended | test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$")) and (.wall_ms | type) == "number" and .wall_ms >= 0 and ((.ended | fromdateiso8601) >= (.started | fromdateiso8601))' "$f" >/dev/null \
    || fail "$(basename "$f") has no ended (RFC 3339 UTC, like started, not before it) or no wall_ms: $(jq -c '{started, ended, wall_ms}' "$f")"
done
jq -e '((.ended | fromdateiso8601) - (.started | fromdateiso8601)) >= 2' "$P" >/dev/null \
  || fail "the plan record's ended is not about 3 s after its started: $(jq -c '{started, ended}' "$P")"
GR=$(records '*-gate-review.json')
[ -n "$GR" ] && jq -e '.schema == "session/v2"' $GR >/dev/null || fail "the gate review record is not session/v2"

# The schemas declare both fields, optional: today's record and one without them validate; a wrong type does not.
(cd "$R" && "$B" schema validate session/v1 "$P") >"$t/out" 2>&1 || fail "the plan record does not validate as session/v1: $(cat "$t/out")"
(cd "$R" && "$B" schema validate session/v2 $GR) >"$t/out" 2>&1 || fail "the gate review record does not validate as session/v2: $(cat "$t/out")"
for s in v1 v2; do
  jq 'del(.ended, .wall_ms)' "$P" >"$t/old.json" || fail jq
  (cd "$R" && "$B" schema validate "session/$s" "$t/old.json") >"$t/out" 2>&1 || fail "a record without ended and wall_ms no longer validates as session/$s: $(cat "$t/out")"
  jq '.wall_ms = "long"' "$P" >"$t/bad.json" || fail jq
  (cd "$R" && "$B" schema validate "session/$s" "$t/bad.json") >"$t/out" 2>&1 && fail "session/$s accepts a wall_ms that is not a number: it does not declare wall_ms"
  jq '.ended = 5' "$P" >"$t/bad.json" || fail jq
  (cd "$R" && "$B" schema validate "session/$s" "$t/bad.json") >"$t/out" 2>&1 && fail "session/$s accepts an ended that is not a string: it does not declare ended"
done
echo "gate T3: ok"
