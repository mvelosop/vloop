#!/bin/sh
# Gate T3 — V2, the verify skill. Prose is the review's to judge; this gate
# holds plugin/skills/verify/SKILL.md to the facts the brief pins: its one
# ```json report example validates as verify/v2 with this tree's binary and
# carries exactly one not_run entry; it names verify/v2 and never verify/v1;
# it validates with `vloop schema validate verify/v2 .vloop/tmp/verify.json`,
# which the verify fence allows; it names the commands the verify fence denies
# (vloop init, vloop run, vloop brief review, vloop brief close, claude), the
# reasons `the operator's, at close` and `covered by`, complete and not_run;
# it makes scratch repositories with `git init -q`, never `git -C`, and the
# verify fence allows that form; it cites no fence/ path; it is at most 1500
# words (wc -w, its ceiling, which this brief does not raise).
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
for v in $(env | sed -n 's/^\(VLOOP_[A-Z0-9_]*\)=.*/\1/p'); do unset "$v"; done
export LC_ALL=C GOTOOLCHAIN=local
top=$(pwd -P)
K=$top/plugin/skills/verify/SKILL.md
[ -f "$K" ] || fail "plugin/skills/verify/SKILL.md is missing"
JQ=$(command -v jq) || fail "jq is not on PATH"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") && t=$(cd "$t" && pwd -P) || fail mktemp
trap 'rm -rf "$t"' EXIT
w=$(wc -w <"$K" | tr -d ' ')
[ "$w" -le 1500 ] || fail "plugin/skills/verify/SKILL.md is $w words, over its ceiling of 1500"
grep -qF 'verify/v1' "$K" && fail "plugin/skills/verify/SKILL.md still names verify/v1: $(grep -nF 'verify/v1' "$K")"
grep -qF 'vloop schema validate verify/v2 .vloop/tmp/verify.json' "$K" || fail "the skill does not validate its report with 'vloop schema validate verify/v2 .vloop/tmp/verify.json'"
grep -qE '(^|[^a-z_])fence/' "$K" && fail "the skill cites a fence/ path: $(grep -nE '(^|[^a-z_])fence/' "$K")"
grep -qF 'git -C' "$K" && fail "the skill still says git -C: $(grep -nF 'git -C' "$K")"
grep -qF 'git init -q' "$K" || fail "the skill does not make scratch repositories with 'git init -q <dir>'"
for s in 'vloop init' 'vloop run' 'vloop brief review' 'vloop brief close' 'claude' 'not_run' 'complete' "the operator's, at close" 'covered by'; do
  grep -qF -- "$s" "$K" || fail "the skill does not name '$s'"
done
# The report example: the skill's ```json blocks, each a verify/v2 report.
awk '/^```json[[:space:]]*$/{n++; on=1; next} /^```/{on=0} on{print > (dir "/ex" n ".json")}' dir="$t" "$K"
ls "$t"/ex*.json >/dev/null 2>&1 || fail "the skill has no \`\`\`json report example"
go build -buildvcs=false -o "$t/vloop" ./cmd/vloop || fail "go build of this tree"
mkdir -p "$t/r" && cd "$t/r" || fail mkdir
V() { env -i PATH=/usr/bin:/bin HOME="$t" LC_ALL=C "$t/vloop" "$@"; }
found=0
for f in "$t"/ex*.json; do
  [ "$("$JQ" -r '.schema // empty' "$f" 2>/dev/null)" = verify/v2 ] || continue
  found=1
  V schema validate verify/v2 "$f" >"$t/o" 2>&1 || fail "the skill's report example is not a valid verify/v2 report: $(cat "$t/o")"
  [ "$("$JQ" '.not_run | length' "$f")" = 1 ] || fail "the skill's report example has $("$JQ" '.not_run | length' "$f") not_run entries, want one"
done
[ "$found" = 1 ] || fail "no \`\`\`json block of the skill is a verify/v2 report"
for c in 'vloop schema validate verify/v2 .vloop/tmp/verify.json' 'git init -q .vloop/tmp/vs/ex'; do
  o=$(V fence check --phase verify --root . --tool bash "$c" 2>&1)
  [ "$o" = allow ] || fail "the verify fence does not allow '$c': $o"
done
echo "gate T3: ok"
