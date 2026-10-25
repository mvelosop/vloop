# Shared harness for B18's gates (the same file in every gate folder). The
# sourcing gate sets G, its own folder. Every repository, binary, stub and HOME
# a gate makes lives under $t, a temporary directory removed on exit. Nothing
# touches this repository's tree or refs, the network or the real HOME: runs
# happen in fixture repositories with the stub claude of stub.sh first on PATH.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT GOFLAGS GOBIN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
# The plan's base: the commit HEAD was at while this plan was written.
BASE=8aa98974a124f45a833844dafb9603d404a858b5
BRIEF=docs/briefs/B20260101-0900-demo.loop-brief.md
BNAME=B20260101-0900-demo.loop-brief
PHASES="plan gate-review work review verify"
# The environment a run gets beyond venv's: variables the base passes through
# to claude, and the ones it drops (the plan hash is set again by the driver
# for the sessions that have one).
CARRY="CLAUDE_CODE_USE_FOUNDRY=1 ANTHROPIC_FOUNDRY_RESOURCE=gate-resource ANTHROPIC_DEFAULT_SONNET_MODEL=gate-sonnet VLOOP_ACTIVE_TASK=leaked VLOOP_GATE_TASK=leaked VLOOP_PLAN_SHA256=leaked"
EFFORTS="VLOOP_EFFORT_PLAN=high VLOOP_EFFORT_WORK=high VLOOP_EFFORT_REVIEW=high VLOOP_EFFORT_GATE_REVIEW=high VLOOP_EFFORT_VERIFY=high"

