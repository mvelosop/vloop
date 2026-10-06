#!/bin/sh
# Gate T2 — one frontmatter reader and rewriter, for defects and interventions
# (brief Q8). A record saved with CRLF line endings, with or without a leading
# BOM, is read by list and show, and rewritten by set: the BOM kept, every line
# still CRLF, and every byte other than the changed line as it was. The
# reference for "as it was" is the same edit on an LF twin of the record in a
# second repository, converted to the twin's endings: a rewrite of a CRLF file
# must differ from the LF rewrite by its line endings and BOM only.
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
B="$t/vloop"
mkdir -p "$t/home" "$t/bin" && printf '#!/bin/sh\nexit 1\n' >"$t/bin/claude" && chmod +x "$t/bin/claude" || fail "fake home"
export HOME="$t/home" PATH="$t/bin:$PATH"
newrepo() {
  R="$t/$1"
  mkdir -p "$R" && git -C "$R" init -q -b main && git -C "$R" commit -q --allow-empty -m "gate base" || fail "git init $1"
  (cd "$R" && "$B" init </dev/null) >"$t/init.out" 2>&1 || fail "vloop init in $1: $(cat "$t/init.out")"
}
run() { (cd "$R" && "$B" "$@") >"$t/out" 2>"$t/err"; }
crlf() { awk '{ printf "%s\r\n", $0 }' "$1"; }
bom() { printf '\357\273\277'; }
# allcrlf <file>: every line ends CRLF.
allcrlf() { awk '{ if (substr($0, length($0), 1) != "\r") bad++ } END { exit bad > 0 }' "$1"; }

newrepo a
BRIEF=$(cd "$R/docs/briefs" && ls *.loop-brief.md | head -1) || fail "vloop init wrote no brief"
BRIEF=${BRIEF%.md}
run defect add "a defect saved with CRLF endings" --found-by operator --brief "$BRIEF" || fail "defect add: $(cat "$t/err")"
DPATH=$(cat "$t/out")
DID=$(basename "$DPATH" .md)
run defect add "a defect saved with a BOM" --found-by operator --brief "$BRIEF" || fail "defect add: $(cat "$t/err")"
D2PATH=$(cat "$t/out")
D2ID=$(basename "$D2PATH" .md)
run intervention add "an intervention saved with a BOM and CRLF" --phase run --kind repair --automatable no --by operator \
  --option "replace the gate" --option "reset the task" --recommended 1 --why "the gate is wrong" --decided-option 1 \
  || fail "intervention add: $(cat "$t/err")"
IPATH=$(cat "$t/out")
IID=$(basename "$IPATH" .md)
[ -f "$R/$DPATH" ] && [ -f "$R/$D2PATH" ] && [ -f "$R/$IPATH" ] || fail "add did not print repo-relative paths of the records it wrote"

# The LF twin: the same records, edited the same way, in a second repository.
mkdir -p "$t/twin" && cp -R "$R/." "$t/twin/" || fail "twin"
cp "$R/$DPATH" "$t/d.lf" && cp "$R/$D2PATH" "$t/d2.lf" && cp "$R/$IPATH" "$t/i.lf" || fail "keep the LF records"
crlf "$t/d.lf" >"$R/$DPATH"
{ bom; cat "$t/d2.lf"; } >"$R/$D2PATH"
{ bom; crlf "$t/i.lf"; } >"$R/$IPATH"

# --- read ---
run defect list || fail "defect list fails on a CRLF and a BOM record: $(cat "$t/err")"
grep -Fq "$DID" "$t/out" || fail "defect list does not list the CRLF record $DID"
grep -Fq "$D2ID" "$t/out" || fail "defect list does not list the BOM record $D2ID"
run intervention list || fail "intervention list fails on a BOM + CRLF record: $(cat "$t/err")"
grep -Fq "$IID" "$t/out" || fail "intervention list does not list the BOM + CRLF record $IID"
run intervention show "$IID" || fail "intervention show fails on a BOM + CRLF record: $(cat "$t/err")"
grep -Fq "replace the gate" "$t/out" || fail "intervention show of a BOM + CRLF record lost its options"

# --- rewrite ---
(cd "$t/twin" && "$B" defect set "$DID" status fixed && "$B" defect set "$D2ID" severity high && "$B" intervention set "$IID" phase verify) \
  >"$t/twin.out" 2>&1 || fail "the LF twin's set failed: $(cat "$t/twin.out")"
run defect set "$DID" status fixed || fail "defect set on a CRLF record exited non-zero: $(cat "$t/err")"
run defect set "$D2ID" severity high || fail "defect set on a BOM record exited non-zero: $(cat "$t/err")"
run intervention set "$IID" phase verify || fail "intervention set on a BOM + CRLF record exited non-zero: $(cat "$t/err")"

grep -q '^status: fixed' "$R/$DPATH" || fail "defect set left the CRLF record's status unchanged"
allcrlf "$R/$DPATH" || fail "defect set left a CRLF record with lines that do not end CRLF"
crlf "$t/twin/$DPATH" >"$t/d.want"
cmp -s "$R/$DPATH" "$t/d.want" || fail "defect set on a CRLF record changed more than the status line (or its endings): $(diff "$t/d.want" "$R/$DPATH" | od -c | head -8)"

{ bom; cat "$t/twin/$D2PATH"; } >"$t/d2.want"
cmp -s "$R/$D2PATH" "$t/d2.want" || fail "defect set on a BOM record did not keep the BOM, or changed more than the severity line"

head -c 3 "$R/$IPATH" | od -An -tx1 | tr -d ' \n' | grep -q '^efbbbf$' || fail "intervention set dropped the record's BOM"
allcrlf "$R/$IPATH" || fail "intervention set left a CRLF record with lines that do not end CRLF"
{ bom; crlf "$t/twin/$IPATH"; } >"$t/i.want"
cmp -s "$R/$IPATH" "$t/i.want" || fail "intervention set on a BOM + CRLF record differs from the same edit on an LF twin by more than its BOM and endings"
run intervention list || fail "intervention list after set: $(cat "$t/err")"
grep "$IID" "$t/out" | grep -q verify || fail "intervention list does not show the new phase of the rewritten record"
echo "gate T2: ok"
