#!/bin/sh
# Gate T4 — R4: Copilot's login, checked, and the experimental warning's new
# text. In a scratch vloop repository with agent.review = copilot, a
# model.review set, a stub copilot and a stub gh first on PATH, and an
# environment holding no token variable (env -i):
#   - gh exits 1: doctor's `copilot login` row is ✗ with the brief's message,
#     doctor exits 1; `vloop run` exits 1 with the brief's line on stderr
#     before any session (the stub claude sees no -p) and leaves no plan;
#   - an empty GH_TOKEN, or a gh that exits 0 printing nothing: still ✗;
#   - gh prints a token, or COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN is
#     set with the failing gh: ✓ copilot login;
#   - no token value appears in doctor's output (text or JSON), in the run's
#     output or anywhere under the fixture's .vloop/;
#   - doctor and the run preflight give the warning's new text;
#   - with agent.review = claude, doctor has no copilot login row.
# Then the text: the old warning wording is gone from what a session or the
# operator reads, and docs/guide/configuration.md's Copilot paragraph says
# "experimental until the fence lands" and how a session logs in.
# No gate runs the real copilot, gh or the network.
set -u
H=.vloop/state/gates/T4
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GOFLAGS GOBIN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/new" ./cmd/vloop || fail "go build of this tree"
pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
BRIEF=docs/briefs/B20260101-0900-demo.loop-brief.md
TOKEN=gho_GATEstubTOKEN0123456789abcdef
ENVTOKEN=gate-env-token-9876543210
MSG='no GitHub login reaches Copilot sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login'
RUNLINE='vloop: agent.review is copilot but no GitHub login reaches its sessions: set COPILOT_GITHUB_TOKEN, GH_TOKEN or GITHUB_TOKEN, or run gh auth login'
WARN='! agent.review is copilot: experimental — review sessions are not fenced yet (docs/guide/configuration.md)'

# The stubs: claude, copilot (exits 0), git, and gh, set by gh <exit> [<printed>].
mkdir -p "$t/stub" "$t/home" "$t/tmp" "$t/repo/docs/briefs" || fail mkdir
ln -s "$GITBIN" "$t/stub/git" && sed "s/@PLUGIN_VERSION@/$pv/" "$H/stub-claude.sh" >"$t/stub/claude" && chmod +x "$t/stub/claude" || fail stub
printf '#!/bin/sh\necho "copilot 1.0.0"\nexit 0\n' >"$t/stub/copilot" && chmod +x "$t/stub/copilot" || fail stub
gh() { printf '#!/bin/sh\n[ -n "%s" ] && echo "%s"\nexit %s\n' "$2" "$2" "$1" >"$t/stub/gh" && chmod +x "$t/stub/gh" || fail "stub gh"; }
printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
# venv [VAR=value...] cmd...: cmd in an environment with no token variable.
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
R="$t/repo"
venv git -C "$R" init -q -b main && venv git -C "$R" config user.name Gate &&
  venv git -C "$R" config user.email gate@example.invalid && venv git -C "$R" config commit.gpgsign false || fail "git init"
