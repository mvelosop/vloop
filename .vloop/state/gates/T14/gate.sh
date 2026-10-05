#!/bin/sh
# Gate T14 — the evals (brief Q10), judged by their files, since no gate runs a
# model. plan-checks-its-gates: the grader that judges the planner's greet.toml
# fixtures (its body speaks of the [style] table) reads files, not the trace
# (which ends before the planner writes its gate folders) nor the final message
# (D20261004-1309). Both changed cases stay well-formed — every grader has a
# known type and a weight, operate-proposes-options keeps a judge grader — and
# both scaffolds still build their repositories in an empty directory. What the
# judges say is the review's and the operator's eval run's to rule on.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
unset NO_COLOR GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE CLAUDE_PLUGIN_ROOT
export LC_ALL=C GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 GIT_TERMINAL_PROMPT=0
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -o "$t/vloop" ./cmd/vloop || fail "go build ./cmd/vloop"
GOOS=linux go build -o "$t/vloop-linux" ./cmd/vloop || fail "GOOS=linux go build ./cmd/vloop"
GOOS=windows go build -o "$t/vloop.exe" ./cmd/vloop || fail "GOOS=windows go build ./cmd/vloop"
E=plugin/evals
# front <file>: the grader's frontmatter, without its fences.
front() { awk 'NR == 1 && $0 == "---" { on = 1; next } on && $0 == "---" { exit } on { print }' "$1"; }
# body <file>: what follows the frontmatter.
body() { awk 'NR == 1 && $0 == "---" { on = 1; next } on && $0 == "---" { on = 0; done = 1; next } done { print }' "$1"; }
# reads_file <file>: its focus or target is {source: file, path: ...}.
reads_file() {
  front "$1" | awk '
    /^(focus|target):/ { blk = 1; if ($0 ~ /source: *file/ && $0 ~ /path:/) ok = 1; next }
    blk && /^[^ ]/ { blk = 0 }
    blk && /source: *file/ { src = 1 }
    blk && /path:/ { pth = 1 }
    END { exit !(ok || (src && pth)) }'
}
wellformed() {
  case_dir=$1
  n=0
  for g in "$case_dir"/graders/*.md; do
    [ -f "$g" ] || continue
    n=$((n + 1))
    ty=$(front "$g" | sed -n 's/^type: *//p' | head -1)
    case "$ty" in regex | tool_used | tool_order | file_exists | llm | baseline) ;; *) fail "$g has no known grader type: '$ty'" ;; esac
    front "$g" | grep -q '^weight: *[0-9]' || fail "$g has no weight"
    [ -n "$(body "$g" | tr -d ' \n')" ] || fail "$g has no description"
  done
  [ "$n" -gt 0 ] || fail "$case_dir has no graders"
}

# --- plan-checks-its-gates ---
C=$E/plan-checks-its-gates
wellformed "$C"
found=0
for g in "$C"/graders/*.md; do
  body "$g" | grep -Fq '[style]' || continue
  found=$((found + 1))
  reads_file "$g" || fail "$g judges the planner's greet.toml fixtures but does not read a file: its focus is the trace (which ends before the gate folders are written) or the final message"
done
[ "$found" -gt 0 ] || fail "plan-checks-its-gates no longer has a grader judging the greet.toml fixtures against the [style] table"

# --- operate-proposes-options ---
C=$E/operate-proposes-options
wellformed "$C"
llm=0
for g in "$C"/graders/*.md; do
  front "$g" | grep -q '^type: *llm' && llm=$((llm + 1))
done
[ "$llm" -gt 0 ] || fail "operate-proposes-options has no judge grader left"

# --- both scaffolds still run ---
for c in plan-checks-its-gates operate-proposes-options; do
  mkdir -p "$t/$c" || fail mkdir
  (cd "$t/$c" && HOME="$t" PATH="$t:$PATH" GIT_AUTHOR_NAME=eval GIT_AUTHOR_EMAIL=eval@test GIT_COMMITTER_NAME=eval GIT_COMMITTER_EMAIL=eval@test bash "$OLDPWD/$E/$c/scaffold.sh") >"$t/$c.out" 2>&1 \
    || fail "the $c scaffold fails in an empty directory: $(tail -5 "$t/$c.out")"
done
echo "gate T14: ok"
