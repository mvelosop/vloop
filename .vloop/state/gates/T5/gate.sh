# T5 (K4): the README, the concepts guide and the domain pages name the two
# brief skills, by the text of the sections the brief names.
G=.vloop/state/gates/T5
. "$G/lib.sh"
cd "$top" || fail cd

# section <file> <heading line> : the lines after the heading, up to the next
# heading of the same or a higher level.
section() {
  awk -v h="$2" '
    $0 == h { on = 1; lvl = index(h, " "); next }
    on && /^#+ / { if (index($0, " ") <= lvl) exit }
    on { print }
  ' "$1"
}
names() {
  for s in $3; do
    printf '%s\n' "$2" | grep -qF "/vloop:$s" || fail "$1 does not name /vloop:$s"
  done
}
ALL="plan gate-review work review verify operate create-brief review-brief"

s=$(section README.md '### Skills')
[ -n "$s" ] || fail "README.md has no ### Skills section"
names "README.md's Skills section" "$s" "$ALL"

s=$(section docs/guide/concepts.md '## Skills')
[ -n "$s" ] || fail "docs/guide/concepts.md has no ## Skills section"
names "docs/guide/concepts.md's Skills section" "$s" "$ALL"

s=$(section docs/domain/execution/session.md '## Skills, as built')
[ -n "$s" ] || fail "docs/domain/execution/session.md has no ## Skills, as built section"
names "session.md's Skills, as built" "$s" "create-brief review-brief"

D=docs/domain/domain-model.md
s=$(awk '/^- A \*\*Brief\*\*/ { on = 1; print; next } on && (/^- / || /^$/ || /^#/) { exit } on { print }' "$D")
[ -n "$s" ] || fail "$D has no '- A **Brief**' paragraph"
names "$D's Brief paragraph" "$s" "create-brief review-brief"
s=$(grep -E '^\| \*\*design act\*\* \|' "$D")
[ -n "$s" ] || fail "$D's glossary has no **design act** row"
names "$D's glossary row for the design act" "$s" "create-brief review-brief"
echo "T5 gate: ok"
