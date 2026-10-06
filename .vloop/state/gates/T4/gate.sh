#!/bin/sh
# Gate T4 — exit 2 is for usage only (brief Q2): an unknown command or flag, a
# wrong argument count, an invalid value for a flag or a settable field. Every
# other failure exits 1. The brief's table of changed codes, then a matrix of
# usage errors that must stay 2 and of failures that must be 1.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
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
# code <want> <args...>: vloop exits want, and a failure says so in one stderr line starting "vloop: ".
code() {
  want=$1; shift
  run "$@"; got=$?
  [ "$got" = "$want" ] || fail "vloop $* exited $got, want $want — stderr: $(cat "$t/err")"
  grep -q '^vloop: ' "$t/err" || fail "vloop $* exited $got without a 'vloop: ' line on stderr: $(cat "$t/err")"
}
# exact <want> <stderr> <args...>: as code, and stderr is exactly that one line.
exact() {
  want=$1; msg=$2; shift 2
  code "$want" "$@"
  [ "$(cat "$t/err")" = "$msg" ] && [ "$(wc -l <"$t/err" | tr -d ' ')" = 1 ] || fail "vloop $* printed '$(cat "$t/err")' on stderr, want exactly '$msg'"
}

BRIEF=$(cd "$R/docs/briefs" && ls *.loop-brief.md | head -1) || fail "vloop init wrote no brief"
run defect add "a defect" --found-by operator --brief "${BRIEF%.md}" || fail "defect add: $(cat "$t/err")"
DID=$(basename "$(cat "$t/out")" .md)
run intervention add "two options" --phase run --kind repair --automatable no --by operator \
  --option "replace the gate" --option "reset the task" --recommended 1 --why "the gate is wrong" --decided-option 1 \
  || fail "intervention add: $(cat "$t/err")"
IID=$(basename "$(cat "$t/out")" .md)
mkdir -p "$R/.vloop/state" && cp "$G/plan.json" "$R/.vloop/state/state.json" || fail "plan fixture"
run task show T1 || fail "the plan fixture does not load: $(cat "$t/err")"

# --- the brief's table ---
exact 1 "vloop: brief not found: docs/briefs/nope.md" run docs/briefs/nope.md
exact 1 "vloop: no such file: missing.md" brief check missing.md
code 2 intervention set "$IID" decided 5
code 2 intervention set "$IID" recommended 9

# --- usage: exit 2 ---
code 2 --bogus
code 2 nosuchcommand
code 2 intervention list --bogus
code 2 status extra-argument
code 2 brief check
code 2 task show
code 2 task show T1 T2
code 2 run --max-attempts 0
code 2 run --plan-only=maybe
code 2 config set language 5x
code 2 task set T1 kind bogus
code 2 defect set "$DID" status bogus
code 2 intervention set "$IID" phase bogus
code 2 intervention add "x" --phase run --kind repair --automatable no --by bogus

# --- not usage: exit 1 ---
code 1 defect set D20990101-0000-nope status fixed
code 1 intervention show I20990101-0000-nope
code 1 intervention set I20990101-0000-nope phase run
code 1 task show T99
code 1 schema validate state/v2 no-such-file.json
code 1 brief check docs/no-such-brief.loop-brief.md
echo "gate T4: ok"
