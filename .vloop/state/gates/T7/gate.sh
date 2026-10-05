#!/bin/sh
# Gate T7 — the remaining messages of brief Q3: brief check on drafts only says
# nothing was checked; plugin path outside a vloop repository refuses and writes
# nothing; intervention show prints the schema and options lines and omits an
# empty brief; intervention migrate --dry-run prints repo-relative paths.
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
newrepo() {
  R="$t/$1"
  mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init $1"
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
}
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }

# --- brief check on drafts only ---
newrepo drafts
run brief new second-draft || fail "brief new: $(cat "$t/err")"
n=$(cd "$R/docs/briefs" && ls *.loop-brief.md | wc -l | tr -d ' ')
[ "$n" = 2 ] || fail "expected two draft briefs, found $n"
(cd "$R" && "$B" brief check docs/briefs/*.loop-brief.md) >"$t/out" 2>"$t/err"; got=$?
[ "$got" = 0 ] || fail "brief check on drafts only exited $got, want 0: $(cat "$t/out" "$t/err")"
cat "$t/out" "$t/err" | grep -qx 'nothing checked: 2 draft brief(s) skipped' \
  || fail "brief check on drafts only does not print 'nothing checked: 2 draft brief(s) skipped': $(cat "$t/out" "$t/err")"
grep -q 'briefs ok' "$t/out" "$t/err" && fail "brief check on drafts only still says 'briefs ok'"

# --- plugin path outside a vloop repository ---
REFUSE="vloop: not in a vloop repository — run vloop init first"
mkdir -p "$t/plain" "$t/gitonly" && git -C "$t/gitonly" init -q -b main || fail "dirs"
for d in "$t/plain" "$t/gitonly"; do
  (cd "$d" && "$B" plugin path) >"$t/out" 2>"$t/err"; got=$?
  [ "$got" = 1 ] || fail "plugin path in $(basename "$d") exited $got, want 1 — stdout: $(cat "$t/out")"
  [ "$(cat "$t/err")" = "$REFUSE" ] || fail "plugin path in $(basename "$d") printed '$(cat "$t/err")', want '$REFUSE'"
  [ -e "$d/.vloop" ] && fail "plugin path outside a vloop repository wrote $(basename "$d")/.vloop"
done
[ -e "$t/home/.vloop" ] && fail "plugin path wrote under the home directory"
newrepo plug
run plugin path || fail "plugin path in a vloop repository failed: $(cat "$t/err")"
grep -q '^\.vloop/tmp/plugin/' "$t/out" || fail "plugin path in a vloop repository does not print .vloop/tmp/plugin/<version>: $(cat "$t/out")"

# --- intervention show ---
newrepo show
BRIEF=$(cd "$R/docs/briefs" && ls *.loop-brief.md | head -1) || fail "no brief"
BRIEF=${BRIEF%.md}
run intervention add "no brief, no options" --phase run --kind repair --automatable no --by operator || fail "add: $(cat "$t/err")"
I1=$(basename "$(cat "$t/out")" .md)
run intervention add "a brief and two options" --brief "$BRIEF" --phase halt --kind repair --automatable partly --by both \
  --option "replace the gate" --option "reset the task" --recommended 1 --why "the gate is wrong" --decided-option 2 \
  || fail "add: $(cat "$t/err")"
I2=$(basename "$(cat "$t/out")" .md)
run intervention show "$I1" || fail "show $I1: $(cat "$t/err")"
grep -q '^schema: *intervention/v2 *$' "$t/out" || fail "intervention show does not print a 'schema: intervention/v2' line: $(cat "$t/out")"
grep -q '^options: *0 *$' "$t/out" || fail "intervention show does not print an 'options: 0' line"
grep -q '^brief:' "$t/out" && fail "intervention show prints an empty brief line for a record with no brief"
run intervention show "$I2" || fail "show $I2: $(cat "$t/err")"
grep -q "^brief: *$BRIEF *\$" "$t/out" || fail "intervention show does not print the record's brief: $(cat "$t/out")"
grep -q '^options: *2 *$' "$t/out" || fail "intervention show does not print an 'options: 2' line for a record with two options"
grep -q '^schema: *intervention/v2 *$' "$t/out" || fail "intervention show does not print the schema line for $I2"

# --- intervention migrate --dry-run ---
V1=I20260101-0900-a-v1-record
mkdir -p "$R/.vloop/interventions" && cp "$G/$V1.md" "$R/.vloop/interventions/$V1.md" || fail "v1 fixture"
for sub in . docs; do
  (cd "$R/$sub" && "$B" intervention migrate --dry-run) >"$t/out" 2>"$t/err" || fail "migrate --dry-run in $sub: $(cat "$t/err")"
  grep -Fq ".vloop/interventions/$V1.md" "$t/out" || fail "migrate --dry-run (from $sub) does not print the repo-relative path .vloop/interventions/$V1.md: $(cat "$t/out")"
  grep -Fq "$t" "$t/out" && fail "migrate --dry-run prints an absolute path"
done
grep -q '^schema:' "$R/.vloop/interventions/$V1.md" && fail "migrate --dry-run wrote the record"
echo "gate T7: ok"
