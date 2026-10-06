#!/bin/sh
# Gate T9 — the rest of brief Q9 that a binary shows: cmd/gendocs refuses an
# argument that starts with "-" and writes no file for it, while still writing
# the reference to a real path; the stray file named --help at the repository
# root is gone; and the CLAUDE.md section vloop init writes lists the session
# kinds plan, gate review, work and review and names /vloop:operate as the
# operator's playbook. The driver's swallowed errors are the review's to judge.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
export GIT_AUTHOR_NAME=gate GIT_AUTHOR_EMAIL=gate@example.invalid GIT_COMMITTER_NAME=gate GIT_COMMITTER_EMAIL=gate@example.invalid
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
go build -o "$t/gendocs" ./cmd/gendocs || fail "go build ./cmd/gendocs"
GOOS=windows go build -o "$t/gendocs.exe" ./cmd/gendocs || fail "GOOS=windows go build ./cmd/gendocs"
B="$t/vloop"

# --- the stray --help file ---
[ -e ./--help ] && fail "the file named --help at the repository root is still there"

# --- gendocs ---
mkdir -p "$t/gd" || fail mkdir
for a in --help -h -o; do
  (cd "$t/gd" && "$t/gendocs" "$a") >"$t/out" 2>"$t/err"; got=$?
  [ "$got" = 0 ] && fail "gendocs $a exited 0"
  [ -e "$t/gd/$a" ] && fail "gendocs $a wrote a file named $a"
done
[ -n "$(ls -A "$t/gd")" ] && fail "gendocs wrote files for a refused argument: $(ls -A "$t/gd")"
(cd "$t/gd" && "$t/gendocs" commands.md) >"$t/out" 2>"$t/err" || fail "gendocs commands.md failed: $(cat "$t/err")"
head -1 "$t/gd/commands.md" | grep -q '^# Command reference' || fail "gendocs did not write the command reference to a real path"

# --- the CLAUDE.md section ---
mkdir -p "$t/home" "$t/r" && export HOME="$t/home" || fail mkdir
git -C "$t/r" init -q -b main || fail "git init"
(cd "$t/r" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init: $(cat "$t/init.out")"
sed -n '/<!-- vloop:begin -->/,/<!-- vloop:end -->/p' "$t/r/CLAUDE.md" >"$t/section"
[ -s "$t/section" ] || fail "vloop init wrote no vloop section to CLAUDE.md"
grep -Fq '/vloop:operate' "$t/section" || fail "the CLAUDE.md section does not name /vloop:operate as the operator's playbook"
grep -Fq 'vloop-operator' "$t/section" && fail "the CLAUDE.md section still names vloop-operator, a skill the plugin does not ship"
grep -i 'plan' "$t/section" | grep -i 'gate[ -]review' | grep -i 'work' | grep -iq 'review' \
  || fail "no line of the CLAUDE.md section lists the session kinds plan, gate review, work and review: $(grep -i 'sessions' "$t/section")"
echo "gate T9: ok"
