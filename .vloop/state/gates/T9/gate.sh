#!/bin/sh
# Gate T9 — the skills' word budgets (brief L10). Runs no repository test: its
# judge is oracle.go, from this folder, which calls the one name this plan pins,
# budget.Check(pluginDir) []string in internal/budget — a message per skill over
# its ceiling, "<skill>: <count> words, over its ceiling of <ceiling>". The gate
# writes synthetic skills (frontmatter, mixed whitespace) exactly at and a random
# amount over each ceiling, and counts as wc -w does. It also counts the real
# skills itself and fails when one is over its ceiling.
# Operator ruling (2026-10-12): the gate review's pins-unnamed finding on
# internal/budget.Check is accepted as the plan's seam: the task's acceptance
# names it, and L10's deliverable is a test, which no gate may run (brief gap).
G=$(cd "$(dirname "$0")" && pwd -P) || { echo "GATE FAIL: no gate folder" >&2; exit 1; }
. "$G/lib.sh"
CEIL='plan 4100
gate-review 1550
work 1800
review 1650
operate 1750'
# sep <n>: n with thousands separators.
sep() { echo "$1" | awk '{ s = $1 ""; o = ""; while (length(s) > 3) { o = "," substr(s, length(s) - 2) o; s = substr(s, 1, length(s) - 3) } print s o }'; }

# The real skills, counted here.
echo "$CEIL" | while read -r k c; do
  n=$(wc -w <"plugin/skills/$k/SKILL.md" | tr -d ' ') || exit 1
  [ "$n" -le "$c" ] || { echo "GATE FAIL: plugin/skills/$k/SKILL.md has $n words, over its ceiling of $c" >&2; exit 1; }
done || exit 1

# The oracle, built in a copy of the module (no gate scratch folder exists).
C="$t/tree"
mkdir -p "$C" || fail mkdir
git ls-files -co --exclude-standard >"$t/files" || fail "git ls-files"
while IFS= read -r f; do [ -e "$f" ] && printf '%s\n' "$f"; done <"$t/files" >"$t/present"
tar -cf - -T "$t/present" | tar -xf - -C "$C" || fail "copy of the tree"
mkdir -p "$C/gateoracle" && cp "$G/oracle.go" "$C/gateoracle/main.go" || fail "copy of the oracle"
(cd "$C" && go build -buildvcs=false -o "$t/oracle" ./gateoracle) >"$t/build.log" 2>&1 ||
  fail "the oracle does not build against internal/budget.Check(pluginDir string) []string:
$(cat "$t/build.log")"
for os in linux windows; do
  (cd "$C" && GOOS=$os go build -buildvcs=false -o "$t/oracle-$os" ./gateoracle) >"$t/build.log" 2>&1 ||
    fail "GOOS=$os: the oracle does not build: $(cat "$t/build.log")"
done

# skill <dir> <name> <words>: a SKILL.md of exactly that many words, as wc -w
# counts them, frontmatter included, spaces, tabs and blank lines mixed.
skill() {
  f="$1/skills/$2/SKILL.md"
  mkdir -p "$1/skills/$2" || fail mkdir
  printf -- '---\nname: %s\ndescription: A synthetic skill, for the gate.\n---\n# The %s skill\n\n' "$2" "$2" >"$f"
  h=$(wc -w <"$f" | tr -d ' ')
  awk -v k=$(($3 - h)) 'BEGIN { for (i = 1; i <= k; i++) { printf "w%d", i; if (i % 11 == 0) printf "\n\n"; else if (i % 3 == 0) printf "\t"; else if (i % 5 == 0) printf "   "; else printf " " } printf "\n" }' >>"$f"
  [ "$(wc -w <"$f" | tr -d ' ')" = "$3" ] || fail "fixture: $f is not $3 words"
}

# Every skill exactly at its ceiling: nothing reported.
A="$t/at"
echo "$CEIL" | while read -r k c; do skill "$A" "$k" "$c"; done || exit 1
"$t/oracle" "$A" >"$t/at.out" 2>"$t/at.err" || fail "budget.Check on skills at their ceilings: $(cat "$t/at.err")"
[ -s "$t/at.out" ] && fail "skills exactly at their ceilings are reported:
$(cat "$t/at.out")"

# Every skill over its ceiling by a different, unpredictable amount.
O="$t/over"
s=$(($$ % 97))
i=0
echo "$CEIL" | while read -r k c; do
  i=$((i + 1)); n=$((c + 1 + (s * i * 7 + i) % 60))
  skill "$O" "$k" "$n"
  echo "$k: $(sep "$n") words, over its ceiling of $(sep "$c")"
done >"$t/want" || exit 1
"$t/oracle" "$O" >"$t/over.out" 2>"$t/over.err" || fail "budget.Check on skills over their ceilings: $(cat "$t/over.err")"
sort "$t/want" >"$t/want.s"; sort "$t/over.out" >"$t/got.s"
cmp -s "$t/want.s" "$t/got.s" || fail "budget.Check on skills over their ceilings: want
$(cat "$t/want.s")
got
$(cat "$t/got.s")"

# One over, the rest under: only that one.
U="$t/one"
echo "$CEIL" | while read -r k c; do
  if [ "$k" = work ]; then skill "$U" "$k" $((c + 1)); else skill "$U" "$k" $((c - 1)); fi
done || exit 1
"$t/oracle" "$U" >"$t/one.out" 2>"$t/one.err" || fail "budget.Check, one skill over: $(cat "$t/one.err")"
[ "$(cat "$t/one.out")" = 'work: 1,801 words, over its ceiling of 1,800' ] || fail "one skill over: want only
work: 1,801 words, over its ceiling of 1,800
got
$(cat "$t/one.out")"
echo "gate T9: ok"