echo x >"$R/docs/x.md"
(cd "$R" && venv "$t/bin/new" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
cfg() { printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n[agent]\nreview = "%s"\n[model]\nreview = "gpt-x"\n' "$1" >"$R/.vloop/config.toml"; }
cfg copilot
cp "$H/brief.md" "$R/$BRIEF" || fail brief
venv git -C "$R" add -A && venv git -C "$R" commit -q -m "gate fixture" || fail "fixture commit"

# doctor <label> [VAR=value...]: doctor's text in $t/d.txt ($drc its exit) and
# its JSON in $t/d.json; $row is the copilot login row as result<TAB>message.
doctor() {
  l=$1; shift
  (cd "$R" && venv "$@" "$t/bin/new" doctor >"$t/d.txt" 2>&1); drc=$?
  (cd "$R" && venv "$@" "$t/bin/new" doctor --json >"$t/d.json" 2>"$t/d.err")
  "$JQ" -e . "$t/d.json" >/dev/null 2>&1 || fail "$l: doctor --json is not JSON: $(head -3 "$t/d.json" "$t/d.err")"
  row=$("$JQ" -r '[.checks[] | select(.name == "copilot login")] | if length == 1 then "\(.[0].result)\t\(.[0].message)" else "rows: \(length)" end' "$t/d.json")
  for v in "$TOKEN" "$ENVTOKEN"; do
    grep -qF -- "$v" "$t/d.txt" "$t/d.json" "$t/d.err" && fail "$l: doctor prints a token"
  done
  return 0
}
# textrow <file>: the copilot login line of doctor's text output, its spaces squeezed.
textrow() { grep -E '^[^ ]+ +copilot login( |$)' "$1" | tr -s ' '; }
want_problem() {
  doctor "$@"
  [ "$row" = "problem	$MSG" ] || fail "$1: doctor's copilot login row is '$row', want a problem with: $MSG"
  [ "$(textrow "$t/d.txt")" = "✗ copilot login $MSG" ] || fail "$1: doctor prints '$(textrow "$t/d.txt")', want '✗ copilot login  $MSG'"
  [ "$drc" = 1 ] || fail "$1: doctor exits $drc with copilot login ✗, want 1"
}
want_pass() {
  doctor "$@"
  [ "$row" = "pass	" ] || fail "$1: doctor's copilot login row is '$row', want pass"
  [ "$(textrow "$t/d.txt")" = "✓ copilot login" ] || fail "$1: doctor prints '$(textrow "$t/d.txt")', want '✓ copilot login'"
}

# No login anywhere.
gh 1 ""
want_problem "no token, gh exits 1"
grep -qxF -- "$WARN" "$t/d.txt" || fail "doctor does not print the warning: $WARN"
grep -qF 'until B20' "$t/d.txt" && fail "doctor still prints the old warning: $(grep -F 'until B20' "$t/d.txt")"
want_problem "GH_TOKEN empty, gh exits 1" GH_TOKEN=
gh 0 ""
want_problem "no token, gh exits 0 and prints nothing"

# The run refuses before planning.
gh 1 ""
: >"$t/stub/sessions.log"
(cd "$R" && venv "$t/bin/new" run "$BRIEF" >"$t/r.out" 2>"$t/r.err"); rc=$?
[ "$rc" = 1 ] || fail "vloop run with no login exits $rc, want 1: $(tail -5 "$t/r.err")"
sed 's/^[[:space:]]*//' "$t/r.err" | grep -qxF -- "$RUNLINE" || fail "vloop run with no login does not print: $RUNLINE — stderr: $(tail -5 "$t/r.err")"
[ ! -s "$t/stub/sessions.log" ] || fail "vloop run with no login started a session: $(cat "$t/stub/sessions.log")"
[ ! -e "$R/.vloop/state/state.json" ] || fail "vloop run with no login left .vloop/state/state.json"

# A login: gh, then each variable with the failing gh.
gh 0 "$TOKEN"
want_pass "gh prints a token"
: >"$t/stub/sessions.log"
(cd "$R" && venv "$t/bin/new" run "$BRIEF" >"$t/r.out" 2>"$t/r.err")
sed 's/^[[:space:]]*//' "$t/r.err" | grep -qxF -- "$WARN" || fail "the run preflight does not print the warning: $WARN — stderr: $(head -5 "$t/r.err")"
grep -qF 'until B20' "$t/r.err" && fail "the run preflight still prints the old warning"
grep -qF -- "$RUNLINE" "$t/r.err" && fail "vloop run refuses with a gh that prints a token"
grep -qF -- "$TOKEN" "$t/r.out" "$t/r.err" && fail "vloop run prints gh's token"
grep -rqF -- "$TOKEN" "$R/.vloop" && fail "gh's token is kept under .vloop/: $(grep -rlF -- "$TOKEN" "$R/.vloop" | sed "s#^$R/##")"
gh 1 ""
for v in COPILOT_GITHUB_TOKEN GH_TOKEN GITHUB_TOKEN; do
  want_pass "$v set, gh exits 1" "$v=$ENVTOKEN"
done
grep -rqF -- "$ENVTOKEN" "$R/.vloop" && fail "a token variable's value is kept under .vloop/"

# claude reviews: no copilot login row.
cfg claude
doctor "agent.review = claude"
[ "$row" = "rows: 0" ] || fail "with agent.review = claude doctor has a copilot login row ($row)"

# The text.
G=docs/guide/configuration.md
old=$(grep -rnF -e 'not fenced until B20' -e 'experimental until B20' plugin internal cmd docs/guide docs/domain README.md 2>/dev/null)
[ -z "$old" ] || fail "the old warning text is still there: $old"
para=$(grep -F 'experimental until the fence lands' "$G")
[ -n "$para" ] || fail "$G does not say the Copilot review sessions are experimental until the fence lands"
for w in COPILOT_GITHUB_TOKEN GH_TOKEN GITHUB_TOKEN 'gh auth'; do
  case "$para" in *"$w"*) ;; *) fail "$G's Copilot paragraph does not say how a session logs in: no $w";; esac
done
echo "gate T4: ok"
