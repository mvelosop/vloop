# Shared harness for B22's gates (the same file in every gate folder). The
# sourcing gate sets G, its own folder. Every repository, binary, stub and HOME
# a gate makes lives under $t, a temporary directory removed on exit. Nothing
# touches this repository's tree or refs, the network or the real HOME: runs
# happen in fixture repositories with the stub claude (stub-claude.sh) and the
# stub copilot (stub-copilot.sh) first on PATH. No gate runs the real copilot
# or gh.
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
BRIEF=docs/briefs/B20260101-0900-demo.loop-brief.md
BNAME=B20260101-0900-demo
# A Copilot model id for the fixtures, and the second reviewer's model, which
# no other session uses: the stub claude tells the second review by it.
MODEL=gpt-5.6-terra
SECOND=gate-second

# newbin: $S, vloop built from this tree.
newbin() {
  mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
  S=$t/bin/new
}
# venv cmd...: cmd in the one environment every fixture invocation gets.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GH_TOKEN=gate-gh-token GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
# fixture [plan file]: $t/repo afresh at the same path — git on main, `vloop
# init` by this tree's binary, one check, the harness brief, all committed —
# with fresh stubs (the plan they hand out is the plan file, plan.json by
# default) and a fake home trusting the repository. $CFG, when set, is lines of
# `vloop config set` arguments run before the commit (one key and value per
# line); it is spent here.
fixture() {
  pf=${1:-plan.json}
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
  pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
  cp "$G/$pf" "$t/stub/plan.json" && ln -s "$GITBIN" "$t/stub/git" && cp "$top/fence/review.json" "$t/stub/review-fence.json" &&
    sed "s/@PLUGIN_VERSION@/$pv/" "$G/stub-claude.sh" >"$t/stub/claude" && cp "$G/stub-copilot.sh" "$t/stub/copilot" &&
    chmod +x "$t/stub/claude" "$t/stub/copilot" || fail "stub"
  printf '%s\n' "$SECOND" >"$t/stub/second.model"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  venv git -C "$t/repo" init -q -b main && venv git -C "$t/repo" config user.name Gate &&
    venv git -C "$t/repo" config user.email gate@example.invalid && venv git -C "$t/repo" config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  (cd "$t/repo" && venv "$S" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  cp "$G/brief.md" "$t/repo/$BRIEF" || fail brief
  if [ -n "${CFG:-}" ]; then
    echo "$CFG" | while read -r k v; do
      [ -n "$k" ] || continue
      (cd "$t/repo" && venv "$S" config set "$k" "$v" >"$t/cfg.out" 2>&1) || fail "vloop config set $k $v: $(cat "$t/cfg.out")"
    done || exit 1
  fi
  CFG=
  venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "gate fixture" || fail "fixture commit"
}
# commitfix: commit whatever the gate added to the fixture.
commitfix() { venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "gate fixture change" || fail "fixture commit"; }
# runfix <out> [env...]: this tree's `vloop run` of the brief in the fixture
# with the extra environment; $rc is its exit code, <out> its stdout and
# stderr, <out>.ev the stubs' events, <out>.sessions the session records,
# <out>.runlog the run log, <out>.iters the iteration records, <out>.reports
# the kept reports.
runfix() {
  o=$1; shift
  (cd "$t/repo" && venv "$@" "$S" run "$BRIEF" >"$o" 2>&1); rc=$?
  if [ -f "$t/stub/events" ]; then mv "$t/stub/events" "$o.ev"; else : >"$o.ev"; fi
  rm -rf "$o.sessions" "$o.reports" && mkdir -p "$o.sessions" "$o.reports" && : >"$o.runlog" && : >"$o.iters"
  for d in "$t/repo"/.vloop/state/runs/*/*/; do
    [ -d "$d" ] || continue
    cp "$d"sessions/*.json "$o.sessions/" 2>/dev/null
    cp "$d"reports/* "$o.reports/" 2>/dev/null
    [ -f "$d/run.log" ] && cat "$d/run.log" >>"$o.runlog"
    [ -f "$d/iterations.jsonl" ] && cat "$d/iterations.jsonl" >>"$o.iters"
  done
  return 0
}
# phaserecs <out> <phase>: the session records of the phase, by file name order.
phaserecs() {
  for f in "$1.sessions"/*.json; do
    [ -f "$f" ] || continue
    [ "$("$JQ" -r .phase "$f")" = "$2" ] && echo "$f"
  done
  return 0
}
# ev <out> <pattern>: the stub event lines matching the extended regex.
ev() { grep -E -- "$2" "$1.ev"; }
# iterblock <runlog> <n>: the run log's lines of iteration n.
iterblock() {
  awk -v n="$2" '/^── iteration / { split($0, a, " "); inb = (a[3] == n); next } /^═══/ { inb = 0 } inb' "$1"
}
