# Shared harness: a fresh build (host, linux, windows), a temporary HOME, and a
# stub claude that answers doctor's two probes and plays every session of
# `vloop run`, logging each session's arguments, one line per session, to
# $t/stub/argv.log. Nothing touches the real HOME or this repository.
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
mkdir -p "$t/stub" "$t/fakehome/gateuser" || fail mkdir
export HOME="$t/fakehome/gateuser"
printf '{"projects":{}}\n' >"$HOME/.claude.json" || fail trust
printf '[{"id":"vloop@vloop","version":"%s","enabled":true}]\n' "$V" >"$t/stub/plugins.json" || fail plugins
cat >"$t/stub/claude" <<'STUB'
#!/bin/sh
dir=$(dirname "$0")
case "$1" in
--version) echo "2.0.0 (Claude Code)"; exit 0 ;;
plugin) [ "$2" = list ] && cat "$dir/plugins.json" && exit 0; exit 1 ;;
esac
echo "$*" >>"$dir/argv.log"
prompt="" MODEL=""
while [ $# -gt 0 ]; do
  case "$1" in -p) prompt=$2; shift 2 ;; --model) MODEL=$2; shift 2 ;; *) shift ;; esac
done
case "${prompt%% *}" in
/vloop:plan) PHASE=plan TASK= ;;
/vloop:gate-review) PHASE=gate-review TASK= ;;
/vloop:work) PHASE=work TASK=${prompt#* } ;;
/vloop:review) PHASE=review TASK=${prompt#* } ;;
*) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64 ;;
esac
mkdir -p .vloop/tmp
case "$PHASE" in
plan) mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json ;;
gate-review) echo '{"schema":"gate-verdict/v1","verdict":"PASS","tasks":[],"notes":"none"}' >.vloop/tmp/gate-verdict.json ;;
work) : >"$TASK.out"
  printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$TASK" "$TASK" >.vloop/tmp/proposal.json ;;
review) printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$TASK" >.vloop/tmp/verdict.json ;;
esac
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":0.1,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"costUSD":0.1}}}\n' "$MODEL"
STUB
chmod +x "$t/stub/claude" || fail chmod
RUN=B20260101-0900-demo
BRIEF=docs/briefs/$RUN.loop-brief.md
# A one-task state/v2 plan, as the stub's plan session writes it.
"$JQ" -n --arg brief "$BRIEF" '{schema: "state/v2", run_id: "", brief: "", base: "", branch: "",
  status: "planning", iteration: 0, created: "2026-01-01T09:00:00Z", updated: "2026-01-01T09:00:00Z",
  shell: "sh", checks: [], gate_scratch: [], gate_review: {rounds: 0, verdict: ""},
  tasks: [{id: "T1", title: "Task T1", goal: "Do T1.", kind: "feature", fixtures: "", references: [],
    depends_on: [], acceptance: ["T1.out exists"], verify: "test -f T1.out", status: "pending",
    attempts: 0, notes: ""}]}' >"$t/stub/plan.json" || fail "plan.json"

# inr <args...>: run in $R with the stub first on PATH.
inr() { (cd "$R" && env PATH="$t/stub:$SYS" "$@"); }
# newrepo <name>: a git repository on main where vloop init ran, with one
# [[check]], trusted in $HOME/.claude.json; $R is its path, $t/<name>.init the
# output of init.
newrepo() {
  R="$t/$1"
  mkdir -p "$R/docs" && git -C "$R" init -q -b main || fail "git init $1"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail identity
  printf 'x\n' >"$R/docs/x.md"
  inr "$B" init </dev/null >"$t/$1.init" 2>&1 || fail "vloop init in $1: $(cat "$t/$1.init")"
  printf '\n[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >>"$R/.vloop/config.toml" || fail config
  "$JQ" --arg p "$R" '.projects[$p] = {hasTrustDialogAccepted: true}' "$HOME/.claude.json" >"$t/trust.json" &&
    mv "$t/trust.json" "$HOME/.claude.json" || fail trust
}
commit() { git -C "$R" add -A && git -C "$R" commit -q -m "$1" || fail "commit $1"; }
# brief: a ready loop brief that passes vloop brief check.
brief() {
  cat <<BRIEF
---
name: $RUN.loop-brief
description: Build a demo
kind: brief
status: ready
created: 2026-01-01
seeds: A run that builds a demo
depends-on: []
---
# Brief

## What it is

A thing.

## Why this shape, and what was rejected

Because.

## Binding references

- \`docs/x.md\` — the error contract

## Behaviour contract

An unknown id exits 1.

## Worked example

\`\`\`
\$ demo
ok
\`\`\`

## Out of scope

- one
- two

## Constraints

- Go.

## Shape

2 to 3 tasks.
BRIEF
}
# runrepo <name>: newrepo with the ready brief, all committed on main.
runrepo() { newrepo "$1"; mkdir -p "$R/docs/briefs" && brief >"$R/$BRIEF" || fail brief; commit harness; }
# vrun <out>: vloop run on the brief in $R; output in $t/<out>, exit in $got.
vrun() { : >"$t/stub/argv.log"; inr "$B" run "$BRIEF" </dev/null >"$t/$1" 2>&1; got=$?; }
# sessions: the logged session lines, one per session.
sessions() { cat "$t/stub/argv.log"; }
