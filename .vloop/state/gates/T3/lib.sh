# Shared harness for B16's gates: a fresh build (host, linux, windows), a
# temporary HOME, and a stub claude that answers doctor's probes and plays
# every session of `vloop run` — plan, gate-review, work, review and verify.
# Nothing touches the real HOME or this repository; every repository the gate
# makes lives under $t.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) GOCACHE=$(go env GOCACHE) GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build ./... || fail "GOOS=linux go build ./..."
GOOS=windows go build ./... || fail "GOOS=windows go build ./..."
B="$t/vloop"
SYS=/usr/bin:/bin
# The plan's base, for "as today" comparisons.
BASE=744f173872ad88141bb820e6969245b18db8c8c1
# basebin: $t/vloop-base, built from the plan's base.
basebin() {
  [ -x "$t/vloop-base" ] && return 0
  mkdir -p "$t/base" && git -C "$top" archive "$BASE" | tar -x -C "$t/base" || fail "git archive $BASE"
  (cd "$t/base" && go build -buildvcs=false -o "$t/vloop-base" ./cmd/vloop) || fail "build of the base"
}
# stubs: the stub claude in $t/stub, and an empty argv log. Files in $t/stub
# steer it:
#   plan.json            the plan the plan session writes
#   <phase>.sh           sourced in that phase's session, in the repository
#   work-<task>.sh       sourced in the work session of that task
#   fail-<task>          the work session of that task leaves its gate failing, once
#   gr-fail              the gate review fails the plan, once
#   cost, cost-<phase>   what a session costs (0.1 by default)
#   exit-<phase>         that phase's session prints nothing and exits with the code
# Each session leaves: fence-<phase>.json (the --settings it got), a line in
# norunning when .vloop/tmp/.running was missing, and verify.json is written by
# the verify session from verify-report.json when no verify.sh says otherwise.
stubs() {
  V=$("$B" version --json | "$JQ" -r .version) && [ -n "$V" ] && [ "$V" != null ] || fail "vloop version --json gave no version"
  mkdir -p "$t/stub" "$t/fakehome/gateuser" || fail mkdir
  export HOME="$t/fakehome/gateuser"
  [ -f "$HOME/.claude.json" ] || printf '{"projects":{}}\n' >"$HOME/.claude.json" || fail trust
  printf '[{"id":"vloop@vloop","version":"%s","enabled":true}]\n' "$V" >"$t/stub/plugins.json" || fail plugins
  ln -s "$B" "$t/stub/vloop" || fail "vloop on the stub's PATH"
  cat >"$t/stub/claude" <<'STUB'
#!/bin/sh
dir=$(dirname "$0")
case "$1" in
--version) echo "2.0.0 (Claude Code)"; exit 0 ;;
plugin) [ "$2" = list ] && cat "$dir/plugins.json" && exit 0; exit 1 ;;
esac
echo "$*" >>"$dir/argv.log"
prompt="" MODEL="" SETTINGS=""
while [ $# -gt 0 ]; do
  case "$1" in -p) prompt=$2; shift 2 ;; --model) MODEL=$2; shift 2 ;; --settings) SETTINGS=$2; shift 2 ;; *) shift ;; esac
done
case "${prompt%% *}" in
/vloop:plan) PHASE=plan TASK= ;;
/vloop:gate-review) PHASE=gate-review TASK= ;;
/vloop:work) PHASE=work TASK=${prompt#* } ;;
/vloop:review) PHASE=review TASK=${prompt#* } ;;
/vloop:verify) PHASE=verify TASK= ;;
*) echo "stub claude: unrecognised prompt: $prompt" >&2; exit 64 ;;
esac
[ "$PHASE" = verify ] && [ "$prompt" != /vloop:verify ] && echo "verify-arg $prompt" >>"$dir/norunning"
[ -f .vloop/tmp/.running ] || echo "$PHASE $TASK" >>"$dir/norunning"
[ -n "$SETTINGS" ] && [ -f "$SETTINGS" ] && cp "$SETTINGS" "$dir/fence-$PHASE.json"
if [ -f "$dir/exit-$PHASE" ]; then exit "$(cat "$dir/exit-$PHASE")"; fi
mkdir -p .vloop/tmp
case "$PHASE" in
plan)
  mkdir -p .vloop/state && cp "$dir/plan.json" .vloop/state/state.json
  if [ -f "$dir/plan.sh" ]; then . "$dir/plan.sh"; fi ;;
