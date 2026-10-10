#!/bin/sh
# Gate T2 — a halt during plan acceptance resumes acceptance (brief R2): a base
# gate that changes .git/config halts with the plan stamped and planning and the
# run folder kept; after restoring it, vloop run <brief> resumes acceptance
# (no second plan session), the plan commit holds both run folders and the
# metrics count both folders' sessions. A halt in the gate review resumes the
# gate review (bare vloop run) before any work. A plan its checks reject is
# still removed.
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

# stamped <label>: the halted plan is stamped with the run id, brief and branch, and still planning.
stamped() {
  br=$(git -C "$R" -c core.fsmonitor=false rev-parse --abbrev-ref HEAD) || fail "$1: no branch"
  jq -e --arg n "$N" --arg b "$BRIEF" --arg br "$br" '.run_id == $n and .brief == $b and .branch == $br and .status == "planning"' \
    "$R/.vloop/state/state.json" >/dev/null 2>&1 \
    || fail "$1: the plan left by a halt during acceptance is not stamped and planning (want run_id $N, brief $BRIEF, branch $br): $(jq -c '{run_id, brief, branch, status}' "$R/.vloop/state/state.json" 2>&1)"
}
count() { grep -c "^-p /vloop:$1" "$S/argv.log"; }
# reviewed_first: the last gate review started before the first work session.
reviewed_first() {
  awk '/^-p \/vloop:gate-review/ { g = NR } /^-p \/vloop:work/ && !w { w = NR } END { exit !(g && w && g < w) }' "$S/argv.log" \
    || fail "$1: work started before the gate review passed: $(grep '^-p ' "$S/argv.log" | cut -c1-30 | tr '\n' ';')"
}
folders() { ls -d "$R"/.vloop/state/runs/"$N"/*/ 2>/dev/null | wc -l | tr -d ' '; }

# --- a base gate halts acceptance; the re-run resumes it ---
runrepo a
jq '.tasks[0].verify = "if [ -f .vloop/tmp/tamper ]; then rm -f .vloop/tmp/tamper; printf '"'"'[core]\\n\\tfsmonitor = false\\n'"'"' >> .git/config; fi; test -f T1.out"' \
  "$G/plan.json" >"$S/plan.json" || fail "plan a"
cat >"$S/script.sh" <<'EOF'
case "$PHASE" in plan) STUB_COST=1.25;; gate-review) STUB_COST=0.5;; esac
EOF
cp "$R/.git/config" "$t/a.config" && mkdir -p "$R/.vloop/tmp" && : >"$R/.vloop/tmp/tamper" || fail "arm the tamper"
vrun a1.out; got=$?
[ "$got" = 9 ] || fail "a base gate that changed .git/config: exit $got, want 9 — $(tail -5 "$t/a1.out.err")"
grep -q 'the gate of T1 changed .git/config' "$t/a1.out.err" || fail "the halt does not blame the gate of T1: $(tail -3 "$t/a1.out.err")"
stamped "a halt in the base gates"
[ "$(folders)" = 1 ] || fail "the halted run folder was not kept"
[ -n "$(records '*-plan.json')" ] || fail "the halted run folder has no plan session record"
cp "$t/a.config" "$R/.git/config" || fail "restore .git/config"
vrun a2.out; got=$?
[ "$got" = 0 ] || fail "the re-run after restoring .git/config exited $got, want 0 — $(tail -5 "$t/a2.out.err")"
grep -Fq "resuming acceptance of $N: round 1" "$t/a2.out" || fail "the re-run does not say 'resuming acceptance of $N: round 1': $(head -5 "$t/a2.out")"
[ "$(count plan)" = 1 ] || fail "the re-run planned again: $(count plan) plan sessions, want 1"
[ "$(count gate-review)" = 1 ] || fail "$(count gate-review) gate review sessions, want 1"
reviewed_first "a halt in the base gates"
[ "$(folders)" = 2 ] || fail "want two run folders after the resume, found $(folders)"
pc=$(git -C "$R" log --format=%H --grep="^\[vloop\] plan $N\$") || fail "git log"
[ "$(echo "$pc" | grep -c .)" = 1 ] || fail "want one plan commit, found: $pc"
inplan=$(git -C "$R" show --name-only --format= "$pc" | sed -n "s|^\.vloop/state/runs/$N/\([^/]*\)/.*|\1|p" | sort -u | wc -l | tr -d ' ')
[ "$inplan" = 2 ] || fail "the plan commit includes $inplan run folder(s), want both"
(cd "$R" && "$B" metrics "$N" --json) >"$t/m.json" 2>"$t/err" || fail "metrics $N --json: $(cat "$t/err")"
jq -s 'map(.cost_usd) | add' $(records '*.json') >"$t/sum" || fail "sum the records"
jq -s 'map(.cost_usd) | add' $(records '*-plan.json') >"$t/plansum" || fail "sum the plan records"
jq -e --slurpfile s "$t/sum" --slurpfile p "$t/plansum" \
  '((.cost.total_usd - $s[0]) | fabs) < 0.001 and ((.cost.plan_usd - $p[0]) | fabs) < 0.001' "$t/m.json" >/dev/null \
  || fail "metrics do not count both run folders' sessions: total $(jq .cost.total_usd "$t/m.json") (records $(cat "$t/sum")), plan $(jq .cost.plan_usd "$t/m.json") (records $(cat "$t/plansum"))"

# --- a halt in the gate review resumes acceptance, never work ---
runrepo b
cat >"$S/script.sh" <<'EOF'
stub_after() {
  [ "$PHASE" = gate-review ] && [ ! -f "$dir/tampered" ] && : >"$dir/tampered" && printf '[core]\n\tfsmonitor = false\n' >>.git/config
  return 0
}
EOF
cp "$R/.git/config" "$t/b.config" || fail "save .git/config"
vrun b1.out; got=$?
[ "$got" = 9 ] || fail "a gate review that changed .git/config: exit $got, want 9 — $(tail -5 "$t/b1.out.err")"
stamped "a halt in the gate review"
cp "$t/b.config" "$R/.git/config" || fail "restore .git/config"
trust
(cd "$R" && env PATH="$S:$PATH" "$B" run) >"$t/b2.out" 2>"$t/b2.out.err"; got=$?
[ "$got" = 0 ] || fail "bare vloop run after a gate review halt exited $got, want 0 — $(tail -5 "$t/b2.out.err")"
grep -Fq "resuming acceptance of $N: round 1" "$t/b2.out" || fail "bare vloop run does not say 'resuming acceptance of $N: round 1': $(head -5 "$t/b2.out")"
[ "$(count plan)" = 1 ] || fail "the re-run after a gate review halt planned again"
[ "$(count gate-review)" = 2 ] || fail "the re-run after a gate review halt ran $(count gate-review) gate review(s) in all, want 2"
reviewed_first "a halt in the gate review"
jq -e '.gate_review.verdict == "PASS" and .status == "complete"' "$R/.vloop/state/state.json" >/dev/null \
  || fail "the resumed plan: $(jq -c '{status, gate_review}' "$R/.vloop/state/state.json")"

# --- a plan its checks reject is still removed ---
runrepo c
jq 'del(.tasks[0].verify)' "$G/plan.json" >"$S/plan.json" || fail "plan c"
vrun c1.out; got=$?
[ "$got" = 1 ] || fail "a plan with no verify: exit $got, want 1"
[ -e "$R/.vloop/state/state.json" ] && fail "a plan its checks rejected was kept"
echo "gate T2: ok"
