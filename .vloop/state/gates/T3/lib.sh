# Shared harness for B29's command gates (the same file in T3 and T4). The
# sourcing gate sets G, its own folder. Every repository, binary, stub and HOME
# a gate makes lives under $t, a temporary directory removed on exit. Nothing
# touches this repository's tree or refs, the network or the real HOME: the
# command runs in fixture repositories with the stub claude (stub-claude.sh)
# and, where the gate has one, the stub copilot (stub-copilot.sh) first on
# PATH. No gate runs the real claude, copilot or gh.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p; s/^\(COPILOT_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT GOFLAGS GOBIN GH_TOKEN GITHUB_TOKEN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
SPID=
trap '[ -n "$SPID" ] && kill "$SPID" 2>/dev/null; chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
RID=B20990101-0000-x
B=docs/briefs/$RID.loop-brief.md
# The Copilot model id of the fixtures.
MODEL=gpt-5.6-terra
# The review-brief fence the brief pins: the gate-review fence of the base
# (gate-review.json in this folder) plus Bash(vloop brief check:*) allowed and
# Bash(vloop brief review:*) denied.
"$JQ" '.permissions.allow += ["Bash(vloop brief check:*)"] | .permissions.deny += ["Bash(vloop brief review:*)"]' "$G/gate-review.json" >"$t/rb-fence.json" || fail "jq"
# The reports the stubs hand out.
REPORT_FIX='{"schema":"brief-review/v1","brief":"B20990101-0000-x.loop-brief","verdict":"FIX","findings":[{"kind":"runnability","section":"Shape","summary":"the go check is red after task 1"},{"kind":"fact","section":"Binding references","summary":"internal/x.go has no func Y","evidence":"grep finds no Y","fix":"name the function that exists"}],"decisions":[],"checked":"every section","header":"**Review:** FIX — 2 findings"}'
REPORT_PASS='{"schema":"brief-review/v1","brief":"B20990101-0000-x.loop-brief","verdict":"PASS","findings":[]}'

