# Shared harness for B30's run gates. The sourcing gate sets G, its own
# folder. Every repository, binary, stub and HOME a gate makes lives under $t,
# a temporary directory removed on exit. Nothing touches this repository's
# tree or refs, the network or the real HOME: runs happen in fixture
# repositories with the stub claude (stub-claude.sh) first on PATH. No gate
# runs the real claude, copilot or gh.
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
RID=B20260101-0900-demo
BRIEF=docs/briefs/$RID.loop-brief.md

# newbin: $S, vloop built from this tree.
newbin() {
  mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
  S=$t/bin/new
}
# venv cmd...: cmd in the one environment every fixture invocation gets.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
# fixture: $t/repo afresh at the same path — git on main, `vloop init` by this
# tree's binary, one check, the brief (brief.md), all committed — with fresh
# stubs handing out plan.json (one task, T1, gated on T1.out) and no verify
# report, and a fake home trusting the repository.
fixture() {
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
  pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
  cp "$G/plan.json" "$t/stub/plan.json" && ln -s "$GITBIN" "$t/stub/git" && ln -s "$JQ" "$t/stub/jq" &&
    sed "s/@PLUGIN_VERSION@/$pv/" "$G/stub-claude.sh" >"$t/stub/claude" && chmod +x "$t/stub/claude" || fail "stub"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  venv git -C "$t/repo" init -q -b main && venv git -C "$t/repo" config user.name Gate &&
    venv git -C "$t/repo" config user.email gate@example.invalid && venv git -C "$t/repo" config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  echo "the fixture's readme" >"$t/repo/README.md"
  (cd "$t/repo" && venv "$S" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  cp "$G/brief.md" "$t/repo/$BRIEF" || fail brief
  venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "gate fixture" || fail "fixture commit"
}
# runcase <name> <report json|-> [stub flag file...]: a fresh fixture whose
# stub verify session hands out the report (none for -) and does what the
# flag files say (see stub-claude.sh), then this tree's `vloop run` of the
# brief. $rc is its exit code; $t/<name>.out its stdout, $t/<name>.err its
# stderr, $t/<name>.ev the stub's events, $t/<name>.report the report the stub
# handed out; $RD the run folder, repo-relative with its trailing slash.
runcase() {
  n=$1; rep=$2; shift 2
  fixture
  rm -f "$t/$n.report"
  if [ "$rep" != - ]; then printf '%s\n' "$rep" >"$t/stub/verify-report.json" && cp "$t/stub/verify-report.json" "$t/$n.report"; fi
  for f in "$@"; do : >"$t/stub/$f"; done
  (cd "$t/repo" && venv "$S" run "$BRIEF" >"$t/$n.out" 2>"$t/$n.err"); rc=$?
  if [ -f "$t/stub/events" ]; then mv "$t/stub/events" "$t/$n.ev"; else : >"$t/$n.ev"; fi
  RD=$(cd "$t/repo" && ls -d .vloop/state/runs/*/*/ 2>/dev/null)
  [ "$(printf '%s\n' "$RD" | grep -c .)" = 1 ] || fail "$n: want one run folder, found: $RD"
  return 0
}
# completed <name>: the run exited 0, its plan is complete, T1 done, and one
# verify session ran — whatever the verify session reported.
completed() {
  [ "$rc" = 0 ] || fail "$1: vloop run exited $rc, want 0: $(tail -8 "$t/$1.err")"
  st=$("$JQ" -r '.status + " " + ([.tasks[].status] | join(","))' "$t/repo/.vloop/state/state.json" 2>/dev/null)
  [ "$st" = "complete done" ] || fail "$1: the plan reads '$st', want 'complete done': the verify report must not change the run's ending or a task"
  [ "$(grep -c '^session verify ' "$t/$1.ev")" = 1 ] || fail "$1: want one verify session, got: $(grep '^session ' "$t/$1.ev" | tr '\n' ';')"
}
# outis <name> <exact line>: stdout holds the line.
outis() { grep -qxF -- "$2" "$t/$1.out" || fail "$1: want '$2' on stdout, got: $(grep -F 'verify' "$t/$1.out")"; }
# erris <name> <exact line>: stderr holds the line.
erris() { grep -qxF -- "$2" "$t/$1.err" || fail "$1: want '$2' on stderr, got: $(grep -F 'verify' "$t/$1.err")"; }
# nowhere <name> <fixed text>: neither stream holds the text.
nowhere() { cat "$t/$1.out" "$t/$1.err" | grep -qF -- "$2" && fail "$1: the output says '$2' and must not: $(cat "$t/$1.out" "$t/$1.err" | grep -F -- "$2")"; return 0; }
# verifylines <name> <n>: stdout holds exactly n lines starting "verify: ".
verifylines() {
  c=$(grep -c '^verify: ' "$t/$1.out")
  [ "$c" = "$2" ] || fail "$1: want $2 'verify: ' line(s) on stdout, got $c: $(grep '^verify: ' "$t/$1.out")"
}
# kept <name>: the run folder keeps the stub's report as reports/verify.json,
# the same JSON, ending with a newline.
kept() {
  k=$t/repo/${RD}reports/verify.json
  [ -f "$k" ] || fail "$1: no ${RD}reports/verify.json kept"
  [ "$("$JQ" -S . "$k" 2>/dev/null)" = "$("$JQ" -S . "$t/$1.report")" ] || fail "$1: the kept report is not the session's: $(cat "$k")"
  [ "$(tail -c 1 "$k" | od -An -c | tr -d ' ')" = '\n' ] || fail "$1: the kept report does not end with a newline"
}
# notkept <name>: no report kept.
notkept() { [ -e "$t/repo/${RD}reports/verify.json" ] && fail "$1: a report was kept as ${RD}reports/verify.json: $(cat "$t/repo/${RD}reports/verify.json")"; return 0; }
# closebare <name>: this tree's `vloop brief close <brief>`, no flag, in the
# fixture as the run left it; $crc its exit, $t/<name>.cout and .cerr.
closebare() { (cd "$t/repo" && venv "$S" brief close "$BRIEF" >"$t/$1.cout" 2>"$t/$1.cerr"); crc=$?; }
