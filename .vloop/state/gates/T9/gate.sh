#!/bin/sh
# Gate T9 — K1: a brief written by vloop brief new (en and es) states in its
# Constraints, beside "No task may weaken a check to pass", that an earlier
# test changes only where it pins behaviour the contract changes, and only to
# the new behaviour. The work and review skills the plugin hands to sessions
# carry the rule (the review's goalpost check now licensed by the contract),
# each within its word ceiling; the architect skill no longer asks for a list
# of earlier tests that must change.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
basebin
# section <file> <heading>: the lines of that level-2 section.
section() { awk -v h="$2" '$0 == h { p = 1; next } p && /^## / { exit } p' "$1"; }
# newbrief <bin> <out>: vloop brief new in $R with that binary; the file in $t/<out>.
newbrief() {
  inr "$1" brief new "demo-$2" </dev/null >"$t/$2.p" 2>&1 || fail "brief new: $(cat "$t/$2.p")"
  cp "$R/$(tail -n 1 "$t/$2.p")" "$t/$2" || fail "brief new printed '$(tail -n 1 "$t/$2.p")'"
}

newrepo en
newbrief "$B" en-new; newbrief "$t/vloop-base" en-base
section "$t/en-new" '## Constraints' >"$t/cn"; section "$t/en-base" '## Constraints' >"$t/cb"
grep -q 'No task may weaken a check to pass' "$t/cn" || fail "the en Constraints lost 'No task may weaken a check to pass'"
grep -Fxv -f "$t/cb" "$t/cn" >"$t/cadd"
grep -qi 'earlier test' "$t/cadd" && grep -qi 'new behaviour' "$t/cadd" && grep -qi 'contract' "$t/cadd" ||
  fail "the en template's Constraints do not state that an earlier test changes only where the contract changes what it pins, and only to the new behaviour; added lines:
$(cat "$t/cadd")"

newrepo es --language es
newbrief "$B" es-new; newbrief "$t/vloop-base" es-base
section "$t/es-new" '## Restricciones' >"$t/sn"; section "$t/es-base" '## Restricciones' >"$t/sb"
grep -q 'Ninguna tarea puede debilitar una comprobación' "$t/sn" || fail "the es Restricciones lost 'Ninguna tarea puede debilitar una comprobación'"
grep -Fxv -f "$t/sb" "$t/sn" >"$t/sadd"
grep -qi 'prueba' "$t/sadd" && grep -qi 'contrato' "$t/sadd" || fail "the es template's Restricciones do not state the rule (an earlier test — prueba — changed only as the contrato licenses); added lines:
$(cat "$t/sadd")"

inr "$B" plugin path </dev/null >"$t/pp" 2>&1 || fail "vloop plugin path: $(cat "$t/pp")"
p=$(tail -n 1 "$t/pp")
case "$p" in /*) P=$p ;; *) P=$R/$p ;; esac
for k in work:1800 review:1650; do
  s=${k%:*} c=${k#*:} F=$P/skills/${k%:*}/SKILL.md
  [ -f "$F" ] || fail "no skills/$s/SKILL.md in the extracted plugin"
  git -C "$top" show "$BASE:plugin/skills/$s/SKILL.md" >"$t/$s.base" || fail "the base's $s skill"
  grep -Fxv -f "$t/$s.base" "$F" >"$t/$s.add"
  grep -qi 'earlier test' "$t/$s.add" && grep -qi 'contract' "$t/$s.add" || fail "the $s skill states no rule for changing an earlier test under the contract; added lines:
$(cat "$t/$s.add")"
  w=$(wc -w <"$F" | tr -d ' ')
  [ "$w" -le "$c" ] || fail "the $s skill is $w words, over its ceiling of $c"
done
grep -qi 'goalpost' "$P/skills/review/SKILL.md" || fail "the review skill no longer names a goalpost move"
grep -qi 'notes' "$t/work.add" || fail "the work skill does not say to name the contract item in the proposal's notes"
grep -Eqi 'cite|item' "$t/review.add" || fail "the review skill does not have the verdict cite the contract item"
A=.claude/skills/vloop-architect/SKILL.md
grep -Fq 'earlier tests that *must* change' "$A" && fail "$A still asks for an explicit list of earlier tests that must change"
grep -qi 'earlier test' "$A" || fail "$A does not state the rule for earlier tests"
echo "gate T9: ok"
