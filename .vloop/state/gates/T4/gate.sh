#!/bin/sh
# Gate T4 — the metrics take session time from wall_ms (brief R3, the metrics):
# per-phase time is the records' wall_ms, not the CLI's duration_ms, and the
# wall runs from the first plan session's started to the last run commit.
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

# A plan session that runs 3 s and reports duration_ms 20; work and review report
# 5 minutes each but take well under a second.
runrepo w
cat >"$S/script.sh" <<'EOF'
case "$PHASE" in plan) STUB_SLEEP=3; STUB_DUR=20;; *) STUB_DUR=300000;; esac
EOF
vrun w.out; got=$?
[ "$got" = 0 ] || fail "the stubbed run exited $got, want 0 — $(tail -5 "$t/w.out.err")"
for f in $(records '*.json'); do
  jq -e '(.wall_ms | type) == "number"' "$f" >/dev/null || fail "$(basename "$f") has no wall_ms — the session record has to carry it first"
done
(cd "$R" && "$B" metrics "$N" --json) >"$t/m.json" 2>"$t/err" || fail "metrics $N --json: $(cat "$t/err")"
# phase <phase>: the wall_ms of the records of that phase, summed. By the record's
# phase, not its file name: *-review.json also matches NNN-gate-review.json.
phase() { jq -s --arg p "$1" 'map(select(.phase == $p) | .wall_ms) | add // 0' $(records '*.json') /dev/null; }
for p in plan:plan_ms work:work_ms review:review_ms gate-review:gate_review_ms; do
  key=${p#*:}
  want=$(phase "${p%%:*}")
  [ "$(jq ".time.$key" "$t/m.json")" = "$want" ] || fail "metrics time.$key is $(jq ".time.$key" "$t/m.json"), want $want — the records' wall_ms, not their duration_ms"
done
w=$(jq '.time.agent_ms' "$t/m.json"); [ "$w" = "$(( $(jq .time.work_ms "$t/m.json") + $(jq .time.review_ms "$t/m.json") ))" ] || fail "agent_ms is not work + review"
# The wall runs from the first plan session's started to the last run commit.
started=$(jq -s 'map(.started | fromdateiso8601) | min' $(records '*-plan.json')) || fail "plan started"
last=$(git -C "$R" log -1 --format=%ct --grep="^\[vloop\] run $N/") || fail "no run commit"
[ -n "$last" ] || fail "no run commit"
want=$(( (last - started) * 1000 ))
jq -e --argjson w "$want" '(.time.wall_ms | type) == "number" and ((.time.wall_ms - $w) | fabs) <= 1000' "$t/m.json" >/dev/null \
  || fail "metrics time.wall_ms is $(jq .time.wall_ms "$t/m.json"), want about $want — from the plan session's started to the last run commit"
echo "gate T4: ok"
