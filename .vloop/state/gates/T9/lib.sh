# Shared harness for B15's gates: a fresh build (host, linux, windows), a
# temporary HOME, and a stub claude that answers doctor's probes and plays
# every session of `vloop run`. Nothing touches the real HOME or this
# repository; every repository the gate makes lives under $t.
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
GOOS=linux go build ./... || fail "GOOS=linux go build ./..."
GOOS=windows go build ./... || fail "GOOS=windows go build ./..."
B="$t/vloop"
SYS=/usr/bin:/bin
# The plan's base, for "as today" comparisons.
BASE=4ece10df6362e2939d4dad3da319e430e3b044bd
# basebin: $t/vloop-base, built from the plan's base.
basebin() {
  mkdir -p "$t/base" && git -C "$top" archive "$BASE" | tar -x -C "$t/base" || fail "git archive $BASE"
  (cd "$t/base" && go build -buildvcs=false -o "$t/vloop-base" ./cmd/vloop) || fail "build of the base"
}
# astcheck <mode> <root>: the gate's source checker, built outside any module.
astcheck() {
  if [ ! -x "$t/astcheck" ]; then
    mkdir -p "$t/ast" && cp "$G/astcheck.go.txt" "$t/ast/astcheck.go" || fail "copy astcheck"
    (cd "$t/ast" && GO111MODULE=on GOFLAGS= go build -o "$t/astcheck" astcheck.go) || fail "build astcheck"
  fi
  "$t/astcheck" "$@"
}
# stubs: the stub claude in $t/stub, and an empty argv log.
stubs() {
  V=$("$B" version --json | "$JQ" -r .version) && [ -n "$V" ] && [ "$V" != null ] || fail "vloop version --json gave no version"
  mkdir -p "$t/stub" "$t/fakehome/gateuser" || fail mkdir
  export HOME="$t/fakehome/gateuser"
  [ -f "$HOME/.claude.json" ] || printf '{"projects":{}}\n' >"$HOME/.claude.json" || fail trust
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
work)
  # fail-<task>: this attempt leaves the gate failing, once.
  if [ -f "$dir/fail-$TASK" ]; then rm -f "$dir/fail-$TASK"; else : >"$TASK.out"; fi
  if [ -f "$dir/work-$TASK.sh" ]; then . "$dir/work-$TASK.sh"; fi
  printf '{"schema":"proposal/v1","task":"%s","outcome":"done","summary":"made %s","files":[],"verified":"ok","notes":"none"}\n' "$TASK" "$TASK" >.vloop/tmp/proposal.json ;;
review) printf '{"schema":"verdict/v1","task":"%s","verdict":"PASS","criteria":[],"findings":[],"notes":"none"}\n' "$TASK" >.vloop/tmp/verdict.json ;;
esac
# cost: what each session costs, 0.1 unless the file says otherwise.
c=0.1; [ -f "$dir/cost" ] && c=$(cat "$dir/cost")
printf '{"type":"result","subtype":"success","is_error":false,"duration_ms":1000,"num_turns":3,"total_cost_usd":%s,"session_id":"stub","permission_denials":[],"modelUsage":{"%s":{"costUSD":%s}}}\n' "$c" "$MODEL" "$c"
STUB
  chmod +x "$t/stub/claude" || fail chmod
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
# newrepo <name>: a git repository on main where vloop init ran, with one
# [[check]], trusted in $HOME/.claude.json; $R is its path.
newrepo() {
  [ -x "$t/stub/claude" ] || stubs
  R="$t/$1"
  mkdir -p "$R/docs/briefs" && git -C "$R" init -q -b main || fail "git init $1"
  git -C "$R" config user.name gate && git -C "$R" config user.email gate@example.invalid || fail identity
  printf 'x\n' >"$R/docs/x.md"
  inr "$B" init </dev/null >"$t/$1.init" 2>&1 || fail "vloop init in $1: $(cat "$t/$1.init")"
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
# vrun <out>: vloop run on the brief in $R; output in $t/<out>, exit in $got.
vrun() { : >"$t/stub/argv.log"; inr "$B" run "$BRIEF" </dev/null >"$t/$1" 2>&1; got=$?; }
# mergemain: merge the current work branch into main with --no-ff, leaving
# main checked out; $MERGED is the merge commit's committer time, in seconds.
mergemain() {
  wb=$(git -C "$R" symbolic-ref --short HEAD) || fail "no branch checked out"
  [ "$wb" != main ] || fail "the run left main checked out, not a work branch"
  git -C "$R" checkout -q main && git -C "$R" merge -q --no-ff -m "merge $wb" "$wb" || fail "merge $wb"
  MERGED=$(git -C "$R" log -1 --format=%ct main) || fail "merge time"
}
# words <file>: the file's lines with runs of blanks squeezed and the ends trimmed.
squeeze() { sed -e 's/[[:space:]][[:space:]]*/ /g' -e 's/^ //' -e 's/ $//' "$1"; }
# stamp <epoch>: the local YYYYMMDD-HHMMSS of that second (BSD date).
stamp() { date -r "$1" +%Y%m%d-%H%M%S; }
# utc <epoch>: that second as created/recorded write it.
utc() { date -u -r "$1" +%Y-%m-%dT%H:%M:%SZ; }
# defectdoc <id> <brief name> <found-by> <created> [<extra frontmatter line>]:
# a defect/v1 record, origin work, as an operator's hand would write it.
defectdoc() {
  printf -- '---\nid: %s\nbrief: %s\norigin: work\nfound-by: %s\nkind: bug\nseverity: low\nstatus: open\nfixed-by: ""\ncase: ""\ncreated: %s\n%s---\nfixture defect %s\n' \
    "$1" "$2" "$3" "$4" "${5:+$5
}" "$1"
}
