#!/bin/sh
# Gate T9 — the skills' word budgets (brief L10), judged on the contract only:
# each plugin skill, counted as wc -w counts the whole SKILL.md, is within the
# ceiling the brief pins; and the five ceilings live beside a test, one comment
# each. The test's failure message is the [[check]]'s and the review's to judge;
# no gate runs the repository's tests.
# Operator ruling (2026-10-12): replaces the gate that called
# internal/budget.Check, an implementation detail the gate review rightly
# refused; the planner owned the defect, not the brief.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
CEIL='plan 4100
gate-review 1550
work 1800
review 1650
operate 1750'

# The real skills, within their ceilings.
echo "$CEIL" | while read -r k c; do
  f="plugin/skills/$k/SKILL.md"
  [ -f "$f" ] || fail "$f does not exist"
  n=$(wc -w <"$f" | tr -d ' ')
  [ "$n" -le "$c" ] || fail "$f has $n words, over its ceiling of $c"
done || exit 1

# The ceilings beside a test: a Go package folder holding a _test.go file and,
# in its other files, each "<skill>": <ceiling> with a comment of its own (on
# the line or the line above).
found=
for d in $(find . -name '*_test.go' -not -path './.git/*' -not -path './.vloop/*' | sed 's|/[^/]*$||' | sort -u); do
  ok=1
  for k in plan gate-review work review operate; do
    c=$(echo "$CEIL" | awk -v k="$k" '$1 == k { print $2 }')
    awk -v k="$k" -v c="$c" '
      { line[NR] = $0 }
      index($0, "\"" k "\":") && $0 ~ ("[^0-9]" c "([^0-9]|$)") {
        if ($0 ~ /\/\/./ || line[NR-1] ~ /^[ \t]*\/\/./) { hit = 1 }
        else { bare = 1 }
      }
      END { exit (hit && !bare) ? 0 : 1 }' $(ls "$d"/*.go | grep -v '_test\.go$') 2>/dev/null || { ok=; break; }
  done
  [ -n "$ok" ] && { found=$d; break; }
done
[ -n "$found" ] || fail "no package beside a test holds the five ceilings (plan 4100, gate-review 1550, work 1800, review 1650, operate 1750), each with a comment of its own"
echo "gate T9: ok (ceilings in $found)"