gate-review)
  if [ -f "$dir/gr-fail" ]; then
    rm -f "$dir/gr-fail"
    echo '{"schema":"gate-verdict/v1","verdict":"FAIL","tasks":[{"task":"T1","verdict":"FAIL","findings":[{"summary":"planted","kind":"wrong-reason"}]}],"notes":"planted"}' >.vloop/tmp/gate-verdict.json
  else
    echo '{"schema":"gate-verdict/v1","verdict":"PASS","tasks":[],"notes":"none"}' >.vloop/tmp/gate-verdict.json
  fi ;;
work)
  # fail-<task>: this attempt leaves the gate failing, once.
  if [ -f "$dir/fail-$TASK" ]; then rm -f "$dir/fail-$TASK"; else : >"$TASK.out"; fi
  if [ -f "$dir/work-$TASK.sh" ]; then . "$dir/work-$TASK.sh"; fi
  printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$TASK" "$TASK" >.vloop/tmp/proposal.json ;;
review)
  if [ -f "$dir/review.sh" ]; then . "$dir/review.sh"; fi
  printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$TASK" >.vloop/tmp/verdict.json ;;
verify)
  if [ -f "$dir/verify.sh" ]; then . "$dir/verify.sh"
  elif [ -f "$dir/verify-report.json" ]; then cp "$dir/verify-report.json" .vloop/tmp/verify.json; fi ;;
