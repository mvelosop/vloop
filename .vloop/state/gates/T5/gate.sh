#!/bin/sh
# Gate T5 — a denied Bash call keeps its command (brief R4): masked, redacted,
# cut to 200 characters with … appended; other tools keep only their file_path;
# the dashed -Users-<user>- form is masked in records and stderr captures.
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

# HOME is $t/Users/gateuser, so the user name the driver masks is gateuser.
U=$(basename "$HOME")
TOKEN=gate-s3cret-value-0123456789
runrepo d
cat >"$S/script.sh" <<'EOF'
if [ "$PHASE" = work ]; then
  STUB_DENIALS=$(jq -cn --arg h "$HOME" --arg u "$(basename "$HOME")" --arg tok "$GATE_API_TOKEN" '[
    {tool_name: "Bash", tool_input: {command: "rm -rf $X/tmp", description: "clean up"}},
    {tool_name: "Bash", tool_input: {command: ("cat " + $h + "/notes && curl -H " + $tok + " x")}},
    {tool_name: "Bash", tool_input: {command: ($h + "/" + ([range(98)] | map("é") | join("")) + ([range(200)] | map("a") | join("")))}},
    {tool_name: "Read", tool_input: {file_path: ($h + "/.claude/projects/-Users-" + $u + "-repo/x")}},
    {tool_name: "Write", tool_input: {file_path: "notes.txt", content: "keep none of this"}},
    {tool_name: "WebFetch", tool_input: {url: "https://example.invalid/", prompt: "keep none of this"}}
  ]')
  echo "trace -Users-$(basename "$HOME")-repo and -home-$(basename "$HOME")-x" >&2
fi
EOF
vrun d.out GATE_API_TOKEN="$TOKEN"; got=$?
[ "$got" = 0 ] || fail "the stubbed run exited $got, want 0 — $(tail -5 "$t/d.out.err")"
W=$(records '*-work.json')
[ "$(echo "$W" | grep -c .)" = 1 ] || fail "want one work session record, found: $W"
# The 300-character command, masked: ~/ and 298 more; kept as its first 200 characters and …
jq -n --arg u "$U" '[
  {tool_name: "Bash", command: "rm -rf $X/tmp"},
  {tool_name: "Bash", command: "cat ~/notes && curl -H <redacted:GATE_API_TOKEN> x"},
  {tool_name: "Bash", command: (("~/" + ([range(98)] | map("é") | join("")) + ([range(200)] | map("a") | join("")))[0:200] + "…")},
  {tool_name: "Read", file_path: "~/.claude/projects/-Users-USER-repo/x"},
  {tool_name: "Write", file_path: "notes.txt"},
  {tool_name: "WebFetch"}
]' >"$t/want.json" || fail "want"
jq -e --slurpfile w "$t/want.json" '.permission_denials == $w[0]' "$W" >/dev/null || {
  jq -c '.permission_denials[]' "$W" >"$t/got.l"; jq -c '.[]' "$t/want.json" >"$t/want.l"
  fail "the work record's permission_denials differ (< got, > want): $(diff "$t/got.l" "$t/want.l" | grep '^[<>]' | cut -c1-160)"
}
jq -e '.permission_denials[2].command | length == 201' "$W" >/dev/null || fail "the long command is not 200 characters and …"
grep -rq "$U" "$R/.vloop/state/runs/$N" && fail "the user name $U is left in the run folder: $(grep -rl "$U" "$R/.vloop/state/runs/$N" | sed "s|$R/||")"
grep -rq "$TOKEN" "$R/.vloop/state/runs/$N" && fail "the secret is left in the run folder"
E=$(ls "$R"/.vloop/state/runs/"$N"/*/work-*.stderr 2>/dev/null | head -1)
[ -n "$E" ] || fail "no work stderr capture"
grep -Fq "trace -Users-USER-repo and -home-USER-x" "$E" || fail "the stderr capture does not mask the dashed form: $(cat "$E")"
echo "gate T5: ok"
