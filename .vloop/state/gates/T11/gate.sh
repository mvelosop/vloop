#!/bin/sh
# Gate T11 — the defect report at close (brief L12): `vloop defect set` takes
# `rule` (<where>: <which>) and `rule-outcome` (fired|missed|absent|none) and
# refuses anything else with exit 2; `vloop brief close` prints, before its last
# line, `squash-merge with the trailer`, the brief's defects with their rule,
# outcome or untriaged, and recurrences, then the untriaged hint; with no
# defects, `defects of <run id>: none`; --dry-run prints the report and writes
# nothing. The operate skill's close step triages, within its ceiling.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
RULE='brief-check: must-change list'
HINT="vloop defect set <id> rule '<where>: <which>' and rule-outcome fired|missed|absent|none"
# set <want exit> <args...>: vloop defect set in $R, expecting that exit code.
dset() { w=$1; shift; inr "$B" defect set "$@" </dev/null >"$t/set.out" 2>&1; c=$?; [ "$c" = "$w" ] || fail "vloop defect set $*: exit $c, want $w: $(cat "$t/set.out")"; }
# fields <line>: the report line's columns, one per line (columns are two or more spaces apart).
fields() { printf '%s\n' "$1" | sed -e 's/^ *//' -e 's/   */|/g' | tr '|' '\n'; }

runrepo d
# Three earlier defects of another brief, a day apart, for the recurrence.
printf -- '---\nname: B20250101-0900-old.loop-brief\ndescription: An old brief\nkind: brief\nstatus: draft\n---\n# Old\n' >"$R/docs/briefs/B20250101-0900-old.loop-brief.md"
mkdir -p "$R/.vloop/defects" || fail mkdir
for i in 1 2 3; do
  cat >"$R/.vloop/defects/D2025010$i-0900-old-$i.md" <<D
---
id: D2025010$i-0900-old-$i
brief: B20250101-0900-old.loop-brief
origin: brief
found-by: operator
kind: spec-gap
severity: low
status: open
fixed-by: ""
case: ""
created: 2025-01-0${i}T09:00:00Z
---
old defect $i
D
done
commit "old defects"
vrun d.run
[ "$got" = 0 ] || fail "vloop run: exit $got: $(tail -n 20 "$t/d.run")"

# No defects yet: the report says none, and a dry run writes nothing.
inr "$B" brief close "$BRIEF" --dry-run --no-findings </dev/null >"$t/none" 2>&1 || fail "close --dry-run --no-findings: $(cat "$t/none")"
grep -Fxq "defects of $RUN: none" "$t/none" || fail "a brief with no defects: no 'defects of $RUN: none' in:
$(cat "$t/none")"

inr "$B" defect add "the must-change list missed a test" --brief "$RUN.loop-brief" --found-by review --kind spec-gap --origin brief </dev/null >"$t/add" 2>&1 || fail "defect add: $(cat "$t/add")"
id1=$(basename "$(head -n 1 "$t/add")" .md)
f1="$R/.vloop/defects/$id1.md"
[ -f "$f1" ] || fail "defect add printed '$(cat "$t/add")', not a defect file"
dset 0 "$id1" rule "$RULE"
dset 0 "$id1" rule-outcome missed
cp "$f1" "$t/before"
dset 2 "$id1" rule-outcome forgot
dset 2 "$id1" rule 'linter: x'
cmp -s "$f1" "$t/before" || fail "a refused defect set changed the record"
for w in 'guideline: D20261011-1455' 'brief-check: x' 'gate: x' 'gate-review: unpassable' 'review: x' 'skill: x'; do dset 0 "$id1" rule "$w"; done
for o in fired missed absent none; do dset 0 "$id1" rule-outcome "$o"; done
dset 0 "$id1" rule "$RULE"
dset 0 "$id1" rule-outcome missed
for i in 1 2 3; do dset 0 "D2025010$i-0900-old-$i" rule "$RULE"; done
commit "triage"

