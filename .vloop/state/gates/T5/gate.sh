#!/bin/sh
# Gate T5 — J1's playbooks and the guide. The operate skill and this
# repository's playbook cite P-8 and do not restate it, cite no P-id the domain
# model does not define, name .vloop/config.toml (where the operator adds the
# check a criterion needs), and the row for exit 2 of each halt table speaks of
# the check-gap halt. docs/guide/concepts.md's "The gate review" paragraph and
# its exit-code row for 2 speak of it too (the gate review's part runs to the
# next bold-led paragraph).
set -u
fail() { echo "GATE FAIL: $*" >&2; exit 1; }
export LC_ALL=C
DM=docs/domain/domain-model.md
flat() { tr '\n' ' ' <"$1" | tr -s ' \t' '  '; }
cites() {
  [ -f "$1" ] || fail "no $1"
  grep -qF 'P-8' "$1" || fail "$1 does not cite P-8"
  f=$(flat "$1")
  for w in "judges the brief's contract from outside the work" "a check, or a test a task writes"; do
    case "$f" in *"$w"*) fail "$1 restates P-8 ('$w') instead of citing it";; esac
  done
  for id in $(grep -oE 'P-[0-9]+' "$1" | sort -u); do
    grep -qF "**$id**" "$DM" || fail "$1 cites $id, which $DM does not define"
  done
}
# gap <text>: the text speaks of a check gap (check gap, check-gap, check gaps).
gap() { echo "$1" | grep -qiE 'check[- ]gaps?'; }
for P in plugin/skills/operate/SKILL.md .claude/skills/vloop-operator/SKILL.md; do
  cites "$P"
  grep -qF '.vloop/config.toml' "$P" || fail "$P does not name .vloop/config.toml, where the operator adds a missing check"
  row=$(grep -E '^\| *`?2`? *\|' "$P" | head -1)
  [ -n "$row" ] || fail "$P has no halt-table row for exit 2"
  gap "$row" || fail "$P's halt-table row for exit 2 does not speak of the check-gap halt: $row"
done
CG=docs/guide/concepts.md
# "The gate review": from its bold lead to the next bold lead or heading.
para=$(awk '/^\*\*The gate review\.\*\*/ { p = 1; print; next } p && (/^\*\*[A-Z]/ || /^#/) { exit } p' "$CG")
[ -n "$para" ] || fail "$CG has no \"The gate review\" paragraph"
gap "$para" || fail "$CG's \"The gate review\" paragraph does not say what a check-gap halt is"
row=$(grep -E '^\| *`2` *\|' "$CG" | head -1)
[ -n "$row" ] || fail "$CG's exit-code table has no row for 2"
gap "$row" || fail "$CG's exit-code row for 2 does not speak of the check-gap halt: $row"
echo "gate T5: ok"
