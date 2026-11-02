#!/bin/sh
# Gate T3 — V2, the plan phase reverts what its sessions change elsewhere.
# This tree's binary, built into a temporary directory, runs `vloop run
# <brief>` in fixture repositories under $t with the stub claude
# (stub-claude.sh) first on PATH and a fake HOME trusting the fixture; nothing
# of this repository's tree, refs or HOME is touched.
#   main: the plan session appends to README.md and creates notes.txt, the
#     gate-review session creates scratch.txt and appends to README.md → the
#     four warnings in order on stderr and in run.log, exit 0, the plan commit
#     holds the plan and none of the three, README.md as at the start,
#     notes.txt and scratch.txt gone.
#   planted, exit 3: the plan session appends to README.md and exits 3 → the
#     warning, then the planning-failure halt, exit 1; README.md as at the
#     start; git status lists nothing outside .vloop/.
#   planted, plan-line: T1's gate is `grep -q plan-line README.md` and the
#     plan session appends plan-line → reverted before the gates are checked
#     on the base, so the plan is accepted in one plan session; the work
#     session appends plan-line and the run completes, plan-line in README.md
#     once.
#   clean: a plan session writing only under .vloop/state/ and .vloop/tmp/ →
#     no "session changed" line, and its gate folder is in the plan commit.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p; s/^\(COPILOT_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT GOFLAGS GOBIN GH_TOKEN GITHUB_TOKEN
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
GITBIN=$(command -v git) || fail "git is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'chmod -R u+w "$t" 2>/dev/null; rm -rf "$t"' EXIT
GOMODCACHE=$(go env GOMODCACHE) && GOCACHE=$(go env GOCACHE) && GOPATH=$(go env GOPATH) || fail "go env"
export GOMODCACHE GOCACHE GOPATH GOPROXY=off
mkdir -p "$t/bin" && go build -buildvcs=false -o "$t/bin/vloop" ./cmd/vloop || fail "go build of this tree"
S=$t/bin/vloop
RID=B20260101-0900-demo
BRIEF=docs/briefs/$RID.loop-brief.md
README="the fixture's readme"
venv() {
  env -i PATH="$t/stub:/usr/bin:/bin" HOME="$t/home" USER=gateuser LOGNAME=gateuser TMPDIR="$t/tmp" LC_ALL=C \
    GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0 "$@"
}
g() { venv git -C "$t/repo" "$@"; }
# runcase <name> <plan json> [stub flag file...]: a fresh fixture — git on
# main, `vloop init`, one check, the brief, README.md, all committed — then
# this tree's `vloop run` of the brief. $rc; $t/<name>.out, .err, .ev; $RD
# the run folder, repo-relative with its trailing slash.
runcase() {
  n=$1; pj=$2; shift 2
  chmod -R u+w "$t/repo" "$t/home" 2>/dev/null; rm -rf "$t/repo" "$t/home" "$t/tmp" "$t/stub"
  mkdir -p "$t/repo/docs/briefs" "$t/home" "$t/tmp" "$t/stub" || fail mkdir
  pv=$("$JQ" -r .version "$top/plugin/.claude-plugin/plugin.json") || fail "plugin version"
  cp "$G/$pj" "$t/stub/plan.json" && ln -s "$GITBIN" "$t/stub/git" && ln -s "$JQ" "$t/stub/jq" &&
    sed "s/@PLUGIN_VERSION@/$pv/" "$G/stub-claude.sh" >"$t/stub/claude" && chmod +x "$t/stub/claude" || fail "stub"
  printf '{"projects":{"%s":{"hasTrustDialogAccepted":true}}}\n' "$t/repo" >"$t/home/.claude.json"
  g init -q -b main && g config user.name Gate && g config user.email gate@example.invalid && g config commit.gpgsign false || fail "git init"
  echo x >"$t/repo/docs/x.md"
  echo "$README" >"$t/repo/README.md"
  (cd "$t/repo" && venv "$S" init >"$t/init.out" 2>&1) || fail "vloop init: $(cat "$t/init.out")"
  printf '[[check]]\nname = "all"\npaths = ["**"]\nrun = "true"\n' >"$t/repo/.vloop/config.toml"
  cp "$G/brief.md" "$t/repo/$BRIEF" || fail brief
  g add -A && g commit -q -m "gate fixture" || fail "fixture commit"
  for f in "$@"; do : >"$t/stub/$f"; done
  (cd "$t/repo" && venv "$S" run "$BRIEF" >"$t/$n.out" 2>"$t/$n.err"); rc=$?
  if [ -f "$t/stub/events" ]; then mv "$t/stub/events" "$t/$n.ev"; else : >"$t/$n.ev"; fi
  RD=$(cd "$t/repo" && ls -d .vloop/state/runs/*/*/ 2>/dev/null)
  [ "$(printf '%s\n' "$RD" | grep -c .)" = 1 ] || fail "$n: want one run folder, found: $RD"
  return 0
}
# completed <name>: exit 0, the plan complete, T1 done.
completed() {
  [ "$rc" = 0 ] || fail "$1: vloop run exited $rc, want 0: $(tail -8 "$t/$1.err")"
  st=$("$JQ" -r '.status + " " + ([.tasks[].status] | join(","))' "$t/repo/.vloop/state/state.json" 2>/dev/null)
  [ "$st" = "complete done" ] || fail "$1: the plan reads '$st', want 'complete done'"
}
# planfiles <name>: $t/<name>.pf, the files of the [vloop] plan <run id> commit.
planfiles() {
  h=$(g log --format='%H %s' | sed -n "s/^\([0-9a-f]*\) \[vloop\] plan $RID\$/\1/p" | head -1)
  [ -n "$h" ] || fail "$1: no '[vloop] plan $RID' commit: $(g log --format=%s | tr '\n' ';')"
  g show --name-only --format= "$h" >"$t/$1.pf" || fail "$1: git show"
}
# warnings <name>: $t/<name>.w, the "session changed … — reverted" lines on stderr.
warnings() { grep -E '^vloop: the (plan|gate-review) session changed ' "$t/$1.err" >"$t/$1.w"; }
readme_is_base() { [ "$(cat "$t/repo/README.md")" = "$README" ] || fail "$1: README.md is not as at the start: $(cat "$t/repo/README.md")"; }

