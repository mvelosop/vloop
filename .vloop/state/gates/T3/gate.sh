#!/bin/sh
# Gate T3 — the skills (brief L1, L5, L10): plan, gate-review, work and review
# each say the repository's guidelines may be appended to the session; the
# gate review's `unpassable` names a gate that needs a session to write under
# .vloop/ outside .vloop/tmp/ (S-4), and its check 1 runs and reads the gate in
# the plan's shell with the host's tools, BSD grep, sed, awk and date on macOS
# (P-5); every skill stays within the ceiling L10 sets.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
S=plugin/skills
for k in plan gate-review work review; do
  [ -f "$S/$k/SKILL.md" ] || fail "no $S/$k/SKILL.md"
  grep -iq 'guidelines' "$S/$k/SKILL.md" || fail "$S/$k/SKILL.md says nothing about the repository's guidelines"
done
GR=$S/gate-review/SKILL.md
# The `unpassable` entry: from its bullet to the next bullet or blank line.
up=$(awk 'p && (/^- `/ || /^$/) { exit } p { print } /^- `unpassable`/ { p = 1; print }' "$GR" | tr '\n' ' ')
[ -n "$up" ] || fail "$GR has no '- \`unpassable\`' entry"
for w in '.vloop/' '.vloop/tmp/' 'S-4'; do
  printf '%s\n' "$up" | grep -Fq -- "$w" || fail "$GR: the unpassable entry does not have '$w': $up"
done
# Check 1: from `1. **It fails on the base` to the next numbered item.
c1=$(awk 'p && /^[0-9]\. \*\*/ { exit } p { print } /^1\. \*\*It fails on the base/ { p = 1; print }' "$GR" | tr '\n' ' ')
[ -n "$c1" ] || fail "$GR has no check 1 ('1. **It fails on the base')"
for w in 'shell' 'P-5' 'macOS' 'BSD' 'grep' 'sed' 'awk' 'date' 'unpassable'; do
  printf '%s\n' "$c1" | grep -Fq -- "$w" || fail "$GR: check 1 does not have '$w': $c1"
done
# The ceilings L10 sets.
for kc in plan:4100 gate-review:1550 work:1800 review:1650 operate:1750; do
  k=${kc%%:*} c=${kc#*:}
  n=$(wc -w <"$S/$k/SKILL.md" | tr -d ' ')
  [ "$n" -le "$c" ] || fail "$S/$k/SKILL.md has $n words, over its ceiling of $c"
done
echo "gate T3: ok"
