#!/bin/sh
# Gate T1 — V3, brief-review/v1's `brief` is a brief's name. This tree's
# binary, built into a temporary directory, validates each report in valid/
# as brief-review/v1 (exit 0) — among them the worked example's
# `B1-x.loop-brief` and `.loop-brief` — and refuses each one in invalid/
# (exit 1, not the unknown-schema exit 2): a `/` path, a `\` path, the file
# name with `.md`, `b`, and the rest of the pattern's edges. Nothing else in
# the schema changes: schemas/brief-review.v1.json without .properties.brief
# equals the base's (base-schema.json in this folder), and its title stays
# brief-review/v1. Real data: every kept
# .vloop/state/brief-reviews/*/*/brief-review.json validates (at least one is
# found), and the review-brief skill's ```json report example validates.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
top=$(pwd -P)
JQ=$(command -v jq) || fail "jq is not on PATH"
S=$top/schemas/brief-review.v1.json
[ "$("$JQ" -r .title "$S")" = "brief-review/v1" ] || fail "schemas/brief-review.v1.json's title is not brief-review/v1"
[ "$("$JQ" -S 'del(.properties.brief)' "$S")" = "$("$JQ" -S 'del(.properties.brief)' "$G/base-schema.json")" ] ||
  fail "schemas/brief-review.v1.json changed outside .properties.brief: only brief's pattern is added"
[ "$("$JQ" -r '.properties.brief.type' "$S")" = string ] || fail "brief is no longer a string"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build of this tree"
mkdir -p "$t/r" || fail mkdir
V() { (cd "$t/r" && env -i PATH=/usr/bin:/bin HOME="$t" LC_ALL=C "$t/vloop" "$@"); }
n=0
for f in "$G"/invalid/*.json; do
  V schema validate brief-review/v1 "$f" >"$t/o" 2>&1; c=$?
  [ "$c" = 1 ] || fail "vloop schema validate brief-review/v1 invalid/$(basename "$f") (brief $("$JQ" -c .brief "$f")) exited $c, want 1: $(cat "$t/o")"
  n=$((n + 1))
done
[ "$n" -ge 6 ] || fail "only $n invalid fixtures found"
n=0
for f in "$G"/valid/*.json; do
  V schema validate brief-review/v1 "$f" >"$t/o" 2>&1; c=$?
  [ "$c" = 0 ] || fail "vloop schema validate brief-review/v1 valid/$(basename "$f") exited $c, want 0: $(cat "$t/o")"
  n=$((n + 1))
done
[ "$n" -ge 3 ] || fail "only $n valid fixtures found"
n=0
for f in "$top"/.vloop/state/brief-reviews/*/*/brief-review.json; do
  [ -f "$f" ] || continue
  V schema validate brief-review/v1 "$f" >"$t/o" 2>&1; c=$?
  [ "$c" = 0 ] || fail "the kept ${f#"$top"/} does not validate as brief-review/v1 (exit $c): $(cat "$t/o")"
  n=$((n + 1))
done
[ "$n" -ge 1 ] || fail "no kept .vloop/state/brief-reviews/*/*/brief-review.json found"
K=$top/plugin/skills/review-brief/SKILL.md
awk '/^```json[[:space:]]*$/{n++; on=1; next} /^```/{on=0} on{print > (dir "/ex" n ".json")}' dir="$t" "$K"
found=0
for f in "$t"/ex*.json; do
  [ -f "$f" ] || continue
  [ "$("$JQ" -r .schema "$f" 2>/dev/null)" = brief-review/v1 ] || continue
  V schema validate brief-review/v1 "$f" >"$t/o" 2>&1 || fail "the review-brief skill's report example does not validate: $(cat "$t/o")"
  found=$((found + 1))
done
[ "$found" -ge 1 ] || fail "plugin/skills/review-brief/SKILL.md has no brief-review/v1 \`\`\`json example"
echo "gate T1: ok"