inr "$B" brief close "$BRIEF" --dry-run --finding a --finding b </dev/null >"$t/dry" 2>&1 || fail "close --dry-run: $(cat "$t/dry")"
[ -z "$(git -C "$R" status --porcelain)" ] || fail "close --dry-run wrote: $(git -C "$R" status --porcelain)"
grep -Fxq "defects of $RUN" "$t/dry" || fail "close --dry-run prints no report: $(cat "$t/dry")"

inr "$B" brief close "$BRIEF" --finding a --finding b </dev/null >"$t/close" 2>&1 || fail "close: exit $?: $(cat "$t/close")"
tail -n 1 "$t/close" | grep -q '^squash-merge with the trailer: ' || fail "close's last line is not the squash-merge line: $(tail -n 1 "$t/close")"
h=$(grep -Fxn "defects of $RUN" "$t/close" | cut -d: -f1)
[ -n "$h" ] || fail "close prints no 'defects of $RUN':
$(cat "$t/close")"
# The report: the lines after the header, up to the untriaged hint.
awk -v h="$h" 'NR <= h { next } !/^ / { exit } !/untriaged: vloop defect set/ { print }' "$t/close" >"$t/rows"
[ "$(wc -l <"$t/rows" | tr -d ' ')" = 3 ] || fail "the report does not have one line per defect of the brief (3):
$(cat "$t/close")"
short1=$(printf '%s' "$id1" | cut -d- -f1-2)
r1=$(grep -E "^ +($id1|$short1) " "$t/rows") || fail "no report line for $id1: $(cat "$t/rows")"
fields "$r1" >"$t/f1"
printf '%s\n' spec-gap review "$RULE" missed 'recurs ×3 (D20250101-0900, D20250102-0900, D20250103-0900)' >"$t/want1"
tail -n +2 "$t/f1" | cmp -s - "$t/want1" || fail "the report line for $id1 is '$r1'; want kind, found-by, rule, outcome and recurs:
$(cat "$t/want1")"
grep -v -E "^ +($id1|$short1) " "$t/rows" >"$t/new"
while IFS= read -r l; do
  fields "$l" | tail -n +2 >"$t/fn"
  printf '%s\n' bug operator '—' untriaged | cmp -s - "$t/fn" || fail "a finding's report line is '$l'; want kind bug, found-by operator, — and untriaged"
done <"$t/new"
awk -v h="$h" 'NR > h' "$t/close" | sed 's/^ *//' | grep -Fxq "2 untriaged: $HINT" || fail "no '2 untriaged: $HINT' line:
$(cat "$t/close")"
hint=$(grep -Fn "untriaged: $HINT" "$t/close" | cut -d: -f1)
[ "$hint" -lt "$(wc -l <"$t/close" | tr -d ' ')" ] || fail "the untriaged hint is not before the squash-merge line"
d1=$(grep -E "^ +($id1|$short1) " "$t/dry") || fail "close --dry-run's report has no line for $id1: $(cat "$t/dry")"
[ "$(fields "$d1")" = "$(fields "$r1")" ] || fail "close --dry-run's report differs on $id1: '$d1' against '$r1'"

inr "$B" defect set --help >"$t/help" 2>&1
grep -Fq 'rule-outcome' "$t/help" || fail "vloop defect set --help does not list rule-outcome"
grep -q '[^-]rule[^-]' "$t/help" || fail "vloop defect set --help does not list rule"

O=$top/plugin/skills/operate/SKILL.md
[ "$(wc -w <"$O" | tr -d ' ')" -le 1750 ] || fail "the operate skill is over its ceiling of 1,750 words"
grep -Fq 'rule-outcome' "$O" || fail "the operate skill's close step does not triage with rule-outcome"
grep -Fq 'untriaged' "$O" || fail "the operate skill does not speak of untriaged defects"
echo "gate T11: ok"
