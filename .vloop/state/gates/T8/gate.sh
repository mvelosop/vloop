#!/bin/sh
# Gate T8 — D1: concepts.md's "The stamped release build" says the version
# comes from plugin.json, the stamped build sets the commit only, and the
# release is prepare, the operator's merge, publish, then install and upgrade;
# the operator skill runs the tool and records the release intervention;
# tools/release has a README under 300 words.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
xbuild

f=docs/guide/concepts.md
awk '/^### The stamped release build/ { p = 1; print; next } p && /^#/ { exit } p { print }' "$f" >"$t/sec"
[ -s "$t/sec" ] || fail "$f has no section 'The stamped release build'"
grep -q 'plugin\.json' "$t/sec" || fail "the section does not say the version comes from plugin.json"
! grep -q 'main\.version' "$t/sec" || fail "the section still stamps main.version: $(grep -n 'main\.version' "$t/sec")"
grep -q 'main\.commit' "$t/sec" || fail "the section's stamped build does not set the commit (-X main.commit)"
first() { grep -n -- "$1" "$t/sec" | head -n 1 | cut -d: -f1; }
p=$(first 'tools/release prepare'); u=$(first 'tools/release publish')
[ -n "$p" ] || fail "the section does not name go run ./tools/release prepare"
# The merge and the upgrade are the first mentions after prepare and publish.
m=$(awk -v p="$p" 'NR >= p && /[Mm]erge/ { print NR; exit }' "$t/sec")
g=$(awk -v u="${u:-0}" 'NR >= u && /vloop upgrade/ { print NR; exit }' "$t/sec")
[ -n "$u" ] || fail "the section does not name go run ./tools/release publish"
[ -n "$g" ] || fail "the section does not name vloop upgrade after publish"
[ -n "$m" ] || fail "the section does not name the operator's merge after prepare"
[ "$p" -le "$m" ] && [ "$m" -le "$u" ] && [ "$u" -le "$g" ] ||
  fail "the section's steps are not prepare, merge, publish, then upgrade (lines $p, $m, $u, $g)"

k=.claude/skills/vloop-operator/SKILL.md
l=$(grep -n 'tools/release' "$k" | head -n 1 | cut -d: -f1)
[ -n "$l" ] || fail "$k does not run the release tool"
# The step: from its first mention of the tool, at most 20 lines, within its section.
awk -v l="$l" 'NR > l && /^#/ { exit } NR >= l && NR < l + 20 { print }' "$k" >"$t/ksec"
grep -q 'tools/release publish' "$t/ksec" || fail "the skill's release step does not run publish"
grep -q 'intervention' "$t/ksec" || fail "the skill's release step does not record the release intervention"

r=tools/release/README.md
[ -f "$r" ] || fail "no $r"
w=$(wc -w <"$r" | tr -d ' ')
[ "$w" -lt 300 ] || fail "$r has $w words, want under 300"
for c in notes prepare publish; do grep -q "$c" "$r" || fail "$r does not describe $c"; done
echo "gate T8: ok"
