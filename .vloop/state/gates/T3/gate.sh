#!/bin/sh
# Gate T3 — vloop init ends with D3's seven lines, exactly, and the CLAUDE.md
# section it writes carries D3's interactive paragraph, the rest unchanged.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
newrepo r
tail -n 7 "$t/r.init" >"$t/next" || fail tail
cmp -s "$t/next" "$G/next.txt" || fail "vloop init's closing lines are not D3's; got:
$(cat "$t/next")"
grep -c '^next:$' "$t/r.init" | grep -qx 1 || fail "vloop init prints next: more than once"
awk '/<!-- vloop:begin -->/{p=1} p{print} /<!-- vloop:end -->/{exit}' "$R/CLAUDE.md" >"$t/sec" || fail awk
grep -Fxq -f "$G/paragraph.txt" "$t/sec" || fail "the CLAUDE.md section has no line equal to D3's interactive paragraph"
sed -e "s/^Written by vloop $V\.\$/Written by vloop VERSION./" -e "s/^In an interactive session you are the operator's hands: .*/PARAGRAPH/" "$t/sec" >"$t/norm" || fail sed
cmp -s "$t/norm" "$G/section.txt" || fail "the rest of the CLAUDE.md section changed:
$(diff "$G/section.txt" "$t/norm")"
echo "gate T3: ok"
