# Shared harness for B19's gates (the same file in every gate folder). The
# sourcing gate sets G, its own folder. Every repository, binary, stub and HOME
# a gate makes lives under $t, a temporary directory removed on exit. Nothing
# touches this repository's tree or refs, the network or the real HOME: runs
# happen in fixture repositories with the stub claude (stub-claude.sh) and the
# stub copilot (stub-copilot.sh) first on PATH. No gate runs the real copilot.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p; s/^\(COPILOT_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT GOFLAGS GOBIN GH_TOKEN GITHUB_TOKEN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
# The plan's base: the commit HEAD was at while this plan was written.
BASE=c2ef299de4890eb8c5b119a62753b170de163a17
BRIEF=docs/briefs/B20260101-0900-demo.loop-brief.md
BNAME=B20260101-0900-demo.loop-brief
MODEL=gpt-5.6-terra
# The stub copilot's credits (session.usage_checkpoint totalNanoAiu / 1e9 of
# its last checkpoint) and the default credit price the brief pins.
CREDITS=14.2
PRICE=0.0066667
# The environment a run gets beyond venv's: what a session must receive
# (an Anthropic setting, a GitHub token, the user's own COPILOT_HOME, which a
# copilot session must not get) and what every session drops.
CARRY="ANTHROPIC_DEFAULT_SONNET_MODEL=gate-sonnet GH_TOKEN=gate-gh-token-value COPILOT_HOME=$t/home/.copilot VLOOP_ACTIVE_TASK=leaked VLOOP_GATE_TASK=leaked VLOOP_PLAN_SHA256=leaked"
LESSON="- A lesson of this repository, for the gate."

# newbin: $t/bin/new, vloop built from this tree.
newbin() {
  mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
}
# basebin: $t/bin/base, vloop built from $BASE with git archive (never a worktree).
basebin() {
  rm -rf "$t/src-base" && mkdir -p "$t/src-base" "$t/bin" &&
    git -C "$top" archive "$BASE" | tar -x -C "$t/src-base" || fail "git archive $BASE"
  (cd "$t/src-base" && go build -buildvcs=false -o "$t/bin/base" ./cmd/vloop) || fail "build of the base"
}
# venv cmd...: cmd in the one environment every fixture invocation gets.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
# fixture [plan file] [guidelines]: $t/repo afresh at the same path — git on
# main, `vloop init` by this tree's binary, one check, the harness brief, all
# committed — with fresh stubs (claude and copilot; the plan they hand out is
# the plan file, plan.json by default) and a fake home trusting the
# repository. With "guidelines" a .vloop/repo-guidelines.md holding $LESSON is
# committed. $CFG, when set, is lines of `vloop config set` arguments run
# before the commit (one key and value per line).
fixture() {
  pf=${1:-plan.json}
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
  pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
  cp "$G/$pf" "$t/stub/plan.json" && ln -s "$GITBIN" "$t/stub/git" &&
    sed "s/@PLUGIN_VERSION@/$pv/" "$G/stub-claude.sh" >"$t/stub/claude" && cp "$G/stub-copilot.sh" "$t/stub/copilot" &&
    chmod +x "$t/stub/claude" "$t/stub/copilot" || fail "stub"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  venv git -C "$t/repo" init -q -b main && venv git -C "$t/repo" config user.name Gate &&
    venv git -C "$t/repo" config user.email gate@example.invalid && venv git -C "$t/repo" config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  (cd "$t/repo" && venv "$t/bin/new" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  cp "$G/brief.md" "$t/repo/$BRIEF" || fail brief
  rm -f "$t/repo/.vloop/repo-guidelines.md"
  if [ "${2:-}" = guidelines ]; then printf -- '%s\n' "$LESSON" >"$t/repo/.vloop/repo-guidelines.md"; fi
  if [ -n "${CFG:-}" ]; then
    echo "$CFG" | while read -r k v; do
      [ -n "$k" ] || continue
      (cd "$t/repo" && venv "$t/bin/new" config set "$k" "$v" >"$t/cfg.out" 2>&1) || fail "vloop config set $k $v: $(cat "$t/cfg.out")"
    done || exit 1
  fi
  # A prefix assignment to a function call outlives the call in some shells:
  # CFG is spent here so the next fixture starts without it.
  CFG=
  venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "gate fixture" || fail "fixture commit"
}
# The configuration of a run with copilot reviews.
COPILOT_CFG="agent.review copilot
model.review $MODEL"
# runfix <out> [env...]: this tree's `vloop run` of the brief in the fixture
# with the extra environment; $rc is its exit code, <out> its stdout and
# stderr, <out>.log the stubs' log (normlog'd), <out>.sessions the records,
# <out>.runlog the run log, <out>.metrics the committed metrics snapshot.
runfix() {
  o=$1; shift
  (cd "$t/repo" && venv "$@" "$t/bin/new" run "$BRIEF" >"$o" 2>&1); rc=$?
  if [ -f "$t/stub/log" ]; then normlog <"$t/stub/log" >"$o.log"; else : >"$o.log"; fi
  rm -f "$t/stub/log"
  rm -rf "$o.sessions" && mkdir -p "$o.sessions" && : >"$o.runlog"
  for d in "$t/repo"/.vloop/state/runs/*/*/; do
    [ -d "$d" ] || continue
    cp "$d"sessions/*.json "$o.sessions/" 2>/dev/null
    [ -f "$d/run.log" ] && cat "$d/run.log" >>"$o.runlog"
  done
  rm -f "$o.metrics"; cp "$t/repo/.vloop/state/metrics/B20260101-0900-demo.json" "$o.metrics" 2>/dev/null
  return 0
}
# normlog: VLOOP_PLAN_SHA256, the SHA-256 of a plan that carries the driver's
# own timestamps, normalized: each distinct 64-hex value becomes H<n> by order
# of first appearance.
normlog() {
  awk '/^env VLOOP_PLAN_SHA256=[0-9a-f]+$/ { v = substr($0, 23); if (length(v) == 64) { if (!(v in h)) h[v] = "H" (++n); $0 = "env VLOOP_PLAN_SHA256=" h[v] } } { print }'
}
# block <log> <agent> <prompt>: the log's entry for the <agent> session started
# with -p <prompt> (the first such).
block() {
  awk -v a="=== $2" -v p="$3" '/^=== / { if (keep) exit; inb = ($0 == a); buf = $0 "\n"; next }
    inb { buf = buf $0 "\n"; if ($0 == "arg " p) keep = 1 }
    END { if (keep) printf "%s", buf }' "$1"
}
# sessions <log> <agent>: the -p prompts of the agent's sessions, in order.
sessions() {
  awk -v a="=== $2" '/^=== / { inb = ($0 == a); prev = ""; next }
    inb && prev == "arg -p" && /^arg / { print substr($0, 5) } { prev = $0 }' "$1"
}
# argafter <block> <flag>: the argument right after the first <flag>.
argafter() { echo "$1" | awk -v f="arg $2" 'take { print substr($0, 5); exit } $0 == f { take = 1 }'; }
# envof <block>: the block's environment lines.
envof() { echo "$1" | grep '^env ' | LC_ALL=C sort; }
# near <a> <b>: a and b, numbers, within 1e-9 of each other.
near() { awk -v a="$1" -v b="$2" 'BEGIN { d = a - b; if (d < 0) d = -d; exit !(d < 1e-9) }'; }