# main
runcase main plan.json plan.readme plan.notes gr.dirty
completed main
warnings main
printf '%s\n' \
  'vloop: the plan session changed README.md — reverted' \
  'vloop: the plan session changed notes.txt — reverted' \
  'vloop: the gate-review session changed README.md — reverted' \
  'vloop: the gate-review session changed scratch.txt — reverted' >"$t/want.w"
cmp -s "$t/want.w" "$t/main.w" || fail "main: the revert warnings on stderr are:
$(cat "$t/main.w")
--- want
$(cat "$t/want.w")"
while IFS= read -r l; do
  grep -qF -- "$l" "$t/repo/${RD}run.log" || fail "main: ${RD}run.log does not hold '$l'"
done <"$t/want.w"
planfiles main
grep -qxF .vloop/state/state.json "$t/main.pf" || fail "main: the plan commit does not hold .vloop/state/state.json: $(tr '\n' ' ' <"$t/main.pf")"
for f in README.md notes.txt scratch.txt; do
  grep -qxF "$f" "$t/main.pf" && fail "main: the plan commit holds $f, a plan-phase session's stray change"
done
readme_is_base main
[ "$(g show HEAD:README.md)" = "$README" ] || fail "main: README.md at HEAD is not as at the start"
for f in notes.txt scratch.txt; do
  [ -e "$t/repo/$f" ] && fail "main: $f exists after the run"
  g cat-file -e "HEAD:$f" 2>/dev/null && fail "main: $f is committed"
done

# planted: the plan session changes README.md and exits 3.
runcase exit3 plan.json plan.readme plan.exit3
[ "$rc" = 1 ] || fail "exit3: vloop run exited $rc, want 1: $(tail -5 "$t/exit3.err")"
wl=$(grep -nxF 'vloop: the plan session changed README.md — reverted' "$t/exit3.err" | head -1 | cut -d: -f1)
hl=$(grep -nF 'planning session failed (claude exited 3)' "$t/exit3.err" | head -1 | cut -d: -f1)
[ -n "$wl" ] || fail "exit3: no 'vloop: the plan session changed README.md — reverted' on stderr: $(cat "$t/exit3.err")"
[ -n "$hl" ] || fail "exit3: no planning-failure halt on stderr: $(cat "$t/exit3.err")"
[ "$wl" -lt "$hl" ] || fail "exit3: the revert warning does not come before the halt: $(cat "$t/exit3.err")"
readme_is_base exit3
out=$(g status --porcelain --untracked-files=all | grep -v '^.. \.vloop/')
[ -z "$out" ] || fail "exit3: git status lists paths outside .vloop/: $out"

# planted: the plan session's line would pass T1's gate on the base.
runcase planline plan-planline.json plan.readme work.planline
completed planline
[ "$(grep -c '^session plan ' "$t/planline.ev")" = 1 ] || fail "planline: want one plan session (the gate fails on the base after the revert), got: $(grep '^session ' "$t/planline.ev" | tr '\n' ';')"
grep -qxF 'vloop: the plan session changed README.md — reverted' "$t/planline.err" || fail "planline: no revert warning: $(tail -5 "$t/planline.err")"
planfiles planline
grep -qxF README.md "$t/planline.pf" && fail "planline: the plan commit holds README.md"
[ "$(grep -c '^plan-line$' "$t/repo/README.md")" = 1 ] || fail "planline: README.md should hold plan-line once, the work session's: $(cat "$t/repo/README.md")"

# clean: nothing outside .vloop/state/ and .vloop/tmp/.
runcase clean plan.json
completed clean
grep -qF 'session changed' "$t/clean.err" "$t/clean.out" && fail "clean: a 'session changed' line with nothing changed: $(grep -F 'session changed' "$t/clean.err" "$t/clean.out")"
planfiles clean
for f in .vloop/state/state.json .vloop/state/gates/T1/gate.sh; do
  grep -qxF "$f" "$t/clean.pf" || fail "clean: the plan commit does not hold $f: $(tr '\n' ' ' <"$t/clean.pf")"
done
echo "gate T3: ok"
