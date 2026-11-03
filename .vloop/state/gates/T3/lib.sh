# Shared harness for B34's gates (the same file in T1, T2 and T3). The
# sourcing gate sets G, its own folder. Every repository, binary, stub and
# HOME a gate makes lives under $t, a temporary directory removed on exit.
# Nothing touches this repository's tree or refs, the network or the real
# HOME: runs happen in fixture repositories with the stub copilot
# (stub-copilot.sh) and stub claude (stub-claude.sh) first on PATH, every
# phase on Copilot. No gate runs the real copilot or gh.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p; s/^\(COPILOT_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT GOFLAGS GOBIN GH_TOKEN GITHUB_TOKEN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
top=$(pwd -P)
[ -f "$top/go.mod" ] && [ -d "$top/cmd/vloop" ] || fail "run from the repository root"
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
BRIEF=docs/briefs/B20260101-0900-demo.loop-brief.md
# The Copilot model id of the fixtures.
MODEL=gpt-5.6-terra
H=$t/gateuser

# The fences as they stood when this plan was written: no task changes them.
for f in "$G"/fence/*.json; do
  cmp -s "$f" "$top/fence/$(basename "$f")" || fail "fence/$(basename "$f") differs from the fence this plan was written against: no task may change a fence's rules"
done

# newbin: $S, vloop built from this tree into $t.
newbin() {
  mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/vloop" ./cmd/vloop >"$t/build.out" 2>&1 ||
    fail "go build ./cmd/vloop: $(tail -n 5 "$t/build.out")"
  S=$t/bin/vloop
}
# venv cmd...: cmd in the one environment every fixture invocation gets.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$H" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GH_TOKEN=gate-gh-token GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
# fixture: $t/repo afresh — git on main, `vloop init` by this tree's binary,
# one check, the brief, every phase on copilot with $MODEL, all committed —
# with fresh stubs (the plan they hand out has one task, T1, whose gate is
# `test -f T1.out`) and a fake home trusting the repository. $CFG, when set,
# is more `vloop config set` lines (key and value); it is spent here.
fixture() {
  chmod -R u+w "$t/repo" "$H" 2>/dev/null; rm -rf "$t/repo" "$H" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$H" "$t/tmp" "$t/stub" || fail mkdir
  pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
  cp "$G/fixture-plan.json" "$t/stub/plan.json" && ln -s "$GITBIN" "$t/stub/git" && ln -s "$JQ" "$t/stub/jq" &&
    sed "s/@PLUGIN_VERSION@/$pv/" "$G/stub-claude.sh" >"$t/stub/claude" && cp "$G/stub-copilot.sh" "$t/stub/copilot" &&
    chmod +x "$t/stub/claude" "$t/stub/copilot" || fail "stub"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$H/.claude.json"
  venv git -C "$t/repo" init -q -b main && venv git -C "$t/repo" config user.name Gate &&
    venv git -C "$t/repo" config user.email gate@example.invalid && venv git -C "$t/repo" config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  (cd "$t/repo" && venv "$S" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  cp "$G/brief.md" "$t/repo/$BRIEF" || fail brief
  all="agent.plan copilot
model.plan $MODEL
agent.gate-review copilot
model.gate-review $MODEL
agent.work copilot
model.work $MODEL
agent.review copilot
model.review $MODEL
agent.second-review copilot
model.second-review $MODEL
agent.verify copilot
model.verify $MODEL
agent.review-brief copilot
model.review-brief $MODEL
${CFG:-}"
  echo "$all" | while read -r k v; do
    [ -n "$k" ] || continue
    (cd "$t/repo" && venv "$S" config set "$k" "$v" >"$t/cfg.out" 2>&1) || fail "vloop config set $k $v: $(cat "$t/cfg.out")"
  done || exit 1
  CFG=
  venv git -C "$t/repo" add -A && venv git -C "$t/repo" commit -q -m "gate fixture" || fail "fixture commit"
  BASEREV=$(venv git -C "$t/repo" rev-parse HEAD)
}
# runfix <name> [binary]: `vloop run` of the brief in the fixture, by this
# tree's binary or the one given; $rc is its exit code, $t/<name>.out and
# .err its stdout and stderr, $t/<name>.ev the stub's events.
runfix() {
  n=$1; b=${2:-$S}
  (cd "$t/repo" && venv "$b" run "$BRIEF" >"$t/$n.out" 2>"$t/$n.err"); rc=$?
  if [ -f "$t/stub/events" ]; then mv "$t/stub/events" "$t/$n.ev"; else : >"$t/$n.ev"; fi
}
# brev <name>: `vloop brief review` of the brief in the fixture, as runfix.
brev() {
  n=$1
  (cd "$t/repo" && venv "$S" brief review "$BRIEF" >"$t/$n.out" 2>"$t/$n.err"); rc=$?
  if [ -f "$t/stub/events" ]; then mv "$t/stub/events" "$t/$n.ev"; else : >"$t/$n.ev"; fi
}
# Tool events as Copilot prints them. start <id> <tool> <arguments JSON>;
# done <id> (completed, no error); denied <id> (denied by the hook).
start() { "$JQ" -cn --arg i "$1" --arg n "$2" --argjson a "$3" '{type: "tool.execution_start", data: {toolCallId: $i, toolName: $n, arguments: $a}}'; }
done_() { "$JQ" -cn --arg i "$1" '{type: "tool.execution_complete", data: {toolCallId: $i, success: true, result: {content: "ok"}}}'; }
denied() { "$JQ" -cn --arg i "$1" '{type: "tool.execution_complete", data: {toolCallId: $i, success: false, error: {code: "denied", message: "Denied by preToolUse hook: vloop fence: no"}}}'; }
bashargs() { "$JQ" -cn --arg c "$1" '{command: $c, description: "d"}'; }
fileargs() { "$JQ" -cn --arg p "$1" '{path: $p, file_text: "x"}'; }
# reason <phase> <tool> [--path <p>] [command]: the reason `vloop fence check`
# gives for the call, which must be a deny.
reason() {
  p=$1; shift
  o=$(cd "$t/repo" && venv "$S" fence check --phase "$p" --root "$t/repo" --tool "$@" 2>&1)
  case $o in
    "deny: "*) printf '%s\n' "${o#deny: }";;
    *) fail "vloop fence check --phase $p --tool $*: '$o', want a deny";;
  esac
}
# mask <text>: the text with the fixture's HOME as ~.
mask() { printf '%s\n' "$1" | awk -v h="$H" '{ while ((i = index($0, h)) > 0) $0 = substr($0, 1, i - 1) "~" substr($0, i + length(h)); print }'; }
# errline <name> <line>: stderr has the exact line.
errline() { grep -qxF -- "$2" "$t/$1.err"; }
# halted <name> <message>: stderr has the message as a line, alone or after "vloop: ".
halted() { awk -v m="$2" '$0 == m || $0 == "vloop: " m { f = 1 } END { exit !f }' "$t/$1.err"; }
# firstof <name> <ERE>: the first stderr line matching the pattern.
firstof() { grep -E -- "$2" "$t/$1.err" | head -n 1; }
# runlog: the fixture's run log (empty when there is none).
runlog() { for f in "$t/repo"/.vloop/state/runs/*/*/run.log; do [ -f "$f" ] && cat "$f"; done; return 0; }
# commits_with <path>: how many commits, on any ref, carry the path.
commits_with() { venv git -C "$t/repo" log --all --format=%H -- "$1" | wc -l | tr -d ' '; }
# show <name>: the run's exit code and the tail of its stderr, for a failure message.
show() { printf 'exit %s; stderr:\n%s' "$rc" "$(tail -n 12 "$t/$1.err")"; }
