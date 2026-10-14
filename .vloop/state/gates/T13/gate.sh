#!/bin/sh
# Gate T13 — D1: the docs, the domain and both operate playbooks describe the
# new model: second-stamped IDs (old ones stay), escaped by time against M-5's
# release, closing's commit carrying intervention files, the triage retired,
# the two per-task columns and $/kloc, the operator count's two sources,
# defects only for escapes; the operate skill within its ceiling of 1,750
# words; docs/guide/commands.md as gendocs makes it now.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
has() { grep -Fq -- "$2" "$1" || fail "$1 does not mention '$2'"; }
hasnt() { grep -Fq -- "$2" "$1" && fail "$1 still has '$2'"; return 0; }
# section <file> <start line regex> <end line regex>: the lines from the start
# up to, not including, the next line matching the end.
section() { awk -v s="$2" -v e="$3" 'p && $0 ~ e { exit } $0 ~ s { p = 1 } p' "$1"; }

D=docs/domain/domain-model.md
grep -E '^\| recorded defect \|' "$D" | grep -Fq 'D<YYYYMMDD-HHMMSS>-<slug>' || fail "$D: the Identifiers table's recorded defect is not D<YYYYMMDD-HHMMSS>-<slug>"
grep -E '^\| intervention \|' "$D" | grep -Fq 'I<YYYYMMDD-HHMMSS>-<slug>' || fail "$D: the Identifiers table's intervention is not I<YYYYMMDD-HHMMSS>-<slug>"
section "$D" '^- \*\*M-4\*\*' '^- \*\*M-5\*\*' >"$t/m4"
[ -s "$t/m4" ] || fail "$D has no M-4"
grep -q 'every other catcher is before it' "$t/m4" && fail "$D: M-4 still decides escaped by catcher alone"
grep -q 'M-5' "$t/m4" || fail "$D: M-4 does not decide escaped against M-5's release"
section "$D" '^- \*\*M-6\*\*' '^- \*\*M-7\*\*' >"$t/m6"
grep -q 'intervention' "$t/m6" || fail "$D: M-6's closing commit does not carry intervention files"
grep -q 'defect files' "$t/m6" && fail "$D: M-6's closing commit still carries defect files"

F=docs/domain/measurement/defect.md
grep -q '^#.*The rule that caught it' "$F" && fail "$F still has 'The rule that caught it'"
has "$F" B15
hasnt "$F" untriaged
grep -qi 'escape' "$F" || fail "$F does not say a recorded defect is an escape"

for m in docs/domain/measurement/metrics.md docs/guide/metrics.md; do
  has "$m" '$/kloc'
  has "$m" 'test:code'
  has "$m" verification-finding
  grep -Eq '(^|[^a-z])t:c([^a-z]|$)' "$m" && fail "$m still names the t:c column"
  hasnt "$m" '$/1k'
  hasnt "$m" 'per 1,000 code lines'
done

F=docs/guide/defects.md
for w in verification-finding --origin --blame; do has "$F" "$w"; done
grep -qi 'escape' "$F" || fail "$F does not say what to record as a defect: escapes"
hasnt "$F" untriaged

O=plugin/skills/operate/SKILL.md
hasnt "$O" untriaged
hasnt "$O" rule-outcome
hasnt "$O" 'defects of'
grep -Eq -- '--finding "[^"]*(brief|plan|work|env|origin)[^"]*: ' "$O" || fail "$O: the close step's --finding does not take its origin"
[ "$(wc -w <"$O" | tr -d ' ')" -le 1750 ] || fail "$O is over its ceiling of 1,750 words"

P=.claude/skills/vloop-operator/SKILL.md
section "$P" '^## 5\. Closing and merging' '^## ' >"$t/p5"
[ -s "$t/p5" ] || fail "$P has no '## 5. Closing and merging'"
grep -q 'verification-finding' "$t/p5" || fail "$P step 5 does not record findings before merge as verification-finding interventions"
grep -q -- '--origin' "$t/p5" || fail "$P step 5 does not give the finding's --origin"
grep -qi 'escape' "$t/p5" || fail "$P step 5 does not keep defects for escapes"

C=docs/guide/commands.md
GOFLAGS= go run ./cmd/gendocs "$t/commands.md" >"$t/gen.out" 2>&1 || fail "go run ./cmd/gendocs: $(cat "$t/gen.out")"
cmp -s "$t/commands.md" "$C" || fail "$C is not what cmd/gendocs makes now: regenerate it"
hasnt "$C" rule-outcome
echo "gate T13: ok"
