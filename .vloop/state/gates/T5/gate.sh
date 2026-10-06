#!/bin/sh
# Gate T5 — messages say what to do next (brief Q3): no plan, no task, no
# brief, a -C that names no directory, and git outside a repository. Each is
# exactly one stderr line, exit 1.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE GIT_CEILING_DIRECTORIES CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
B="$t/vloop"
mkdir -p "$t/home" "$t/bin" && printf '#!/bin/sh\nexit 1\n' >"$t/bin/claude" && chmod +x "$t/bin/claude" || fail "fake home"
export HOME="$t/home" PATH="$t/bin:$PATH"
R="$t/r"
mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init"
(cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init: $(cat "$t/init.out")"
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }
exact() {
  want=$1; msg=$2; shift 2
  run "$@"; got=$?
  [ "$got" = "$want" ] || fail "vloop $* exited $got, want $want — stderr: $(cat "$t/err")"
  [ "$(cat "$t/err")" = "$msg" ] && [ "$(wc -l <"$t/err" | tr -d ' ')" = 1 ] || fail "vloop $* printed '$(cat "$t/err")' on stderr, want exactly '$msg'"
}

# --- no plan ---
NOPLAN="vloop: no plan — vloop run <brief> makes one"
exact 1 "$NOPLAN" task show T1
exact 1 "$NOPLAN" status
exact 1 "$NOPLAN" task list

# --- no task ---
mkdir -p "$R/.vloop/state" && cp "$G/plan.json" "$R/.vloop/state/state.json" || fail "plan fixture"
exact 1 "vloop: no task T99 — vloop task list shows the plan's tasks" task show T99
run task show T1 || fail "task show T1 fails on a valid plan: $(cat "$t/err")"

# --- no brief, and a brief with no runs ---
exact 1 "vloop: no brief nope — vloop brief list shows the briefs" metrics nope
BRIEF=$(cd "$R/docs/briefs" && ls *.loop-brief.md | head -1) || fail "vloop init wrote no brief"
run metrics "${BRIEF%.md}"; got=$?
[ "$got" = 1 ] && grep -q '^vloop: no runs for ' "$t/err" || fail "vloop metrics on a brief with no runs exited $got with '$(cat "$t/err")', want exit 1 and 'vloop: no runs for <name>'"

# --- -C naming no directory, for every command ---
NODIR="$t/no-such-dir"
for c in "status" "task list" "task show T1" "metrics" "metrics export" "brief list" "brief check x.md" "defect list" \
  "intervention list" "config list" "config get language" "doctor" "plugin path" "schema list" "version" "init" "upgrade" "run"; do
  # shellcheck disable=SC2086
  (cd "$R" && "$B" -C "$NODIR" $c) >"$t/out" 2>"$t/err"; got=$?
  [ "$got" = 1 ] || fail "vloop -C <no such dir> $c exited $got, want 1 — stderr: $(cat "$t/err")"
  [ "$(tail -1 "$t/err")" = "vloop: -C $NODIR: no such directory" ] || fail "vloop -C <no such dir> $c printed '$(cat "$t/err")', want 'vloop: -C $NODIR: no such directory'"
done
[ -e "$NODIR" ] && fail "a -C naming no directory created it"

# --- git outside a repository ---
mkdir -p "$t/plain/sub" || fail mkdir
(cd "$t/plain/sub" && GIT_CEILING_DIRECTORIES="$t" "$B" metrics) >"$t/out" 2>"$t/err"; got=$?
[ "$got" = 1 ] || fail "vloop metrics outside a git repository exited $got, want 1 — stderr: $(cat "$t/err")"
[ "$(wc -l <"$t/err" | tr -d ' ')" = 1 ] && grep -q '^vloop: .* is not a git repository$' "$t/err" \
  || fail "vloop metrics outside a git repository printed '$(cat "$t/err")', want 'vloop: <dir> is not a git repository'"
grep -q 'exit status' "$t/err" && fail "a raw git error reached the user: $(cat "$t/err")"
echo "gate T5: ok"
