#!/bin/sh
# Gate T5 — V2's skill: the plugin vloop extracts for sessions carries
# skills/verify/SKILL.md (name verify), which names the report, its schema
# check, the three kinds and the report's fields; it is within its word
# ceiling of 1,500, and the budget reports a verify skill one word past it.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newrepo p5
inr "$B" plugin path </dev/null >"$t/pp" 2>&1 || fail "vloop plugin path: $(cat "$t/pp")"
p=$(tail -n 1 "$t/pp")
case "$p" in /*) P=$p ;; *) P=$R/$p ;; esac
F=$P/skills/verify/SKILL.md
[ -f "$F" ] || fail "the extracted plugin has no skills/verify/SKILL.md: $(ls "$P/skills" 2>&1)"
cmp -s "$F" "$top/plugin/skills/verify/SKILL.md" || fail "the extracted verify skill is not plugin/skills/verify/SKILL.md"
[ "$(head -n 1 "$F")" = "---" ] || fail "the verify skill has no frontmatter"
awk '$0 == "---" { n++; next } n == 1' "$F" >"$t/fm"
grep -Fxq 'name: verify' "$t/fm" || fail "the verify skill's frontmatter has no 'name: verify': $(cat "$t/fm")"
grep -q '^description: ..*' "$t/fm" || fail "the verify skill's frontmatter has no description"
w=$(wc -w <"$F" | tr -d ' ')
[ "$w" -le 1500 ] || fail "the verify skill is $w words, over its ceiling of 1,500"
for s in .vloop/tmp/verify.json 'vloop schema validate verify/v1 .vloop/tmp/verify.json' verify/v1 \
  example contract surface claim observed where example_lines sentences surfaces; do
  grep -Fq -- "$s" "$F" || fail "the verify skill does not mention '$s'"
done
mkdir -p "$t/ov" && cp "$G/budget_gate_test.go.txt" "$t/ov/gate_test.go" || fail "copy the budget judge"
"$JQ" -n --arg a "$top/internal/budget/zz_gate_t5_test.go" --arg b "$t/ov/gate_test.go" '{Replace: {($a): $b}}' >"$t/ov/overlay.json" || fail overlay
GATE_TOP=$top go test -count=1 -overlay "$t/ov/overlay.json" -run '^TestGateT5VerifyCeiling$' ./internal/budget >"$t/bt" 2>&1 ||
  fail "the budget does not hold the verify skill to 1,500 words:
$(cat "$t/bt")"
echo "gate T5: ok"
