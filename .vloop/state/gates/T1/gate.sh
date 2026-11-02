#!/bin/sh
# Gate T1 — the plan skill (G1, G2). plugin/skills/plan/SKILL.md, joined (its
# newlines turned into spaces, runs of whitespace collapsed to one), must hold
# the worked example's five phrases and not the old one, and must equal the
# joined form of plan-expected.md in this folder: the base's skill with exactly
# the brief's edits (the "Naming" paragraph replaced, the "Known trap"
# parenthesis, the "Gates are judged" sentence, the two revision-round
# sentences). Wrapping is free; no other word may change.
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
F=plugin/skills/plan/SKILL.md
[ -f "$F" ] || fail "$F is missing"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
joined() { tr -s ' \t\n' '   ' <"$1" | sed -e 's/^ //' -e 's/ $//'; }
joined "$F" >"$t/got"
joined "$G/plan-expected.md" >"$t/want"
for s in '**Where the brief is silent, judge through what it names.**' \
  "state each pinned name in its task's acceptance criteria and report it." \
  'A gate you leave as it was gets the same finding again.' \
  'an import of a module the brief names)' \
  'pins no name the brief does not, beyond a pin "Naming" allows.'; do
  grep -qF -- "$s" "$t/got" || fail "$F does not say: $s"
done
grep -qF -- 'you may pin what a gate needs' "$t/got" && fail "$F still says: you may pin what a gate needs"
if ! cmp -s "$t/got" "$t/want"; then
  tr ' ' '\n' <"$t/got" >"$t/gw"; tr ' ' '\n' <"$t/want" >"$t/ww"
  fail "$F differs from the brief's edits elsewhere (word diff, < got, > want):
$(diff "$t/gw" "$t/ww" | head -20)"
fi
echo "gate T1: ok"