# newbin: $t/bin/new, vloop built from this tree.
newbin() {
  mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
}
# basebin <name> [sed expression]: $t/bin/<name>, vloop built from $BASE with
# git archive (never a worktree), with the base's internal/driver/session.go
# edited by the expression when one is given.
basebin() {
  rm -rf "$t/src-$1" && mkdir -p "$t/src-$1" "$t/bin" &&
    git -C "$top" archive "$BASE" | tar -x -C "$t/src-$1" || fail "git archive $BASE"
  if [ $# -gt 1 ]; then
    f="$t/src-$1/internal/driver/session.go"
    sed "$2" "$f" >"$t/s.go" && ! cmp -s "$f" "$t/s.go" && mv "$t/s.go" "$f" || fail "planting into the base copy changed nothing"
  fi
  (cd "$t/src-$1" && go build -buildvcs=false -o "$t/bin/$1" ./cmd/vloop) || fail "build of the base ($1)"
}
# venv cmd...: cmd in the one environment every fixture invocation gets.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
# fixture <init binary> [guidelines]: $t/repo afresh at the same path — git on
# main, `vloop init` by the given binary, one check, the harness brief, all
# committed — with a fresh stub and a fake home trusting the repository. With
# "guidelines" a non-empty .vloop/repo-guidelines.md is committed; without, the
# fixture has none.
fixture() {
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
  cp "$G/plan.json" "$t/stub/plan.json" && ln -s "$GITBIN" "$t/stub/git" &&
    sed "s/@PLUGIN_VERSION@/$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json")/" "$G/stub.sh" >"$t/stub/claude" &&
    chmod +x "$t/stub/claude" || fail "stub"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  venv git -C "$t/repo" init -q -b main && venv git -C "$t/repo" config user.name Gate &&
    venv git -C "$t/repo" config user.email gate@example.invalid && venv git -C "$t/repo" config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  (cd "$t/repo" && venv "$1" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  cp "$G/brief.md" "$t/repo/$BRIEF" || fail brief
  rm -f "$t/repo/.vloop/repo-guidelines.md"
  if [ "${2:-}" = guidelines ]; then printf -- '- A lesson of this repository.\n' >"$t/repo/.vloop/repo-guidelines.md"; fi
  venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "gate fixture" || fail "fixture commit"
}
# runfix <binary> <out> [env...]: `vloop run` of the brief in the fixture with
# the extra environment; $rc is its exit code, <out> its stdout and stderr
# with the run folder's stamp and the durations normalized, <out>.log the
# stub's log (normlog'd), <out>.sessions the records, <out>.stderr the
# sessions' captured stderr, <out>.runlog the run log (normalized).
runfix() {
  b=$1 o=$2; shift 2
  (cd "$t/repo" && venv "$@" "$b" run "$BRIEF" >"$o.raw" 2>&1); rc=$?
  normout <"$o.raw" >"$o"
  if [ -f "$t/stub/log" ]; then normlog <"$t/stub/log" >"$o.log"; else : >"$o.log"; fi
  rm -f "$t/stub/log"
  rm -rf "$o.sessions" "$o.stderr" && mkdir -p "$o.sessions" "$o.stderr"
  for d in "$t/repo"/.vloop/state/runs/*/*/; do
    [ -d "$d" ] || continue
    cp "$d"sessions/*.json "$o.sessions/" 2>/dev/null
    cp "$d"*.stderr "$o.stderr/" 2>/dev/null
    [ -f "$d/run.log" ] && normout <"$d/run.log" >"$o.runlog"
  done
  return 0
}
# normout: the run folder's stamp and session durations, which differ between
# any two runs, replaced by placeholders.
normout() { sed -E 's/[0-9]{8}-[0-9]{6}/STAMP/g; s/[0-9]+m[0-9]{2}s/DUR/g'; }
# normlog: the one value of the stub's log normalized — VLOOP_PLAN_SHA256, the
# SHA-256 of a plan that carries the driver's own timestamps, so it differs
# between any two runs. Each distinct 64-hex value becomes H<n> by order of
# first appearance, which keeps which sessions got a hash and which got the
# same one; a value that is not 64 hex digits is kept as it is.
normlog() {
  awk '/^env VLOOP_PLAN_SHA256=[0-9a-f]+$/ { v = substr($0, 23); if (length(v) == 64) { if (!(v in h)) h[v] = "H" (++n); $0 = "env VLOOP_PLAN_SHA256=" h[v] } } { print }'
}
# sessionblock <log> <prompt>: the log's entry for the session started with -p <prompt>.
sessionblock() {
  awk -v p="$2" '/^=== claude$/ { if (keep) exit; buf = $0 "\n"; inb = 1; next }
    inb { buf = buf $0 "\n"; if ($0 == "arg " p) keep = 1 }
    END { if (keep) printf "%s", buf }' "$1"
}
# compare <name> [guidelines] [env...]: one run with the base binary and one
# with this tree's, each in a fresh fixture at the same path, with the same
# environment and the same stub modes ($MODES: words touched in the stub's
# folder after the fixture is made, name=value writing value into name). Their
# stub logs (argv, working directory and environment of every claude
# invocation), exit codes and output up to the closing summary must be the
# same. $brc is the base's exit code; the changed run's files are
# $t/<name>.new*, the base's $t/<name>.base*.
compare() {
  n=$1; shift
  gl=""; if [ "${1:-}" = guidelines ]; then gl=guidelines; shift; fi
  for side in base new; do
    fixture "$t/bin/base" $gl
    for m in ${MODES:-}; do
      case "$m" in *=*) printf '%s\n' "${m#*=}" >"$t/stub/${m%%=*}";; *) : >"$t/stub/$m";; esac
    done
    runfix "$t/bin/$side" "$t/$n.$side" $CARRY "$@"
    eval "${side}rc=\$rc"
  done
  brc=$baserc
  [ "$newrc" = "$brc" ] || fail "$n: vloop run exited $newrc, the base's $brc: $(tail -5 "$t/$n.new")"
  diff "$t/$n.base.log" "$t/$n.new.log" >"$t/$n.diff" ||
    fail "$n: what claude received differs from the base's (< base, > this tree):
$(head -40 "$t/$n.diff")"
  prelude "$t/$n.base" >"$t/$n.base.pre"; prelude "$t/$n.new" >"$t/$n.new.pre"
  diff "$t/$n.base.pre" "$t/$n.new.pre" >"$t/$n.prediff" ||
    fail "$n: vloop run's output differs from the base's (< base, > this tree):
$(head -20 "$t/$n.prediff")"
}
# allphases <name>: the changed run of compare <name> completed and started
# one session of every phase.
allphases() {
  [ "$brc" = 0 ] || fail "$1: the run exited $brc: $(tail -5 "$t/$1.base")"
  for p in $PHASES; do
    grep -q "^arg /vloop:$p" "$t/$1.new.log" || fail "$1: no $p session was started"
  done
}
# prelude <out>: the run's output up to the metrics summary at its end — the
# part that is the same, line for line, as the base's.
prelude() { awk '/^B20260101-0900-demo  / { exit } { print }' "$1"; }