# newbin: $S, vloop built from this tree.
newbin() {
  mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
  S=$t/bin/new
}
# venv cmd...: cmd in the one environment every fixture invocation gets, with
# a GitHub token for the Copilot login check. venvnl: the same with none.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GH_TOKEN=gate-gh-token GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
venvnl() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
# fixture: $t/repo afresh at the same path — git on main with a user, `vloop
# init` by this tree's binary, one check, the brief (status draft) and a
# notes.md, all committed — with fresh stubs (no report until a case sets one)
# and a fake home trusting the repository. $CFG, when set, is lines of key and
# value written to .vloop/config.toml with `vloop config set` before the
# commit; it is spent here.
fixture() {
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
  pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
  ln -s "$GITBIN" "$t/stub/git" && ln -s "$JQ" "$t/stub/jq" && cp "$t/rb-fence.json" "$t/stub/review-brief-fence.json" &&
    sed "s/@PLUGIN_VERSION@/$pv/" "$G/stub-claude.sh" >"$t/stub/claude" && chmod +x "$t/stub/claude" || fail "stub"
  if [ -f "$G/stub-copilot.sh" ]; then cp "$G/stub-copilot.sh" "$t/stub/copilot" && chmod +x "$t/stub/copilot" || fail "stub"; fi
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  venv git -C "$t/repo" init -q -b main && venv git -C "$t/repo" config user.name Gate &&
    venv git -C "$t/repo" config user.email gate@example.invalid && venv git -C "$t/repo" config commit.gpgsign false || fail "git init"
  echo "the fixture's notes" >"$t/repo/notes.md"
  (cd "$t/repo" && venv "$S" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >>"$t/repo/.vloop/config.toml"
  cp "$G/brief.md" "$t/repo/$B" || fail brief
  if [ -n "${CFG:-}" ]; then
    echo "$CFG" | while read -r k v; do
      [ -n "$k" ] || continue
      (cd "$t/repo" && venv "$S" config set "$k" "$v" >"$t/cfg.out" 2>&1) || fail "vloop config set $k $v: $(cat "$t/cfg.out")"
    done || exit 1
  fi
  CFG=
  venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "gate fixture" || fail "fixture commit"
  cp "$t/repo/$B" "$t/brief.orig"
}
# report <json>: the report the stubs hand out.
report() { printf '%s\n' "$1" >"$t/stub/report.json"; }
# refs: every ref and HEAD of the fixture.
refs() { venv git -C "$t/repo" for-each-ref --format='%(refname) %(objectname)'; venv git -C "$t/repo" rev-parse HEAD; venv git -C "$t/repo" symbolic-ref -q HEAD; }
# rev <out> [arg] [env...]: this tree's `vloop brief review <arg>` in the
# fixture (arg defaults to the brief); $rc its exit, <out>.out its stdout,
# <out>.err its stderr, <out>.ev the stubs' events; $F the review folder,
# repo-relative, when exactly one was made ("" otherwise). revnl: the same
# with no GitHub token.
rev() { o=$1; shift; a=${1:-$B}; [ $# -gt 0 ] && shift; refs >"$o.refs0"
  (cd "$t/repo" && venv "$@" "$S" brief review "$a" >"$o.out" 2>"$o.err"); rc=$?; collect "$o"; }
revnl() { o=$1; shift; a=${1:-$B}; [ $# -gt 0 ] && shift; refs >"$o.refs0"
  (cd "$t/repo" && venvnl "$@" "$S" brief review "$a" >"$o.out" 2>"$o.err"); rc=$?; collect "$o"; }
collect() {
  if [ -f "$t/stub/events" ]; then mv "$t/stub/events" "$1.ev"; else : >"$1.ev"; fi
  F=$(cd "$t/repo" && ls -d .vloop/state/brief-reviews/*/*/ 2>/dev/null)
  if [ "$(printf '%s' "$F" | grep -c .)" = 1 ]; then F=${F%/}; else F=; fi
  return 0
}
# ev <out> <pattern>: the stub event lines matching the extended regex.
ev() { grep -E -- "$2" "$1.ev"; }
# nosession <out> <what>: no stub session was started.
nosession() { [ -s "$1.ev" ] && fail "$2: a session was started: $(head -3 "$1.ev")"; return 0; }
# noreview <out> <what>: no session and no review folder.
noreview() { nosession "$1" "$2"; [ -e "$t/repo/.vloop/state/brief-reviews" ] && fail "$2: a review folder was made: $(cd "$t/repo" && find .vloop/state/brief-reviews)"; return 0; }
# folder <out> <what>: exactly one review folder, named by the brief's run id
# and a YYYYMMDD-HHMMSS stamp.
folder() {
  [ -n "$F" ] || fail "$2: want one folder .vloop/state/brief-reviews/$RID/<YYYYMMDD-HHMMSS>/, found: $(cd "$t/repo" && find .vloop/state/brief-reviews -maxdepth 2 2>/dev/null | tr '\n' ' ')"
  printf '%s\n' "$F" | grep -qxE "\.vloop/state/brief-reviews/$RID/[0-9]{8}-[0-9]{6}" || fail "$2: the review folder is $F, want .vloop/state/brief-reviews/$RID/<YYYYMMDD-HHMMSS>"
}
# errline <out> <line>: stderr holds the line (leading blanks allowed).
errline() { sed 's/^[[:space:]]*//' "$1.err" | grep -qxF -- "$2"; }
# lastline <out> <exact line> <what>: the run ended with that `vloop:` line on
# stderr, nothing on stdout.
endswith() {
  errline "$1" "$2" || fail "$3: want '$2' on stderr, got: $(cat "$1.err")"
  [ -s "$1.out" ] && fail "$3: want nothing on stdout, got: $(cat "$1.out")"
  return 0
}
# briefsame <what>: the brief's bytes are as committed.
briefsame() { cmp -s "$t/brief.orig" "$t/repo/$B" || fail "$1: the brief's bytes changed: $(diff "$t/brief.orig" "$t/repo/$B")"; }
# refsame <out> <what>: no ref moved.
refsame() { refs >"$1.refs1"; cmp -s "$1.refs0" "$1.refs1" || fail "$2: git refs moved: $(diff "$1.refs0" "$1.refs1")"; }
# block <verdict> <report file>: the verdict block the report must print.
block() {
  n=$("$JQ" '.findings | length' "$2")
  echo "verdict: $1 — $n finding(s)"
  "$JQ" -r '.findings[] | [.kind, .section, .summary] | @tsv' "$2" | while IFS="$(printf '\t')" read -r k s m; do
    printf '  %-12s  %s: %s\n' "$k" "$s" "$m"
  done
}
# record <what> <agent> <model>: the folder's one session record.
record() {
  ls "$t/repo/$F/sessions/" >"$t/recs" 2>/dev/null
  [ "$(cat "$t/recs")" = 001-review-brief.json ] || fail "$1: want the one record sessions/001-review-brief.json in $F, got: $(tr '\n' ' ' <"$t/recs")"
  r=$t/repo/$F/sessions/001-review-brief.json
  "$JQ" -e --arg a "$2" --arg m "$3" --arg id "$RID" '.schema == "session/v5" and .run_id == $id and .iteration == 0 and .phase == "review-brief" and .agent == $a and .model == $m' "$r" >/dev/null ||
    fail "$1: the record is $("$JQ" -c '{schema, run_id, iteration, phase, agent, model}' "$r"), want session/v5, run_id $RID, iteration 0, phase review-brief, agent $2, model $3"
  (cd "$t/repo" && venv "$S" schema validate session/v5 "$F/sessions/001-review-brief.json" >"$t/val" 2>&1) || fail "$1: the record is not valid session/v5: $(cat "$t/val")"
}
# skill <dir>: a project skill directory in the fixture.
skill() { mkdir -p "$t/repo/$1" && printf -- '---\nname: %s\ndescription: a project skill\n---\nBody.\n' "$(basename "$1")" >"$t/repo/$1/SKILL.md"; }
