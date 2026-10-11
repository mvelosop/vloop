#!/bin/sh
# Gate T11 — the close: the worked example together on a fresh build, then the
# real-data check: this repository read by the new binary and by the base's
# (v0.8.0's code): metrics --json for every consumed brief equal except the
# wall time, which only grows; defect list, intervention list and brief list
# equal. The changed wall times are printed for the run record.
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

BASE=615dd19c754de2f82c977c0434d796dcd214955f
# --- the worked example, together, on a fresh build: one run whose work session
# writes editor keys, is denied a Bash command, and whose plan session runs 2 s
# while claude reports 20 ms ---
CONFIG='run.git-config-ignore = ["gitlens.*"]
' runrepo w
cat >"$S/script.sh" <<'EOF'
case "$PHASE" in
  plan) STUB_SLEEP=2; STUB_DUR=20;;
  work) STUB_COST=1.84; STUB_DENIALS='[{"tool_name":"Bash","tool_input":{"command":"rm -rf $X/tmp"}}]';;
esac
stub_after() {
  [ "$PHASE" = work ] || return 0
  printf '[branch "%s"]\n\tvscode-merge-base = refs/heads/main\n' "$(git rev-parse --abbrev-ref HEAD)" >>.git/config
  git config gitlens.x 1
}
EOF
vrun w.out; got=$?
[ "$got" = 0 ] || fail "the worked example's run exited $got, want 0 — $(tail -5 "$t/w.out.err")"
grep -Eq '^   work T1 · [0-9]+m[0-5][0-9]s · \$1\.84 · run \$[0-9]+\.[0-9]{2} / \$40\.00$' "$t/w.out" || fail "no per-session line for work T1 at \$1.84: $(grep ' · ' "$t/w.out")"
P=$(records '*-plan.json'); W=$(records '*-work.json')
jq -e '.wall_ms >= 2000 and .duration_ms == 20' "$P" >/dev/null || fail "the plan record: $(jq -c '{wall_ms, duration_ms}' "$P")"
jq -e '.permission_denials == [{"tool_name": "Bash", "command": "rm -rf $X/tmp"}]' "$W" >/dev/null || fail "the work record's denials: $(jq -c .permission_denials "$W")"
(cd "$R" && "$B" metrics "$N" --json) >"$t/m.json" 2>"$t/err" || fail "metrics: $(cat "$t/err")"
jq -e --slurpfile p "$P" '.time.plan_ms == $p[0].wall_ms' "$t/m.json" >/dev/null || fail "metrics plan time $(jq .time.plan_ms "$t/m.json") is not the plan record's wall_ms"

# --- the real-data check: this repository, read by this build and by the base's ---
mkdir -p "$t/base" && (git archive "$BASE" | tar -xf - -C "$t/base") || fail "git archive $BASE"
(cd "$t/base" && go build -o "$t/vloop-base" ./cmd/vloop) || fail "build the base"
cd "$top" || fail "cd"
"$B" brief list --json >"$t/briefs.json" || fail "brief list --json on this repository"
jq -r '.[] | select(.status == "consumed") | .path' "$t/briefs.json" >"$t/consumed"
[ "$(wc -l <"$t/consumed" | tr -d ' ')" -ge 10 ] || fail "fewer than ten consumed briefs listed: $(cat "$t/consumed")"
: >"$t/changed"
while read -r b; do
  "$t/vloop-base" metrics --json "$b" >"$t/m.old" 2>"$t/err" || fail "the base's metrics --json $b: $(cat "$t/err")"
  "$B" metrics --json "$b" >"$t/m.new" 2>"$t/err" || fail "metrics --json $b: $(cat "$t/err")"
  for f in m.old m.new; do jq -S 'del(.time.wall_ms)' "$t/$f" >"$t/$f.cmp" || fail "metrics --json $b is not JSON"; done
  cmp -s "$t/m.old.cmp" "$t/m.new.cmp" || fail "metrics --json for $b differs from the base's beyond the wall time: $(diff "$t/m.old.cmp" "$t/m.new.cmp" | head -8)"
  o=$(jq '.time.wall_ms' "$t/m.old"); n=$(jq '.time.wall_ms' "$t/m.new")
  if [ "$o" != "$n" ]; then
    [ "$o" = null ] || [ "$n" = null ] && fail "$b: the wall time went from $o to $n"
    [ "$n" -gt "$o" ] || fail "$b: the wall time shrank from $o to $n ms — planning comes before the plan commit"
    echo "$b: wall_ms $o -> $n" >>"$t/changed"
  fi
done <"$t/consumed"
[ -s "$t/changed" ] || fail "no consumed brief's wall time changed — the wall still starts at the plan commit"
for c in "defect list" "intervention list" "brief list"; do
  # shellcheck disable=SC2086
  "$t/vloop-base" $c >"$t/l.old" 2>&1 || fail "the base's $c failed: $(cat "$t/l.old")"
  # shellcheck disable=SC2086
  "$B" $c >"$t/l.new" 2>&1 || fail "vloop $c failed on this repository: $(cat "$t/l.new")"
  cmp -s "$t/l.old" "$t/l.new" || fail "vloop $c reads this repository's records differently from the base: $(diff "$t/l.old" "$t/l.new" | head -6)"
done
echo "changed wall times (for the run record):"; cat "$t/changed"
echo "gate T11: ok"
