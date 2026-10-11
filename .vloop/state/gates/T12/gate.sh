#!/bin/sh
# Gate T12 — the docs (brief L8, with L9–L12): docs/guide/concepts.md gains
# `## Repository guidelines` and the .vloop/ layout lists the file; the brief
# domain doc lists the two warnings; the defect and metrics domain docs carry
# the defect's rule fields and the instructions object.
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
C=docs/guide/concepts.md
# sect <file> <heading>: the lines under that exact heading, up to the next ## heading.
sect() { awk -v h="$2" 'p && /^## / { exit } p { print } $0 == h { p = 1 }' "$1"; }
# need <file> <section text> <section name> <string>...: each string is in the section.
need() {
  f=$1 s=$2 n=$3; shift 3
  for w in "$@"; do
    printf '%s\n' "$s" | grep -Fq -- "$w" || fail "$f, '$n', does not have: $w"
  done
}
grep -Fxq '## Repository guidelines' "$C" || fail "$C has no '## Repository guidelines' heading"
rg=$(sect "$C" '## Repository guidelines')
need "$C" "$rg" 'Repository guidelines' '.vloop/repo-guidelines.md' 'vloop run' '## Pinning tests' '## External formats' \
  '- when `' 'must-change list:' 'external format:' 'vloop defect set <id> fixed-by guideline' 'vloop doctor'
printf '%s\n' "$rg" | grep -Eq '(^|[^0-9])20([^0-9]|$)' || fail "$C, 'Repository guidelines', does not state the budget of 20 rules"
lay=$(sect "$C" '## The .vloop/ layout')
[ -n "$lay" ] || fail "$C has no '## The .vloop/ layout' section"
printf '%s\n' "$lay" | grep -q '^- `\.vloop/repo-guidelines\.md`' || fail "$C: the .vloop/ layout list does not list .vloop/repo-guidelines.md"
D=docs/domain/briefing/brief.md
need "$D" "$(cat "$D")" "$D" 'must-change list' 'external format'
D=docs/domain/measurement/defect.md
need "$D" "$(cat "$D")" "$D" 'rule-outcome' 'fired' 'missed' 'absent' 'untriaged' 'guideline' 'brief-check' 'gate-review' 'skill'
D=docs/domain/measurement/metrics.md
need "$D" "$(cat "$D")" "$D" 'instructions' 'guidelines_rules'
echo "gate T12: ok"
