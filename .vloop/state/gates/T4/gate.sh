#!/bin/sh
# Gate T4 — vloop brief check reads `## Pinning tests` in .vloop/repo-guidelines.md
# (brief L3): a rule whose phrase the Behaviour contract contains, and whose
# names Constraints do not all contain, is one warning naming the missing ones
# in the rule's order; sections by the repository's heading set, never the
# whole brief; no file, nothing; a warning never changes the exit code. And the
# real-data check: four consumed briefs warn naming the tests their recorded
# defects say they omitted.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
. "$G/briefs.sh"
W='must-change list: the contract mentions "config key" but Constraints do not name'
ALL='TestConfigListDefaults, TestConfigListJSONOrderAndNull, TestDefaultsEveryRow, TestWorkedExampleEnglishSession'
TAIL='(.vloop/repo-guidelines.md)'

newrepo en
# The seed, with two malformed lines under its pinning tests: they are ignored.
awk '{ print } $0 == "## Pinning tests" { print ""; print "- when config key: TestBroken"; print "- when `config key` `TestBroken2`" }' "$G/seed.md" >"$R/$GF" || fail seed

wbrief 'A thing.' 'It adds a config key.' '- Keep TestConfigListDefaults green.'
A=$F
check a "$A"
[ "$got" = 0 ] || fail "a: a brief with only a warning exits $got, not 0: $(cat "$t/a.raw")"
has a "! $W TestConfigListJSONOrderAndNull, TestDefaultsEveryRow, TestWorkedExampleEnglishSession $TAIL"
[ "$(grep -c '^! must-change list:' "$t/a")" = 1 ] || fail "a: not exactly one must-change warning: $(cat "$t/a.raw")"
grep -q 'TestBroken' "$t/a" && fail "a: a malformed rule was read: $(cat "$t/a.raw")"

wbrief 'A thing.' 'It adds a config key.' "- Keep $ALL green."
check b "$F"; none b 'must-change list:'

wbrief 'It adds a config key.' 'Nothing else.' '- Keep TestConfigListDefaults green.'
check c "$F"; none c 'must-change list:'

wbrief 'A thing.' 'It ADDS A Config Key.' '- Keep TestConfigListDefaults green.'
check d "$F"
has d "! $W TestConfigListJSONOrderAndNull, TestDefaultsEveryRow, TestWorkedExampleEnglishSession $TAIL"

# The contract runs through its ### subsections; the names count only in Constraints.
wbrief "Tests $ALL." '### L1 — keys

The brief adds a config key.

### Decided here

Nothing.' '- Nothing to keep.'
check e "$F"
has e "! $W $ALL $TAIL"

# The phrase only in the worked example is not the contract.
wbrief 'A thing.' 'Nothing.' '- Nothing.'
sed 's/^\$ demo$/$ demo --config key/' "$R/$F" >"$t/f.md" && mv "$t/f.md" "$R/$F" || fail sed
check f "$F"; none f 'must-change list:'

# No guidelines file: the output is today's, byte for byte, as the plan's base builds it.
rm "$R/$GF"
check nofile "$A"
none nofile 'must-change list:'
basebin
inr "$t/vloop-base" brief check "$A" </dev/null >"$t/today" 2>&1
cmp -s "$t/nofile.raw" "$t/today" || fail "with no guidelines file the report is not today's:
$(diff "$t/today" "$t/nofile.raw")"

# Spanish: the sections are found by the repository's heading set.
R="$t/es"
mkdir -p "$R/docs/briefs" && git -C "$R" init -q -b main || fail "git init es"
printf 'x\n' >"$R/docs/x.md"
inr "$B" init --language es </dev/null >"$t/es.init" 2>&1 || fail "vloop init --language es: $(cat "$t/es.init")"
cp "$G/seed.md" "$R/$GF" || fail seed
cat >"$R/docs/briefs/B20260101-0900-es.loop-brief.md" <<'B'
---
name: B20260101-0900-es.loop-brief
description: Una demo
kind: brief
status: ready
created: 2026-01-01
seeds: Una demo
depends-on: []
---
# Brief

## Qué es

Una cosa.

## Contrato de comportamiento

Añade una config key.

## Restricciones

- TestConfigListDefaults.
B
check esb docs/briefs/B20260101-0900-es.loop-brief.md
has esb "! $W TestConfigListJSONOrderAndNull, TestDefaultsEveryRow, TestWorkedExampleEnglishSession $TAIL"

# The real-data check: copies of four consumed briefs, set to ready, with the seed.
newrepo real
cp "$G/seed.md" "$R/$GF" || fail seed
mkdir -p "$R/docs/briefs"
for b in B20261003-2049-gate-model B20261004-1434-interventions-options B20261008-2144-b12-field-reliability B20261011-2100-b13-docs-onboarding; do
  awk '!d && $0 == "status: consumed" { print "status: ready"; d = 1; next } { print }' "$top/docs/briefs/$b.loop-brief.md" >"$R/docs/briefs/$b.loop-brief.md" || fail "copy $b"
  grep -Fxq 'status: ready' "$R/docs/briefs/$b.loop-brief.md" || fail "$b: no 'status: consumed' to set to ready"
  check "$b" "docs/briefs/$b.loop-brief.md"
done
P='! must-change list: the contract mentions'
has B20261003-2049-gate-model "$P \"schemas/\" but Constraints do not name TestWorkedExampleB2Commands, TestSchemasAreEmbedded, TestWorkedExampleB6RealData $TAIL"
has B20261004-1434-interventions-options "$P \"schemas/\" but Constraints do not name TestSchemasAreEmbedded, TestWorkedExampleB6RealData $TAIL"
has B20261008-2144-b12-field-reliability "$P \"config key\" but Constraints do not name TestConfigListDefaults, TestConfigListJSONOrderAndNull, TestDefaultsEveryRow $TAIL"
has B20261011-2100-b13-docs-onboarding "$P \"closing lines\" but Constraints do not name TestInitPointsAtBothWaysToLoadThePlugin $TAIL"
echo "gate T4: ok"
