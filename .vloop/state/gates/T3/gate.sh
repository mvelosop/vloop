#!/bin/sh
# Gate T3 — briefs and brief close through the one frontmatter reader and
# rewriter (brief Q8). A ready brief saved with a BOM and CRLF endings checks and
# lists like its LF twin; brief close rewrites only the first `**Status:**` line
# (the one the template starts with), not a later one nor one quoted mid-line,
# keeps the BOM, and leaves every line CRLF, the run record it appends included.
# The reference for the closed file is the same close of an LF twin in a second
# repository, converted to CRLF with the BOM.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C TZ=UTC GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
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
N=B20260101-0900-demo.loop-brief
F=docs/briefs/$N.md
crlf() { awk '{ printf "%s\r\n", $0 }' "$1"; }
bom() { printf '\357\273\277'; }
allcrlf() { awk '{ if (substr($0, length($0), 1) != "\r") bad++ } END { exit bad > 0 }' "$1"; }
# brepo <name> <crlf|lf>: an init'd repository holding the fixture brief (BOM +
# CRLF, or plain LF), one run folder for it, all committed, on a work branch.
brepo() {
  R="$t/$1"
  mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init $1"
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
  mkdir -p "$R/docs/briefs" && echo x >"$R/docs/x.md" || fail "docs"
  if [ "$2" = crlf ]; then { bom; crlf "$G/brief.md"; } >"$R/$F"; else cp "$G/brief.md" "$R/$F"; fi
  mkdir -p "$R/.vloop/state/runs/B20260101-0900-demo/20260101-090000" && echo "planning from $F" >"$R/.vloop/state/runs/B20260101-0900-demo/20260101-090000/run.log" || fail "run folder"
  git -C "$R" add -A && git -C "$R" commit -q -m fixture && git -C "$R" checkout -q -b work || fail "commit the fixture in $1"
}
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }

brepo lf lf
LF=$R
run brief check "$F" || fail "the LF fixture brief does not check: $(cat "$t/out" "$t/err")"
brepo crlf crlf

# --- read ---
run brief check "$F" || fail "brief check fails on a ready brief saved with a BOM and CRLF endings: $(cat "$t/out" "$t/err" | head -5)"
run brief list --json || fail "brief list --json: $(cat "$t/err")"
jq -e --arg n "$N" 'map(select(.name == $n)) | length == 1 and .[0].status == "ready" and .[0].path == "docs/briefs/\($n).md"' "$t/out" >/dev/null \
  || fail "brief list does not read the BOM + CRLF brief's name and status: $(cat "$t/out")"

# --- close ---
(cd "$LF" && "$B" brief close "$N" --abandon "the gate abandons it" --no-findings) >"$t/lf.out" 2>&1 || fail "brief close of the LF twin failed: $(cat "$t/lf.out")"
run brief close "$N" --abandon "the gate abandons it" --no-findings || fail "brief close of a BOM + CRLF brief failed: $(cat "$t/err")"

# Only the first status line changes, in the LF twin as in the CRLF brief.
grep -n 'Status:\*\*' "$LF/$F" >"$t/status.lines"
awk 'NR == 1' "$t/status.lines" | grep -q '^12:\*\*Status:\*\* abandoned — the gate abandons it$' \
  || fail "brief close did not rewrite the first **Status:** line to 'abandoned — <reason>': $(head -1 "$t/status.lines")"
grep -q '^16:A thing\. Quoted later: \*\*Status:\*\* ready to plan, as the template says\.$' "$t/status.lines" \
  || fail "brief close rewrote a **Status:** quoted in the middle of a line: $(grep '^16:' "$t/status.lines")"
[ "$(grep -c ':\*\*Status:\*\* ready to plan$' "$t/status.lines")" = 1 ] \
  || fail "brief close rewrote a later **Status:** line too: $(cat "$t/status.lines")"
sed -n '/^---$/,/^---$/p' "$LF/$F" | grep -q '^status: abandoned$' || fail "brief close did not set status: abandoned in the frontmatter"

head -c 3 "$R/$F" | od -An -tx1 | tr -d ' \n' | grep -q '^efbbbf$' || fail "brief close dropped the brief's BOM"
allcrlf "$R/$F" || fail "brief close left a CRLF brief with lines that do not end CRLF (mixed line endings)"
{ bom; crlf "$LF/$F"; } >"$t/want"
cmp -s "$R/$F" "$t/want" || fail "brief close of a BOM + CRLF brief differs from the close of its LF twin by more than the BOM and endings: $(diff "$t/want" "$R/$F" | head -6 | od -c | head -6)"
run brief list --json || fail "brief list --json after close: $(cat "$t/err")"
jq -e --arg n "$N" 'map(select(.name == $n)) | .[0].status == "abandoned"' "$t/out" >/dev/null \
  || fail "brief list does not read the closed BOM + CRLF brief as abandoned"
echo "gate T3: ok"
