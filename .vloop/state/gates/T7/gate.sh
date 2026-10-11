#!/bin/sh
# Gate T7 — vloop status shows the live run (brief R5, the lock and status):
# while a stub plan session runs, with no plan yet, status prints the running
# line and exits 0; during the work session of T1 it prints the plan's line,
# then the running line, --json carries running, and the lock records the
# session; with no live run (finished, or a lock whose process is gone) neither.
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

runrepo s
cat >"$S/script.sh" <<'EOF'
case "$PHASE" in
  plan) : >"$dir/in-plan"; STUB_SLEEP=3;;
  work) : >"$dir/in-work"; STUB_SLEEP=3;;
esac
EOF
trust
(cd "$R" && env PATH="$S:$PATH" "$B" run "$BRIEF") >"$t/run.out" 2>"$t/run.err" &
pid=$!
# waitfor <file>: poll for the stub's marker, up to 30 s.
waitfor() {
  n=0
  while [ ! -f "$1" ]; do
    n=$((n + 1)); [ "$n" -gt 300 ] && { kill "$pid" 2>/dev/null; fail "the stubbed run never reached $(basename "$1"): $(tail -5 "$t/run.err")"; }
    sleep 0.1
  done
}
st() { (cd "$R" && "$B" status "$@") >"$t/st.out" 2>"$t/st.err"; }

# --- planning, no plan yet: the running line, exit 0 ---
waitfor "$S/in-plan"
sleep 1
[ -e "$R/.vloop/state/state.json" ] && fail "the fixture wrote the plan before the status check"
st; got=$?
[ "$got" = 0 ] || { kill "$pid"; fail "status while a live run plans exited $got, want 0: $(cat "$t/st.err")"; }
grep -Eq '^running: plan for 0m0[0-3]s · \$0\.00 of \$40\.00$' "$t/st.out" || { kill "$pid"; fail "status while planning printed: $(cat "$t/st.out")"; }

# --- working T1: the plan's line, then the running line; the lock records the session ---
waitfor "$S/in-work"
sleep 1
spent=$(jq -s 'map(.cost_usd) | add // 0' $(records '*.json') /dev/null)
st; got=$?
[ "$got" = 0 ] || { kill "$pid"; fail "status during a work session exited $got: $(cat "$t/st.err")"; }
want=$(printf 'running: work T1 for 0m0[0-3]s · \\$%.2f of \\$40\\.00' "$spent" | sed 's/\./\\./')
sed -n 2p "$t/st.out" | grep -Eq "^$want\$" || { kill "$pid"; fail "status during the work session of T1 printed (want 'running: work T1 for 0m0<s>s · \$$(printf '%.2f' "$spent") of \$40.00' second): $(cat "$t/st.out")"; }
head -1 "$t/st.out" | grep -q "^$N — " || { kill "$pid"; fail "status's first line is not the plan's: $(head -1 "$t/st.out")"; }
st --json; got=$?
jq -e --argjson s "$spent" '.running.phase == "work" and .running.task == "T1" and ((.running.spent - $s) | fabs) < 0.001 and .running.ceiling == 40
  and (.running.session_started | test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}"))' "$t/st.out" >/dev/null \
  || { kill "$pid"; fail "status --json during the work session of T1: $(cat "$t/st.out")"; }
jq -e --argjson s "$spent" '.phase == "work" and .task == "T1" and (.session_started | type) == "string" and ((.spent - $s) | fabs) < 0.001' "$R/.vloop/tmp/.running" >/dev/null \
  || { kill "$pid"; fail "the run lock does not record the session in flight: $(cat "$R/.vloop/tmp/.running")"; }

wait "$pid"; got=$?
[ "$got" = 0 ] || fail "the stubbed run exited $got, want 0 — $(tail -5 "$t/run.err")"

# --- no live run: no running line, no running key ---
st; grep -q '^running:' "$t/st.out" && fail "status prints a running line with no live run"
st --json; jq -e 'has("running") | not' "$t/st.out" >/dev/null || fail "status --json has running with no live run"
sh -c 'exit 0' & dead=$!; wait "$dead"
jq -n --argjson p "$dead" '{pid: $p, branch: "x", started: "2026-01-01T00:00:00Z", run: "x", phase: "work", task: "T1", session_started: "2026-01-01T00:00:00Z", spent: 1}' >"$R/.vloop/tmp/.running"
st --json; jq -e 'has("running") | not' "$t/st.out" >/dev/null || fail "status --json shows a run whose process is gone"
rm -f "$R/.vloop/tmp/.running"
echo "gate T7: ok"