esac
# cost: what each session costs, 0.1 unless a file says otherwise.
c=0.1; [ -f "$dir/cost" ] && c=$(cat "$dir/cost"); [ -f "$dir/cost-$PHASE" ] && c=$(cat "$dir/cost-$PHASE")
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":%s,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"costUSD":%s}}}\n' "$c" "$MODEL" "$c"
STUB
  chmod +x "$t/stub/claude" || fail chmod
  : >"$t/stub/argv.log"
}
# resetstub: forget every steering file and every trace a run left.
resetstub() {
  [ -d "$t/stub" ] || return 0
  (cd "$t/stub" && rm -f ./*.sh fail-* gr-fail cost cost-* exit-* fence-*.json norunning verify-report.json) || fail resetstub
  : >"$t/stub/argv.log"
}
RUN=B20260101-0900-demo
BRIEF=docs/briefs/$RUN.loop-brief.md
# mkplan <task id>...: the state/v2 plan the stub's plan session writes; task
# <id> passes when <id>.out exists, and depends on the one before it.
mkplan() {
  printf '%s\n' "$@" | "$JQ" -R . | "$JQ" -s '[ to_entries[] | {id: .value, title: ("Task " + .value),
    goal: ("Do " + .value + "."), kind: "feature", fixtures: "", references: [],
    depends_on: (if .key == 0 then [] else ["T" + (.key | tostring)] end),
    acceptance: [(.value + ".out exists")], verify: ("test -f " + .value + ".out"),
    status: "pending", attempts: 0, notes: ""} ]' >"$t/tasks.json" || fail tasks.json
  "$JQ" -n --slurpfile ts "$t/tasks.json" '{schema: "state/v2", run_id: "", brief: "", base: "", branch: "",
    status: "planning", iteration: 0, created: "2026-01-01T09:00:00Z", updated: "2026-01-01T09:00:00Z",
    shell: "sh", checks: [], gate_scratch: [], gate_review: {rounds: 0, verdict: ""}, tasks: $ts[0]}' >"$t/stub/plan.json" || fail plan.json
}
# inr <args...>: run in $R with the stub first on PATH.
inr() { (cd "$R" && env PATH="$t/stub:$SYS" "$@"); }
# newrepo <name> [<init args>...]: a git repository on main where vloop init
# ran, with one [[check]], trusted in $HOME/.claude.json; $R is its path.
newrepo() {
  [ -x "$t/stub/claude" ] || stubs
  R="$t/$1"; shift
  mkdir -p "$R/docs/briefs" && git -C "$R" init -q -b main || fail "git init $R"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail identity
  printf 'x\n' >"$R/docs/x.md"
  printf '# demo\n\nThe demo greets.\n' >"$R/README.md"
  inr "$B" init "$@" </dev/null >"$R.init" 2>&1 || fail "vloop init in $R: $(cat "$R.init")"
  printf '\n[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >>"$R/.vloop/config.toml" || fail config
  "$JQ" --arg p "$R" '.projects[$p] = {hasTrustDialogAccepted: true}' "$HOME/.claude.json" >"$t/trust.json" &&
    mv "$t/trust.json" "$HOME/.claude.json" || fail trust
}
commit() { git -C "$R" add -A && git -C "$R" commit -q -m "$1" || fail "commit $1"; }
# briefdoc <run id> <status>: a loop brief that passes vloop brief check.
briefdoc() {
  cat <<BRIEF
---
name: $1.loop-brief
description: Build a demo
kind: brief
status: $2
created: 2026-01-01
seeds: A run that builds a demo
depends-on: []
---
# Brief

**Status:** ready to plan

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
# putbrief <run id> <status>: write the brief into $R (not committed).
putbrief() { briefdoc "$1" "$2" >"$R/docs/briefs/$1.loop-brief.md" || fail "brief $1"; }
# runrepo <name> <task id>...: newrepo with the ready brief $RUN and a plan of
# those tasks, all committed on main.
runrepo() { n=$1; shift; newrepo "$n"; mkplan "$@"; putbrief "$RUN" ready; commit harness; }
# vrun <out> [<run args>...]: vloop run on the brief in $R; output in $t/<out>,
# exit in $got.
vrun() { o=$1; shift; : >"$t/stub/argv.log"; inr "$B" run "$BRIEF" "$@" </dev/null >"$t/$o" 2>&1; got=$?; }
# lastdir: the newest run folder of $RUN in $R, repo-relative, ending in /.
lastdir() {
  d=$(ls -1 "$R/.vloop/state/runs/$RUN" 2>/dev/null | sort | tail -n 1)
  [ -n "$d" ] || fail "no run folder for $RUN"
  echo ".vloop/state/runs/$RUN/$d/"
}
# vreport <file> <n>: a verify/v1 report with the brief's first <n> example findings.
vreport() {
  "$JQ" -n --argjson n "$2" '{schema: "verify/v1",
    findings: ([{kind: "example", claim: "exit 2", observed: "exit 1", where: "Worked example, line 3; vloop greet --loud"},
               {kind: "surface", claim: "greet prints in upper case with --loud", observed: "README.md says --loud is not supported", where: "README.md:12"}][0:$n]),
    checked: {example_lines: 3, sentences: 4, surfaces: 2}, notes: "stub"}' >"$1" || fail "verify report $1"
}
# sq <file>: the file's lines with runs of blanks squeezed and the ends trimmed.
sq() { sed -e 's/[[:space:]][[:space:]]*/ /g' -e 's/^ //' -e 's/ $//' "$1"; }
# retire <brief check output>: that output with the retired must-change list:
# and external format: lines dropped and its warning count lowered by as many.
retire() {
  k=$(grep -c -e 'must-change list:' -e 'external format:' "$1")
  grep -v -e 'must-change list:' -e 'external format:' "$1" | awk -v k="$k" '
    / warning\(s\)$/ { n = $0; sub(/ warning\(s\)$/, "", n); sub(/.* /, "", n); sub(/[0-9]+ warning\(s\)$/, (n - k) " warning(s)") } { print }'
}
