#!/bin/sh
# Gate T2 — the gate review (G3) and its ceiling.
# plugin/skills/gate-review/SKILL.md, joined (newlines turned into spaces, runs
# of whitespace collapsed to one), holds the worked example's two phrases and
# equals the joined form of gate-review-expected.md in this folder: the base's
# skill with exactly G3's two edits, wrapping free. Its wc -w is at most 1600.
# internal/budget/budget.go's Ceilings entry for "gate-review" is 1600 with the
# brief's comment (the brief names the file, the entry, the value and the text).
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
G=$(cd "$(dirname "$0")" && pwd -P) || fail "no gate folder"
F=plugin/skills/gate-review/SKILL.md
B=internal/budget/budget.go
[ -f "$F" ] || fail "$F is missing"
[ -f "$B" ] || fail "$B is missing"
t=$(mktemp -d "${TMPDIR:-/tmp}/gate.XXXXXX") || fail mktemp
trap 'rm -rf "$t"' EXIT
joined() { tr -s ' \t\n' '   ' <"$1" | sed -e 's/^ //' -e 's/ $//'; }
joined "$F" >"$t/got"
joined "$G/gate-review-expected.md" >"$t/want"
for s in 'or internals the brief does not name (beyond a sound pin, as below),' \
  "A pin is sound only where no surface the brief names (a command, a route, a file) could judge the same criterion, and its task's acceptance states the name."; do
  grep -qF -- "$s" "$t/got" || fail "$F does not say: $s"
done
if ! cmp -s "$t/got" "$t/want"; then
  tr ' ' '\n' <"$t/got" >"$t/gw"; tr ' ' '\n' <"$t/want" >"$t/ww"
  fail "$F differs from the brief's edits elsewhere (word diff, < got, > want):
$(diff "$t/gw" "$t/ww" | head -20)"
fi
n=$(wc -w <"$F" | tr -d ' ')
[ "$n" -le 1600 ] || fail "$F has $n words, over 1600"
# The Ceilings map's "gate-review" entry, whitespace collapsed.
e=$(awk '/^var Ceilings = map\[string\]int\{/ { f = 1; next } f && /^\}/ { exit } f && /"gate-review":/ { print }' "$B" | tr -s ' \t' '  ' | sed -e 's/^ //' -e 's/ $//')
[ -n "$e" ] || fail "$B: Ceilings has no \"gate-review\" entry"
want='"gate-review": 1600, // 2026-11-02, B32 (#84): 1550 (B14) plus the sound-pin sentence of pins-unnamed and its reference in not-contract'
[ "$e" = "$want" ] || fail "$B: Ceilings[\"gate-review\"] is
$e
--- want
$want"
echo "gate T2: ok"
