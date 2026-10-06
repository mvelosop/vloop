#!/bin/sh
# Gate T11 — metrics that count right (brief Q7), on stubbed vloop runs:
# 1. a task reverted by a gate regression and redone is not first-pass: T2's
#    work breaks T1 once, T1 is redone, so first-pass is 1 of 2 — and a clean
#    two-task run is still 2 of 2;
# 2. the convergence ratio divides this run's iterations by the tasks this run
#    closed: a first run closes T1 and T2 and halts at its iteration budget; the
#    resumed run closes nothing in its two iterations, so with
#    run.convergence-min 2 and run.convergence-max 1.5 it halts not converging,
#    exit 5, before T3 reaches its attempt ceiling (3) and blocks.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
mkdir -p "$t/home" || fail mkdir
export HOME="$t/home"
N=B20260101-0900-demo
BRIEF=docs/briefs/$N.loop-brief.md
# runrepo <name> <plan>: a vloop repository on main with a ready brief and a
# check, committed and trusted; its own stub claude playing <plan>.
runrepo() {
  R="$t/$1" S="$t/$1.stub"
  mkdir -p "$R" "$S" && git -C "$R" init -q -b main || fail "git init $1"
  cp "$G/claude" "$S/claude" && chmod +x "$S/claude" && cp "$G/$2" "$S/plan.json" || fail "stub for $1"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail "identity"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$R" >"$HOME/.claude.json"
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$R/.vloop/config.toml" || fail config
  mkdir -p "$R/docs/briefs" && echo x >"$R/docs/x.md" && cp "$G/brief.md" "$R/$BRIEF" || fail brief
  git -C "$R" add -A && git -C "$R" commit -q -m harness || fail "commit $1"
}
# vrun <out> [VAR=value...]: vloop run in $R with the stub on PATH; returns its exit code.
vrun() {
  o=$1; shift
  (cd "$R" && env PATH="$S:$PATH" "$@" "$B" run "$BRIEF") >"$t/$o" 2>&1
}
firstpass() { (cd "$R" && "$B" metrics "$N" --json) 2>"$t/err" | jq -r '"\(.tasks.first_pass)/\(.tasks.planned)"'; }

# --- first-pass ---
runrepo clean two.json
vrun clean.out || fail "the clean stubbed run failed: $(tail -15 "$t/clean.out")"
[ "$(firstpass)" = 2/2 ] || fail "a clean two-task run is not first-pass 2/2: $(firstpass) $(cat "$t/err")"
runrepo reg two.json
cp "$G/clobber.sh" "$S/script.sh" || fail script
vrun reg.out || fail "the regression run failed: $(tail -15 "$t/reg.out")"
grep -q 'GATE REGRESSION T1' "$t/reg.out" || fail "the fixture run did not regress T1: $(tail -20 "$t/reg.out")"
[ "$(firstpass)" = 1/2 ] || fail "first-pass is $(firstpass) after T1 was reverted by a gate regression and redone, want 1/2"

# --- convergence per run ---
runrepo conv three.json
vrun conv1.out VLOOP_RUN_MAX_ITERATIONS=2; got=$?
[ "$got" = 4 ] || fail "the first run did not halt at its iteration budget (exit $got, want 4): $(tail -15 "$t/conv1.out")"
[ "$(cd "$R" && "$B" status --json | jq -r '.done')" = 2 ] || fail "the first run did not close T1 and T2"
cp "$G/stuck.sh" "$S/script.sh" || fail script
vrun conv2.out VLOOP_RUN_CONVERGENCE_MIN=2 VLOOP_RUN_CONVERGENCE_MAX=1.5 VLOOP_RUN_MAX_ATTEMPTS=3; got=$?
[ "$got" = 5 ] || fail "the resumed run, two iterations and no task closed, exited $got, want 5 (not converging): the ratio counts tasks earlier runs closed — $(grep -i 'converg\|blocked\|attempt' "$t/conv2.out" | head -3)"
grep -qi 'not converging' "$t/conv2.out" || fail "the resumed run does not say it is not converging"
echo "gate T11: ok"
