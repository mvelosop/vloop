#!/bin/sh
# Gate T5 — vloop brief check reads `## External formats` in
# .vloop/repo-guidelines.md (brief L4): a rule whose phrase the Behaviour
# contract contains, and whose path Binding references do not cite, is one
# warning; sections, not the whole brief; no file, nothing; a warning never
# changes the exit code.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
. "$G/briefs.sh"
W='! external format: the contract mentions "eval case" but Binding references do not cite docs/guide/evals.md (.vloop/repo-guidelines.md)'

newrepo en
cp "$G/seed.md" "$R/$GF" || fail seed
mkdir -p "$R/docs/guide" && printf '# Evals guide\n' >"$R/docs/guide/evals.md" || fail evals.md

wbrief 'A thing.' 'It adds a new eval case.' ''
A=$F
check a "$A"
[ "$got" = 0 ] || fail "a: a brief with only a warning exits $got, not 0: $(cat "$t/a.raw")"
has a "$W"
[ "$(grep -c '^! external format:' "$t/a")" = 1 ] || fail "a: not exactly one external format warning: $(cat "$t/a.raw")"

wbrief 'A thing.' 'It adds a new eval case.' '' '- `docs/guide/evals.md` — the eval case format'
check b "$F"; none b 'external format:'

wbrief 'A thing.' 'It adds a new EVAL Case.' '- Follow `docs/guide/evals.md`.'
check c "$F"; has c "$W"

wbrief 'It adds a new eval case.' 'Nothing else.' ''
check d "$F"; none d 'external format:'

# The contract runs through its ### subsections.
wbrief 'A thing.' '### L6 — a case

A new eval case under plugin/evals/.' ''
check e "$F"; has e "$W"

# No guidelines file: the output is today's, byte for byte, as the plan's base builds it.
rm "$R/$GF"
check nofile "$A"
none nofile 'external format:'
basebin
inr "$t/vloop-base" brief check "$A" </dev/null >"$t/today" 2>&1
cmp -s "$t/nofile.raw" "$t/today" || fail "with no guidelines file the report is not today's:
$(diff "$t/today" "$t/nofile.raw")"
echo "gate T5: ok"
